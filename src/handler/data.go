package handler

import (
	"net/http"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"
)

// SearchDeleteHandler ...
func SearchDeleteHandler(c *gin.Context) {

	user := getAuthedUser(c)

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing parameter: ID",
			"status":  "Bad Request",
		})
		c.Abort()
		return
	}

	if err := store.DeleteSearchCriterion(c.Request.Context(), id); err != nil {
		logger.Printf("Error getting criteria data: %v", err)
		errJSONAndAbort(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Search criterion deleted",
		"status":  "Success",
		"user":    user,
	})

}

// DashboardDataHandler ...
func DashboardDataHandler(c *gin.Context) {

	forUser := getAuthedUser(c)
	result, err := worker.GetSummaryForUser(c.Request.Context(), forUser)
	if err != nil {

		logger.Printf("Error while quering data service: %v", err)

		// if anything else by no data found, err
		if !store.IsDataNotFoundError(err) {
			errJSONAndAbort(c)
			return
		}

		// update only when need to
		logger.Printf("No data found, updating data for: %v", forUser.Username)
		if err := worker.ExecuteFollowerUpdate(c.Request.Context(), forUser); err != nil {
			logger.Printf("Error while updating after nil results: %v", err)
			errJSONAndAbort(c)
			return
		}

		// get data once more after update
		result, err = worker.GetSummaryForUser(c.Request.Context(), forUser)
		if err != nil {
			logger.Printf("Error while getting summary after nil results: %v", err)
			errJSONAndAbort(c)
			return
		}

	}

	c.JSON(http.StatusOK, result)

}

// errJSONAndAbort throws JSON error and abort prevents pending handlers from being called
func errJSONAndAbort(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Internal server error, see logs for details",
		"status":  "Error",
	})
	c.Abort()
	return
}
