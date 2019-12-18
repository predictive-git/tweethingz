package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
)

const (
	recentUsersPerDayLimit   = 10 // TODO: make that UI parameter
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
	RecentUserPerDayLimit int `firestore:"recent_users_per_day_limit" json:"recent_users_per_day_limit"`
	NumDaysPeriod         int `firestore:"num_days_period" json:"num_days_period"`
}

// GetSummaryForUser retreaves all summary data for that user
func GetSummaryForUser(ctx context.Context, username string) (data *SummaryData, err error) {

	if username == "" {
		return nil, errors.New("username required")
	}

	// ============================================================================
	// Init data
	// ============================================================================
	data = &SummaryData{
		FollowerCountSeries:   map[string]int{},
		FollowedEventSeries:   map[string]int{},
		UnfollowedEventSeries: map[string]int{},
		Meta: &QueryCriteria{
			RecentUserPerDayLimit: recentUsersPerDayLimit,
			NumDaysPeriod:         recentEventDefaultPeriod,
		},
		RecentFollowers:   make([]*SimpleUserEvent, 0),
		RecentUnfollowers: make([]*SimpleUserEvent, 0),
	}

	// ============================================================================
	// User's saved twitter profile data
	// ============================================================================
	self, err := GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	data.Self = self

	// ============================================================================
	// User follower series
	// ============================================================================
	periodStartDate := time.Now().AddDate(0, 0, -data.Meta.NumDaysPeriod)
	followerData, err := GetDailyFollowerStatesSince(ctx, username, periodStartDate)
	if err != nil {
		return nil, errors.Wrap(err, "error getting followe count")
	}
	for _, item := range followerData {
		data.FollowerCountSeries[item.StateOn] = item.FollowerCount
		data.FollowedEventSeries[item.StateOn] = item.NewFollowerCount
		data.UnfollowedEventSeries[item.StateOn] = -item.UnfollowerCount
	}

	// ============================================================================
	// User follower events for today
	// ============================================================================
	list, err := GetUserEventsSince(ctx, username, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "error quering user events")
	}
	logger.Printf("found %d events for %s", len(list), username)
	for _, item := range list {
		if item.EventType == FollowedEventType {
			data.RecentFollowers = append(data.RecentFollowers, item)
		} else if item.EventType == UnfollowedEventType {
			data.RecentUnfollowers = append(data.RecentUnfollowers, item)
		} else {
			return nil, fmt.Errorf("invalid event type: %s", item.EventType)
		}
	}

	sort.Sort(UserEventByDate(data.RecentFollowers))
	sort.Sort(UserEventByDate(data.RecentUnfollowers))

	data.RecentFollowerCount = len(data.RecentFollowers)
	data.RecentUnfollowerCount = len(data.RecentUnfollowers)

	// ============================================================================
	// Trim results after sort
	// This is inly necessary because the lack of support for compounded queries
	// (you can only perform range comparisons (<, <=, >, >=) on a single field)
	// so I look for each day since, hence the need for after sort and trim
	// ============================================================================
	if len(data.RecentFollowers) > recentUsersPerDayLimit {
		data.RecentFollowers = data.RecentFollowers[0 : recentUsersPerDayLimit-1]
	}
	if len(data.RecentUnfollowers) > recentUsersPerDayLimit {
		data.RecentUnfollowers = data.RecentUnfollowers[0 : recentUsersPerDayLimit-1]
	}

	// return loaded object
	return data, nil

}
