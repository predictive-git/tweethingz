package main

import (
	"log"
	"os"

	"github.com/mchmarny/tweethingz/pkg/worker"
)

var (
	logger = log.New(os.Stdout, "service - ", 0)
)

func main() {
	logger.Println("Starting service...")

	if err := worker.Run(); err != nil {
		logger.Fatalf("Error: %v", err)
	}

	logger.Println("Done")

}
