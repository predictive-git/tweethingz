package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// OAuthConfig defines the active OAuth2 configuration
type OAuthConfig struct {
	CommonConfig
	OAuthClientID     string `envconfig:"oauth_client_id" required:"true"`
	OAuthClientSecret string `envconfig:"oauth_client_secret" required:"true"`
	ForceHTTPS        bool   `envconfig:"oauth_force_https"`
}

// GetOAuthConfig loads OAuth2 configs
func GetOAuthConfig() (cfg *OAuthConfig, err error) {
	var c OAuthConfig
	if e := envconfig.Process(appConfigPrefix, &c); e != nil {
		return nil, errors.Wrap(e, "Error parsing OAuth config")
	}
	return &c, nil
}
