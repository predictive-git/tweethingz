package main

import (
	"encoding/json"
	"net/http"

	"github.com/mchmarny/tweethingz/pkg/data"
)

func datadHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	uid := getCurrentUserIDFromCookie(r)
	if uid == "" {
		logger.Println("User not authenticated")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	result, err := data.GetSummaryForUser(uid)
	if err != nil {
		logger.Printf("Error while quering data service: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		logger.Printf("Error while encoding response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}
