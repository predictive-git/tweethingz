package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserEventsByType(t *testing.T) {

	ctx := context.Background()
	username := "knativeproject"
	yesterday := time.Now().AddDate(0, 0, -1)

	data, err := GetUserEventsByDate(ctx, username, yesterday)
	assert.Nil(t, err)
	assert.NotNil(t, data)

}

func TestToUserEventDateID(t *testing.T) {

	date := time.Now().AddDate(0, 0, -1)
	id1 := toUserEventDateID("aaa", FollowedEventType, date)
	id2 := toUserEventDateID("aaa", FollowedEventType, date)
	assert.Equal(t, id1, id2)

	id1 = toUserEventDateID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", FollowedEventType, date)
	id2 = toUserEventDateID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa2", FollowedEventType, date)
	assert.NotEqual(t, id1, id2)

}
