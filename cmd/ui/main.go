package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	ev "github.com/mchmarny/gcputil/env"
)

var (
	logger = log.New(os.Stdout, "", 0)
	port   = ev.MustGetEnvVar("PORT", "8080")
)

func main() {

	mux := http.NewServeMux()

	// Static
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))

	// Handlers
	mux.HandleFunc("/", defaultHandler)

	mux.HandleFunc("/auth/login", authLoginHandler)
	mux.HandleFunc("/auth/callback", authCallbackHandler)
	mux.HandleFunc("/auth/logout", logOutHandler)

	mux.HandleFunc("/view", viewHandler)

	mux.HandleFunc("/_health", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	})

	// Server
	hostPort := net.JoinHostPort("0.0.0.0", port)
	server := &http.Server{
		Addr:    hostPort,
		Handler: mux,
	}

	logger.Printf("Server starting: %s \n", hostPort)
	logger.Fatal(server.ListenAndServe())

}
