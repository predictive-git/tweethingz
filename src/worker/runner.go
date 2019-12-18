package worker

import (
	"context"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
	"time"
)

// RunItemResult represent run item result
type RunItemResult struct {
	ForUser *store.AuthedUser
	Error   error
}

// UpdateUserData runs the background service
func UpdateUserData(ctx context.Context, username string) error {

	logger.Printf("Starting worker for %s...", username)
	forUser, err := store.GetAuthedUser(ctx, username)
	if err != nil {
		errors.Wrapf(err, "error getting authed user for: %s", username)
	}

	logger.Printf("Refreshing twitter details for %s...", forUser.Username)
	if err := refreshUserOwnDetails(ctx, forUser); err != nil {
		return errors.Wrapf(err, "error getting twitter %s deails", forUser.Username)
	}

	logger.Printf("Getting %s twitter followers...", forUser.Username)
	currentFollowerIDs, err := getFollowerIDs(forUser)
	if err != nil {
		return errors.Wrap(err, "error getting follower IDs")
	}

	logger.Printf("Getting previous day state for %s...", forUser.Username)
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayState, err := store.GetDailyFollowerState(ctx, forUser.Username, yesterday)
	if err != nil {
		return errors.Wrap(err, "error getting yesterday's state")
	}

	logger.Printf("Getting current state for %s...", forUser.Username)
	newDailyState, err := store.GetDailyFollowerState(ctx, forUser.Username, time.Now())
	if err != nil {
		return errors.Wrap(err, "error getting today's state")
	}

	logger.Printf("Identifying new followers for %s...", forUser.Username)
	var newFollowerIDs []int64

	// if didn't have data yesterday but has today than this is a subsequent view on the first day
	if yesterdayState.FollowerCount == 0 && newDailyState.FollowerCount > 0 {
		logger.Printf("View: 2nd+ view on the first day for %s...", forUser.Username)
		newFollowerIDs = getArrayDiff(newDailyState.Followers, currentFollowerIDs)
	} else {
		logger.Printf("View: with yesterday data for for %s...", forUser.Username)
		newFollowerIDs = getArrayDiff(yesterdayState.Followers, currentFollowerIDs)
	}

	logger.Printf("Yesterday:%d, Current:%d, New:%d",
		yesterdayState.FollowerCount, newDailyState.FollowerCount, len(newFollowerIDs))

	// update the current state
	newDailyState.Followers = currentFollowerIDs
	newDailyState.FollowerCount = len(currentFollowerIDs)
	newDailyState.NewFollowerCount = len(newFollowerIDs)

	logger.Printf("Process new followers for %s...", forUser.Username)
	if err := pageDownloadFollowerDetail(ctx, forUser, store.FollowedEventType, newFollowerIDs); err != nil {
		return errors.Wrapf(err, "error downloading new follower detail for %s",
			forUser.Username)
	}

	logger.Printf("Deriving unfollowers for %s...", forUser.Username)
	newUnfollowerIDs := getArrayDiff(newDailyState.Followers, yesterdayState.Followers)
	newDailyState.UnfollowerCount = len(newUnfollowerIDs)

	logger.Printf("Process unfollowers for %s...", forUser.Username)
	if err := pageDownloadFollowerDetail(ctx, forUser, store.UnfollowedEventType, newUnfollowerIDs); err != nil {
		return errors.Wrapf(err, "error downloading unfollower detail for %s",
			forUser.Username)
	}

	logger.Printf("Saving current state for %s...", forUser.Username)
	err = store.SaveDailyFollowerState(ctx, newDailyState)
	if err != nil {
		return errors.Wrap(err, "error saving daily state")
	}

	return nil

}

func refreshUserOwnDetails(ctx context.Context, forUser *store.AuthedUser) error {
	// this returns array of 1
	users, err := GetUserDetails(forUser)
	if err != nil {
		return errors.Wrap(err, "Error getting user details")
	}

	// save tweeter details for the authed user
	err = store.SaveUsers(ctx, users)
	if err != nil {
		return errors.Wrap(err, "Error saving new follower events")
	}

	return nil
}

func pageDownloadFollowerDetail(ctx context.Context, forUser *store.AuthedUser, eventType string, ids []int64) error {

	if len(ids) == 0 {
		return nil
	}

	pageIDs := []int64{}

	// page in 100s
	for _, id := range ids {
		pageIDs = append(pageIDs, id)
		if len(pageIDs) == 100 { //max twitter page size
			err := saveFollowerDetails(ctx, forUser, eventType, pageIDs)
			if err != nil {
				return err
			}
			pageIDs = []int64{}
		}
	}

	// process left overs
	if len(pageIDs) > 0 { //are there any left over?
		err := saveFollowerDetails(ctx, forUser, eventType, pageIDs)
		if err != nil {
			return err
		}
	}

	return nil

}

func saveFollowerDetails(ctx context.Context, forUser *store.AuthedUser, eventType string, ids []int64) error {

	if len(ids) == 0 {
		return nil
	}

	// details
	users, err := GetUsersFromIDs(forUser, ids)
	if err != nil {
		return errors.Wrap(err, "error getting users details")
	}

	// save details
	if err = store.SaveUsers(ctx, users); err != nil {
		return errors.Wrap(err, "error saving users")
	}

	events := make([]*store.SimpleUserEvent, 0)
	for _, u := range users {
		ue := &store.SimpleUserEvent{
			EventDate:  time.Now().Format(store.ISODateFormat),
			EventType:  eventType,
			EventUser:  forUser.Username,
			SimpleUser: *u,
		}
		events = append(events, ue)
	}

	// save events
	if err = store.SaveUserEvents(ctx, events); err != nil {
		return errors.Wrap(err, "error saving events")
	}

	return nil
}

// returns items from b that are NOT in a
func getArrayDiff(a, b []int64) (diff []int64) {

	m := make(map[int64]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
