package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kurrik/oauth1a"
	"github.com/mchmarny/tweethingz/pkg/config"
	"github.com/mchmarny/tweethingz/pkg/data"
)

const (
	googleOAuthURL   = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	stateCookieName  = "tweethingz"
	userIDCookieName = "user_id"
	authIDCookieName = "auth_id"
)

var (
	authedUserCookieDuration = 30 * 24 * 60
)

func getOAuthService(r *http.Request) *oauth1a.Service {

	// HTTPS or HTTP
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	cfg, err := config.GetTwitterConfig()
	if err != nil {
		logger.Printf("Error parsing Twitter config: %v", err)
	}

	baseURL := fmt.Sprintf("%s://%s", proto, r.Host)
	logger.Printf("External URL: %s", baseURL)

	return &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authorize",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    cfg.ConsumerKey,
			ConsumerSecret: cfg.ConsumerSecret,
			CallbackURL:    fmt.Sprintf("%s/auth/callback", baseURL),
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}

}

func authLoginHandler(w http.ResponseWriter, r *http.Request) {

	uid := getCurrentUserIDFromCookie(r)
	if uid != "" {
		logger.Printf("User ID from previous visit: %s", uid)
		http.Redirect(w, r, "/view", http.StatusSeeOther)
		return
	}

	logger.Printf("Auth handled: %s", uid)

	service := getOAuthService(r)

	httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}
	if err := userConfig.GetRequestToken(service, httpClient); err != nil {
		err := errors.Wrap(err, "Could not get request token")
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	url, err := userConfig.GetAuthorizeURL(service)
	if err != nil {
		err := errors.Wrap(err, "Could not get authorization URL")
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	logger.Printf("Redirecting user to %s", url)

	sessionID := getNewSessionID()
	log.Printf("Starting session %s", sessionID)

	data.SaveAuthSession(sessionID, userConfigToString(userConfig))

	http.SetCookie(w, getSessionStartCookie(sessionID))
	http.Redirect(w, r, url, http.StatusFound)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {

	logger.Println("Auth callback...")
	sessionID, err := setSessionID(r)
	if err != nil {
		err := errors.Wrap(err, "Error, callback with no session id")
		errorHandler(w, r, err, http.StatusUnauthorized)
		return
	}

	content, err := data.GetAuthSession(sessionID, 20)
	if err != nil || content == "" {
		err := errors.Wrapf(err, "Unable to find auth config for this sessions ID: %s", sessionID)
		errorHandler(w, r, err, http.StatusUnauthorized)
		return
	}

	userConfig, err := userConfigFromString(content)
	if err != nil {
		err := errors.New("Error decoding user config in sessions storage")
		errorHandler(w, r, err, http.StatusUnauthorized)
		return
	}

	service := getOAuthService(r)

	token, verifier, err := userConfig.ParseAuthorize(r, service)
	if err != nil {
		err := errors.Wrap(err, "Error, Could not parse authorization")
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	httpClient := new(http.Client)
	if err = userConfig.GetAccessToken(token, verifier, service, httpClient); err != nil {
		err := errors.Wrap(err, "Error getting access token")
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	logger.Printf("Ending session %s", sessionID)
	if err := data.DeleteAuthSession(sessionID); err != nil {
		err := errors.Wrap(err, "Error deleting session")
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, getSessionEndCookie())

	authedUser := &data.AuthedUser{
		Username:          userConfig.AccessValues.Get("screen_name"),
		UserID:            userConfig.AccessValues.Get("user_id"),
		AccessTokenKey:    userConfig.AccessTokenKey,
		AccessTokenSecret: userConfig.AccessTokenSecret,
		UpdatedAt:         time.Now(),
	}

	if err = data.SaveAuthUser(authedUser); err != nil {
		e := errors.Wrap(err, "Error saving authenticated user")
		errorHandler(w, r, e, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, getUserAuthCookie(authedUser.Username))
	// logger.Printf("Authed User: %+v", authedUser)

	http.Redirect(w, r, "/view", http.StatusSeeOther)

}

func logOutHandler(w http.ResponseWriter, r *http.Request) {

	year := time.Duration(365*24 + time.Hour)
	cookie := http.Cookie{
		Name:    userIDCookieName,
		Path:    "/",
		Value:   "",
		MaxAge:  -1,
		Expires: time.Now().Add(-year),
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/index", http.StatusSeeOther) // index
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

func getCurrentUserIDFromCookie(r *http.Request) string {
	c, _ := r.Cookie(userIDCookieName)
	if c != nil {
		return c.Value
	}
	return ""
}

func getUserAuthCookie(id string) *http.Cookie {
	return &http.Cookie{
		Name:   userIDCookieName,
		Value:  id,
		MaxAge: authedUserCookieDuration,
		Secure: false,
		Path:   "/",
	}
}

func getSessionStartCookie(id string) *http.Cookie {
	return &http.Cookie{
		Name:   authIDCookieName,
		Value:  id,
		MaxAge: 60,
		Secure: false,
		Path:   "/",
	}
}

func getSessionEndCookie() *http.Cookie {
	return &http.Cookie{
		Name:   authIDCookieName,
		Value:  "",
		MaxAge: 0,
		Secure: false,
		Path:   "/",
	}
}

func setSessionID(r *http.Request) (id string, err error) {
	c, e := r.Cookie(authIDCookieName)
	if err != nil {
		return "", e
	}
	return c.Value, nil
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
