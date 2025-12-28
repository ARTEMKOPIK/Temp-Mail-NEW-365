package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

var (
	// ErrInvalidKey is returned when encryption key is invalid
	ErrInvalidKey = errors.New("invalid encryption key")
	// ErrEncryptionFailed is returned when encryption fails
	ErrEncryptionFailed = errors.New("encryption failed")
	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errors.New("decryption failed")
	// ErrInvalidCiphertext is returned when ciphertext is invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Encryption provides AES-256-GCM encryption/decryption
type Encryption struct {
	key []byte
}

// NewEncryption creates a new Encryption service with the given key
// The key will be derived using PBKDF2 if it's not 32 bytes
func NewEncryption(key string) *Encryption {
	// Derive a 32-byte key using PBKDF2
	derivedKey := pbkdf2.Key([]byte(key), []byte("tempmail-salt"), 100000, 32, sha256.New)
	
	return &Encryption{
		key: derivedKey,
	}
}

// Encrypt encrypts plaintext using AES-256-GCM
// Returns base64-encoded ciphertext with nonce prepended
func (e *Encryption) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: failed to generate nonce: %v", ErrEncryptionFailed, err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (e *Encryption) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64: %v", ErrDecryptionFailed, err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Check ciphertext length
	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext too short", ErrInvalidCiphertext)
	}

	// Extract nonce and ciphertext
	nonce, cipherData := decoded[:nonceSize], decoded[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts byte slice
func (e *Encryption) EncryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	encrypted, err := e.Encrypt(string(data))
	if err != nil {
		return nil, err
	}

	return []byte(encrypted), nil
}

// DecryptBytes decrypts to byte slice
func (e *Encryption) DecryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	decrypted, err := e.Decrypt(string(data))
	if err != nil {
		return nil, err
	}

	return []byte(decrypted), nil
}

// ValidateKey checks if the encryption key is valid
func (e *Encryption) ValidateKey() error {
	if len(e.key) != 32 {
		return ErrInvalidKey
	}

	// Test encryption/decryption
	testData := "test"
	encrypted, err := e.Encrypt(testData)
	if err != nil {
		return fmt.Errorf("encryption test failed: %w", err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("decryption test failed: %w", err)
	}

	if decrypted != testData {
		return errors.New("encryption validation failed: data mismatch")
	}

	return nil
}

// RotateKey creates a new Encryption instance with a new key
// This is useful for key rotation scenarios
func RotateKey(oldEncryption *Encryption, newKey string) (*Encryption, error) {
	newEncryption := NewEncryption(newKey)
	
	// Validate new key
	if err := newEncryption.ValidateKey(); err != nil {
		return nil, fmt.Errorf("new key validation failed: %w", err)
	}

	return newEncryption, nil
}

// ReEncrypt re-encrypts data with a new encryption instance
// Useful for key rotation
func ReEncrypt(oldEncryption, newEncryption *Encryption, ciphertext string) (string, error) {
	// Decrypt with old key
	plaintext, err := oldEncryption.Decrypt(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with old key: %w", err)
	}

	// Encrypt with new key
	newCiphertext, err := newEncryption.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with new key: %w", err)
	}

	return newCiphertext, nil
}

