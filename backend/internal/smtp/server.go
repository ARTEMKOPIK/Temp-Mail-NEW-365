package smtp

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/artemkopik/tempmail/internal/config"
	"github.com/artemkopik/tempmail/internal/models"
	"github.com/artemkopik/tempmail/internal/repository"
	"github.com/artemkopik/tempmail/internal/websocket"
	"github.com/emersion/go-smtp"
)

// Server represents the SMTP server
type Server struct {
	server         *smtp.Server
	mailboxRepo    *repository.MailboxRepository
	emailRepo      *repository.EmailRepository
	websocketHub   *websocket.Hub
	config         *config.Config
}

// Backend implements smtp.Backend interface
type Backend struct {
	mailboxRepo  *repository.MailboxRepository
	emailRepo    *repository.EmailRepository
	websocketHub *websocket.Hub
	config       *config.Config
}

// Session implements smtp.Session interface
type Session struct {
	backend *Backend
	from    string
	to      []string
}

// NewServer creates a new SMTP server
func NewServer(cfg *config.Config, mailboxRepo *repository.MailboxRepository, 
	emailRepo *repository.EmailRepository, hub *websocket.Hub) *Server {
	
	backend := &Backend{
		mailboxRepo:  mailboxRepo,
		emailRepo:    emailRepo,
		websocketHub: hub,
		config:       cfg,
	}

	s := smtp.NewServer(backend)
	s.Addr = fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	s.Domain = cfg.EmailDomain
	s.MaxMessageBytes = 10 * 1024 * 1024 // 10MB
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true
	s.WriteTimeout = 10 * time.Second
	s.ReadTimeout = 10 * time.Second

	return &Server{
		server:       s,
		mailboxRepo:  mailboxRepo,
		emailRepo:    emailRepo,
		websocketHub: hub,
		config:       cfg,
	}
}

// ListenAndServe starts the SMTP server
func (s *Server) ListenAndServe() error {
	log.Printf("SMTP server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Close gracefully shuts down the SMTP server
func (s *Server) Close() error {
	log.Println("Shutting down SMTP server...")
	return s.server.Close()
}

// NewSession implements smtp.Backend
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		backend: b,
	}, nil
}

// AuthPlain implements smtp.Backend (we don't require auth)
func (b *Backend) AuthPlain(conn *smtp.Conn, username, password string) (smtp.Session, error) {
	// Accept any credentials (no authentication required)
	return &Session{
		backend: b,
	}, nil
}

// Mail implements smtp.Session (MAIL FROM command)
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	log.Printf("MAIL FROM: %s", from)
	return nil
}

// Rcpt implements smtp.Session (RCPT TO command)
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	// Extract address from angle brackets if present
	cleanTo := extractEmailAddress(to)

	// Check if recipient mailbox exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := s.backend.mailboxRepo.Exists(ctx, cleanTo)
	if err != nil {
		log.Printf("Error checking mailbox existence: %v", err)
		return fmt.Errorf("mailbox check failed")
	}

	if !exists {
		log.Printf("Mailbox does not exist: %s", cleanTo)
		// Accept anyway to avoid leaking mailbox information
	}

	s.to = append(s.to, cleanTo)
	log.Printf("RCPT TO: %s", cleanTo)
	return nil
}

// Data implements smtp.Session (email data transfer)
func (s *Session) Data(r io.Reader) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Read entire email
	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("Error reading email data: %v", err)
		return fmt.Errorf("failed to read email data")
	}

	log.Printf("Received email: From=%s, To=%v, Size=%d bytes", s.from, s.to, len(data))

	// Parse email
	email, err := ParseEmail(data, s.from, s.to)
	if err != nil {
		log.Printf("Error parsing email: %v", err)
		return fmt.Errorf("failed to parse email")
	}

	// Set expiry time
	email.ExpiresAt = time.Now().Add(time.Duration(s.backend.config.EmailExpiryMinutes) * time.Minute)

	// Store email for each recipient
	for _, recipient := range s.to {
		// Check if mailbox exists
		exists, err := s.backend.mailboxRepo.Exists(ctx, recipient)
		if err != nil || !exists {
			log.Printf("Skipping non-existent mailbox: %s", recipient)
			continue
		}

		// Create email copy for this recipient
		emailCopy := *email
		emailCopy.Address = recipient

		// Save to database
		if err := s.backend.emailRepo.Create(ctx, &emailCopy); err != nil {
			log.Printf("Error saving email for %s: %v", recipient, err)
			continue
		}

		log.Printf("Email saved for %s: ID=%s", recipient, emailCopy.ID.Hex())

		// Send WebSocket notification
		notification := &models.NewEmailPayload{
			EmailID:    emailCopy.ID.Hex(),
			From:       emailCopy.From,
			Subject:    emailCopy.Subject,
			ReceivedAt: emailCopy.ReceivedAt,
		}

		s.backend.websocketHub.BroadcastNewEmail(recipient, notification)
		log.Printf("WebSocket notification sent for %s", recipient)
	}

	return nil
}

// Reset implements smtp.Session
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

// Logout implements smtp.Session
func (s *Session) Logout() error {
	return nil
}

// extractEmailAddress extracts email address from strings like "<user@domain.com>"
func extractEmailAddress(email string) string {
	email = strings.TrimSpace(email)
	email = strings.Trim(email, "<>")
	return strings.ToLower(email)
}

