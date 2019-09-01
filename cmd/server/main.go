package main

import (
	"log"
	"os"

	"github.com/mchmarny/twitterd/pkg/worker"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

func main() {

	logger.Println("Starting...")

	list, err := worker.GetFollowers()
	if err != nil {
		logger.Fatalf("Error getting followers: %v", err)
	}

	logger.Printf("Followers: %d", len(list))

	logger.Println("Done")
}
