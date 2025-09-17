// Package services provides business logic for the Status Page UI Service.
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/statuspage-status-ui-service/internal/models"
	"go.uber.org/zap"
)

// StatusPageService handles status page UI business logic (frontend only).
type StatusPageService struct {
	logger *zap.Logger
	config *Config
}

// Config represents the service configuration.
type Config struct {
	TenantAdminServiceURL string
}

// NewStatusPageService creates a new status page service.
func NewStatusPageService(logger *zap.Logger, config *Config) *StatusPageService {
	return &StatusPageService{
		logger: logger,
		config: config,
	}
}

// GetStatusPageData retrieves all data needed for rendering the status page from tenant-admin-service.
func (s *StatusPageService) GetStatusPageData(tenantID uint, slug string) (*models.StatusPageData, error) {
	// Make HTTP request to tenant-admin-service
	url := fmt.Sprintf("%s/api/v1/status-pages/%s/data?tenant_id=%d", s.config.TenantAdminServiceURL, slug, tenantID)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get status page data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tenant admin service returned status %d", resp.StatusCode)
	}

	var data models.StatusPageData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode status page data: %w", err)
	}

	s.logger.Info("Retrieved status page data", 
		zap.Uint("tenant_id", tenantID),
		zap.String("slug", slug))

	return &data, nil
}

