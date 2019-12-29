package store

import (
	"context"
	"sort"
	"testing"
	"time"

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

func TestSearchResultIDSort(t *testing.T) {

	cIDs := 5
	days := 10
	keys := 7

	ids := make([]string, 0)

	for c := 0; c < cIDs; c++ {
		cID := NewID()
		for d := 0; d < days; d++ {
			day := time.Now().UTC().AddDate(0, 0, -d)
			ids = append(ids, ToSearchResultPagingKey(cID, day, ""))
			for k := 0; k < keys; k++ {
				ids = append(ids, ToSearchResultPagingKey(cID, day, NewID()))
			} // keys
		} // days
	} // criteria

	sort.Strings(ids)

	// for i, id := range ids {
	// 	t.Logf("id[%d] %s", i, id)
	// }

}
