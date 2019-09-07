package worker

import (
	"github.com/mchmarny/tweethingz/pkg/data"
	"github.com/pkg/errors"
)

// BackfillFollowers downloads details for users who don't have details in DB
func BackfillFollowers() error {

	logger.Println("Starting backfill run...")
	ids, err := data.GetUserIDsToBackfill()
	if err != nil {
		return errors.Wrap(err, "Error getting backfill IDs")
	}

	if len(ids) == 0 {
		return nil
	}

	err = GetAndSaveUsers(ids)
	if err != nil {
		return err
	}

	logger.Println("Done backfill run")
	return nil

}
