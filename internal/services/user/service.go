package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService handles user-related operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// ServeHTTP implements the http.Handler interface
func (s *UserService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Route requests based on path and method
	switch {
	case r.URL.Path == "/health":
		s.handleHealthCheck(w, r)
	case r.URL.Path == "/admin/users" && r.Method == "GET":
		s.handleGetUsers(w, r)
	case r.URL.Path == "/admin/users" && r.Method == "POST":
		s.handleCreateUser(w, r)
	case r.URL.Path == "/admin/users" && r.Method == "PUT":
		s.handleUpdateUser(w, r)
	case r.URL.Path == "/admin/users" && r.Method == "DELETE":
		s.handleDeleteUser(w, r)
	case r.URL.Path == "/admin/users" && r.Method == "GET":
		s.handleGetUser(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleHealthCheck handles health check requests
func (s *UserService) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "user-service",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	json.NewEncoder(w).Encode(response)
}

// handleGetUsers handles GET /admin/users
func (s *UserService) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get users from database
	var users []models.User
	var total int64

	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		logger.Log.Error("Failed to count users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := s.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		logger.Log.Error("Failed to get users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleCreateUser handles POST /admin/users
func (s *UserService) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Email == "" || user.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Create user
	if err := s.db.Create(&user).Error; err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return created user (without password)
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// handleGetUser handles GET /admin/users/:id
func (s *UserService) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID := r.URL.Path[len("/admin/users/"):]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get user from database
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Log.Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Remove password from response
	user.Password = ""

	// Return user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleUpdateUser handles PUT /admin/users/:id
func (s *UserService) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID := r.URL.Path[len("/admin/users/"):]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get existing user
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Log.Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse update data
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update user
	updateData["updated_at"] = time.Now()
	if err := s.db.Model(&user).Updates(updateData).Error; err != nil {
		logger.Log.Error("Failed to update user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get updated user
	if err := s.db.First(&user, userID).Error; err != nil {
		logger.Log.Error("Failed to get updated user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Remove password from response
	user.Password = ""

	// Return updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleDeleteUser handles DELETE /admin/users/:id
func (s *UserService) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID := r.URL.Path[len("/admin/users/"):]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Log.Error("Failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete user
	if err := s.db.Delete(&user).Error; err != nil {
		logger.Log.Error("Failed to delete user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "User deleted successfully",
		"user_id": userID,
	}
	json.NewEncoder(w).Encode(response)
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	// In a real implementation, this would use bcrypt
	// For now, return a placeholder
	return fmt.Sprintf("hashed_%s", password), nil
}
