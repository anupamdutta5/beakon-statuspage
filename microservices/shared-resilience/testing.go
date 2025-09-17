package resilience

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// TestDatabaseManager manages test database connections
type TestDatabaseManager struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	logger *zap.Logger
}

// NewTestDatabaseManager creates a new test database manager with proper cleanup
func NewTestDatabaseManager(t *testing.T) *TestDatabaseManager {
	logger := zaptest.NewLogger(t)

	// Use in-memory SQLite for tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.New(
			&testGormLogger{logger: logger},
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormlogger.Silent, // Reduce noise in tests
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Get underlying sql.DB for cleanup
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Configure connection pool for tests
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	manager := &TestDatabaseManager{
		db:     db,
		sqlDB:  sqlDB,
		logger: logger,
	}

	// Register cleanup function
	t.Cleanup(func() {
		manager.Close()
	})

	return manager
}

// GetDB returns the GORM database instance
func (tdm *TestDatabaseManager) GetDB() *gorm.DB {
	return tdm.db
}

// GetLogger returns the test logger
func (tdm *TestDatabaseManager) GetLogger() *zap.Logger {
	return tdm.logger
}

// AutoMigrate performs migration for test models
func (tdm *TestDatabaseManager) AutoMigrate(models ...interface{}) error {
	return tdm.db.AutoMigrate(models...)
}

// Truncate removes all data from specified tables
func (tdm *TestDatabaseManager) Truncate(tables ...string) error {
	for _, table := range tables {
		if err := tdm.db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}

// Close closes the test database connection
func (tdm *TestDatabaseManager) Close() error {
	if tdm.sqlDB != nil {
		return tdm.sqlDB.Close()
	}
	return nil
}

// TestTransaction executes a function within a test transaction that gets rolled back
func (tdm *TestDatabaseManager) TestTransaction(fn func(*gorm.DB) error) error {
	tx := tdm.db.Begin()
	defer tx.Rollback()

	return fn(tx)
}

// testGormLogger implements GORM logger interface for tests
type testGormLogger struct {
	logger *zap.Logger
}

func (l *testGormLogger) Printf(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

// TestConfig creates a test configuration
func TestConfig() *Config {
	return &Config{
		Environment: "test",

		CircuitBreaker: CircuitBreakerConfigs{
			Database: CircuitBreakerConfig{
				Name:         "test-database",
				MaxRequests:  3,
				Interval:     5 * time.Second,
				Timeout:      10 * time.Second,
				FailureRatio: 0.5,
				Enabled:      true,
			},
			External: CircuitBreakerConfig{
				Name:         "test-external",
				MaxRequests:  2,
				Interval:     5 * time.Second,
				Timeout:      10 * time.Second,
				FailureRatio: 0.5,
				Enabled:      true,
			},
		},

		Cache: CacheConfig{
			Enabled:         true,
			DefaultTTL:      1 * time.Minute,
			MaxSize:         100,
			CleanupInterval: 30 * time.Second,
			Type:            "memory",
		},

		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
			Burst:             10,
			CleanupInterval:   1 * time.Minute,
			RedisKeyPrefix:    "test:rate_limit:",
			ExcludedPaths:     []string{"/health", "/metrics"},
		},

		Security: SecurityConfig{
			SanitizationEnabled:    true,
			MaxStringLength:        1000,
			StrictMode:             false,
			AllowedFileTypes:       []string{"jpg", "png", "pdf"},
			MaxFileSize:            1024 * 1024, // 1MB
			SecurityHeadersEnabled: true,
		},

		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "test",
			Password:        "test",
			Name:            ":memory:",
			SSLMode:         "disable",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
		},

		Redis: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           1, // Use different DB for tests
			MaxRetries:   3,
			PoolSize:     5,
			MinIdleConns: 2,
			KeyPrefix:    "test:",
		},

		JWT: JWTConfig{
			Secret:            "test-jwt-secret-for-testing-only-32-chars",
			Expiration:        1 * time.Hour,
			RefreshExpiration: 24 * time.Hour,
			Issuer:            "test-beakon",
			Audience:          "test-users",
		},

		CORS: CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Correlation-ID"},
			AllowCredentials: false,
			MaxAge:           3600,
			ExposeHeaders:    []string{"X-Correlation-ID"},
		},

		Monitoring: MonitoringConfig{
			Enabled:         false, // Disable monitoring in tests
			MetricsEnabled:  false,
			TracingEnabled:  false,
			MetricsPath:     "/metrics",
			HealthPath:      "/health",
			JaegerEndpoint:  "",
			TracingSampling: 0.0,
		},
	}
}

// TestGinEngine creates a test Gin engine with minimal setup
func TestGinEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Add minimal middleware for testing
	engine.Use(RequestIDMiddleware())

	return engine
}

