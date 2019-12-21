package worker

import (
	"context"
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
	return t
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
	}
}

func getTwitterFollowerIDs(byUser *store.AuthedUser) (ids []int64, err error) {

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

func isInIntRange(v int, r *store.IntRange) bool {

	if r == nil {
		return true
	}

	if r.Min == 0 && r.Max == 0 {
		return true
	}

	if r.Min > 0 && v < r.Min {
		return false
	}

	if r.Max > 0 && v > r.Max {
		return false
	}

	return true

}

func isInFollowerRange(following, followers int, r *store.FloatRange) bool {

	if r == nil {
		return true
	}

	if r.Min == 0 && r.Max == 0 {
		return true
	}

	v := float64(followers / following)

	if r.Min > 0 && v < r.Min {
		return false
	}

	if r.Max > 0 && v > r.Max {
		return false
	}

	return true

}

func getSearchResults(ctx context.Context, u *store.AuthedUser, c *store.SearchCriteria) (list []*store.SimpleTweet, err error) {

	tc, err := getClient(u)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing twitter client")
	}

	qp := &twitter.SearchTweetParams{
		Query:           c.Query.Value,
		Lang:            c.Query.Lang,
		Count:           100,
		SinceID:         c.Query.SinceID,
		IncludeEntities: twitter.Bool(true),
	}

	list = make([]*store.SimpleTweet, 0)

	for {

		logger.Printf("Searching since ID: %d", c.Query.SinceID)
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

			if shouldFilterOut(&t, c.Filter) {
				continue
			}

			item := &store.SimpleTweet{
				ID:            t.IDStr,
				CriteriaID:    c.ID,
				CreatedAt:     convertTwitterTime(t.CreatedAt),
				FavoriteCount: t.FavoriteCount,
				ReplyCount:    t.ReplyCount,
				RetweetCount:  t.RetweetCount,
				Text:          t.Text,
				IsRT:          t.RetweetedStatus != nil,
				Author:        toSimpleUser(t.User),
			}

			list = append(list, item)

			// tweets come in newest first order so just make sure we capture the highest number
			// and start from there the next time
			if t.ID > c.Query.SinceID {
				c.Query.SinceID = t.ID
				qp.SinceID = t.ID
			}

		}

		// page has less than the max == last page
		if len(search.Statuses) < qp.Count {
			logger.Printf("Page size (List:%d, Page:%d)", len(list), len(search.Statuses))
			return list, nil
		}

	}

}

func shouldFilterOut(t *twitter.Tweet, c *store.SimpleFilter) bool {

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

	// Author
	if c.Author != nil {

		// Post Count
		if !isInIntRange(t.User.StatusesCount, c.Author.PostCount) {
			return true
		}

		// Fave Count
		if !isInIntRange(t.User.FavouritesCount, c.Author.FaveCount) {
			return true
		}

		// Following Count
		if !isInIntRange(t.User.FriendsCount, c.Author.FollowingCount) {
			return true
		}

		// Followers Count
		if !isInIntRange(t.User.FollowersCount, c.Author.FollowerCount) {
			return true
		}

		// Follower Count
		if !isInFollowerRange(t.User.FriendsCount, t.User.FollowersCount, c.Author.FollowerRatio) {
			return true
		}

	}

	return false

}
