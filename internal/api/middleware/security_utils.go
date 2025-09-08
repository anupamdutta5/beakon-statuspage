package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// GenerateSecureSecret generates a secure random secret
func GenerateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GetOrGenerateSecret gets a secret from environment or generates a new one
func GetOrGenerateSecret(envVar string, defaultLength int) (string, error) {
	secret := os.Getenv(envVar)
	if secret == "" {
		generated, err := GenerateSecureSecret(defaultLength)
		if err != nil {
			return "", fmt.Errorf("failed to generate secret for %s: %w", envVar, err)
		}
		return generated, nil
	}
	return secret, nil
}

// ValidateSecretStrength validates if a secret meets security requirements
func ValidateSecretStrength(secret string, minLength int) error {
	if len(secret) < minLength {
		return fmt.Errorf("secret must be at least %d characters long", minLength)
	}

	// Check for common weak secrets
	weakSecrets := []string{
		"password", "123456", "admin", "secret", "changeme",
		"default", "test", "demo", "example", "temp",
	}

	secretLower := strings.ToLower(secret)
	for _, weak := range weakSecrets {
		if strings.Contains(secretLower, weak) {
			return fmt.Errorf("secret contains weak pattern: %s", weak)
		}
	}

	return nil
}

// MaskSecret masks a secret for logging purposes
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey(prefix string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	key := hex.EncodeToString(bytes)
	if prefix != "" {
		return fmt.Sprintf("%s_%s", prefix, key), nil
	}
	return key, nil
}

// GenerateSessionID generates a secure session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// SanitizeFilename sanitizes a filename to prevent path traversal
func SanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "~", "")

	// Remove control characters
	var result strings.Builder
	for _, char := range filename {
		if char >= 32 && char <= 126 {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	// Basic email validation
	if len(email) < 5 || !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]
	if len(local) == 0 || len(domain) == 0 {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// ValidatePasswordStrength validates password strength
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= 33 && char <= 126 && !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// IsSecureEnvironment checks if the environment is secure
func IsSecureEnvironment() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "production" || env == "staging"
}

// GetSecureDefaults returns secure default values based on environment
func GetSecureDefaults() map[string]interface{} {
	defaults := map[string]interface{}{
		"jwt_secret_length":     64,
		"session_secret_length": 32,
		"api_key_length":        32,
		"min_password_length":   8,
		"max_login_attempts":    5,
		"session_timeout":       3600, // 1 hour
		"rate_limit_enabled":    true,
		"csrf_protection":       true,
		"secure_cookies":        IsSecureEnvironment(),
		"https_only":            IsSecureEnvironment(),
	}

	return defaults
}
