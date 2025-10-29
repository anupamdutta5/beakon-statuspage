// Package models provides SSO/SAML data models for the Tenant Admin Service.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttributeMapping represents the mapping from IdP attributes to user fields
type AttributeMapping map[string]string

// Scan implements sql.Scanner for AttributeMapping
func (am *AttributeMapping) Scan(value interface{}) error {
	if value == nil {
		*am = make(AttributeMapping)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan AttributeMapping")
	}

	return json.Unmarshal(bytes, am)
}

// Value implements driver.Valuer for AttributeMapping
func (am AttributeMapping) Value() (driver.Value, error) {
	if am == nil {
		return json.Marshal(make(map[string]string))
	}
	return json.Marshal(am)
}

// SSOMetadata represents generic metadata stored as JSONB
type SSOMetadata map[string]interface{}

// Scan implements sql.Scanner for SSOMetadata
func (m *SSOMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(SSOMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SSOMetadata")
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for SSOMetadata
func (m SSOMetadata) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(make(map[string]interface{}))
	}
	return json.Marshal(m)
}

// SSOProvider represents an SSO identity provider configuration
// Multi-tenant: Each tenant can have multiple SSO providers
type SSOProvider struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Multi-tenant association
	TenantID uuid.UUID `gorm:"type:uuid;not null;index:idx_sso_providers_tenant" json:"tenant_id"`

	// Organization association
	OrganizationDomain string `gorm:"not null;index:idx_sso_providers_domain" json:"organization_domain"` // @company.com
	OrganizationName   string `gorm:"not null" json:"organization_name"`

	// Provider configuration
	ProviderType string `gorm:"not null" json:"provider_type"` // 'saml', 'oauth', 'oidc'
	ProviderName string `gorm:"not null" json:"provider_name"` // 'Okta', 'Azure AD', 'Google Workspace'

	// SAML-specific fields
	EntityID         string `json:"entity_id"`          // SP Entity ID
	IdpEntityID      string `json:"idp_entity_id"`      // IdP Entity ID
	SSOURL           string `json:"sso_url"`            // Single Sign-On URL
	SLOURL           string `json:"slo_url"`            // Single Logout URL
	IdpMetadataURL   string `json:"idp_metadata_url"`   // IdP Metadata URL
	IdpMetadataXML   string `gorm:"type:text" json:"-"` // Cached metadata (don't expose in JSON)
	IdpCertificate   string `gorm:"type:text" json:"-"` // X.509 cert (don't expose)
	SPCertificate    string `gorm:"type:text" json:"-"` // SP cert
	SPPrivateKey     string `gorm:"type:text" json:"-"` // SP private key (encrypted)

	// OAuth/OIDC fields
	ClientID              string `json:"client_id,omitempty"`
	ClientSecret          string `gorm:"type:text" json:"-"` // Encrypted
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string `json:"token_endpoint,omitempty"`
	UserinfoEndpoint      string `json:"userinfo_endpoint,omitempty"`
	JWKSURI               string `json:"jwks_uri,omitempty"`

	// Configuration
	IsEnabled         bool `gorm:"default:true" json:"is_enabled"`
	IsDefault         bool `gorm:"default:false" json:"is_default"`
	EnforceSSO        bool `gorm:"default:false" json:"enforce_sso"`
	AllowIdpInitiated bool `gorm:"default:true" json:"allow_idp_initiated"`

	// Attribute mapping
	AttributeMapping AttributeMapping `gorm:"type:jsonb;default:'{}'" json:"attribute_mapping"`

	// JIT Provisioning
	EnableJITProvisioning bool   `gorm:"default:true" json:"enable_jit_provisioning"`
	DefaultRole           string `gorm:"default:viewer" json:"default_role"` // Changed from 'user' to 'viewer'

	// Metadata
	Metadata           SSOMetadata `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	LastMetadataUpdate *time.Time  `json:"last_metadata_update,omitempty"`

	// Relationships
	Tenant         Tenant            `gorm:"foreignKey:TenantID" json:"-"`
	UserIdentities []SSOUserIdentity `gorm:"foreignKey:SSOProviderID" json:"-"`
}

// SSOUserIdentity links a user to their SSO identity from an IdP
type SSOUserIdentity struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// User and provider association
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_sso_identities_user" json:"user_id"`
	SSOProviderID uuid.UUID `gorm:"type:uuid;not null;index:idx_sso_identities_provider" json:"sso_provider_id"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index:idx_sso_identities_tenant" json:"tenant_id"`

	// IdP-specific identity
	IdpUserID   string `gorm:"not null" json:"idp_user_id"` // NameID for SAML
	IdpUsername string `json:"idp_username,omitempty"`
	IdpEmail    string `gorm:"index" json:"idp_email,omitempty"`

	// SAML session
	SessionIndex string `json:"session_index,omitempty"`  // For SLO
	NameIDFormat string `json:"name_id_format,omitempty"` // SAML NameID format

	// Attributes from IdP
	Attributes SSOMetadata `gorm:"type:jsonb;default:'{}'" json:"attributes"`

	// Last login tracking
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP string     `json:"last_login_ip,omitempty"`
	LoginCount  int        `gorm:"default:0" json:"login_count"`

	// Relationships
	User        User        `gorm:"foreignKey:UserID" json:"-"`
	SSOProvider SSOProvider `gorm:"foreignKey:SSOProviderID" json:"-"`
	Tenant      Tenant      `gorm:"foreignKey:TenantID" json:"-"`
}

