package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
)

const (
	followerCollectionName = "thingz_follower"

	// FollowedEventType when user followes
	FollowedEventType = "followed"

	// UnfollowedEventType when user unfollows
	UnfollowedEventType = "unfollowed"
)

// DailyFollowerState represents daily follower state
type DailyFollowerState struct {
	Username         string  `firestore:"username" json:"username"`
	StateOn          string  `firestore:"date" json:"date"`
	Followers        []int64 `firestore:"followers" json:"followers"`
	FollowerCount    int     `firestore:"follower_count" json:"follower_count"`
	NewFollowers     []int64 `firestore:"new_followers" json:"new_followers"`
	NewFollowerCount int     `firestore:"new_follower_count" json:"new_follower_count"`
	Unfollowers      []int64 `firestore:"unfollowers" json:"unfollowers"`
	UnfollowerCount  int     `firestore:"unfollower_count" json:"unfollower_count"`
}

// DailyFollowerStateByDate is a custom data structure for array of DailyFollowerState
type DailyFollowerStateByDate []*DailyFollowerState

func (s DailyFollowerStateByDate) Len() int           { return len(s) }
func (s DailyFollowerStateByDate) Less(i, j int) bool { return s[i].StateOn < s[j].StateOn }
func (s DailyFollowerStateByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// NewDailyFollowerState creates a new instance of the DailyFollowerState
func NewDailyFollowerState(username string, date time.Time) *DailyFollowerState {
	return &DailyFollowerState{
		Username:     username,
		StateOn:      date.Format(ISODateFormat),
		Followers:    make([]int64, 0),
		NewFollowers: make([]int64, 0),
		Unfollowers:  make([]int64, 0),
	}
}

//============================================================================
// User
//============================================================================

// SimpleUser represents simplified Twitter user
type SimpleUser struct {

	// User details
	Username     string    `firestore:"username" json:"username"`
	Name         string    `firestore:"name" json:"name"`
	Description  string    `firestore:"description" json:"description"`
	ProfileImage string    `firestore:"profile_image" json:"profile_image"`
	CreatedAt    time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt    time.Time `firestore:"updated_at" json:"updated_at"`

	// geo
	Lang     string `firestore:"lang" json:"lang"`
	Location string `firestore:"location" json:"location"`
	Timezone string `firestore:"time_zone" json:"time_zone"`

	// counts
	PostCount      int `firestore:"post_count" json:"post_count"`
	FaveCount      int `firestore:"fave_count" json:"fave_count"`
	FollowingCount int `firestore:"following_count" json:"following_count"`
	FollowerCount  int `firestore:"followers_count" json:"followers_count"`
}

// FormatedCreatedAt returns RFC822 formated CreatedAt
func (s *SimpleUser) FormatedCreatedAt() string {
	if s == nil || s.CreatedAt.IsZero() {
		return ""
	}
	return s.CreatedAt.Format(time.RFC822)
}

//============================================================================
// Event
//============================================================================

// SimpleUserEvent wraps simple twitter user as an time event
type SimpleUserEvent struct {
	SimpleUser
	EventDate string `firestore:"event_at" json:"event_at"`
	EventType string `firestore:"event_type" json:"event_type"`
	EventUser string `firestore:"event_user" json:"event_user"`
}

// UserEventByDate is a custom data structure for array of SimpleUserEvent
type UserEventByDate []*SimpleUserEvent

func (s UserEventByDate) Len() int           { return len(s) }
func (s UserEventByDate) Less(i, j int) bool { return s[i].EventDate < s[j].EventDate }
func (s UserEventByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func toUserDateID(username string, date time.Time) string {
	return ToID(fmt.Sprintf("%s-%s", date.Format(ISODateFormat), NormalizeString(username)))
}

// SaveDailyFollowerState saves daily follower state
func SaveDailyFollowerState(ctx context.Context, data *DailyFollowerState) error {
	if data == nil {
		return errors.New("data required")
	}
	docID := toUserDateID(data.Username, time.Now().UTC())
	return save(ctx, followerCollectionName, docID, data)
}

// GetDailyFollowerState retreaves follower data for specific date
func GetDailyFollowerState(ctx context.Context, username string, day time.Time) (data *DailyFollowerState, err error) {

	docID := toUserDateID(username, day)
	data = &DailyFollowerState{}
	err = getByID(ctx, followerCollectionName, docID, data)
	if err != nil {
		if IsDataNotFoundError(err) {
			// logger.Printf("no state data for %s on %v, using defaults", username, day)
			return NewDailyFollowerState(username, day), nil
		}

		return nil, fmt.Errorf("error getting data by id %s: %v", docID, err)
	}

	return

}

// GetDailyFollowerStatesSince retrieves map of dates and follower count since the specified date
// HACK: workaround for lack of support for compounded queries.
// You can only perform range comparisons (<, <=, >, >=) on a single field
func GetDailyFollowerStatesSince(ctx context.Context, username string, since time.Time) (data []*DailyFollowerState, err error) {

	data = make([]*DailyFollowerState, 0)
	for _, d := range getDateRange(since) {
		s, e := GetDailyFollowerState(ctx, username, d)
		if e != nil {
			return nil, e
		}
		data = append(data, s)
	}

	sort.Sort(DailyFollowerStateByDate(data))

	return

}
