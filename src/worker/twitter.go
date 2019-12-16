package worker

import (
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
)

// GetUserDetails retreaves details about the user
func GetUserDetails(byUser *store.AuthedUser) (users []*store.SimpleUser, err error) {
	logger.Printf("User: %s", byUser.Username)
	return getUsersByParams(byUser, &twitter.UserLookupParams{
		ScreenName:      []string{byUser.Username},
		IncludeEntities: twitter.Bool(true),
	})
}

// GetUsersFromIDs retreaves details about the user
func GetUsersFromIDs(byUser *store.AuthedUser, ids []int64) (users []*store.SimpleUser, err error) {
	logger.Printf("IDs: %d", len(ids))
	return getUsersByParams(byUser, &twitter.UserLookupParams{
		UserID:          ids,
		IncludeEntities: twitter.Bool(true),
	})
}

func getUsersByParams(byUser *store.AuthedUser, listParam *twitter.UserLookupParams) (users []*store.SimpleUser, err error) {

	client, err := getClient(byUser)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing client")
	}

	list := []*store.SimpleUser{}
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

		usr := &store.SimpleUser{
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

func getFollowerIDs(byUser *store.AuthedUser) (ids []int64, err error) {

	client, err := getClient(byUser)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing client")
	}

	listParam := &twitter.FollowerIDParams{
		ScreenName: byUser.Username,
		Count:      5000, // max per page
	}

	list := []int64{}

	for {
		page, resp, err := client.Followers.IDs(listParam)
		if err != nil {
			return nil, errors.Wrapf(err, "Error paging follower IDs (%s): %v", resp.Status, err)
		}

		// debug
		logger.Printf("Page size:%d, Next:%d\n", len(page.IDs), page.NextCursor)

		list = append(list, page.IDs...)

		// has more IDs?
		if page.NextCursor < 1 {
			break
		}

		// reset cursor
		listParam.Cursor = page.NextCursor
	}

	return list, nil
}
