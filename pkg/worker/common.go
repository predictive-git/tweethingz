package worker

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stdout, "worker: ", 0)
)
