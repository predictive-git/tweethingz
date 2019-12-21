package worker

import (
	"context"
	"fmt"
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
	filter := &store.SearchCriteria{
		IncludeRT: false,
	}
	assert.True(t, shouldFilterOut(tweet, filter))
	tweet = makeTweet(false, false)
	assert.False(t, shouldFilterOut(tweet, filter))
}

func TestLinkFilter(t *testing.T) {

	tweet := makeTweet(false, true)
	filter := &store.SearchCriteria{
		HasLink: true,
	}
	assert.False(t, shouldFilterOut(tweet, filter))

	tweet = makeTweet(false, false)
	assert.True(t, shouldFilterOut(tweet, filter))

}

func TestAuthorFilter(t *testing.T) {

	tweet := makeTweet(false, false)
	tweet.User.FollowersCount = 9

	filter := &store.SearchCriteria{
		FollowerCountMin: 10,
	}
	assert.True(t, shouldFilterOut(tweet, filter))

	tweet.User.FollowersCount = 11
	assert.False(t, shouldFilterOut(tweet, filter))

}

func TestTwitterSearchWorker(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	username := store.NormalizeString(os.Getenv("TEST_TW_ACCOUNT"))

	usr, err := store.GetAuthedUser(ctx, username)
	assert.Nil(t, err)

	sc := &store.SearchCriteria{
		ID:        store.NewID(),
		User:      usr.Username,
		Name:      "Test Search",
		UpdatedOn: time.Now(),
		Value:     "serverless",
		Lang:      "en",
		SinceID:   0,
	}

	list, err := getSearchResults(ctx, usr, sc)
	assert.Nil(t, err)
	assert.NotNil(t, list)

	err = store.SaveSearchResults(ctx, list)
	assert.Nil(t, err)

	pageSize := 10

	// get full page size of records
	initialPagingKey := store.ToSearchResultPagingKey(sc.ID, time.Now(), "")
	list2, err := store.GetSavedSearchResults(ctx, initialPagingKey, pageSize)
	assert.Nil(t, err)
	assert.NotNil(t, list2)
	assert.Len(t, list2, pageSize)

	lastPageKey := list2[len(list2)-1].Key
	pagingKey := store.ToSearchResultPagingKey(sc.ID, time.Now(), lastPageKey)
	list4, err := store.GetSavedSearchResults(ctx, pagingKey, pageSize)
	assert.Nil(t, err)
	assert.NotNil(t, list4)

}

func getTestSearchResults(criteriaID string, num int) []*store.SimpleTweet {
	list := make([]*store.SimpleTweet, 0)
	for i := 0; i < num; i++ {
		item := &store.SimpleTweet{
			ID:         fmt.Sprintf("id-%d-%s", i, store.NewID()),
			CriteriaID: criteriaID,
			CreatedAt:  time.Now(),
			Text:       fmt.Sprintf("Test tweet %d", i),
		}
		list = append(list, item)
	}
	return list
}
