# Configuration Loading Strategy

## Overview

As of November 2025, the Beakon platform uses a unified configuration loading strategy that works seamlessly across all environments (Docker, local development, and tests) without requiring manual path adjustments.

## How It Works

The `shared-resilience` library's `ConfigLoader` automatically determines the correct configuration path based on the runtime context. Services simply call:

```go
loader := resilience.NewConfigLoader("configs")
```

The library then intelligently searches for configuration files in the following priority order:

### 1. CONFIG_PATH Environment Variable (Highest Priority)
If set, this overrides all other path resolution:
```bash
export CONFIG_PATH=/custom/path/to/configs
```

### 2. Provided Path Check
If the provided path (e.g., `"configs"`) contains a valid `config.yml`, it's used directly. This handles:
- **Docker containers**: Where configs are mounted at `/app/configs`
- **Any valid local path**: If configs exist at the specified location

### 3. Automatic Service Discovery
If the provided path doesn't exist, the library:
1. Determines the current service name from the working directory
2. Searches for configs in multiple locations:
   - `../../docker-deployment/<service-name>/configs` (for local development)
   - `../../../../docker-deployment/<service-name>/configs` (for tests in subdirectories)
   - `docker-deployment/<service-name>/configs` (from project root)

### 4. Fallback
If no configuration is found, the original provided path is used (which will likely fail with a clear error message).

## Environment-Specific Behavior

### Docker (Production/Staging)
```yaml
# docker-compose.yml
volumes:
  - ./configs:/app/configs:ro
```
- Configs are mounted at `/app/configs`
- The library finds them at the provided path `"configs"`
- No special configuration needed

### Local Development
```bash
cd microservices/tenant-admin-service
go run cmd/main.go
```
- Working directory: `microservices/tenant-admin-service`
- Library automatically finds: `../../docker-deployment/tenant-admin-service/configs`
- No manual path configuration needed

### Tests
```bash
cd microservices/tenant-admin-service
go test ./tests/unit
```
- Working directory: `microservices/tenant-admin-service`
- Library automatically finds: `../../docker-deployment/tenant-admin-service/configs`
- Tests use the same `config.Load()` as production code

## Service Implementation

### In main.go
```go
// Standard configuration loading - works in all environments
cfg, err := config.Load()
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### In config/config.go
```go
func Load() (*Config, error) {
    // Create configuration loader pointing to "configs" directory
    // The library will resolve this to the correct path automatically
    loader := resilience.NewConfigLoader("configs")

    var cfg Config
    if err := loader.Load(&cfg); err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return &cfg, nil
}
```

### In Tests
```go
func loadTestConfig(t *testing.T) *config.Config {
    // Use development environment for tests
    os.Setenv("ENVIRONMENT", "development")
    defer os.Unsetenv("ENVIRONMENT")

    // Use the same config.Load() as production
    cfg, err := config.Load()
    if err != nil {
        t.Fatalf("Failed to load test configuration: %v", err)
    }

    // Override database for in-memory testing
    cfg.Database.Host = "localhost"
    cfg.Database.Name = ":memory:"

    return cfg
}
```

## Benefits

1. **Zero Configuration**: No manual path adjustments needed for different environments
2. **Consistent API**: Same code works everywhere - `NewConfigLoader("configs")`
3. **Test Simplicity**: Tests use the same configuration loading as production
4. **Maintainability**: No hardcoded paths, no environment-specific code
5. **Flexibility**: Can override with `CONFIG_PATH` when needed

## Migration Guide

If you have existing services with hardcoded paths:

### Before (Hardcoded):
```go
// In tests
configPath := "../../docker-deployment/service-name/configs"
loader := resilience.NewConfigLoader(configPath)

// In main.go
loader := resilience.NewConfigLoader("/app/configs")
```

### After (Automatic):
```go
// Everywhere - tests, main.go, local dev
loader := resilience.NewConfigLoader("configs")
// Or use the service's config.Load() function
cfg, err := config.Load()
```

## Troubleshooting

### Config Not Found
If you see "failed to load config.yml", check:
1. Current working directory: `pwd`
2. Service name detection: Ensure you're in a standard service directory
3. Config file exists: `ls ../../docker-deployment/<service>/configs/config.yml`

### Override Path
If automatic detection doesn't work for your use case:
```bash
export CONFIG_PATH=/path/to/your/configs
go run cmd/main.go
```

### Debug Path Resolution
The library will use the first valid path it finds. To debug:
```bash
# Check what path would be used
find . -name config.yml 2>/dev/null
ls docker-deployment/*/configs/config.yml
```

## Summary

The unified configuration loading strategy eliminates the complexity of managing different configuration paths for different environments. Services can use a simple, consistent API (`"configs"`) that automatically resolves to the correct location whether running in Docker, locally, or in tests. This approach follows the principle of "convention over configuration" while still allowing flexibility through the `CONFIG_PATH` environment variable when needed.