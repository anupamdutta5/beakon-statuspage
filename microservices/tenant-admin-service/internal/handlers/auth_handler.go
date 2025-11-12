package handlers

import (
	"github.com/google/uuid"

	"errors"
	"net/http"
	"os"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// GetLoginPage renders the login page.
func (h *TenantAdminHandler) GetLoginPage(c *gin.Context) {
	// Check if tenant context exists - only allow login for active tenants
	_, hasTenant := middleware.GetTenantID(c)

	// If tenant doesn't exist, show error
	if !hasTenant {
		h.logger.Warn("Tenant not found - login access denied",
			zap.String("host", c.Request.Host),
			zap.String("path", c.Request.URL.Path))

		c.HTML(http.StatusNotFound, "login.html", gin.H{
			"title": "Tenant Not Found",
			"error": "Tenant not found",
		})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Status Page Admin - Login",
	})
}

// Login handles user login.
func (h *TenantAdminHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Email      string `json:"email" binding:"required,email"`
		Password   string `json:"password" binding:"required"`
		RememberMe bool   `json:"remember_me"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get tenant context
	tenantID, hasTenantID := c.Get("tenant_id")
	if !hasTenantID {
		RespondWithError(c, http.StatusBadRequest, "Tenant context not found")
		return
	}

	// Authenticate user with database
	user, err := h.service.AuthenticateUser(c.Request.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		h.logger.Warn("Authentication failed",
			zap.String("email", loginRequest.Email),
			zap.Error(err))
		// Use typed error checking for better error handling
		if errors.Is(err, services.ErrUnauthorized) {
			RespondWithError(c, http.StatusUnauthorized, "Invalid credentials")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Authentication service error")
		}
		return
	}

	// Generate JWT access token (short-lived: 15 minutes)
	accessTokenExpiry := time.Minute * 15

	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(accessTokenExpiry).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		h.logger.Error("JWT_SECRET not configured")
		RespondWithError(c, http.StatusInternalServerError, "Authentication service configuration error")
		return
	}

	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Generate refresh token (long-lived: 7 days)
	tenantIDUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		h.logger.Error("Invalid tenant_id type", zap.Any("tenant_id", tenantID))
		RespondWithError(c, http.StatusInternalServerError, "Invalid tenant context")
		return
	}

	refreshToken, err := h.sessionService.CreateRefreshToken(
		c.Request.Context(),
		user.ID,
		tenantIDUUID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		h.logger.Error("Failed to create refresh token", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.String("user_id", user.ID.String()),
		zap.Any("tenant_id", tenantID))

	// Create RBAC session automatically for protected route access
	var sessionID string
	session, err := h.rbacService.CreateSession(c.Request.Context(), user.ID, tenantIDUUID, c.ClientIP(), c.Request.UserAgent(), accessTokenExpiry)
	if err != nil {
		h.logger.Warn("Failed to create session (non-critical)",
			zap.Error(err),
			zap.String("user_id", user.ID.String()))
		// Don't fail login if session creation fails - log warning and continue
	} else {
		sessionID = session.ID
		h.logger.Info("Session created automatically",
			zap.String("session_id", sessionID),
			zap.String("user_id", user.ID.String()))
	}

	// Set auth token cookie with comprehensive security attributes
	// Security features:
	// - HttpOnly: Prevents XSS attacks (JavaScript cannot access cookie)
	// - SameSite=Lax: Prevents CSRF attacks while allowing normal navigation
	// - Secure: HTTPS only in production environment
	// - Max-Age: Matches access token expiry (15 minutes)
	cookieMaxAge := int(accessTokenExpiry.Seconds())

	// Use Gin's SetCookie method properly
	secure := false
	if os.Getenv("ENVIRONMENT") == "production" {
		secure = true
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token",              // name
		accessToken,               // value
		cookieMaxAge,              // maxAge
		"/",                       // path
		"",                        // domain
		secure,                    // secure
		true,                      // httpOnly
	)

	// Set refresh token cookie (long-lived, 7 days)
	refreshCookieMaxAge := 7 * 24 * 60 * 60 // 7 days in seconds
	c.SetCookie(
		"refresh_token",           // name
		refreshToken,              // value
		refreshCookieMaxAge,       // maxAge
		"/",                       // path
		"",                        // domain
		secure,                    // secure
		true,                      // httpOnly
	)

	// Return access token, refresh token, and user info
	response := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int(accessTokenExpiry.Seconds()), // 900 seconds (15 minutes)
		"token_type":    "Bearer",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"tenant_id": tenantID,
		},
	}

	// Include session_id if created successfully
	if sessionID != "" {
		response["session_id"] = sessionID
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout.
func (h *TenantAdminHandler) Logout(c *gin.Context) {
	// Get refresh token from cookie or request body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Try to get from request body
		var logoutRequest struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&logoutRequest); err == nil && logoutRequest.RefreshToken != "" {
			refreshToken = logoutRequest.RefreshToken
		}
	}

	// Revoke refresh token if present
	if refreshToken != "" {
		if err := h.sessionService.RevokeRefreshToken(c.Request.Context(), refreshToken); err != nil {
			h.logger.Warn("Failed to revoke refresh token during logout",
				zap.Error(err),
				zap.String("token_prefix", refreshToken[:min(10, len(refreshToken))]+"..."))
			// Continue with logout even if revocation fails
		} else {
			h.logger.Info("Refresh token revoked during logout",
				zap.String("token_prefix", refreshToken[:min(10, len(refreshToken))]+"..."))
		}
	}

	// Clear the auth_token cookie by setting MaxAge to -1
	c.SetCookie(
		"auth_token",  // name
		"",            // value (empty)
		-1,            // maxAge (-1 deletes the cookie)
		"/",           // path
		"",            // domain
		false,         // secure
		true,          // httpOnly
	)

	// Clear the refresh_token cookie
	c.SetCookie(
		"refresh_token", // name
		"",              // value (empty)
		-1,              // maxAge (-1 deletes the cookie)
		"/",             // path
		"",              // domain
		false,           // secure
		true,            // httpOnly
	)

	h.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// min is a helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RefreshToken handles refresh token rotation and access token renewal.
// This endpoint implements token rotation security best practice:
// 1. Validates the old refresh token
// 2. Revokes the old refresh token
// 3. Generates new refresh token
// 4. Generates new access token
// 5. Returns both tokens
//
// POST /api/v1/auth/refresh
// Request Body: { "refresh_token": "base64_token..." }
// OR refresh_token cookie
//
// Response (200 OK):
//
//	{
//	  "access_token": "eyJhbGc...",
//	  "refresh_token": "new_base64_token...",
//	  "expires_in": 900,
//	  "token_type": "Bearer"
//	}
func (h *TenantAdminHandler) RefreshToken(c *gin.Context) {
	// Get refresh token from cookie or request body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Try to get from request body
		var refreshRequest struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&refreshRequest); err != nil {
			RespondWithError(c, http.StatusBadRequest, "Refresh token is required")
			return
		}
		refreshToken = refreshRequest.RefreshToken
	}

	if refreshToken == "" {
		RespondWithError(c, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Validate refresh token and get user/tenant IDs
	userID, tenantID, err := h.sessionService.ValidateRefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token",
			zap.Error(err),
			zap.String("token_prefix", refreshToken[:min(10, len(refreshToken))]+"..."))
		RespondWithError(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// Get user details for JWT claims
	user, err := h.service.GetUserByID(c.Request.Context(), tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to get user for token refresh",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("tenant_id", tenantID.String()))
		RespondWithError(c, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	// Rotate refresh token (revoke old, create new) - security best practice
	newRefreshToken, err := h.sessionService.RotateRefreshToken(
		c.Request.Context(),
		refreshToken,
		userID,
		tenantID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		h.logger.Error("Failed to rotate refresh token",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		RespondWithError(c, http.StatusInternalServerError, "Failed to rotate refresh token")
		return
	}

	// Generate new JWT access token (15 minutes)
	accessTokenExpiry := time.Minute * 15
	claims := jwt.MapClaims{
		"user_id":   userID,
		"email":     user.Email,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(accessTokenExpiry).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		h.logger.Error("JWT_SECRET not configured")
		RespondWithError(c, http.StatusInternalServerError, "Authentication service configuration error")
		return
	}

	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		h.logger.Error("Failed to generate new access token", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.logger.Info("Token refreshed successfully",
		zap.String("user_id", userID.String()),
		zap.String("tenant_id", tenantID.String()))

	// Set new cookies
	cookieMaxAge := int(accessTokenExpiry.Seconds())
	secure := false
	if os.Getenv("ENVIRONMENT") == "production" {
		secure = true
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token",
		accessToken,
		cookieMaxAge,
		"/",
		"",
		secure,
		true,
	)

	refreshCookieMaxAge := 7 * 24 * 60 * 60 // 7 days
	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		refreshCookieMaxAge,
		"/",
		"",
		secure,
		true,
	)

	// Return new tokens
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    int(accessTokenExpiry.Seconds()),
		"token_type":    "Bearer",
	})
}

// VerifyToken verifies the JWT token.
func (h *TenantAdminHandler) VerifyToken(c *gin.Context) {
	// In production, this would verify the JWT token signature and expiration
	// For now, we return a mock response for development
	h.logger.Info("Token verification requested")
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":        1,
			"email":     "admin@example.com",
			"username":  "admin",
			"role":      "admin",
			"tenant_id": 1,
		},
	})
}

// ValidateSubdomain validates if a subdomain corresponds to a valid tenant.
// This endpoint is used by the frontend middleware to check if a subdomain exists
// before allowing access to the login page.
func (h *TenantAdminHandler) ValidateSubdomain(c *gin.Context) {
	subdomain := c.Query("subdomain")
	if subdomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "subdomain parameter is required",
			"valid": false,
		})
		return
	}

	h.logger.Info("Validating subdomain", zap.String("subdomain", subdomain))

	// Check if tenant exists with this subdomain
	exists, err := h.service.ExistsBySubdomain(c.Request.Context(), subdomain)
	if err != nil {
		h.logger.Error("Failed to validate subdomain",
			zap.String("subdomain", subdomain),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate subdomain",
			"valid": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     exists,
		"subdomain": subdomain,
	})
}
