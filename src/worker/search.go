package worker

import (
	"context"

	"github.com/mchmarny/tweethingz/src/store"
	"github.com/pkg/errors"
)

// ExecuteUserSearches runs the background service
func ExecuteUserSearches(ctx context.Context, username string) error {

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
	// Criteria
	// ============================================================================
	criteria, err := store.GetSearchCriteria(ctx, forUser.Username)
	if err != nil {
		if err == store.ErrDataNotFound {
			return nil
		}
		return errors.Wrapf(err, "error getting search criteria user for: %s", username)
	}

	if criteria == nil || len(criteria) == 0 {
		return nil
	}

	for _, c := range criteria {

		logger.Printf("executing criteria %s...", c.Name)

		tweets, err := getSearchResults(ctx, forUser, c)
		if err != nil {
			return errors.Wrapf(err, "error executing search criteria %s user for %s: %v", c.ID, c.User, err)
		}

		if err = store.SaveSearchResults(ctx, tweets); err != nil {
			return errors.Wrapf(err, "error saving search criteria %s results for %s: %v", c.ID, c.User, err)
		}

		// save the updated search criteria (lastID and exec time)
		if err = store.SaveSearchCriteria(ctx, c); err != nil {
			return errors.Wrapf(err, "error saving search criteria %s for %s: %v", c.ID, c.User, err)
		}

	}

	return nil

}