// SAMLRequest tracks pending SAML authentication requests
type SAMLRequest struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// Request tracking
	RequestID     string    `gorm:"uniqueIndex;not null" json:"request_id"`
	SSOProviderID uuid.UUID `gorm:"type:uuid;not null" json:"sso_provider_id"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index:idx_saml_requests_tenant" json:"tenant_id"`

	// Request details
	RelayState string `gorm:"type:text" json:"relay_state,omitempty"` // Return URL after auth
	ACSURL     string `gorm:"type:text;not null" json:"acs_url"`      // Assertion Consumer Service URL

	// Request metadata
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `gorm:"type:text" json:"user_agent,omitempty"`

	// Expiration
	ExpiresAt   time.Time  `gorm:"not null;index" json:"expires_at"`
	IsCompleted bool       `gorm:"default:false;index" json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Relationships
	SSOProvider SSOProvider `gorm:"foreignKey:SSOProviderID" json:"-"`
	Tenant      Tenant      `gorm:"foreignKey:TenantID" json:"-"`
}

// SSOAuditLog provides an audit trail for all SSO-related events
type SSOAuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`

	// Multi-tenant association
	TenantID uuid.UUID `gorm:"type:uuid;not null;index:idx_sso_audit_tenant" json:"tenant_id"`

	// Association
	SSOProviderID *uuid.UUID `gorm:"type:uuid;index" json:"sso_provider_id,omitempty"`
	UserID        *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`

	// Event details
	EventType        string `gorm:"not null;index" json:"event_type"` // 'login_success', 'login_failure', etc.
	EventDescription string `gorm:"type:text" json:"event_description,omitempty"`

	// Request details
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `gorm:"type:text" json:"user_agent,omitempty"`

	// SAML-specific
	SAMLRequestID      string `json:"saml_request_id,omitempty"`
	SAMLResponseStatus string `json:"saml_response_status,omitempty"`

	// Metadata
	Metadata SSOMetadata `gorm:"type:jsonb;default:'{}'" json:"metadata"`

	// Relationships
	Tenant      Tenant       `gorm:"foreignKey:TenantID" json:"-"`
	SSOProvider *SSOProvider `gorm:"foreignKey:SSOProviderID" json:"-"`
	User        *User        `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for SSOProvider
func (SSOProvider) TableName() string {
	return "sso_providers"
}

// TableName returns the table name for SSOUserIdentity
func (SSOUserIdentity) TableName() string {
	return "sso_user_identities"
}

// TableName returns the table name for SAMLRequest
func (SAMLRequest) TableName() string {
	return "saml_requests"
}

// TableName returns the table name for SSOAuditLog
func (SSOAuditLog) TableName() string {
	return "sso_audit_logs"
}

// Common SSO event types
const (
	EventLoginSuccess    = "login_success"
	EventLoginFailure    = "login_failure"
	EventLogout          = "logout"
	EventConfigChange    = "config_change"
	EventJITProvision    = "jit_provision"
	EventMetadataRefresh = "metadata_refresh"
	EventCertificateWarn = "certificate_expiring"
)

// Common SAML NameID formats
const (
	NameIDFormatPersistent  = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
	NameIDFormatTransient   = "urn:oasis:names:tc:SAML:2.0:nameid-format:transient"
	NameIDFormatEmail       = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	NameIDFormatUnspecified = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
)

// Provider types
const (
	ProviderTypeSAML  = "saml"
	ProviderTypeOAuth = "oauth"
	ProviderTypeOIDC  = "oidc"
)

// Common provider names
const (
	ProviderNameOkta            = "Okta"
	ProviderNameAzureAD         = "Azure AD"
	ProviderNameGoogleWorkspace = "Google Workspace"
	ProviderNameOneLogin        = "OneLogin"
	ProviderNameAuth0           = "Auth0"
	ProviderNameGenericSAML     = "Generic SAML"
)
