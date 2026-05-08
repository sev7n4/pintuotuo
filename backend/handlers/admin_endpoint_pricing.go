package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type EndpointPricing struct {
	ID           int       `json:"id"`
	EndpointType string    `json:"endpoint_type"`
	ProviderCode string    `json:"provider_code"`
	UnitType     string    `json:"unit_type"`
	UnitPrice    float64   `json:"unit_price"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type EndpointPricingCreateRequest struct {
	EndpointType string  `json:"endpoint_type" binding:"required"`
	ProviderCode string  `json:"provider_code" binding:"required"`
	UnitType     string  `json:"unit_type" binding:"required"`
	UnitPrice    float64 `json:"unit_price" binding:"required"`
}

type EndpointPricingUpdateRequest struct {
	EndpointType string  `json:"endpoint_type"`
	ProviderCode string  `json:"provider_code"`
	UnitType     string  `json:"unit_type"`
	UnitPrice    float64 `json:"unit_price"`
}

var validEndpointTypes = map[string]bool{
	services.EndpointTypeChatCompletions:     true,
	services.EndpointTypeEmbeddings:          true,
	services.EndpointTypeImagesGenerations:   true,
	services.EndpointTypeImagesVariations:    true,
	services.EndpointTypeImagesEdits:         true,
	services.EndpointTypeAudioTranscriptions: true,
	services.EndpointTypeAudioTranslations:   true,
	services.EndpointTypeAudioSpeech:         true,
	services.EndpointTypeModerations:         true,
	services.EndpointTypeResponses:           true,
}

var validUnitTypes = map[string]bool{
	string(billing.BillingUnitToken):     true,
	string(billing.BillingUnitImage):     true,
	string(billing.BillingUnitSecond):    true,
	string(billing.BillingUnitCharacter): true,
	string(billing.BillingUnitRequest):   true,
}

func AdminListEndpointPricing(c *gin.Context) {
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

	endpointTypeFilter := c.Query("endpoint_type")
	providerCodeFilter := c.Query("provider_code")

	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if endpointTypeFilter != "" {
		whereClause += " WHERE endpoint_type = $" + strconv.Itoa(argIdx)
		args = append(args, endpointTypeFilter)
		argIdx++
	}
	if providerCodeFilter != "" {
		if whereClause == "" {
			whereClause += " WHERE provider_code = $" + strconv.Itoa(argIdx)
		} else {
			whereClause += " AND provider_code = $" + strconv.Itoa(argIdx)
		}
		args = append(args, providerCodeFilter)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM endpoint_pricing" + whereClause
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	query := "SELECT id, endpoint_type, provider_code, unit_type, unit_price, created_at, updated_at FROM endpoint_pricing" +
		whereClause + " ORDER BY id ASC LIMIT $" + strconv.Itoa(argIdx) + " OFFSET $" + strconv.Itoa(argIdx+1)
	args = append(args, perPageNum, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	items := make([]EndpointPricing, 0)
	for rows.Next() {
		var item EndpointPricing
		if err := rows.Scan(&item.ID, &item.EndpointType, &item.ProviderCode, &item.UnitType, &item.UnitPrice, &item.CreatedAt, &item.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
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

func AdminCreateEndpointPricing(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	var req EndpointPricingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if !validEndpointTypes[req.EndpointType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint_type"})
		return
	}
	if !validUnitTypes[req.UnitType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit_type"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var item EndpointPricing
	err := db.QueryRow(
		`INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES ($1, $2, $3, $4) RETURNING id, endpoint_type, provider_code, unit_type, unit_price, created_at, updated_at`,
		req.EndpointType, req.ProviderCode, req.UnitType, req.UnitPrice,
	).Scan(&item.ID, &item.EndpointType, &item.ProviderCode, &item.UnitType, &item.UnitPrice, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "endpoint pricing already exists for this endpoint_type and provider_code"})
		return
	}

	ctx := context.Background()
	cache.InvalidatePatterns(ctx, "endpoint_pricing:*")

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func AdminUpdateEndpointPricing(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req EndpointPricingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var existing EndpointPricing
	err = db.QueryRow("SELECT id, endpoint_type, provider_code, unit_type, unit_price, created_at, updated_at FROM endpoint_pricing WHERE id = $1", id).
		Scan(&existing.ID, &existing.EndpointType, &existing.ProviderCode, &existing.UnitType, &existing.UnitPrice, &existing.CreatedAt, &existing.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint pricing not found"})
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	endpointType := existing.EndpointType
	if req.EndpointType != "" {
		if !validEndpointTypes[req.EndpointType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint_type"})
			return
		}
		endpointType = req.EndpointType
	}

	providerCode := existing.ProviderCode
	if req.ProviderCode != "" {
		providerCode = req.ProviderCode
	}

	unitType := existing.UnitType
	if req.UnitType != "" {
		if !validUnitTypes[req.UnitType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit_type"})
			return
		}
		unitType = req.UnitType
	}

	unitPrice := existing.UnitPrice
	if req.UnitPrice > 0 {
		unitPrice = req.UnitPrice
	}

	var item EndpointPricing
	err = db.QueryRow(
		`UPDATE endpoint_pricing SET endpoint_type = $1, provider_code = $2, unit_type = $3, unit_price = $4, updated_at = NOW() WHERE id = $5 RETURNING id, endpoint_type, provider_code, unit_type, unit_price, created_at, updated_at`,
		endpointType, providerCode, unitType, unitPrice, id,
	).Scan(&item.ID, &item.EndpointType, &item.ProviderCode, &item.UnitType, &item.UnitPrice, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	ctx := context.Background()
	cache.InvalidatePatterns(ctx, "endpoint_pricing:*")

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func AdminDeleteEndpointPricing(c *gin.Context) {
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

	result, err := db.Exec("DELETE FROM endpoint_pricing WHERE id = $1", id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint pricing not found"})
		return
	}

	ctx := context.Background()
	cache.InvalidatePatterns(ctx, "endpoint_pricing:*")

	c.JSON(http.StatusOK, gin.H{"message": "endpoint pricing deleted successfully"})
}
