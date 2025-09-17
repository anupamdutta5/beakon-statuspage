// Package models provides data transfer objects for the Status Page UI Service.
// These are simplified DTOs for frontend consumption, not database models.
package models

// StatusPageData represents the data structure for rendering the status page.
type StatusPageData struct {
	// Site Configuration
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	LogoURL         string `json:"logo_url"`
	Favicon         string `json:"favicon"`
	OGImage         string `json:"og_image"`
	SiteURL         string `json:"site_url"`
	SubscribeURL    string `json:"subscribe_url"`
	HistoryURL      string `json:"history_url"`
	APIURL          string `json:"api_url"`
	WebSocketURL    string `json:"websocket_url"`
	PoweredBy       string `json:"powered_by"`
	PoweredByURL    string `json:"powered_by_url"`
	LastUpdated     string `json:"last_updated"`

	// Overall Status
	OverallStatus     string `json:"overall_status"`
	StatusTitle       string `json:"status_title"`
	StatusDescription string `json:"status_description"`

	// Components
	Components []ComponentStatus `json:"components"`

	// Incidents
	RecentIncidents []IncidentStatus `json:"recent_incidents"`

	// Maintenance
	ScheduledMaintenance []MaintenanceStatus `json:"scheduled_maintenance"`
}

// ComponentStatus represents component status for the status page.
type ComponentStatus struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	StatusText  string  `json:"status_text"`
	Uptime      float64 `json:"uptime"`
}

// IncidentStatus represents incident status for the status page.
type IncidentStatus struct {
	ID          uint             `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      string           `json:"status"`
	StatusText  string           `json:"status_text"`
	CreatedAt   string           `json:"created_at"`
	Updates     []IncidentUpdate `json:"updates"`
}

// IncidentUpdate represents an incident update for the status page.
type IncidentUpdate struct {
	ID        uint   `json:"id"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// MaintenanceStatus represents maintenance status for the status page.
type MaintenanceStatus struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	Impact        string `json:"impact"`
	ImpactText    string `json:"impact_text"`
	ScheduledTime string `json:"scheduled_time"`
}




