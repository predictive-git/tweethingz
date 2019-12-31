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
	logger = log.New(os.Stdout, "", 0)
	port   = env.MustGetEnvVar("PORT", "8080")
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

	// auth (authing itself)
	auth := r.Group("/auth")
	{
		auth.GET("/login", handler.AuthLoginHandler)
		auth.GET("/callback", handler.AuthCallbackHandler)
		auth.GET("/logout", handler.LogOutHandler)
	}

	// authed routes
	view := r.Group("/view")
	view.Use(handler.AuthRequired(false))
	{
		view.GET("/board", handler.DashboardHandler)
		view.GET("/search", handler.SearchListHandler)
		view.GET("/search/:cid", handler.SearchDetailHandler)
		view.GET("/tweet/:cid", handler.TweetHandler)
		view.GET("/day/:day", handler.DayHandler)
	}

	data := r.Group("/data")
	data.Use(handler.AuthRequired(true))
	{
		data.GET("/view", handler.DashboardDataHandler)
		data.DELETE("/search/:id", handler.SearchDeleteHandler)
		data.POST("/search", handler.SearchDataSubmitHandler)
	}

	// api (token validation)
	api := r.Group("/api")
	api.Use(handler.APITokenRequired())
	{
		v1 := api.Group("/v1")
		{
			// exec user follower update
			v1.POST("/follower", handler.ExecuteFollowerUpdateHandler)
		}
	}

	// port
	hostPort := net.JoinHostPort("0.0.0.0", port)
	logger.Printf("Server starting: %s \n", hostPort)
	if err := r.Run(hostPort); err != nil {
		logger.Fatal(err)
	}

}
