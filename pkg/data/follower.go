package data

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

func getFollowerDailyCount(username string, day time.Time) (count int64, err error) {

	if err := initDB(); err != nil {
		return 0, err
	}

	row := db.QueryRow(`SELECT count(*) FROM followers
		WHERE username = ? AND on_day = ?`, username, day.Format(isoDateFormat))

	var idCount int64
	err = row.Scan(&idCount)
	if err != nil {
		return 0, errors.Wrap(err, "Error parsing follower count results")
	}

	return idCount, nil

}

func getFollowerDiff(username string, d1, d2 time.Time) (list []int64, err error) {

	if err := initDB(); err != nil {
		return nil, err
	}

	// select all deltas where event has not been captured already
	res, err := db.Query(`SELECT follower_id FROM followers
		WHERE username = ? AND on_day = ? AND follower_id NOT IN (
		SELECT follower_id FROM followers WHERE username = ? AND on_day = ?)
		AND follower_id NOT IN (SELECT follower_id FROM follower_events
			WHERE username = ? AND on_day = ?)`,
		username, d1.Format(isoDateFormat),
		username, d2.Format(isoDateFormat),
		username, time.Now().Format(isoDateFormat))
	if err != nil {
		return nil, errors.Wrap(err, "Error executing statement")
	}

	ids := []int64{}
	for res.Next() {
		var id int64
		err := res.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing results")
		}
		ids = append(ids, id)
	}

	logger.Printf("Found %d records for %s", len(ids), username)

	return ids, nil

}

// GetNewFollowerIDs users who started following since yesterday
func GetNewFollowerIDs(username string) (list []int64, err error) {
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayFollowers, err := getFollowerDailyCount(username, yesterday)
	if err != nil {
		return nil, err
	}
	if yesterdayFollowers == 0 {
		return []int64{}, nil
	}
	return getFollowerDiff(username, time.Now(), yesterday)
}

// GetStopFollowerIDs users who stopped following since yesterday
func GetStopFollowerIDs(username string) (list []int64, err error) {
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayFollowers, err := getFollowerDailyCount(username, yesterday)
	if err != nil {
		return nil, err
	}
	if yesterdayFollowers == 0 {
		return []int64{}, nil
	}
	return getFollowerDiff(username, time.Now().AddDate(0, 0, -1), time.Now())
}

// SaveDailyFollowers in single statement saves all followers for this day
func SaveDailyFollowers(username string, followerIDs []int64) error {

	if len(followerIDs) == 0 {
		return nil
	}

	if err := initDB(); err != nil {
		return err
	}

	sqlStr := "INSERT INTO followers (username, on_day, follower_id) VALUES "
	prms := []string{}
	vals := []interface{}{}

	day := time.Now()
	for _, id := range followerIDs {
		prms = append(prms, "(?, ?, ?)")
		vals = append(vals, username, day, id)
	}

	prmStr := strings.Join(prms, ",")
	sqlStr = sqlStr + prmStr + " ON DUPLICATE KEY UPDATE username = username"

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrap(err, "Error preparing bulk save followers statement")
	}

	res, err := stmt.Exec(vals...)
	if err != nil {
		return errors.Wrap(err, "Error executing save followers")
	}

	err = stmt.Close()
	if err != nil {
		return errors.Wrap(err, "Error closing save followers statement")
	}

	rowCount, _ := res.RowsAffected()
	logger.Printf("Saved %d from %d records for %s",
		rowCount, len(followerIDs), username)

	return nil

}

// SaveFollowerEvents saves follower events for given account
func SaveFollowerEvents(username, eventType string, followerIDs []int64) error {

	if len(followerIDs) == 0 {
		return nil
	}

	if err := initDB(); err != nil {
		return err
	}

	sqlStr := `INSERT INTO follower_events
		(username, on_day, follower_id, event_type) VALUES `
	prms := []string{}
	vals := []interface{}{}

	day := time.Now()
	for _, id := range followerIDs {
		prms = append(prms, "(?, ?, ?, ?)")
		vals = append(vals, username, day, id, eventType)
	}

	prmStr := strings.Join(prms, ",")
	sqlStr = sqlStr + prmStr + " ON DUPLICATE KEY UPDATE username = username"

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrap(err, "Error preparing bulk save follower event statement")
	}

	res, err := stmt.Exec(vals...)
	if err != nil {
		return errors.Wrap(err, "Error executing save follower events")
	}

	err = stmt.Close()
	if err != nil {
		return errors.Wrap(err, "Error closing save follower events")
	}

	rowCount, _ := res.RowsAffected()
	logger.Printf("Saved %d events from %d records for %s",
		rowCount, len(followerIDs), username)

	return nil

}
