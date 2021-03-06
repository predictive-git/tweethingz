package worker

import (
	"context"
	"time"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
)

const (
	recentEventDefaultPeriod = 6 // inclusive to 6 past + today == 7 days
)

// SummaryData represents aggregate data view
type SummaryData struct {
	Self                  *store.SimpleUser  `firestore:"user" json:"user"`
	FollowerCountSeries   map[string]int     `firestore:"follower_count_series" json:"follower_count_series"`
	FollowedEventSeries   map[string]int     `firestore:"followed_event_series" json:"followed_event_series"`
	UnfollowedEventSeries map[string]int     `firestore:"unfollowed_event_series" json:"unfollowed_event_series"`
	AvgEventSeries        map[string]float32 `firestore:"avg_event_series" json:"avg_event_series"`
	NewFollowerCount      int                `firestore:"recent_follower_count" json:"recent_follower_count"`
	UnfollowerCount       int                `firestore:"recent_unfollower_count" json:"recent_unfollower_count"`
	Meta                  *QueryMetaData     `firestore:"meta" json:"meta"`
	UpdatedOn             time.Time          `firestore:"updated_on" json:"updated_on"`
}

// QueryMetaData represents scope of the query
// default for now, will pass this in as criteria
type QueryMetaData struct {
	NumDaysPeriod int `firestore:"num_days_period" json:"num_days_period"`
}

// GetSummaryForUser retreaves all summary data for that user
func GetSummaryForUser(ctx context.Context, forUser *store.AuthedUser) (data *SummaryData, err error) {

	if forUser == nil {
		return nil, errors.New("user required")
	}

	// ============================================================================
	// Init data
	// ============================================================================
	data = &SummaryData{
		FollowerCountSeries:   map[string]int{},
		FollowedEventSeries:   map[string]int{},
		UnfollowedEventSeries: map[string]int{},
		AvgEventSeries:        map[string]float32{},
		Meta: &QueryMetaData{
			NumDaysPeriod: recentEventDefaultPeriod,
		},
	}

	// ============================================================================
	// User's Twitter profile data
	// ============================================================================
	self, err := GetTwitterUserDetails(forUser)
	if err != nil {
		return nil, err
	}
	data.Self = self
	data.UpdatedOn = self.UpdatedAt

	// ============================================================================
	// User follower series
	// ============================================================================
	periodStartDate := time.Now().UTC().AddDate(0, 0, -data.Meta.NumDaysPeriod)
	followerData, err := store.GetDailyFollowerStatesSince(ctx, forUser.Username, periodStartDate)
	if err != nil {
		return nil, errors.Wrap(err, "error getting followe count")
	}
	var runSum float32 = 0
	for i, item := range followerData {
		day := i + 1
		// total
		data.FollowerCountSeries[item.StateOn] = item.FollowerCount
		// followers (+/-)
		data.FollowedEventSeries[item.StateOn] = item.NewFollowerCount
		data.UnfollowedEventSeries[item.StateOn] = -item.UnfollowerCount
		// avg
		runSum += float32(item.NewFollowerCount - item.UnfollowerCount)
		data.AvgEventSeries[item.StateOn] = runSum / float32(day)
		// logger.Printf("day[%d] +:%d -%d a:%f ra:%f",
		//      day, item.NewFollowerCount, item.UnfollowerCount, runSum, data.AvgEventSeries[item.StateOn])
	}

	if followerData != nil {
		lastDay := followerData[len(followerData)-1]
		data.NewFollowerCount = lastDay.NewFollowerCount
		data.UnfollowerCount = lastDay.UnfollowerCount
	}

	// return loaded object
	return data, nil

}
