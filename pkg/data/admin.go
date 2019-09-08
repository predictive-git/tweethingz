package data

import (
	"database/sql"

	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound represents static error when user is not found
	ErrUserNotFound = errors.New("User not found")
)

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

	row := db.QueryRow(`select username, user_id, access_token_key,
						access_token_secret, updated_on
						from authed_users
						where username = ?`, email)

	u := &AuthedUser{}
	err = row.Scan(&u.Username, &u.UserID, &u.AccessTokenKey, &u.AccessTokenSecret, &u.UpdatedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
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

	rows, e := db.Query(`select username, user_id, access_token_key,
						access_token_secret, updated_on
						from authed_users
						order by updated_on desc`)
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


