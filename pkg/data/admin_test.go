package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidLookupUIUser(t *testing.T) {
	testEmail := "test@domain.com"
	testUser, err := LookupUIUser(testEmail, "test")
	assert.Nil(t, err)
	assert.NotNil(t, testUser)
}

func TestInvalidLookupUIUser(t *testing.T) {
	testEmail := "test2@domain.com"
	testUser, err := LookupUIUser(testEmail, "test")
	assert.Nil(t, err)
	assert.Empty(t, testUser)
}
