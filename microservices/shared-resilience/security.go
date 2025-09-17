package resilience

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
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
	return tg.GenerateHexToken(32)
}

// GenerateAPIKey generates a secure API key with prefix
func (tg *TokenGenerator) GenerateAPIKey(prefix string) (string, error) {
	if prefix == "" {
		prefix = "bk"
	}

	keyPart, err := tg.GenerateHexToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	return fmt.Sprintf("%s_%s", prefix, keyPart), nil
}

// GenerateVerificationToken generates a time-based verification token
func (tg *TokenGenerator) GenerateVerificationToken() (string, error) {
	timestamp := time.Now().Unix()
	randomPart, err := tg.GenerateHexToken(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}

	combined := fmt.Sprintf("%d_%s", timestamp, randomPart)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:16]), nil
}

// GenerateCorrelationID generates a correlation ID for request tracing
func (tg *TokenGenerator) GenerateCorrelationID() (string, error) {
	segments := []int{8, 4, 4, 4, 12}
	var parts []string

	for _, segmentLen := range segments {
		segment, err := tg.GenerateHexToken(segmentLen / 2)
		if err != nil {
			return "", fmt.Errorf("failed to generate correlation ID segment: %w", err)
		}
		parts = append(parts, segment)
	}

	return strings.Join(parts, "-"), nil
}

// PasswordHasher provides secure password hashing functionality
type PasswordHasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// NewPasswordHasher creates a new password hasher with secure defaults
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		memory:      64 * 1024, // 64 MB
		iterations:  3,         // 3 iterations
		parallelism: 2,         // 2 threads
		saltLength:  16,        // 16 bytes salt
		keyLength:   32,        // 32 bytes key
	}
}

// HashPassword hashes a password using Argon2id
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt := make([]byte, ph.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, ph.iterations, ph.memory, ph.parallelism, ph.keyLength)

	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		ph.memory, ph.iterations, ph.parallelism, saltEncoded, hashEncoded), nil
}

// VerifyPassword verifies a password against its hash
func (ph *PasswordHasher) VerifyPassword(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, fmt.Errorf("password and hash cannot be empty")
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, fmt.Errorf("invalid hash format")
	}

	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1, nil
}

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

// String returns string representation of password strength
func (ps PasswordStrength) String() string {
	switch ps {
	case PasswordWeak:
		return "weak"
	case PasswordMedium:
		return "medium"
	case PasswordStrong:
		return "strong"
	case PasswordVeryStrong:
		return "very-strong"
	default:
		return "unknown"
	}
}

// PasswordRequirements defines password complexity requirements
type PasswordRequirements struct {
	MinLength          int
	MaxLength          int
	RequireUpper       bool
	RequireLower       bool
	RequireDigit       bool
	RequireSpecial     bool
	ForbiddenPasswords []string
}

// DefaultPasswordRequirements returns secure default password requirements
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:    12,
		MaxLength:    128,
		RequireUpper: true,
		RequireLower: true,
		RequireDigit: true,
		RequireSpecial: true,
		ForbiddenPasswords: []string{
			"password", "123456", "password123", "admin", "user",
			"qwerty", "letmein", "welcome", "monkey", "dragon",
		},
	}
}

// ValidatePassword validates a password against requirements
func ValidatePassword(password string, requirements PasswordRequirements) []string {
	var errors []string

	if len(password) < requirements.MinLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters long", requirements.MinLength))
	}

	if len(password) > requirements.MaxLength {
		errors = append(errors, fmt.Sprintf("password must be no longer than %d characters", requirements.MaxLength))
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	if requirements.RequireUpper && !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}

	if requirements.RequireLower && !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}

	if requirements.RequireDigit && !hasDigit {
		errors = append(errors, "password must contain at least one digit")
	}

	if requirements.RequireSpecial && !hasSpecial {
		errors = append(errors, "password must contain at least one special character")
	}

	passwordLower := strings.ToLower(password)
	for _, forbidden := range requirements.ForbiddenPasswords {
		if passwordLower == strings.ToLower(forbidden) || strings.Contains(passwordLower, strings.ToLower(forbidden)) {
			errors = append(errors, "password contains forbidden words or patterns")
			break
		}
	}

	return errors
}

// Global instances
var (
	defaultTokenGenerator = NewTokenGenerator()
	defaultPasswordHasher = NewPasswordHasher()
)

// Package-level convenience functions

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	return defaultTokenGenerator.GenerateSecureToken(length)
}

// GenerateHexToken generates a secure hex-encoded token
func GenerateHexToken(byteLength int) (string, error) {
	return defaultTokenGenerator.GenerateHexToken(byteLength)
}

// GenerateSessionID generates a secure session ID
func GenerateSessionID() (string, error) {
	return defaultTokenGenerator.GenerateSessionID()
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey(prefix string) (string, error) {
	return defaultTokenGenerator.GenerateAPIKey(prefix)
}

// GenerateVerificationToken generates a verification token
func GenerateVerificationToken() (string, error) {
	return defaultTokenGenerator.GenerateVerificationToken()
}

// GenerateCorrelationID generates a correlation ID
func GenerateCorrelationID() (string, error) {
	return defaultTokenGenerator.GenerateCorrelationID()
}

// HashPassword hashes a password using the default hasher
func HashPassword(password string) (string, error) {
	return defaultPasswordHasher.HashPassword(password)
}

// VerifyPassword verifies a password using the default hasher
func VerifyPassword(password, hash string) (bool, error) {
	return defaultPasswordHasher.VerifyPassword(password, hash)
}