package main

import (
	"log"
	"os"
	"os/signal"

	"syscall"

	"golang.org/x/net/context"

	"github.com/mchmarny/tweethingz/pkg/worker"
)

var (
	logger = log.New(os.Stdout, "cli - ", 0)
)

func main() {
	logger.Println("Starting service...")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Wait for SIGINT and SIGTERM (HIT CTRL-C)
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		cancel()
		os.Exit(0)
	}()

	if err := worker.Run(ctx); err != nil {
		logger.Fatalf("Error: %v", err)
	}

	logger.Println("Done")

}
