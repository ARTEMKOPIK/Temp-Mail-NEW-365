package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/artemkopik/tempmail/internal/repository"
	"github.com/artemkopik/tempmail/internal/websocket"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (configure properly in production)
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *websocket.Hub
	mailboxRepo *repository.MailboxRepository
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(hub *websocket.Hub, mailboxRepo *repository.MailboxRepository) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		mailboxRepo: mailboxRepo,
	}
}

// HandleConnection upgrades HTTP connection to WebSocket for a mailbox
// GET /api/ws/mailbox/:address
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address is required",
		})
		return
	}

	// Verify mailbox exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := h.mailboxRepo.Exists(ctx, address)
	if err != nil {
		log.Printf("Error checking mailbox: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to verify mailbox",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "mailbox not found or expired",
		})
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := websocket.NewClient(h.hub, conn, address)

	// Register client
	h.hub.register <- client

	// Start client goroutines
	client.Start()

	log.Printf("WebSocket connection established for %s", address)
}

