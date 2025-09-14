// Package unit provides unit tests for the Database Service.
package unit

import (
	"context"
	"testing"

	"github.com/enterprise-status/statuspage-database-service/internal/config"
	"github.com/enterprise-status/statuspage-database-service/internal/models"
	"github.com/enterprise-status/statuspage-database-service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseService_ListDatabases(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	databaseService, _ := services.NewDatabaseService(cfg, logger)
	// Override the database connection with our test database
	databaseService.SetDB(db)

	// Test
	databases, err := databaseService.ListDatabases(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, databases)
	assert.IsType(t, []*models.Database{}, databases)
}

func TestDatabaseService_CreateDatabase(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	databaseService, _ := services.NewDatabaseService(cfg, logger)
	// Override the database connection with our test database
	databaseService.SetDB(db)

	// Test data
	database := &models.Database{
		Name:     "test-db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 1,
		Status:   "active",
		Metadata: `{"source": "test"}`,
	}

	// Test
	err := databaseService.CreateDatabase(context.Background(), database)

	// Assertions
	assert.NoError(t, err)
	assert.NotZero(t, database.ID)
}

func TestDatabaseService_GetDatabase(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	databaseService, _ := services.NewDatabaseService(cfg, logger)
	// Override the database connection with our test database
	databaseService.SetDB(db)

	// Create a database first
	database := &models.Database{
		Name:     "test-db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 1,
		Status:   "active",
		Metadata: `{"source": "test"}`,
	}
	err := databaseService.CreateDatabase(context.Background(), database)
	require.NoError(t, err)

	// Test
	retrievedDatabase, err := databaseService.GetDatabase(context.Background(), database.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, retrievedDatabase)
	assert.Equal(t, database.Name, retrievedDatabase.Name)
	assert.Equal(t, database.Type, retrievedDatabase.Type)
	assert.Equal(t, database.Host, retrievedDatabase.Host)
	assert.Equal(t, database.Port, retrievedDatabase.Port)
}

func TestDatabaseService_UpdateDatabase(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	databaseService, _ := services.NewDatabaseService(cfg, nil)

	// Create a database first
	database := &models.Database{
		Name:     "test-db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 1,
		Status:   "active",
		Metadata: `{"source": "test"}`,
	}
	err := databaseService.CreateDatabase(context.Background(), database)
	require.NoError(t, err)

	// Update data
	database.Name = "updated-db"
	database.Port = 5433
	database.MaxConns = 20

	// Test
	err = databaseService.UpdateDatabase(context.Background(), database.ID, database)

	// Assertions
	assert.NoError(t, err)

	// Verify update
	updatedDatabase, err := databaseService.GetDatabase(context.Background(), database.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-db", updatedDatabase.Name)
	assert.Equal(t, 5433, updatedDatabase.Port)
	assert.Equal(t, 20, updatedDatabase.MaxConns)
}

func TestDatabaseService_DeleteDatabase(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	databaseService, _ := services.NewDatabaseService(cfg, nil)

	// Create a database first
	database := &models.Database{
		Name:     "test-db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 1,
		Status:   "active",
		Metadata: `{"source": "test"}`,
	}
	err := databaseService.CreateDatabase(context.Background(), database)
	require.NoError(t, err)

	// Test
	err = databaseService.DeleteDatabase(context.Background(), database.ID)

	// Assertions
	assert.NoError(t, err)

	// Verify deletion
	_, err = databaseService.GetDatabase(context.Background(), database.ID)
	assert.Error(t, err)
}

func TestDatabaseService_ListTables(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	databaseService, _ := services.NewDatabaseService(cfg, nil)

	// Create a database first
	database := &models.Database{
		Name:     "test-db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 1,
		Status:   "active",
		Metadata: `{"source": "test"}`,
	}
	err := databaseService.CreateDatabase(context.Background(), database)
	require.NoError(t, err)

	// Test
	tables, err := databaseService.ListTables(context.Background(), database.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, tables)
	assert.IsType(t, []*models.Table{}, tables)
}

func TestDatabaseService_ExecuteQuery(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	databaseService, _ := services.NewDatabaseService(cfg, nil)

	// Create a database first
	database := &models.Database{
		Name:     "test-db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MinConns: 1,
		Status:   "active",
		Metadata: `{"source": "test"}`,
	}
	err := databaseService.CreateDatabase(context.Background(), database)
	require.NoError(t, err)

	// Test
	query := "SELECT 1 as test_column"
	result, err := databaseService.ExecuteQuery(context.Background(), database.ID, query)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "rows")
	assert.Contains(t, result, "columns")
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Database{},
		&models.Table{},
		&models.Column{},
		&models.Index{},
		&models.Backup{},
		&models.Migration{},
	)
	require.NoError(t, err)

	return db
}
