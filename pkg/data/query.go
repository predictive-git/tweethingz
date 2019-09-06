package data

import (
	"database/sql"

	"github.com/mchmarny/twitterd/pkg/twitter"
	"github.com/pkg/errors"
)

const (
	recentUsersDefaultLimit = 10
	followerCountPageSize   = 100
)

// SummaryData represents aggregate data view
type SummaryData struct {
	Username              string                       `json:"username"`
	FollowerCount         int64                        `json:"follower_count"`
	FollowerCountDate     string                       `json:"follower_count_on"`
	FollowerCountSeries   map[string]int64             `json:"follower_count_series"`
	FollowedEventSeries   map[string]int64             `json:"followed_event_series"`
	UnfollowedEventSeries map[string]int64             `json:"unfollowed_event_series"`
	RecentFollowers       []*twitter.SimpleTwitterUser `json:"recent_follower_list"`
	RecentUnfollowers     []*twitter.SimpleTwitterUser `json:"recent_unfollower_list"`
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
		RecentFollowers:       []*twitter.SimpleTwitterUser{},
		RecentUnfollowers:     []*twitter.SimpleTwitterUser{},
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
						   LIMIT ?`, username, followerCountPageSize)

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
						LIMIT ?`, EventNewFollower, EventUnFollowing, username, followerCountPageSize)

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

	// new followers
	list, err := getEventUsers(username, EventNewFollower)
	if err != nil {
		return nil, errors.Wrap(err, "Error quering new follower event users")
	}
	r.RecentFollowers = list

	// new unfollowers
	list, err = getEventUsers(username, EventUnFollowing)
	if err != nil {
		return nil, errors.Wrap(err, "Error quering new unfollower event users")
	}
	r.RecentFollowers = list

	// return loaded object
	return r, err

}

func getEventUsers(username, eventType string) (users []*twitter.SimpleTwitterUser, err error) {

	// follower events
	rows, err := db.Query(`select
			u.id, u.username, u.name, u.description, u.profile_image, u.created_at, u.lang,
			u.location, u.timezone, u.post_count, u.fave_count, u.following_count, u.follower_count
		from users u
		join follower_events e on u.id = e.follower_id
		where e.username = ?
		and e.event_type = ?
		order by e.on_day desc
		limit ?`, username, eventType, recentUsersDefaultLimit)

	if err != nil {
		return nil, errors.Wrap(err, "Error quering event users")
	}

	list := []*twitter.SimpleTwitterUser{}
	for rows.Next() {
		u := &twitter.SimpleTwitterUser{}
		err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Description, &u.ProfileImage, &u.CreatedAt,
			&u.Lang, &u.Location, &u.Timezone, &u.PostCount, &u.FaveCount, &u.FollowingCount,
			&u.FollowerCount)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing follower events series")
		}
		list = append(list, u)
	}

	return list, nil
}
