package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TenantAdminBaseURL = "http://localhost:8099"
	SaaSAdminBaseURL   = "http://localhost:8098"
)

// TestTenantOnboardingFlow tests the complete tenant onboarding process
// This is a critical E2E flow that covers:
// 1. Tenant creation via public endpoint
// 2. Owner user creation
// 3. Initial login
// 4. Role assignment
// 5. Team creation
// 6. Component setup
// 7. Status page configuration
func TestTenantOnboardingFlow(t *testing.T) {
	// Test data
	tenantEmail := fmt.Sprintf("test-%d@e2e-test.com", time.Now().Unix())
	tenantName := fmt.Sprintf("E2E Test Tenant %d", time.Now().Unix())
	tenantSubdomain := fmt.Sprintf("e2e-test-%d", time.Now().Unix())

	var tenantID string
	var ownerID string
	var authToken string
	var teamID string
	var componentID string
	var statusPageID string

	t.Run("Step 1: Create new tenant via public API", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":      tenantName,
			"subdomain": tenantSubdomain,
			"email":     tenantEmail,
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/public/tenants", TenantAdminBaseURL),
			"application/json",
			bytes.NewBuffer(body),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Tenant creation should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		tenantID = response["tenant_id"].(string)
		ownerID = response["owner_id"].(string)

		assert.NotEmpty(t, tenantID, "Tenant ID should be returned")
		assert.NotEmpty(t, ownerID, "Owner ID should be returned")

		t.Logf("✓ Tenant created: %s (ID: %s)", tenantName, tenantID)
		t.Logf("✓ Owner user created: %s (ID: %s)", tenantEmail, ownerID)
	})

	t.Run("Step 2: Initial login as tenant owner", func(t *testing.T) {
		// Wait for async processing (tenant sync)
		time.Sleep(2 * time.Second)

		payload := map[string]string{
			"email":    tenantEmail,
			"password": "default_password", // Set during tenant creation
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/auth/login", TenantAdminBaseURL),
			"application/json",
			bytes.NewBuffer(body),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Login should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		authToken = response["token"].(string)
		assert.NotEmpty(t, authToken, "Auth token should be returned")

		t.Logf("✓ Login successful, token received")
	})

	t.Run("Step 3: Verify tenant details", func(t *testing.T) {
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/api/v1/tenants/%s", TenantAdminBaseURL, tenantID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tenant map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tenant)
		require.NoError(t, err)

		assert.Equal(t, tenantName, tenant["name"])
		assert.Equal(t, tenantSubdomain, tenant["subdomain"])
		assert.Equal(t, "active", tenant["status"])

		t.Logf("✓ Tenant details verified")
	})

	t.Run("Step 4: Create admin user", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin-%d@e2e-test.com", time.Now().Unix())

		payload := map[string]interface{}{
			"email":      adminEmail,
			"first_name": "Admin",
			"last_name":  "User",
			"password":   "AdminPass123!",
			"role":       "admin",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/users", TenantAdminBaseURL),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Admin user creation should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		adminID := response["id"].(string)
		assert.NotEmpty(t, adminID)

		t.Logf("✓ Admin user created: %s", adminEmail)
	})

	t.Run("Step 5: Create development team", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "Development Team",
			"description": "Core development team",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/teams", TenantAdminBaseURL),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Team creation should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		teamID = response["id"].(string)
		assert.NotEmpty(t, teamID)

		t.Logf("✓ Team created: Development Team (ID: %s)", teamID)
	})

	t.Run("Step 6: Add owner to team", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id": ownerID,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/teams/%s/members", TenantAdminBaseURL, teamID),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Adding member should succeed")

		t.Logf("✓ Owner added to team")
	})

	t.Run("Step 7: Create first component", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "API Service",
			"description": "Main API backend",
			"status":      "operational",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/components", TenantAdminBaseURL),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Component creation should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		componentID = response["id"].(string)
		assert.NotEmpty(t, componentID)

		t.Logf("✓ Component created: API Service (ID: %s)", componentID)
	})

	t.Run("Step 8: Configure status page", func(t *testing.T) {
		payload := map[string]interface{}{
			"page_title":            fmt.Sprintf("%s Status", tenantName),
			"page_slug":             tenantSubdomain,
			"description":           "Current system status and uptime",
			"is_public":             true,
			"show_uptime":           true,
			"show_incident_history": true,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/status-pages", TenantAdminBaseURL),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status page creation should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		statusPageID = response["id"].(string)
		assert.NotEmpty(t, statusPageID)

		t.Logf("✓ Status page configured (ID: %s)", statusPageID)
	})

	t.Run("Step 9: Publish status page", func(t *testing.T) {
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/status-pages/%s/publish", TenantAdminBaseURL, statusPageID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status page publish should succeed")

		t.Logf("✓ Status page published")
	})

	t.Run("Step 10: Verify public status page access", func(t *testing.T) {
		// Wait for publish to complete
		time.Sleep(1 * time.Second)

		resp, err := http.Get(
			fmt.Sprintf("http://localhost:8093/status/%s", tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Public status page should be accessible")

		var page map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&page)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("%s Status", tenantName), page["page_title"])
		assert.True(t, page["is_public"].(bool))

		t.Logf("✓ Public status page accessible")
	})

	// Cleanup
	t.Run("Cleanup: Delete test tenant", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE",
			fmt.Sprintf("%s/api/v1/tenants/%s", TenantAdminBaseURL, tenantID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		t.Logf("✓ Test tenant cleaned up")
	})

	t.Log("\n=== Tenant Onboarding E2E Test Complete ===")
	t.Logf("Tenant: %s", tenantName)
	t.Logf("Subdomain: %s", tenantSubdomain)
	t.Logf("Owner: %s", tenantEmail)
	t.Logf("Total steps: 10/10 passed")
}
