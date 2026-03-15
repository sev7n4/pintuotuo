package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var dbConn *sql.DB

// Init initializes the database connection
func Init() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db?sslmode=disable"
	}

	var err error
	dbConn, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(5)
	dbConn.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return dbConn
}

// Close closes the database connection
func Close() error {
	if dbConn != nil {
		return dbConn.Close()
	}
	return nil
}
