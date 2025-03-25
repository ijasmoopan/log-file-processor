package config

import (
	"os"
)

type Config struct {
	ServerAddress     string
	UploadDir         string
	MaxFileSize       int64
	AllowedTypes      []string
	ProcessingService string
	RedisAddress      string
	ProcessingChannel string
	ProgressChannel   string
	ResultChannel     string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress:     getEnvOrDefault("SERVER_ADDRESS", ":8080"),
		UploadDir:         getEnvOrDefault("UPLOAD_DIR", "../uploads"),
		MaxFileSize:       500 * 1024 * 1024, // 500MB
		AllowedTypes:      []string{".log", ".txt", ".csv", ".json"},
		ProcessingService: getEnvOrDefault("PROCESSING_SERVICE_URL", "http://localhost:8081"),
		RedisAddress:      getEnvOrDefault("REDIS_ADDRESS", "localhost:6379"),
		ProcessingChannel: getEnvOrDefault("PROCESSING_CHANNEL", "processing_channel"),
		ProgressChannel:   getEnvOrDefault("PROGRESS_CHANNEL", "progress_channel"),
		ResultChannel:     getEnvOrDefault("RESULT_CHANNEL", "result_channel"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func EnsureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
