package handler

import (
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"

	"net/http"
)

// WorkerHandler ...
func WorkerHandler(c *gin.Context) {

	user := c.Param("username")

	logger.Printf("Starting service for: %s...", user)

	r := worker.Run(user)

	c.JSON(http.StatusOK, r)

}
