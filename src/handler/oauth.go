package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/kurrik/oauth1a"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/data"

	"github.com/gin-gonic/gin"
)

const (
	googleOAuthURL   = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	stateCookieName  = "tweethingz"
	userIDCookieName = "user_id"
	authIDCookieName = "auth_id"
)

var (
	logger                   = log.New(os.Stdout, "handler: ", 0)
	authedUserCookieDuration = 30 * 24 * 60
	consumerKey              = env.MustGetEnvVar("TW_KEY", "")
	consumerSecret           = env.MustGetEnvVar("TW_SECRET", "")
)

func getOAuthService(r *http.Request) *oauth1a.Service {

	// HTTPS or HTTP
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	baseURL := fmt.Sprintf("%s://%s", proto, r.Host)
	logger.Printf("External URL: %s", baseURL)

	return &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authorize",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			CallbackURL:    fmt.Sprintf("%s/auth/callback", baseURL),
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}

}

// AuthLoginHandler ...
func AuthLoginHandler(c *gin.Context) {

	uid, _ := c.Cookie(userIDCookieName)
	if uid != "" {
		c.Redirect(http.StatusSeeOther, "view")
		return
	}

	service := getOAuthService(c.Request)
	httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}
	if err := userConfig.GetRequestToken(service, httpClient); err != nil {
		err := errors.Wrap(err, "Could not get request token")
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	url, err := userConfig.GetAuthorizeURL(service)
	if err != nil {
		err := errors.Wrap(err, "Could not get authorization URL")
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	logger.Printf("Redirecting user to %s", url)

	sessionID := getNewSessionID()
	log.Printf("Starting session %s", sessionID)

	data.SaveAuthSession(sessionID, userConfigToString(userConfig))

	c.SetCookie(authIDCookieName, sessionID, 60, "/", c.Request.Host, false, true)

	c.Redirect(http.StatusFound, url)

}

// AuthCallbackHandler ...
func AuthCallbackHandler(c *gin.Context) {

	logger.Println("Auth callback...")
	sessionID, err := c.Cookie(authIDCookieName)
	if err != nil {
		err := errors.Wrap(err, "Error, callback with no session id")
		errorHandler(c, err, http.StatusUnauthorized)
		return
	}

	// TODO: make the session age decision here
	content, err := data.GetAuthSession(sessionID, 20)
	if err != nil || content == "" {
		err := errors.Wrapf(err, "Unable to find auth config for this sessions ID: %s", sessionID)
		errorHandler(c, err, http.StatusUnauthorized)
		return
	}

	userConfig, err := userConfigFromString(content)
	if err != nil {
		err := errors.New("Error decoding user config in sessions storage")
		errorHandler(c, err, http.StatusUnauthorized)
		return
	}

	service := getOAuthService(c.Request)

	token, verifier, err := userConfig.ParseAuthorize(c.Request, service)
	if err != nil {
		err := errors.Wrap(err, "Error, Could not parse authorization")
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	httpClient := new(http.Client)
	if err = userConfig.GetAccessToken(token, verifier, service, httpClient); err != nil {
		err := errors.Wrap(err, "Error getting access token")
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	logger.Printf("Ending session %s", sessionID)
	if err := data.DeleteAuthSession(sessionID); err != nil {
		err := errors.Wrap(err, "Error deleting session")
		errorHandler(c, err, http.StatusInternalServerError)
		return
	}

	c.SetCookie(authIDCookieName, "", 0, "/", c.Request.Host, false, true)

	authedUser := &data.AuthedUser{
		Username:          userConfig.AccessValues.Get("screen_name"),
		AccessTokenKey:    userConfig.AccessTokenKey,
		AccessTokenSecret: userConfig.AccessTokenSecret,
		UpdatedAt:         time.Now(),
	}

	if err = data.SaveAuthUser(authedUser); err != nil {
		e := errors.Wrap(err, "Error saving authenticated user")
		errorHandler(c, e, http.StatusInternalServerError)
		return
	}

	c.SetCookie(userIDCookieName, authedUser.Username, authedUserCookieDuration,
		"/", c.Request.Host, false, true)

	c.Redirect(http.StatusSeeOther, "/view")

}

// LogOutHandler ...
func LogOutHandler(c *gin.Context) {

	c.SetCookie(userIDCookieName, "", -1, "/", c.Request.Host, false, true)
	c.Redirect(http.StatusSeeOther, "/")

}

func getNewSessionID() string {
	c := 128
	b := make([]byte, c)
	n, err := io.ReadFull(rand.Reader, b)
	if n != len(b) || err != nil {
		panic("Could not generate random number")
	}
	return base64.URLEncoding.EncodeToString(b)
}

func userConfigToString(config *oauth1a.UserConfig) string {
	b, _ := json.Marshal(config)
	return hex.EncodeToString(b)
}

func userConfigFromString(content string) (conf *oauth1a.UserConfig, err error) {
	b, e := hex.DecodeString(content)
	if e != nil {
		return nil, e
	}
	var c oauth1a.UserConfig
	if e := json.Unmarshal(b, &c); e != nil {
		return nil, e
	}

	return &c, nil
}
