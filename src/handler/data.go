package handler

import (
	"net/http"
	"time"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"
)

// SearchDataHandler ...
func SearchDataHandler(c *gin.Context) {

	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"status":  "Unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Not implemented",
		"status":  "Not implemented",
	})

}

// SearchDataSubmitHandler ...
func SearchDataSubmitHandler(c *gin.Context) {

	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"status":  "Unauthorized",
		})
		return
	}

	sc := &store.SearchCriteria{}
	if err := c.ShouldBind(&sc); err != nil {
		logger.Printf("error binding: %v", err)
	}

	sc.ID = store.NewID()
	sc.User = username
	sc.UpdatedOn = time.Now()

	// logger.Printf("Search Criteria: %+v", sc)

	if err := store.SaveSearchCriteria(c.Request.Context(), sc); err != nil {
		logger.Printf("error saving search criteria: %v", err)
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error":       "Server error, details captured in service logs",
			"status_code": http.StatusInternalServerError,
			"status":      http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/search")
	return

}

// ViewDataHandler ...
func ViewDataHandler(c *gin.Context) {

	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"status":  "Unauthorized",
		})
		return
	}

	if err := worker.UpdateUserData(c.Request.Context(), username); err != nil {
		logger.Printf("Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error refreshing data, see logs for details",
			"status":  "Internal Server Error",
		})
		return
	}

	result, err := store.GetSummaryForUser(c.Request.Context(), username)
	if err != nil {
		logger.Printf("Error while quering data service: %v", err)

		if store.IsDataNotFoundError(err) {
			c.JSON(http.StatusNoContent, gin.H{
				"message": "Your data is still being loaded. Please try again in a few min",
				"status":  "Error",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error while quering data service",
			"status":  "Error",
		})
		return
	}

	c.JSON(http.StatusOK, result)

}
