package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/config"
	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/models"
	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/processor"
	"github.com/ijasmoopan/intucloud-task/log-processor-service/internal/redis"
)

type Server struct {
	config    *config.Config
	redis     *redis.Client
	processor *processor.Processor
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize Redis client
	redisClient, err := redis.NewClient(cfg.RedisAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		config:    cfg,
		redis:     redisClient,
		processor: processor.NewProcessor(cfg.NumWorkers, cfg.UploadDir),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (s *Server) Start() error {
	// Subscribe to the processing channel
	pubsub, err := s.redis.Subscribe(s.config.ProcessingChannel)
	if err != nil {
		return fmt.Errorf("failed to subscribe to Redis channel: %v", err)
	}
	defer pubsub.Close()

	// Create a channel to handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle shutdown signals
	go func() {
		<-sigChan
		log.Println("\nReceived shutdown signal. Cleaning up...")
		s.cancel()
	}()

	log.Println("Log processor started. Waiting for messages...")

	// Process messages from Redis
	for {
		select {
		case <-s.ctx.Done():
			log.Println("Shutting down...")
			return nil
		default:
			msg, err := pubsub.ReceiveMessage(s.ctx)
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				continue
			}

			if err := s.handleMessage(&redis.Message{Payload: msg.Payload}); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}
}

func (s *Server) handleMessage(msg *redis.Message) error {
	var processingMsg redis.ProcessingMessage
	if err := json.Unmarshal([]byte(msg.Payload), &processingMsg); err != nil {
		return fmt.Errorf("error unmarshaling message: %v", err)
	}

	log.Printf("Received processing request for files: %v ~ %s", processingMsg.FileNames, processingMsg.ClientID)

	// Create progress callback function
	progressCb := func(fileName string, progress int, status string, err error) {
		progressMsg := redis.ProgressMessage{
			ClientID:    processingMsg.ClientID,
			FileName:    fileName,
			Progress:    progress,
			Status:      status,
			ProcessedAt: time.Now(),
		}

		if err != nil {
			progressMsg.Error = err.Error()
		}

		if err := s.redis.Publish(s.config.ProgressChannel, progressMsg); err != nil {
			log.Printf("Error publishing progress message: %v", err)
		}
		log.Printf("Progress update for client %s: File: %s, Progress: %d%%, Status: %s",
			processingMsg.ClientID, fileName, progress, status)
	}

	// Process the files with progress updates
	results, err := s.processor.ProcessFiles(processingMsg.FileNames, progressCb)
	if err != nil {
		return fmt.Errorf("error processing files: %v", err)
	}

	// Calculate totals
	totalResult := models.Result{FilePath: "Total"}
	for _, result := range results {
		totalResult.Add(result)
	}

	fmt.Printf("Total -> Error Count: %d, Warn Count: %d\n",
		totalResult.ErrorCount, totalResult.WarnCount)

	return nil
}

func (s *Server) Stop() {
	s.cancel()
	s.redis.Close()
}
