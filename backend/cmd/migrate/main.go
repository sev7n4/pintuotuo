package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Failed to close database connection: %v", closeErr)
		}
	}()

	if pingErr := db.Ping(); pingErr != nil {
		log.Fatalf("Failed to ping database: %v", pingErr)
	}

	log.Println("Connected to database")

	migrationsDir := "./migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		migrationPath := filepath.Join(migrationsDir, file.Name())
		log.Printf("Running migration: %s", file.Name())

		sqlBytes, readErr := os.ReadFile(migrationPath)
		if readErr != nil {
			log.Fatalf("Failed to read migration file %s: %v", file.Name(), readErr)
		}

		if _, execErr := db.Exec(string(sqlBytes)); execErr != nil {
			log.Fatalf("Failed to execute migration %s: %v", file.Name(), execErr)
		}

		log.Printf("Completed migration: %s", file.Name())
	}

	log.Println("All migrations completed successfully")
}
