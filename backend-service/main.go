package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
	"github.com/ijasmoopan/intucloud-task/backend-service/handlers"
	"github.com/ijasmoopan/intucloud-task/backend-service/middleware"
	"github.com/ijasmoopan/intucloud-task/backend-service/websocket"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})

	// Initialize WebSocket manager
	wsManager := websocket.NewManager(redisClient, db)
	go wsManager.Run()

	// Set up Gin router
	router := gin.Default()

	// Configure CORS based on environment
	corsConfig := cors.DefaultConfig()
	if os.Getenv("APP_ENV") == "prod" {
		// Production CORS settings
		corsConfig.AllowOrigins = []string{"http://15.206.174.223:3000"}
	} else {
		// Development CORS settings
		corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * 60 * 60 // 12 hours

	router.Use(cors.New(corsConfig))

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Create upload directory if it doesn't exist
	if err := config.EnsureUploadDir(cfg.UploadDir); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	// Routes
	api := router.Group("/api/v1")
	{
		api.POST("/upload", middleware.ValidateFiles(), handlers.UploadFile(cfg))
		api.GET("/files", handlers.ListFiles(cfg))
		api.POST("/process", handlers.ProcessFiles(cfg))
		api.GET("/ws", wsManager.HandleWebSocket)
		api.GET("/results", handlers.GetResults(db))
		api.GET("/results/:id", handlers.GetResultByID(db))
		api.GET("/results/filename/:filename", handlers.GetResultByFilename(db))
	}

	// Start server
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
