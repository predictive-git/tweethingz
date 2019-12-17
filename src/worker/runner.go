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

// Run runs the background service
func Run(ctx context.Context, username string) error {

	logger.Printf("Starting worker for %s...", username)
	forUser, err := store.GetAuthedUser(ctx, username)
	if err != nil {
		errors.Wrapf(err, "error getting authed user for: %s", username)
	}

	logger.Printf("Refreshing twitter details for %s...", forUser.Username)
	if err := refreshUserDetails(ctx, forUser); err != nil {
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
	isCurrentStateNew := newDailyState.FollowerCount == 0
	newDailyState.Followers = currentFollowerIDs
	newDailyState.FollowerCount = len(currentFollowerIDs)

	logger.Printf("Identifying new followers for %s...", forUser.Username)
	newFollowerIDs := getArrayDiff(yesterdayState.Followers, newDailyState.Followers)
	newDailyState.NewFollowerCount = len(newFollowerIDs)

	logger.Printf("Yesterday:%d, Today:%d", yesterdayState.FollowerCount, newDailyState.FollowerCount)

	// to avoid long refreshes the first day run only if
	// A) there was a state the previous date (2nd day+ run)
	// B) the current day state is new (1st run)
	// C) the number of new followers is so small that it doesn't matter
	if yesterdayState.FollowerCount > 0 || isCurrentStateNew || len(newFollowerIDs) < 100 {

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

	}

	logger.Printf("Saving current state for %s...", forUser.Username)
	err = store.SaveDailyFollowerState(ctx, newDailyState)
	if err != nil {
		return errors.Wrap(err, "error saving daily state")
	}

	return nil

}

func refreshUserDetails(ctx context.Context, forUser *store.AuthedUser) error {
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
			EventDate:  time.Now(),
			EventType:  eventType,
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
