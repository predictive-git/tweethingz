package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserEventsByType(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	username := "KnativeProject"
	yesterday := time.Now().AddDate(0, 0, -1)

	data, err := GetUserEventsSince(ctx, username, yesterday)
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.True(t, len(data) > 0)

}

func TestToUserEventDateID(t *testing.T) {

	date := time.Now().AddDate(0, 0, -1).Format(ISODateFormat)
	id1 := toUserEventDateID("aaa", FollowedEventType, date)
	id2 := toUserEventDateID("aaa", FollowedEventType, date)
	assert.Equal(t, id1, id2)

	id1 = toUserEventDateID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", FollowedEventType, date)
	id2 = toUserEventDateID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa2", FollowedEventType, date)
	assert.NotEqual(t, id1, id2)

}
