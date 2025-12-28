package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/artemkopik/tempmail/internal/repository"
	"github.com/gin-gonic/gin"
)

// EmailHandler handles email-related requests
type EmailHandler struct {
	emailRepo *repository.EmailRepository
}

// NewEmailHandler creates a new EmailHandler
func NewEmailHandler(emailRepo *repository.EmailRepository) *EmailHandler {
	return &EmailHandler{
		emailRepo: emailRepo,
	}
}

// GetEmail retrieves a single email by ID with full content
// GET /api/email/:id
func (h *EmailHandler) GetEmail(c *gin.Context) {
	emailID := c.Param("id")

	if emailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email, err := h.emailRepo.GetByID(ctx, emailID)
	if err != nil {
		log.Printf("Error getting email %s: %v", emailID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "email not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email": email,
	})
}

// DeleteEmail deletes a single email
// DELETE /api/email/:id
func (h *EmailHandler) DeleteEmail(c *gin.Context) {
	emailID := c.Param("id")

	if emailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.emailRepo.DeleteByID(ctx, emailID); err != nil {
		log.Printf("Error deleting email %s: %v", emailID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete email",
			"message": err.Error(),
		})
		return
	}

	log.Printf("Email deleted: %s", emailID)

	c.JSON(http.StatusOK, gin.H{
		"message": "email deleted successfully",
	})
}

