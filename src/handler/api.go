package handler

import (
	"fmt"

	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"

	"net/http"
)

var (
	expectedToken = env.MustGetEnvVar("TOKEN", "")
)

// ExecuteSearchHandler ...
func ExecuteSearchHandler(c *gin.Context) {

	token := c.Query("token")
	if token != expectedToken {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"status":  "Unauthorized",
		})
		return
	}

	user := c.Param("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username not defined",
			"status":  "Bad Request",
		})
		return
	}

	user = store.NormalizeString(user)
	logger.Printf("Starting background search worker for: %s...", user)
	if err := worker.ExecuteUserSearches(c.Request.Context(), user); err != nil {
		logger.Printf("error while executing user search: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error running worker",
			"status":  "Internal Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Refreshed %s data", user),
		"status":  "Success",
	})

}

// RefreshUserDataHandler ...
func RefreshUserDataHandler(c *gin.Context) {

	user := c.Param("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username not defined",
			"status":  "Bad Request",
		})
		return
	}

	user = store.NormalizeString(user)
	logger.Printf("Starting background worker for: %s...", user)
	if err := worker.UpdateUserData(c.Request.Context(), user); err != nil {
		logger.Printf("error while updating user data: %v", err)
		if err == store.ErrDataNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("User %s must authenticate through UI first", user),
				"status":  "Data Not Found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error running worker",
				"status":  "Internal Error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Refreshed %s data", user),
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
