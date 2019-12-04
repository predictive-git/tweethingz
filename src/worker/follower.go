package worker

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/mchmarny/tweethingz/src/data"
	"github.com/pkg/errors"
)

// GetFollowerIDs retreaves follower IDs for config specified user
func GetFollowerIDs(byUser *data.AuthedUser) (ids []int64, err error) {

	client, err := getClient(byUser)
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing client")
	}

	listParam := &twitter.FollowerIDParams{
		ScreenName: byUser.Username,
		Count:      5000, // max per page
	}

	list := []int64{}

	for {
		page, resp, err := client.Followers.IDs(listParam)
		if err != nil {
			return nil, errors.Wrapf(err, "Error paging follower IDs (%s): %v", resp.Status, err)
		}

		// debug
		logger.Printf("Page size:%d, Next:%d\n", len(page.IDs), page.NextCursor)

		list = append(list, page.IDs...)

		// has more IDs?
		if page.NextCursor < 1 {
			break
		}

		// reset cursor
		listParam.Cursor = page.NextCursor
	}

	return list, nil
}
