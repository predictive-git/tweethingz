package main

import (
	"log"
	"os"

	"github.com/mchmarny/twitterd/pkg/worker"
)

var (
	logger = log.New(os.Stdout, "server - ", 0)
)

func main() {
	logger.Println("Starting...")

	if err := worker.BackfillFollowers(); err != nil {
		logger.Fatalf("Error: %v", err)
	}

	logger.Println("Done")
}
