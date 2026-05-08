package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

func AdminListResponseStorage(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)
	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}
	offset := (pageNum - 1) * perPageNum

	statusFilter := c.Query("status")
	userIDFilter := c.Query("user_id")

	whereClause := " WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if statusFilter != "" {
		whereClause += " AND status = $" + strconv.Itoa(argIdx)
		args = append(args, statusFilter)
		argIdx++
	}
	if userIDFilter != "" {
		whereClause += " AND user_id = $" + strconv.Itoa(argIdx)
		args = append(args, userIDFilter)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM stored_responses" + whereClause
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	query := "SELECT id, response_id, user_id, merchant_id, model, status, error_message, background_job_id, created_at, expires_at FROM stored_responses" +
		whereClause + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIdx) + " OFFSET $" + strconv.Itoa(argIdx+1)
	args = append(args, perPageNum, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	type ResponseStorageItem struct {
		ID              int    `json:"id"`
		ResponseID      string `json:"response_id"`
		UserID          int    `json:"user_id"`
		MerchantID      int    `json:"merchant_id"`
		Model           string `json:"model"`
		Status          string `json:"status"`
		ErrorMessage    string `json:"error_message,omitempty"`
		BackgroundJobID string `json:"background_job_id,omitempty"`
		CreatedAt       string `json:"created_at"`
		ExpiresAt       string `json:"expires_at"`
	}

	items := make([]ResponseStorageItem, 0)
	for rows.Next() {
		var item ResponseStorageItem
		var errMsg, bgJobID *string
		if err := rows.Scan(&item.ID, &item.ResponseID, &item.UserID, &item.MerchantID, &item.Model, &item.Status, &errMsg, &bgJobID, &item.CreatedAt, &item.ExpiresAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if errMsg != nil {
			item.ErrorMessage = *errMsg
		}
		if bgJobID != nil {
			item.BackgroundJobID = *bgJobID
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     items,
	})
}

func AdminDeleteResponseStorage(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	result, err := db.Exec("UPDATE stored_responses SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "response not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "response deleted successfully"})
}

func AdminCleanExpiredResponses(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	svc := services.NewResponseStorageService(db)
	deleted, err := svc.CleanExpiredResponses(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "expired responses cleaned",
		"deleted_count": deleted,
	})
}
