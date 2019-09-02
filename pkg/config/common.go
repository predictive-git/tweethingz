package config

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stdout, "config - ", 0)
)
