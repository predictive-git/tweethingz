package worker

import (
	"github.com/mchmarny/twitterd/pkg/data"
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

	pageIDs := []int64{}
	for _, id := range ids {
		pageIDs = append(pageIDs, id)
		if len(pageIDs) == 100 { //max twitter page size
			//TODO process
			logger.Printf("Page: %d", len(pageIDs))
			pageIDs = []int64{}
		}
	}

	if len(pageIDs) > 0 { //are there any left over?
		//TODO process
		logger.Printf("Page: %d", len(pageIDs))
	}

	logger.Println("Done backfill run")
	return nil

}
