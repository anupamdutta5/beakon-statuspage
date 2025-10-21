package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

//go:embed *.sql
var migrationsFS embed.FS

// Migration represents a database migration.
type Migration struct {
	Name string
	SQL  string
}

// RunMigrations runs all migrations that haven't been applied yet.
func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files
	migrations, err := readMigrations()
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Apply migrations in order
	for _, m := range migrations {
		if _, exists := applied[m.Name]; !exists {
			log.Printf("Applying migration: %s\n", m.Name)
			
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Execute migration
			if _, err := tx.Exec(m.SQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", m.Name, err)
			}

			// Record migration
			if _, err := tx.Exec(
				"INSERT INTO schema_migrations (name) VALUES ($1)",
				m.Name,
			); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %w", m.Name, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			log.Printf("Applied migration: %s\n", m.Name)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	migrations := make(map[string]bool)

	rows, err := db.Query("SELECT name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations[name] = true
	}

	return migrations, rows.Err()
}

func readMigrations() ([]Migration, error) {
	var migrations []Migration

	dirEntries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return nil, err
	}

	// Filter and sort migration files
	var migrationFiles []fs.DirEntry
	for _, entry := range dirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry)
		}
	}

	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Name() < migrationFiles[j].Name()
	})

	// Read migration files
	for _, entry := range migrationFiles {
		name := entry.Name()
		content, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		migrations = append(migrations, Migration{
			Name: strings.TrimSuffix(name, ".up.sql"),
			SQL:  string(content),
		})
	}

	return migrations, nil
}
