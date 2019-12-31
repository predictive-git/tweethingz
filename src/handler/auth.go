package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kurrik/oauth1a"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/mchmarny/tweethingz/src/worker"

	"github.com/gin-gonic/gin"
)

const (
	googleOAuthURL        = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	stateCookieName       = "tweethingz"
	userIDCookieName      = "user_id"
	authIDCookieName      = "auth_id"
	allowedUsersUndefined = "undefined"
)

var (
	logger                   = log.New(os.Stdout, "handler: ", 0)
	authedUserCookieDuration = 30 * 24 * 60 // sec
	maxSessionAge            = 5.0          // min
	sessionCookieAge         = 5 * 60       // maxSessionAge in secs
	consumerKey              = env.MustGetEnvVar("TW_KEY", "")
	consumerSecret           = env.MustGetEnvVar("TW_SECRET", "")
	allowedUsers             = env.MustGetEnvVar("TW_USERS", allowedUsersUndefined)
)

func getOAuthService(r *http.Request) *oauth1a.Service {

	// HTTPS or HTTP
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	baseURL := fmt.Sprintf("%s://%s", proto, r.Host)

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
		c.Redirect(http.StatusSeeOther, "/view/board")
		return
	}

	service := getOAuthService(c.Request)
	httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}
	if err := userConfig.GetRequestToken(service, httpClient); err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting request token")
		return
	}

	AuthURL, err := userConfig.GetAuthorizeURL(service)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting authorization URL")
		return
	}

	authSession := &store.AuthSession{
		ID:     store.NewID(), // already URL escaped
		Config: userConfigToString(userConfig),
		On:     time.Now().UTC(),
	}

	store.SaveAuthSession(c.Request.Context(), authSession)

	c.SetCookie(authIDCookieName, authSession.ID, sessionCookieAge, "/", c.Request.Host, false, true)

	c.Redirect(http.StatusFound, AuthURL)

}

// AuthCallbackHandler ...
func AuthCallbackHandler(c *gin.Context) {

	sessionID, err := c.Cookie(authIDCookieName)
	if err != nil {
		viewErrorHandler(c, http.StatusUnauthorized, err, "Error handling callback with no session id")
		return
	}

	authSession, err := store.GetAuthSession(c.Request.Context(), sessionID)
	if err != nil || authSession == nil {
		viewErrorHandler(c, http.StatusUnauthorized, err, fmt.Sprintf("Unable to find auth config for this sessions ID: %s", sessionID))
		return
	}

	sessionAge := time.Now().UTC().Sub(authSession.On)
	if sessionAge.Minutes() > maxSessionAge {
		viewErrorHandler(c, http.StatusUnauthorized, err, fmt.Sprintf("session %s expired. Age %v, expected %f min", sessionAge, maxSessionAge, maxSessionAge))
		return
	}

	userConfig, err := userConfigFromString(authSession.Config)
	if err != nil {
		viewErrorHandler(c, http.StatusUnauthorized, err, "Error decoding user config in sessions storage")
		return
	}

	service := getOAuthService(c.Request)

	token, verifier, err := userConfig.ParseAuthorize(c.Request, service)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Could not parse authorization")
		return
	}

	httpClient := new(http.Client)
	if err = userConfig.GetAccessToken(token, verifier, service, httpClient); err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting access token")
		return
	}

	if err := store.DeleteAuthSession(c.Request.Context(), sessionID); err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error deleting session")
		return
	}

	c.SetCookie(authIDCookieName, "", 0, "/", c.Request.Host, false, true)

	username := store.NormalizeString(userConfig.AccessValues.Get("screen_name"))

	if !isUserAllowed(username) {
		viewErrorHandler(c, http.StatusUnauthorized, nil, fmt.Sprintf("User %s not authorized to access this service", username))
		return
	}

	authedUser := &store.AuthedUser{
		Username:          username,
		AccessTokenKey:    userConfig.AccessTokenKey,
		AccessTokenSecret: userConfig.AccessTokenSecret,
		UpdatedAt:         time.Now().UTC(),
	}

	self, err := worker.GetTwitterUserDetails(authedUser)
	if err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting user twitter details")
		return
	}

	authedUser.Profile = self
	if err = store.SaveAuthUser(c.Request.Context(), authedUser); err != nil {
		viewErrorHandler(c, http.StatusInternalServerError, err, "Error saving authenticated user")
		return
	}

	c.SetCookie(userIDCookieName, authedUser.Username, authedUserCookieDuration,
		"/", c.Request.Host, false, true)

	c.Redirect(http.StatusSeeOther, "/view/board")

}

// LogOutHandler ...
func LogOutHandler(c *gin.Context) {
	c.SetCookie(userIDCookieName, "", -1, "/", c.Request.Host, false, true)
	c.Redirect(http.StatusSeeOther, "/")
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
	conf = &oauth1a.UserConfig{}
	if e := json.Unmarshal(b, conf); e != nil {
		return nil, e
	}
	return
}

// AuthRequired is a authentication midleware
func AuthRequired(isJSON bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, _ := c.Cookie(userIDCookieName)
		if username == "" {
			if isJSON {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "User not authenticated",
					"status":  "Unauthorized",
				})
			} else {
				c.Redirect(http.StatusSeeOther, "/")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func getAuthedUser(c *gin.Context) *store.AuthedUser {
	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		logger.Fatalln("This should never happen, nil auth cookie")
	}

	usr, err := store.GetAuthedUser(c.Request.Context(), username)
	if err != nil || usr == nil {
		logger.Fatalf("This should never happen, error getting authed user: %v", err)
	}

	return usr
}

func isUserAllowed(user string) bool {

	if user == "" {
		return false
	}

	// check if the user is allowed if allowedUsers defined
	if allowedUsers == "" || allowedUsers == allowedUsersUndefined {
		return true
	}

	for _, u := range strings.Split(allowedUsers, ",") {
		if strings.EqualFold(strings.TrimSpace(u), strings.TrimSpace(user)) {
			return true
		}
	}

	return false
}
