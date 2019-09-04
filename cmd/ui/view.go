package main

import (
	"html/template"
	"net/http"
)

var (
	templates *template.Template
)

func initTemplates() {
	if templates != nil {
		return
	}
	tmpls, err := template.ParseGlob("template/*.html")
	if err != nil {
		logger.Fatalf("Error while parsing templates: %v", err)
	}
	templates = tmpls
}

func getCurrentUserID(r *http.Request) string {
	c, _ := r.Cookie(userIDCookieName)
	if c != nil {
		return c.Value
	}
	return ""
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	initTemplates()
	data := make(map[string]interface{})
	data["version"] = version

	if err := templates.ExecuteTemplate(w, "index", data); err != nil {
		logger.Printf("Error in index template: %s", err)
	}

}

func errorHandler(w http.ResponseWriter, r *http.Request, err error, code int) {

	initTemplates()
	logger.Printf("Error: %v", err)

	w.WriteHeader(code)
	if err := templates.ExecuteTemplate(w, "error", map[string]interface{}{
		"error":       "Server error, details captured in service logs",
		"status_code": code,
		"status":      http.StatusText(code),
	}); err != nil {
		logger.Printf("Error in error template: %s", err)
	}

}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	initTemplates()
	uid := getCurrentUserID(r)
	if uid == "" {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["twitter_username"] = uid
	data["version"] = version
	if err := templates.ExecuteTemplate(w, "view", data); err != nil {
		logger.Printf("Error in view template: %s", err)
	}

}
