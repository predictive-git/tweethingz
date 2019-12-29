package worker

import (
	"context"
	"os"
	"testing"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/stretchr/testify/assert"
)

func TestExecuteUserSearches(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	username := os.Getenv("TEST_TW_ACCOUNT")

	forUser, err := store.GetAuthedUser(ctx, username)
	assert.Nil(t, err)

	err = ExecuteUserSearches(ctx, forUser)
	assert.Nil(t, err)

}
