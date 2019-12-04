package handler

import (
	"net/http"

	"github.com/mchmarny/tweethingz/src/data"

	"github.com/gin-gonic/gin"
)

// DataHandler ...
func DataHandler(c *gin.Context) {

	uid, _ := c.Cookie(userIDCookieName)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"status":  "Unauthorized",
		})
		return
	}

	result, err := data.GetSummaryForUser(uid)
	if err != nil {
		logger.Printf("Error while quering data service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error while quering data service",
			"status":  "Error",
		})
		return
	}

	c.JSON(http.StatusOK, result)

}
