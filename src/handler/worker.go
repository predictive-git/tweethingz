package handler

import (
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"

	"net/http"
)

var (
	expectedToken = env.MustGetEnvVar("TOKEN", "")
)

// WorkerHandler ...
func WorkerHandler(c *gin.Context) {

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

	logger.Printf("Starting worker for: %s...", user)
	result := worker.Run(user)
	logger.Printf("Result: %+v", result)

	c.JSON(http.StatusOK, result)

}
