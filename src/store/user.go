package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	userCollectionName      = "tweethingz_twitter_user_store"
	userEventCollectionName = "tweethingz_twitter_user_event_store"

	// FollowedEventType when user followes
	FollowedEventType = "followed"

	// UnfollowedEventType when user unfollows
	UnfollowedEventType = "unfollowed"
)

// SimpleUserEvent wraps simple twitter user as an time event
type SimpleUserEvent struct {
	SimpleUser
	EventDate time.Time `firestore:"event_at"`
	EventType string    `firestore:"event_type" json:"event_type"`
}

// UserEventByDate is a custom data structure for array of SimpleUserEvent
type UserEventByDate []*SimpleUserEvent

func (s UserEventByDate) Len() int           { return len(s) }
func (s UserEventByDate) Less(i, j int) bool { return s[i].EventDate.Before(s[j].EventDate) }
func (s UserEventByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

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

	// Meta
	UpdatedOn time.Time `firestore:"updated_on" json:"updated_on"`
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
		docRef := col.Doc(toID(u.Username))
		batch.Set(docRef, u)
	}

	_, err = batch.Commit(ctx)
	return err

}

// GetUser retreaves single user
func GetUser(ctx context.Context, username string) (user *SimpleUser, err error) {

	user = &SimpleUser{}
	err = getByID(ctx, userCollectionName, toID(username), user)

	return user, err
}

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
		docRef := col.Doc(toUserEventDateID(u.Username, u.EventType, u.EventDate))
		batch.Set(docRef, u)
	}

	_, err = batch.Commit(ctx)
	return err

}

func toUserEventDateID(username, eventType string, date time.Time) string {
	return toID(fmt.Sprintf("%s-%s-%s", date.Format(isoDateFormat), username, eventType))
}

// GetUserEventsByDate retreaves user events since date
func GetUserEventsByDate(ctx context.Context, username string, since time.Time) (data []*SimpleUserEvent, err error) {

	col, err := getCollection(ctx, userEventCollectionName)
	if err != nil {
		return nil, err
	}

	data = make([]*SimpleUserEvent, 0)
	for _, d := range getDateRange(since) {
		s, e := getUserEventsForDate(ctx, col, username, d)
		if e != nil {
			return nil, e
		}
		data = append(data, s...)
	}

	sort.Sort(UserEventByDate(data))

	return

}

func getUserEventsForDate(ctx context.Context, col *firestore.CollectionRef, username string, since time.Time) (data []*SimpleUserEvent, err error) {

	docs, err := col.
		Where("username", "==", username).
		Where("event_at", "==", since.Format(isoDateFormat)).
		Documents(ctx).
		GetAll()

	data = make([]*SimpleUserEvent, 0)

	for _, doc := range docs {
		state := &SimpleUserEvent{}
		if err := doc.DataTo(state); err != nil {
			return nil, fmt.Errorf("error retreiveing user events from %v: %v", doc.Data(), err)
		}
	}

	return

}
