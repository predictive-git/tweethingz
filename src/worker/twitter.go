package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
)

var (
	logger         = log.New(os.Stdout, "worker: ", 0)
	consumerKey    = env.MustGetEnvVar("TW_KEY", "")
	consumerSecret = env.MustGetEnvVar("TW_SECRET", "")
)

func getClient(byUser *store.AuthedUser) (client *twitter.Client, err error) {

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(byUser.AccessTokenKey, byUser.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient), nil
}

// GetTwitterUserDetails retreaves details about the user
func GetTwitterUserDetails(byUser *store.AuthedUser) (user *store.SimpleUser, err error) {
	logger.Printf("User: %s", byUser.Username)
	users, err := getUsersByParams(byUser, &twitter.UserLookupParams{
		ScreenName:      []string{byUser.Username},
		IncludeEntities: twitter.Bool(true),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Error quering Twitter for user: %s", byUser.Username)
	}
	if users == nil {
		return nil, fmt.Errorf("Expected 1 user, found 0")
	}
	if len(users) != 1 {
		return nil, fmt.Errorf("Expected 1 user, found ")
	}
	return users[0], nil
}

// GetTwitterUserDetailsFromIDs retreaves details about the user
func GetTwitterUserDetailsFromIDs(byUser *store.AuthedUser, ids []int64) (users []*store.SimpleUser, err error) {
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

	users = make([]*store.SimpleUser, 0)
	items, resp, err := client.Users.Lookup(listParam)
	if err != nil {
		// TODO: find cleaner way of parsing error status code (17) from API error
		if resp.StatusCode == 404 && strings.Contains(err.Error(), "No user matches") {
			return users, nil
		}
		return nil, errors.Wrapf(err, "Error paging followers (%s): %v", resp.Status, err)
	}

	for _, u := range items {
		usr := toSimpleUser(&u)
		users = append(users, usr)
	}

	return
}

func convertTwitterTime(v string) time.Time {
	t, err := time.Parse(time.RubyDate, v)
	if err != nil {
		t = time.Now()
	}
	return t.UTC()
}

func toSimpleUser(u *twitter.User) *store.SimpleUser {
	return &store.SimpleUser{
		Username:       store.NormalizeString(u.ScreenName),
		Name:           u.Name,
		Description:    u.Description,
		ProfileImage:   u.ProfileImageURLHttps,
		CreatedAt:      convertTwitterTime(u.CreatedAt),
		Lang:           u.Lang,
		Location:       u.Location,
		Timezone:       u.Timezone,
		PostCount:      u.StatusesCount,
		FaveCount:      u.FavouritesCount,
		FollowingCount: u.FriendsCount,
		FollowerCount:  u.FollowersCount,
		ListedCount:    u.ListedCount,
		UpdatedAt:      time.Now().UTC(),
	}
}

// GetTwitterFollowerIDs returns all follower IDs for authed user
func GetTwitterFollowerIDs(byUser *store.AuthedUser) (ids []int64, err error) {

	client, err := getClient(byUser)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing client")
	}

	listParam := &twitter.FollowerIDParams{
		ScreenName: byUser.Username,
		Count:      5000, // max per page
	}

	ids = make([]int64, 0)
	for {
		page, resp, err := client.Followers.IDs(listParam)
		if err != nil {
			return nil, errors.Wrapf(err, "Error paging follower IDs (%s): %v", resp.Status, err)
		}

		// debug
		logger.Printf("   Page size:%d, Next:%d", len(page.IDs), page.NextCursor)

		ids = append(ids, page.IDs...)

		// has more IDs?
		if page.NextCursor < 1 {
			break
		}

		// reset cursor
		listParam.Cursor = page.NextCursor
	}

	return
}

func isInIntRange(v, min, max int) bool {

	if min == 0 && max == 0 {
		return true
	}

	if min > 0 && v < min {
		return false
	}

	if max > 0 && v > max {
		return false
	}

	return true

}

func isInFollowerRange(following, followers int, min, max float32) bool {

	if min == 0 && max == 0 {
		return true
	}

	v := float32(0)

	if following > 0 {
		v = float32(followers / following)
	}

	if min > 0 && v < min {
		return false
	}

	if max > 0 && v > max {
		return false
	}

	return true

}

// GetSearchResults returns all
func GetSearchResults(ctx context.Context, u *store.AuthedUser, c *store.SearchCriteria) (list []*store.SimpleTweet, err error) {

	tc, err := getClient(u)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing twitter client")
	}

	resultType := "popular"
	if c.Latest {
		resultType = "recent"
	}

	qp := &twitter.SearchTweetParams{
		Query:           c.Value,
		Lang:            c.Lang,
		Count:           100,
		SinceID:         c.SinceID,
		IncludeEntities: twitter.Bool(true),
		ResultType:      resultType,
		TweetMode:       "extended",
	}

	c.ExecutedOn = time.Now().UTC()
	list = make([]*store.SimpleTweet, 0)

	for {

		logger.Printf("Searching since ID: %d", c.SinceID)
		search, resp, err := tc.Search.Tweets(qp)

		if err != nil {
			return nil, errors.Wrapf(err, "Error executing search %+v - %v", qp, resp.Status)
		}

		// page has no data, search has no results or previous page was exactly the size
		if search == nil || search.Statuses == nil || len(search.Statuses) == 0 {
			return list, nil
		}

		logger.Printf("Page processing (List:%d, Page:%d)", len(list), len(search.Statuses))
		for _, t := range search.Statuses {

			// tweets come in newest first order so just make sure we capture the highest number
			// and start from there the next time
			if t.ID >= c.SinceID {
				c.SinceID = t.ID
				qp.SinceID = t.ID
			}

			if shouldFilterOut(&t, c) {
				continue
			}

			item := &store.SimpleTweet{
				ID:            t.IDStr,
				CriteriaID:    c.ID,
				CreatedAt:     convertTwitterTime(t.CreatedAt),
				FavoriteCount: t.FavoriteCount,
				ReplyCount:    t.ReplyCount,
				RetweetCount:  t.RetweetCount,
				Text:          t.FullText,
				IsRT:          t.RetweetedStatus != nil,
				Author:        toSimpleUser(t.User),
			}

			list = append(list, item)

		}

		logger.Printf("Page size (List:%d, Page:%d)", len(list), len(search.Statuses))

		// page has less than the max == last page
		if len(search.Statuses) < qp.Count {
			return list, nil
		}

	}

}

func shouldFilterOut(t *twitter.Tweet, c *store.SearchCriteria) bool {

	isRT := t.RetweetedStatus != nil

	// qualify the twee based on filter
	if c == nil {
		return false
	}

	// link
	if c.HasLink && (t.Entities == nil || t.Entities.Urls == nil || len(t.Entities.Urls) == 0) {
		return true
	}

	// RT
	if c.IncludeRT == false && isRT {
		return true
	}

	// Post Count
	if !isInIntRange(t.User.StatusesCount, c.PostCountMin, c.PostCountMax) {
		return true
	}

	// Fave Count
	if !isInIntRange(t.User.FavouritesCount, c.FaveCountMin, c.FaveCountMax) {
		return true
	}

	// Following Count
	if !isInIntRange(t.User.FriendsCount, c.FollowingCountMin, c.FollowingCountMax) {
		return true
	}

	// Followers Count
	if !isInIntRange(t.User.FollowersCount, c.FollowerCountMin, c.FollowerCountMax) {
		return true
	}

	// Follower Count
	if !isInFollowerRange(t.User.FriendsCount, t.User.FollowersCount, c.FollowerRatioMin, c.FollowerRatioMax) {
		return true
	}

	return false

}
