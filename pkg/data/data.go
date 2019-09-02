package data

import (
	"log"
	"os"
	"strings"
	"time"

	"database/sql"

	"github.com/mchmarny/twitterd/pkg/config"
	"github.com/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
)

const (
	isoDateFormat = "2006-01-02"
)

var (
	logger = log.New(os.Stdout, "data - ", 0)
)

// DB represents the application DB
type DB struct {
	conn *sql.DB
}

// GetDB creates initialized DB client
func GetDB() (db *DB, err error) {

	cfg, err := config.GetDataConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Twitter config")
	}

	c, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "Error connecting to DB")
	}

	if err := c.Ping(); err != nil {
		c.Close()
		return nil, errors.Wrap(err, "Error pinging DB")
	}

	d := &DB{
		conn: c,
	}

	return d, nil

}

// Finalize cleans up all DB resources
func (d *DB) Finalize() {
	if d.conn != nil {
		d.conn.Close()
	}
}

func (d *DB) getFollowerDiff(d1, d2 time.Time) (list []int64, err error) {

	stmt, err := d.conn.Prepare(`SELECT user_id FROM followers
		WHERE on_day = ? AND user_id NOT IN (SELECT user_id
		FROM followers WHERE on_day = ?)`)
	if err != nil {
		return nil, errors.Wrap(err, "Error on new followers prepare")
	}

	res, err := stmt.Query(d1.Format(isoDateFormat), d2.Format(isoDateFormat))
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

	logger.Printf("Found %d records", len(ids))

	return ids, nil

}

// GetNewFollowerIDs users who started following since yesterday
func (d *DB) GetNewFollowerIDs() (list []int64, err error) {
	return d.getFollowerDiff(time.Now(), time.Now().AddDate(0, 0, -1))
}

// GetStopFollowerIDs users who stopped following since yesterday
func (d *DB) GetStopFollowerIDs() (list []int64, err error) {
	return d.getFollowerDiff(time.Now().AddDate(0, 0, -1), time.Now())
}

// SaveDailyFollowers in single statement saves all followers for this day
func (d *DB) SaveDailyFollowers(list []int64) error {

	sqlStr := "INSERT INTO followers(on_day, user_id) VALUES "
	prms := []string{}
	vals := []interface{}{}

	day := time.Now()
	for _, id := range list {
		prms = append(prms, "(?, ?)")
		vals = append(vals, day, id)
	}

	prmStr := strings.Join(prms, ",")
	sqlStr = sqlStr + prmStr

	stmt, err := d.conn.Prepare(sqlStr)
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
	logger.Printf("Saved %d from %d records", rowCount, len(list))

	return nil

}