// TestGinEngineWithMiddleware creates a test Gin engine with full middleware stack
func TestGinEngineWithMiddleware(config *Config, logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Add middleware stack
	middleware := DefaultMiddlewareStack(config, logger)
	for _, mw := range middleware {
		engine.Use(mw)
	}

	return engine
}

// TestModel represents a sample model for testing
type TestModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Email       string `gorm:"size:255;unique;not null" json:"email"`
	Description string `gorm:"type:text" json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// TenantTestModel represents a tenant-aware test model
type TenantTestModel struct {
	TenantAwareModel
	Name     string `gorm:"size:255;not null" json:"name"`
	Value    int    `json:"value"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

// CreateTestData creates sample test data
func CreateTestData(db *gorm.DB) error {
	// Migrate test models
	if err := db.AutoMigrate(&TestModel{}, &TenantTestModel{}); err != nil {
		return fmt.Errorf("failed to migrate test models: %w", err)
	}

	// Create test data
	testModels := []TestModel{
		{Name: "Test Item 1", Email: "test1@example.com", Description: "First test item", IsActive: true},
		{Name: "Test Item 2", Email: "test2@example.com", Description: "Second test item", IsActive: true},
		{Name: "Test Item 3", Email: "test3@example.com", Description: "Third test item", IsActive: false},
	}

	for _, model := range testModels {
		if err := db.Create(&model).Error; err != nil {
			return fmt.Errorf("failed to create test model: %w", err)
		}
	}

	tenantTestModels := []TenantTestModel{
		{TenantAwareModel: TenantAwareModel{TenantID: "tenant1"}, Name: "Tenant 1 Item 1", Value: 100, IsActive: true},
		{TenantAwareModel: TenantAwareModel{TenantID: "tenant1"}, Name: "Tenant 1 Item 2", Value: 200, IsActive: true},
		{TenantAwareModel: TenantAwareModel{TenantID: "tenant2"}, Name: "Tenant 2 Item 1", Value: 150, IsActive: true},
	}

	for _, model := range tenantTestModels {
		if err := db.Create(&model).Error; err != nil {
			return fmt.Errorf("failed to create tenant test model: %w", err)
		}
	}

	return nil
}

// AssertDatabaseHealth checks if database is healthy
func AssertDatabaseHealth(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("Database health check failed: %v", err)
	}
}

// TestEnvironmentSetup sets up test environment variables
func TestEnvironmentSetup() func() {
	// Store original values
	originalValues := make(map[string]string)
	testEnvVars := map[string]string{
		"ENVIRONMENT":            "test",
		"LOG_LEVEL":              "error",
		"DB_HOST":                "localhost",
		"DB_PORT":                "5432",
		"DB_USER":                "test",
		"DB_PASSWORD":            "test",
		"DB_NAME":                "test_db",
		"REDIS_HOST":             "localhost",
		"REDIS_PORT":             "6379",
		"REDIS_DB":               "1",
		"JWT_SECRET":             "test-jwt-secret-for-testing-only-32-chars",
		"CACHE_ENABLED":          "true",
		"CACHE_TYPE":             "memory",
		"RATE_LIMIT_ENABLED":     "true",
		"MONITORING_ENABLED":     "false",
		"METRICS_ENABLED":        "false",
		"TRACING_ENABLED":        "false",
	}

	// Set test environment variables
	for key, value := range testEnvVars {
		originalValues[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key, originalValue := range originalValues {
			if originalValue != "" {
				os.Setenv(key, originalValue)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

// BenchmarkHelper provides utilities for benchmark tests
type BenchmarkHelper struct {
	config *Config
	logger *zap.Logger
	db     *gorm.DB
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	logger := zap.NewNop() // No-op logger for benchmarks
	config := TestConfig()

	// Use in-memory SQLite for benchmarks
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		b.Fatalf("Failed to create benchmark database: %v", err)
	}

	helper := &BenchmarkHelper{
		config: config,
		logger: logger,
		db:     db,
	}

	// Setup cleanup
	b.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

	return helper
}

// GetConfig returns the test configuration
func (bh *BenchmarkHelper) GetConfig() *Config {
	return bh.config
}

// GetLogger returns the benchmark logger
func (bh *BenchmarkHelper) GetLogger() *zap.Logger {
	return bh.logger
}

// GetDB returns the benchmark database
func (bh *BenchmarkHelper) GetDB() *gorm.DB {
	return bh.db
}

// MockErrorHandler creates a mock error handler for testing
type MockErrorHandler struct {
	Errors []error
}

// Handle captures errors for testing
func (meh *MockErrorHandler) Handle(c *gin.Context, err error) {
	meh.Errors = append(meh.Errors, err)
}

// GetLastError returns the last captured error
func (meh *MockErrorHandler) GetLastError() error {
	if len(meh.Errors) == 0 {
		return nil
	}
	return meh.Errors[len(meh.Errors)-1]
}

// Reset clears captured errors
func (meh *MockErrorHandler) Reset() {
	meh.Errors = nil
}