package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pintuotuo/backend/config"
	"github.com/stretchr/testify/assert"
)

func setupReferralTestDB() (*sql.DB, error) {
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
		referral_code TEXT,
		referred_by INTEGER,
		total_referrals INTEGER DEFAULT 0,
		total_rewards REAL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS referral_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		code TEXT UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS referrals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		referrer_id INTEGER,
		referee_id INTEGER,
		code_used TEXT,
		status TEXT DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (referrer_id) REFERENCES users(id),
		FOREIGN KEY (referee_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS referral_rewards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		referrer_id INTEGER,
		referee_id INTEGER,
		order_id INTEGER,
		amount REAL,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		paid_at TIMESTAMP,
		FOREIGN KEY (referrer_id) REFERENCES users(id),
		FOREIGN KEY (referee_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		product_id INTEGER,
		group_id INTEGER,
		quantity INTEGER,
		total_price REAL,
		status TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	// Insert test users
	_, err = db.Exec("INSERT INTO users (email, name, password, role) VALUES (?, ?, ?, ?)", "user1@example.com", "User 1", "password1", "user")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("INSERT INTO users (email, name, password, role) VALUES (?, ?, ?, ?)", "user2@example.com", "User 2", "password2", "user")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestGenerateReferralCode(t *testing.T) {
	code := generateReferralCode()
	assert.Len(t, code, referralCodeLength)
	assert.Regexp(t, `^[A-Z0-9]+$`, code)
}

func TestGetMyReferralCode(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Call handler
	GetMyReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "code")
	assert.Len(t, response["code"], referralCodeLength)

	// Verify code is stored in database
	var code string
	err = db.QueryRow("SELECT code FROM referral_codes WHERE user_id = ?", 1).Scan(&code)
	assert.NoError(t, err)
	assert.Equal(t, response["code"], code)

	// Verify code is stored in user table
	var userCode string
	err = db.QueryRow("SELECT referral_code FROM users WHERE id = ?", 1).Scan(&userCode)
	assert.NoError(t, err)
	assert.Equal(t, response["code"], userCode)
}

func TestValidateReferralCode(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test referral code
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: "TESTCODE"}}

	// Call handler
	ValidateReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["valid"].(bool))
	assert.Equal(t, float64(1), response["referrer_id"])
	assert.Equal(t, "User 1", response["referrer_name"])
}

func TestValidateReferralCode_Invalid(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: "INVALID1"}} // 8 characters

	// Call handler
	ValidateReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["valid"].(bool))
}

func TestGetReferralStats(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test data
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	_, err = db.Exec("UPDATE users SET referred_by = ? WHERE id = ?", 1, 2)
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO referrals (referrer_id, referee_id, code_used) VALUES (?, ?, ?)", 1, 2, "TESTCODE")
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO referral_rewards (referrer_id, referee_id, amount, status) VALUES (?, ?, ?, ?)", 1, 2, 10.0, "pending")
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO referral_rewards (referrer_id, referee_id, amount, status) VALUES (?, ?, ?, ?)", 1, 2, 20.0, "paid")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Call handler
	GetReferralStats(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["total_referrals"])
	assert.Equal(t, 30.0, response["total_rewards"])
	assert.Equal(t, 10.0, response["pending_rewards"])
	assert.Equal(t, 20.0, response["paid_rewards"])
}

