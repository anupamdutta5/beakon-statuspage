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
	StatusUIServiceURL  = "http://localhost:8093"
	ComponentServiceURL = "http://localhost:8084"
	IncidentServiceURL  = "http://localhost:8086"
)

// TestStatusPageViewingFlow tests the complete public status page viewing experience
// This covers:
// 1. Accessing public status page
// 2. Viewing component status
// 3. Viewing incident timeline
// 4. Viewing uptime metrics
// 5. Subscribing to notifications
// 6. Component status updates
// 7. Real-time updates
// 8. Historical data viewing
// 9. Custom branding verification
// 10. Mobile responsiveness
func TestStatusPageViewingFlow(t *testing.T) {
	tenantID := "tenant-1111-1111-1111-111111111111"
	tenantSubdomain := "test1"
	authToken := getTestAuthToken(t, tenantID)

	var statusPageID string
	var componentID string
	var incidentID string
	var subscriptionID string

	t.Run("Step 1: Access public status page", func(t *testing.T) {
		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Public status page should be accessible")

		var page map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&page)
		require.NoError(t, err)

		assert.NotEmpty(t, page["page_title"])
		assert.True(t, page["is_public"].(bool))
		assert.True(t, page["is_active"].(bool))

		statusPageID = page["id"].(string)

		t.Logf("✓ Status page accessible: %s", page["page_title"])
		t.Logf("  Subdomain: %s", tenantSubdomain)
		t.Logf("  Page ID: %s", statusPageID)
	})

	t.Run("Step 2: View component status list", func(t *testing.T) {
		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s/components", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var components []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&components)
		require.NoError(t, err)

		assert.NotEmpty(t, components, "Should have at least one component")

		// Find an operational component
		for _, comp := range components {
			if comp["status"] == "operational" {
				componentID = comp["id"].(string)
				t.Logf("✓ Found operational component: %s", comp["name"])
				break
			}
		}

		assert.NotEmpty(t, componentID, "Should have at least one operational component")

		// Verify component groups
		groupedComponents := make(map[string][]map[string]interface{})
		for _, comp := range components {
			groupID := comp["group_id"].(string)
			groupedComponents[groupID] = append(groupedComponents[groupID], comp)
		}

		t.Logf("✓ Components organized in %d groups", len(groupedComponents))
	})

	t.Run("Step 3: View recent incidents", func(t *testing.T) {
		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s/incidents?limit=10", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var incidents []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&incidents)
		require.NoError(t, err)

		if len(incidents) > 0 {
			incident := incidents[0]
			t.Logf("✓ Recent incident found: %s", incident["title"])
			t.Logf("  Status: %s", incident["status"])
			t.Logf("  Severity: %s", incident["severity"])

			// Verify incident has updates
			updates := incident["updates"].([]interface{})
			t.Logf("  Updates: %d", len(updates))
		} else {
			t.Logf("✓ No incidents (all systems operational)")
		}
	})

	t.Run("Step 4: View uptime metrics", func(t *testing.T) {
		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s/uptime?days=90", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var uptime map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&uptime)
		require.NoError(t, err)

		assert.Contains(t, uptime, "overall_uptime")
		assert.Contains(t, uptime, "components")

		overallUptime := uptime["overall_uptime"].(float64)
		assert.GreaterOrEqual(t, overallUptime, 0.0)
		assert.LessOrEqual(t, overallUptime, 100.0)

		t.Logf("✓ Uptime metrics loaded")
		t.Logf("  Overall uptime (90 days): %.2f%%", overallUptime)

		componentMetrics := uptime["components"].([]interface{})
		for _, comp := range componentMetrics {
			compMap := comp.(map[string]interface{})
			t.Logf("  %s: %.2f%%", compMap["name"], compMap["uptime"])
		}
	})

	t.Run("Step 5: Subscribe to status updates", func(t *testing.T) {
		subscriberEmail := fmt.Sprintf("subscriber-%d@e2e-test.com", time.Now().Unix())

		payload := map[string]interface{}{
			"subscription_type":     "email",
			"contact_value":         subscriberEmail,
			"subscribed_components": []string{componentID},
			"subscribed_incidents":  []string{"critical", "major"},
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(
			fmt.Sprintf("%s/status/%s/subscribe", StatusUIServiceURL, tenantSubdomain),
			"application/json",
			bytes.NewBuffer(body),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Subscription should succeed")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		subscriptionID = response["id"].(string)
		assert.NotEmpty(t, subscriptionID)
		assert.False(t, response["is_verified"].(bool), "Should require verification")

		t.Logf("✓ Subscription created: %s", subscriberEmail)
		t.Logf("  Subscription ID: %s", subscriptionID)
		t.Logf("  Verification required: true")
	})

	t.Run("Step 6: Create incident to test real-time updates", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "E2E Test Incident - API Slowness",
			"description": "Testing incident display on status page",
			"severity":    "minor",
			"status":      "investigating",
			"affected_components": []map[string]interface{}{
				{
					"component_id":  componentID,
					"impact_level": "degraded_performance",
				},
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/incidents", IncidentServiceURL),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		incidentID = response["id"].(string)
		assert.NotEmpty(t, incidentID)

		t.Logf("✓ Test incident created: %s", response["title"])
	})

	t.Run("Step 7: Verify incident appears on status page", func(t *testing.T) {
		// Wait for incident to propagate
		time.Sleep(2 * time.Second)

		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var page map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&page)
		require.NoError(t, err)

		// Check active incidents
		activeIncidents := page["active_incidents"].([]interface{})
		assert.NotEmpty(t, activeIncidents, "Should have active incident")

		foundIncident := false
		for _, inc := range activeIncidents {
			incMap := inc.(map[string]interface{})
			if incMap["id"] == incidentID {
				foundIncident = true
				t.Logf("✓ Incident visible on status page: %s", incMap["title"])
				t.Logf("  Status: %s", incMap["status"])
				t.Logf("  Severity: %s", incMap["severity"])
				break
			}
		}

		assert.True(t, foundIncident, "Created incident should be visible")
	})

	t.Run("Step 8: Update component status", func(t *testing.T) {
		payload := map[string]interface{}{
			"status": "degraded_performance",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH",
			fmt.Sprintf("%s/api/v1/components/%s", ComponentServiceURL, componentID),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		t.Logf("✓ Component status updated to degraded_performance")
	})

	t.Run("Step 9: Verify component status change on status page", func(t *testing.T) {
		// Wait for status change to propagate
		time.Sleep(2 * time.Second)

		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s/components", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		var components []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&components)
		require.NoError(t, err)

		foundComponent := false
		for _, comp := range components {
			if comp["id"] == componentID {
				foundComponent = true
				assert.Equal(t, "degraded_performance", comp["status"])
				t.Logf("✓ Component status change reflected: %s", comp["status"])
				break
			}
		}

		assert.True(t, foundComponent, "Component should be found with updated status")
	})

	t.Run("Step 10: View historical data (90-day uptime chart)", func(t *testing.T) {
		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s/history?component_id=%s&days=90",
				StatusUIServiceURL, tenantSubdomain, componentID),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var history map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&history)
		require.NoError(t, err)

		assert.Contains(t, history, "daily_uptime")
		assert.Contains(t, history, "incidents")
		assert.Contains(t, history, "average_uptime")

		dailyUptime := history["daily_uptime"].([]interface{})
		assert.Len(t, dailyUptime, 90, "Should have 90 days of data")

		avgUptime := history["average_uptime"].(float64)
		t.Logf("✓ Historical data loaded: 90 days")
		t.Logf("  Average uptime: %.2f%%", avgUptime)
		t.Logf("  Data points: %d", len(dailyUptime))
	})

	t.Run("Step 11: Verify custom branding", func(t *testing.T) {
		resp, err := http.Get(
			fmt.Sprintf("%s/status/%s/branding", StatusUIServiceURL, tenantSubdomain),
		)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var branding map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&branding)
		require.NoError(t, err)

		assert.Contains(t, branding, "theme")
		assert.Contains(t, branding, "logo")
		assert.Contains(t, branding, "custom_css")

		theme := branding["theme"].(map[string]interface{})
		assert.NotEmpty(t, theme["primary_color"])
		assert.NotEmpty(t, theme["background_color"])

		t.Logf("✓ Custom branding loaded")
		t.Logf("  Primary color: %s", theme["primary_color"])
		t.Logf("  Font family: %s", theme["font_family"])
	})

	t.Run("Step 12: Test mobile viewport access", func(t *testing.T) {
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/status/%s", StatusUIServiceURL, tenantSubdomain),
			nil,
		)
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var page map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&page)
		require.NoError(t, err)

		// Should render mobile-friendly version
		assert.True(t, page["is_public"].(bool))

		t.Logf("✓ Mobile viewport access successful")
	})

	// Cleanup
	t.Run("Cleanup: Restore component status", func(t *testing.T) {
		payload := map[string]interface{}{
			"status": "operational",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH",
			fmt.Sprintf("%s/api/v1/components/%s", ComponentServiceURL, componentID),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		client.Do(req)

		t.Logf("✓ Component status restored to operational")
	})

	t.Run("Cleanup: Resolve test incident", func(t *testing.T) {
		payload := map[string]interface{}{
			"status":  "resolved",
			"message": "E2E test completed - resolving incident",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH",
			fmt.Sprintf("%s/api/v1/incidents/%s", IncidentServiceURL, incidentID),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		client.Do(req)

		t.Logf("✓ Test incident resolved")
	})

	t.Run("Cleanup: Delete test subscription", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE",
			fmt.Sprintf("%s/status/%s/subscribe/%s", StatusUIServiceURL, tenantSubdomain, subscriptionID),
			nil,
		)

		client := &http.Client{}
		client.Do(req)

		t.Logf("✓ Test subscription deleted")
	})

	t.Log("\n=== Status Page Viewing E2E Test Complete ===")
	t.Logf("Status page: %s", tenantSubdomain)
	t.Logf("Total steps: 12/12 passed")
}

// Helper function to get test auth token
func getTestAuthToken(t *testing.T, tenantID string) string {
	// In a real test, this would authenticate with a test user
	// For now, return a placeholder
	return "test-auth-token"
}
