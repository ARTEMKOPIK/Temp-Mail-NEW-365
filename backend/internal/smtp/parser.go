package smtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"github.com/artemkopik/tempmail/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParseEmail parses raw email data into an Email model
func ParseEmail(data []byte, from string, to []string) (*models.Email, error) {
	// Parse email message
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	// Extract headers
	subject := decodeHeader(msg.Header.Get("Subject"))
	fromHeader := msg.Header.Get("From")
	if fromHeader != "" {
		from = extractEmailFromHeader(fromHeader)
	}

	// Parse date
	dateStr := msg.Header.Get("Date")
	receivedAt := time.Now()
	if dateStr != "" {
		if parsedDate, err := mail.ParseDate(dateStr); err == nil {
			receivedAt = parsedDate
		}
	}

	// Parse message body
	contentType := msg.Header.Get("Content-Type")
	htmlBody, textBody, hasAttachments := parseBody(msg.Body, contentType)

	// Create email model
	email := &models.Email{
		ID:             primitive.NewObjectID(),
		From:           from,
		To:             to,
		Subject:        subject,
		HTMLBody:       htmlBody,
		TextBody:       textBody,
		ReceivedAt:     receivedAt,
		Size:           int64(len(data)),
		HasAttachments: hasAttachments,
	}

	return email, nil
}

// parseBody parses the email body based on content type
func parseBody(body io.Reader, contentType string) (html, text string, hasAttachments bool) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Default to plain text if parsing fails
		data, _ := io.ReadAll(body)
		return "", string(data), false
	}

	// Handle multipart messages
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			log.Println("Warning: multipart message without boundary")
			return "", "", false
		}

		return parseMultipart(body, boundary)
	}

	// Handle single part messages
	data, _ := io.ReadAll(body)
	
	// Decode based on transfer encoding
	transferEncoding := params["transfer-encoding"]
	decoded := decodeBody(data, transferEncoding)

	if strings.Contains(mediaType, "text/html") {
		return decoded, "", false
	}

	return "", decoded, false
}

// parseMultipart parses a multipart email message
func parseMultipart(body io.Reader, boundary string) (html, text string, hasAttachments bool) {
	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading multipart: %v", err)
			break
		}

		// Get content type
		contentType := part.Header.Get("Content-Type")
		mediaType, params, _ := mime.ParseMediaType(contentType)

		// Handle nested multipart
		if strings.HasPrefix(mediaType, "multipart/") {
			nestedBoundary := params["boundary"]
			if nestedBoundary != "" {
				nestedHTML, nestedText, nestedAttachments := parseMultipart(part, nestedBoundary)
				if nestedHTML != "" {
					html = nestedHTML
				}
				if nestedText != "" {
					text = nestedText
				}
				if nestedAttachments {
					hasAttachments = true
				}
			}
			continue
		}

		// Read part data
		partData, err := io.ReadAll(part)
		if err != nil {
			log.Printf("Error reading part data: %v", err)
			continue
		}

		// Decode transfer encoding
		transferEncoding := part.Header.Get("Content-Transfer-Encoding")
		decoded := decodeBody(partData, transferEncoding)

		// Check for attachment
		disposition := part.Header.Get("Content-Disposition")
		if disposition != "" && strings.HasPrefix(disposition, "attachment") {
			hasAttachments = true
			log.Printf("Found attachment: %s", part.FileName())
			// We're not storing attachments, just noting their presence
			continue
		}

		// Handle text parts
		if strings.Contains(mediaType, "text/html") {
			html = decoded
		} else if strings.Contains(mediaType, "text/plain") {
			text = decoded
		} else if strings.HasPrefix(mediaType, "application/") || 
			strings.HasPrefix(mediaType, "image/") ||
			strings.HasPrefix(mediaType, "audio/") ||
			strings.HasPrefix(mediaType, "video/") {
			// These are attachments
			hasAttachments = true
		}
	}

	return html, text, hasAttachments
}

// decodeBody decodes email body based on transfer encoding
func decodeBody(data []byte, encoding string) string {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	switch encoding {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			log.Printf("Base64 decode error: %v", err)
			return string(data)
		}
		return string(decoded)

	case "quoted-printable":
		reader := quotedprintable.NewReader(bytes.NewReader(data))
		decoded, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Quoted-printable decode error: %v", err)
			return string(data)
		}
		return string(decoded)

	case "7bit", "8bit", "binary", "":
		return string(data)

	default:
		log.Printf("Unknown transfer encoding: %s", encoding)
		return string(data)
	}
}

// decodeHeader decodes MIME encoded-word headers (RFC 2047)
func decodeHeader(header string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(header)
	if err != nil {
		// If decoding fails, return original
		return header
	}
	return decoded
}

// extractEmailFromHeader extracts email address from header like "Name <email@domain.com>"
func extractEmailFromHeader(header string) string {
	// Try to parse as address
	addr, err := mail.ParseAddress(header)
	if err != nil {
		// Fall back to simple extraction
		start := strings.Index(header, "<")
		end := strings.Index(header, ">")
		if start != -1 && end != -1 && end > start {
			return strings.ToLower(header[start+1 : end])
		}
		return strings.ToLower(strings.TrimSpace(header))
	}

	return strings.ToLower(addr.Address)
}

