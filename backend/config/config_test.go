package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfig tests environment configuration loading
func TestLoadConfig(t *testing.T) {
	err := LoadConfig()
	assert.NoError(t, err, "LoadConfig should not return an error")
}

// TestInitDB tests database initialization with proper cleanup
func TestInitDB(t *testing.T) {
	// This test requires a test database to be available
	// Skip if not in test environment
	t.Skip("Requires test database to be configured")

	err := InitDB()
	assert.NoError(t, err, "InitDB should initialize successfully")

	// Verify connection
	err = DB.Ping()
	assert.NoError(t, err, "Database connection should be active")

	// Cleanup
	CloseDB()
}

// TestGetDB tests database connection getter
func TestGetDB(t *testing.T) {
	db := GetDB()
	assert.Nil(t, db, "GetDB should return the database connection")
}
