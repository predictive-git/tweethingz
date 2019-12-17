package store

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const (
	recentUsersDefaultLimit  = 10
	recentEventDefaultPeriod = 7
)

// SummaryData represents aggregate data view
type SummaryData struct {
	Self                  *SimpleUser        `firestore:"user" json:"user"`
	FollowerCountSeries   map[string]int     `firestore:"follower_count_series" json:"follower_count_series"`
	FollowedEventSeries   map[string]int     `firestore:"followed_event_series" json:"followed_event_series"`
	UnfollowedEventSeries map[string]int     `firestore:"unfollowed_event_series" json:"unfollowed_event_series"`
	RecentFollowers       []*SimpleUserEvent `firestore:"recent_follower_list" json:"recent_follower_list"`
	RecentUnfollowers     []*SimpleUserEvent `firestore:"recent_unfollower_list" json:"recent_unfollower_list"`
	RecentFollowerCount   int                `firestore:"recent_follower_count" json:"recent_follower_count"`
	RecentUnfollowerCount int                `firestore:"recent_unfollower_count" json:"recent_unfollower_count"`
	Meta                  *QueryCriteria     `firestore:"meta" json:"meta"`
}

// QueryCriteria represents scope of the query
// default for now, will pass this in as criteria
type QueryCriteria struct {
	NumRecentUsers int `firestore:"num_recent_users" json:"num_recent_users"`
	NumDaysPeriod  int `firestore:"num_days_period" json:"num_days_period"`
}

// GetSummaryForUser retreaves all summary data for that user
func GetSummaryForUser(ctx context.Context, username string) (data *SummaryData, err error) {

	if username == "" {
		return nil, errors.New("Null username parameter")
	}

	data = &SummaryData{
		FollowerCountSeries:   map[string]int{},
		FollowedEventSeries:   map[string]int{},
		UnfollowedEventSeries: map[string]int{},
		Meta: &QueryCriteria{
			NumRecentUsers: recentUsersDefaultLimit,
			NumDaysPeriod:  recentEventDefaultPeriod,
		},
	}

	// user details
	self, err := GetUser(ctx, username)
	if err == ErrDataNotFound {
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error getting user details")
	}
	data.Self = self
	sinceDate := time.Now().AddDate(0, 0, -data.Meta.NumDaysPeriod)

	// follower series
	followerData, err := GetDailyFollowerStatesSince(ctx, username, sinceDate)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting followe count")
	}
	for _, item := range followerData {
		data.FollowerCountSeries[item.StateOn] = item.FollowerCount
		data.FollowedEventSeries[item.StateOn] = item.NewFollowerCount
		data.UnfollowedEventSeries[item.StateOn] = item.UnfollowerCount
	}

	// new followers
	list, err := GetUserEventsByDate(ctx, username, sinceDate)
	if err != nil {
		return nil, errors.Wrap(err, "error quering new follower event users")
	}

	for _, item := range list {
		if item.EventType == FollowedEventType {
			data.RecentFollowers = append(data.RecentFollowers, item)
		} else if item.EventType == UnfollowedEventType {
			data.RecentUnfollowers = append(data.RecentUnfollowers, item)
		} else {
			return nil, fmt.Errorf("invalid event type: %s", item.EventType)
		}
	}

	data.RecentFollowerCount = len(data.RecentFollowers)
	data.RecentUnfollowerCount = len(data.RecentUnfollowers)

	// return loaded object
	return data, nil

}
