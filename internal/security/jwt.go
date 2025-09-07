package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey     string        `json:"secret_key"`
	PublicKey     string        `json:"public_key"`
	PrivateKey    string        `json:"private_key"`
	Expiration    time.Duration `json:"expiration"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
	Issuer        string        `json:"issuer"`
	Audience      string        `json:"audience"`
	Algorithm     string        `json:"algorithm"`
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID    string   `json:"user_id"`
	TenantID  string   `json:"tenant_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	Scopes    []string `json:"scopes"`
	SessionID string   `json:"session_id"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	config     JWTConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config JWTConfig) (*JWTManager, error) {
	manager := &JWTManager{
		config: config,
	}

	// Set default values
	if config.Expiration == 0 {
		config.Expiration = 15 * time.Minute
	}
	if config.RefreshExpiry == 0 {
		config.RefreshExpiry = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "statuspage"
	}
	if config.Audience == "" {
		config.Audience = "statuspage-api"
	}
	if config.Algorithm == "" {
		config.Algorithm = "RS256"
	}

	manager.config = config

	// Load or generate keys
	if config.PrivateKey != "" && config.PublicKey != "" {
		err := manager.loadKeys()
		if err != nil {
			return nil, fmt.Errorf("failed to load keys: %w", err)
		}
	} else {
		err := manager.generateKeys()
		if err != nil {
			return nil, fmt.Errorf("failed to generate keys: %w", err)
		}
	}

	return manager, nil
}

// loadKeys loads RSA keys from PEM format
func (jm *JWTManager) loadKeys() error {
	// Load private key
	block, _ := pem.Decode([]byte(jm.config.PrivateKey))
	if block == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	jm.privateKey = privateKey

	// Load public key
	block, _ = pem.Decode([]byte(jm.config.PublicKey))
	if block == nil {
		return fmt.Errorf("failed to decode public key PEM")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not RSA")
	}

	jm.publicKey = rsaPublicKey

	return nil
}

// generateKeys generates new RSA key pair
func (jm *JWTManager) generateKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	jm.privateKey = privateKey
	jm.publicKey = &privateKey.PublicKey

	return nil
}

// GenerateToken generates a new JWT token
func (jm *JWTManager) GenerateToken(claims JWTClaims) (string, error) {
	// Set standard claims
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    jm.config.Issuer,
		Audience:  []string{jm.config.Audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.config.Expiration)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        claims.SessionID,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign token
	tokenString, err := token.SignedString(jm.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token
func (jm *JWTManager) GenerateRefreshToken(userID, sessionID string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.config.Issuer,
			Audience:  []string{jm.config.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.config.RefreshExpiry)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        sessionID,
		},
	}

	return jm.GenerateToken(claims)
}

// ValidateToken validates a JWT token
func (jm *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validate issuer
	if claims.Issuer != jm.config.Issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	// Validate audience
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == jm.config.Audience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, fmt.Errorf("invalid audience")
	}

	return claims, nil
}

// RefreshToken refreshes a JWT token
func (jm *JWTManager) RefreshToken(refreshToken string, newClaims JWTClaims) (string, error) {
	// Validate refresh token
	claims, err := jm.ValidateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is a refresh token (no roles/scopes)
	if len(claims.Roles) > 0 || len(claims.Scopes) > 0 {
		return "", fmt.Errorf("token is not a refresh token")
	}

	// Generate new token with updated claims
	newClaims.UserID = claims.UserID
	newClaims.SessionID = claims.SessionID

	return jm.GenerateToken(newClaims)
}

// RevokeToken revokes a token (in a real implementation, this would be stored in a blacklist)
func (jm *JWTManager) RevokeToken(tokenString string) error {
	// In a real implementation, you would:
	// 1. Parse the token to get the session ID
	// 2. Add the session ID to a blacklist/revocation list
	// 3. Store the blacklist in Redis or database

	// For now, we'll just validate the token to ensure it's valid
	_, err := jm.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// TODO: Implement token revocation logic
	return nil
}

// GetPublicKey returns the public key in PEM format
func (jm *JWTManager) GetPublicKey() (string, error) {
	if jm.publicKey == nil {
		return "", fmt.Errorf("public key not available")
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(jm.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// GetPrivateKey returns the private key in PEM format
func (jm *JWTManager) GetPrivateKey() (string, error) {
	if jm.privateKey == nil {
		return "", fmt.Errorf("private key not available")
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(jm.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

// ExtractClaimsFromToken extracts claims from a token without validation
func (jm *JWTManager) ExtractClaimsFromToken(tokenString string) (*JWTClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

// IsTokenExpired checks if a token is expired
func (jm *JWTManager) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := jm.ExtractClaimsFromToken(tokenString)
	if err != nil {
		return true, err
	}

	if claims.ExpiresAt == nil {
		return true, fmt.Errorf("token has no expiration")
	}

	return time.Now().After(claims.ExpiresAt.Time), nil
}

// GetTokenExpiration returns the token expiration time
func (jm *JWTManager) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := jm.ExtractClaimsFromToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, fmt.Errorf("token has no expiration")
	}

	return claims.ExpiresAt.Time, nil
}

// Global JWT manager instance
var globalJWTManager *JWTManager

// InitGlobalJWT initializes the global JWT manager
func InitGlobalJWT(config JWTConfig) error {
	manager, err := NewJWTManager(config)
	if err != nil {
		return err
	}
	globalJWTManager = manager
	return nil
}

// GetGlobalJWT returns the global JWT manager
func GetGlobalJWT() *JWTManager {
	return globalJWTManager
}

// GenerateToken is a convenience function for generating tokens
func GenerateToken(claims JWTClaims) (string, error) {
	if globalJWTManager != nil {
		return globalJWTManager.GenerateToken(claims)
	}
	return "", fmt.Errorf("JWT manager not initialized")
}

// ValidateToken is a convenience function for validating tokens
func ValidateToken(tokenString string) (*JWTClaims, error) {
	if globalJWTManager != nil {
		return globalJWTManager.ValidateToken(tokenString)
	}
	return nil, fmt.Errorf("JWT manager not initialized")
}

// RefreshToken is a convenience function for refreshing tokens
func RefreshToken(refreshToken string, newClaims JWTClaims) (string, error) {
	if globalJWTManager != nil {
		return globalJWTManager.RefreshToken(refreshToken, newClaims)
	}
	return "", fmt.Errorf("JWT manager not initialized")
}
