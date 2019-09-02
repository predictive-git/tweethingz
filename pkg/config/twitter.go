package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	twitterConfigPrefix = "T"
)

// TwitterConfig defines the active twiter configuration
type TwitterConfig struct {
	ConsumerKey    string `envconfig:"consumer_key" required:"true"`
	ConsumerSecret string `envconfig:"consumer_secret" required:"true"`
	AccessToken    string `envconfig:"access_token" required:"true"`
	AccessSecret   string `envconfig:"access_secret" required:"true"`
}

// GetTwitterConfig loads twitter configs
func GetTwitterConfig() (cfg *TwitterConfig, err error) {
	var c TwitterConfig
	if e := envconfig.Process(twitterConfigPrefix, &c); e != nil {
		return nil, errors.Wrap(e, "Error parsing twitter config")
	}
	return &c, nil
}