package config

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	twitterConfigPrefix = "T"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// TwitterConfig defines the active twiter configuration
type TwitterConfig struct {
	Username       string `envconfig:"username" required:"true"`
	ConsumerKey    string `envconfig:"consumer_key" required:"true"`
	ConsumerSecret string `envconfig:"consumer_secret" required:"true"`
	AccessToken    string `envconfig:"access_token" required:"true"`
	AccessSecret   string `envconfig:"access_secret" required:"true"`
}

// GetTwitterConfig reads twitter config from env vars
func GetTwitterConfig() (cfg *TwitterConfig, err error) {
	var c TwitterConfig
	if e := envconfig.Process(twitterConfigPrefix, &c); e != nil {
		return nil, errors.Wrapf(e, "Error parsing twitter config")
	}
	return &c, nil
}
