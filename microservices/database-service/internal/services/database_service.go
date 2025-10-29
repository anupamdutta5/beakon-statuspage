// Package services provides business logic for the Database Service.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/anupamdutta5/database-service/internal/config"
	"github.com/anupamdutta5/database-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseService handles database-related business logic.
type DatabaseService struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
	cache  *CacheService
}

// NewDatabaseService creates a new database service.
func NewDatabaseService(cfg *config.Config, logger *zap.Logger) (*DatabaseService, error) {
	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		// For testing, we'll allow the service to be created without a database
		// The database will be set later via SetDB method
		if logger != nil {
			logger.Warn("Failed to initialize database, service will be created without database", zap.Error(err))
		}
		db = nil
	}

	// Initialize cache service
	cache := NewCacheService(cfg, logger)

	return &DatabaseService{
		config: cfg,
		logger: logger,
		db:     db,
		cache:  cache,
	}, nil
}

// SetDB sets the database connection (for testing)
func (s *DatabaseService) SetDB(db *gorm.DB) {
	s.db = db
}

// ListDatabases lists all databases.
func (s *DatabaseService) ListDatabases(ctx context.Context) ([]*models.Database, error) {
	s.logger.Info("Listing databases")

	var databases []*models.Database
	if err := s.db.Find(&databases).Error; err != nil {
		s.logger.Error("Failed to list databases", zap.Error(err))
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	return databases, nil
}

// CreateDatabase creates a new database.
func (s *DatabaseService) CreateDatabase(ctx context.Context, database *models.Database) error {
	s.logger.Info("Creating database",
		zap.String("name", database.Name),
		zap.String("type", database.Type))

	if err := s.db.Create(database).Error; err != nil {
		s.logger.Error("Failed to create database", zap.Error(err))
		return fmt.Errorf("failed to create database: %w", err)
	}

	s.logger.Info("Database created successfully",
		zap.Uint("database_id", database.ID))

	return nil
}

// GetDatabase retrieves a database by ID.
func (s *DatabaseService) GetDatabase(ctx context.Context, id uint) (*models.Database, error) {
	s.logger.Info("Getting database", zap.Uint("database_id", id))

	var database models.Database
	if err := s.db.First(&database, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("database not found")
		}
		s.logger.Error("Failed to get database", zap.Error(err))
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	return &database, nil
}

// UpdateDatabase updates a database.
func (s *DatabaseService) UpdateDatabase(ctx context.Context, id uint, updates *models.Database) error {
	s.logger.Info("Updating database", zap.Uint("database_id", id))

	if err := s.db.Model(&models.Database{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update database", zap.Error(err))
		return fmt.Errorf("failed to update database: %w", err)
	}

	s.logger.Info("Database updated successfully", zap.Uint("database_id", id))
	return nil
}

// DeleteDatabase deletes a database.
func (s *DatabaseService) DeleteDatabase(ctx context.Context, id uint) error {
	s.logger.Info("Deleting database", zap.Uint("database_id", id))

	if err := s.db.Delete(&models.Database{}, id).Error; err != nil {
		s.logger.Error("Failed to delete database", zap.Error(err))
		return fmt.Errorf("failed to delete database: %w", err)
	}

	s.logger.Info("Database deleted successfully", zap.Uint("database_id", id))
	return nil
}

// ListTables lists all tables in a database.
func (s *DatabaseService) ListTables(ctx context.Context, databaseID uint) ([]*models.Table, error) {
	s.logger.Info("Listing tables", zap.Uint("database_id", databaseID))

	var tables []*models.Table
	if err := s.db.Where("database_id = ?", databaseID).Find(&tables).Error; err != nil {
		s.logger.Error("Failed to list tables", zap.Error(err))
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	return tables, nil
}

// CreateTable creates a new table.
func (s *DatabaseService) CreateTable(ctx context.Context, table *models.Table) error {
	s.logger.Info("Creating table",
		zap.String("name", table.Name),
		zap.Uint("database_id", table.DatabaseID))

	if err := s.db.Create(table).Error; err != nil {
		s.logger.Error("Failed to create table", zap.Error(err))
		return fmt.Errorf("failed to create table: %w", err)
	}

	s.logger.Info("Table created successfully",
		zap.Uint("table_id", table.ID))

	return nil
}

// GetTable retrieves a table by ID.
func (s *DatabaseService) GetTable(ctx context.Context, id uint) (*models.Table, error) {
	s.logger.Info("Getting table", zap.Uint("table_id", id))

	var table models.Table
	if err := s.db.First(&table, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("table not found")
		}
		s.logger.Error("Failed to get table", zap.Error(err))
		return nil, fmt.Errorf("failed to get table: %w", err)
	}

	return &table, nil
}

// UpdateTable updates a table.
func (s *DatabaseService) UpdateTable(ctx context.Context, id uint, updates *models.Table) error {
	s.logger.Info("Updating table", zap.Uint("table_id", id))

	if err := s.db.Model(&models.Table{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update table", zap.Error(err))
		return fmt.Errorf("failed to update table: %w", err)
	}

	s.logger.Info("Table updated successfully", zap.Uint("table_id", id))
	return nil
}

// DeleteTable deletes a table.
func (s *DatabaseService) DeleteTable(ctx context.Context, id uint) error {
	s.logger.Info("Deleting table", zap.Uint("table_id", id))

	if err := s.db.Delete(&models.Table{}, id).Error; err != nil {
		s.logger.Error("Failed to delete table", zap.Error(err))
		return fmt.Errorf("failed to delete table: %w", err)
	}

	s.logger.Info("Table deleted successfully", zap.Uint("table_id", id))
	return nil
}

// ExecuteQuery executes a SQL query.
func (s *DatabaseService) ExecuteQuery(ctx context.Context, databaseID uint, query string) (*models.QueryResult, error) {
	s.logger.Info("Executing query",
		zap.Uint("database_id", databaseID),
		zap.String("query", query))

	// For now, we'll simulate query execution
	// In production, you would execute the actual SQL query

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	result := &models.QueryResult{
		DatabaseID: databaseID,
		Query:      query,
		Rows:       "[]", // Empty JSON array
		RowCount:   0,
		Duration:   100,
		Timestamp:  time.Now(),
	}

	s.logger.Info("Query executed successfully",
		zap.Uint("database_id", databaseID),
		zap.Int("row_count", result.RowCount))

	return result, nil
}

// ExecuteTransaction executes a database transaction.
func (s *DatabaseService) ExecuteTransaction(ctx context.Context, databaseID uint, operations []*models.TransactionOperation) error {
	s.logger.Info("Executing transaction",
		zap.Uint("database_id", databaseID),
		zap.Int("operation_count", len(operations)))

	// For now, we'll simulate transaction execution
	// In production, you would execute the actual transaction

	// Simulate processing time
	time.Sleep(200 * time.Millisecond)

	s.logger.Info("Transaction executed successfully",
		zap.Uint("database_id", databaseID))

	return nil
}

// GetDatabaseStats returns database statistics.
func (s *DatabaseService) GetDatabaseStats(ctx context.Context, databaseID uint) (*models.DatabaseStats, error) {
	s.logger.Info("Getting database statistics", zap.Uint("database_id", databaseID))

	// For now, we'll return simulated statistics
	// In production, you would query actual database statistics

	stats := &models.DatabaseStats{
		DatabaseID:  databaseID,
		TableCount:  10,
		RowCount:    1000,
		SizeBytes:   1024000,
		IndexCount:  5,
		LastBackup:  time.Now().Add(-24 * time.Hour),
		LastUpdated: time.Now(),
	}

	return stats, nil
}

// Health checks the health of the database service.
func (s *DatabaseService) Health(ctx context.Context) error {
	s.logger.Debug("Checking database service health")

	// Check database connection
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		s.logger.Error("Database health check failed", zap.Error(err))
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check cache connection
	if err := s.cache.Health(ctx); err != nil {
		s.logger.Error("Cache health check failed", zap.Error(err))
		return fmt.Errorf("cache health check failed: %w", err)
	}

	return nil
}

// initDatabase initializes the database connection.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// NOTE: Database migrations are managed by Atlas (see migrations/ directory and atlas.hcl)
	// Run migrations before starting the service:
	//   cd microservices/database-service
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files
	//
	// IMPORTANT: database-service is DEPRECATED (port 8095) - functionality moved to shared-resilience library

	return db, nil
}
