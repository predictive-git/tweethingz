package worker

import (
	"log"
	"os"

	"github.com/mchmarny/twitterd/pkg/data"
	"github.com/mchmarny/twitterd/pkg/twitter"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "worker: ", 0)
)

// RunDaily executes the main worker
func RunDaily(username string) error {

	logger.Printf("Starting daily run for %s", username)
	defer data.Finalize()

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

	// stopped following
	logger.Println("Getting stopped followes...")
	list, err = data.GetStopFollowerIDs(username)
	if err != nil {
		return errors.Wrap(err, "Error getting stopped followes")
	}
	logger.Printf("Found %d stopped followers", len(list))

	return nil

}
