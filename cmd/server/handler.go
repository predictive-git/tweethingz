package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mchmarny/twitterd/worker"
)

func okHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func apiRequestHandler(c *gin.Context) {

	usr := c.Param("u")
	logger.Printf("User: %s", usr)
	if usr == "" {
		logger.Println("Error on nil usr parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Null Argument",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := worker.ProcessFollowers(usr)
	if err != nil {
		logger.Println("Error processing followers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Internal Error",
			"status":  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"status":  http.StatusOK,
	})

}
