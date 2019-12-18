package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDailyFollowerStatesSince(t *testing.T) {

	ctx := context.Background()
	username := os.Getenv("TW_USER")
	yesterday := time.Now().AddDate(0, 0, -1)

	data, err := GetDailyFollowerStatesSince(ctx, username, yesterday)
	assert.Nil(t, err)
	assert.NotNil(t, data)

}
