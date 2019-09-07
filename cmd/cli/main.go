package main

import (
	"flag"
	"log"
	"os"

	"github.com/mchmarny/tweethingz/pkg/worker"
)

var (
	logger = log.New(os.Stdout, "server - ", 0)
)

func main() {
	logger.Println("Starting...")

	usr := flag.String("u", "", "Twitter username")
	flag.Parse()

	if *usr != "" {
		if err := worker.ProcessFollowers(*usr); err != nil {
			logger.Fatalf("Error: %v", err)
		}
	} else {
		if err := worker.BackfillFollowers(); err != nil {
			logger.Fatalf("Error: %v", err)
		}
	}

	logger.Println("Done")
}
