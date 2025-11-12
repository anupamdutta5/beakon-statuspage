# Unified Configuration Loading - Same Code, All Environments

## The Same Code Works Everywhere

Every service uses the exact same configuration loading code, regardless of the environment:

```go
// In config/config.go - SAME for production, dev, and tests
func Load() (*Config, error) {
    // Always use "configs" - the library figures out the actual path
    loader := resilience.NewConfigLoader("configs")

    var cfg Config
    if err := loader.Load(&cfg); err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }

    return &cfg, nil
}
```

```go
// In cmd/main.go - SAME for production and dev
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    // ... rest of service initialization
}
```

```go
// In tests - SAME pattern, just with test overrides
func loadTestConfig(t *testing.T) *config.Config {
    os.Setenv("ENVIRONMENT", "development")

    cfg, err := config.Load()  // Same Load() function!
    if err != nil {
        t.Fatalf("Failed to load test configuration: %v", err)
    }

    // Test-specific overrides
    cfg.Database.Name = ":memory:"
    return cfg
}
```

## How It Works in Each Environment

### 1. Production (Docker Container)

**Setup:**
```yaml
# docker-compose.yml
services:
  tenant-admin-service:
    volumes:
      - ./configs:/app/configs:ro  # Configs mounted at /app/configs
```

**What happens:**
```
1. Service calls: NewConfigLoader("configs")
2. Library checks: Does "configs/config.yml" exist?
3. In Docker: YES! (mounted at /app/configs)
4. Result: Uses /app/configs
```

**Verification:**
```bash
docker exec tenant-admin-service ls /app/configs
# Output: config.yml  .env  service-endpoints.yml
```

### 2. Local Development

**Setup:**
```
Directory structure:
/Users/.../Beakon/
├── microservices/
│   └── tenant-admin-service/     <-- You run from here
└── docker-deployment/
    └── tenant-admin-service/
        └── configs/              <-- Configs are here
            ├── config.yml
            └── .env
```

**What happens:**
```
1. Service calls: NewConfigLoader("configs")
2. Library checks: Does "configs/config.yml" exist?
3. Locally: NO (no local configs directory)
4. Library extracts service name: "tenant-admin-service"
5. Library searches and finds: ../../docker-deployment/tenant-admin-service/configs
6. Result: Uses docker-deployment configs
```

**Verification:**
```bash
cd microservices/tenant-admin-service
ENVIRONMENT=development go run cmd/main.go
# Logs: "Starting Tenant Admin Service (v2.0)" with port 8099 from config
```

### 3. Tests

**Setup:**
```
Running tests from:
/Users/.../Beakon/microservices/tenant-admin-service/
```

**What happens:**
```
1. Test calls: config.Load() → NewConfigLoader("configs")
2. Library checks: Does "configs/config.yml" exist?
3. In tests: NO
4. Library extracts service name: "tenant-admin-service"
5. Library searches and finds: ../../docker-deployment/tenant-admin-service/configs
6. Result: Uses docker-deployment configs (same as local dev!)
```

**Verification:**
```bash
cd microservices/tenant-admin-service
go test ./tests/unit -run TestTenantAdminHandler_HealthCheck
# PASS - configs loaded automatically
```

## The Magic: Automatic Path Resolution

The `resolveConfigPath` function in shared-resilience handles all the intelligence:

```go
func resolveConfigPath(configPath string) string {
    // 1. Override via environment variable (optional)
    if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
        return envPath
    }

    // 2. Use provided path if it exists (Docker case)
    if _, err := os.Stat(filepath.Join(configPath, "config.yml")); err == nil {
        return configPath  // Found at "configs" - use it!
    }

    // 3. Auto-discover for local dev and tests
    serviceName := extractServiceName(os.Getwd())
    // Search in docker-deployment/<service>/configs
    // ... (finds the right path)

    return foundPath
}
```

## Benefits of This Approach

1. **One Codebase**: No `if docker { ... } else if local { ... }` conditionals
2. **Zero Configuration**: Developers don't need to set CONFIG_PATH
3. **Convention Over Configuration**: Follow the standard directory structure, everything works
4. **Flexibility**: Can still override with CONFIG_PATH if needed
5. **Consistency**: Same `config.Load()` everywhere - production, dev, tests

## Real-World Usage

### Starting Services

**Production:**
```bash
docker-compose up tenant-admin-service
# Automatically uses /app/configs (mounted volume)
```

**Local Development:**
```bash
cd microservices/tenant-admin-service
go run cmd/main.go
# Automatically finds ../../docker-deployment/tenant-admin-service/configs
```

**Tests:**
```bash
cd microservices/tenant-admin-service
go test ./...
# Automatically finds ../../docker-deployment/tenant-admin-service/configs
```

### No More Path Issues

**Before (Error-Prone):**
```go
// Different paths for different environments
if os.Getenv("DOCKER") == "true" {
    loader := NewConfigLoader("/app/configs")
} else if testing {
    loader := NewConfigLoader("../../docker-deployment/service/configs")
} else {
    loader := NewConfigLoader("./configs")
}
```

**Now (Unified):**
```go
// Same code everywhere!
loader := NewConfigLoader("configs")
```

## Override When Needed

If you need a custom config location:
```bash
CONFIG_PATH=/my/custom/configs go run cmd/main.go
```

## Summary

The unified configuration loading strategy means:
- **Production**: Reads from mounted `/app/configs`
- **Development**: Auto-finds `docker-deployment/<service>/configs`
- **Tests**: Auto-finds `docker-deployment/<service>/configs`

All using the **exact same code**: `NewConfigLoader("configs")`

This is what you asked for: "It's not just tests, for production and dev as well" - the same pattern now works everywhere!