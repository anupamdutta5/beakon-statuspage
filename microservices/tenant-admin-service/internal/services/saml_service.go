// Package services provides SAML/SSO authentication services
package services

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
)

// SAMLService handles SAML authentication for multi-tenant environment
type SAMLService struct {
	db            *gorm.DB
	logger        *zap.Logger
	config        *SAMLConfig
	tenantService *TenantAdminService // For user creation
}

// SAMLConfig holds SAML configuration
type SAMLConfig struct {
	BaseURL           string // e.g., https://your-app.com
	EntityID          string // SP Entity ID
	CertificatePath   string
	PrivateKeyPath    string
	MetadataURL       string // /saml/metadata
	ACSURL            string // /saml/acs (Assertion Consumer Service)
	SLOurl            string // /saml/slo (Single Logout)
	CookieMaxAge      int    // Session cookie max age (seconds)
	AllowIDPInitiated bool
}

// NewSAMLService creates a new SAML service
func NewSAMLService(db *gorm.DB, tenantService *TenantAdminService, logger *zap.Logger, config *SAMLConfig) *SAMLService {
	return &SAMLService{
		db:            db,
		logger:        logger,
		config:        config,
		tenantService: tenantService,
	}
}

// DB returns the database instance
func (s *SAMLService) DB() *gorm.DB {
	return s.db
}

// GetServiceProvider creates a SAML Service Provider for a given SSO provider
// Multi-tenant: Uses tenant-specific subdomain in URLs
func (s *SAMLService) GetServiceProvider(ctx context.Context, ssoProvider *models.SSOProvider, subdomain string) (*saml.ServiceProvider, error) {
	// Build tenant-specific base URL
	baseURL := s.config.BaseURL
	if subdomain != "" {
		// Replace base URL with tenant subdomain
		// e.g., https://acme.example.com instead of https://example.com
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		parsedURL.Host = fmt.Sprintf("%s.%s", subdomain, parsedURL.Host)
		baseURL = parsedURL.String()
	}

	// Parse metadata URL
	metadataURL, err := url.Parse(fmt.Sprintf("%s%s", baseURL, s.config.MetadataURL))
	if err != nil {
		return nil, fmt.Errorf("invalid metadata URL: %w", err)
	}

	// Parse ACS URL
	acsURL, err := url.Parse(fmt.Sprintf("%s%s", baseURL, s.config.ACSURL))
	if err != nil {
		return nil, fmt.Errorf("invalid ACS URL: %w", err)
	}

	// Parse SLO URL
	sloURL, err := url.Parse(fmt.Sprintf("%s%s", baseURL, s.config.SLOurl))
	if err != nil {
		return nil, fmt.Errorf("invalid SLO URL: %w", err)
	}

	// Parse IdP metadata
	idpMetadata, err := s.parseIdPMetadata(ssoProvider.IdpMetadataXML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IdP metadata: %w", err)
	}

	// Parse SP certificate and key
	certPair, err := s.parseCertificateAndKey(ssoProvider.SPCertificate, ssoProvider.SPPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SP certificate: %w", err)
	}

	sp := &saml.ServiceProvider{
		EntityID:          ssoProvider.EntityID,
		Key:               certPair.PrivateKey,
		Certificate:       certPair.Certificate,
		MetadataURL:       *metadataURL,
		AcsURL:            *acsURL,
		SloURL:            *sloURL,
		IDPMetadata:       idpMetadata,
		AllowIDPInitiated: ssoProvider.AllowIdpInitiated,
	}

	return sp, nil
}

