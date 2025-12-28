package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/artemkopik/tempmail/internal/crypto"
	"github.com/artemkopik/tempmail/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	emailsCollection = "emails"
)

// EmailRepository handles email CRUD operations
type EmailRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	encryption *crypto.Encryption
}

// NewEmailRepository creates a new EmailRepository
func NewEmailRepository(db *mongo.Database, encryption *crypto.Encryption) *EmailRepository {
	return &EmailRepository{
		db:         db,
		collection: db.Collection(emailsCollection),
		encryption: encryption,
	}
}

// Create inserts a new email into the database with encryption
func (r *EmailRepository) Create(ctx context.Context, email *models.Email) error {
	if email == nil {
		return fmt.Errorf("email cannot be nil")
	}

	// Set timestamps
	if email.ReceivedAt.IsZero() {
		email.ReceivedAt = time.Now()
	}

	// Encrypt email bodies
	if email.HTMLBody != "" {
		encryptedHTML, err := r.encryption.Encrypt(email.HTMLBody)
		if err != nil {
			return fmt.Errorf("failed to encrypt HTML body: %w", err)
		}
		email.HTMLBody = encryptedHTML
		email.IsHTMLEncrypted = true
	}

	if email.TextBody != "" {
		encryptedText, err := r.encryption.Encrypt(email.TextBody)
		if err != nil {
			return fmt.Errorf("failed to encrypt text body: %w", err)
		}
		email.TextBody = encryptedText
		email.IsTextEncrypted = true
	}

	// Insert into MongoDB
	result, err := r.collection.InsertOne(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to insert email: %w", err)
	}

	email.ID = result.InsertedID.(primitive.ObjectID)
	
	log.Printf("Email created successfully: ID=%s, Address=%s, From=%s", email.ID.Hex(), email.Address, email.From)

	return nil
}

// GetByID retrieves an email by ID with automatic decryption
func (r *EmailRepository) GetByID(ctx context.Context, id string) (*models.Email, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid email ID: %w", err)
	}

	var email models.Email
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("email not found")
		}
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	// Decrypt bodies
	if err := r.decryptEmail(&email); err != nil {
		return nil, fmt.Errorf("failed to decrypt email: %w", err)
	}

	return &email, nil
}

// GetByAddress retrieves all emails for a specific address
func (r *EmailRepository) GetByAddress(ctx context.Context, address string) ([]models.EmailListItem, error) {
	// Sort by received_at descending (newest first)
	opts := options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"address": address}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find emails: %w", err)
	}
	defer cursor.Close(ctx)

	var emails []models.EmailListItem
	for cursor.Next(ctx) {
		var email models.Email
		if err := cursor.Decode(&email); err != nil {
			log.Printf("Warning: failed to decode email: %v", err)
			continue
		}

		emails = append(emails, email.ToListItem())
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	if emails == nil {
		emails = []models.EmailListItem{}
	}

	return emails, nil
}

// Count counts emails for a specific address
func (r *EmailRepository) Count(ctx context.Context, address string) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"address": address})
	if err != nil {
		return 0, fmt.Errorf("failed to count emails: %w", err)
	}
	return count, nil
}

// DeleteByAddress deletes all emails for a specific address
func (r *EmailRepository) DeleteByAddress(ctx context.Context, address string) error {
	result, err := r.collection.DeleteMany(ctx, bson.M{"address": address})
	if err != nil {
		return fmt.Errorf("failed to delete emails: %w", err)
	}

	log.Printf("Deleted %d emails for address %s", result.DeletedCount, address)

	return nil
}

// DeleteByID deletes a single email by ID
func (r *EmailRepository) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid email ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("email not found")
	}

	log.Printf("Email deleted successfully: ID=%s", id)

	return nil
}

// DeleteExpired deletes all expired emails (backup for TTL index)
func (r *EmailRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired emails: %w", err)
	}

	if result.DeletedCount > 0 {
		log.Printf("Deleted %d expired emails", result.DeletedCount)
	}

	return result.DeletedCount, nil
}

// GetLatest retrieves the latest N emails for an address
func (r *EmailRepository) GetLatest(ctx context.Context, address string, limit int64) ([]models.EmailListItem, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "received_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{"address": address}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find emails: %w", err)
	}
	defer cursor.Close(ctx)

	var emails []models.EmailListItem
	for cursor.Next(ctx) {
		var email models.Email
		if err := cursor.Decode(&email); err != nil {
			log.Printf("Warning: failed to decode email: %v", err)
			continue
		}

		emails = append(emails, email.ToListItem())
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	if emails == nil {
		emails = []models.EmailListItem{}
	}

	return emails, nil
}

// Update updates an email (for future use)
func (r *EmailRepository) Update(ctx context.Context, id string, updates bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid email ID: %w", err)
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("email not found")
	}

	log.Printf("Email updated successfully: ID=%s", id)

	return nil
}

// GetStats returns statistics about emails
func (r *EmailRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	totalCount, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total emails: %w", err)
	}

	expiredCount, err := r.collection.CountDocuments(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count expired emails: %w", err)
	}

	stats := map[string]interface{}{
		"total":   totalCount,
		"expired": expiredCount,
		"active":  totalCount - expiredCount,
	}

	return stats, nil
}

// decryptEmail decrypts email bodies in place
func (r *EmailRepository) decryptEmail(email *models.Email) error {
	if email.IsHTMLEncrypted && email.HTMLBody != "" {
		decrypted, err := r.encryption.Decrypt(email.HTMLBody)
		if err != nil {
			return fmt.Errorf("failed to decrypt HTML body: %w", err)
		}
		email.HTMLBody = decrypted
		email.IsHTMLEncrypted = false
	}

	if email.IsTextEncrypted && email.TextBody != "" {
		decrypted, err := r.encryption.Decrypt(email.TextBody)
		if err != nil {
			return fmt.Errorf("failed to decrypt text body: %w", err)
		}
		email.TextBody = decrypted
		email.IsTextEncrypted = false
	}

	return nil
}

// Exists checks if an email exists by ID
func (r *EmailRepository) Exists(ctx context.Context, id string) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, fmt.Errorf("invalid email ID: %w", err)
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": objectID})
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

