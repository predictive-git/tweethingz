package data

import (
	"log"
	"os"

	"database/sql"

	"github.com/mchmarny/twitterd/pkg/config"
	"github.com/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
)

const (
	isoDateFormat = "2006-01-02"
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

// Finalize cleans up all DB resources
func Finalize() {
	if db != nil {
		db.Close()
	}
}
