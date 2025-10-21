package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertRoutingRule represents a rule for routing alerts to specific integrations
type AlertRoutingRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Priority    int       `gorm:"default:0;index" json:"priority"` // Higher priority = evaluated first
	IsActive    bool      `gorm:"default:true" json:"is_active"`

	// Conditions (all must match for rule to apply)
	MonitorIDs       string `gorm:"type:text" json:"monitor_ids"`             // Comma-separated, empty = all
	MonitorTypes     string `gorm:"type:text" json:"monitor_types"`           // http,tcp,ping (empty = all)
	Severities       string `gorm:"type:text" json:"severities"`              // down,degraded (empty = all)
	TimeRangeStart   string `gorm:"size:10" json:"time_range_start"`          // HH:MM format, empty = always
	TimeRangeEnd     string `gorm:"size:10" json:"time_range_end"`            // HH:MM format, empty = always
	DaysOfWeek       string `gorm:"type:text" json:"days_of_week"`            // mon,tue,wed,thu,fri,sat,sun (empty = all)
	ComponentIDs     string `gorm:"type:text" json:"component_ids"`           // Comma-separated UUIDs, empty = all
	Tags             string `gorm:"type:text" json:"tags"`                    // Comma-separated tags, empty = all
	CustomConditions string `gorm:"type:jsonb" json:"custom_conditions"`      // JSON for advanced conditions

	// Actions (what integrations to route to)
	RouteToEmail     bool   `gorm:"default:false" json:"route_to_email"`
	RouteToSMS       bool   `gorm:"default:false" json:"route_to_sms"`
	RouteToSlack     bool   `gorm:"default:false" json:"route_to_slack"`
	RouteToTeams     bool   `gorm:"default:false" json:"route_to_teams"`
	RouteToWebhook   bool   `gorm:"default:false" json:"route_to_webhook"`
	RouteToPagerDuty bool   `gorm:"default:false" json:"route_to_pagerduty"`
	StopOnMatch      bool   `gorm:"default:false" json:"stop_on_match"` // Stop evaluating rules after this one matches

	// Integration-specific settings
	IntegrationIDs string `gorm:"type:text" json:"integration_ids"` // Comma-separated integration IDs to route to

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertRoutingLog represents a log entry for alert routing decisions
type AlertRoutingLog struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MonitorID        uint      `gorm:"not null;index" json:"monitor_id"`
	AlertSeverity    string    `gorm:"size:50" json:"alert_severity"`
	RulesEvaluated   int       `gorm:"default:0" json:"rules_evaluated"`
	RulesMatched     int       `gorm:"default:0" json:"rules_matched"`
	MatchedRuleIDs   string    `gorm:"type:text" json:"matched_rule_ids"`   // Comma-separated rule IDs
	IntegrationsUsed string    `gorm:"type:text" json:"integrations_used"`  // Comma-separated integration types
	RoutingDecision  string    `gorm:"type:text" json:"routing_decision"`   // Human-readable decision log
	ProcessingTimeMs int       `gorm:"default:0" json:"processing_time_ms"` // Time taken to evaluate rules
	CreatedAt        time.Time `json:"created_at"`
}

// AlertContext contains information about an alert for routing decisions
type AlertContext struct {
	TenantID      uuid.UUID
	MonitorID     uint
	MonitorName   string
	MonitorType   string
	ComponentID   uuid.UUID
	Severity      string // down, degraded, up, maintenance
	ErrorMessage  string
	Tags          []string
	CustomData    map[string]interface{}
	Timestamp     time.Time
}

// RoutingDecision represents the result of evaluating routing rules
type RoutingDecision struct {
	ShouldRouteToEmail     bool
	ShouldRouteToSMS       bool
	ShouldRouteToSlack     bool
	ShouldRouteToTeams     bool
	ShouldRouteToWebhook   bool
	ShouldRouteToPagerDuty bool
	MatchedRules           []AlertRoutingRule
	IntegrationIDs         []uint
	DecisionLog            []string
}

