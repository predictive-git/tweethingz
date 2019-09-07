package data

import (
	"log"
	"os"

	"database/sql"

	"github.com/mchmarny/tweethingz/pkg/config"
	"github.com/pkg/errors"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

const (
	isoDateFormat = "2006-01-02"

	// EventNewFollower event type
	EventNewFollower = "followed"
	// EventUnFollowing event type
	EventUnFollowing = "unfollowed"
)

var (
	logger = log.New(os.Stdout, "data: ", 0)
	db     *sql.DB
)

func initDB() error {

	if db != nil {
		return nil
	}

	cfg, err := config.GetDataConfig()
	if err != nil {
		return errors.Wrap(err, "Error getting data config")
	}

	d, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return errors.Wrap(err, "Error connecting to DB")
	}

	if err := d.Ping(); err != nil {
		d.Close()
		return errors.Wrap(err, "Error pinging DB")
	}

	db = d

	return nil

}
