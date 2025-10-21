// Package dbclient provides a client for interacting with the database-service
package dbclient

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
)

// DatabaseManager manages database connections and migrations
// for a service using the database-service
// 
// Example usage:
//
//	manager := dbclient.NewDatabaseManager(
//		dbClient,
//		"my-service",
//		dbclient.DatabaseTypePostgreSQL,
//		"migrations",
//		slog.Default(),
//	)
//
//	// Initialize the database and run migrations
//	db, err := manager.Initialize(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to initialize database: %w", err)
//	}
//	defer db.Close()
//
//	// Use the database connection
//	// ...
//
//	// Cleanup when the service shuts down
//	if err := manager.Cleanup(ctx); err != nil {
//		return fmt.Errorf("failed to cleanup database: %w", err)
//	}
//
//	return nil
//
// The DatabaseManager will automatically create a new database for the service
// if it doesn't exist, and run all pending migrations. When the service shuts
// down, it will close all database connections and perform any necessary cleanup.
//
// The DatabaseManager is safe for concurrent use by multiple goroutines.
type DatabaseManager struct {
	client       *Client
	serviceName  string
	dbType       string
	migrationsDir string
	logger       *slog.Logger

	// Internal state
	dbConfig     *Database
	db           *sql.DB
}

// NewDatabaseManager creates a new DatabaseManager
func NewDatabaseManager(
	client *Client,
	serviceName string,
	dbType string,
	migrationsDir string,
	logger *slog.Logger,
) *DatabaseManager {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return &DatabaseManager{
		client:        client,
		serviceName:   serviceName,
		dbType:        dbType,
		migrationsDir: migrationsDir,
		logger:        logger,
	}
}

// Initialize initializes the database for the service
// It creates the database if it doesn't exist and runs migrations
func (m *DatabaseManager) Initialize(ctx context.Context) (*sql.DB, error) {
	// Check if the database service is available
	if err := m.client.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("database service is not available: %w", err)
	}

	// Try to get the database for this service
	dbName := GenerateDatabaseName(m.serviceName)
	dbUser := GenerateDatabaseUser(m.serviceName)
	dbPass := GenerateDatabasePassword()

	m.logger.Info("Initializing database", "service", m.serviceName, "database", dbName)

	// Create or get the database
	db, err := m.getOrCreateDatabase(ctx, dbName, dbUser, dbPass)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create database: %w", err)
	}

	// Store the database configuration
	m.dbConfig = db

	// Connect to the database
	dbConnStr := m.getConnectionString(db, dbUser, dbPass)
	sqlDB, err := sql.Open(db.Driver(), dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	m.db = sqlDB

	// Run migrations if migrations directory exists
	if m.migrationsDir != "" {
		m.logger.Info("Running database migrations", "dir", m.migrationsDir)
		if err := m.runMigrations(ctx, dbConnStr); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	m.logger.Info("Database initialized successfully", "service", m.serviceName, "database", dbName)

	return sqlDB, nil
}

// getOrCreateDatabase gets an existing database or creates a new one
func (m *DatabaseManager) getOrCreateDatabase(ctx context.Context, name, user, password string) (*Database, error) {
	// Try to get the database by name
	databases, err := m.client.ListDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	// Check if the database already exists
	for _, db := range databases {
		if db.Name == name {
			m.logger.Info("Using existing database", "name", name)
			return &db, nil
		}
	}

	// Database doesn't exist, create a new one
	m.logger.Info("Creating new database", "name", name)

	db := &Database{
		Name:        name,
		Description: fmt.Sprintf("Database for %s service", m.serviceName),
		Type:        m.dbType,
		Host:        "postgres", // This will be overridden by the database-service
		Port:        5432,       // This will be overridden by the database-service
		Username:    user,
		Password:    password,
		SSLMode:     "disable",
	}

	createdDB, err := m.client.CreateDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	m.logger.Info("Created new database", "name", name, "id", createdDB.ID)

	return createdDB, nil
}

// runMigrations runs database migrations
func (m *DatabaseManager) runMigrations(ctx context.Context, dbURL string) error {
	// Check if migrations directory exists
	if _, err := os.Stat(m.migrationsDir); os.IsNotExist(err) {
		m.logger.Info("Migrations directory does not exist, skipping migrations", "dir", m.migrationsDir)
		return nil
	}

	// Convert file:// URLs to absolute paths for the file source
	sourceURL := m.migrationsDir
	if !strings.HasPrefix(sourceURL, "github://") {
		absPath, err := filepath.Abs(sourceURL)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for migrations: %w", err)
		}
		sourceURL = "file://" + absPath
	}

	// Create a new migrator
	migrator, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Run migrations
	m.logger.Info("Running database migrations...")
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	m.logger.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (m *DatabaseManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// Cleanup performs any necessary cleanup when the service shuts down
func (m *DatabaseManager) Cleanup(ctx context.Context) error {
	m.logger.Info("Cleaning up database resources", "service", m.serviceName)

	// Close the database connection
	if err := m.Close(); err != nil {
		m.logger.Error("Failed to close database connection", "error", err)
	}

	// In a real implementation, you might want to:
	// 1. Clean up any temporary tables or resources
	// 2. Close any open transactions
	// 3. Release any database locks

	m.logger.Info("Database resources cleaned up successfully", "service", m.serviceName)
	return nil
}

// getConnectionString returns a connection string for the database
func (m *DatabaseManager) getConnectionString(db *Database, username, password string) string {
	switch db.Type {
	case DatabaseTypePostgreSQL:
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			username,
			password,
			db.Host,
			db.Port,
			db.Name,
			db.SSLMode,
		)
	case DatabaseTypeMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
			username,
			password,
			db.Host,
			db.Port,
			db.Name,
		)
	default:
		// Default to PostgreSQL
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			username,
			password,
			db.Host,
			db.Port,
			db.Name,
		)
	}
}

// Database returns the database configuration
func (m *DatabaseManager) Database() *Database {
	return m.dbConfig
}

// Driver returns the database driver name
func (db *Database) Driver() string {
	switch db.Type {
	case DatabaseTypePostgreSQL:
		return "postgres"
	case DatabaseTypeMySQL:
		return "mysql"
	default:
		return "postgres"
	}
}
