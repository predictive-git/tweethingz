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

	worker.GetFollowers()

	logger.Println("Done")
}
