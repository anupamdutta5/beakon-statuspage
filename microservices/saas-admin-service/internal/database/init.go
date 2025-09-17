// Package database provides database initialization and management
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/config"
)

// InitDatabase initializes the database connection using the provided configuration
func InitDatabase(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	// Use direct database connection
	return initDirectDatabase(cfg, logger)
}

// initDirectDatabase initializes a direct database connection
func initDirectDatabase(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("Initializing direct database connection",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name))

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdle)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetime) * time.Second)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Migration support disabled - no migration package available
	logger.Info("Migration support disabled - no migration package available")

	logger.Info("Successfully connected to database")
	return db, nil
}


// CloseDatabase closes the database connection
func CloseDatabase(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
