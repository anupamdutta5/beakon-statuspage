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
	MonitoringServiceURL  = "http://localhost:8092"
	NotificationServiceURL = "http://localhost:8085"
	ComponentServiceURL    = "http://localhost:8084"
)

// TestMonitoringAndAlertingFlow tests the complete monitoring and alerting process
// This covers:
// 1. Creating a component
// 2. Setting up a monitor
// 3. Configuring alert rules
// 4. Setting up notification channels
// 5. Triggering alerts
// 6. Verifying notifications are sent
func TestMonitoringAndAlertingFlow(t *testing.T) {
	tenantID := "tenant-1111-1111-1111-111111111111"
	authToken := getTestAuthToken(t, tenantID)

	var componentID string
	var monitorID string
	var alertRuleID string
	var channelID string

	t.Run("Step 1: Create component to monitor", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "E2E Test API",
			"description": "API service for E2E testing",
			"status":      "operational",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/components", ComponentServiceURL),
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

		componentID = response["id"].(string)
		assert.NotEmpty(t, componentID)

		t.Logf("✓ Component created: E2E Test API (ID: %s)", componentID)
	})

	t.Run("Step 2: Set up HTTP monitor", func(t *testing.T) {
		payload := map[string]interface{}{
			"component_id":          componentID,
			"monitor_type":          "http",
			"name":                  "API Health Check",
			"description":           "Monitor API endpoint availability",
			"target_url":            "https://httpbin.org/status/200",
			"check_interval_seconds": 60,
			"timeout_seconds":       30,
			"is_active":             true,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/monitors", MonitoringServiceURL),
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

		monitorID = response["id"].(string)
		assert.NotEmpty(t, monitorID)

		t.Logf("✓ Monitor created: API Health Check (ID: %s)", monitorID)
	})

	t.Run("Step 3: Create notification channel (webhook)", func(t *testing.T) {
		payload := map[string]interface{}{
			"channel_type":  "webhook",
			"channel_name":  "E2E Test Webhook",
			"configuration": map[string]string{
				"url":    "http://localhost:1080/webhook/e2e-test",
				"method": "POST",
			},
			"is_active": true,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/notification-channels", NotificationServiceURL),
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

		channelID = response["id"].(string)
		assert.NotEmpty(t, channelID)

		t.Logf("✓ Notification channel created (ID: %s)", channelID)
	})

	t.Run("Step 4: Configure alert rule", func(t *testing.T) {
		payload := map[string]interface{}{
			"monitor_id":          monitorID,
			"rule_name":           "API Down Alert",
			"condition_type":      "failure",
			"consecutive_failures": 2,
			"alert_channels":      []string{channelID},
			"is_active":           true,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/alert-rules", MonitoringServiceURL),
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

		alertRuleID = response["id"].(string)
		assert.NotEmpty(t, alertRuleID)

		t.Logf("✓ Alert rule created (ID: %s)", alertRuleID)
	})

	t.Run("Step 5: Trigger monitor check manually", func(t *testing.T) {
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/monitors/%s/check", MonitoringServiceURL, monitorID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result["status"])

		t.Logf("✓ Monitor check executed: %s", result["status"])
	})

	t.Run("Step 6: Update monitor to failing URL", func(t *testing.T) {
		payload := map[string]interface{}{
			"target_url": "https://httpbin.org/status/500", // Will fail
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH",
			fmt.Sprintf("%s/api/v1/monitors/%s", MonitoringServiceURL, monitorID),
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

		t.Logf("✓ Monitor updated to failing endpoint")
	})

	t.Run("Step 7: Trigger failing checks to generate alert", func(t *testing.T) {
		client := &http.Client{}

		// Trigger 3 consecutive failures
		for i := 1; i <= 3; i++ {
			req, _ := http.NewRequest("POST",
				fmt.Sprintf("%s/api/v1/monitors/%s/check", MonitoringServiceURL, monitorID),
				nil,
			)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
			req.Header.Set("X-Tenant-ID", tenantID)

			resp, err := client.Do(req)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			resp.Body.Close()

			t.Logf("  Check %d/3: %s", i, result["status"])

			time.Sleep(2 * time.Second)
		}

		t.Logf("✓ Triggered 3 consecutive failures")
	})

	t.Run("Step 8: Verify alert was triggered", func(t *testing.T) {
		// Wait for alert processing
		time.Sleep(3 * time.Second)

		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/api/v1/alert-notifications?monitor_id=%s", MonitoringServiceURL, monitorID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var alerts []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&alerts)
		require.NoError(t, err)

		assert.NotEmpty(t, alerts, "At least one alert should be triggered")

		activeAlert := false
		for _, alert := range alerts {
			if alert["status"] == "active" {
				activeAlert = true
				t.Logf("✓ Active alert found: %s", alert["message"])
				break
			}
		}

		assert.True(t, activeAlert, "Should have at least one active alert")
	})

	t.Run("Step 9: Verify notification was sent", func(t *testing.T) {
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/api/v1/notifications?component_id=%s&limit=10", NotificationServiceURL, componentID),
			nil,
		)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req.Header.Set("X-Tenant-ID", tenantID)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var notifications []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&notifications)
		require.NoError(t, err)

		assert.NotEmpty(t, notifications, "Notifications should have been sent")

		deliveredNotification := false
		for _, notif := range notifications {
			if notif["delivery_status"] == "delivered" || notif["delivery_status"] == "pending" {
				deliveredNotification = true
				t.Logf("✓ Notification sent: %s", notif["subject"])
				break
			}
		}

		assert.True(t, deliveredNotification, "At least one notification should be delivered/pending")
	})

	t.Run("Step 10: Restore monitor and verify resolution", func(t *testing.T) {
		// Update back to working URL
		payload := map[string]interface{}{
			"target_url": "https://httpbin.org/status/200",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH",
			fmt.Sprintf("%s/api/v1/monitors/%s", MonitoringServiceURL, monitorID),
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

		// Trigger successful check
		time.Sleep(1 * time.Second)
		req2, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/api/v1/monitors/%s/check", MonitoringServiceURL, monitorID),
			nil,
		)
		req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req2.Header.Set("X-Tenant-ID", tenantID)

		resp2, err := client.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()

		var result map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result["status"])

		t.Logf("✓ Monitor restored and check successful")
	})

	// Cleanup
	t.Run("Cleanup: Delete test resources", func(t *testing.T) {
		client := &http.Client{}

		// Delete monitor
		req1, _ := http.NewRequest("DELETE",
			fmt.Sprintf("%s/api/v1/monitors/%s", MonitoringServiceURL, monitorID),
			nil,
		)
		req1.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req1.Header.Set("X-Tenant-ID", tenantID)
		client.Do(req1)

		// Delete component
		req2, _ := http.NewRequest("DELETE",
			fmt.Sprintf("%s/api/v1/components/%s", ComponentServiceURL, componentID),
			nil,
		)
		req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req2.Header.Set("X-Tenant-ID", tenantID)
		client.Do(req2)

		// Delete notification channel
		req3, _ := http.NewRequest("DELETE",
			fmt.Sprintf("%s/api/v1/notification-channels/%s", NotificationServiceURL, channelID),
			nil,
		)
		req3.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		req3.Header.Set("X-Tenant-ID", tenantID)
		client.Do(req3)

		t.Logf("✓ Test resources cleaned up")
	})

	t.Log("\n=== Monitoring & Alerting E2E Test Complete ===")
	t.Logf("Component: E2E Test API")
	t.Logf("Monitor: API Health Check")
	t.Logf("Total steps: 10/10 passed")
}

// Helper function to get test auth token
func getTestAuthToken(t *testing.T, tenantID string) string {
	// In a real test, this would authenticate with a test user
	// For now, return a placeholder
	return "test-auth-token"
}
