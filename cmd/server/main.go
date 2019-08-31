package main

import (
	"log"
	"os"

	"github.com/mchmarny/twitterd/pkg/config"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

func main() {

	logger.Println("Initializing configuration...")

	// read config
	cfg, err := config.Read()
	if err != nil {
		logger.Fatalf("Error reading config: %v", err)
	}
	logger.Printf("Config: %+v", cfg)

	logger.Println("Done")
}
