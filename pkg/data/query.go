package data

import (
	"database/sql"

	"github.com/pkg/errors"
)

// SummaryData represents aggregate data view
type SummaryData struct {
	Username              string           `json:"username"`
	FollowerCount         int64            `json:"follower_count"`
	FollowerCountDate     string           `json:"follower_count_on"`
	FollowerCountSeries   map[string]int64 `json:"follower_count_series"`
	FollowedEventSeries   map[string]int64 `json:"followed_event_series"`
	UnfollowedEventSeries map[string]int64 `json:"unfollowed_event_series"`
}

// GetSummaryForUser retreaves all summary data for that user
func GetSummaryForUser(username string) (data *SummaryData, err error) {

	if username == "" {
		return nil, errors.New("Null username parameter")
	}

	if err := initDB(); err != nil {
		return nil, err
	}

	r := &SummaryData{
		Username:              username,
		FollowerCountSeries:   map[string]int64{},
		FollowedEventSeries:   map[string]int64{},
		UnfollowedEventSeries: map[string]int64{},
	}

	// follower counts
	row := db.QueryRow(`SELECT
							DATE_FORMAT(MAX(on_day), "%Y-%m-%d") as count_date,
							COUNT(*) as num_of_followers
						FROM followers
						WHERE username = ?
						AND on_day = (
							SELECT MAX(on_day) FROM followers WHERE username = ?
						)`, username, username)

	err = row.Scan(&r.FollowerCountDate, &r.FollowerCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "Error parsing follower count select results")
	}

	// follower count series
	rows, err := db.Query(`SELECT DATE_FORMAT(on_day, "%Y-%m-%d") as count_date,
								count(*) as num_of_followers
						   FROM followers
						   WHERE username = ?
						   GROUP BY count_date
						   ORDER BY count_date
						   LIMIT 100`, username)

	if err != nil {
		return nil, errors.Wrap(err, "Error quering follower count series")
	}

	for rows.Next() {
		var countDate string
		var followerCount int64
		err := rows.Scan(&countDate, &followerCount)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing follower count series")
		}
		r.FollowerCountSeries[countDate] = followerCount
	}

	// follower events
	rows, err = db.Query(`SELECT
							DATE_FORMAT(on_day, "%Y-%m-%d") as count_date,
							SUM(CASE WHEN event_type = ? THEN 1 ELSE 0 END) as followed,
							SUM(CASE WHEN event_type = ? THEN 1 ELSE 0 END) as unhallowed
						FROM follower_events
						WHERE
							username = ?
						GROUP BY
							count_date
						ORDER BY count_date
						LIMIT 100`, EventNewFollower, EventUnFollowing, username)

	if err != nil {
		return nil, errors.Wrap(err, "Error quering follower events series")
	}

	for rows.Next() {
		var countDate string
		var followerCount int64
		var unfollowerCount int64
		err := rows.Scan(&countDate, &followerCount, &unfollowerCount)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing follower events series")
		}
		r.FollowedEventSeries[countDate] = followerCount
		r.UnfollowedEventSeries[countDate] = unfollowerCount - unfollowerCount*2
	}

	// return loaded object
	return r, err

}
