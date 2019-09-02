package twitter

import (
	"log"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mchmarny/twitterd/pkg/config"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "worker - ", 0)
)

func getClient(cfg *config.TwitterConfig) *twitter.Client {
	config := oauth1.NewConfig(cfg.ConsumerKey, cfg.ConsumerSecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}

// GetFollowerIDs retreaves follower IDs for config specified user
func GetFollowerIDs() (ids []int64, err error) {

	cfg, err := config.GetTwitterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Twitter config")
	}

	listParam := &twitter.FollowerIDParams{
		ScreenName: cfg.Username,
		Count:      5000, // max per page
	}

	list := []int64{}

	for {
		page, resp, err := getClient(cfg).Followers.IDs(listParam)
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

// GetFollowers retreaves followers for config specified user
func GetFollowers() (followers []*SimpleTwitterUser, err error) {

	cfg, err := config.GetTwitterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Twitter config")
	}

	listParam := &twitter.FollowerListParams{
		ScreenName:          cfg.Username,
		SkipStatus:          twitter.Bool(true),
		IncludeUserEntities: twitter.Bool(true),
		Count:               200, // max per page
	}

	list := []*SimpleTwitterUser{}

	for {
		page, resp, err := getClient(cfg).Followers.List(listParam)
		if err != nil {
			return nil, errors.Wrapf(err, "Error paging followers (%s): %v", resp.Status, err)
		}

		// debug
		logger.Printf("%d, Next:%d\n", len(page.Users), page.NextCursor)

		// parse page users
		for _, u := range page.Users {

			ca, err := time.Parse(time.RubyDate, u.CreatedAt)
			if err != nil {
				return nil, errors.Wrapf(err, "Error parsing created timestamp: %s", u.CreatedAt)
			}

			usr := &SimpleTwitterUser{
				ID:             u.IDStr,
				Username:       u.ScreenName,
				Name:           u.Name,
				Description:    u.Description,
				ProfileImage:   u.ProfileImageURLHttps,
				CreatedAt:      ca,
				Lang:           u.Lang,
				Location:       u.Location,
				Timezone:       u.Timezone,
				IsFollower:     u.Following,
				PostCount:      u.StatusesCount,
				FaveCount:      u.FavouritesCount,
				FollowingCount: u.FriendsCount,
				FollowerCount:  u.FollowersCount,
			}

			list = append(list, usr)

		}

		// check if last page
		if page.NextCursor < 1 {
			break
		}

		// reset cursor
		listParam.Cursor = page.NextCursor
	}

	return list, nil
}

// Search searches twitter for specified query results
func Search(query string) error {

	cfg, err := config.GetTwitterConfig()
	if err != nil {
		return errors.Wrap(err, "Error getting Twitter config")
	}

	logger.Printf("Starting search for %s", query)
	list, resp, err := getClient(cfg).Search.Tweets(&twitter.SearchTweetParams{
		Query:           query,
		Count:           100,
		SinceID:         0,
		IncludeEntities: twitter.Bool(true),
	})

	if err != nil {
		return errors.Wrapf(err, "Error executing search %s - %v", resp.Status, err)
	}

	logger.Printf("Results: %+v", list)

	return nil
}
