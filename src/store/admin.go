package store

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound represents static error when user is not found
	ErrUserNotFound = errors.New("User not found")
)

// AuthSession represents the authenticated user session
type AuthSession struct {
	ID     string    `json:"id"`
	Config string    `json:"config"`
	On     time.Time `json:"on"`
}

// SaveAuthSession persists authenticated user session config
func SaveAuthSession(ctx context.Context, s *AuthSession) error {

	if s == nil {
		return errors.New("Nil auh session")
	}

	if err := initStore(ctx); err != nil {
		return err
	}

	if err := save(ctx, toID(s.ID), s); err != nil {
		return errors.Wrap(err, "Error executing save auth session")
	}

	return nil

}

// GetAuthSession retreaves previous saved session config
func GetAuthSession(ctx context.Context, id string) (content *AuthSession, err error) {

	if id == "" {
		return nil, errors.New("Null id parameter")
	}

	if err := initStore(ctx); err != nil {
		return nil, err
	}

	s, e := getByID(ctx, toID(id), AuthSession{})
	if e != nil {
		return nil, errors.Wrap(err, "Error getting session")
	}

	v, ok := s.(AuthSession)
	if !ok {
		return nil, errors.New("Retreaved data is not of the AuthSession type")
	}

	return &v, nil

}

// DeleteAuthSession deletes session once it has been used
func DeleteAuthSession(ctx context.Context, id string) error {

	if id == "" {
		return errors.New("Null id parameter")
	}

	if err := initStore(ctx); err != nil {
		return err
	}

	return deleteByID(ctx, toID(id))

}

// SaveAuthUser saves multiple users
func SaveAuthUser(ctx context.Context, u *AuthedUser) error {

	if u == nil {
		return errors.New("Nil user argument")
	}

	if err := initStore(ctx); err != nil {
		return err
	}

	if err := save(ctx, toID(u.Username), u); err != nil {
		return errors.Wrap(err, "Error executing save auth session")
	}

	return nil

}

// GetAuthedUser check if the authed email is in UI users and creates UI event
func GetAuthedUser(ctx context.Context, email string) (user *AuthedUser, err error) {

	if email == "" {
		return nil, errors.New("Null email or context parameter")
	}

	if err := initStore(ctx); err != nil {
		return nil, err
	}

	u, e := getByID(ctx, toID(email), AuthedUser{})
	if e != nil {
		return nil, errors.Wrap(err, "Error getting session")
	}

	v, ok := u.(AuthedUser)
	if !ok {
		return nil, errors.New("Retreaved data is not of the AuthSession type")
	}

	return &v, nil

}
