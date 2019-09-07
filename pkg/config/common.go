package config

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stdout, "config: ", 0)
)

const (
	appConfigPrefix = "TW"
)

// CommonConfig represents common configuration options
type CommonConfig struct {
	Debug bool `envconfig:"debug"`
}
