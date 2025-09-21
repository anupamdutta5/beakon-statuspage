package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/anupamdutta5/saas-admin-service/internal/db/seed"
	"github.com/stretchr/testify/assert"
)

func TestSeedDatabase(t *testing.T) {
	// Set up test database connection
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5433/saas_admin?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Run the seed function
	err = seed.Seed(db)
	assert.NoError(t, err, "Seed function should not return an error")

	// Verify that plans were created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM saas_plans").Scan(&count)
	assert.NoError(t, err, "Should be able to query saas_plans table")
	assert.Greater(t, count, 0, "Should have at least one plan created")

	// Verify that pricing tiers were created
	err = db.QueryRow("SELECT COUNT(*) FROM pricing_tiers").Scan(&count)
	assert.NoError(t, err, "Should be able to query pricing_tiers table")
	assert.Greater(t, count, 0, "Should have at least one pricing tier created")

	// Verify that plan features were created
	err = db.QueryRow("SELECT COUNT(*) FROM plan_features").Scan(&count)
	assert.NoError(t, err, "Should be able to query plan_features table")
	assert.Greater(t, count, 0, "Should have at least one plan feature created")
}