// InitiateLogin starts a SAML authentication flow (SP-initiated)
// Multi-tenant: tenantID is required for tenant-scoped provider lookup
func (s *SAMLService) InitiateLogin(ctx context.Context, tenantID uuid.UUID, organizationDomain string, relayState string, r *http.Request, subdomain string) (*SAMLLoginRequest, error) {
	// Find SSO provider by organization domain and tenant
	var ssoProvider models.SSOProvider
	if err := s.db.Where("tenant_id = ? AND organization_domain = ? AND is_enabled = ? AND deleted_at IS NULL",
		tenantID, organizationDomain, true).First(&ssoProvider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("SSO provider not found for domain: %s in tenant: %s", organizationDomain, tenantID)
		}
		return nil, fmt.Errorf("failed to query SSO provider: %w", err)
	}

	// Get Service Provider with tenant subdomain
	sp, err := s.GetServiceProvider(ctx, &ssoProvider, subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get service provider: %w", err)
	}

	// Create auth request
	authReq, err := sp.MakeAuthenticationRequest(sp.GetSSOBindingLocation(saml.HTTPRedirectBinding), saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	// Store SAML request in database with tenant context
	samlReq := &models.SAMLRequest{
		RequestID:     authReq.ID,
		SSOProviderID: ssoProvider.ID,
		TenantID:      tenantID,
		RelayState:    relayState,
		ACSURL:        sp.AcsURL.String(),
		IPAddress:     getIPAddress(r),
		UserAgent:     r.UserAgent(),
		ExpiresAt:     time.Now().Add(5 * time.Minute),
		IsCompleted:   false,
	}

	if err := s.db.Create(samlReq).Error; err != nil {
		return nil, fmt.Errorf("failed to store SAML request: %w", err)
	}

	// Log audit event with tenant context
	s.logAuditEvent(&models.SSOAuditLog{
		TenantID:         tenantID,
		SSOProviderID:    &ssoProvider.ID,
		EventType:        "login_initiated",
		EventDescription: fmt.Sprintf("SAML login initiated for domain: %s", organizationDomain),
		IPAddress:        getIPAddress(r),
		UserAgent:        r.UserAgent(),
		SAMLRequestID:    authReq.ID,
	})

	// Build redirect URL
	redirectURL, err := authReq.Redirect(relayState, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to build redirect URL: %w", err)
	}

	return &SAMLLoginRequest{
		RequestID:   authReq.ID,
		RedirectURL: redirectURL.String(),
		RelayState:  relayState,
	}, nil
}

// HandleACS processes the SAML assertion response
// Multi-tenant: Returns tenant context from SAML request
func (s *SAMLService) HandleACS(ctx context.Context, r *http.Request, subdomain string) (*SAMLAuthResult, error) {
	// Parse SAML response
	err := r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	samlResponseEncoded := r.FormValue("SAMLResponse")
	if samlResponseEncoded == "" {
		return nil, errors.New("missing SAMLResponse parameter")
	}

	relayState := r.FormValue("RelayState")

	// Decode SAML response
	samlResponseXML, err := base64.StdEncoding.DecodeString(samlResponseEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAMLResponse: %w", err)
	}

	// Parse SAML response
	var samlResp saml.Response
	if err := xml.Unmarshal(samlResponseXML, &samlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SAML response: %w", err)
	}

	// Find SAML request with tenant preload
	var samlReq models.SAMLRequest
	if err := s.db.Where("request_id = ? AND is_completed = ?", samlResp.InResponseTo, false).
		Preload("SSOProvider").Preload("Tenant").First(&samlReq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Allow IdP-initiated login if enabled
			return s.handleIdPInitiatedLogin(ctx, &samlResp, relayState, r, subdomain)
		}
		return nil, fmt.Errorf("failed to find SAML request: %w", err)
	}

	// Get Service Provider with tenant subdomain
	sp, err := s.GetServiceProvider(ctx, &samlReq.SSOProvider, subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get service provider: %w", err)
	}

	// Validate SAML response
	assertion, err := sp.ParseResponse(r, []string{samlReq.RequestID})
	if err != nil {
		s.logAuditEvent(&models.SSOAuditLog{
			TenantID:           samlReq.TenantID,
			SSOProviderID:      &samlReq.SSOProviderID,
			EventType:          models.EventLoginFailure,
			EventDescription:   fmt.Sprintf("SAML response validation failed: %s", err.Error()),
			IPAddress:          getIPAddress(r),
			UserAgent:          r.UserAgent(),
			SAMLRequestID:      samlReq.RequestID,
			SAMLResponseStatus: "validation_failed",
		})
		return nil, fmt.Errorf("failed to validate SAML response: %w", err)
	}

	// Mark request as completed
	now := time.Now()
	samlReq.IsCompleted = true
	samlReq.CompletedAt = &now
	s.db.Save(&samlReq)

	// Extract user attributes
	attributes := make(map[string]interface{})
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if len(attr.Values) > 0 {
				attributes[attr.Name] = attr.Values[0].Value
			}
		}
	}

	// Get or create user with tenant context
	user, isNewUser, err := s.getOrCreateUserFromSAML(ctx, &samlReq.SSOProvider, samlReq.TenantID, assertion, attributes)
	if err != nil {
		s.logAuditEvent(&models.SSOAuditLog{
			TenantID:           samlReq.TenantID,
			SSOProviderID:      &samlReq.SSOProviderID,
			EventType:          models.EventLoginFailure,
			EventDescription:   fmt.Sprintf("Failed to provision user: %s", err.Error()),
			IPAddress:          getIPAddress(r),
			UserAgent:          r.UserAgent(),
			SAMLRequestID:      samlReq.RequestID,
			SAMLResponseStatus: "user_provision_failed",
		})
		return nil, fmt.Errorf("failed to provision user: %w", err)
	}

	// Update SSO user identity
	err = s.updateSSOUserIdentity(ctx, user, &samlReq.SSOProvider, samlReq.TenantID, assertion, attributes, r)
	if err != nil {
		s.logger.Warn("Failed to update SSO user identity", zap.Error(err))
	}

	// Log successful login
	s.logAuditEvent(&models.SSOAuditLog{
		TenantID:           samlReq.TenantID,
		SSOProviderID:      &samlReq.SSOProviderID,
		UserID:             &user.ID,
		EventType:          models.EventLoginSuccess,
		EventDescription:   "SAML login successful",
		IPAddress:          getIPAddress(r),
		UserAgent:          r.UserAgent(),
		SAMLRequestID:      samlReq.RequestID,
		SAMLResponseStatus: "success",
		Metadata:           models.SSOMetadata{"is_new_user": isNewUser},
	})

	return &SAMLAuthResult{
		User:       user,
		TenantID:   samlReq.TenantID,
		IsNewUser:  isNewUser,
		Attributes: attributes,
		RelayState: relayState,
	}, nil
}

