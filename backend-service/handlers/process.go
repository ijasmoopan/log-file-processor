package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
	"github.com/ijasmoopan/intucloud-task/backend-service/redis"
)

type ProcessRequest struct {
	FileNames []string `json:"file_names" binding:"required"`
}

type ProcessResult struct {
	FileName    string    `json:"file_name"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
	Error       string    `json:"error,omitempty"`
}

type ProcessResponse struct {
	Message string `json:"message"`
}

func ProcessFiles(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			return
		}

		var req ProcessRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		validFiles := make([]string, 0)
		for _, fileName := range req.FileNames {
			if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
				continue
			}
			filePath := filepath.Join(cfg.UploadDir, fileName)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				continue
			}
			validFiles = append(validFiles, fileName)
		}

		if len(validFiles) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid files found for processing"})
			return
		}

		// Initialize Redis client
		redisClient, err := redis.NewClient(cfg.RedisAddress)
		if err != nil {
			log.Printf("Failed to connect to Redis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to connect to Redis service",
			})
			return
		}
		defer redisClient.Close()

		// Generate a unique client ID for this request
		clientID := time.Now().Format("20060102150405")

		// Create a message with the valid file names and client ID
		message := struct {
			Files    []string `json:"file_names"`
			ClientID string   `json:"client_id"`
		}{
			Files:    validFiles,
			ClientID: clientID,
		}

		// Publish the message to Redis
		if err := redisClient.Publish(cfg.ProcessingChannel, message); err != nil {
			log.Printf("Failed to publish message to Redis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to publish processing request",
			})
			return
		}

		// Return immediate response
		c.JSON(http.StatusAccepted, gin.H{
			"message":   "Processing request accepted",
			"client_id": clientID,
		})
	}
}
