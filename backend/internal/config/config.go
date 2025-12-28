package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	BackendPort int
	SMTPPort    int
	SMTPHost    string

	// Database configuration
	MongoURL      string
	MongoDatabase string
	RedisURL      string
	RedisPassword string

	// Email configuration
	EmailDomain       string
	EmailExpiryMinutes int

	// Security
	EncryptionKey string
	JWTSecret     string

	// CORS
	FrontendURL string

	// WebSocket configuration
	WSPingInterval   time.Duration
	WSMaxConnections int

	// Rate limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	cfg := &Config{
		BackendPort:       getEnvAsInt("BACKEND_PORT", 8080),
		SMTPPort:          getEnvAsInt("SMTP_PORT", 2525),
		SMTPHost:          getEnv("SMTP_HOST", "0.0.0.0"),
		MongoURL:          getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:     getEnv("MONGODB_DATABASE", "tempmail"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		EmailDomain:       getEnv("EMAIL_DOMAIN", "your-tempmail.info"),
		EmailExpiryMinutes: getEnvAsInt("EMAIL_EXPIRY_MINUTES", 60),
		EncryptionKey:     getEnv("ENCRYPTION_KEY", "change-this-32-byte-key-in-prod"),
		JWTSecret:         getEnv("JWT_SECRET", "change-this-jwt-secret"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		WSPingInterval:    getEnvAsDuration("WS_PING_INTERVAL", 30*time.Second),
		WSMaxConnections:  getEnvAsInt("WS_MAX_CONNECTIONS", 1000),
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute),
	}

	// Validate critical configuration
	cfg.validate()

	log.Printf("Configuration loaded successfully")
	log.Printf("Backend Port: %d, SMTP Port: %d", cfg.BackendPort, cfg.SMTPPort)
	log.Printf("Email Domain: %s", cfg.EmailDomain)

	return cfg
}

// validate checks that all critical configuration values are set
func (c *Config) validate() {
	if c.EncryptionKey == "" || c.EncryptionKey == "change-this-32-byte-key-in-prod" {
		log.Println("WARNING: Using default encryption key. Change ENCRYPTION_KEY in production!")
	}

	if len(c.EncryptionKey) < 32 {
		log.Fatal("ENCRYPTION_KEY must be at least 32 bytes long")
	}

	if c.EmailDomain == "" {
		log.Fatal("EMAIL_DOMAIN must be set")
	}

	if c.MongoURL == "" {
		log.Fatal("MONGODB_URI must be set")
	}

	if c.RedisURL == "" {
		log.Fatal("REDIS_URL must be set")
	}
}

// Helper functions to read environment variables

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default %d", key, defaultValue)
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid duration value for %s, using default %v", key, defaultValue)
		return defaultValue
	}

	return value
}

