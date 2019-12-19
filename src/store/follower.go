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
)

// DailyFollowerState represents daily follower state
type DailyFollowerState struct {
	Username         string  `firestore:"username" json:"username"`
	StateOn          string  `firestore:"date" json:"date"`
	Followers        []int64 `firestore:"followers" json:"followers"`
	FollowerCount    int     `firestore:"follower_count" json:"follower_count"`
	NewFollowerCount int     `firestore:"new_follower_count" json:"new_follower_count"`
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
		Username: username,
		StateOn:  date.Format(ISODateFormat),
	}
}

func toUserDateID(username string, date time.Time) string {
	return ToID(fmt.Sprintf("%s-%s", date.Format(ISODateFormat), username))
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
			data.Username = username
			data.StateOn = day.Format(ISODateFormat)
			data.Followers = make([]int64, 0)
			return data, nil
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

// GetDailyFollowerStatesSince retrieves map of dates and follower count since the specified date
// func GetDailyFollowerStatesSince(ctx context.Context, username string, since time.Time) (data []*DailyFollowerState, err error) {

// 	col, err := getCollection(ctx, followerCollectionName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	docs, err := col.
// 		Where("username", "==", username).
// 		Where("date", ">=", since.Format(isoDateFormat)).
// 		OrderBy("date", firestore.Desc).
// 		Documents(ctx).
// 		GetAll()

// 	data = make([]*DailyFollowerState, 0)

// 	for _, doc := range docs {
// 		state := &DailyFollowerState{}
// 		if err := doc.DataTo(state); err != nil {
// 			return nil, fmt.Errorf("error retreiveing daily follower state from %v: %v", doc.Data(), err)
// 		}
// 	}

// 	return

// }
