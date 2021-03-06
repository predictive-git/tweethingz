package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTestTwitterAccount() string {
	return os.Getenv("TEST_TW_ACCOUNT")
}

func TestTwitterTestAccount(t *testing.T) {
	val := getTestTwitterAccount()
	if !testing.Short() {
		t.Logf("Test User: %s", val)
	}
	assert.NotEmpty(t, val)
}

func TestGetDailyFollowerStatesSince(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	data, err := GetDailyFollowerStatesSince(context.Background(), getTestTwitterAccount(), yesterday)
	assert.Nil(t, err)
	assert.NotNil(t, data)

}
