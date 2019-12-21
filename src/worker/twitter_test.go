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

	err = store.SaveSearchResults(ctx, list)
	assert.Nil(t, err)

	pageSize := 10

	// get full page size of records
	list2, err := store.GetSavedSearchResults(ctx, sc.ID, time.Now(), "", pageSize)
	assert.Nil(t, err)
	assert.NotNil(t, list2)
	assert.Len(t, list2, pageSize)

	// get all saved records when page size is large enough
	list3, err := store.GetSavedSearchResults(ctx, sc.ID, time.Now(), "", 10000)
	assert.Nil(t, err)
	assert.NotNil(t, list3)
	assert.Len(t, list3, len(list))

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

func TestTwitterSearchPaging(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	criteriaID := store.NewID()

	totalSize := 51
	list := getTestSearchResults(criteriaID, totalSize)
	assert.NotNil(t, list)
	assert.Len(t, list, totalSize)

	err := store.SaveSearchResults(ctx, list)
	assert.Nil(t, err)

	list1, err := store.GetSavedSearchResults(ctx, criteriaID, time.Now(), "", 10000)
	assert.Nil(t, err)
	assert.NotNil(t, list1)
	assert.Len(t, list1, totalSize)

	for i, d := range list1 {
		t.Logf("1--C: %s, T[%d]: %s on %s", d.CriteriaID, i, d.ID, d.CreatedAt.Format(store.ISODateFormat))
	}

	pageSize := 10

	// get full page size of records
	list2, err := store.GetSavedSearchResults(ctx, criteriaID, time.Now(), "", pageSize)
	assert.Nil(t, err)
	assert.NotNil(t, list2)
	assert.Len(t, list2, pageSize)

	lastPageKey := list2[len(list2)-1].PageKey

	// get all saved records when page size is large enough
	list3, err := store.GetSavedSearchResults(ctx, criteriaID, time.Now(), lastPageKey, pageSize)
	assert.Nil(t, err)
	assert.NotNil(t, list3)

	for i, d := range list3 {
		t.Logf("2--C: %s, T[%d]: %s on %s", d.CriteriaID, i, d.ID, d.CreatedAt.Format(store.ISODateFormat))
	}

}
