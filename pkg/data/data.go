package data

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

func getFollowerDiff(username string, d1, d2 time.Time) (list []int64, err error) {

	if err := initDB(); err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(`SELECT follower_id FROM followers
		WHERE username = ? AND on_day = ? AND follower_id NOT IN (
		SELECT follower_id FROM followers WHERE username = ? AND on_day = ?)`)
	if err != nil {
		return nil, errors.Wrap(err, "Error on new followers prepare")
	}

	res, err := stmt.Query(username, d1.Format(isoDateFormat),
		username, d2.Format(isoDateFormat))
	if err != nil {
		return nil, errors.Wrap(err, "Error executing statement")
	}
	defer stmt.Close()

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
	return getFollowerDiff(username, time.Now(), time.Now().AddDate(0, 0, -1))
}

// GetStopFollowerIDs users who stopped following since yesterday
func GetStopFollowerIDs(username string) (list []int64, err error) {
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

	sqlStr := "INSERT INTO followers(username, on_day, follower_id) VALUES "
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
