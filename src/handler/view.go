package handler

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/gcputil/env"
)

var (
	version   = env.MustGetEnvVar("RELEASE", "v0.0.1-manual")
	templates *template.Template
)

// DefaultHandler ...
func DefaultHandler(c *gin.Context) {

	uid, _ := c.Cookie(userIDCookieName)
	if uid != "" {
		logger.Printf("user already authenticated -> view")
		c.Redirect(http.StatusSeeOther, "/view")
		return
	}

	c.HTML(http.StatusOK, "index", gin.H{
		"version": version,
	})

}

func errorHandler(c *gin.Context, err error, code int) {

	logger.Printf("Error: %v", err)
	c.HTML(code, "error", gin.H{
		"error":       "Server error, details captured in service logs",
		"status_code": code,
		"status":      http.StatusText(code),
	})

}

// ViewHandler ...
func ViewHandler(c *gin.Context) {

	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "view", gin.H{
		"twitter_username": username,
		"version":          version,
		"refresh":          c.Query("refresh"),
	})

}
