package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
	"github.com/ijasmoopan/intucloud-task/backend-service/handlers"
	"github.com/ijasmoopan/intucloud-task/backend-service/middleware"
	"github.com/ijasmoopan/intucloud-task/backend-service/websocket"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})

	// Initialize WebSocket manager
	wsManager := websocket.NewManager(redisClient)
	go wsManager.Run()

	// Set up Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Create upload directory if it doesn't exist
	if err := config.EnsureUploadDir(cfg.UploadDir); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Routes
	api := router.Group("/api/v1")
	{
		api.POST("/upload", middleware.ValidateFiles(), handlers.UploadFile(cfg))
		api.GET("/files", handlers.ListFiles(cfg))
		api.POST("/process", handlers.ProcessFiles(cfg))
		api.GET("/ws", wsManager.HandleWebSocket)
	}

	// Start server
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
