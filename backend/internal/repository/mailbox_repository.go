package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/artemkopik/tempmail/internal/database"
	"github.com/artemkopik/tempmail/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	mailboxPrefix      = "mailbox:"
	subscribersPrefix  = "subscribers:"
	defaultExpiryMins  = 60
	minAddressLength   = 8
	maxAddressLength   = 16
)

// MailboxRepository handles mailbox operations in Redis
type MailboxRepository struct {
	client *redis.Client
	domain string
}

// NewMailboxRepository creates a new MailboxRepository
func NewMailboxRepository(client *redis.Client, domain string) *MailboxRepository {
	return &MailboxRepository{
		client: client,
		domain: domain,
	}
}

// Create generates and stores a new temporary mailbox
func (r *MailboxRepository) Create(ctx context.Context, expiryMinutes int) (*models.Mailbox, error) {
	if expiryMinutes <= 0 {
		expiryMinutes = defaultExpiryMins
	}

	// Generate unique address
	address, err := r.generateUniqueAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address: %w", err)
	}

	// Create mailbox
	mailbox := models.CreateMailbox(address, expiryMinutes)

	// Store in Redis with expiration
	key := mailboxPrefix + address
	expiration := time.Duration(expiryMinutes) * time.Minute

	if err := database.RedisSet(ctx, r.client, key, mailbox.CreatedAt.Unix(), expiration); err != nil {
		return nil, fmt.Errorf("failed to store mailbox: %w", err)
	}

	log.Printf("Mailbox created: %s (expires in %d minutes)", address, expiryMinutes)

	return mailbox, nil
}

// Get retrieves a mailbox by address
func (r *MailboxRepository) Get(ctx context.Context, address string) (*models.Mailbox, error) {
	key := mailboxPrefix + address

	// Check if mailbox exists
	exists, err := database.RedisExists(ctx, r.client, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check mailbox existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("mailbox not found or expired")
	}

	// Get created timestamp
	createdAtStr, err := database.RedisGet(ctx, r.client, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}

	// Parse created timestamp
	var createdAtUnix int64
	fmt.Sscanf(createdAtStr, "%d", &createdAtUnix)
	createdAt := time.Unix(createdAtUnix, 0)

	// Get TTL
	ttl, err := database.RedisTTL(ctx, r.client, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox TTL: %w", err)
	}

	if ttl < 0 {
		return nil, fmt.Errorf("mailbox has expired")
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(ttl)
	expiryMinutes := int(ttl.Minutes())

	mailbox := &models.Mailbox{
		Address:    address,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
		ExpiryTime: expiryMinutes,
	}

	return mailbox, nil
}

// Exists checks if a mailbox exists
func (r *MailboxRepository) Exists(ctx context.Context, address string) (bool, error) {
	key := mailboxPrefix + address
	return database.RedisExists(ctx, r.client, key)
}

// Delete removes a mailbox and its subscribers
func (r *MailboxRepository) Delete(ctx context.Context, address string) error {
	mailboxKey := mailboxPrefix + address
	subscribersKey := subscribersPrefix + address

	// Delete both mailbox and subscribers
	if err := database.RedisDelete(ctx, r.client, mailboxKey, subscribersKey); err != nil {
		return fmt.Errorf("failed to delete mailbox: %w", err)
	}

	log.Printf("Mailbox deleted: %s", address)

	return nil
}

// ExtendExpiry extends the expiry time of a mailbox
func (r *MailboxRepository) ExtendExpiry(ctx context.Context, address string, additionalMinutes int) error {
	key := mailboxPrefix + address

	// Get current TTL
	ttl, err := database.RedisTTL(ctx, r.client, key)
	if err != nil {
		return fmt.Errorf("failed to get mailbox TTL: %w", err)
	}

	if ttl < 0 {
		return fmt.Errorf("mailbox has expired")
	}

	// Add additional time
	newExpiry := ttl + time.Duration(additionalMinutes)*time.Minute

	// Set new expiry
	if err := database.RedisExpire(ctx, r.client, key, newExpiry); err != nil {
		return fmt.Errorf("failed to extend mailbox expiry: %w", err)
	}

	log.Printf("Mailbox expiry extended: %s (+%d minutes)", address, additionalMinutes)

	return nil
}

// AddSubscriber adds a WebSocket subscriber to a mailbox
func (r *MailboxRepository) AddSubscriber(ctx context.Context, address, subscriberID string) error {
	key := subscribersPrefix + address

	if err := database.RedisSAdd(ctx, r.client, key, subscriberID); err != nil {
		return fmt.Errorf("failed to add subscriber: %w", err)
	}

	// Set expiry on subscribers set (same as mailbox)
	mailboxKey := mailboxPrefix + address
	ttl, err := database.RedisTTL(ctx, r.client, mailboxKey)
	if err == nil && ttl > 0 {
		database.RedisExpire(ctx, r.client, key, ttl)
	}

	return nil
}

// RemoveSubscriber removes a WebSocket subscriber from a mailbox
func (r *MailboxRepository) RemoveSubscriber(ctx context.Context, address, subscriberID string) error {
	key := subscribersPrefix + address

	if err := database.RedisSRem(ctx, r.client, key, subscriberID); err != nil {
		return fmt.Errorf("failed to remove subscriber: %w", err)
	}

	return nil
}

// GetSubscribers retrieves all subscribers for a mailbox
func (r *MailboxRepository) GetSubscribers(ctx context.Context, address string) ([]string, error) {
	key := subscribersPrefix + address

	subscribers, err := database.RedisSMembers(ctx, r.client, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers: %w", err)
	}

	return subscribers, nil
}

// generateUniqueAddress generates a unique email address
func (r *MailboxRepository) generateUniqueAddress(ctx context.Context) (string, error) {
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		// Generate random string
		randomStr, err := generateRandomString(12)
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}

		address := fmt.Sprintf("%s@%s", randomStr, r.domain)

		// Check if address already exists
		exists, err := r.Exists(ctx, address)
		if err != nil {
			return "", fmt.Errorf("failed to check address existence: %w", err)
		}

		if !exists {
			return address, nil
		}

		log.Printf("Address collision detected, retrying... (attempt %d/%d)", i+1, maxAttempts)
	}

	return "", fmt.Errorf("failed to generate unique address after %d attempts", maxAttempts)
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes)[:length], nil
}

// GetAllMailboxes retrieves all active mailboxes (for admin/stats purposes)
func (r *MailboxRepository) GetAllMailboxes(ctx context.Context) ([]string, error) {
	pattern := mailboxPrefix + "*"

	var allKeys []string
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Remove prefix to get address
		address := key[len(mailboxPrefix):]
		allKeys = append(allKeys, address)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan mailboxes: %w", err)
	}

	return allKeys, nil
}

// GetStats returns statistics about mailboxes
func (r *MailboxRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	mailboxes, err := r.GetAllMailboxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailboxes: %w", err)
	}

	stats := map[string]interface{}{
		"total_active_mailboxes": len(mailboxes),
	}

	return stats, nil
}

