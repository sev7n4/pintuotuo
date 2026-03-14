package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

// ListAPIKeys retrieves all API keys for current user
func ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()

	rows, err := db.Query(
		"SELECT id, user_id, name, status, last_used_at, created_at, updated_at FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var apiKeys []map[string]interface{}
	for rows.Next() {
		var id, userIDScanned int
		var name, status string
		var createdAt, updatedAt time.Time
		var lastUsedAt *time.Time

		err := rows.Scan(&id, &userIDScanned, &name, &status, &lastUsedAt, &createdAt, &updatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		apiKeys = append(apiKeys, map[string]interface{}{
			"id":           id,
			"user_id":      userIDScanned,
			"name":         name,
			"status":       status,
			"last_used_at": lastUsedAt,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
		})
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
	keyHash := hashAPIKey(apiKey)

	db := config.GetDB()

	var id int
	err = db.QueryRow(
		"INSERT INTO api_keys (user_id, key_hash, name, status) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, keyHash, req.Name, "active",
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

	// Return the key (only time it's visible)
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

	c.JSON(http.StatusOK, gin.H{
		"id":         apiID,
		"user_id":    apiUserID,
		"name":       apiName,
		"status":     apiStatus,
		"created_at": apiCreatedAt,
		"updated_at": apiUpdatedAt,
	})
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

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// Helper function to hash API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
