package worker

import (
	"context"
	"os"
	"testing"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/stretchr/testify/assert"
)

func TestDerivingArrayDiff(t *testing.T) {

	a1 := []int64{1, 2, 3, 4, 5}
	a2 := []int64{6, 7, 3, 4, 5}
	a3 := getArrayDiff(a1, a2)
	assert.Len(t, a3, 2)

	a4 := getArrayDiff(a2, a1)
	assert.Len(t, a4, 2)
}

func TestUpdateFollowerDataWorker(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	username := os.Getenv("TEST_TW_ACCOUNT")

	forUser, err := store.GetAuthedUser(ctx, username)
	assert.Nil(t, err)

	err = ExecuteFollowerUpdate(ctx, forUser)
	assert.Nil(t, err)

}
