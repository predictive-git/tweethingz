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
	username := "knativeproject"
	yesterday := time.Now().AddDate(0, 0, -1)

	data, err := GetUserEventsSince(ctx, username, yesterday)
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.True(t, len(data) > 0)

}
