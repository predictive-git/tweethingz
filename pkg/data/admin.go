package data

import (
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound represents static error when user is not found
	ErrUserNotFound = errors.New("User not found")
)

// SaveAuthSession persists oauth session config
func SaveAuthSession(id, content string) error {

	if id == "" || content == "" {
		return errors.New("Nil id or content argument")
	}

	if err := initDB(); err != nil {
		return err
	}

	_, err := db.Exec(`INSERT INTO auth_sessions (session_id, auth_config,
		created_on) VALUES (?, ?, ?)`, id, content, time.Now())

	if err != nil {
		return errors.Wrap(err, "Error executing save auth session")
	}

	return nil

}

// GetAuthSession retreaves previous saved session config
func GetAuthSession(id string, maxAge int) (content string, err error) {

	if id == "" {
		return "", errors.New("Null id parameter")
	}

	if err := initDB(); err != nil {
		return "", err
	}

	row := db.QueryRow(`SELECT auth_config
						FROM auth_sessions
						WHERE session_id = ?
						AND TIMESTAMPDIFF(MINUTE,NOW(),created_on) < ?`, id, maxAge)

	var c string
	err = row.Scan(&c)
	if err != nil {
		return "", errors.Wrap(err, "Error parsing authed user")
	}

	return c, nil

}

// DeleteAuthSession deletes session once it has been used
func DeleteAuthSession(id string) error {

	if id == "" {
		return errors.New("Null id parameter")
	}

	if err := initDB(); err != nil {
		return err
	}

	_, e := db.Exec(`DELETE FROM auth_sessions WHERE session_id = ?`, id)
	if e != nil {
		return errors.Wrap(e, "Error deleteting session")
	}

	return nil

}

// SaveAuthUser saves multiple users
func SaveAuthUser(user *AuthedUser) error {

	if user == nil {
		return errors.New("Nil user argument")
	}

	if err := initDB(); err != nil {
		return err
	}

	_, err := db.Exec(`INSERT INTO authed_users (
			username, user_id, access_token_key, access_token_secret, updated_on
			) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE
			access_token_key = ?, access_token_secret = ?, updated_on = ?`,
		user.Username, user.UserID, user.AccessTokenKey, user.AccessTokenSecret,
		user.UpdatedAt, user.AccessTokenKey, user.AccessTokenSecret, user.UpdatedAt)

	if err != nil {
		return errors.Wrapf(err, "Error executing save auth user for: %+v", user)
	}

	logger.Printf("Saved authed users: %s", user.Username)

	return nil

}

// GetAuthedUser check if the authed email is in UI users and creates UI event
func GetAuthedUser(email string) (user *AuthedUser, err error) {

	if email == "" {
		return nil, errors.New("Null email or context parameter")
	}

	if err := initDB(); err != nil {
		return nil, err
	}

	row := db.QueryRow(`SELECT username, user_id, access_token_key,
						access_token_secret, updated_on
						FROM authed_users
						WHERE username = ?`, email)

	u := &AuthedUser{}
	err = row.Scan(&u.Username, &u.UserID, &u.AccessTokenKey, &u.AccessTokenSecret, &u.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing authed user")
	}

	return u, nil

}

// GetAuthedUsers gets all authed users
func GetAuthedUsers() (users []*AuthedUser, err error) {

	if err := initDB(); err != nil {
		return nil, err
	}

	rows, e := db.Query(`SELECT username, user_id, access_token_key,
						access_token_secret, updated_on
						FROM authed_users
						ORDER BY updated_on DESC`)
	if e != nil {
		return nil, errors.Wrap(e, "Error quering all authed users")
	}

	list := []*AuthedUser{}
	for rows.Next() {
		u := &AuthedUser{}
		e = rows.Scan(&u.Username, &u.UserID, &u.AccessTokenKey, &u.AccessTokenSecret, &u.UpdatedAt)
		if e != nil {
			return nil, errors.Wrap(e, "Error parsing authed users")
		}
		list = append(list, u)
	}

	return list, nil

}
