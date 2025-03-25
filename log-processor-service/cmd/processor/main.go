package main

import (
	"log"

	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/config"
	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/server"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Initialize configuration
	cfg := config.NewConfig()

	// Create and start server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
