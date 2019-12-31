package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchCRUD(t *testing.T) {

	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()
	usr := "test"

	c1 := &SearchCriteria{
		ID:               NewID(),
		Name:             "Test Criteria",
		User:             usr,
		Value:            "knative",
		Lang:             "en",
		HasLink:          true,
		FollowerRatioMin: 1.5,
		FollowerCountMax: 2000,
	}

	err := SaveSearchCriteria(ctx, c1)
	assert.Nil(t, err)

	c1.ID = NewID()
	c1.Name = "test2"
	err = SaveSearchCriteria(ctx, c1)
	assert.Nil(t, err)

	list, err := GetSearchCriteria(ctx, usr)
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list, 2)

	for _, c := range list {
		err = SaveSearchCriteria(ctx, c)
		assert.Nil(t, err)
	}

	for _, c := range list {
		err = DeleteSearchCriterion(ctx, c.ID)
		assert.Nil(t, err)
	}

}
