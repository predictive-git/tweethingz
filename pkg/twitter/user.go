package twitter

import (
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/mchmarny/tweethingz/pkg/data"
	"github.com/pkg/errors"
)

// GetUserDetails retreaves details about the user
func GetUserDetails(byUser *data.AuthedUser) (users []*data.SimpleUser, err error) {
	logger.Printf("User: %s", byUser.Username)
	return getUsersByParams(byUser, &twitter.UserLookupParams{
		ScreenName:      []string{byUser.Username},
		IncludeEntities: twitter.Bool(true),
	})
}

// GetUsersFromIDs retreaves details about the user
func GetUsersFromIDs(byUser *data.AuthedUser, ids []int64) (users []*data.SimpleUser, err error) {
	logger.Printf("IDs: %d", len(ids))
	return getUsersByParams(byUser, &twitter.UserLookupParams{
		UserID:          ids,
		IncludeEntities: twitter.Bool(true),
	})
}

func getUsersByParams(byUser *data.AuthedUser, listParam *twitter.UserLookupParams) (users []*data.SimpleUser, err error) {

	client, err := getClient(byUser)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing client")
	}

	list := []*data.SimpleUser{}
	items, resp, err := client.Users.Lookup(listParam)
	if err != nil {
		return nil, errors.Wrapf(err, "Error paging followers (%s): %v", resp.Status, err)
	}

	logger.Printf("Found %d users for %s", len(items), byUser.Username)

	// parse page users
	for _, u := range items {

		ca, err := time.Parse(time.RubyDate, u.CreatedAt)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing created timestamp: %s", u.CreatedAt)
		}

		usr := &data.SimpleUser{
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
