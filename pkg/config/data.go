package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	dbConfigPrefix = "DB"
)

// DataConfig defines the active data configuration
type DataConfig struct {
	DSN string `envconfig:"dsn" required:"true"`
}

// GetDataConfig loads db configs
func GetDataConfig() (cfg *DataConfig, err error) {
	var c DataConfig
	if e := envconfig.Process(dbConfigPrefix, &c); e != nil {
		return nil, errors.Wrap(e, "Error parsing data config")
	}
	return &c, nil
}
