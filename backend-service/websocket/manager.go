package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
	"github.com/ijasmoopan/intucloud-task/backend-service/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
	mu   sync.Mutex
}

type Manager struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	redis      *redis.Client
	mu         sync.RWMutex
	db         *gorm.DB
}

// ProgressMessage represents the progress update from the log processor
type ProgressMessage struct {
	ClientID    string    `json:"client_id"`
	FileName    string    `json:"file_name"`
	Progress    int       `json:"progress"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
}

type ResultMessage struct {
	ClientID   string `json:"client_id,omitempty"`
	FilePath   string `json:"file_path"`
	ErrorCount int    `json:"error_count"`
	WarnCount  int    `json:"warn_count"`
}

func NewManager(redisClient *redis.Client, db *gorm.DB) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      redisClient,
		db:         db,
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()

		case client := <-m.unregister:
			if _, ok := m.clients[client.ID]; ok {
				m.mu.Lock()
				delete(m.clients, client.ID)
				m.mu.Unlock()
				close(client.Send)
			}
			client.Conn.Close()

		case message := <-m.broadcast:
			m.mu.RLock()
			for _, client := range m.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.clients, client.ID)
					client.Conn.Close()
				}
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) HandleWebSocket(c *gin.Context) {
	cfg := config.NewConfig()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	clientID := c.Query("client_id")
	if clientID == "" {
		conn.Close()
		return
	}

	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	m.register <- client

	// Start goroutines for reading and writing
	go m.writePump(client)
	go m.readPump(client)

	// Subscribe to Redis channel for this client
	go m.subscribeToRedis(cfg, clientID)
}

func (m *Manager) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
		m.unregister <- client
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.mu.Lock()
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				client.mu.Unlock()
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				client.mu.Unlock()
				return
			}
			w.Write(message)

			n := len(client.Send)
			for range n {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				client.mu.Unlock()
				return
			}
			client.mu.Unlock()
		}
	}
}

func (m *Manager) readPump(client *Client) {
	defer func() {
		m.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Handle incoming messages if needed
		log.Printf("Received message from client %s: %s", client.ID, message)
	}
}

func (m *Manager) subscribeToRedis(cfg *config.Config, clientID string) {
	ctx := context.Background()
	pubsub := m.redis.Subscribe(ctx, cfg.ProgressChannel, cfg.ResultChannel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Redis subscription error: %v", err)
			return
		}

		if msg.Channel == cfg.ProgressChannel {
			var progressMsg ProgressMessage
			if err := json.Unmarshal([]byte(msg.Payload), &progressMsg); err != nil {
				log.Printf("Error unmarshaling progress message: %v", err)
				continue
			}

			// Only broadcast to the specific client
			if progressMsg.ClientID == clientID {
				// Log the progress update
				log.Printf("Progress update for client %s: File: %s, Progress: %d%%, Status: %s",
					clientID, progressMsg.FileName, progressMsg.Progress, progressMsg.Status)

				if progressMsg.Error != "" {
					log.Printf("Error in progress update: %s", progressMsg.Error)
				}

				// Forward the message to the client
				m.broadcast <- []byte(msg.Payload)
			}

		} else if msg.Channel == cfg.ResultChannel {
			var resultMsg ResultMessage
			if err := json.Unmarshal([]byte(msg.Payload), &resultMsg); err != nil {
				log.Printf("Error unmarshaling result message: %v", err)
				fileResult := models.FileResult{
					FileName: "unknown", // We don't know the filename as unmarshal failed
					ClientID: clientID,
					Status:   "failed",
					Error:    fmt.Sprintf("Failed to parse result message: %v. Raw message: %s", err, msg.Payload),
				}
				if err := m.db.Create(&fileResult).Error; err != nil {
					log.Printf("Error storing failed result in database: %v", err)
				}
				continue
			}

			fileResult := models.FileResult{
				FileName:   filepath.Base(resultMsg.FilePath),
				ClientID:   clientID,
				Status:     "completed",
				ErrorCount: &resultMsg.ErrorCount,
				WarnCount:  &resultMsg.WarnCount,
			}

			if resultMsg.ClientID != clientID {
				fileResult.Status = "failed"
				fileResult.Error = fmt.Sprintf("Result message client ID %s does not match current client ID %s", resultMsg.ClientID, clientID)
			} else {
				log.Printf("Result update for client %s: File: %s, Error Count: %d, Warn Count: %d",
					clientID, resultMsg.FilePath, resultMsg.ErrorCount, resultMsg.WarnCount)
			}

			if err := m.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "file_name"}},
				UpdateAll: true,
			}).Create(&fileResult).Error; err != nil {
				log.Printf("Error storing failed result in database: %v", err)
			}
		}
	}
}
