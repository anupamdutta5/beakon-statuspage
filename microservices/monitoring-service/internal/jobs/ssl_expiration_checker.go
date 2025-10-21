// Package jobs provides background job runners for the Monitoring Service.
package jobs

import (
	"log"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/events"
	"github.com/anupamdutta5/monitoring-service/internal/services"
)

// SSLExpirationChecker is a background job that checks for expiring SSL certificates
// and sends warning events via RabbitMQ.
type SSLExpirationChecker struct {
	db             *gorm.DB
	sslService     *services.SSLScannerService
	eventPublisher *events.EventPublisher
	interval       time.Duration
	stopChan       chan struct{}
}

// NewSSLExpirationChecker creates a new SSL expiration checker job.
func NewSSLExpirationChecker(db *gorm.DB, eventPublisher *events.EventPublisher, checkInterval time.Duration) *SSLExpirationChecker {
	return &SSLExpirationChecker{
		db:             db,
		sslService:     services.NewSSLScannerService(db),
		eventPublisher: eventPublisher,
		interval:       checkInterval,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the SSL expiration checking job.
func (j *SSLExpirationChecker) Start() {
	log.Printf("🚀 Starting SSL Expiration Checker (interval: %s)", j.interval)

	// Run immediately on start
	j.checkExpirations()

	// Then run on interval
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.checkExpirations()
		case <-j.stopChan:
			log.Println("⏹️  Stopping SSL Expiration Checker")
			return
		}
	}
}

// Stop stops the SSL expiration checking job.
func (j *SSLExpirationChecker) Stop() {
	close(j.stopChan)
}

// checkExpirations performs the SSL expiration check.
func (j *SSLExpirationChecker) checkExpirations() {
	log.Println("🔍 Checking for expiring SSL certificates...")

	certs, err := j.sslService.GetCertificatesNeedingWarning()
	if err != nil {
		log.Printf("❌ Error fetching expiring certificates: %v", err)
		return
	}

	if len(certs) == 0 {
		log.Println("✅ No expiring certificates found")
		return
	}

	log.Printf("⚠️  Found %d certificate(s) needing warnings", len(certs))

	sentCount := 0
	for _, cert := range certs {
		shouldSend, warningType := cert.ShouldSendWarning()
		if !shouldSend {
			continue
		}

		// Create and publish SSL expiring event
		event := events.SSLExpiringEvent{
			TenantID:        cert.TenantID,
			CertificateID:   cert.ID,
			Domain:          cert.Domain,
			DaysUntilExpiry: *cert.DaysUntilExpiry,
			ValidUntil:      *cert.ValidUntil,
			WarningType:     warningType,
		}

		if err := j.eventPublisher.PublishSSLExpiring(event); err != nil {
			log.Printf("❌ Failed to publish SSL expiring event for %s: %v", cert.Domain, err)
			continue
		}

		// Mark warning as sent
		if err := j.sslService.MarkWarningSent(cert.ID, warningType); err != nil {
			log.Printf("❌ Failed to mark warning as sent for %s: %v", cert.Domain, err)
			continue
		}

		log.Printf("✅ Sent %s warning for %s (expires in %d days)", warningType, cert.Domain, *cert.DaysUntilExpiry)
		sentCount++
	}

	log.Printf("📤 Sent %d SSL expiration warning(s)", sentCount)
}

// CertificateRescanJob handles periodic rescanning of SSL certificates.
type CertificateRescanJob struct {
	db         *gorm.DB
	sslService *services.SSLScannerService
	interval   time.Duration
	stopChan   chan struct{}
}

// NewCertificateRescanJob creates a new certificate rescan job.
func NewCertificateRescanJob(db *gorm.DB, rescanInterval time.Duration) *CertificateRescanJob {
	return &CertificateRescanJob{
		db:         db,
		sslService: services.NewSSLScannerService(db),
		interval:   rescanInterval,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the certificate rescanning job.
func (j *CertificateRescanJob) Start() {
	log.Printf("🚀 Starting Certificate Rescan Job (interval: %s)", j.interval)

	// Run on interval
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.rescanCertificates()
		case <-j.stopChan:
			log.Println("⏹️  Stopping Certificate Rescan Job")
			return
		}
	}
}

// Stop stops the certificate rescanning job.
func (j *CertificateRescanJob) Stop() {
	close(j.stopChan)
}

// rescanCertificates performs the certificate rescanning.
func (j *CertificateRescanJob) rescanCertificates() {
	log.Println("🔄 Rescanning outdated SSL certificates...")

	scannedCount, errors := j.sslService.RescanExpiredCertificates()

	if len(errors) > 0 {
		log.Printf("⚠️  Rescanned %d certificates with %d errors", scannedCount, len(errors))
		for _, err := range errors {
			log.Printf("  ❌ %v", err)
		}
	} else {
		log.Printf("✅ Successfully rescanned %d certificate(s)", scannedCount)
	}
}

// HeartbeatCheckerJob checks for overdue heartbeats and sends alerts.
type HeartbeatCheckerJob struct {
	db                *gorm.DB
	heartbeatService  *services.HeartbeatService
	logger            *zap.Logger
	interval          time.Duration
	stopChan          chan struct{}
}

// NewHeartbeatCheckerJob creates a new heartbeat checker job.
func NewHeartbeatCheckerJob(db *gorm.DB, logger *zap.Logger, checkInterval time.Duration) *HeartbeatCheckerJob {
	return &HeartbeatCheckerJob{
		db:               db,
		heartbeatService: services.NewHeartbeatService(db, logger),
		logger:           logger,
		interval:         checkInterval,
		stopChan:         make(chan struct{}),
	}
}

