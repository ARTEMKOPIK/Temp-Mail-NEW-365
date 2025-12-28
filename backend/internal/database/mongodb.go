package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// EmailsCollection is the name of the emails collection
	EmailsCollection = "emails"
	
	// DefaultMaxPoolSize is the default maximum number of connections in the pool
	DefaultMaxPoolSize = 100
	
	// DefaultMinPoolSize is the default minimum number of connections in the pool
	DefaultMinPoolSize = 10
	
	// DefaultMaxConnIdleTime is the default maximum time a connection can be idle
	DefaultMaxConnIdleTime = 30 * time.Minute
)

// NewMongoClient creates a new MongoDB client with connection pooling
func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	// Configure client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(DefaultMaxPoolSize).
		SetMinPoolSize(DefaultMinPoolSize).
		SetMaxConnIdleTime(DefaultMaxConnIdleTime).
		SetRetryWrites(true).
		SetRetryReads(true)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB")

	return client, nil
}

// InitializeIndexes creates necessary indexes for the database
func InitializeIndexes(ctx context.Context, db *mongo.Database) error {
	emailsCollection := db.Collection(EmailsCollection)

	// Create TTL index on expires_at field (automatic deletion)
	ttlIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "expires_at", Value: 1},
		},
		Options: options.Index().
			SetExpireAfterSeconds(0). // Delete immediately when expires_at is reached
			SetName("expires_at_ttl"),
	}

	// Create compound index on address and received_at for efficient querying
	addressIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "address", Value: 1},
			{Key: "received_at", Value: -1}, // Descending for latest first
		},
		Options: options.Index().
			SetName("address_received_at"),
	}

	// Create index on address only
	addressOnlyIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "address", Value: 1},
		},
		Options: options.Index().
			SetName("address"),
	}

	// Create indexes
	indexes := []mongo.IndexModel{ttlIndex, addressIndex, addressOnlyIndex}
	
	indexNames, err := emailsCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Printf("Created MongoDB indexes: %v", indexNames)

	return nil
}

// HealthCheck performs a health check on the MongoDB connection
func HealthCheck(ctx context.Context, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	return nil
}

// GetCollection returns a collection from the database
func GetCollection(db *mongo.Database, collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

// CloseConnection closes the MongoDB connection gracefully
func CloseConnection(ctx context.Context, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("MongoDB connection closed successfully")
	return nil
}

// WithRetry executes a function with automatic retry logic
func WithRetry(ctx context.Context, maxRetries int, fn func() error) error {
	var err error
	
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Wait before retry with exponential backoff
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * time.Second
			log.Printf("Operation failed, retrying in %v... (attempt %d/%d)", waitTime, i+1, maxRetries)
			
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

