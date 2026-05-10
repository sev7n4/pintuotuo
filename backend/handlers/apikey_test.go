package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/utils"
	"github.com/stretchr/testify/assert"
)

func setupAPIKeyTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		password TEXT,
		role TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		key_hash TEXT,
		name TEXT,
		status TEXT,
		last_used_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		key_encrypted TEXT,
		key_preview TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	// Insert test user
	_, err = db.Exec("INSERT INTO users (email, name, password, role) VALUES (?, ?, ?, ?)", "user1@example.com", "User 1", "password1", "user")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestHashAPIKey(t *testing.T) {
	key := "ptd_test_key"
	hash1 := utils.HashUserAPIKey(key)
	hash2 := utils.HashUserAPIKey(key)

	// 相同密钥应该产生相同的哈希
	assert.Equal(t, hash1, hash2)

	// 不同密钥应该产生不同的哈希
	differentKey := "ptd_different_key"
	differentHash := utils.HashUserAPIKey(differentKey)
	assert.NotEqual(t, hash1, differentHash)

	// 哈希长度应该是64个字符（sha256的十六进制表示）
	assert.Len(t, hash1, 64)
}

func TestListAPIKeys(t *testing.T) {
	db, err := setupAPIKeyTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	assert.NoError(t, utils.InitEncryption())

	// Insert test API key
	_, err = db.Exec("INSERT INTO api_keys (user_id, key_hash, name, status) VALUES (?, ?, ?, ?)", 1, "test_hash", "Test Key", "active")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Call handler
	ListAPIKeys(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "Test Key", response[0]["name"])
	assert.Equal(t, "active", response[0]["status"])
	assert.Equal(t, "", response[0]["key_preview"])
	assert.Equal(t, false, response[0]["can_reveal"])
}

func TestCreateAPIKey(t *testing.T) {
	db, err := setupAPIKeyTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	assert.NoError(t, utils.InitEncryption())

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Set request body
	reqBody := map[string]string{"name": "Test API Key"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/apikeys", bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	CreateAPIKey(c)

	// Check response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "key")
	assert.Equal(t, "Test API Key", response["name"])
	assert.Equal(t, "active", response["status"])

	// Verify API key is stored in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE user_id = ?", 1).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	var enc, preview string
	err = db.QueryRow("SELECT key_encrypted, key_preview FROM api_keys WHERE user_id = ?", 1).Scan(&enc, &preview)
	assert.NoError(t, err)
	assert.NotEmpty(t, enc)
	assert.Contains(t, preview, "ptd_")
}

func TestRevealAPIKey(t *testing.T) {
	db, err := setupAPIKeyTestDB()
	assert.NoError(t, err)
	defer db.Close()

	config.DB = db
	defer func() { config.DB = nil }()

	assert.NoError(t, utils.InitEncryption())

	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Set("user_id", 1)
	reqBody := map[string]string{"name": "Reveal Key"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	cCreate.Request, _ = http.NewRequest("POST", "/tokens/keys", bytes.NewBuffer(reqBodyBytes))
	cCreate.Request.Header.Set("Content-Type", "application/json")
	CreateAPIKey(cCreate)
	assert.Equal(t, http.StatusCreated, wCreate.Code)

	var created map[string]interface{}
	assert.NoError(t, json.Unmarshal(wCreate.Body.Bytes(), &created))
	keyID := int(created["id"].(float64))
	fullKey := created["key"].(string)

	wReveal := httptest.NewRecorder()
	cReveal, _ := gin.CreateTestContext(wReveal)
	cReveal.Set("user_id", 1)
	cReveal.Params = gin.Params{{Key: "id", Value: strconv.Itoa(keyID)}}
	cReveal.Request, _ = http.NewRequest("POST", "/tokens/keys/"+strconv.Itoa(keyID)+"/reveal", nil)
	RevealAPIKey(cReveal)
	assert.Equal(t, http.StatusOK, wReveal.Code)
	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(wReveal.Body.Bytes(), &out))
	assert.Equal(t, fullKey, out["key"])
}

func TestUpdateAPIKey(t *testing.T) {
	db, err := setupAPIKeyTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test API key
	var apiKeyID int
	err = db.QueryRow("INSERT INTO api_keys (user_id, key_hash, name, status) VALUES (?, ?, ?, ?) RETURNING id", 1, "test_hash", "Old Name", "active").Scan(&apiKeyID)
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(apiKeyID)}}

	// Set request body
	reqBody := map[string]string{"name": "New Name", "status": "inactive"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("PUT", "/api/apikeys/"+strconv.Itoa(apiKeyID), bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	UpdateAPIKey(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", response["name"])
	assert.Equal(t, "inactive", response["status"])

	// Verify API key is updated in database
	var name, status string
	err = db.QueryRow("SELECT name, status FROM api_keys WHERE id = ?", apiKeyID).Scan(&name, &status)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", name)
	assert.Equal(t, "inactive", status)
}

func TestDeleteAPIKey(t *testing.T) {
	db, err := setupAPIKeyTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test API key
	var apiKeyID int
	err = db.QueryRow("INSERT INTO api_keys (user_id, key_hash, name, status) VALUES (?, ?, ?, ?) RETURNING id", 1, "test_hash", "Test Key", "active").Scan(&apiKeyID)
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(apiKeyID)}}

	// Call handler
	DeleteAPIKey(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "API key deleted successfully", response["message"])

	// Verify API key is deleted from database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE id = ?", apiKeyID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