// Start begins the heartbeat checking job.
func (j *HeartbeatCheckerJob) Start() {
	j.logger.Info("Starting Heartbeat Checker Job", zap.Duration("interval", j.interval))

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.checkHeartbeats()
		case <-j.stopChan:
			j.logger.Info("Stopping Heartbeat Checker Job")
			return
		}
	}
}

// Stop stops the heartbeat checking job.
func (j *HeartbeatCheckerJob) Stop() {
	close(j.stopChan)
}

// checkHeartbeats performs the heartbeat check.
func (j *HeartbeatCheckerJob) checkHeartbeats() {
	j.logger.Debug("Checking for overdue heartbeats")

	err := j.heartbeatService.CheckOverdueHeartbeats()
	if err != nil {
		j.logger.Error("Failed to check overdue heartbeats", zap.Error(err))
	}
}

// MaintenanceWindowJob automatically starts and completes maintenance windows based on schedule.
type MaintenanceWindowJob struct {
	db                  *gorm.DB
	maintenanceService  *services.MaintenanceManagementService
	logger              *zap.Logger
	interval            time.Duration
	stopChan            chan struct{}
}

// NewMaintenanceWindowJob creates a new maintenance window job.
func NewMaintenanceWindowJob(db *gorm.DB, logger *zap.Logger, checkInterval time.Duration) *MaintenanceWindowJob {
	return &MaintenanceWindowJob{
		db:                 db,
		maintenanceService: services.NewMaintenanceManagementService(db, logger),
		logger:             logger,
		interval:           checkInterval,
		stopChan:           make(chan struct{}),
	}
}

// Start begins the maintenance window job.
func (j *MaintenanceWindowJob) Start() {
	j.logger.Info("Starting Maintenance Window Job", zap.Duration("interval", j.interval))

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.processMaintenanceWindows()
		case <-j.stopChan:
			j.logger.Info("Stopping Maintenance Window Job")
			return
		}
	}
}

// Stop stops the maintenance window job.
func (j *MaintenanceWindowJob) Stop() {
	close(j.stopChan)
}

// processMaintenanceWindows checks and auto-starts/completes maintenance windows.
func (j *MaintenanceWindowJob) processMaintenanceWindows() {
	j.logger.Debug("Processing maintenance windows")

	// Auto-start scheduled maintenance windows
	err := j.maintenanceService.AutoStartMaintenanceWindows()
	if err != nil {
		j.logger.Error("Failed to auto-start maintenance windows", zap.Error(err))
	}

	// Auto-complete expired maintenance windows
	err = j.maintenanceService.AutoCompleteMaintenanceWindows()
	if err != nil {
		j.logger.Error("Failed to auto-complete maintenance windows", zap.Error(err))
	}
}

// EscalationProcessorJob processes active escalations and escalates to next level when needed.
type EscalationProcessorJob struct {
	db                 *gorm.DB
	escalationService  *services.EscalationService
	logger             *zap.Logger
	interval           time.Duration
	stopChan           chan struct{}
}

// NewEscalationProcessorJob creates a new escalation processor job.
func NewEscalationProcessorJob(
	db *gorm.DB,
	logger *zap.Logger,
	smsService *services.SMSService,
	onCallService *services.OnCallService,
	checkInterval time.Duration,
) *EscalationProcessorJob {
	return &EscalationProcessorJob{
		db:                db,
		escalationService: services.NewEscalationService(db, logger, smsService, onCallService),
		logger:            logger,
		interval:          checkInterval,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the escalation processor job.
func (j *EscalationProcessorJob) Start() {
	j.logger.Info("Starting Escalation Processor Job", zap.Duration("interval", j.interval))

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.processEscalations()
		case <-j.stopChan:
			j.logger.Info("Stopping Escalation Processor Job")
			return
		}
	}
}

// Stop stops the escalation processor job.
func (j *EscalationProcessorJob) Stop() {
	close(j.stopChan)
}

// processEscalations performs the escalation processing.
func (j *EscalationProcessorJob) processEscalations() {
	j.logger.Debug("Processing active escalations")

	err := j.escalationService.ProcessEscalations()
	if err != nil {
		j.logger.Error("Failed to process escalations", zap.Error(err))
	}
}

// WebhookRetryJob retries failed webhook deliveries with exponential backoff.
type WebhookRetryJob struct {
	db             *gorm.DB
	webhookService *services.WebhookService
	logger         *zap.Logger
	interval       time.Duration
	stopChan       chan struct{}
}

// NewWebhookRetryJob creates a new webhook retry job.
func NewWebhookRetryJob(db *gorm.DB, logger *zap.Logger, checkInterval time.Duration) *WebhookRetryJob {
	return &WebhookRetryJob{
		db:             db,
		webhookService: services.NewWebhookService(db, logger),
		logger:         logger,
		interval:       checkInterval,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the webhook retry job.
func (j *WebhookRetryJob) Start() {
	j.logger.Info("Starting Webhook Retry Job", zap.Duration("interval", j.interval))

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.retryFailedWebhooks()
		case <-j.stopChan:
			j.logger.Info("Stopping Webhook Retry Job")
			return
		}
	}
}

// Stop stops the webhook retry job.
func (j *WebhookRetryJob) Stop() {
	close(j.stopChan)
}

// retryFailedWebhooks retries webhooks that are due for retry.
func (j *WebhookRetryJob) retryFailedWebhooks() {
	j.logger.Debug("Retrying failed webhook deliveries")

	err := j.webhookService.RetryFailedDeliveries()
	if err != nil {
		j.logger.Error("Failed to retry webhook deliveries", zap.Error(err))
	}
}
