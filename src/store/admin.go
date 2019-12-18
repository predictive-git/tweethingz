package store

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

const (
	sessionCollectionName = "thingz_session"
	authCollectionName    = "thingz_auth"
)

// AuthSession represents the authenticated user session
type AuthSession struct {
	ID     string    `firestore:"id" json:"id"`
	Config string    `firestore:"config" json:"config"`
	On     time.Time `firestore:"on" json:"on"`
}

// AuthedUser represents authenticated user
type AuthedUser struct {

	// User details
	Username string `firestore:"username" json:"username"`

	AccessTokenKey    string `firestore:"access_token_key" json:"access_token_key"`
	AccessTokenSecret string `firestore:"access_token_secret" json:"access_token_secret"`

	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

// SaveAuthSession persists authenticated user session config
func SaveAuthSession(ctx context.Context, s *AuthSession) error {

	if s == nil {
		return errors.New("Nil auh session")
	}

	if err := save(ctx, sessionCollectionName, toID(s.ID), s); err != nil {
		return errors.Wrap(err, "Error executing save auth session")
	}

	return nil

}

// GetAuthSession retreaves previous saved session config
func GetAuthSession(ctx context.Context, id string) (content *AuthSession, err error) {

	if id == "" {
		return nil, errors.New("Null id parameter")
	}

	s := &AuthSession{}
	e := getByID(ctx, sessionCollectionName, toID(id), s)
	if e != nil {
		return nil, errors.Wrap(err, "Error getting session")
	}

	return s, nil

}

// DeleteAuthSession deletes session once it has been used
func DeleteAuthSession(ctx context.Context, username string) error {

	if username == "" {
		return errors.New("Null id parameter")
	}

	return deleteByID(ctx, sessionCollectionName, toID(username))

}

// SaveAuthUser saves multiple users
func SaveAuthUser(ctx context.Context, u *AuthedUser) error {

	if u == nil {
		return errors.New("Nil user argument")
	}

	if err := save(ctx, authCollectionName, toID(u.Username), u); err != nil {
		return errors.Wrap(err, "Error executing save auth session")
	}

	return nil

}

// GetAuthedUser check if the authed username is in UI users and creates UI event
func GetAuthedUser(ctx context.Context, username string) (user *AuthedUser, err error) {

	if username == "" {
		return nil, errors.New("username required")
	}

	user = &AuthedUser{}
	e := getByID(ctx, authCollectionName, toID(username), user)
	if e != nil {
		return nil, e
	}

	return

}
