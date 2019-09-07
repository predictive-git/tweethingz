package twitter

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/mchmarny/tweethingz/pkg/config"
	"github.com/pkg/errors"
)

// GetFollowerIDs retreaves follower IDs for config specified user
func GetFollowerIDs(username string) (ids []int64, err error) {

	cfg, err := config.GetTwitterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Twitter config")
	}

	listParam := &twitter.FollowerIDParams{
		ScreenName: username,
		Count:      5000, // max per page
	}

	list := []int64{}

	for {
		page, resp, err := getClient(cfg).Followers.IDs(listParam)
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
