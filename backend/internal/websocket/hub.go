package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/artemkopik/tempmail/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	newEmailChannel = "new_email"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients mapped by mailbox address
	clients map[string]map[*Client]bool

	// Inbound messages from clients
	broadcast chan *BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Redis client for pub/sub
	redisClient *redis.Client

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Context for graceful shutdown
	ctx context.Context
	
	// Cancel function
	cancel context.CancelFunc
}

// BroadcastMessage represents a message to broadcast to specific mailbox
type BroadcastMessage struct {
	Address string
	Message models.WebSocketMessage
}

// NewHub creates a new Hub instance
func NewHub(redisClient *redis.Client) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Hub{
		clients:     make(map[string]map[*Client]bool),
		broadcast:   make(chan *BroadcastMessage, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	// Start Redis pub/sub listener
	go h.listenRedis()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAddress(message)

		case <-h.ctx.Done():
			log.Println("Hub shutting down...")
			h.cleanup()
			return
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.address] == nil {
		h.clients[client.address] = make(map[*Client]bool)
	}

	h.clients[client.address][client] = true
	
	log.Printf("WebSocket client registered: %s (total for address: %d)", 
		client.address, len(h.clients[client.address]))
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.address]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)

			// Remove empty address entry
			if len(clients) == 0 {
				delete(h.clients, client.address)
			}

			log.Printf("WebSocket client unregistered: %s (remaining for address: %d)", 
				client.address, len(h.clients[client.address]))
		}
	}
}

// broadcastToAddress sends a message to all clients subscribed to an address
func (h *Hub) broadcastToAddress(msg *BroadcastMessage) {
	h.mu.RLock()
	clients := h.clients[msg.Address]
	h.mu.RUnlock()

	if clients == nil {
		log.Printf("No clients subscribed to address: %s", msg.Address)
		return
	}

	// Marshal message to JSON
	messageJSON, err := json.Marshal(msg.Message)
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		return
	}

	// Send to all clients for this address
	successCount := 0
	for client := range clients {
		select {
		case client.send <- messageJSON:
			successCount++
		default:
			// Client's send channel is full, close it
			log.Printf("Client send buffer full, closing connection for %s", client.address)
			close(client.send)
			h.mu.Lock()
			delete(h.clients[client.address], client)
			h.mu.Unlock()
		}
	}

	log.Printf("Broadcast message to %d clients for address %s", successCount, msg.Address)
}

// BroadcastNewEmail broadcasts a new email notification
func (h *Hub) BroadcastNewEmail(address string, email *models.NewEmailPayload) {
	message := models.WebSocketMessage{
		Type:    "new_email",
		Payload: email,
	}

	h.broadcast <- &BroadcastMessage{
		Address: address,
		Message: message,
	}

	// Also publish to Redis for distributed scenarios
	h.publishToRedis(address, message)
}

// BroadcastError broadcasts an error message
func (h *Hub) BroadcastError(address string, errorMsg string, code string) {
	message := models.WebSocketMessage{
		Type: "error",
		Payload: models.ErrorPayload{
			Message: errorMsg,
			Code:    code,
		},
	}

	h.broadcast <- &BroadcastMessage{
		Address: address,
		Message: message,
	}
}

// publishToRedis publishes a message to Redis pub/sub
func (h *Hub) publishToRedis(address string, message models.WebSocketMessage) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message for Redis: %v", err)
		return
	}

	channel := newEmailChannel + ":" + address
	
	if err := h.redisClient.Publish(h.ctx, channel, messageJSON).Err(); err != nil {
		log.Printf("Error publishing to Redis: %v", err)
	}
}

// listenRedis listens for messages from Redis pub/sub
func (h *Hub) listenRedis() {
	pubsub := h.redisClient.PSubscribe(h.ctx, newEmailChannel+":*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Println("Started listening to Redis pub/sub")

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Extract address from channel name
			address := msg.Channel[len(newEmailChannel)+1:]

			// Parse message
			var wsMessage models.WebSocketMessage
			if err := json.Unmarshal([]byte(msg.Payload), &wsMessage); err != nil {
				log.Printf("Error parsing Redis message: %v", err)
				continue
			}

			// Broadcast to local clients
			h.broadcast <- &BroadcastMessage{
				Address: address,
				Message: wsMessage,
			}

		case <-h.ctx.Done():
			log.Println("Redis listener shutting down...")
			return
		}
	}
}

// GetClientCount returns the number of connected clients for an address
func (h *Hub) GetClientCount(address string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[address]; ok {
		return len(clients)
	}

	return 0
}

// GetTotalClientCount returns the total number of connected clients
func (h *Hub) GetTotalClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}

	return total
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients":    h.GetTotalClientCount(),
		"active_addresses": len(h.clients),
	}
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	log.Println("Shutting down WebSocket hub...")
	h.cancel()
}

// cleanup closes all client connections
func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for address, clients := range h.clients {
		for client := range clients {
			close(client.send)
			log.Printf("Closed client connection for %s", address)
		}
	}

	h.clients = make(map[string]map[*Client]bool)
	log.Println("Hub cleanup completed")
}

