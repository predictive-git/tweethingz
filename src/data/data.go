package data

import (
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/mchmarny/gcputil/env"
	"github.com/pkg/errors"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

const (
	isoDateFormat = "2006-01-02"

	// EventNewFollower event type
	EventNewFollower = "followed"
	// EventUnFollowing event type
	EventUnFollowing = "unfollowed"
)

var (
	logger = log.New(os.Stdout, "data: ", 0)
	dsn    = env.MustGetEnvVar("DSN", "")

	db *sql.DB
)

func initDB() error {

	if db != nil {
		return nil
	}

	d, err := sql.Open("mysql", dsn)
	if err != nil {
		return errors.Wrap(err, "Error connecting to DB")
	}

	if err := d.Ping(); err != nil {
		d.Close()
		return errors.Wrap(err, "Error pinging DB")
	}

	db = d

	return nil

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
