package handler

import (
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"

	"net/http"
)

var (
	expectedToken = env.MustGetEnvVar("TOKEN", "")
)

// ExecuteFollowerUpdateHandler ...
func ExecuteFollowerUpdateHandler(c *gin.Context) {

	users, err := store.GetAllAuthedUsers(c.Request.Context())
	if err != nil {
		logger.Printf("error while getting authed users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error running worker",
			"status":  "Internal Error",
		})
		c.Abort()
		return
	}

	for _, user := range users {
		logger.Printf("Starting follower update for: %s...", user.Username)
		if err := worker.ExecuteFollowerUpdate(c.Request.Context(), user); err != nil {
			logger.Printf("error while updating user data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error running worker",
				"status":  "Internal Error",
			})
			c.Abort()
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Follower update executed",
		"status":  "Success",
	})

}

// APITokenRequired is a authentication midleware
func APITokenRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User not authenticated",
				"status":  "Unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
