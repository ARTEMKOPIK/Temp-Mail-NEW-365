package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artemkopik/tempmail/internal/api"
	"github.com/artemkopik/tempmail/internal/config"
	"github.com/artemkopik/tempmail/internal/crypto"
	"github.com/artemkopik/tempmail/internal/database"
	"github.com/artemkopik/tempmail/internal/repository"
	"github.com/artemkopik/tempmail/internal/smtp"
	"github.com/artemkopik/tempmail/internal/websocket"
)

func main() {
	log.Println("🚀 Starting Temp Mail 365 Service...")

	// Load configuration
	cfg := config.LoadConfig()
	log.Println("✅ Configuration loaded successfully")

	// Initialize encryption service
	encryptionService, err := crypto.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("❌ Failed to initialize encryption service: %v", err)
	}
	log.Println("✅ Encryption service initialized")

	// Connect to MongoDB
	mongoClient, err := database.NewMongoClient(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	log.Println("✅ Connected to MongoDB")

	// Initialize MongoDB indexes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := database.InitializeMongoIndexes(ctx, mongoClient, cfg.MongoDB); err != nil {
		log.Printf("⚠️  Warning: Failed to initialize MongoDB indexes: %v", err)
	} else {
		log.Println("✅ MongoDB indexes initialized")
	}
	cancel()

	// Connect to Redis
	redisClient, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// Initialize repositories
	emailRepo := repository.NewEmailRepository(
		mongoClient.Database(cfg.MongoDB),
		encryptionService,
	)
	mailboxRepo := repository.NewMailboxRepository(redisClient, cfg.EmailDomain)
	log.Println("✅ Repositories initialized")

	// Initialize WebSocket hub
	hub := websocket.NewHub(redisClient)
	go hub.Run()
	log.Println("✅ WebSocket hub started")

	// Initialize SMTP server
	smtpServer := smtp.NewServer(cfg, mailboxRepo, emailRepo, hub)
	go func() {
		log.Printf("📧 SMTP server starting on %s:%d", cfg.SMTPHost, cfg.SMTPPort)
		if err := smtpServer.ListenAndServe(); err != nil {
			log.Printf("❌ SMTP server error: %v", err)
		}
	}()

	// Initialize HTTP router
	router := api.NewRouter(cfg, emailRepo, mailboxRepo, hub)
	
	// Start HTTP server in goroutine
	go func() {
		log.Printf("🌐 API server starting on port %d", cfg.APIPort)
		if err := router.Run(cfg.APIPort); err != nil {
			log.Fatalf("❌ Failed to start API server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down gracefully...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Close SMTP server
	if err := smtpServer.Close(); err != nil {
		log.Printf("⚠️  Error closing SMTP server: %v", err)
	} else {
		log.Println("✅ SMTP server closed")
	}

	// Shutdown WebSocket hub
	hub.Shutdown()
	log.Println("✅ WebSocket hub closed")

	// Disconnect from MongoDB
	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		log.Printf("⚠️  Error disconnecting from MongoDB: %v", err)
	} else {
		log.Println("✅ MongoDB disconnected")
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("⚠️  Error closing Redis connection: %v", err)
	} else {
		log.Println("✅ Redis connection closed")
	}

	log.Println("👋 Temp Mail 365 Service stopped")
}

