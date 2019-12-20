package worker

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/stretchr/testify/assert"
)

func makeTweet(rt, links bool) *twitter.Tweet {
	t := &twitter.Tweet{
		User: &twitter.User{
			StatusesCount:   0,
			FollowersCount:  0,
			FriendsCount:    0,
			FavouritesCount: 0,
		},
	}
	if rt {
		t.RetweetedStatus = &twitter.Tweet{}
	}
	if links {
		t.Entities = &twitter.Entities{
			Urls: []twitter.URLEntity{
				twitter.URLEntity{
					URL: "http://test.local",
				},
			},
		}
	}
	return t
}

func TestRetweetFilter(t *testing.T) {
	tweet := makeTweet(true, false)
	filter := &store.SimpleFilter{
		IncludeRT: false,
	}
	assert.True(t, shouldFilterOut(tweet, filter))
	tweet = makeTweet(false, false)
	assert.False(t, shouldFilterOut(tweet, filter))
}

func TestLinkFilter(t *testing.T) {

	tweet := makeTweet(false, true)
	filter := &store.SimpleFilter{
		HasLink: true,
	}
	assert.False(t, shouldFilterOut(tweet, filter))

	tweet = makeTweet(false, false)
	assert.True(t, shouldFilterOut(tweet, filter))

}

func TestAuthorFilter(t *testing.T) {

	tweet := makeTweet(false, false)
	tweet.User.FollowersCount = 9

	filter := &store.SimpleFilter{
		Author: &store.AuthorFilter{
			FollowerCount: &store.IntRange{
				Min: 10,
			},
		},
	}
	assert.True(t, shouldFilterOut(tweet, filter))

	tweet.User.FollowersCount = 11
	assert.False(t, shouldFilterOut(tweet, filter))

}

func TestSearch(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	username := os.Getenv("TEST_TW_ACCOUNT")

	usr, err := store.GetAuthedUser(ctx, username)
	assert.Nil(t, err)

	sc := &store.SearchCriteria{
		ID:        "testID",
		User:      usr.Username,
		Name:      "Test Search",
		UpdatedOn: time.Now(),
		Query: &store.SimpleQuery{
			Value:   "serverless",
			Lang:    "en",
			SinceID: 0,
		},
		Filter: &store.SimpleFilter{},
	}

	list, err := getSearchResults(ctx, usr, sc)
	assert.Nil(t, err)
	assert.NotNil(t, list)

}
