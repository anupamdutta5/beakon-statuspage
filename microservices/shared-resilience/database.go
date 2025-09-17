package resilience

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager manages database connections and operations
type DatabaseManager struct {
	db     *gorm.DB
	config DatabaseConfig
	logger *zap.Logger
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(config DatabaseConfig, zapLogger *zap.Logger) (*DatabaseManager, error) {
	db, err := NewDatabaseConnection(config, zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &DatabaseManager{
		db:     db,
		config: config,
		logger: zapLogger,
	}, nil
}

// NewDatabaseConnection creates a new database connection with proper configuration
func NewDatabaseConnection(config DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// Create GORM logger that integrates with zap
	gormLogger := logger.New(
		&gormZapLogger{zapLogger: zapLogger},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.GetDSN()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zapLogger.Info("Database connection established",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Name),
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
	)

	return db, nil
}

// GetDB returns the GORM database instance
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// HealthCheck performs a database health check
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func (dm *DatabaseManager) GetStats() (sql.DBStats, error) {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return sql.DBStats{}, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Stats(), nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	dm.logger.Info("Closing database connection")
	return sqlDB.Close()
}

// WithContext returns a database instance with the given context
func (dm *DatabaseManager) WithContext(ctx context.Context) *gorm.DB {
	return dm.db.WithContext(ctx)
}

// WithTenant returns a database instance scoped to a specific tenant
func (dm *DatabaseManager) WithTenant(tenantID string) *gorm.DB {
	return dm.db.Where("tenant_id = ?", tenantID)
}

// Transaction executes a function within a database transaction
func (dm *DatabaseManager) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return dm.db.WithContext(ctx).Transaction(fn)
}

// BaseModel provides common fields for all database models
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TenantAwareModel extends BaseModel with tenant awareness
type TenantAwareModel struct {
	BaseModel
	TenantID string `gorm:"not null;index" json:"tenant_id"`
}

// AuditModel extends BaseModel with audit fields
type AuditModel struct {
	BaseModel
	CreatedBy string `gorm:"size:255" json:"created_by"`
	UpdatedBy string `gorm:"size:255" json:"updated_by"`
}

// TenantAuditModel combines tenant awareness and audit fields
type TenantAuditModel struct {
	BaseModel
	TenantID  string `gorm:"not null;index" json:"tenant_id"`
	CreatedBy string `gorm:"size:255" json:"created_by"`
	UpdatedBy string `gorm:"size:255" json:"updated_by"`
}

// BeforeCreate GORM hook to set tenant ID from context
func (m *TenantAwareModel) BeforeCreate(tx *gorm.DB) error {
	if m.TenantID == "" {
		if tenantID, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			m.TenantID = tenantID
		}
	}
	return nil
}

// BeforeUpdate GORM hook to prevent tenant ID changes
func (m *TenantAwareModel) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("tenant_id") {
		return fmt.Errorf("tenant_id cannot be modified")
	}
	return nil
}

// gormZapLogger integrates GORM with zap logger
type gormZapLogger struct {
	zapLogger *zap.Logger
}

func (l *gormZapLogger) Printf(format string, v ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(format, v...))
}

// DatabaseHealthChecker provides health checking functionality
type DatabaseHealthChecker struct {
	manager *DatabaseManager
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(manager *DatabaseManager) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{manager: manager}
}

// Check performs a comprehensive health check
func (hc *DatabaseHealthChecker) Check(ctx context.Context) map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
	}

	// Basic connectivity check
	if err := hc.manager.HealthCheck(ctx); err != nil {
		result["status"] = "unhealthy"
		result["error"] = err.Error()
		return result
	}

	// Get connection statistics
	stats, err := hc.manager.GetStats()
	if err != nil {
		result["status"] = "degraded"
		result["error"] = err.Error()
	} else {
		result["stats"] = map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":          stats.InUse,
			"idle":            stats.Idle,
			"wait_count":      stats.WaitCount,
			"wait_duration":   stats.WaitDuration.String(),
			"max_idle_closed": stats.MaxIdleClosed,
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		}

		// Check for potential issues
		if stats.OpenConnections >= hc.manager.config.MaxOpenConns-1 {
			result["status"] = "degraded"
			result["warning"] = "connection pool nearly exhausted"
		}

		if stats.WaitCount > 0 {
			result["status"] = "degraded"
			result["warning"] = "connections are waiting"
		}
	}

	return result
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB, logger *zap.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger,
	}
}

// AutoMigrate performs automatic migration for the given models
func (mm *MigrationManager) AutoMigrate(models ...interface{}) error {
	for _, model := range models {
		mm.logger.Info("Migrating model", zap.String("model", fmt.Sprintf("%T", model)))
		if err := mm.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}
	return nil
}

// Repository provides base repository functionality
type Repository[T any] struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewRepository creates a new repository for the given model type
func NewRepository[T any](db *gorm.DB, logger *zap.Logger) *Repository[T] {
	return &Repository[T]{
		db:     db,
		logger: logger,
	}
}

// Create creates a new record
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID retrieves a record by ID
func (r *Repository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update updates an existing record
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete soft deletes a record
func (r *Repository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// List retrieves a list of records with pagination
func (r *Repository[T]) List(ctx context.Context, offset, limit int) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&entities).Error
	return entities, err
}

// Count returns the total number of records
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Count(&count).Error
	return count, err
}