package handler

import (
	"net/http"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"
)

// DataHandler ...
func DataHandler(c *gin.Context) {

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
