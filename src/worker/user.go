package worker

import (
	"context"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
	"time"
)

// UpdateUserData runs the background service
func UpdateUserData(ctx context.Context, username string) error {

	// ============================================================================
	// Config
	// ============================================================================
	logger.Printf("Getting config info for %s...", username)
	forUser, err := store.GetAuthedUser(ctx, username)
	if err != nil {
		if err == store.ErrDataNotFound {
			return err
		}
		return errors.Wrapf(err, "error getting authed user for: %s", username)
	}

	// ============================================================================
	// Updating Twitter Details
	// ============================================================================
	logger.Printf("Refreshing twitter details for %s...", forUser.Username)
	twitterUser, err := refreshUserOwnDetails(ctx, forUser)
	if err != nil {
		return errors.Wrapf(err, "error getting twitter %s deails", forUser.Username)
	}

	// ============================================================================
	// IDs of all followers from Twitter
	// ============================================================================
	logger.Printf("Getting IDs of %s twitter followers...", forUser.Username)
	currentFollowerIDs, err := getTwitterFollowerIDs(forUser)
	if err != nil {
		return errors.Wrap(err, "error getting follower IDs")
	}
	logger.Printf("   Follower counts for %s (Profile:%d, IDs:%d)",
		twitterUser.Username, twitterUser.FollowerCount, len(currentFollowerIDs))

	// ============================================================================
	// Yesterday State
	// ============================================================================
	logger.Printf("Getting previous day state for %s...", forUser.Username)
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	yesterdayState, err := store.GetDailyFollowerState(ctx, forUser.Username, yesterday)
	if err != nil {
		return errors.Wrap(err, "error getting yesterday's state")
	}
	logger.Printf("   Yesterday (#%d, +%d, -%d)",
		yesterdayState.FollowerCount, yesterdayState.NewFollowerCount, yesterdayState.UnfollowerCount)

	// ============================================================================
	//  Today State
	// ============================================================================
	logger.Printf("Getting current day state for %s...", forUser.Username)
	todayState, err := store.GetDailyFollowerState(ctx, forUser.Username, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "error getting today's state")
	}
	logger.Printf("   Today (#%d, +%d, -%d)",
		todayState.FollowerCount, todayState.NewFollowerCount, todayState.UnfollowerCount)

	// ============================================================================
	// New Followers
	// ============================================================================
	logger.Printf("Comparing to find new followers for %s...", forUser.Username)
	newFollowerIDs := getArrayDiff(todayState.Followers, currentFollowerIDs)
	logger.Printf("   Yesterday:%d, Today:%d, New Followers:%d",
		yesterdayState.FollowerCount, todayState.FollowerCount, len(newFollowerIDs))

	logger.Printf("Getting new follower details for %s...", forUser.Username)
	if err := pageDownloadFollowerDetail(ctx, forUser, store.FollowedEventType, newFollowerIDs); err != nil {
		return errors.Wrapf(err, "error downloading new follower detail for %s", forUser.Username)
	}

	// ============================================================================
	// New Unfollowers
	// ============================================================================
	logger.Printf("Comparing to find unfollowers for %s...", forUser.Username)
	newUnfollowerIDs := getArrayDiff(todayState.Followers, yesterdayState.Followers)
	logger.Printf("   Yesterday:%d, Today:%d, New Unfollowers:%d",
		yesterdayState.UnfollowerCount, todayState.UnfollowerCount, len(newUnfollowerIDs))

	logger.Printf("Getting unfollowers details for %s...", forUser.Username)
	if err := pageDownloadFollowerDetail(ctx, forUser, store.UnfollowedEventType, newUnfollowerIDs); err != nil {
		return errors.Wrapf(err, "error downloading unfollower detail for %s", forUser.Username)
	}

	// ============================================================================
	// Saving State
	// ============================================================================
	logger.Printf("Saving current state for %s...", forUser.Username)
	// update the current state
	todayState.Followers = currentFollowerIDs
	todayState.FollowerCount = len(currentFollowerIDs)
	todayState.NewFollowerCount = len(newFollowerIDs)
	todayState.UnfollowerCount = len(newUnfollowerIDs)

	// handle first day, don't want to all followers to show as new
	// messes up visualization for the first week
	if todayState.NewFollowerCount == todayState.FollowerCount {
		todayState.NewFollowerCount = 0
	}

	err = store.SaveDailyFollowerState(ctx, todayState)
	if err != nil {
		return errors.Wrap(err, "error saving daily state")
	}

	return nil

}

func refreshUserOwnDetails(ctx context.Context, forUser *store.AuthedUser) (twitterUser *store.SimpleUser, err error) {
	// this returns array of 1
	users, err := GetUserDetails(forUser)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user details")
	}

	if users == nil || len(users) != 1 {
		return nil, errors.Wrapf(err, "expected 1 user, got %d", len(users))
	}

	// save tweeter details for the authed user
	err = store.SaveUsers(ctx, users)
	if err != nil {
		return nil, errors.Wrap(err, "error saving new follower events")
	}

	twitterUser = users[0]

	return
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
			EventDate:  time.Now().UTC().Format(store.ISODateFormat),
			EventType:  eventType,
			EventUser:  store.NormalizeString(forUser.Username),
			SimpleUser: *u,
		}
		events = append(events, ue)
	}

	// save events
	if err = store.SaveUserEvents(ctx, events); err != nil {
		return errors.Wrap(err, "error saving events")
	}

	logger.Printf("   Found %d twitter users and saved %d events for %s",
		len(users), len(events), forUser.Username)

	return nil
}

// returns items from b that are NOT in a
func getArrayDiff(a, b []int64) (diff []int64) {

	if len(b) == 0 {
		return a
	}

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
