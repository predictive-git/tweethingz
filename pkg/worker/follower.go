package worker

import (
	"github.com/mchmarny/tweethingz/pkg/data"
	"github.com/mchmarny/tweethingz/pkg/twitter"
	"github.com/pkg/errors"
)

// ProcessFollowers finds new and stopped followers
func ProcessFollowers(username string) error {

	logger.Printf("Starting daily run for %s", username)

	logger.Println("Getting followers...")
	ids, err := twitter.GetFollowerIDs(username)
	if err != nil {
		return errors.Wrap(err, "Error getting follower IDs")
	}

	// followers right now
	logger.Printf("Saving %d followers...", len(ids))
	err = data.SaveDailyFollowers(username, ids)
	if err != nil {
		return errors.Wrap(err, "Error saving followers")
	}

	// new followers
	logger.Println("Getting new followes...")
	list, err := data.GetNewFollowerIDs(username)
	if err != nil {
		return errors.Wrap(err, "Error getting new followes")
	}
	logger.Printf("Found %d new followers", len(list))

	if len(list) > 0 {
		err = getAndSaveUserDetails(username, data.EventNewFollower, list)
		if err != nil {
			return errors.Wrap(err, "Error saving users details")
		}
	}

	// stopped following
	logger.Println("Getting stopped followes...")
	list, err = data.GetStopFollowerIDs(username)
	if err != nil {
		return errors.Wrap(err, "Error getting stopped followes")
	}
	logger.Printf("Found %d stopped followers", len(list))

	if len(list) > 0 {
		err = getAndSaveUserDetails(username, data.EventUnFollowing, list)
		if err != nil {
			return errors.Wrap(err, "Error saving users details")
		}
	}

	return nil

}

func getAndSaveUserDetails(username, eventType string, ids []int64) error {

	// save events
	err := data.SaveFollowerEvents(username, eventType, ids)
	if err != nil {
		return errors.Wrapf(err, "Error saving events: %s for %s", eventType, username)
	}

	return GetAndSaveUsers(ids)

}

// GetAndSaveUsers retreaves and saves users
func GetAndSaveUsers(ids []int64) error {

	if len(ids) == 0 {
		return nil
	}

	pageIDs := []int64{}

	// page in 100s
	for _, id := range ids {
		pageIDs = append(pageIDs, id)
		if len(pageIDs) == 100 { //max twitter page size
			err := getAndSaveUsersPaged(pageIDs)
			if err != nil {
				return err
			}
			pageIDs = []int64{}
		}
	}

	// process left overs
	if len(pageIDs) > 0 { //are there any left over?
		err := getAndSaveUsersPaged(pageIDs)
		if err != nil {
			return err
		}
	}

	return nil

}

func getAndSaveUsersPaged(ids []int64) error {

	if len(ids) == 0 {
		return nil
	}

	// details
	users, err := twitter.GetUsers(ids)
	if err != nil {
		return errors.Wrap(err, "Error getting users details")
	}

	// save details
	err = data.SaveUsers(users)
	if err != nil {
		return errors.Wrap(err, "Error saving new follower events")
	}

	return nil

}
