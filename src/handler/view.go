package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
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

func errorHandler(c *gin.Context, err error, code int) {

	logger.Printf("Error: %v", err)
	c.HTML(code, "error", errResult)

}

// DashboardHandler ...
func DashboardHandler(c *gin.Context) {
	username := getAuthedUsername(c)
	c.HTML(http.StatusOK, "view", gin.H{
		"twitter_username": username,
		"version":          version,
		"refresh":          c.Query("refresh"),
	})
}

// SearchListHandler ...
func SearchListHandler(c *gin.Context) {
	username := getAuthedUsername(c)
	list, err := store.GetSearchCriteria(c.Request.Context(), username)
	if err != nil {
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "search", gin.H{
		"twitter_username": username,
		"version":          version,
		"list":             list,
	})
}

// SearchDetailHandler ...
func SearchDetailHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	id := c.Param("cid")
	if id == "" {
		errorHandler(c, errors.New("Search ID required"), http.StatusInternalServerError)
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
			errorHandler(c, err, http.StatusInternalServerError)
			return
		}
	}

	logger.Printf("Lang: %s", detail.Lang)

	c.HTML(http.StatusOK, "search", gin.H{
		"twitter_username": username,
		"version":          version,
		"detail":           detail,
	})

}

const tweetPageSize = 10

// TweetHandler ...
func TweetHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	cid := c.Param("cid")
	if cid == "" {
		errorHandler(c, errors.New("Search query ID required (param: cid)"), http.StatusBadRequest)
		return
	}

	criteria, err := store.GetSearchCriterion(c.Request.Context(), cid)
	if err != nil {
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	sinceKey := c.Query("key")
	if sinceKey == "" {
		sinceKey = store.ToSearchResultPagingKey(criteria.ID, time.Now(), "")
	}

	results, err := store.GetSavedSearchResults(c.Request.Context(), sinceKey, tweetPageSize)
	if err != nil {
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	data := gin.H{
		"twitter_username": username,
		"version":          version,
		"criteria":         criteria,
		"results":          results,
	}

	if len(results) == tweetPageSize {
		data["next_key"] = results[len(results)-1].Key
	}

	c.HTML(http.StatusOK, "tweet", data)

}

// DayHandler ...
func DayHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	day := c.Param("day")
	if day == "" {
		errorHandler(c, errors.New("Day required (param: day)"), http.StatusBadRequest)
		return
	}

	list, err := store.GetUserDailyEvents(c.Request.Context(), username, day)
	if err != nil {
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	followers := make([]*store.SimpleUserEvent, 0)
	unfollowers := make([]*store.SimpleUserEvent, 0)

	for _, item := range list {
		if item.EventType == store.FollowedEventType {
			followers = append(followers, item)
		} else if item.EventType == store.UnfollowedEventType {
			unfollowers = append(unfollowers, item)
		} else {
			logger.Printf("invalid event type: %s", item.EventType)
			errorHandler(c, err, http.StatusInternalServerError)
			return
		}
	}

	logger.Printf("List:%d (f:%d, u:%d)", len(list), len(followers), len(unfollowers))

	data := gin.H{
		"username":    username,
		"version":     version,
		"date":        day,
		"followers":   followers,
		"unfollowers": unfollowers,
	}

	c.HTML(http.StatusOK, "day", data)

}
