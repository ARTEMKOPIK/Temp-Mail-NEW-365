package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Email represents an email message stored in MongoDB
type Email struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Address        string             `bson:"address" json:"address"`
	From           string             `bson:"from" json:"from"`
	To             []string           `bson:"to" json:"to"`
	Subject        string             `bson:"subject" json:"subject"`
	HTMLBody       string             `bson:"html_body" json:"htmlBody"`
	TextBody       string             `bson:"text_body" json:"textBody"`
	IsHTMLEncrypted bool              `bson:"is_html_encrypted" json:"-"`
	IsTextEncrypted bool              `bson:"is_text_encrypted" json:"-"`
	ReceivedAt     time.Time          `bson:"received_at" json:"receivedAt"`
	ExpiresAt      time.Time          `bson:"expires_at" json:"expiresAt"`
	Size           int64              `bson:"size" json:"size"`
	HasAttachments bool               `bson:"has_attachments" json:"hasAttachments"`
}

// EmailListItem represents a simplified email for list view
type EmailListItem struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	Subject    string    `json:"subject"`
	ReceivedAt time.Time `json:"receivedAt"`
	Size       int64     `json:"size"`
}

// Mailbox represents a temporary mailbox stored in Redis
type Mailbox struct {
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	ExpiryTime int       `json:"expiryTime"` // in minutes
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewEmailPayload represents the payload for new email notification
type NewEmailPayload struct {
	EmailID    string    `json:"emailId"`
	From       string    `json:"from"`
	Subject    string    `json:"subject"`
	ReceivedAt time.Time `json:"receivedAt"`
}

// ErrorPayload represents an error message payload
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ToListItem converts Email to EmailListItem
func (e *Email) ToListItem() EmailListItem {
	return EmailListItem{
		ID:         e.ID.Hex(),
		From:       e.From,
		Subject:    e.Subject,
		ReceivedAt: e.ReceivedAt,
		Size:       e.Size,
	}
}

// CreateMailbox creates a new Mailbox with expiry time
func CreateMailbox(address string, expiryMinutes int) *Mailbox {
	now := time.Now()
	return &Mailbox{
		Address:    address,
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Duration(expiryMinutes) * time.Minute),
		ExpiryTime: expiryMinutes,
	}
}

