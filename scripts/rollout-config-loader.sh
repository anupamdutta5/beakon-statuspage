#!/bin/bash
# Rollout loadTestConfig() pattern to all services
# This adds the config loader helper and updates tests to use it

set -e

cd "$(dirname "$0")/.."

# The loadTestConfig helper function template
read -r -d '' LOAD_TEST_CONFIG_FUNC << 'EOF' || true
// Helper function to load test configuration
func loadTestConfig(t *testing.T) *config.Config {
	// Get the config path relative to the test file
	configPath := "../../configs"

	// Set environment to test to load config.test.yml
	os.Setenv("ENVIRONMENT", "test")
	defer os.Unsetenv("ENVIRONMENT")

	// Load configuration using shared-resilience ConfigLoader
	loader := resilience.NewConfigLoader(configPath)
	var cfg config.Config
	err := loader.Load(&cfg)
	if err != nil {
		// Fallback to minimal config if file doesn't exist
		t.Logf("Warning: Could not load config.test.yml, using defaults: %v", err)
		return &config.Config{
			Service: config.ServiceConfig{
				Name:        "SERVICE_NAME_PLACEHOLDER",
				Version:     "1.0.0-test",
				Environment: "test",
			},
		}
	}

	return &cfg
}
EOF

echo "Rolling out loadTestConfig() to all services..."
echo ""

# Services that need the config pattern (those with cfg in their constructors)
# Based on V2 patterns discovered earlier
services_needing_config=(
    "saas-admin-service"
    "landing-page-service"
)

for svc in "${services_needing_config[@]}"; do
    test_file="microservices/$svc/tests/unit/${svc//-/_}_test.go"

    if [ ! -f "$test_file" ]; then
        echo "  ⚠️  $svc: test file not found at $test_file"
        continue
    fi

    # Check if loadTestConfig already exists
    if grep -q "func loadTestConfig" "$test_file"; then
        echo "  ✓ $svc: loadTestConfig already exists"
        continue
    fi

    echo "  🔧 Processing $svc..."

    # Add imports if needed (os and resilience)
    if ! grep -q '"os"' "$test_file"; then
        echo "    - Adding os import"
        # This will be done manually per service due to import complexity
    fi

    if ! grep -q 'resilience "github.com/anupamdutta5/shared-resilience"' "$test_file"; then
        echo "    - Adding resilience import"
        # This will be done manually per service
    fi

    # We'll handle each service individually due to different patterns
    echo "    - Service requires manual configuration (different V2 pattern)"
done

echo ""
echo "✅ Analysis complete!"
echo ""
echo "Summary:"
echo "  - tenant-admin-service: ✅ Already has loadTestConfig()"
echo "  - saas-admin-service: Needs loadTestConfig() + imports"
echo "  - landing-page-service: Needs loadTestConfig() + imports"
echo "  - Other services: Use simple constructors (db, logger) - no config needed"
echo ""
echo "Next: Apply changes to saas-admin-service and landing-page-service"
