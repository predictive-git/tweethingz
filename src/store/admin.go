package store

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

const (
	sessionCollectionName = "thingz_session"
	authCollectionName    = "thingz_auth"
	maxAuthedUsers        = 100
)

//============================================================================
// User
//============================================================================

// AuthSession represents the authenticated user session
type AuthSession struct {
	ID     string    `firestore:"id" json:"id"`
	Config string    `firestore:"config" json:"config"`
	On     time.Time `firestore:"on" json:"on"`
}

// SaveAuthSession persists authenticated user session config
func SaveAuthSession(ctx context.Context, s *AuthSession) error {

	if s == nil {
		return errors.New("Nil auh session")
	}

	if err := save(ctx, sessionCollectionName, s.ID, s); err != nil {
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
	e := getByID(ctx, sessionCollectionName, id, s)
	if e != nil {
		return nil, errors.Wrap(err, "Error getting session")
	}

	return s, nil

}

// DeleteAuthSession deletes session once it has been used
func DeleteAuthSession(ctx context.Context, id string) error {

	if id == "" {
		return errors.New("Null id parameter")
	}

	return deleteByID(ctx, sessionCollectionName, id)

}

//============================================================================
// User
//============================================================================

// AuthedUser represents authenticated user
type AuthedUser struct {
	Username          string      `firestore:"username" json:"username"`
	Profile           *SimpleUser `firestore:"profile" json:"profile"`
	AccessTokenKey    string      `firestore:"access_token_key" json:"access_token_key"`
	AccessTokenSecret string      `firestore:"access_token_secret" json:"access_token_secret"`
	UpdatedAt         time.Time   `firestore:"updated_at" json:"updated_at"`
}

// SaveAuthUser saves multiple users
func SaveAuthUser(ctx context.Context, u *AuthedUser) error {
	if u == nil {
		return errors.New("Nil user argument")
	}

	if err := save(ctx, authCollectionName, ToID(u.Username), u); err != nil {
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
	err = getByID(ctx, authCollectionName, ToID(username), user)

	return
}

// GetAllAuthedUsers retreaves all authenticated users
func GetAllAuthedUsers(ctx context.Context) (users []*AuthedUser, err error) {

	users = make([]*AuthedUser, 0)

	col, err := getCollection(ctx, authCollectionName)
	if err != nil {
		return nil, err
	}

	// all docs with the most recent updated authed user first
	docs := col.OrderBy("updated_at", firestore.Desc).Limit(maxAuthedUsers).Documents(ctx)

	for {
		d, e := docs.Next()
		if e == iterator.Done {
			break
		}
		if e != nil {
			return nil, e
		}

		item := &AuthedUser{}
		if e := d.DataTo(item); e != nil {
			return nil, e
		}
		users = append(users, item)
	}

	logger.Printf("found %d authenticated users", len(users))

	return

}
