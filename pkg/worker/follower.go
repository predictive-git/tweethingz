package worker

import (
	"log"
	"os"

	"github.com/mchmarny/twitterd/pkg/data"
	"github.com/mchmarny/twitterd/pkg/twitter"
	"github.com/pkg/errors"
)

const (
	newFollowerEvent      = "followed"
	stoppedFollowingEvent = "unhallowed"
)

var (
	logger = log.New(os.Stdout, "worker: ", 0)
)

// ProcessFollowers finds new and stopped followers
func ProcessFollowers(username string) error {

	logger.Printf("Starting daily run for %s", username)

	logger.Println("Getting followers...")
	ids, err := twitter.GetFollowerIDs(username)
	if err != nil {
		return errors.Wrap(err, "Error getting follower IDs")
	}

	logger.Printf("Saving %d followers...", len(ids))
	err = data.SaveDailyFollowers(username, ids)
	if err != nil {
		return errors.Wrap(err, "Error saving followers")
	}

	// new
	logger.Println("Getting new followes...")
	list, err := data.GetNewFollowerIDs(username)
	if err != nil {
		return errors.Wrap(err, "Error getting new followes")
	}
	logger.Printf("Found %d new followers", len(list))

	if len(list) > 0 {
		err = data.SaveFollowerEvents(username, newFollowerEvent, list)
		if err != nil {
			return errors.Wrap(err, "Error saving new follower events")
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
		err = data.SaveFollowerEvents(username, stoppedFollowingEvent, list)
		if err != nil {
			return errors.Wrap(err, "Error saving stopped following events")
		}
	}

	return nil

}
