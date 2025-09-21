// Package services provides RBAC (Role-Based Access Control) business logic for the Tenant Admin Service.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RBACService provides role-based access control functionality
type RBACService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewRBACService creates a new RBAC service instance
func NewRBACService(db *gorm.DB, logger *zap.Logger) *RBACService {
	return &RBACService{
		db:     db,
		logger: logger,
	}
}

// Role Management

// CreateRole creates a new role
func (s *RBACService) CreateRole(ctx context.Context, role *models.Role) error {
	// Check if role name already exists for this tenant
	var existing models.Role
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND name = ?", role.TenantID, role.Name).First(&existing).Error
	if err == nil {
		return errors.New("role with this name already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing role: %w", err)
	}

	if err := s.db.WithContext(ctx).Create(role).Error; err != nil {
		s.logger.Error("Failed to create role", zap.Error(err))
		return fmt.Errorf("failed to create role: %w", err)
	}

	s.logger.Info("Role created successfully",
		zap.Uint("role_id", role.ID),
		zap.String("role_name", role.Name),
		zap.Uint("tenant_id", role.TenantID))

	return nil
}

// GetRoles retrieves roles for a tenant
func (s *RBACService) GetRoles(ctx context.Context, tenantID uint, includeInactive bool) ([]models.Role, error) {
	var roles []models.Role
	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Preload("Permissions").Find(&roles).Error
	if err != nil {
		s.logger.Error("Failed to get roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	// Add computed fields
	for i := range roles {
		s.db.Model(&models.Permission{}).Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
			Where("role_permissions.role_id = ?", roles[i].ID).Count(&roles[i].PermissionCount)

		s.db.Model(&models.UserRole{}).Where("role_id = ? AND is_active = ?", roles[i].ID, true).Count(&roles[i].UserCount)
	}

	return roles, nil
}

// GetRole retrieves a specific role
func (s *RBACService) GetRole(ctx context.Context, roleID, tenantID uint) (*models.Role, error) {
	var role models.Role
	err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", roleID, tenantID).
		Preload("Permissions").First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		s.logger.Error("Failed to get role", zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// UpdateRole updates an existing role
func (s *RBACService) UpdateRole(ctx context.Context, roleID, tenantID uint, updates map[string]interface{}) error {
	result := s.db.WithContext(ctx).Model(&models.Role{}).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).Updates(updates)

	if result.Error != nil {
		s.logger.Error("Failed to update role", zap.Error(result.Error))
		return fmt.Errorf("failed to update role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("role not found")
	}

	s.logger.Info("Role updated successfully",
		zap.Uint("role_id", roleID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// DeleteRole soft deletes a role
func (s *RBACService) DeleteRole(ctx context.Context, roleID, tenantID uint) error {
	// Check if role is being used
	var userRoleCount int64
	s.db.WithContext(ctx).Model(&models.UserRole{}).Where("role_id = ? AND is_active = ?", roleID, true).Count(&userRoleCount)

	if userRoleCount > 0 {
		return errors.New("cannot delete role that is assigned to users")
	}

	result := s.db.WithContext(ctx).Model(&models.Role{}).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to delete role", zap.Error(result.Error))
		return fmt.Errorf("failed to delete role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("role not found")
	}

	s.logger.Info("Role deleted successfully",
		zap.Uint("role_id", roleID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// Permission Management

// GetPermissions retrieves all available permissions
func (s *RBACService) GetPermissions(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	err := s.db.WithContext(ctx).Find(&permissions).Error
	if err != nil {
		s.logger.Error("Failed to get permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return permissions, nil
}

// AssignPermissionsToRole assigns permissions to a role
func (s *RBACService) AssignPermissionsToRole(ctx context.Context, roleID, tenantID uint, permissionIDs []uint) error {
	// Verify role exists and belongs to tenant
	var role models.Role
	err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	// Get permissions
	var permissions []models.Permission
	err = s.db.WithContext(ctx).Where("id IN ?", permissionIDs).Find(&permissions).Error
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// Replace existing permissions
	err = s.db.WithContext(ctx).Model(&role).Association("Permissions").Replace(permissions)
	if err != nil {
		s.logger.Error("Failed to assign permissions to role", zap.Error(err))
		return fmt.Errorf("failed to assign permissions: %w", err)
	}

	s.logger.Info("Permissions assigned to role successfully",
		zap.Uint("role_id", roleID),
		zap.Int("permission_count", len(permissionIDs)))

	return nil
}

// User Role Management

// AssignRoleToUser assigns a role to a user
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID, tenantID, assignedBy uint, expiresAt *time.Time) error {
	// Check if assignment already exists
	var existing models.UserRole
	err := s.db.WithContext(ctx).Where("user_id = ? AND role_id = ? AND tenant_id = ? AND is_active = ?",
		userID, roleID, tenantID, true).First(&existing).Error
	if err == nil {
		return errors.New("user already has this role")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}

	userRole := &models.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		TenantID:   tenantID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
		ExpiresAt:  expiresAt,
		IsActive:   true,
	}

	if err := s.db.WithContext(ctx).Create(userRole).Error; err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	s.logger.Info("Role assigned to user successfully",
		zap.Uint("user_id", userID),
		zap.Uint("role_id", roleID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID, tenantID uint) error {
	result := s.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ? AND tenant_id = ?", userID, roleID, tenantID).
		Update("is_active", false)

	if result.Error != nil {
		s.logger.Error("Failed to remove role from user", zap.Error(result.Error))
		return fmt.Errorf("failed to remove role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("role assignment not found")
	}

	s.logger.Info("Role removed from user successfully",
		zap.Uint("user_id", userID),
		zap.Uint("role_id", roleID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// GetUserRoles retrieves roles for a user
func (s *RBACService) GetUserRoles(ctx context.Context, userID, tenantID uint) ([]models.UserRole, error) {
	var userRoles []models.UserRole
	err := s.db.WithContext(ctx).Where("user_id = ? AND tenant_id = ? AND is_active = ?", userID, tenantID, true).
		Preload("Role").Preload("Role.Permissions").Find(&userRoles).Error
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return userRoles, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, userID, tenantID uint) ([]string, error) {
	var permissions []string

	// Get permissions from direct role assignments
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.tenant_id = ? AND ur.is_active = true
		AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
	`

	err := s.db.WithContext(ctx).Raw(query, userID, tenantID).Scan(&permissions).Error
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Get permissions from team role assignments
	teamQuery := `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN team_roles tr ON rp.role_id = tr.role_id
		JOIN team_members tm ON tr.team_id = tm.team_id
		WHERE tm.user_id = ? AND tm.tenant_id = ? AND tm.is_active = true
		AND tr.is_active = true AND (tr.expires_at IS NULL OR tr.expires_at > NOW())
	`

	var teamPermissions []string
	err = s.db.WithContext(ctx).Raw(teamQuery, userID, tenantID).Scan(&teamPermissions).Error
	if err != nil {
		s.logger.Error("Failed to get team permissions", zap.Error(err))
		return permissions, nil // Return direct permissions even if team permissions fail
	}

	// Merge permissions
	permissionSet := make(map[string]bool)
	for _, perm := range permissions {
		permissionSet[perm] = true
	}
	for _, perm := range teamPermissions {
		permissionSet[perm] = true
	}

	result := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		result = append(result, perm)
	}

	return result, nil
}

// CheckPermission checks if a user has a specific permission
func (s *RBACService) CheckPermission(ctx context.Context, userID, tenantID uint, permission string) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permission {
			return true, nil
		}
	}

	return false, nil
}

// Team Management

// CreateTeam creates a new team
func (s *RBACService) CreateTeam(ctx context.Context, team *models.Team) error {
	if err := s.db.WithContext(ctx).Create(team).Error; err != nil {
		s.logger.Error("Failed to create team", zap.Error(err))
		return fmt.Errorf("failed to create team: %w", err)
	}

	s.logger.Info("Team created successfully",
		zap.Uint("team_id", team.ID),
		zap.String("team_name", team.Name),
		zap.Uint("tenant_id", team.TenantID))

	return nil
}

// AddUserToTeam adds a user to a team
func (s *RBACService) AddUserToTeam(ctx context.Context, teamID, userID, tenantID, addedBy uint, role string) error {
	// Check if user is already in team
	var existing models.TeamMember
	err := s.db.WithContext(ctx).Where("team_id = ? AND user_id = ? AND is_active = ?",
		teamID, userID, true).First(&existing).Error
	if err == nil {
		return errors.New("user is already a member of this team")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing membership: %w", err)
	}

	if role == "" {
		role = "member"
	}

	teamMember := &models.TeamMember{
		TeamID:   teamID,
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		AddedBy:  addedBy,
		AddedAt:  time.Now(),
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(teamMember).Error; err != nil {
		s.logger.Error("Failed to add user to team", zap.Error(err))
		return fmt.Errorf("failed to add user to team: %w", err)
	}

	s.logger.Info("User added to team successfully",
		zap.Uint("user_id", userID),
		zap.Uint("team_id", teamID),
		zap.String("role", role))

	return nil
}

// Session Management

// CreateSession creates a new user session
func (s *RBACService) CreateSession(ctx context.Context, userID, tenantID uint, ipAddress, userAgent string, duration time.Duration) (*models.Session, error) {
	// Generate secure session ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &models.Session{
		ID:        sessionID,
		UserID:    userID,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
		LastSeen:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session and updates last seen
func (s *RBACService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	var session models.Session
	err := s.db.WithContext(ctx).Where("id = ? AND is_active = ? AND expires_at > NOW()", sessionID, true).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid session")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	// Update last seen
	s.db.WithContext(ctx).Model(&session).Update("last_seen", time.Now())

	return &session, nil
}

// Audit Logging

// LogAuditEvent logs an audit event
func (s *RBACService) LogAuditEvent(ctx context.Context, tenantID, userID uint, action, resource string, resourceID *uint, details, ipAddress, userAgent string, success bool, errorMessage string) error {
	auditLog := &models.AuditLog{
		TenantID:     tenantID,
		UserID:       userID,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      success,
		ErrorMessage: errorMessage,
	}

	if err := s.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		s.logger.Error("Failed to log audit event", zap.Error(err))
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

// InitializeDefaultRoles initializes default roles and permissions for a tenant
func (s *RBACService) InitializeDefaultRoles(ctx context.Context, tenantID uint) error {
	// Create permissions if they don't exist
	for _, perm := range models.DefaultPermissions {
		var existing models.Permission
		err := s.db.WithContext(ctx).Where("name = ?", perm.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.db.WithContext(ctx).Create(&perm).Error; err != nil {
				s.logger.Error("Failed to create permission", zap.Error(err), zap.String("permission", perm.Name))
			}
		}
	}

	// Create default roles for the tenant
	for _, roleTemplate := range models.DefaultRoles {
		role := roleTemplate
		role.TenantID = tenantID

		var existing models.Role
		err := s.db.WithContext(ctx).Where("tenant_id = ? AND name = ?", tenantID, role.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.db.WithContext(ctx).Create(&role).Error; err != nil {
				s.logger.Error("Failed to create default role", zap.Error(err), zap.String("role", role.Name))
				continue
			}

			// Assign permissions to role
			if permNames, exists := models.DefaultRolePermissions[role.Name]; exists {
				var permissions []models.Permission
				s.db.WithContext(ctx).Where("name IN ?", permNames).Find(&permissions)
				s.db.WithContext(ctx).Model(&role).Association("Permissions").Append(permissions)
			}
		}
	}

	s.logger.Info("Default roles initialized for tenant", zap.Uint("tenant_id", tenantID))
	return nil
}

// Helper methods

func (s *RBACService) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}