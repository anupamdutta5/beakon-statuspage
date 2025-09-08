package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuth2Config represents OAuth2 configuration
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	Provider     string   `json:"provider"` // google, github, microsoft, etc.
}

// OAuth2Provider represents an OAuth2 provider
type OAuth2Provider struct {
	config OAuth2Config
	client *http.Client
}

// OAuth2Token represents an OAuth2 token response
type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// OAuth2UserInfo represents user information from OAuth2 provider
type OAuth2UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider string `json:"provider"`
}

// OAuth2State represents OAuth2 state for CSRF protection
type OAuth2State struct {
	State       string    `json:"state"`
	RedirectURL string    `json:"redirect_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// NewOAuth2Provider creates a new OAuth2 provider
func NewOAuth2Provider(config OAuth2Config) *OAuth2Provider {
	return &OAuth2Provider{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateAuthURL generates an OAuth2 authorization URL
func (op *OAuth2Provider) GenerateAuthURL(state string) (string, error) {
	params := url.Values{}
	params.Set("client_id", op.config.ClientID)
	params.Set("redirect_uri", op.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(op.config.Scopes, " "))
	params.Set("state", state)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")

	authURL := fmt.Sprintf("%s?%s", op.config.AuthURL, params.Encode())
	return authURL, nil
}

// ExchangeCodeForToken exchanges authorization code for access token
func (op *OAuth2Provider) ExchangeCodeForToken(code string) (*OAuth2Token, error) {
	data := url.Values{}
	data.Set("client_id", op.config.ClientID)
	data.Set("client_secret", op.config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", op.config.RedirectURL)

	req, err := http.NewRequest("POST", op.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := op.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var token OAuth2Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}

// RefreshToken refreshes an access token using refresh token
func (op *OAuth2Provider) RefreshToken(refreshToken string) (*OAuth2Token, error) {
	data := url.Values{}
	data.Set("client_id", op.config.ClientID)
	data.Set("client_secret", op.config.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", op.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := op.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var token OAuth2Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}

// GetUserInfo retrieves user information using access token
func (op *OAuth2Provider) GetUserInfo(accessToken string) (*OAuth2UserInfo, error) {
	req, err := http.NewRequest("GET", op.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := op.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo OAuth2UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	userInfo.Provider = op.config.Provider
	return &userInfo, nil
}

// ValidateToken validates an access token
func (op *OAuth2Provider) ValidateToken(accessToken string) (bool, error) {
	// Try to get user info to validate token
	_, err := op.GetUserInfo(accessToken)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GenerateState generates a random state for CSRF protection
func (op *OAuth2Provider) GenerateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// OAuth2Manager manages OAuth2 providers
type OAuth2Manager struct {
	providers map[string]*OAuth2Provider
	states    map[string]*OAuth2State
}

// NewOAuth2Manager creates a new OAuth2 manager
func NewOAuth2Manager() *OAuth2Manager {
	return &OAuth2Manager{
		providers: make(map[string]*OAuth2Provider),
		states:    make(map[string]*OAuth2State),
	}
}

// RegisterProvider registers an OAuth2 provider
func (om *OAuth2Manager) RegisterProvider(name string, config OAuth2Config) {
	om.providers[name] = NewOAuth2Provider(config)
}

// GetProvider returns an OAuth2 provider by name
func (om *OAuth2Manager) GetProvider(name string) (*OAuth2Provider, error) {
	provider, exists := om.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GenerateAuthURL generates an OAuth2 authorization URL for a provider
func (om *OAuth2Manager) GenerateAuthURL(providerName, redirectURL string) (string, string, error) {
	provider, err := om.GetProvider(providerName)
	if err != nil {
		return "", "", err
	}

	// Generate state
	state, err := provider.GenerateState()
	if err != nil {
		return "", "", err
	}

	// Store state
	om.states[state] = &OAuth2State{
		State:       state,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	// Generate auth URL
	authURL, err := provider.GenerateAuthURL(state)
	if err != nil {
		return "", "", err
	}

	return authURL, state, nil
}

// HandleCallback handles OAuth2 callback
func (om *OAuth2Manager) HandleCallback(providerName, code, state string) (*OAuth2Token, *OAuth2UserInfo, error) {
	// Validate state
	stateData, exists := om.states[state]
	if !exists {
		return nil, nil, fmt.Errorf("invalid state")
	}

	if time.Now().After(stateData.ExpiresAt) {
		delete(om.states, state)
		return nil, nil, fmt.Errorf("state expired")
	}

	// Clean up state
	delete(om.states, state)

	// Get provider
	provider, err := om.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	// Exchange code for token
	token, err := provider.ExchangeCodeForToken(code)
	if err != nil {
		return nil, nil, err
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	return token, userInfo, nil
}

// RefreshAccessToken refreshes an access token
func (om *OAuth2Manager) RefreshAccessToken(providerName, refreshToken string) (*OAuth2Token, error) {
	provider, err := om.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.RefreshToken(refreshToken)
}

// ValidateAccessToken validates an access token
func (om *OAuth2Manager) ValidateAccessToken(providerName, accessToken string) (bool, error) {
	provider, err := om.GetProvider(providerName)
	if err != nil {
		return false, err
	}

	return provider.ValidateToken(accessToken)
}

// GetUserInfo retrieves user information using access token
func (om *OAuth2Manager) GetUserInfo(providerName, accessToken string) (*OAuth2UserInfo, error) {
	provider, err := om.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetUserInfo(accessToken)
}

// CleanupExpiredStates removes expired states
func (om *OAuth2Manager) CleanupExpiredStates() {
	now := time.Now()
	for state, stateData := range om.states {
		if now.After(stateData.ExpiresAt) {
			delete(om.states, state)
		}
	}
}

// OAuth2Middleware provides HTTP middleware for OAuth2
type OAuth2Middleware struct {
	manager *OAuth2Manager
}

// NewOAuth2Middleware creates a new OAuth2 middleware
func NewOAuth2Middleware(manager *OAuth2Manager) *OAuth2Middleware {
	return &OAuth2Middleware{
		manager: manager,
	}
}

// AuthHandler handles OAuth2 authentication
func (om *OAuth2Middleware) AuthHandler(providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := r.URL.Query().Get("redirect_url")
		if redirectURL == "" {
			redirectURL = "/"
		}

		authURL, _, err := om.manager.GenerateAuthURL(providerName, redirectURL)
		if err != nil {
			http.Error(w, "Failed to generate auth URL", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

// CallbackHandler handles OAuth2 callback
func (om *OAuth2Middleware) CallbackHandler(providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" || state == "" {
			http.Error(w, "Missing code or state", http.StatusBadRequest)
			return
		}

		token, userInfo, err := om.manager.HandleCallback(providerName, code, state)
		if err != nil {
			http.Error(w, "Failed to handle callback", http.StatusInternalServerError)
			return
		}

		// TODO: Create or update user in database
		// TODO: Generate JWT token
		// TODO: Set session cookie

		// For now, just return the user info
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"token":     token,
			"user_info": userInfo,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Global OAuth2 manager instance
var globalOAuth2Manager *OAuth2Manager

// InitGlobalOAuth2 initializes the global OAuth2 manager
func InitGlobalOAuth2() {
	globalOAuth2Manager = NewOAuth2Manager()
}

// GetGlobalOAuth2 returns the global OAuth2 manager
func GetGlobalOAuth2() *OAuth2Manager {
	return globalOAuth2Manager
}

// RegisterProvider is a convenience function for registering providers
func RegisterProvider(name string, config OAuth2Config) {
	if globalOAuth2Manager != nil {
		globalOAuth2Manager.RegisterProvider(name, config)
	}
}

// GenerateAuthURL is a convenience function for generating auth URLs
func GenerateAuthURL(providerName, redirectURL string) (string, string, error) {
	if globalOAuth2Manager != nil {
		return globalOAuth2Manager.GenerateAuthURL(providerName, redirectURL)
	}
	return "", "", fmt.Errorf("OAuth2 manager not initialized")
}

// HandleCallback is a convenience function for handling callbacks
func HandleCallback(providerName, code, state string) (*OAuth2Token, *OAuth2UserInfo, error) {
	if globalOAuth2Manager != nil {
		return globalOAuth2Manager.HandleCallback(providerName, code, state)
	}
	return nil, nil, fmt.Errorf("OAuth2 manager not initialized")
}
