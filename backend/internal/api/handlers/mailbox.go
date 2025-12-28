package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/artemkopik/tempmail/internal/repository"
	"github.com/gin-gonic/gin"
)

// MailboxHandler handles mailbox-related requests
type MailboxHandler struct {
	mailboxRepo *repository.MailboxRepository
	emailRepo   *repository.EmailRepository
}

// NewMailboxHandler creates a new MailboxHandler
func NewMailboxHandler(mailboxRepo *repository.MailboxRepository, emailRepo *repository.EmailRepository) *MailboxHandler {
	return &MailboxHandler{
		mailboxRepo: mailboxRepo,
		emailRepo:   emailRepo,
	}
}

// CreateMailboxRequest represents the request body for creating a mailbox
type CreateMailboxRequest struct {
	ExpiryMinutes int `json:"expiryMinutes" binding:"required,min=5,max=120"`
}

// CreateMailbox creates a new temporary mailbox
// POST /api/mailbox
func (h *MailboxHandler) CreateMailbox(c *gin.Context) {
	var req CreateMailboxRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"message": "expiryMinutes must be between 5 and 120",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create mailbox
	mailbox, err := h.mailboxRepo.Create(ctx, req.ExpiryMinutes)
	if err != nil {
		log.Printf("Error creating mailbox: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create mailbox",
			"message": err.Error(),
		})
		return
	}

	log.Printf("Mailbox created: %s", mailbox.Address)

	c.JSON(http.StatusCreated, gin.H{
		"mailbox": mailbox,
	})
}

// GetMailbox retrieves mailbox information
// GET /api/mailbox/:address
func (h *MailboxHandler) GetMailbox(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mailbox, err := h.mailboxRepo.Get(ctx, address)
	if err != nil {
		log.Printf("Error getting mailbox %s: %v", address, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "mailbox not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mailbox": mailbox,
	})
}

// DeleteMailbox deletes a mailbox and all its emails
// DELETE /api/mailbox/:address
func (h *MailboxHandler) DeleteMailbox(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete all emails for this mailbox
	if err := h.emailRepo.DeleteByAddress(ctx, address); err != nil {
		log.Printf("Error deleting emails for %s: %v", address, err)
	}

	// Delete mailbox
	if err := h.mailboxRepo.Delete(ctx, address); err != nil {
		log.Printf("Error deleting mailbox %s: %v", address, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete mailbox",
			"message": err.Error(),
		})
		return
	}

	log.Printf("Mailbox deleted: %s", address)

	c.JSON(http.StatusOK, gin.H{
		"message": "mailbox deleted successfully",
	})
}

// GetEmails retrieves all emails for a mailbox
// GET /api/mailbox/:address/emails
func (h *MailboxHandler) GetEmails(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if mailbox exists
	exists, err := h.mailboxRepo.Exists(ctx, address)
	if err != nil {
		log.Printf("Error checking mailbox existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to check mailbox",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "mailbox not found or expired",
		})
		return
	}

	// Get emails
	emails, err := h.emailRepo.GetByAddress(ctx, address)
	if err != nil {
		log.Printf("Error getting emails for %s: %v", address, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get emails",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"count":  len(emails),
	})
}