// AlertRoutingService handles alert routing logic
type AlertRoutingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAlertRoutingService creates a new alert routing service
func NewAlertRoutingService(db *gorm.DB, logger *zap.Logger) *AlertRoutingService {
	return &AlertRoutingService{
		db:     db,
		logger: logger,
	}
}

// CreateRule creates a new alert routing rule
func (s *AlertRoutingService) CreateRule(rule *AlertRoutingRule) error {
	// Validate rule
	if err := s.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	return s.db.Create(rule).Error
}

// UpdateRule updates an existing alert routing rule
func (s *AlertRoutingService) UpdateRule(rule *AlertRoutingRule) error {
	if err := s.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	return s.db.Save(rule).Error
}

// DeleteRule deletes an alert routing rule
func (s *AlertRoutingService) DeleteRule(id uint, tenantID uuid.UUID) error {
	return s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&AlertRoutingRule{}).Error
}

// GetRule retrieves a specific alert routing rule
func (s *AlertRoutingService) GetRule(id uint, tenantID uuid.UUID) (*AlertRoutingRule, error) {
	var rule AlertRoutingRule
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&rule).Error
	return &rule, err
}

// GetRulesByTenant retrieves all alert routing rules for a tenant
func (s *AlertRoutingService) GetRulesByTenant(tenantID uuid.UUID) ([]AlertRoutingRule, error) {
	var rules []AlertRoutingRule
	err := s.db.Where("tenant_id = ?", tenantID).Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}

// GetActiveRulesByTenant retrieves all active alert routing rules for a tenant (sorted by priority)
func (s *AlertRoutingService) GetActiveRulesByTenant(tenantID uuid.UUID) ([]AlertRoutingRule, error) {
	var rules []AlertRoutingRule
	err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("priority DESC, id ASC").
		Find(&rules).Error
	return rules, err
}

// EvaluateRouting evaluates all routing rules for an alert and returns routing decision
func (s *AlertRoutingService) EvaluateRouting(ctx context.Context, alertCtx *AlertContext) (*RoutingDecision, error) {
	startTime := time.Now()

	// Get active rules for tenant
	rules, err := s.GetActiveRulesByTenant(alertCtx.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routing rules: %w", err)
	}

	decision := &RoutingDecision{
		MatchedRules: make([]AlertRoutingRule, 0),
		DecisionLog:  make([]string, 0),
	}

	if len(rules) == 0 {
		decision.DecisionLog = append(decision.DecisionLog, "No routing rules configured - using default routing")
		// Default: route to all integrations
		decision.ShouldRouteToEmail = true
		decision.ShouldRouteToSMS = true
		decision.ShouldRouteToSlack = true
		decision.ShouldRouteToTeams = true
		decision.ShouldRouteToWebhook = true
		decision.ShouldRouteToPagerDuty = true
		return decision, nil
	}

	// Evaluate rules in priority order
	for _, rule := range rules {
		decision.DecisionLog = append(decision.DecisionLog, fmt.Sprintf("Evaluating rule: %s (priority %d)", rule.Name, rule.Priority))

		if s.ruleMatches(&rule, alertCtx) {
			decision.DecisionLog = append(decision.DecisionLog, fmt.Sprintf("✅ Rule matched: %s", rule.Name))
			decision.MatchedRules = append(decision.MatchedRules, rule)

			// Apply routing actions
			decision.ShouldRouteToEmail = decision.ShouldRouteToEmail || rule.RouteToEmail
			decision.ShouldRouteToSMS = decision.ShouldRouteToSMS || rule.RouteToSMS
			decision.ShouldRouteToSlack = decision.ShouldRouteToSlack || rule.RouteToSlack
			decision.ShouldRouteToTeams = decision.ShouldRouteToTeams || rule.RouteToTeams
			decision.ShouldRouteToWebhook = decision.ShouldRouteToWebhook || rule.RouteToWebhook
			decision.ShouldRouteToPagerDuty = decision.ShouldRouteToPagerDuty || rule.RouteToPagerDuty

			// Add integration IDs
			if rule.IntegrationIDs != "" {
				ids := strings.Split(rule.IntegrationIDs, ",")
				for _, idStr := range ids {
					if id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32); err == nil {
						decision.IntegrationIDs = append(decision.IntegrationIDs, uint(id))
					}
				}
			}

			// Stop if this rule says so
			if rule.StopOnMatch {
				decision.DecisionLog = append(decision.DecisionLog, "⛔ Stop-on-match rule - halting evaluation")
				break
			}
		} else {
			decision.DecisionLog = append(decision.DecisionLog, fmt.Sprintf("❌ Rule did not match: %s", rule.Name))
		}
	}

	// Log routing decision
	processingTime := time.Since(startTime).Milliseconds()
	s.logRoutingDecision(alertCtx, decision, len(rules), int(processingTime))

	return decision, nil
}

