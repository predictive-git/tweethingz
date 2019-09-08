package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// WorkerConfig defines the active twiter configuration
type WorkerConfig struct {
	CommonConfig
	ConcurentRefreshLimit int `envconfig:"concurent_refresh_limit" required:"true"`
}

// GetWorkerConfig loads twitter configs
func GetWorkerConfig() (cfg *WorkerConfig, err error) {
	var c WorkerConfig
	if e := envconfig.Process(appConfigPrefix, &c); e != nil {
		return nil, errors.Wrap(e, "Error parsing worker config")
	}
	return &c, nil
}
