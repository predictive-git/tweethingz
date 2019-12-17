package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToID(t *testing.T) {

	vals := []string{"@1e6", "a/Aa", "B-bB", "1", "d75ac5fa-69e6-4f92-8b33-1cdaa3dd2275"}
	date := time.Now().AddDate(0, 0, -1)

	for _, d := range vals {
		id := toUserDateID(d, date)
		// is at least 10 char ("id-" + 9)
		assert.True(t, len(id) > 10)
		// the non prefix part is numeric
		assert.True(t, isNumeric(id[3:]))
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
