package data

import (
	"time"

	"github.com/pkg/errors"
)

const (
	authedEventType = "authed"
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
	selectStmt, err := db.Prepare("SELECT twitter_username FROM ui_users WHERE email = ?")
	if err != nil {
		return "", errors.Wrap(err, "Error preparing select statement")
	}

	row := selectStmt.QueryRow(email)
	if row != nil {
		return "", errors.Wrapf(err, "Error while selecting row for %s", email)
	}

	var twUsername string
	if err := row.Scan(&twUsername); err != nil {
		return "", errors.Wrap(err, "Error parsing session incrementing results")
	}

	if twUsername == "" {
		return "", nil
	}

	// insert event
	insertStmt, err := db.Prepare(`INSERT INTO ui_events (
			email, event_at, event_type, description) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return "", errors.Wrap(err, "Error preparing insert statement")
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(email, time.Now().Format("2006-01-02 15:04:05"),
		authedEventType, context)
	if err != nil {
		return "", errors.Wrap(err, "Error executing event insert")
	}

	return twUsername, nil

}
