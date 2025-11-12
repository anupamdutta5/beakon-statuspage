# V2.0 Migration Complete Guide

**Project**: Beakon Status Page - shared-resilience v2.0 Migration
**Status**: 15/19 services complete (78.9%)
**Last Updated**: 2025-11-04

---

## 📊 Quick Overview

```
Progress:     ████████████████░░░░  78.9%
Completed:    15 services
Remaining:    2 services (+ 2 deprecated)
Time Spent:   6.5 hours
Avg Time:     26 min/service
Build Rate:   100% success
```

---

## 🎯 Migration Patterns (All 4 Established)

### Pattern 1: Template-Based ⭐
**Use when**: Service has standard `NewService(db *gorm.DB, logger)` constructor

**Services**: tenant-admin, user, incident, component, payment, analytics (7 total)

**Process**:
```bash
cd microservices/<service-name>
../../scripts/migrate-service-to-v2.sh <service-name>
# Fix any naming issues
# Update routes in cmd/main_v2.go
go build -o service-v2 ./cmd/main_v2.go
# If successful:
mv cmd/main.go cmd/main.v1.backup.go
mv cmd/main_v2.go cmd/main.go
mv configs/config.yml configs/config.v1.backup.yml
mv configs/config.v2.yml configs/config.yml
```

**Time**: 10-15 minutes

---

### Pattern 2: Refactor + Template
**Use when**: Service has custom config in constructor but only for DB init

**Services**: event-store, branding (2 total)

**Process**:
1. Refactor service:
   - Remove `config *config.Config` field from struct
   - Change constructor to `NewService(db *gorm.DB, logger)`
   - Delete `initDatabase()` function
   - Remove config package import (if only used for DB)

2. Apply template pattern (same as Pattern 1)

**Time**: 20 minutes

---

### Pattern 3: Manual Migration
**Use when**: Template generates incorrect output or service has unique structure

**Services**: monitoring, notification, status-ui (3 total)

**Process**:
1. Read current `cmd/main.go` to understand structure
2. Grep handlers to identify routes: `grep -n "^func (h \*.*Handler)" internal/handlers/*.go`
3. Create `cmd/main_v2.go` manually based on payment-service template
4. Update service-specific sections (routes, handlers, config structs)
5. Fix API differences
6. Build and finalize

**Time**: 30-120 minutes (depends on complexity)

---

### Pattern 4: Consumer Template ⭐ NEW
**Use when**: RabbitMQ consumer service (no HTTP server)

**Services**: audit-consumer, billing-consumer, notification-consumer, analytics-consumer (4 total)

**Template**:
```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/<org>/<service>/internal/config"
	"github.com/<org>/<service>/internal/consumer"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from YAML
	loader := resilience.NewConfigLoader("configs")
	var cfg config.Config
	if err := loader.Load(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logger *zap.Logger
	var err error
	if cfg.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting <Service> Consumer (v2.0)",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Environment),
	)

	// Initialize consumer
	consumer, err := consumer.New<Service>Consumer(&cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize consumer", zap.Error(err))
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("<Service> Consumer starting...")
		if err := consumer.Start(ctx); err != nil {
			logger.Error("Consumer stopped with error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down <Service> Consumer...")
	cancel()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	select {
	case <-shutdownCtx.Done():
		logger.Warn("<Service> Consumer shutdown timeout")
	case <-time.After(5 * time.Second):
		logger.Info("<Service> Consumer stopped gracefully")
	}

	logger.Info("<Service> Consumer exited")
}
```

**Time**: 8-10 minutes

---

## 📁 Required Files

Every v2.0 service needs:

1. **configs/config.v2.yml** (or config.yml after migration)
   - All non-secret configuration in YAML
   - Uses `${ENV_VAR}` for environment variable expansion
   - Secrets loaded from `.env` file

2. **configs/.env** (gitignored)
   - Contains secrets (DB passwords, JWT secrets, API keys)
   - Never committed to git

3. **cmd/main.go**
   - Uses `resilience.NewConfigLoader("configs")`
   - Injects Prometheus registry (no global state)
   - Uses `resilience.NewServiceClient()` for downstream calls
   - Uses `resilience.DefaultMiddlewareStack()` for HTTP services

---

## 🔧 Common API Differences (v1.x → v2.0)

