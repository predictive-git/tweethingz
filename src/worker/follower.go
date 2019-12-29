package worker

import (
	"context"
	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
	"time"
)

// ExecuteFollowerUpdate runs the background service
func ExecuteFollowerUpdate(ctx context.Context, forUser *store.AuthedUser) error {

	if forUser == nil {
		return errors.New("authUser parameter required")
	}

	// ============================================================================
	// Twitter Details
	// ============================================================================
	twitterUser, err := GetTwitterUserDetails(forUser)
	if err != nil {
		return errors.Wrapf(err, "error getting twitter %s deails", forUser.Username)
	}

	// ============================================================================
	// IDs of all followers from Twitter
	// ============================================================================
	followerIDs, err := GetTwitterFollowerIDs(forUser)
	if err != nil {
		return errors.Wrap(err, "error getting follower IDs")
	}
	logger.Printf("Follower counts for %s (Profile:%d, IDs:%d)",
		twitterUser.Username, twitterUser.FollowerCount, len(followerIDs))

	// ============================================================================
	// Yesterday State
	// ============================================================================
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	yesterdayState, err := store.GetDailyFollowerState(ctx, forUser.Username, yesterday)
	if err != nil {
		return errors.Wrap(err, "error getting yesterday's state")
	}
	logger.Printf("Yesterday (count:%d, +%d, -%d)",
		yesterdayState.FollowerCount, yesterdayState.NewFollowerCount, yesterdayState.UnfollowerCount)

	// ============================================================================
	//  Today State
	// ============================================================================
	todayState, err := store.GetDailyFollowerState(ctx, forUser.Username, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "error getting today's state")
	}
	logger.Printf("Today (count:%d, +%d, -%d)",
		todayState.FollowerCount, todayState.NewFollowerCount, todayState.UnfollowerCount)

	// ============================================================================
	// New Followers
	// ============================================================================
	newFollowerIDs := getArrayDiff(yesterdayState.Followers, followerIDs)
	logger.Printf("New Followers (y:%d, +:%d)", yesterdayState.FollowerCount, len(newFollowerIDs))

	// ============================================================================
	// New Unfollowers
	// ============================================================================
	newUnfollowerIDs := getArrayDiff(followerIDs, yesterdayState.Followers)
	logger.Printf("Unfollowers (y:%d, -:%d)", yesterdayState.FollowerCount, len(newUnfollowerIDs))

	// ============================================================================
	// Update State
	// ============================================================================
	todayState.Followers = followerIDs
	todayState.FollowerCount = len(followerIDs)

	// populate diffs only if not first day
	if yesterdayState.FollowerCount > 0 {
		todayState.NewFollowers = newFollowerIDs
		todayState.NewFollowerCount = len(newFollowerIDs)
		todayState.Unfollowers = newUnfollowerIDs
		todayState.UnfollowerCount = len(newUnfollowerIDs)
	}

	// ============================================================================
	// Save State
	// ============================================================================
	logger.Printf("Saving followers for %s (: #%d, +%d, -%d)", forUser.Username,
		todayState.FollowerCount, todayState.NewFollowerCount, todayState.UnfollowerCount)
	err = store.SaveDailyFollowerState(ctx, todayState)
	if err != nil {
		return errors.Wrap(err, "error saving daily state")
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
