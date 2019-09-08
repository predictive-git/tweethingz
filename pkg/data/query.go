package data

import (
	"time"

	"github.com/pkg/errors"
)

const (
	recentUsersDefaultLimit  = 10
	recentEventDefaultPeriod = 7
)

// SummaryData represents aggregate data view
type SummaryData struct {
	Self                  *SimpleUser        `json:"user"`
	FollowerCountSeries   map[string]int64   `json:"follower_count_series"`
	FollowedEventSeries   map[string]int64   `json:"followed_event_series"`
	UnfollowedEventSeries map[string]int64   `json:"unfollowed_event_series"`
	RecentFollowers       []*SimpleUserEvent `json:"recent_follower_list"`
	RecentUnfollowers     []*SimpleUserEvent `json:"recent_unfollower_list"`
	RecentFollowerCount   int64              `json:"recent_follower_count"`
	RecentUnfollowerCount int64              `json:"recent_unfollower_count"`
	Meta                  *QueryCriteria     `json:"meta"`
}

// QueryCriteria represents scope of the query
// default for now, will pass this in as criteria
type QueryCriteria struct {
	NumRecentUsers int `json:"num_recent_users"`
	NumDaysPeriod  int `json:"num_days_period"`
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
		FollowerCountSeries:   map[string]int64{},
		FollowedEventSeries:   map[string]int64{},
		UnfollowedEventSeries: map[string]int64{},
		Meta: &QueryCriteria{
			NumRecentUsers: recentUsersDefaultLimit,
			NumDaysPeriod:  recentEventDefaultPeriod,
		},
	}

	// user details
	self, err := getUser(username)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting user details")
	}
	r.Self = self

	// follower count series
	rows, err := db.Query(`SELECT DATE_FORMAT(on_day, "%Y-%m-%d") as count_date,
								count(*) as num_of_followers
						   FROM followers
						   WHERE username = ?
						   GROUP BY count_date
						   ORDER BY count_date
						   LIMIT ?`, username, r.Meta.NumDaysPeriod)
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
						LIMIT ?`, EventNewFollower, EventUnFollowing, username, r.Meta.NumDaysPeriod)

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
	r.RecentUnfollowers = list

	// calculate period event counts
	r.RecentFollowerCount = getRecentEventCount(r.FollowedEventSeries, r.Meta.NumDaysPeriod)
	r.RecentUnfollowerCount = getRecentEventCount(r.UnfollowedEventSeries, r.Meta.NumDaysPeriod)

	// return loaded object
	return r, err

}

func getRecentEventCount(events map[string]int64, forLastDays int) int64 {

	if events == nil {
		return 0
	}

	var result int64
	sinceDate := time.Now().AddDate(0, 0, -forLastDays).Format(isoDateFormat)
	for k, v := range events {
		if k >= sinceDate {
			result = result + v
		}
	}

	return result
}

func getEventUsers(username, eventType string) (users []*SimpleUserEvent, err error) {

	// follower events
	rows, e := db.Query(`select
			u.id, u.username, u.name, u.description, u.profile_image, u.created_at,
			u.lang, u.location, u.timezone, u.post_count, u.fave_count, u.following_count,
			u.follower_count, e.on_day
		from users u
		join follower_events e on u.id = e.follower_id
		where e.username = ?
		and e.event_type = ?
		order by e.on_day desc
		limit ?`, username, eventType, recentUsersDefaultLimit)

	if e != nil {
		return nil, errors.Wrap(err, "Error quering event users")
	}

	list := []*SimpleUserEvent{}
	for rows.Next() {
		u := &SimpleUserEvent{}
		e := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Description, &u.ProfileImage, &u.CreatedAt,
			&u.Lang, &u.Location, &u.Timezone, &u.PostCount, &u.FaveCount, &u.FollowingCount,
			&u.FollowerCount, &u.EventDate)
		if e != nil {
			return nil, errors.Wrap(e, "Error parsing follower events series")
		}
		list = append(list, u)
	}

	return list, nil
}

func getUser(username string) (user *SimpleUser, err error) {

	// follower events
	row := db.QueryRow(`select id, username, name, description,
		profile_image, created_at, lang, location, timezone,
		post_count, fave_count, following_count, follower_count, updated_on
		from users where username = ?`, username)

	u := &SimpleUser{}
	e := row.Scan(&u.ID, &u.Username, &u.Name, &u.Description, &u.ProfileImage, &u.CreatedAt,
		&u.Lang, &u.Location, &u.Timezone, &u.PostCount, &u.FaveCount, &u.FollowingCount,
		&u.FollowerCount, &u.UpdatedOn)
	if e != nil {
		return nil, errors.Wrap(e, "Error parsing user")
	}

	return u, nil
}
