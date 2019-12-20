package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToID(t *testing.T) {
	values := []string{"@1e6", "ed", "a/Aa", "B-bB", "1", "d75ac5fa-69e6-4f92-8b33-1cdaa3dd2275"}
	date := time.Now().AddDate(0, 0, -1)

	for _, u := range values {
		id1 := toUserDateID(u, date)
		id2 := toUserDateID(u, date)
		assert.True(t, len(id1) == 35) // 3 in prefix + 32 MD5 hash
		assert.True(t, len(id2) == 35) // 3 in prefix + 32 MD5 hash
		assert.Equal(t, id1, id2)
	}
}

func TestDateRangeYesterday(t *testing.T) {

	r := getDateRange(time.Now().AddDate(0, 0, -1))
	assert.NotNil(t, r)
	assert.Len(t, r, 2)
}

func TestDateRangeToday(t *testing.T) {
	r := getDateRange(time.Now())
	assert.NotNil(t, r)
	assert.Len(t, r, 1)
}

func TestDateRangeWeek(t *testing.T) {
	r := getDateRange(time.Now().AddDate(0, 0, -7))
	assert.NotNil(t, r)
	assert.Len(t, r, 8)
}
