package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// DatabaseURL returns the PostgreSQL connection string used by InitDB (for LISTEN / tooling).
func DatabaseURL() string {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}
	//nolint:gosec,G101
	return "postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db?sslmode=disable"
}

// LoadConfig loads environment variables and initializes configuration
func LoadConfig() error {
	// Environment variables are loaded from .env files
	// This function is called during init()
	return nil
}

// InitDB initializes database connection
func InitDB() error {
	dbURL := DatabaseURL()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	DB = db
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}
