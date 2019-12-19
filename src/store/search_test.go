package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchCRUD(t *testing.T) {

	ctx := context.Background()

	c1 := &SearchCriteria{
		ID:        NewID(),
		Name:      "Test Criteria",
		User:      "test",
		UpdatedOn: time.Now(),
		Query: &SimpleQuery{
			Value: "knative",
			Lang:  "en",
		},
		Filter: &SimpleFilter{
			HasLink: true,
			Author: &AuthorFilter{
				FollowerRatio: &FloatRange{
					Min: 1.5,
				},
				FollowerCount: &IntRange{
					Min: 2000,
				},
			},
		},
	}

	err := SaveSearchCriteria(ctx, c1)
	assert.Nil(t, err)

	c2, err := GetSearchCriterion(ctx, c1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, c1)
	assert.Equal(t, c1.Name, c2.Name)

	c2.ID = NewID()
	c2.Name = "test2"
	err = SaveSearchCriteria(ctx, c2)
	assert.Nil(t, err)

	list, err := GetSearchCriteria(ctx, "test")
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list, 2)

	err = DeleteSearchCriterion(ctx, c1.ID)
	assert.Nil(t, err)

	err = DeleteSearchCriterion(ctx, c2.ID)
	assert.Nil(t, err)

}
