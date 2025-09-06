package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RBACService struct {
	db *gorm.DB
}

func NewRBACService() *RBACService {
	return &RBACService{
		db: database.GetDB(),
	}
}

// Permission constants
const (
	PermissionCreateIncident     = "incidents:create"
	PermissionUpdateIncident     = "incidents:update"
	PermissionDeleteIncident     = "incidents:delete"
	PermissionCreateMaintenance  = "maintenance:create"
	PermissionUpdateMaintenance  = "maintenance:update"
	PermissionDeleteMaintenance  = "maintenance:delete"
	PermissionCreateService      = "services:create"
	PermissionUpdateService      = "services:update"
	PermissionDeleteService      = "services:delete"
	PermissionCreateMonitor      = "monitors:create"
	PermissionUpdateMonitor      = "monitors:update"
	PermissionDeleteMonitor      = "monitors:delete"
	PermissionManageUsers        = "users:manage"
	PermissionManageBranding     = "branding:manage"
	PermissionViewAuditLogs      = "audit:view"
	PermissionManageIntegrations = "integrations:manage"
)

// Role permissions mapping
var rolePermissions = map[string][]string{
	models.RoleAdmin: {
		PermissionCreateIncident,
		PermissionUpdateIncident,
		PermissionDeleteIncident,
		PermissionCreateMaintenance,
		PermissionUpdateMaintenance,
		PermissionDeleteMaintenance,
		PermissionCreateService,
		PermissionUpdateService,
		PermissionDeleteService,
		PermissionCreateMonitor,
		PermissionUpdateMonitor,
		PermissionDeleteMonitor,
		PermissionManageUsers,
		PermissionManageBranding,
		PermissionViewAuditLogs,
		PermissionManageIntegrations,
	},
	models.RoleEditor: {
		PermissionCreateIncident,
		PermissionUpdateIncident,
		PermissionDeleteIncident,
		PermissionCreateMaintenance,
		PermissionUpdateMaintenance,
		PermissionDeleteMaintenance,
		PermissionCreateService,
		PermissionUpdateService,
		PermissionDeleteService,
		PermissionCreateMonitor,
		PermissionUpdateMonitor,
		PermissionDeleteMonitor,
	},
	models.RoleViewer: {
		// Viewers have read-only access, no specific permissions needed
	},
}

// HasPermission checks if a user has a specific permission based on their role
func (s *RBACService) HasPermission(userID uint, permission string) bool {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		logger.Error("Failed to find user for permission check", zap.Error(err))
		return false
	}

	permissions, exists := rolePermissions[user.Role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetUserRole returns the role of a user
func (s *RBACService) GetUserRole(userID uint) (string, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", err
	}
	return user.Role, nil
}

// UpdateUserRole updates a user's role
func (s *RBACService) UpdateUserRole(userID uint, newRole string) error {
	// Validate role
	if _, exists := rolePermissions[newRole]; !exists {
		return gorm.ErrInvalidData
	}

	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", newRole).Error
}

// GetRolePermissions returns all permissions for a given role
func (s *RBACService) GetRolePermissions(role string) []string {
	permissions, exists := rolePermissions[role]
	if !exists {
		return []string{}
	}
	return permissions
}

// GetAllRoles returns all available roles
func (s *RBACService) GetAllRoles() []string {
	roles := make([]string, 0, len(rolePermissions))
	for role := range rolePermissions {
		roles = append(roles, role)
	}
	return roles
}

// CreateUser creates a new user with the specified role
func (s *RBACService) CreateUser(username, password, role string) (*models.User, error) {
	// Validate role
	if _, exists := rolePermissions[role]; !exists {
		return nil, gorm.ErrInvalidData
	}

	user := &models.User{
		Username: username,
		Password: password,
		Role:     role,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers returns all users with their roles
func (s *RBACService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteUser deletes a user
func (s *RBACService) DeleteUser(userID uint) error {
	return s.db.Delete(&models.User{}, userID).Error
}
