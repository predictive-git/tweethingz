package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// DataConfig defines the active data configuration
type DataConfig struct {
	CommonConfig
	DSN string `envconfig:"dsn" required:"true"`
}

// GetDataConfig loads db configs
func GetDataConfig() (cfg *DataConfig, err error) {
	var c DataConfig
	if e := envconfig.Process(appConfigPrefix, &c); e != nil {
		return nil, errors.Wrap(e, "Error parsing data config")
	}
	return &c, nil
}