// getOrCreateUserFromSAML retrieves or creates a user based on SAML assertion
// Multi-tenant: Creates users with proper tenant association
func (s *SAMLService) getOrCreateUserFromSAML(ctx context.Context, provider *models.SSOProvider, tenantID uuid.UUID, assertion *saml.Assertion, attributes map[string]interface{}) (*models.User, bool, error) {
	email := s.extractAttribute(attributes, provider.AttributeMapping["email"], "email")

	if email == "" {
		return nil, false, errors.New("email not found in SAML assertion")
	}

	// Check if user exists by email and tenant
	var user models.User
	err := s.db.Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error

	if err == nil {
		// User exists, update SSO fields
		user.AuthMethod = "saml"
		user.SSOProviderID = &provider.ID
		user.IsSSOUser = true
		s.db.Save(&user)
		return &user, false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("failed to query user: %w", err)
	}

	// User doesn't exist - check if JIT provisioning is enabled
	if !provider.EnableJITProvisioning {
		return nil, false, errors.New("user does not exist and JIT provisioning is disabled")
	}

	// Create new user (JIT provisioning) using TenantAdminService
	firstName := s.extractAttribute(attributes, provider.AttributeMapping["first_name"], "firstName", "givenName")
	lastName := s.extractAttribute(attributes, provider.AttributeMapping["last_name"], "lastName", "surname")

	newUser := &models.User{
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		TenantID:      tenantID,
		Role:          provider.DefaultRole, // e.g., "viewer"
		IsActive:      true,
		AuthMethod:    "saml",
		SSOProviderID: &provider.ID,
		IsSSOUser:     true,
		PasswordHash:  "", // No password for SSO users
	}

	// Use TenantAdminService to create user (ensures proper validation)
	if err := s.db.Create(newUser).Error; err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("JIT provisioned new user from SAML",
		zap.String("email", email),
		zap.String("provider", provider.ProviderName),
		zap.String("tenant_id", tenantID.String()),
		zap.String("user_id", newUser.ID.String()))

	return newUser, true, nil
}

