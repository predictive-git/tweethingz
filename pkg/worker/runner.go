package worker

import (
	"log"
	"os"

	"github.com/mchmarny/tweethingz/pkg/config"
	"github.com/mchmarny/tweethingz/pkg/data"
	"github.com/mchmarny/tweethingz/pkg/twitter"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "worker: ", 0)
)

// RunItemResult represent run item result
type RunItemResult struct {
	ForUser *data.AuthedUser
	Error   error
}

// Run runs the background service
func Run() error {

	logger.Println("Starting service...")

	c, e := config.GetWorkerConfig()
	if e != nil {
		return errors.Wrap(e, "Error getting worker config")
	}

	// users
	users, err := data.GetAuthedUsers()
	if err != nil {
		return errors.Wrap(err, "Error getting authed users")
	}

	// result
	resultCh := make(chan *RunItemResult, c.ConcurentRefreshLimit)

	// run
	userJobs := make(map[string]*data.AuthedUser)
	for _, u := range users {
		userJobs[u.Username] = u
		go refreshUser(u, resultCh)
	}

	// wait
	for {
		select {
		case r := <-resultCh:
			logger.Printf("Refresh for %s completed with: %s",
				r.ForUser.Username, errorToMessage(r.Error))
			delete(userJobs, r.ForUser.Username)
			if len(userJobs) == 0 {
				return nil
			}
		}
	}

}

func errorToMessage(err error) string {
	if err == nil {
		return "success"
	}
	return err.Error()
}

func refreshUser(forUser *data.AuthedUser, result chan<- *RunItemResult) {

	logger.Printf("Starting refresh for %s", forUser.Username)

	logger.Printf("Refreshing %s details...", forUser.Username)
	if err := refreshUserDetails(forUser); err != nil {
		result <- &RunItemResult{
			ForUser: forUser,
			Error: errors.Wrapf(err, "Error getting %s deails",
				forUser.Username),
		}
		return
	}

	logger.Printf("Reconciling new followers for %s...", forUser.Username)
	if err := reconcileNewFollowers(forUser); err != nil {
		result <- &RunItemResult{
			ForUser: forUser,
			Error: errors.Wrapf(err, "Error reconciling new followers for %s",
				forUser.Username),
		}
		return
	}

	logger.Printf("Reconciling stopped followers for %s...", forUser.Username)
	if err := reconcileStoppedFollowers(forUser); err != nil {
		result <- &RunItemResult{
			ForUser: forUser,
			Error: errors.Wrapf(err, "Error reconciling stopped followers for %s",
				forUser.Username),
		}
		return
	}

	logger.Printf("Refreshing %s followers...", forUser.Username)
	if err := refreshUserFollowers(forUser); err != nil {
		result <- &RunItemResult{
			ForUser: forUser,
			Error: errors.Wrapf(err, "Error refreshing follower IDs for %s",
				forUser.Username),
		}
		return
	}

	// final result
	result <- &RunItemResult{
		ForUser: forUser,
		Error:   nil,
	}

	//TODO: validate that all the new and stopped followers IDs have
	//      corresponding records in the user table

}

func refreshUserDetails(forUser *data.AuthedUser) error {
	// this returns array of 1
	users, err := twitter.GetUserDetails(forUser)
	if err != nil {
		return errors.Wrap(err, "Error getting user details")
	}

	// save details
	err = data.SaveUsers(users)
	if err != nil {
		return errors.Wrap(err, "Error saving new follower events")
	}

	return nil
}

func refreshUserFollowers(forUser *data.AuthedUser) error {

	logger.Printf("Getting %s followers...", forUser.Username)
	ids, err := twitter.GetFollowerIDs(forUser)
	if err != nil {
		return errors.Wrap(err, "Error getting follower IDs")
	}

	// followers right now
	logger.Printf("Saving %d followers for %s...", len(ids), forUser.Username)
	err = data.SaveUserFollowersIDs(forUser.Username, ids)
	if err != nil {
		return errors.Wrap(err, "Error saving followers")
	}

	return nil
}

func reconcileNewFollowers(forUser *data.AuthedUser) error {

	logger.Printf("Getting new followes for %s ...", forUser.Username)
	list, err := data.GetNewFollowerIDs(forUser.Username)
	if err != nil {
		return errors.Wrap(err, "Error getting new followes")
	}

	logger.Printf("Found %d new followers for %s", len(list), forUser.Username)

	if len(list) > 0 {
		err = updateFollowerDetailByEvent(forUser, data.EventNewFollower, list)
		if err != nil {
			return errors.Wrap(err, "Error saving users details")
		}
	}

	return nil
}

func reconcileStoppedFollowers(forUser *data.AuthedUser) error {

	logger.Printf("Getting stopped followes for %s ...", forUser.Username)
	list, err := data.GetStopFollowerIDs(forUser.Username)
	if err != nil {
		return errors.Wrap(err, "Error getting stopped followes")
	}
	logger.Printf("Found %d stopped followers for %s", len(list), forUser.Username)

	if len(list) > 0 {
		err = updateFollowerDetailByEvent(forUser, data.EventUnFollowing, list)
		if err != nil {
			return errors.Wrap(err, "Error saving users details")
		}
	}

	return nil
}

func updateFollowerDetailByEvent(forUser *data.AuthedUser, eventType string, ids []int64) error {

	err := data.SaveFollowerEvents(forUser.Username, eventType, ids)
	if err != nil {
		return errors.Wrapf(err, "Error saving events: %s for %s", eventType, forUser.Username)
	}

	pageIDs := []int64{}

	// page in 100s
	for _, id := range ids {
		pageIDs = append(pageIDs, id)
		if len(pageIDs) == 100 { //max twitter page size
			err := saveFollowerDetailPage(forUser, pageIDs)
			if err != nil {
				return err
			}
			pageIDs = []int64{}
		}
	}

	// process left overs
	if len(pageIDs) > 0 { //are there any left over?
		err := saveFollowerDetailPage(forUser, pageIDs)
		if err != nil {
			return err
		}
	}

	return nil

}

func saveFollowerDetailPage(forUser *data.AuthedUser, ids []int64) error {

	if len(ids) == 0 {
		return nil
	}

	// details
	users, err := twitter.GetUsersFromIDs(forUser, ids)
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
