package api

import (
	"fmt"
	"time"

	"github.com/artemkopik/tempmail/internal/api/handlers"
	"github.com/artemkopik/tempmail/internal/api/middleware"
	"github.com/artemkopik/tempmail/internal/config"
	"github.com/artemkopik/tempmail/internal/repository"
	"github.com/artemkopik/tempmail/internal/websocket"
	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures the Gin router
func NewRouter(cfg *config.Config, emailRepo *repository.EmailRepository, 
	mailboxRepo *repository.MailboxRepository, hub *websocket.Hub) *gin.Engine {
	
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.FrontendURL))

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	router.Use(rateLimiter.Middleware())

	// Initialize handlers
	mailboxHandler := handlers.NewMailboxHandler(mailboxRepo, emailRepo)
	emailHandler := handlers.NewEmailHandler(emailRepo)
	wsHandler := handlers.NewWebSocketHandler(hub, mailboxRepo)

	// API routes
	api := router.Group("/api")
	{
		// Mailbox routes
		api.POST("/mailbox", mailboxHandler.CreateMailbox)
		api.GET("/mailbox/:address", mailboxHandler.GetMailbox)
		api.DELETE("/mailbox/:address", mailboxHandler.DeleteMailbox)
		api.GET("/mailbox/:address/emails", mailboxHandler.GetEmails)

		// Email routes
		api.GET("/email/:id", emailHandler.GetEmail)
		api.DELETE("/email/:id", emailHandler.DeleteEmail)

		// WebSocket route
		api.GET("/ws/mailbox/:address", wsHandler.HandleConnection)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	return router
}

// Run starts the HTTP server
func (router *gin.Engine) Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return router.Run(addr)
}

