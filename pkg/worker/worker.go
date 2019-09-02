package worker

import (
	"log"
	"os"

	"github.com/mchmarny/twitterd/pkg/data"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "worker - ", 0)
)

// Run executes the main worker
func Run() error {

	logger.Println("Initializing data...")
	db, err := data.GetDB()
	if err != nil {
		return errors.Wrap(err, "Error initializing data")
	}
	defer db.Finalize()

	// logger.Println("Getting followers...")
	// ids, err := twitter.GetFollowerIDs()
	// if err != nil {
	// 	return errors.Wrap(err, "Error getting follower IDs")
	// }

	// logger.Printf("Saving %d followers...", len(ids))
	// err = db.SaveDailyFollowers(ids)
	// if err != nil {
	// 	return errors.Wrap(err, "Error saving followers")
	// }

	// new
	logger.Println("Getting new followes...")
	list, err := db.GetNewFollowerIDs()
	if err != nil {
		return errors.Wrap(err, "Error getting new followes")
	}
	logger.Printf("Found %d new followers", len(list))

	// stopped following
	logger.Println("Getting stopped followes...")
	list, err = db.GetStopFollowerIDs()
	if err != nil {
		return errors.Wrap(err, "Error getting stopped followes")
	}
	logger.Printf("Found %d stopped followers", len(list))

	return nil

}
