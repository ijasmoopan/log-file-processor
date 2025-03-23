package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ijasmoopan/intucloud-task/backend-service/config"
	"github.com/redis/go-redis/v9"
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

func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      redisClient,
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
	log.Printf("Subscribing to Redis channel: %s", cfg.ProgressChannel)
	ctx := context.Background()
	pubsub := m.redis.Subscribe(ctx, cfg.ProgressChannel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Redis subscription error: %v", err)
			return
		}

		log.Printf("Received message from Redis: %s", msg.Payload)

		var progressMsg ProgressMessage
		if err := json.Unmarshal([]byte(msg.Payload), &progressMsg); err != nil {
			log.Printf("Error unmarshaling progress message: %v", err)
			continue
		}

		log.Printf("*** Progress update for client %s: File: %s, Progress: %d%%, Status: %s",
			clientID, progressMsg.FileName, progressMsg.Progress, progressMsg.Status)

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
	}
}
