package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/notification-service/internal/services"
)

// SlackHandler handles Slack OAuth and integration endpoints
type SlackHandler struct {
	db           *gorm.DB
	logger       *zap.Logger
	slackService *services.SlackIntegrationService
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewSlackHandler creates a new Slack handler
func NewSlackHandler(db *gorm.DB, logger *zap.Logger, slackService *services.SlackIntegrationService) *SlackHandler {
	return &SlackHandler{
		db:           db,
		logger:       logger,
		slackService: slackService,
		clientID:     os.Getenv("SLACK_CLIENT_ID"),
		clientSecret: os.Getenv("SLACK_CLIENT_SECRET"),
		redirectURI:  os.Getenv("SLACK_REDIRECT_URI"),
	}
}

// InstallSlack initiates the Slack OAuth flow
// GET /api/v1/integrations/slack/install
func (h *SlackHandler) InstallSlack(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	// Validate tenant_id is a valid UUID
	if _, err := uuid.Parse(tenantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Build Slack OAuth URL
	state := fmt.Sprintf("%s:%s", tenantID, uuid.New().String()) // tenant_id:random_uuid
	scope := "chat:write,incoming-webhook"

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		url.QueryEscape(h.clientID),
		url.QueryEscape(scope),
		url.QueryEscape(h.redirectURI),
		url.QueryEscape(state),
	)

	h.logger.Info("Initiating Slack OAuth flow",
		zap.String("tenant_id", tenantID),
		zap.String("state", state),
	)

	// Redirect to Slack OAuth
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// SlackOAuthResponse represents Slack OAuth token response
type SlackOAuthResponse struct {
	OK          bool   `json:"ok"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	AppID       string `json:"app_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	IncomingWebhook struct {
		Channel          string `json:"channel"`
		ChannelID        string `json:"channel_id"`
		ConfigurationURL string `json:"configuration_url"`
		URL              string `json:"url"`
	} `json:"incoming_webhook"`
	Error string `json:"error,omitempty"`
}

// OAuthCallback handles the Slack OAuth callback
// GET /api/v1/integrations/slack/callback
func (h *SlackHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorMsg := c.Query("error")

	// Check for OAuth errors
	if errorMsg != "" {
		h.logger.Error("Slack OAuth error", zap.String("error", errorMsg))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Slack OAuth error: %s", errorMsg)})
		return
	}

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state parameter"})
		return
	}

	// Parse state to extract tenant_id
	tenantID, err := h.parseTenantFromState(state)
	if err != nil {
		h.logger.Error("Invalid state parameter", zap.String("state", state), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	h.logger.Info("Slack OAuth callback received",
		zap.String("tenant_id", tenantID.String()),
		zap.String("code", code[:10]+"..."),
	)

	// Exchange code for access token
	token, err := h.exchangeCodeForToken(code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete Slack OAuth"})
		return
	}

	if !token.OK {
		h.logger.Error("Slack OAuth token exchange failed", zap.String("error", token.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Slack error: %s", token.Error)})
		return
	}

	// Save integration to database
	integration := &services.SlackIntegration{
		TenantID:      tenantID,
		WorkspaceName: token.Team.Name,
		WebhookURL:    token.IncomingWebhook.URL,
		DefaultChannel: token.IncomingWebhook.Channel,
		IsActive:      true,
		NotifyOnDown:  true,
		NotifyOnUp:    true,
	}

	if err := h.slackService.CreateIntegration(integration); err != nil {
		h.logger.Error("Failed to save Slack integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save integration"})
		return
	}

	h.logger.Info("Slack integration created successfully",
		zap.String("tenant_id", tenantID.String()),
		zap.String("workspace", token.Team.Name),
		zap.String("channel", token.IncomingWebhook.Channel),
	)

	// Return success HTML page
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Slack Integration Complete</title>
			<style>
				body {
					font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
					background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
				}
				.container {
					background: white;
					padding: 40px;
					border-radius: 12px;
					box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
					text-align: center;
					max-width: 500px;
				}
				h1 { color: #2d3748; margin-bottom: 10px; }
				p { color: #718096; margin-bottom: 20px; }
				.success { color: #48bb78; font-size: 48px; margin-bottom: 20px; }
				.button {
					background: #667eea;
					color: white;
					padding: 12px 24px;
					border-radius: 6px;
					text-decoration: none;
					display: inline-block;
					margin-top: 20px;
				}
				.details {
					background: #f7fafc;
					padding: 15px;
					border-radius: 6px;
					margin-top: 20px;
					text-align: left;
					font-size: 14px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="success">✅</div>
				<h1>Slack Integration Complete!</h1>
				<p>Your Beakon status page is now connected to Slack.</p>
				<div class="details">
					<strong>Workspace:</strong> %s<br>
					<strong>Channel:</strong> %s<br>
					<strong>Status:</strong> Active
				</div>
				<a href="/admin/integrations" class="button">Return to Dashboard</a>
				<p style="margin-top: 20px; font-size: 12px; color: #a0aec0;">
					You can now close this window.
				</p>
			</div>
		</body>
		</html>
	`, token.Team.Name, token.IncomingWebhook.Channel)
}

// exchangeCodeForToken exchanges OAuth code for access token
func (h *SlackHandler) exchangeCodeForToken(code string) (*SlackOAuthResponse, error) {
	tokenURL := "https://slack.com/api/oauth.v2.access"

	data := url.Values{}
	data.Set("client_id", h.clientID)
	data.Set("client_secret", h.clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", h.redirectURI)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to post token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp SlackOAuthResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// parseTenantFromState extracts tenant_id from OAuth state parameter
func (h *SlackHandler) parseTenantFromState(state string) (uuid.UUID, error) {
	// State format: "tenant_id:random_uuid"
	// Extract tenant_id part before the colon
	tenantIDStr := state
	if idx := len(state); idx > 36 {
		tenantIDStr = state[:36] // UUID is 36 characters
	}

	// Try to find colon separator
	for i := 0; i < len(state); i++ {
		if state[i] == ':' {
			tenantIDStr = state[:i]
			break
		}
	}

	return uuid.Parse(tenantIDStr)
}

// GetIntegrations returns all Slack integrations for a tenant
// GET /api/v1/integrations/slack
func (h *SlackHandler) GetIntegrations(c *gin.Context) {
	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	integrations, err := h.slackService.GetIntegrationsByTenant(tenantID)
	if err != nil {
		h.logger.Error("Failed to get Slack integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve integrations"})
		return
	}

	// Return first integration if exists
	var integration *services.SlackIntegration
	if len(integrations) > 0 {
		integration = &integrations[0]
	}

	if integration == nil {
		c.JSON(http.StatusOK, gin.H{"integration": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integration": integration})
}

// DeleteIntegration deletes a Slack integration
// DELETE /api/v1/integrations/slack/:id
func (h *SlackHandler) DeleteIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	if err := h.slackService.DeleteIntegration(id, tenantID); err != nil {
		h.logger.Error("Failed to delete Slack integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	h.logger.Info("Slack integration deleted",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted successfully"})
}

// TestIntegration sends a test notification to Slack
// POST /api/v1/integrations/slack/:id/test
func (h *SlackHandler) TestIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration ID"})
		return
	}

	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
		return
	}

	// Get integration
	integration, err := h.slackService.GetIntegration(id, tenantID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	// Send test notification using SendMonitorAlert
	testData := map[string]interface{}{
		"error":               "This is a test notification from Beakon Status Page",
		"consecutive_failures": 1,
	}

	if err := h.slackService.SendMonitorAlert(c.Request.Context(), 0, tenantID, "down", "Test Monitor", "", testData); err != nil {
		h.logger.Error("Failed to send test notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test notification"})
		return
	}

	h.logger.Info("Test Slack notification sent",
		zap.Uint("id", id),
		zap.String("tenant_id", tenantID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "test notification sent successfully"})
}
