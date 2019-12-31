package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToID(t *testing.T) {
	values := []string{"@1e6", "ed", "a/Aa", "b-bB", "1", "d75ac5fa-69e6-4f92-8b33-1cdaa3dd2275"}
	date := time.Now().UTC().AddDate(0, 0, -1)

	for _, u := range values {
		// t.Logf("Values[%d]: %v", i, u)
		id1 := toUserDateID(u, date)
		id2 := toUserDateID(u, date)
		assert.True(t, len(id1) == 35) // 3 in prefix + 32 MD5 hash
		assert.True(t, len(id2) == 35) // 3 in prefix + 32 MD5 hash
		assert.Equal(t, id1, id2)
	}
}

func TestDateRangeYesterday(t *testing.T) {

	r := getDateRange(time.Now().UTC().AddDate(0, 0, -1))
	assert.NotNil(t, r)
	assert.Len(t, r, 2)
}

func TestDateRangeToday(t *testing.T) {
	r := getDateRange(time.Now().UTC())
	assert.NotNil(t, r)
	assert.Len(t, r, 1)
}

func TestDateRangeWeek(t *testing.T) {
	r := getDateRange(time.Now().UTC().AddDate(0, 0, -7))
	assert.NotNil(t, r)
	assert.Len(t, r, 8)
}

// func TestPrettyDurationSince(t *testing.T) {

// 	d1 := time.Now().AddDate(0, 0, -1)
// 	s1 := PrettyDurationSince(d1)
// 	t.Logf("1 day: %s", s1)

// 	d2 := time.Now().AddDate(0, 0, -29)
// 	s2 := PrettyDurationSince(d2)
// 	t.Logf("29 days: \n%s", s2)

// 	d3 := time.Now().AddDate(0, -4, -3)
// 	s3 := PrettyDurationSince(d3)
// 	t.Logf("4 months, and 3 days: \n%s", s3)

// 	d4 := time.Now().AddDate(-2, -3, -4)
// 	s4 := PrettyDurationSince(d4)
// 	t.Logf("2 years, 3 months, and 4 days: \n%s", s4)

// }