func TestGetReferralList(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test data
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	_, err = db.Exec("UPDATE users SET referred_by = ? WHERE id = ?", 1, 2)
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO referrals (referrer_id, referee_id, code_used) VALUES (?, ?, ?)", 1, 2, "TESTCODE")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Call handler
	GetReferralList(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(20), response["per_page"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestGetReferralRewards(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test data
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	_, err = db.Exec("UPDATE users SET referred_by = ? WHERE id = ?", 1, 2)
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO referral_rewards (referrer_id, referee_id, amount, status) VALUES (?, ?, ?, ?)", 1, 2, 10.0, "pending")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Call handler
	GetReferralRewards(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["total"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestBindReferralCode(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test referral code
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 2)

	// Set request body
	reqBody := map[string]string{"code": "TESTCODE"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/referral/bind", bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	BindReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify referred_by is set
	var referredBy int
	err = db.QueryRow("SELECT referred_by FROM users WHERE id = ?", 2).Scan(&referredBy)
	assert.NoError(t, err)
	assert.Equal(t, 1, referredBy)

	// Verify referral record is created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM referrals WHERE referrer_id = ? AND referee_id = ?", 1, 2).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify total_referrals is updated
	var totalReferrals int
	err = db.QueryRow("SELECT total_referrals FROM users WHERE id = ?", 1).Scan(&totalReferrals)
	assert.NoError(t, err)
	assert.Equal(t, 1, totalReferrals)
}

func TestBindReferralCode_AlreadyBound(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test referral code
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	// Set user 2 to already be referred by someone
	_, err = db.Exec("UPDATE users SET referred_by = ? WHERE id = ?", 1, 2)
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 2)

	// Set request body
	reqBody := map[string]string{"code": "TESTCODE"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/referral/bind", bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	BindReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBindReferralCode_InvalidCode(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 2)

	// Set request body
	reqBody := map[string]string{"code": "INVALID"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/referral/bind", bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	BindReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBindReferralCode_OwnCode(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test referral code for user 1
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Set request body
	reqBody := map[string]string{"code": "TESTCODE"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/referral/bind", bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	BindReferralCode(c)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCalculateReferralReward(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test referral code
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	// Set user 2 to be referred by user 1
	_, err = db.Exec("UPDATE users SET referred_by = ? WHERE id = ?", 1, 2)
	assert.NoError(t, err)

	// Insert test order
	_, err = db.Exec("INSERT INTO orders (id, user_id, product_id, quantity, total_price, status) VALUES (?, ?, ?, ?, ?, ?)", 1, 2, 1, 1, 100.0, "paid")
	assert.NoError(t, err)

	// Calculate reward
	err = CalculateReferralReward(1, 2, 100.0)
	assert.NoError(t, err)

	// Verify reward is created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM referral_rewards WHERE referrer_id = ? AND referee_id = ? AND order_id = ?", 1, 2, 1).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify reward amount is correct (5% of 100)
	var amount float64
	err = db.QueryRow("SELECT amount FROM referral_rewards WHERE referrer_id = ? AND referee_id = ? AND order_id = ?", 1, 2, 1).Scan(&amount)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, amount)

	// Verify total_rewards is updated
	var totalRewards float64
	err = db.QueryRow("SELECT total_rewards FROM users WHERE id = ?", 1).Scan(&totalRewards)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, totalRewards)
}

func TestPayReferralRewards(t *testing.T) {
	db, err := setupReferralTestDB()
	assert.NoError(t, err)
	defer db.Close()

	// Set the database
	config.DB = db
	defer func() { config.DB = nil }()

	// Insert test referral code
	_, err = db.Exec("INSERT INTO referral_codes (user_id, code) VALUES (?, ?)", 1, "TESTCODE")
	assert.NoError(t, err)

	// Set user 2 to be referred by user 1
	_, err = db.Exec("UPDATE users SET referred_by = ? WHERE id = ?", 1, 2)
	assert.NoError(t, err)

	// Insert test reward
	_, err = db.Exec("INSERT INTO referral_rewards (id, referrer_id, referee_id, amount, status) VALUES (?, ?, ?, ?, ?)", 1, 1, 2, 5.0, "pending")
	assert.NoError(t, err)

	// Create gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Set request body
	reqBody := map[string]interface{}{"reward_ids": []int{1}}
	reqBodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/referral/pay", bytes.NewBuffer(reqBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	PayReferralRewards(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify reward status is updated
	var status string
	err = db.QueryRow("SELECT status FROM referral_rewards WHERE id = ?", 1).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "paid", status)

	// Verify paid_at is set
	var paidAt time.Time
	err = db.QueryRow("SELECT paid_at FROM referral_rewards WHERE id = ?", 1).Scan(&paidAt)
	assert.NoError(t, err)
	assert.NotZero(t, paidAt)
}
