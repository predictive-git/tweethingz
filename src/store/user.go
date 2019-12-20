package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	userCollectionName      = "thingz_user"
	userEventCollectionName = "thingz_event"

	// FollowedEventType when user followes
	FollowedEventType = "followed"

	// UnfollowedEventType when user unfollows
	UnfollowedEventType = "unfollowed"
)

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

// SaveUsers saves multiple users
func SaveUsers(ctx context.Context, users []*SimpleUser) error {

	if len(users) == 0 {
		return nil
	}

	col, err := getCollection(ctx, userCollectionName)
	if err != nil {
		return err
	}

	batch := fsClient.Batch()

	for _, u := range users {
		docRef := col.Doc(ToID(u.Username))
		batch.Set(docRef, u)
	}

	_, err = batch.Commit(ctx)
	return err

}

// GetUser retreaves single user
func GetUser(ctx context.Context, username string) (user *SimpleUser, err error) {
	user = &SimpleUser{}
	err = getByID(ctx, userCollectionName, ToID(username), user)
	return user, err
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

// SaveUserEvents saves multiple user events
func SaveUserEvents(ctx context.Context, users []*SimpleUserEvent) error {

	if len(users) == 0 {
		return nil
	}

	col, err := getCollection(ctx, userEventCollectionName)
	if err != nil {
		return err
	}

	batch := fsClient.Batch()

	for _, u := range users {
		docRef := col.Doc(NewID())
		batch.Set(docRef, u)
	}

	_, err = batch.Commit(ctx)
	return err

}

// GetUserEventsSince retreaves user events since date
// HACK: workaround for lack of support for compounded queries so we look for each day since
// You can only perform range comparisons (<, <=, >, >=) on a single field
func GetUserEventsSince(ctx context.Context, username string, since time.Time) (data []*SimpleUserEvent, err error) {

	col, err := getCollection(ctx, userEventCollectionName)
	if err != nil {
		return nil, err
	}

	data = make([]*SimpleUserEvent, 0)
	for _, d := range getDateRange(since) {
		items, e := GetUserEventsForDate(ctx, col, username, d)
		if e != nil {
			return nil, e
		}
		data = append(data, items...)
	}

	sort.Sort(UserEventByDate(data))

	return

}

// GetUserEventsForDate returns user events for specific date
func GetUserEventsForDate(ctx context.Context, col *firestore.CollectionRef, username string, since time.Time) (data []*SimpleUserEvent, err error) {

	docs, err := col.
		Where("event_user", "==", NormalizeString(username)).
		Where("event_at", "==", since.Format(ISODateFormat)).
		Documents(ctx).
		GetAll()

	// logger.Printf("query for %s user and %s day found %d events",
	// username, since.Format(ISODateFormat), len(docs))

	data = make([]*SimpleUserEvent, 0)

	for _, doc := range docs {
		state := &SimpleUserEvent{}
		if err := doc.DataTo(state); err != nil {
			return nil, fmt.Errorf("error retreiveing user events from %v: %v", doc.Data(), err)
		}
		data = append(data, state)
	}

	return

}
