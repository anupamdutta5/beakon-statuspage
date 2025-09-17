package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// TokenGenerator provides secure token generation functionality
type TokenGenerator struct {
	charset string
}

// NewTokenGenerator creates a new secure token generator
func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{
		charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}
}

// GenerateSecureToken generates a cryptographically secure random token
func (tg *TokenGenerator) GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	token := make([]byte, length)
	charsetLen := big.NewInt(int64(len(tg.charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		token[i] = tg.charset[randomIndex.Int64()]
	}

	return string(token), nil
}

// GenerateHexToken generates a secure hex-encoded token
func (tg *TokenGenerator) GenerateHexToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("byte length must be positive")
	}

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// GenerateSessionID generates a secure session ID
func (tg *TokenGenerator) GenerateSessionID() (string, error) {
	// Generate 32 random bytes (256 bits) for session ID
	return tg.GenerateHexToken(32)
}

// GenerateAPIKey generates a secure API key with prefix
func (tg *TokenGenerator) GenerateAPIKey(prefix string) (string, error) {
	if prefix == "" {
		prefix = "bk"
	}

	// Generate 32 random bytes for the key part
	keyPart, err := tg.GenerateHexToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	return fmt.Sprintf("%s_%s", prefix, keyPart), nil
}

// GenerateVerificationToken generates a time-based verification token
func (tg *TokenGenerator) GenerateVerificationToken() (string, error) {
	// Generate timestamp component
	timestamp := time.Now().Unix()

	// Generate random component (16 bytes)
	randomPart, err := tg.GenerateHexToken(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Combine timestamp and random part
	combined := fmt.Sprintf("%d_%s", timestamp, randomPart)

	// Hash the combined string for additional security
	hash := sha256.Sum256([]byte(combined))

	return hex.EncodeToString(hash[:16]), nil // Return first 16 bytes of hash
}

// GenerateCSRFToken generates a CSRF token
func (tg *TokenGenerator) GenerateCSRFToken() (string, error) {
	return tg.GenerateHexToken(32)
}

// GeneratePasswordResetToken generates a secure password reset token
func (tg *TokenGenerator) GeneratePasswordResetToken() (string, error) {
	// Generate a longer token for password reset (48 bytes = 96 hex chars)
	return tg.GenerateHexToken(48)
}

// GenerateCorrelationID generates a correlation ID for request tracing
func (tg *TokenGenerator) GenerateCorrelationID() (string, error) {
	// Generate UUID-like format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	segments := []int{8, 4, 4, 4, 12}
	var parts []string

	for _, segmentLen := range segments {
		segment, err := tg.GenerateHexToken(segmentLen / 2) // Divide by 2 because hex encoding doubles length
		if err != nil {
			return "", fmt.Errorf("failed to generate correlation ID segment: %w", err)
		}
		parts = append(parts, segment)
	}

	return strings.Join(parts, "-"), nil
}

// ValidateTokenLength validates if a token meets minimum security requirements
func ValidateTokenLength(token string, minLength int) error {
	if len(token) < minLength {
		return fmt.Errorf("token length %d is below minimum required length %d", len(token), minLength)
	}
	return nil
}

// ValidateTokenCharset validates if a token contains only allowed characters
func ValidateTokenCharset(token, allowedCharset string) error {
	for _, char := range token {
		if !strings.ContainsRune(allowedCharset, char) {
			return fmt.Errorf("token contains invalid character: %c", char)
		}
	}
	return nil
}

// IsSecureToken performs basic security validation on a token
func IsSecureToken(token string, minLength int) bool {
	if len(token) < minLength {
		return false
	}

	// Check for minimum entropy (basic check)
	charMap := make(map[rune]bool)
	for _, char := range token {
		charMap[char] = true
	}

	// Require at least 1/4 of token length to be unique characters
	if len(charMap) < len(token)/4 {
		return false
	}

	return true
}

// Global token generator instance
var defaultGenerator = NewTokenGenerator()

// Package-level convenience functions

// GenerateSecureToken generates a cryptographically secure random token using the default generator
func GenerateSecureToken(length int) (string, error) {
	return defaultGenerator.GenerateSecureToken(length)
}

// GenerateHexToken generates a secure hex-encoded token using the default generator
func GenerateHexToken(byteLength int) (string, error) {
	return defaultGenerator.GenerateHexToken(byteLength)
}

// GenerateSessionID generates a secure session ID using the default generator
func GenerateSessionID() (string, error) {
	return defaultGenerator.GenerateSessionID()
}

// GenerateAPIKey generates a secure API key using the default generator
func GenerateAPIKey(prefix string) (string, error) {
	return defaultGenerator.GenerateAPIKey(prefix)
}

// GenerateVerificationToken generates a verification token using the default generator
func GenerateVerificationToken() (string, error) {
	return defaultGenerator.GenerateVerificationToken()
}

// GenerateCSRFToken generates a CSRF token using the default generator
func GenerateCSRFToken() (string, error) {
	return defaultGenerator.GenerateCSRFToken()
}

// GeneratePasswordResetToken generates a password reset token using the default generator
func GeneratePasswordResetToken() (string, error) {
	return defaultGenerator.GeneratePasswordResetToken()
}

// GenerateCorrelationID generates a correlation ID using the default generator
func GenerateCorrelationID() (string, error) {
	return defaultGenerator.GenerateCorrelationID()
}