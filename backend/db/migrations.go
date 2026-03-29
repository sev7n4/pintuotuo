package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

type MigrationStatus struct {
	Version    string
	ExecutedAt sql.NullTime
}

const migrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version VARCHAR(255) PRIMARY KEY,
	executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`

func RunMigrations() error {
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}

	migrationsDir := "/app/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "./migrations"
	}

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Println("Migrations directory not found, skipping migrations")
		return nil
	}

	if _, err := dbConn.Exec(migrationsTableSQL); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	log.Println("Ensured schema_migrations table exists")

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}
		sqlFiles = append(sqlFiles, file.Name())
	}

	sort.Strings(sqlFiles)

	executed := 0
	skipped := 0

	for _, filename := range sqlFiles {
		var alreadyExecuted bool
		err := dbConn.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
			filename,
		).Scan(&alreadyExecuted)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		}

		if alreadyExecuted {
			log.Printf("Skipping already executed migration: %s", filename)
			skipped++
			continue
		}

		migrationPath := filepath.Join(migrationsDir, filename)
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		log.Printf("Running migration: %s", filename)

		tx, err := dbConn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			filename,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("Completed migration: %s", filename)
		executed++
	}

	log.Printf("Migrations complete: %d executed, %d skipped", executed, skipped)
	return nil
}

func GetMigrationStatus() ([]MigrationStatus, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := dbConn.Query(
		"SELECT version, executed_at FROM schema_migrations ORDER BY version",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var statuses []MigrationStatus
	for rows.Next() {
		var s MigrationStatus
		if err := rows.Scan(&s.Version, &s.ExecutedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %w", err)
		}
		statuses = append(statuses, s)
	}

	return statuses, nil
}
