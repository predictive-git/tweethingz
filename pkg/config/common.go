package config

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stdout, "config: ", 0)
)

const (
	appConfigPrefix = "T"
)

// CommonConfig represents common configuration options
type CommonConfig struct {
	Debug bool `envconfig:"debug"`
}
