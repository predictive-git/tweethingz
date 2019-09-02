package twitter

import (
	"time"
)

// SimpleTwitterUser represents simplified Twitter user
type SimpleTwitterUser struct {

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
	IsFollower     bool `json:"is_follower"`
	PostCount      int  `json:"post_count"`
	FaveCount      int  `json:"fave_count"`
	FollowingCount int  `json:"following_count"`
	FollowerCount  int  `json:"followers_count"`
}
