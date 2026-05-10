package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/utils"
)

const APIKeyListTTL = 5 * time.Minute

func APIKeyListKey(userID int) string {
	return fmt.Sprintf("apikeys:user:%d", userID)
}

// ListAPIKeys retrieves all API keys for current user
func ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	cacheKey := APIKeyListKey(userIDInt)

	if cachedKeys, err := cache.Get(ctx, cacheKey); err == nil {
		var apiKeys []map[string]interface{}
		if err := json.Unmarshal([]byte(cachedKeys), &apiKeys); err == nil {
			c.JSON(http.StatusOK, apiKeys)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(`
		SELECT id, user_id, name, status, last_used_at, created_at, updated_at,
		       COALESCE(NULLIF(TRIM(key_preview), ''), '') AS key_preview,
		       key_encrypted
		FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Failed to close rows: %v\n", err)
		}
	}()

	var apiKeys []map[string]interface{}
	for rows.Next() {
		var id, userIDScanned int
		var name, status string
		var createdAt, updatedAt time.Time
		var lastUsedAt *time.Time
		var keyPreview string
		var keyEnc sql.NullString

		err := rows.Scan(&id, &userIDScanned, &name, &status, &lastUsedAt, &createdAt, &updatedAt, &keyPreview, &keyEnc)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		canReveal := keyEnc.Valid && strings.TrimSpace(keyEnc.String) != ""

		apiKeys = append(apiKeys, map[string]interface{}{
			"id":           id,
			"user_id":      userIDScanned,
			"name":         name,
			"status":       status,
			"last_used_at": lastUsedAt,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
			"key_preview":  keyPreview,
			"can_reveal":   canReveal,
		})
	}

	if keysJSON, err := json.Marshal(apiKeys); err == nil {
		_ = cache.Set(ctx, cacheKey, string(keysJSON), APIKeyListTTL)
	}

	c.JSON(http.StatusOK, apiKeys)
}

// CreateAPIKey generates a new API key for user
func CreateAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,min=3"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Generate API key
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"KEY_GENERATION_FAILED",
			"Failed to generate API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	apiKey := "ptd_" + hex.EncodeToString(keyBytes)
	keyHash := utils.HashUserAPIKey(apiKey)
	keyPreview := utils.PlatformAPIKeyPreview(apiKey)
	keyEncrypted, encErr := utils.Encrypt(apiKey)
	if encErr != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_ENCRYPT_FAILED",
			"Failed to encrypt API key",
			http.StatusInternalServerError,
			encErr,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var id int
	err = db.QueryRow(
		`INSERT INTO api_keys (user_id, key_hash, name, status, key_encrypted, key_preview)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		userID, keyHash, req.Name, "active", keyEncrypted, keyPreview,
	).Scan(&id)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_CREATION_FAILED",
			"Failed to create API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	userIDInt, _ := userID.(int)
	cache.Delete(context.Background(), APIKeyListKey(userIDInt))

	c.JSON(http.StatusCreated, gin.H{
		"id":     id,
		"key":    apiKey,
		"name":   req.Name,
		"status": "active",
	})
}

// UpdateAPIKey updates API key metadata
func UpdateAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")

	var req struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Verify ownership
	var ownerID int
	err := db.QueryRow("SELECT user_id FROM api_keys WHERE id = $1", id).Scan(&ownerID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrAPIKeyNotFound)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok || ownerID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}

	// Actually, let's use proper scanning
	var apiID, apiUserID int
	var apiName, apiStatus string
	var apiCreatedAt, apiUpdatedAt time.Time

	err = db.QueryRow(
		"UPDATE api_keys SET name = COALESCE(NULLIF($1, ''), name), status = COALESCE(NULLIF($2, ''), status) WHERE id = $3 RETURNING id, user_id, name, status, created_at, updated_at",
		req.Name, req.Status, id,
	).Scan(&apiID, &apiUserID, &apiName, &apiStatus, &apiCreatedAt, &apiUpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_UPDATE_FAILED",
			"Failed to update API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	cache.Delete(context.Background(), APIKeyListKey(userIDInt))

	c.JSON(http.StatusOK, gin.H{
		"id":         apiID,
		"user_id":    apiUserID,
		"name":       apiName,
		"status":     apiStatus,
		"created_at": apiCreatedAt,
		"updated_at": apiUpdatedAt,
	})
}

// RevealAPIKey returns the full platform API key for the owner (decrypted from key_encrypted).
func RevealAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var ownerID int
	var keyEnc sql.NullString
	err := db.QueryRow(
		`SELECT user_id, key_encrypted FROM api_keys WHERE id = $1`,
		id,
	).Scan(&ownerID, &keyEnc)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrAPIKeyNotFound)
		return
	}
	if ownerID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}
	if !keyEnc.Valid || strings.TrimSpace(keyEnc.String) == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"LEGACY_API_KEY_CANNOT_REVEAL",
			"该密钥创建于加密存储上线前，无法再次查看完整内容。请新建密钥并停用或删除本密钥。",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	plain, decErr := utils.Decrypt(keyEnc.String)
	if decErr != nil || plain == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_DECRYPT_FAILED",
			"密钥解密失败，请联系管理员或新建密钥",
			http.StatusInternalServerError,
			decErr,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": plain})
}

// DeleteAPIKey deletes an API key
func DeleteAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Verify ownership
	var ownerID int
	err := db.QueryRow("SELECT user_id FROM api_keys WHERE id = $1", id).Scan(&ownerID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrAPIKeyNotFound)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok || ownerID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM api_keys WHERE id = $1", id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_DELETE_FAILED",
			"Failed to delete API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	cache.Delete(context.Background(), APIKeyListKey(userIDInt))

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}