// ruleMatches checks if a rule matches the alert context
func (s *AlertRoutingService) ruleMatches(rule *AlertRoutingRule, alertCtx *AlertContext) bool {
	// Check monitor IDs
	if rule.MonitorIDs != "" {
		if !s.matchesMonitorID(rule.MonitorIDs, alertCtx.MonitorID) {
			return false
		}
	}

	// Check monitor types
	if rule.MonitorTypes != "" {
		if !s.matchesMonitorType(rule.MonitorTypes, alertCtx.MonitorType) {
			return false
		}
	}

	// Check severities
	if rule.Severities != "" {
		if !s.matchesSeverity(rule.Severities, alertCtx.Severity) {
			return false
		}
	}

	// Check time range
	if rule.TimeRangeStart != "" && rule.TimeRangeEnd != "" {
		if !s.matchesTimeRange(rule.TimeRangeStart, rule.TimeRangeEnd, alertCtx.Timestamp) {
			return false
		}
	}

	// Check days of week
	if rule.DaysOfWeek != "" {
		if !s.matchesDayOfWeek(rule.DaysOfWeek, alertCtx.Timestamp) {
			return false
		}
	}

	// Check component IDs
	if rule.ComponentIDs != "" {
		if !s.matchesComponentID(rule.ComponentIDs, alertCtx.ComponentID) {
			return false
		}
	}

	// Check tags
	if rule.Tags != "" {
		if !s.matchesTags(rule.Tags, alertCtx.Tags) {
			return false
		}
	}

	// All conditions matched
	return true
}

// matchesMonitorID checks if monitor ID is in the list
func (s *AlertRoutingService) matchesMonitorID(monitorIDs string, monitorID uint) bool {
	ids := strings.Split(monitorIDs, ",")
	monitorIDStr := fmt.Sprintf("%d", monitorID)

	for _, id := range ids {
		if strings.TrimSpace(id) == monitorIDStr {
			return true
		}
	}
	return false
}

// matchesMonitorType checks if monitor type is in the list
func (s *AlertRoutingService) matchesMonitorType(types string, monitorType string) bool {
	typeList := strings.Split(types, ",")
	for _, t := range typeList {
		if strings.ToLower(strings.TrimSpace(t)) == strings.ToLower(monitorType) {
			return true
		}
	}
	return false
}

// matchesSeverity checks if severity is in the list
func (s *AlertRoutingService) matchesSeverity(severities string, severity string) bool {
	sevList := strings.Split(severities, ",")
	for _, sev := range sevList {
		if strings.ToLower(strings.TrimSpace(sev)) == strings.ToLower(severity) {
			return true
		}
	}
	return false
}

