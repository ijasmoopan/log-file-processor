package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Message struct {
	Payload string
}

type ProcessingMessage struct {
	FileNames []string `json:"file_names"`
	ClientID  string   `json:"client_id"`
}

type ProgressMessage struct {
	ClientID    string    `json:"client_id"`
	FileName    string    `json:"file_name"`
	Progress    int       `json:"progress"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Client struct {
	client *redis.Client
}

func NewClient(addr string) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create a context with timeout for connection check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", addr, err)
	}

	fmt.Printf("Successfully connected to Redis at %s\n", addr)
	return &Client{client: client}, nil
}

func (c *Client) Subscribe(channel string) (*redis.PubSub, error) {
	ctx := context.Background()
	return c.client.Subscribe(ctx, channel), nil
}

func (c *Client) Publish(channel string, message interface{}) error {
	ctx := context.Background()

	// Convert message to JSON if it's not already a string
	var messageStr string
	switch v := message.(type) {
	case string:
		messageStr = v
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %v", err)
		}
		messageStr = string(jsonBytes)
	}

	return c.client.Publish(ctx, channel, messageStr).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}
