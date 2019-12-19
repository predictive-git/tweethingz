package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestViewRedirectSansAuthCookie(t *testing.T) {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", ViewHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)

}
