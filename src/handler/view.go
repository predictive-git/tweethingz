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
	"github.com/pkg/errors"
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
	username := getAuthedUsername(c)
	c.HTML(http.StatusOK, "view", gin.H{
		"username": username,
		"version":  version,
		"refresh":  c.Query("refresh"),
	})
}

// SearchListHandler ...
func SearchListHandler(c *gin.Context) {
	username := getAuthedUsername(c)
	list, err := store.GetSearchCriteria(c.Request.Context(), username)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "search", gin.H{
		"username": username,
		"version":  version,
		"list":     list,
	})
}

// SearchDetailHandler ...
func SearchDetailHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	id := c.Param("cid")
	if id == "" {
		viewErrorHandler(c, http.StatusInternalServerError, errors.New("Search ID required"))
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
			viewErrorHandler(c, http.StatusInternalServerError, err)
			return
		}
	}

	logger.Printf("Lang: %s", detail.Lang)

	c.HTML(http.StatusOK, "search", gin.H{
		"username": username,
		"version":  version,
		"detail":   detail,
	})

}

// SearchDataSubmitHandler ...
func SearchDataSubmitHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	sc := &store.SearchCriteria{}
	if err := c.ShouldBind(&sc); err != nil {
		logger.Printf("error binding: %v", err)
	}

	if sc.ID == "" {
		sc.ID = store.NewID()
	}
	sc.User = username

	// logger.Printf("Search Criteria: %+v", sc)
	if err := store.SaveSearchCriteria(c.Request.Context(), sc); err != nil {
		logger.Printf("error saving search criteria: %v", err)
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/view/search")
	return

}

const tweetPageSize = 10

// TweetHandler ...
func TweetHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	cid := c.Param("cid")
	if cid == "" {
		viewErrorHandler(c, http.StatusBadRequest, errors.New("Search query ID required (param: cid)"))
		return
	}

	criteria, err := store.GetSearchCriterion(c.Request.Context(), cid)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var results []*store.SimpleTweet
	var resultsErr error

	sinceKey := c.Query("key")
	if sinceKey == "" {
		results, resultsErr = store.GetSearchResultsForDay(c.Request.Context(), cid, time.Now(), tweetPageSize)
	} else {
		results, resultsErr = store.GetSearchResultsFromKey(c.Request.Context(), sinceKey, tweetPageSize)
	}

	if resultsErr != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	data := gin.H{
		"username": username,
		"version":  version,
		"criteria": criteria,
		"results":  results,
	}

	if len(results) == tweetPageSize {
		data["next_key"] = results[len(results)-1].Key
	}

	c.HTML(http.StatusOK, "tweet", data)

}

// DayHandler ...
func DayHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	forUser, err := store.GetAuthedUser(c.Request.Context(), username)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	isoDate := c.Param("day")
	if isoDate == "" {
		viewErrorHandler(c, http.StatusBadRequest, errors.New("Day required (param: day)"))
		return
	}

	day, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		viewErrorHandler(c, http.StatusBadRequest, fmt.Errorf("Invalid day parameter format (expected YYYY-MM-DD, got: %s)", isoDate))
		return
	}

	dayState, err := store.GetDailyFollowerState(c.Request.Context(), username, day)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	followers, err := ToUserEvent(forUser, dayState.NewFollowers, isoDate, store.FollowedEventType)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	unfollowers, err := ToUserEvent(forUser, dayState.Unfollowers, isoDate, store.UnfollowedEventType)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err)
		return
	}

	data := gin.H{
		"username":    username,
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

func viewErrorHandler(c *gin.Context, code int, err error) {
	logger.Printf("Error: %v", err)
	c.HTML(code, "error", gin.H{
		"error":  err.Error,
		"status": "Internal Server Error",
	})
	c.Abort()
}
