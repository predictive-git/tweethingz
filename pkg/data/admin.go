package data

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

const (
	authedEventType = "authed"
)

var (
	// ErrUserNotFound represents static error when user is not found
	ErrUserNotFound = errors.New("User not found")
)

// LookupUIUser check if the authed email is in UI users and creates UI event
func LookupUIUser(email, context string) (twitterUsername string, err error) {

	if email == "" || context == "" {
		return "", errors.New("Null email or context parameter")
	}

	if err := initDB(); err != nil {
		return "", err
	}

	// select
	logger.Printf("Quering for: %s", email)
	row := db.QueryRow("SELECT twitter_username FROM ui_users WHERE email = ?;", email)

	var twUsername string
	err = row.Scan(&twUsername)
	if err != nil && err != sql.ErrNoRows {
		return "", errors.Wrap(err, "Error parsing select results")
	}

	if twUsername == "" {
		return "", ErrUserNotFound
	}

	// insert event
	_, err = db.Exec(`INSERT INTO ui_events
		(email, event_at, event_type, description) VALUES (?, ?, ?, ?)`,
		email, time.Now().Format("2006-01-02 15:04:05"), authedEventType, context)
	if err != nil {
		return "", errors.Wrap(err, "Error executing event insert")
	}

	return twUsername, nil

}
