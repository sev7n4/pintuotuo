package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

// GetAdminCatalogModelKeys GET /admin/model-catalog-keys — 上架模型 id 列表，供 fallback 配置参考。
func GetAdminCatalogModelKeys(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	keys, err := services.ListCatalogModelKeys(c.Request.Context(), db)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

// ListModelFallbackRules GET /admin/model-fallback-rules
func ListModelFallbackRules(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	rules, err := services.ListModelFallbackRules(c.Request.Context(), db)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

type modelFallbackWriteBody struct {
	SourceModel    string   `json:"source_model" binding:"required"`
	FallbackModels []string `json:"fallback_models"`
	Enabled        *bool    `json:"enabled"`
	Notes          string   `json:"notes"`
}

// CreateModelFallbackRule POST /admin/model-fallback-rules
func CreateModelFallbackRule(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	var req modelFallbackWriteBody
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	src, chain, vErr := services.ValidateFallbackRule(c.Request.Context(), db, 0, req.SourceModel, req.FallbackModels, enabled)
	if vErr != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_FALLBACK_VALIDATION_FAILED",
			vErr.Error(),
			http.StatusBadRequest,
			vErr,
		))
		return
	}
	var id int
	var notesArg interface{}
	if t := strings.TrimSpace(req.Notes); t != "" {
		notesArg = t
	}
	if insErr := db.QueryRowContext(c.Request.Context(), `
		INSERT INTO model_fallback_rules (source_model, fallback_models, enabled, notes)
		VALUES ($1, $2::text[], $3, $4)
		RETURNING id`,
		src, pq.Array(chain), enabled, notesArg,
	).Scan(&id); insErr != nil {
		if pqErr, ok := insErr.(*pq.Error); ok && pqErr.Code == "23505" {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"MODEL_FALLBACK_SOURCE_EXISTS",
				"该主模型已存在 fallback 规则",
				http.StatusConflict,
				nil,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

// PatchModelFallbackRule PATCH /admin/model-fallback-rules/:id
func PatchModelFallbackRule(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	var req modelFallbackWriteBody
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	src, chain, vErr := services.ValidateFallbackRule(c.Request.Context(), db, id, req.SourceModel, req.FallbackModels, enabled)
	if vErr != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_FALLBACK_VALIDATION_FAILED",
			vErr.Error(),
			http.StatusBadRequest,
			vErr,
		))
		return
	}
	notes := sql.NullString{}
	if t := strings.TrimSpace(req.Notes); t != "" {
		notes = sql.NullString{String: t, Valid: true}
	}
	res, err := db.ExecContext(c.Request.Context(), `
		UPDATE model_fallback_rules
		SET source_model = $1, fallback_models = $2::text[], enabled = $3, notes = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5`,
		src, pq.Array(chain), enabled, notes, id,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"MODEL_FALLBACK_SOURCE_EXISTS",
				"该主模型已存在其他 fallback 规则",
				http.StatusConflict,
				nil,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_FALLBACK_NOT_FOUND",
			"规则不存在",
			http.StatusNotFound,
			nil,
		))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// DeleteModelFallbackRule DELETE /admin/model-fallback-rules/:id
func DeleteModelFallbackRule(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	res, err := db.ExecContext(c.Request.Context(), `DELETE FROM model_fallback_rules WHERE id = $1`, id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_FALLBACK_NOT_FOUND",
			"规则不存在",
			http.StatusNotFound,
			nil,
		))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
