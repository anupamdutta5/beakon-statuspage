package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordHasher provides secure password hashing functionality
type PasswordHasher struct {
	// Argon2 parameters
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

	// Generate a cryptographically secure random salt
	salt := make([]byte, ph.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash using Argon2id
	hash := argon2.IDKey([]byte(password), salt, ph.iterations, ph.memory, ph.parallelism, ph.keyLength)

	// Encode the hash in the format: $argon2id$v=19$m=memory,t=iterations,p=parallelism$salt$hash
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		ph.memory, ph.iterations, ph.parallelism, saltEncoded, hashEncoded), nil
}

// VerifyPassword verifies a password against its hash
func (ph *PasswordHasher) VerifyPassword(password, hash string) (bool, error) {
	if password == "" {
		return false, fmt.Errorf("password cannot be empty")
	}

	if hash == "" {
		return false, fmt.Errorf("hash cannot be empty")
	}

	// Parse the hash
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported hash algorithm: %s", parts[1])
	}

	if parts[2] != "v=19" {
		return false, fmt.Errorf("unsupported argon2 version: %s", parts[2])
	}

	// Parse parameters
	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Compute hash for the provided password
	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedHash)))

	// Use constant-time comparison to prevent timing attacks
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
	MinLength    int
	MaxLength    int
	RequireUpper bool
	RequireLower bool
	RequireDigit bool
	RequireSpecial bool
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

	// Check length
	if len(password) < requirements.MinLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters long", requirements.MinLength))
	}

	if len(password) > requirements.MaxLength {
		errors = append(errors, fmt.Sprintf("password must be no longer than %d characters", requirements.MaxLength))
	}

	// Check character requirements
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

	// Check forbidden passwords
	passwordLower := strings.ToLower(password)
	for _, forbidden := range requirements.ForbiddenPasswords {
		if passwordLower == strings.ToLower(forbidden) || strings.Contains(passwordLower, strings.ToLower(forbidden)) {
			errors = append(errors, "password contains forbidden words or patterns")
			break
		}
	}

	return errors
}

// CalculatePasswordStrength calculates the strength of a password
func CalculatePasswordStrength(password string) PasswordStrength {
	score := 0

	// Length scoring
	if len(password) >= 8 {
		score += 1
	}
	if len(password) >= 12 {
		score += 1
	}
	if len(password) >= 16 {
		score += 1
	}

	// Character variety scoring
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
		default:
			hasSpecial = true
		}
	}

	if hasUpper {
		score += 1
	}
	if hasLower {
		score += 1
	}
	if hasDigit {
		score += 1
	}
	if hasSpecial {
		score += 1
	}

	// Determine strength based on score
	switch {
	case score >= 7:
		return PasswordVeryStrong
	case score >= 5:
		return PasswordStrong
	case score >= 3:
		return PasswordMedium
	default:
		return PasswordWeak
	}
}

// GenerateSecurePassword generates a cryptographically secure password
func GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8 characters")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;':\",./<>?"

	password := make([]byte, length)
	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}

	// Ensure the generated password meets requirements
	requirements := DefaultPasswordRequirements()
	if errors := ValidatePassword(string(password), requirements); len(errors) > 0 {
		// Recursively generate until we get a valid password (very unlikely to fail)
		return GenerateSecurePassword(length)
	}

	return string(password), nil
}

// Global password hasher instance
var defaultHasher = NewPasswordHasher()

// Package-level convenience functions

// HashPassword hashes a password using the default hasher
func HashPassword(password string) (string, error) {
	return defaultHasher.HashPassword(password)
}

// VerifyPassword verifies a password using the default hasher
func VerifyPassword(password, hash string) (bool, error) {
	return defaultHasher.VerifyPassword(password, hash)
}