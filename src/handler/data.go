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

	username := getAuthedUsername(c)

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
		"message":  "Search criterion deleted",
		"status":   "Success",
		"username": username,
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
		c.HTML(http.StatusInternalServerError, "error", errResult)
		return
	}

	c.Redirect(http.StatusSeeOther, "/view/search")
	return

}

// ViewDashboardHandler ...
func ViewDashboardHandler(c *gin.Context) {

	username := getAuthedUsername(c)
	result, err := store.GetSummaryForUser(c.Request.Context(), username)
	if err != nil {

		logger.Printf("Error while quering data service: %v", err)

		// if anything else by no data found, err
		if !store.IsDataNotFoundError(err) {
			c.JSON(http.StatusInternalServerError, errResult)
			return
		}

		// update only when need to
		logger.Printf("No data found, updating data for: %v", username)
		if err := worker.UpdateUserData(c.Request.Context(), username); err != nil {
			logger.Printf("Error while updating after nil results: %v", err)
			c.JSON(http.StatusInternalServerError, errResult)
			return
		}

		// get data once more after update
		result, err = store.GetSummaryForUser(c.Request.Context(), username)
		if err != nil {
			logger.Printf("Error while getting summary after nil results: %v", err)
			c.JSON(http.StatusInternalServerError, errResult)
			return
		}

	}

	c.JSON(http.StatusOK, result)

}
