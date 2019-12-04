package store

import (
	"log"
	"os"
	"time"

	"context"
	"errors"
	"fmt"
	"github.com/mchmarny/gcputil/env"
	"hash/fnv"

	"cloud.google.com/go/firestore"
	"github.com/mchmarny/gcputil/project"
)

const (
	isoDateFormat         = "2006-01-02"
	collectionDefaultName = "tweethingz"
	recordIDPrefix        = "id-"

	// EventNewFollower event type
	EventNewFollower = "followed"
	// EventUnFollowing event type
	EventUnFollowing = "unfollowed"
)

var (
	logger         = log.New(os.Stdout, "data: ", 0)
	collectionName = env.MustGetEnvVar("DB_NAME", collectionDefaultName)
	storePath      = env.MustGetEnvVar("DB_PATH", "")
	projectID      = project.GetIDOrFail()

	errNilDocRef = errors.New("firestore: nil DocumentRef")

	fsClient  *firestore.Client
	stateColl *firestore.CollectionRef
)

func initStore(ctx context.Context) error {

	c, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("Error while creating Firestore client: %v", err)
	}
	fsClient = c
	stateColl = c.Collection(collectionName)
	return nil
}

func deleteByID(ctx context.Context, id string) error {

	if id == "" {
		return errors.New("Nil id")
	}

	_, err := stateColl.Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("Error getting state: %v", err)
	}

	return nil
}

func getByID(ctx context.Context, id string, in interface{}) (out interface{}, err error) {

	if id == "" {
		return nil, errors.New("Nil id")
	}

	d, err := stateColl.Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error getting state: %v", err)
	}

	if err := d.DataTo(in); err != nil {
		return nil, fmt.Errorf("Stored data not user: %v", err)
	}

	return in, nil
}

func save(ctx context.Context, id string, in interface{}) error {

	if in == nil {
		return errors.New("Nil state")
	}

	_, err := stateColl.Doc(id).Set(ctx, in)
	if err != nil {
		return fmt.Errorf("Error on save: %v", err)
	}
	return nil
}

func toID(query string) string {
	h := fnv.New32a()
	h.Write([]byte(query))
	return fmt.Sprintf("%s%d", recordIDPrefix, h.Sum32())
}

// SimpleUserEvent wraps simple twitter user as an time event
type SimpleUserEvent struct {
	SimpleUser
	EventDate time.Time `json:"event_at"`
}

// SimpleUser represents simplified Twitter user
type SimpleUser struct {

	// ID is global identifier
	ID string `json:"id"`

	// User details
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ProfileImage string    `json:"profile_image"`
	CreatedAt    time.Time `json:"created_at"`

	// geo
	Lang     string `json:"lang"`
	Location string `json:"location"`
	Timezone string `json:"time_zone"`

	// counts
	PostCount      int `json:"post_count"`
	FaveCount      int `json:"fave_count"`
	FollowingCount int `json:"following_count"`
	FollowerCount  int `json:"followers_count"`

	// Meta
	UpdatedOn time.Time `json:"updated_on"`
}

// AuthedUser represents authenticated user
type AuthedUser struct {

	// User details
	Username string `json:"username"`

	AccessTokenKey    string `json:"access_token_key"`
	AccessTokenSecret string `json:"access_token_secret"`

	UpdatedAt time.Time `json:"updated_at"`
}
