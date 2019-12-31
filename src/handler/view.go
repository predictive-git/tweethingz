package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"
)

var (
	version   = env.MustGetEnvVar("RELEASE", "v0.0.1-manual")
	templates *template.Template
)

// DefaultHandler ...
func DefaultHandler(c *gin.Context) {

	uid, _ := c.Cookie(userIDCookieName)
	if uid != "" {
		logger.Printf("user already authenticated -> view")
		c.Redirect(http.StatusSeeOther, "/view/board")
		return
	}

	c.HTML(http.StatusOK, "index", gin.H{
		"version": version,
	})

}

// DashboardHandler ...
func DashboardHandler(c *gin.Context) {
	forUser := getAuthedUser(c)
	c.HTML(http.StatusOK, "view", gin.H{
		"user":    forUser,
		"version": version,
		"refresh": c.Query("refresh"),
	})
}

// SearchListHandler ...
func SearchListHandler(c *gin.Context) {
	forUser := getAuthedUser(c)
	list, err := store.GetSearchCriteria(c.Request.Context(), forUser.Username)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting search criteria")
		return
	}

	c.HTML(http.StatusOK, "search", gin.H{
		"user":    forUser,
		"version": version,
		"list":    list,
	})
}

// SearchDetailHandler ...
func SearchDetailHandler(c *gin.Context) {
	forUser := getAuthedUser(c)
	id := c.Param("cid")
	if id == "" {
		viewErrorHandler(c, http.StatusInternalServerError, nil, "Search ID required")
		return
	}
	logger.Printf("Search ID: %s", id)

	var detail *store.SearchCriteria
	var err error

	if id == "0" {
		detail = &store.SearchCriteria{}
	} else {
		detail, err = store.GetSearchCriterion(c.Request.Context(), id)
		if err != nil {
			viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting search criteria")
			return
		}
	}

	logger.Printf("Lang: %s", detail.Lang)

	c.HTML(http.StatusOK, "search", gin.H{
		"user":    forUser,
		"version": version,
		"detail":  detail,
	})

}

// SearchDataSubmitHandler ...
func SearchDataSubmitHandler(c *gin.Context) {

	forUser := getAuthedUser(c)
	sc := &store.SearchCriteria{}
	if err := c.ShouldBind(&sc); err != nil {
		logger.Printf("error binding: %v", err)
	}

	if sc.ID == "" {
		sc.ID = store.NewID()
	}

	sc.User = forUser.Username
	sc.SinceID = 0
	sc.ExecutedOn = time.Time{}

	// logger.Printf("Search Criteria: %+v", sc)
	if err := store.SaveSearchCriteria(c.Request.Context(), sc); err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error saving search criteria")
		return
	}

	c.Redirect(http.StatusSeeOther, "/view/search")
	return

}

const tweetPageSize = 10

// TweetHandler ...
func TweetHandler(c *gin.Context) {

	ctx := c.Request.Context()
	forUser := getAuthedUser(c)
	cid := c.Param("cid")
	if cid == "" {
		viewErrorHandler(c, http.StatusBadRequest, nil, "Search query ID required (param: cid)")
		return
	}

	criteria, err := store.GetSearchCriterion(ctx, cid)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, nil, "Error getting search criteria")
		return
	}

	view := c.Query("view")
	if view == "latest" {
		criteria.SinceID = 0
	}

	results, err := worker.GetSearchResults(ctx, forUser, criteria)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting search results")
		return
	}

	logger.Printf("saving criteria %s (since: %d, on: %v)", criteria.Name, criteria.SinceID, criteria.ExecutedOn)
	if err = store.SaveSearchCriteria(ctx, criteria); err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting search criteria")
		return
	}

	data := gin.H{
		"user":     forUser,
		"version":  version,
		"criteria": criteria,
		"results":  results,
	}

	c.HTML(http.StatusOK, "tweet", data)

}

// DayHandler ...
func DayHandler(c *gin.Context) {

	forUser := getAuthedUser(c)
	isoDate := c.Param("day")
	if isoDate == "" {
		viewErrorHandler(c, http.StatusBadRequest, nil, "Day required (param: day)")
		return
	}

	day, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		viewErrorHandler(c, http.StatusBadRequest, nil, fmt.Sprintf("Invalid day parameter format (expected YYYY-MM-DD, got: %s)", isoDate))
		return
	}

	dayState, err := store.GetDailyFollowerState(c.Request.Context(), forUser.Username, day)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting daily follower state")
		return
	}

	followers, err := ToUserEvent(forUser, dayState.NewFollowers, isoDate, store.FollowedEventType)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting new follower events")
		return
	}

	unfollowers, err := ToUserEvent(forUser, dayState.Unfollowers, isoDate, store.UnfollowedEventType)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting unfollower events")
		return
	}

	data := gin.H{
		"user":        forUser,
		"version":     version,
		"date":        isoDate,
		"state":       dayState,
		"followers":   followers,
		"unfollowers": unfollowers,
	}

	c.HTML(http.StatusOK, "day", data)

}

// ToUserEvent retreaves users and builds user events for list of IDs as
func ToUserEvent(forUser *store.AuthedUser, ids []int64, isoDate, eventType string) (list []*store.SimpleUserEvent, err error) {

	list = make([]*store.SimpleUserEvent, 0)

	if len(ids) > 0 {
		users, detailErr := worker.GetTwitterUserDetailsFromIDs(forUser, ids)
		if detailErr != nil {
			return nil, detailErr
		}
		for _, u := range users {
			event := &store.SimpleUserEvent{
				SimpleUser: *u,
				EventDate:  isoDate,
				EventType:  eventType,
				EventUser:  forUser.Username,
			}
			list = append(list, event)
		}
	}

	return

}

func viewErrorHandler(c *gin.Context, code int, err error, msg string) {
	logger.Printf("Error: %v - Msg: %s", err, msg)
	c.HTML(code, "error", gin.H{
		"code": code,
		"msg":  msg,
	})
	c.Abort()
}