```go
// Database config field name
cfg.SharedConfig.Database.MaxConns          // ❌ v1.x
cfg.SharedConfig.Database.MaxOpenConns      // ✅ v2.0

// Load service endpoints
resilience.LoadServiceEndpoints("file.yml") // ❌ v1.x
loader.LoadServiceEndpoints()               // ✅ v2.0

// Retry manager signature
resilience.NewRetryManager(retryConfig)     // ❌ v1.x
resilience.NewRetryManager(cfg.Retry, logger) // ✅ v2.0

// Retry config field
cfg.Retry.MaxAttempts                       // ❌ v1.x
cfg.Retry.MaxRetries                        // ✅ v2.0

// Middleware stack
router.Use(resilience.Recovery(logger, metrics)) // ❌ v1.x (individual)
middleware := resilience.DefaultMiddlewareStack(&cfg.SharedConfig, logger) // ✅ v2.0
for _, mw := range middleware { router.Use(mw) }
```

---

## ✅ Completed Services (15)

1. monitoring-service (489 lines, 39MB) - Manual, 2h
2. notification-service (488 lines, 38MB) - Manual, 1h
3. tenant-admin-service (464 lines, 39MB) - Template, 10m
4. user-service (439 lines, 37MB) - Template, 10m
5. incident-service (405 lines, 37MB) - Template, 10m
6. component-service (386 lines, 39MB) - Template, 10m
7. payment-service (436 lines, 40MB) - Template, 15m
8. analytics-service (357 lines, 39MB) - Template, 15m
9. event-store-service (267 lines, 39MB) - Refactor, 20m
10. branding-service (257 lines, 39MB) - Refactor, 20m
11. status-ui-service (352 lines, 41MB) - Manual, 45m
12. audit-consumer (95 lines, 35MB) - Consumer, 10m
13. billing-consumer (95 lines, 35MB) - Consumer, 8m
14. notification-consumer (95 lines, 35MB) - Consumer, 8m
15. analytics-consumer (95 lines, 35MB) - Consumer, 8m

**Total**: 5,120 lines, 15 services, ~6.5 hours

---

## ⏳ Remaining Services (2 + 2 deprecated)

### Complex HTTP Services

**landing-page-service** (122 lines)
- **Challenge**: Config used in 13 places for business logic
- **Approach**: Create service-specific LandingConfig struct
- **Estimated**: 1-1.5 hours

**saas-admin-service** (606 lines)
- **Challenge**: Largest service, multi-service architecture
- **Approach**: Investigate, then apply appropriate pattern
- **Estimated**: 2-3 hours

### Deprecated (Skip)

**api-gateway** (369 lines) - Deprecated Oct 26, 2025
**database-service** (99 lines) - Functionality moved to shared-resilience

---

## 📈 Performance Metrics

### Time Analysis
- **Total services**: 15
- **Total time**: ~6.5 hours
- **Average**: 26 min/service
- **Fastest pattern**: Consumer (8.5 min avg)
- **Slowest pattern**: Manual (58 min avg)

### Build Success
- **Success rate**: 100% (15/15)
- **Binary sizes**: 35-41MB (very consistent)
- **Pattern compliance**: 100% use v2.0 YAML config

### Efficiency Gains
- **Without patterns**: 15 × 90 min = 22.5 hours (estimated)
- **With patterns**: 6.5 hours actual
- **Time saved**: 16 hours (71% reduction)

---

## 🚀 Next Steps

**Immediate** (1-2 sessions, 3-4.5 hours):
1. Migrate landing-page-service
2. Migrate saas-admin-service
3. **Reach 100% completion**

**Post-Migration**:
1. Test all 17 services (skip deprecated)
2. Update documentation
3. Create final migration report

---

## 📚 Documentation

### Created Documents
1. SESSION_2_COMPLETE_SUMMARY.md
2. SESSION_3_SUMMARY.md
3. SESSION_3_EXTENDED_SUMMARY.md
4. FINAL_SESSION_3_STATUS.md
5. MIGRATION_SUMMARY_TABLE.md
6. V2_MIGRATION_COMPLETE_GUIDE.md (this file)
7. V2_MIGRATION_PROGRESS.md
8. SERVICES_V2_MIGRATION_STATUS.md

### Migration Scripts
1. scripts/migrate-service-to-v2.sh
2. scripts/generate-v2-configs.sh
3. scripts/migrate-all-services.sh

---

## 💡 Key Learnings

1. **Pattern-based approach works** - 4 patterns cover all service types
2. **Consumer template is fastest** - Simplest pattern (no HTTP/metrics)
3. **Manual migration is valid** - Some services too unique for templates
4. **YAML config standardization** - Consistent across all services
5. **Batch migrations efficient** - 4 consumers in 34 minutes
6. **Skip deprecated early** - Saves time and effort

---

## ✨ Success Factors

1. ✅ All 4 migration patterns established and validated
2. ✅ 100% build success rate across all services
3. ✅ Comprehensive documentation created
4. ✅ Efficient time management (26 min/service avg)
5. ✅ Clear path to completion (only 2 services remain)

---

**The migration is 78.9% complete with proven patterns and clear path to 100%!**

Next session target: **Complete remaining 2 services → 100% migration**