// matchesTimeRange checks if current time is within the specified time range
func (s *AlertRoutingService) matchesTimeRange(startTime, endTime string, timestamp time.Time) bool {
	// Parse start time (HH:MM format)
	startParts := strings.Split(startTime, ":")
	if len(startParts) != 2 {
		return true // Invalid format, allow by default
	}

	// Parse end time (HH:MM format)
	endParts := strings.Split(endTime, ":")
	if len(endParts) != 2 {
		return true // Invalid format, allow by default
	}

	startHour, _ := strconv.Atoi(startParts[0])
	startMin, _ := strconv.Atoi(startParts[1])
	endHour, _ := strconv.Atoi(endParts[0])
	endMin, _ := strconv.Atoi(endParts[1])

	currentHour := timestamp.Hour()
	currentMin := timestamp.Minute()

	// Convert to minutes since midnight for easier comparison
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin
	currentMinutes := currentHour*60 + currentMin

	// Handle time ranges that cross midnight
	if startMinutes > endMinutes {
		// Range crosses midnight (e.g., 22:00 to 06:00)
		return currentMinutes >= startMinutes || currentMinutes <= endMinutes
	}

	// Normal time range
	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

// matchesDayOfWeek checks if current day is in the list
func (s *AlertRoutingService) matchesDayOfWeek(daysOfWeek string, timestamp time.Time) bool {
	daysList := strings.Split(strings.ToLower(daysOfWeek), ",")
	currentDay := strings.ToLower(timestamp.Format("Mon"))

	for _, day := range daysList {
		if strings.TrimSpace(day) == currentDay {
			return true
		}
	}
	return false
}

// matchesComponentID checks if component ID is in the list
func (s *AlertRoutingService) matchesComponentID(componentIDs string, componentID uuid.UUID) bool {
	ids := strings.Split(componentIDs, ",")
	componentIDStr := componentID.String()

	for _, id := range ids {
		if strings.TrimSpace(id) == componentIDStr {
			return true
		}
	}
	return false
}

// matchesTags checks if any of the alert tags match the rule tags
func (s *AlertRoutingService) matchesTags(ruleTags string, alertTags []string) bool {
	ruleTagList := strings.Split(strings.ToLower(ruleTags), ",")

	for _, alertTag := range alertTags {
		alertTagLower := strings.ToLower(strings.TrimSpace(alertTag))
		for _, ruleTag := range ruleTagList {
			if strings.TrimSpace(ruleTag) == alertTagLower {
				return true
			}
		}
	}
	return false
}

// validateRule validates a routing rule
func (s *AlertRoutingService) validateRule(rule *AlertRoutingRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	// Validate time range format
	if rule.TimeRangeStart != "" {
		if !s.isValidTimeFormat(rule.TimeRangeStart) {
			return fmt.Errorf("invalid time_range_start format, expected HH:MM")
		}
	}

	if rule.TimeRangeEnd != "" {
		if !s.isValidTimeFormat(rule.TimeRangeEnd) {
			return fmt.Errorf("invalid time_range_end format, expected HH:MM")
		}
	}

	// At least one routing action must be enabled
	if !rule.RouteToEmail && !rule.RouteToSMS && !rule.RouteToSlack &&
		!rule.RouteToTeams && !rule.RouteToWebhook && !rule.RouteToPagerDuty {
		return fmt.Errorf("at least one routing action must be enabled")
	}

	return nil
}

// isValidTimeFormat checks if time is in HH:MM format
func (s *AlertRoutingService) isValidTimeFormat(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}

	hour, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return false
	}

	return hour >= 0 && hour <= 23 && min >= 0 && min <= 59
}

// logRoutingDecision logs a routing decision to the database
func (s *AlertRoutingService) logRoutingDecision(alertCtx *AlertContext, decision *RoutingDecision, rulesEvaluated int, processingTimeMs int) {
	// Build matched rule IDs
	var matchedRuleIDs []string
	for _, rule := range decision.MatchedRules {
		matchedRuleIDs = append(matchedRuleIDs, fmt.Sprintf("%d", rule.ID))
	}

	// Build integrations used
	var integrationsUsed []string
	if decision.ShouldRouteToEmail {
		integrationsUsed = append(integrationsUsed, "email")
	}
	if decision.ShouldRouteToSMS {
		integrationsUsed = append(integrationsUsed, "sms")
	}
	if decision.ShouldRouteToSlack {
		integrationsUsed = append(integrationsUsed, "slack")
	}
	if decision.ShouldRouteToTeams {
		integrationsUsed = append(integrationsUsed, "teams")
	}
	if decision.ShouldRouteToWebhook {
		integrationsUsed = append(integrationsUsed, "webhook")
	}
	if decision.ShouldRouteToPagerDuty {
		integrationsUsed = append(integrationsUsed, "pagerduty")
	}

	log := &AlertRoutingLog{
		TenantID:         alertCtx.TenantID,
		MonitorID:        alertCtx.MonitorID,
		AlertSeverity:    alertCtx.Severity,
		RulesEvaluated:   rulesEvaluated,
		RulesMatched:     len(decision.MatchedRules),
		MatchedRuleIDs:   strings.Join(matchedRuleIDs, ","),
		IntegrationsUsed: strings.Join(integrationsUsed, ","),
		RoutingDecision:  strings.Join(decision.DecisionLog, " | "),
		ProcessingTimeMs: processingTimeMs,
	}

	if err := s.db.Create(log).Error; err != nil {
		s.logger.Error("Failed to log routing decision", zap.Error(err))
	}
}

