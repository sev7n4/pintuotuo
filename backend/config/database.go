package config

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// LoadConfig loads environment variables and initializes configuration
func LoadConfig() error {
	// Environment variables are loaded from .env files
	// This function is called during init()
	return nil
}

// InitDB initializes database connection
func InitDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to development database
		dbURL = "postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings based on environment
	maxOpenConns := 100
	maxIdleConns := 25

	// Reduce connections in CI/CD or test mode to avoid hitting PostgreSQL limits
	if os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TEST_MODE") == "true" {
		maxOpenConns = 15
		maxIdleConns = 5
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	DB = db

	// Truncate and seed once if in test mode
	if os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TEST_MODE") == "true" {
		TruncateAndSeed()
	}

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

var seedOnce sync.Once

// TruncateAndSeed cleans the database and seeds basic test data
// Used primarily in CI/CD environments to ensure test isolation
func TruncateAndSeed() {
	seedOnce.Do(func() {
		db := GetDB()
		if db == nil {
			return
		}

		// Truncate all relevant tables in correct order
		tables := []string{"token_transactions", "tokens", "payments", "group_members", "orders", "groups", "products", "users"}
		for _, table := range tables {
			_, _ = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		}

		// Seed basic data with ON CONFLICT DO NOTHING
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (1, 'test@example.com', 'Test User', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (2, 'test2@example.com', 'Test User 2', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (3, 'test3@example.com', 'Test User 3', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (4, 'test4@example.com', 'Test User 4', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (5, 'test5@example.com', 'Test User 5', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (6, 'test6@example.com', 'Test User 6', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (7, 'test7@example.com', 'Test User 7', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (8, 'test8@example.com', 'Test User 8', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (9, 'test9@example.com', 'Test User 9', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (10, 'test10@example.com', 'Test User 10', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (15, 'test15@example.com', 'Test User 15', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (16, 'test16@example.com', 'Test User 16', 'hash') ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO users (id, email, name, password_hash) VALUES (17, 'test17@example.com', 'Test User 17', 'hash') ON CONFLICT DO NOTHING")

		_, _ = db.Exec("INSERT INTO products (id, merchant_id, name, price, stock) VALUES (1, 1, 'Test Product', 99.99, 1000) ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO tokens (user_id, balance) VALUES (1, 1000.00) ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO tokens (user_id, balance) VALUES (2, 1000.00) ON CONFLICT DO NOTHING")
		_, _ = db.Exec("INSERT INTO tokens (user_id, balance) VALUES (10, 1000.00) ON CONFLICT DO NOTHING")
	})
}
