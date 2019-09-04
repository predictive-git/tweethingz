package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSummaryForUser(t *testing.T) {
	data, err := GetSummaryForUser("mchmarny")
	logger.Printf("Data: %+v", data)
	assert.Nil(t, err)
	assert.NotNil(t, data)
}
