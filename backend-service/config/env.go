package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads the appropriate environment file based on APP_ENV
func LoadEnv() {
	// Get the environment from APP_ENV, default to "local"
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	// Load the appropriate .env file based on environment
	var envFile string
	switch env {
	case "prod":
		envFile = ".env.prod"
	case "local":
		envFile = ".env.local"
	default:
		envFile = ".env.local"
	}

	// Try to load the environment file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Error loading %s file: %v", envFile, err)
	}

	// Log which environment is being used
	log.Printf("Running in %s environment", env)
}
