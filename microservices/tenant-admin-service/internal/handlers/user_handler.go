// Package handlers provides HTTP request handlers for the Tenant Admin Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/anupamdutta5/tenant-admin-service/internal/services"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	service *services.TenantAdminService
	logger  *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *services.TenantAdminService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"` // admin, user, manager, viewer
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	IsActive  *bool   `json:"is_active"`
	Role      *string `json:"role"`
}

// GetUsers retrieves all users for a tenant
// GET /api/v1/tenants/:tenant_id/users
func (h *UserHandler) GetUsers(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Fetch users
	users, total, err := h.service.GetUsers(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get users",
			zap.String("tenant_id", tenantIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUser retrieves a single user by ID
// GET /api/v1/tenants/:tenant_id/users/:user_id
func (h *UserHandler) GetUser(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Fetch user
	user, err := h.service.GetUserByID(c.Request.Context(), tenantID, uint(userID))
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to get user",
			zap.String("tenant_id", tenantIDStr),
			zap.String("user_id", userIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUser creates a new user for a tenant
// POST /api/v1/tenants/:tenant_id/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	user, err := h.service.CreateUser(
		c.Request.Context(),
		tenantID,
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		req.Role,
	)

	if err != nil {
		// Check if it's a max_users limit error
		if err.Error() == "tenant has reached maximum user limit" ||
		   err.Error() == "tenant has reached maximum user limit of "+strconv.Itoa(int(tenantID.ID()))+" users" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "MAX_USERS_EXCEEDED",
			})
			return
		}

		h.logger.Error("Failed to create user",
			zap.String("tenant_id", tenantIDStr),
			zap.String("email", req.Email),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    user,
	})
}

// UpdateUser updates user information
// PUT /api/v1/tenants/:tenant_id/users/:user_id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update user
	if err := h.service.UpdateUser(c.Request.Context(), tenantID, uint(userID), updates); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to update user",
			zap.String("tenant_id", tenantIDStr),
			zap.String("user_id", userIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser soft-deletes a user
// DELETE /api/v1/tenants/:tenant_id/users/:user_id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete user
	if err := h.service.DeleteUser(c.Request.Context(), tenantID, uint(userID)); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.Error("Failed to delete user",
			zap.String("tenant_id", tenantIDStr),
			zap.String("user_id", userIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetUserStats retrieves user count and limit statistics for a tenant
// GET /api/v1/users/:tenant_id/stats
func (h *UserHandler) GetUserStats(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Use the existing GetTenantUserStats method
	stats, err := h.service.GetTenantUserStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant user stats",
			zap.String("tenant_id", tenantIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
