package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidLookupUIUser(t *testing.T) {

	testUser, err := LookupUIUser("mark@chmarny.com", "test")
	logger.Printf("User: %s, Error: %v", testUser, err)
	assert.Nil(t, err)
	assert.NotEmpty(t, testUser)
}

func TestInvalidLookupUIUser(t *testing.T) {
	testUser, err := LookupUIUser("test2@domain.com", "test")
	logger.Printf("User: %s, Error: %v", testUser, err)
	assert.Equal(t, ErrUserNotFound, err, "Should have thrown not found error")
	assert.Empty(t, testUser)
}
