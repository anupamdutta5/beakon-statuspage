package services

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService() *AuditService {
	return &AuditService{
		db: database.GetDB(),
	}
}

// AuditAction represents different types of audit actions
const (
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionLogin    = "login"
	ActionLogout   = "logout"
	ActionView     = "view"
	ActionExport   = "export"
	ActionImport   = "import"
)

// AuditResource represents different types of resources
const (
	ResourceIncident     = "incident"
	ResourceMaintenance  = "maintenance"
	ResourceService      = "service"
	ResourceMonitor      = "monitor"
	ResourceUser         = "user"
	ResourceSubscriber   = "subscriber"
	ResourceBranding     = "branding"
	ResourceIntegration  = "integration"
	ResourceAuditLog     = "audit_log"
)

// LogAuditEvent logs an audit event
func (s *AuditService) LogAuditEvent(userID uint, action, resource string, resourceID uint, details interface{}, req *http.Request) error {
	var detailsJSON string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			logger.Error("Failed to marshal audit details", zap.Error(err))
			detailsJSON = "{}"
		} else {
			detailsJSON = string(detailsBytes)
		}
	}

	var ipAddress, userAgent string
	if req != nil {
		ipAddress = s.getClientIP(req)
		userAgent = req.UserAgent()
	}

	auditLog := &models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    detailsJSON,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		logger.Error("Failed to create audit log", zap.Error(err))
		return err
	}

	logger.Info("Audit event logged",
		zap.Uint("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.Uint("resource_id", resourceID),
		zap.String("ip_address", ipAddress),
	)

	return nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(userID *uint, action, resource string, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := s.db.Preload("User")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetAuditLogsByDateRange retrieves audit logs within a date range
func (s *AuditService) GetAuditLogsByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := s.db.Preload("User").Where("timestamp BETWEEN ? AND ?", startDate, endDate)

	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetAuditLogsByResource retrieves audit logs for a specific resource
func (s *AuditService) GetAuditLogsByResource(resource string, resourceID uint, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := s.db.Preload("User").Where("resource = ? AND resource_id = ?", resource, resourceID)

	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// getClientIP extracts the real client IP from the request
func (s *AuditService) getClientIP(req *http.Request) string {
	// Check for forwarded headers first
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := req.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	
	// Fall back to RemoteAddr
	return req.RemoteAddr
}

// AuditMiddleware creates a middleware function for logging requests
func (s *AuditService) AuditMiddleware(action, resource string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context (assuming it's set by auth middleware)
			userID, ok := r.Context().Value("user_id").(uint)
			if !ok {
				userID = 0 // Anonymous user
			}

			// Log the audit event
			go func() {
				if err := s.LogAuditEvent(userID, action, resource, 0, nil, r); err != nil {
					logger.Error("Failed to log audit event", zap.Error(err))
				}
			}()

			next(w, r)
		}
	}
}
