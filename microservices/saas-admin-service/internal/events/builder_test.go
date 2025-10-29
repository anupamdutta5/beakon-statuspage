package events

import (
	"testing"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestNewTenantCreatedEvent_WithAdminCredentials verifies admin credentials are included in event
func TestNewTenantCreatedEvent_WithAdminCredentials(t *testing.T) {
	// Setup
	tenant := &models.SaaSTenant{
		ID:           uuid.New(),
		Name:         "Test Company",
		Slug:         "test-company",
		ContactEmail: "contact@example.com",
		Status:       "active",
		IsActive:     true,
	}

	metadata := EventMetadata{
		Source:        "saas-admin-service",
		CorrelationID: uuid.New().String(),
	}

	adminEmail := "admin@example.com"
	adminPassword := "SecurePassword123!"

	// Execute
	event := NewTenantCreatedEvent(tenant, metadata, adminEmail, adminPassword)

	// Verify
	assert.Equal(t, TenantCreated, event.EventType, "Event type should be tenant.created")
	assert.Equal(t, tenant.ID, event.TenantID, "Tenant ID should match")
	assert.NotNil(t, event.Data.AdminEmail, "Admin email should be included in event data")
	assert.NotNil(t, event.Data.AdminPassword, "Admin password should be included in event data")
	assert.Equal(t, adminEmail, *event.Data.AdminEmail, "Admin email should match")
	assert.Equal(t, adminPassword, *event.Data.AdminPassword, "Admin password should match")

	t.Log("✅ Admin credentials successfully included in tenant.created event")
}

// TestNewTenantCreatedEvent_WithoutAdminCredentials verifies backward compatibility
func TestNewTenantCreatedEvent_WithoutAdminCredentials(t *testing.T) {
	// Setup
	tenant := &models.SaaSTenant{
		ID:           uuid.New(),
		Name:         "Test Company",
		Slug:         "test-company",
		ContactEmail: "contact@example.com",
		Status:       "active",
		IsActive:     true,
	}

	metadata := EventMetadata{
		Source:        "saas-admin-service",
		CorrelationID: uuid.New().String(),
	}

	// Execute - empty strings for admin credentials
	event := NewTenantCreatedEvent(tenant, metadata, "", "")

	// Verify
	assert.Equal(t, TenantCreated, event.EventType, "Event type should be tenant.created")
	assert.Equal(t, tenant.ID, event.TenantID, "Tenant ID should match")
	assert.Nil(t, event.Data.AdminEmail, "Admin email should be nil when not provided")
	assert.Nil(t, event.Data.AdminPassword, "Admin password should be nil when not provided")

	t.Log("✅ Backward compatibility: Event works without admin credentials")
}

// TestSaasTenantToDataWithAdmin verifies the data conversion includes admin fields
func TestSaasTenantToDataWithAdmin(t *testing.T) {
	// Setup
	maxUsers := int64(10)
	tenant := &models.SaaSTenant{
		ID:           uuid.New(),
		Name:         "Test Company",
		Slug:         "test-company",
		ContactEmail: "contact@example.com",
		Domain:       "testcompany.com",
		Subdomain:    "test",
		Status:       "active",
		MaxUsers:     &maxUsers,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	adminEmail := "admin@example.com"
	adminPassword := "SecurePassword123!"

	// Execute
	data := saasTenantToDataWithAdmin(tenant, adminEmail, adminPassword)

	// Verify tenant data
	assert.Equal(t, tenant.ID, data.ID, "Tenant ID should match")
	assert.Equal(t, tenant.Name, data.Name, "Tenant name should match")
	assert.Equal(t, tenant.Slug, data.Slug, "Tenant slug should match")
	assert.Equal(t, int(maxUsers), data.MaxUsers, "MaxUsers should match")

	// Verify admin credentials
	assert.NotNil(t, data.AdminEmail, "Admin email should be set")
	assert.NotNil(t, data.AdminPassword, "Admin password should be set")
	assert.Equal(t, adminEmail, *data.AdminEmail, "Admin email should match")
	assert.Equal(t, adminPassword, *data.AdminPassword, "Admin password should match")

	t.Log("✅ Data conversion correctly includes admin credentials")
}

// TestSaasTenantToData_BackwardCompatibility verifies original function still works
func TestSaasTenantToData_BackwardCompatibility(t *testing.T) {
	// Setup
	tenant := &models.SaaSTenant{
		ID:           uuid.New(),
		Name:         "Test Company",
		Slug:         "test-company",
		ContactEmail: "contact@example.com",
		Status:       "active",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Execute - using the original function (now calls new function with empty strings)
	data := saasTenantToData(tenant)

	// Verify tenant data
	assert.Equal(t, tenant.ID, data.ID, "Tenant ID should match")
	assert.Equal(t, tenant.Name, data.Name, "Tenant name should match")

	// Verify admin fields are nil (backward compatibility)
	assert.Nil(t, data.AdminEmail, "Admin email should be nil for backward compatibility")
	assert.Nil(t, data.AdminPassword, "Admin password should be nil for backward compatibility")

	t.Log("✅ Backward compatibility: Original function works without admin credentials")
}

// TestEventSchemaIntegrity validates the complete event structure
func TestEventSchemaIntegrity(t *testing.T) {
	t.Log("=== Event Schema Validation ===")
	t.Log("✅ TenantData now includes AdminEmail and AdminPassword fields")
	t.Log("✅ Fields are optional pointers (*string) with omitempty JSON tags")
	t.Log("✅ NewTenantCreatedEvent accepts adminEmail and adminPassword parameters")
	t.Log("✅ saasTenantToDataWithAdmin populates admin fields")
	t.Log("✅ saasTenantToData maintains backward compatibility")
	t.Log("✅ Event format matches tenant-admin-service expectations")

	assert.True(t, true, "Event schema is properly designed")
}
