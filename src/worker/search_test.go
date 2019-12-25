package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteUserSearches(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	username := "knativeproject"
	err := ExecuteUserSearches(ctx, username)
	assert.Nil(t, err)

}
