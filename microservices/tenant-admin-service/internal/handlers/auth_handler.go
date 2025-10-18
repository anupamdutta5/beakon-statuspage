package handlers

import (
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

	// Generate JWT token
	tokenExpiry := time.Hour * 24 // Default 24 hours
	if loginRequest.RememberMe {
		tokenExpiry = time.Hour * 24 * 30 // 30 days if remember me
	}

	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(tokenExpiry).Unix(),
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

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.Uint("user_id", user.ID),
		zap.Any("tenant_id", tenantID))

	// Create RBAC session automatically for protected route access
	var sessionID string
	if tenantIDStr, ok := tenantID.(string); ok {
		// Call RBAC service to create session
		session, err := h.rbacService.CreateSession(c.Request.Context(), user.ID, tenantIDStr, c.ClientIP(), c.Request.UserAgent(), tokenExpiry)
		if err != nil {
			h.logger.Warn("Failed to create session (non-critical)",
				zap.Error(err),
				zap.Uint("user_id", user.ID))
			// Don't fail login if session creation fails - log warning and continue
		} else {
			sessionID = session.ID
			h.logger.Info("Session created automatically",
				zap.String("session_id", sessionID),
				zap.Uint("user_id", user.ID))
		}
	}

	// Set auth token cookie with comprehensive security attributes
	// Security features:
	// - HttpOnly: Prevents XSS attacks (JavaScript cannot access cookie)
	// - SameSite=Lax: Prevents CSRF attacks while allowing normal navigation
	// - Secure: HTTPS only in production environment
	// - Max-Age: Matches JWT token expiry (respects RememberMe setting)
	cookieMaxAge := int(tokenExpiry.Seconds())

	// Use Gin's SetCookie method properly
	secure := false
	if os.Getenv("ENVIRONMENT") == "production" {
		secure = true
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token",              // name
		tokenString,               // value
		cookieMaxAge,              // maxAge
		"/",                       // path
		"",                        // domain
		secure,                    // secure
		true,                      // httpOnly
	)

	// Return JWT token and session ID
	response := gin.H{
		"token": tokenString,
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
	// In production, this would invalidate the JWT token on the server side
	// For now, we rely on client-side token removal
	h.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
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
