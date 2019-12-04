package main

import (
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/handler"
)

var (
	logger  = log.New(os.Stdout, "", 0)
	port    = env.MustGetEnvVar("PORT", "8080")
	version = env.MustGetEnvVar("RELEASE", "v0.0.1-manual")
)

func main() {

	gin.SetMode(gin.ReleaseMode)

	// router
	r := gin.Default()
	r.Use(gin.Recovery())

	// static
	r.LoadHTMLGlob("template/*")
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/img/favicon.ico")

	// routes
	r.GET("/", handler.DefaultHandler)
	r.GET("/view", handler.ViewHandler)
	r.GET("/data", handler.DataHandler)

	// auth
	auth := r.Group("/auth")
	{
		auth.GET("/login", handler.AuthLoginHandler)
		auth.GET("/callback", handler.AuthCallbackHandler)
		auth.GET("/logout", handler.LogOutHandler)
	}

	// api
	v1 := r.Group("/v1")
	{
		v1.POST("/worker/:user", handler.WorkerHandler)
	}

	// port
	hostPort := net.JoinHostPort("0.0.0.0", port)
	logger.Printf("Server starting: %s \n", hostPort)
	if err := r.Run(hostPort); err != nil {
		logger.Fatal(err)
	}

}