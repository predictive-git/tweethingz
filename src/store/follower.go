package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
)

const (
	followerCollectionName = "tweethingz_followers"
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

// NewDailyFollowerState creates a new instance of the DailyFollowerState
func NewDailyFollowerState(username string, date time.Time) *DailyFollowerState {
	return &DailyFollowerState{
		Username: username,
		StateOn:  date.Format(isoDateFormat),
	}
}

func getUserDayKey(username string, date time.Time) string {
	return toID(fmt.Sprintf("%s-%s", date.Format(isoDateFormat), username))
}

// SaveDailyFollowerState saves daily follower state
func SaveDailyFollowerState(ctx context.Context, data *DailyFollowerState) error {

	if data == nil {
		return errors.New("data required")
	}

	docID := getUserDayKey(data.Username, time.Now())

	return save(ctx, followerCollectionName, toID(docID), data)

}

// GetDailyFollowerState retreaves follower data for specific date
func GetDailyFollowerState(ctx context.Context, username string, day time.Time) (data *DailyFollowerState, err error) {

	docID := getUserDayKey(username, day)

	data = &DailyFollowerState{}
	err = getByID(ctx, followerCollectionName, docID, data)
	if err != nil {
		if IsDataNotFoundError(err) {
			logger.Printf("no state data for %s on %v, using defaults", username, day)
			data.Username = username
			data.StateOn = day.Format(isoDateFormat)
			data.Followers = make([]int64, 0)
			return data, nil
		}

		return nil, fmt.Errorf("error getting data by id %s: %v", docID, err)
	}

	return

}

// GetDailyFollowerStatesSince retrieves map of dates and follower count since the specified date
func GetDailyFollowerStatesSince(ctx context.Context, username string, since time.Time) (data []*DailyFollowerState, err error) {

	col, err := getCollection(ctx, followerCollectionName)
	if err != nil {
		return nil, err
	}

	// TODO: figure out if the toID generates sortable values
	// replace dual where with key or figure out how to create an index programmatically
	docs, err := col.
		Where("username", "==", username).
		Where("date", ">=", since.Format(isoDateFormat)).
		OrderBy("date", firestore.Desc).
		Documents(ctx).
		GetAll()

	data = make([]*DailyFollowerState, 0)

	for _, doc := range docs {
		state := &DailyFollowerState{}
		if err := doc.DataTo(state); err != nil {
			return nil, fmt.Errorf("error retreiveing daily follower state from %v: %v", doc.Data(), err)
		}
	}

	return

}
