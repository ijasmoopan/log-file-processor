package config

import (
	"os"
)

type Config struct {
	RedisAddress      string
	ProcessingChannel string
	ProgressChannel   string
	NumWorkers        int
	UploadDir         string
}

func NewConfig() *Config {
	return &Config{
		RedisAddress:      getEnvOrDefault("REDIS_ADDRESS", "localhost:6379"),
		ProcessingChannel: getEnvOrDefault("PROCESSING_CHANNEL", "processing_channel"),
		ProgressChannel:   getEnvOrDefault("PROGRESS_CHANNEL", "progress_channel"),
		NumWorkers:        4,
		UploadDir:         getEnvOrDefault("UPLOAD_DIR", "../uploads"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
