package config

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	appConfigName = "TWD"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// AppConfig defines the active configuration for this service
type AppConfig struct {
	// TwitterUser is the twitter username
	TwitterUser string `envconfig:"twitter_user" required:"true"`
}

// Read reads config from env vars
func Read() (cfg *AppConfig, err error) {
	var c AppConfig
	if e := envconfig.Process(appConfigName, &c); e != nil {
		return nil, errors.Wrapf(e, "Error parsing config")
	}
	return &c, nil
}
