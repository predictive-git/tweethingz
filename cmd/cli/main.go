package main

import (
	"log"
	"os"

	"github.com/mchmarny/tweethingz/pkg/worker"
)

var (
	logger = log.New(os.Stdout, "cli - ", 0)
)

func main() {
	logger.Println("Starting...")

	if err := worker.Run(); err != nil {
		logger.Fatalf("Error: %v", err)
	}

	logger.Println("Done")
}