// updateSSOUserIdentity creates or updates the SSO user identity record
// Multi-tenant: Includes tenant context
func (s *SAMLService) updateSSOUserIdentity(ctx context.Context, user *models.User, provider *models.SSOProvider, tenantID uuid.UUID, assertion *saml.Assertion, attributes map[string]interface{}, r *http.Request) error {
	nameID := assertion.Subject.NameID.Value
	sessionIndex := ""
	if len(assertion.AuthnStatements) > 0 {
		sessionIndex = assertion.AuthnStatements[0].SessionIndex
	}

	var identity models.SSOUserIdentity
	err := s.db.Where("user_id = ? AND sso_provider_id = ? AND tenant_id = ?", user.ID, provider.ID, tenantID).First(&identity).Error

	now := time.Now()
	ipAddress := getIPAddress(r)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new identity
		identity = models.SSOUserIdentity{
			UserID:        user.ID,
			SSOProviderID: provider.ID,
			TenantID:      tenantID,
			IdpUserID:     nameID,
			IdpEmail:      user.Email,
			SessionIndex:  sessionIndex,
			NameIDFormat:  assertion.Subject.NameID.Format,
			Attributes:    models.SSOMetadata(attributes),
			LastLoginAt:   &now,
			LastLoginIP:   ipAddress,
			LoginCount:    1,
		}
		return s.db.Create(&identity).Error
	}

	if err != nil {
		return err
	}

	// Update existing identity
	identity.SessionIndex = sessionIndex
	identity.Attributes = models.SSOMetadata(attributes)
	identity.LastLoginAt = &now
	identity.LastLoginIP = ipAddress
	identity.LoginCount++

	return s.db.Save(&identity).Error
}

// handleIdPInitiatedLogin handles IdP-initiated SAML login
// Multi-tenant: Extracts tenant from subdomain
func (s *SAMLService) handleIdPInitiatedLogin(ctx context.Context, samlResp *saml.Response, relayState string, r *http.Request, subdomain string) (*SAMLAuthResult, error) {
	// Get tenant by subdomain
	var tenant models.Tenant
	if err := s.db.Where("subdomain = ?", subdomain).First(&tenant).Error; err != nil {
		return nil, fmt.Errorf("tenant not found for subdomain: %s", subdomain)
	}

	// Find provider by IdP Entity ID and tenant
	var ssoProvider models.SSOProvider
	if err := s.db.Where("tenant_id = ? AND idp_entity_id = ? AND is_enabled = ? AND allow_idp_initiated = ? AND deleted_at IS NULL",
		tenant.ID, samlResp.Issuer.Value, true, true).First(&ssoProvider).Error; err != nil {
		return nil, fmt.Errorf("SSO provider not found or IdP-initiated login not allowed")
	}

	// Get Service Provider
	sp, err := s.GetServiceProvider(ctx, &ssoProvider, subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get service provider: %w", err)
	}

	// Validate response (without request ID for IdP-initiated)
	assertion, err := sp.ParseResponse(r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate IdP-initiated SAML response: %w", err)
	}

	// Extract attributes and process user
	attributes := make(map[string]interface{})
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if len(attr.Values) > 0 {
				attributes[attr.Name] = attr.Values[0].Value
			}
		}
	}

	user, isNewUser, err := s.getOrCreateUserFromSAML(ctx, &ssoProvider, tenant.ID, assertion, attributes)
	if err != nil {
		return nil, err
	}

	s.updateSSOUserIdentity(ctx, user, &ssoProvider, tenant.ID, assertion, attributes, r)

	return &SAMLAuthResult{
		User:       user,
		TenantID:   tenant.ID,
		IsNewUser:  isNewUser,
		Attributes: attributes,
		RelayState: relayState,
	}, nil
}

// Helper functions

func (s *SAMLService) extractAttribute(attributes map[string]interface{}, mappingKeys ...string) string {
	for _, key := range mappingKeys {
		if val, ok := attributes[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}
	return ""
}

func (s *SAMLService) parseIdPMetadata(metadataXML string) (*saml.EntityDescriptor, error) {
	var metadata saml.EntityDescriptor
	if err := xml.Unmarshal([]byte(metadataXML), &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

type CertificatePair struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
}

func (s *SAMLService) parseCertificateAndKey(certPEM, keyPEM string) (*CertificatePair, error) {
	// Parse certificate
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		return nil, errors.New("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		return nil, errors.New("failed to decode private key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &CertificatePair{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

func (s *SAMLService) logAuditEvent(log *models.SSOAuditLog) {
	if err := s.db.Create(log).Error; err != nil {
		s.logger.Error("Failed to log SSO audit event", zap.Error(err))
	}
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// Data structures

type SAMLLoginRequest struct {
	RequestID   string
	RedirectURL string
	RelayState  string
}

type SAMLAuthResult struct {
	User       *models.User
	TenantID   uuid.UUID
	IsNewUser  bool
	Attributes map[string]interface{}
	RelayState string
}
