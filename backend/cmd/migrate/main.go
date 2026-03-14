package main

import (
	"database/sql"
	"io/ioutil"
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
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✓ Connected to database")

	// Read and execute migration files
	migrationsDir := "./migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		migrationPath := filepath.Join(migrationsDir, file.Name())
		log.Printf("Running migration: %s", file.Name())

		sqlBytes, err := ioutil.ReadFile(migrationPath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file.Name(), err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", file.Name(), err)
		}

		log.Printf("✓ Completed migration: %s", file.Name())
	}

	log.Println("✓ All migrations completed successfully")
}
