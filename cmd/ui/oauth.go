package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	ev "github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/twitterd/pkg/data"
)

const (
	googleOAuthURL   = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	stateCookieName  = "twitterd"
	userIDCookieName = "uid"
)

var (
	longTimeAgo    = time.Duration(3650 * 24 * time.Hour)
	cookieDuration = time.Duration(30 * 24 * time.Hour)
	oauthConfig    *oauth2.Config
)

func getOAuthConfig(r *http.Request) *oauth2.Config {

	if oauthConfig != nil {
		return oauthConfig
	}

	// HTTPS or HTTP
	proto := r.Header.Get("x-forwarded-proto")
	if proto == "" {
		proto = "http"
	}
	if ev.MustGetEnvVar("FORCE_HTTPS", "NO") == "yes" {
		proto = "https"
	}

	baseURL := fmt.Sprintf("%s://%s", proto, r.Host)
	logger.Printf("External URL: %s", baseURL)

	// OAuth
	oauthConfig = &oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/auth/callback", baseURL),
		ClientID:     ev.MustGetEnvVar("OAUTH_CLIENT_ID", ""),
		ClientSecret: ev.MustGetEnvVar("OAUTH_CLIENT_SECRET", ""),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	return oauthConfig

}

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	uid := getCurrentUserID(r)
	if uid != "" {
		logger.Printf("User ID from previous visit: %s", uid)
	}

	u := getOAuthConfig(r).AuthCodeURL(generateStateOauthCookie(w))
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {

	oauthState, _ := r.Cookie(stateCookieName)

	// checking state of the callback
	if r.FormValue("state") != oauthState.Value {
		err := errors.New("invalid oauth state from Google")
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	// parsing callback data
	authData, err := getOAuthedUserData(r)
	if err != nil {
		logger.Printf("Error while parsing user data %v", err)
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	dataMap := make(map[string]interface{})
	json.Unmarshal(authData, &dataMap)

	email := strings.ToLower(dataMap["email"].(string))
	authContext := fmt.Sprintf("Google: %s", string(authData))
	logger.Printf("Email: %s", email)

	twUser, err := data.LookupUIUser(email, authContext)
	if err != nil {
		logger.Printf("Error while looking up user: %v", err)
		errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	// set cookie for 30 days
	cookie := http.Cookie{
		Name:    userIDCookieName,
		Path:    "/",
		Value:   twUser,
		Expires: time.Now().Add(cookieDuration),
	}
	http.SetCookie(w, &cookie)

	// redirect on success
	http.Redirect(w, r, "/view", http.StatusSeeOther)

}

func logOutHandler(w http.ResponseWriter, r *http.Request) {

	uid := getCurrentUserID(r)
	logger.Printf("User logging out: %s", uid)

	cookie := http.Cookie{
		Name:    userIDCookieName,
		Path:    "/",
		Value:   "",
		MaxAge:  -1,
		Expires: time.Now().Add(-longTimeAgo),
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/index", http.StatusSeeOther) // index
}

func generateStateOauthCookie(w http.ResponseWriter) string {

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:    stateCookieName,
		Value:   state,
		Expires: time.Now().Add(cookieDuration),
	}
	http.SetCookie(w, &cookie)

	return state
}

func getOAuthedUserData(r *http.Request) ([]byte, error) {

	// exchange code
	token, err := getOAuthConfig(r).Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		return nil, fmt.Errorf("Got wrong exchange code: %v", err)
	}

	// user info
	response, err := http.Get(googleOAuthURL + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Error getting user info: %v", err)
	}
	defer response.Body.Close()

	// parse body
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}

	return contents, nil
}
