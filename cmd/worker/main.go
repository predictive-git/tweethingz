package main

import (
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"

	ev "github.com/mchmarny/gcputil/env"
)

const (
	defaultPort      = "8080"
	portVariableName = "PORT"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// routes
	r.GET("/", okHandler)
	r.GET("/health", okHandler)

	// api
	v1 := r.Group("/v1")
	{
		v1.POST("/followers/:username", followerScheduleHandler)
		v1.POST("/backfill", backfillScheduleHandler)
	}

	// server
	port := ev.MustGetEnvVar(portVariableName, defaultPort)
	hostPost := net.JoinHostPort("0.0.0.0", port)
	logger.Printf("Server starting: %s \n", hostPost)
	if err := r.Run(hostPost); err != nil {
		logger.Fatal(err)
	}
}
