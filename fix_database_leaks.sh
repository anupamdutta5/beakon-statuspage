#!/bin/bash

# Script to fix database connection leaks across microservices

echo "Fixing database connection leaks..."

# Create enhanced database manager template
create_database_manager() {
    local service_path="$1"
    local service_name="$2"

    # Create internal/database directory if it doesn't exist
    mkdir -p "$service_path/internal/database"

    # Create enhanced database manager
    cat > "$service_path/internal/database/manager.go" << 'EOF'
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/anupamdutta5/statuspage-shared-resilience"
)

// Manager handles database connections with proper lifecycle management
type Manager struct {
	sqlDB        *sql.DB
	gormDB       *gorm.DB
	logger       *zap.Logger
	healthTicker *time.Ticker
	closeOnce    sync.Once
	closed       chan struct{}
	config       Config
}

// Config represents database configuration
type Config struct {
	Host            string `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port            int    `yaml:"port" env:"DB_PORT" default:"5432"`
	User            string `yaml:"user" env:"DB_USER" default:"postgres"`
	Password        string `yaml:"password" env:"DB_PASSWORD"`
	Name            string `yaml:"name" env:"DB_NAME" default:"statuspage"`
	SSLMode         string `yaml:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"5"`
	MaxLifetime     int    `yaml:"max_lifetime" env:"DB_MAX_LIFETIME" default:"3600"`
	ConnectTimeout  int    `yaml:"connect_timeout" env:"DB_CONNECT_TIMEOUT" default:"10"`
	HealthCheckInterval int `yaml:"health_check_interval" env:"DB_HEALTH_CHECK_INTERVAL" default:"30"`
}

// NewManager creates a new database manager with proper connection pooling
func NewManager(config Config, logger *zap.Logger) (*Manager, error) {
	manager := &Manager{
		logger: logger,
		closed: make(chan struct{}),
		config: config,
	}

	if err := manager.connect(); err != nil {
		return nil, err
	}

	// Start health check monitoring
	manager.startHealthCheck()

	return manager, nil
}

// connect establishes database connections with proper configuration
func (m *Manager) connect() error {
	m.logger.Info("Connecting to database",
		zap.String("host", m.config.Host),
		zap.Int("port", m.config.Port),
		zap.String("database", m.config.Name),
	)

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		m.config.Host,
		m.config.Port,
		m.config.User,
		m.config.Password,
		m.config.Name,
		m.config.SSLMode,
		m.config.ConnectTimeout,
	)

	// Open SQL connection
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(m.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(m.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(m.config.MaxLifetime) * time.Second)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.config.ConnectTimeout)*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create GORM connection
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.NewGormLogger(m.logger),
	})
	if err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to create GORM connection: %w", err)
	}

	m.sqlDB = sqlDB
	m.gormDB = gormDB

	m.logger.Info("Successfully connected to database",
		zap.Int("max_open_conns", m.config.MaxOpenConns),
		zap.Int("max_idle_conns", m.config.MaxIdleConns),
		zap.Int("max_lifetime", m.config.MaxLifetime),
	)

	return nil
}

// startHealthCheck starts periodic database health checks
func (m *Manager) startHealthCheck() {
	m.healthTicker = time.NewTicker(time.Duration(m.config.HealthCheckInterval) * time.Second)

	go func() {
		for {
			select {
			case <-m.healthTicker.C:
				if err := m.healthCheck(); err != nil {
					m.logger.Error("Database health check failed", zap.Error(err))

					// Attempt to reconnect
					if err := m.reconnect(); err != nil {
						m.logger.Error("Failed to reconnect to database", zap.Error(err))
					}
				}
			case <-m.closed:
				return
			}
		}
	}()
}

// healthCheck performs a database health check
func (m *Manager) healthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.sqlDB.PingContext(ctx)
}

// reconnect attempts to reconnect to the database
func (m *Manager) reconnect() error {
	m.logger.Warn("Attempting to reconnect to database...")

	// Close existing connections
	if m.sqlDB != nil {
		m.sqlDB.Close()
	}

	// Reconnect
	return m.connect()
}

// GetDB returns the GORM database instance
func (m *Manager) GetDB() *gorm.DB {
	return m.gormDB
}

// GetSQLDB returns the SQL database instance
func (m *Manager) GetSQLDB() *sql.DB {
	return m.sqlDB
}

// GetStats returns database connection statistics
func (m *Manager) GetStats() sql.DBStats {
	if m.sqlDB == nil {
		return sql.DBStats{}
	}
	return m.sqlDB.Stats()
}

// HealthCheck returns the health status of the database
func (m *Manager) HealthCheck(ctx context.Context) error {
	if m.sqlDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return m.sqlDB.PingContext(ctx)
}

// Close gracefully closes all database connections
func (m *Manager) Close() error {
	var err error

	m.closeOnce.Do(func() {
		m.logger.Info("Closing database connections...")

		// Signal health check to stop
		close(m.closed)

		// Stop health check ticker
		if m.healthTicker != nil {
			m.healthTicker.Stop()
		}

		// Close GORM connection
		if m.gormDB != nil {
			if sqlDB, gormErr := m.gormDB.DB(); gormErr == nil {
				sqlDB.Close()
			}
		}

		// Close SQL connection
		if m.sqlDB != nil {
			if closeErr := m.sqlDB.Close(); closeErr != nil {
				err = fmt.Errorf("failed to close SQL database connection: %w", closeErr)
				m.logger.Error("Failed to close SQL database connection", zap.Error(closeErr))
			}
		}

		m.logger.Info("Database connections closed successfully")
	})

	return err
}

// WithTransaction executes a function within a database transaction
func (m *Manager) WithTransaction(fn func(*gorm.DB) error) error {
	tx := m.gormDB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			m.logger.Error("Transaction rolled back due to panic", zap.Any("panic", r))
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
EOF

    echo "Created enhanced database manager for $service_name"
}

# Create GORM logger adapter
create_gorm_logger() {
    local service_path="$1"

    cat > "$service_path/internal/database/logger.go" << 'EOF'
package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger adapts zap logger for GORM
type GormLogger struct {
	ZapLogger *zap.Logger
	LogLevel  gormlogger.LogLevel
}

// NewGormLogger creates a new GORM logger adapter
func NewGormLogger(zapLogger *zap.Logger) gormlogger.Interface {
	return &GormLogger{
		ZapLogger: zapLogger,
		LogLevel:  gormlogger.Info,
	}
}

// LogMode sets the log level
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &GormLogger{
		ZapLogger: l.ZapLogger,
		LogLevel:  level,
	}
}

// Info logs info messages
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warn messages
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs error messages
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL queries
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.ZapLogger.Error("Database query error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case elapsed > 200*time.Millisecond && l.LogLevel >= gormlogger.Warn:
		l.ZapLogger.Warn("Slow database query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.LogLevel == gormlogger.Info:
		l.ZapLogger.Info("Database query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
EOF
}

# Fix server implementations to use proper database management
fix_server_database_usage() {
    local service_path="$1"
    local service_name="$2"

    local server_file="$service_path/internal/server/server.go"
    local main_file="$service_path/cmd/main.go"

    if [ -f "$server_file" ]; then
        echo "Fixing database usage in $server_file"

        # Create backup
        cp "$server_file" "$server_file.backup"

        # Remove problematic defer in New function and fix database lifecycle
        sed -i '' '
            /defer func() {/,/}()/ {
                /defer func() {/ {
                    i\
\	// Database connection will be managed by the Manager lifecycle
                    d
                }
                /if err := database\.CloseDatabase/ d
                /logger\.Error.*Failed to close database/ d
                /}()/ d
            }
        ' "$server_file"
    fi

    if [ -f "$main_file" ]; then
        echo "Updating main.go for proper database lifecycle management"

        # This is a placeholder - actual implementation would require
        # more sophisticated text processing to update main.go properly
        echo "// TODO: Update $main_file to use database.Manager instead of direct connections"
    fi
}

# Find all services with database usage and fix them
find microservices -mindepth 1 -maxdepth 1 -type d -not -name "shared-resilience" | while read -r service_dir; do
    service_name=$(basename "$service_dir")

    # Check if service uses database
    if find "$service_dir" -name "*.go" -type f -exec grep -l "sql\.Open\|gorm\.Open\|database" {} \; | head -1 >/dev/null; then
        echo "Processing service with database usage: $service_name"

        create_database_manager "$service_dir" "$service_name"
        create_gorm_logger "$service_dir"
        fix_server_database_usage "$service_dir" "$service_name"
    else
        echo "Skipping service without database usage: $service_name"
    fi
done

echo ""
echo "Database connection leak fixes implemented!"
echo ""
echo "Manual steps required:"
echo "1. Update each service's main.go to use database.Manager"
echo "2. Ensure proper database.Manager.Close() in shutdown handlers"
echo "3. Update service constructors to accept database.Manager instead of direct connections"
echo "4. Test connection pooling and health checks"
echo ""
echo "Example usage pattern:"
echo 'dbManager, err := database.NewManager(config.Database, logger)'
echo 'defer dbManager.Close()'
echo 'service := services.NewService(dbManager.GetDB(), logger)'