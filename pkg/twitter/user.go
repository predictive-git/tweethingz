package twitter

import (
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/mchmarny/tweethingz/pkg/config"
	"github.com/pkg/errors"
)

// GetUsers retreaves details about the user
func GetUsers(ids []int64) (users []*SimpleUser, err error) {

	cfg, err := config.GetTwitterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Twitter config")
	}

	logger.Printf("IDs: %d\n", len(ids))

	listParam := &twitter.UserLookupParams{
		UserID:          ids,
		IncludeEntities: twitter.Bool(true),
	}

	list := []*SimpleUser{}

	items, resp, err := getClient(cfg).Users.Lookup(listParam)
	if err != nil {
		return nil, errors.Wrapf(err, "Error paging followers (%s): %v", resp.Status, err)
	}

	// debug
	logger.Printf("Users: %d\n", len(items))

	// parse page users
	for _, u := range items {

		ca, err := time.Parse(time.RubyDate, u.CreatedAt)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing created timestamp: %s", u.CreatedAt)
		}

		usr := &SimpleUser{
			ID:             u.IDStr,
			Username:       u.ScreenName,
			Name:           u.Name,
			Description:    u.Description,
			ProfileImage:   u.ProfileImageURLHttps,
			CreatedAt:      ca,
			Lang:           u.Lang,
			Location:       u.Location,
			Timezone:       u.Timezone,
			PostCount:      u.StatusesCount,
			FaveCount:      u.FavouritesCount,
			FollowingCount: u.FriendsCount,
			FollowerCount:  u.FollowersCount,
		}

		list = append(list, usr)
	} // for users loop

	return list, nil
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
}