// GetRoutingLogs retrieves routing logs for a tenant
func (s *AlertRoutingService) GetRoutingLogs(tenantID uuid.UUID, limit int) ([]AlertRoutingLog, error) {
	var logs []AlertRoutingLog
	err := s.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetRoutingStats retrieves routing statistics for a tenant
func (s *AlertRoutingService) GetRoutingStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var totalAlerts int64
	var avgProcessingTime float64
	var topRules []map[string]interface{}

	// Total alerts routed
	s.db.Model(&AlertRoutingLog{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Count(&totalAlerts)

	// Average processing time
	s.db.Model(&AlertRoutingLog{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Select("AVG(processing_time_ms) as avg_time").
		Scan(&avgProcessingTime)

	// Top rules by usage
	type RuleUsage struct {
		RuleID     uint   `json:"rule_id"`
		RuleName   string `json:"rule_name"`
		MatchCount int64  `json:"match_count"`
	}

	var ruleUsages []RuleUsage
	s.db.Raw(`
		SELECT
			CAST(UNNEST(string_to_array(matched_rule_ids, ',')) AS INTEGER) as rule_id,
			COUNT(*) as match_count
		FROM alert_routing_logs
		WHERE tenant_id = ? AND created_at BETWEEN ? AND ?
		AND matched_rule_ids != ''
		GROUP BY rule_id
		ORDER BY match_count DESC
		LIMIT 5
	`, tenantID, startDate, endDate).Scan(&ruleUsages)

	// Get rule names
	for i := range ruleUsages {
		var rule AlertRoutingRule
		if err := s.db.First(&rule, ruleUsages[i].RuleID).Error; err == nil {
			ruleUsages[i].RuleName = rule.Name
		}
	}

	// Convert to map
	for _, usage := range ruleUsages {
		topRules = append(topRules, map[string]interface{}{
			"rule_id":     usage.RuleID,
			"rule_name":   usage.RuleName,
			"match_count": usage.MatchCount,
		})
	}

	// Integration usage stats
	var integrationStats []map[string]interface{}
	integrationTypes := []string{"email", "sms", "slack", "teams", "webhook", "pagerduty"}

	for _, intType := range integrationTypes {
		var count int64
		s.db.Model(&AlertRoutingLog{}).
			Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
			Where("integrations_used LIKE ?", "%"+intType+"%").
			Count(&count)

		if count > 0 {
			integrationStats = append(integrationStats, map[string]interface{}{
				"integration": intType,
				"count":       count,
			})
		}
	}

	return map[string]interface{}{
		"total_alerts":         totalAlerts,
		"avg_processing_ms":    avgProcessingTime,
		"top_rules":            topRules,
		"integration_usage":    integrationStats,
	}, nil
}

// TestRule tests a routing rule against sample data without saving
func (s *AlertRoutingService) TestRule(rule *AlertRoutingRule, alertCtx *AlertContext) (bool, string) {
	if err := s.validateRule(rule); err != nil {
		return false, fmt.Sprintf("Rule validation failed: %v", err)
	}

	matches := s.ruleMatches(rule, alertCtx)
	if matches {
		return true, "Rule matches the provided alert context"
	}

	return false, "Rule does not match the provided alert context"
}
