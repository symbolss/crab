package main

import (
	"log"

	"github.com/family/crab-server/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting worker (environment: %s)", cfg.Environment)

	// TODO: Implement worker logic
	// - Process background jobs
	// - Consume from Redis queues
	// - Handle scheduled tasks

	select {}
}
