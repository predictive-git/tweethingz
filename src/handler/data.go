package handler

import (
	"net/http"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"
)

var (
	errResult = gin.H{
		"message": "Internal server error, see logs for details",
		"status":  "Error",
	}
)

// SearchDeleteHandler ...
func SearchDeleteHandler(c *gin.Context) {

	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"status":  "Unauthorized",
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing parameter: ID",
			"status":  "Bad Request",
		})
		return
	}

	if err := store.DeleteSearchCriterion(c.Request.Context(), id); err != nil {
		logger.Printf("Error getting criteria data: %v", err)
		c.JSON(http.StatusInternalServerError, errResult)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Search criterion deleted",
		"status":  "Success",
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

	if sc.ID == "" {
		sc.ID = store.NewID()
	}
	sc.User = username

	// logger.Printf("Search Criteria: %+v", sc)

	if err := store.SaveSearchCriteria(c.Request.Context(), sc); err != nil {
		logger.Printf("error saving search criteria: %v", err)
		c.HTML(http.StatusInternalServerError, "error", errResult)
		return
	}

	c.Redirect(http.StatusSeeOther, "/search")
	return

}

// ViewDashboardHandler ...
func ViewDashboardHandler(c *gin.Context) {

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
		c.JSON(http.StatusInternalServerError, errResult)
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

		c.JSON(http.StatusInternalServerError, errResult)
		return
	}

	c.JSON(http.StatusOK, result)

}
