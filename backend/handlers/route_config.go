package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type ProviderRouteConfig struct {
	ID             int                    `json:"id"`
	Code           string                 `json:"code"`
	Name           string                 `json:"name"`
	ProviderRegion string                 `json:"provider_region"`
	RouteStrategy  map[string]interface{} `json:"route_strategy"`
	Endpoints      map[string]interface{} `json:"endpoints"`
	Status         string                 `json:"status"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

type UpdateProviderRouteConfigRequest struct {
	ProviderRegion string                 `json:"provider_region"`
	RouteStrategy  map[string]interface{} `json:"route_strategy"`
	Endpoints      map[string]interface{} `json:"endpoints"`
}

func GetProviderRouteConfigs(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	region := c.Query("region")
	status := c.Query("status")

	query := `SELECT id, code, name, COALESCE(provider_region, 'domestic'), 
		COALESCE(route_strategy, '{}'::jsonb), COALESCE(endpoints, '{}'::jsonb),
		status, created_at, updated_at
		FROM model_providers WHERE 1=1`
	var args []interface{}
	argPos := 1

	if region != "" {
		query += " AND provider_region = $" + strconv.Itoa(argPos)
		args = append(args, region)
		argPos++
	}

	if status != "" {
		query += " AND status = $" + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}

	query += " ORDER BY id ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var configs []ProviderRouteConfig
	for rows.Next() {
		var cfg ProviderRouteConfig
		var routeStrategyJSON, endpointsJSON []byte

		err := rows.Scan(
			&cfg.ID, &cfg.Code, &cfg.Name, &cfg.ProviderRegion,
			&routeStrategyJSON, &endpointsJSON,
			&cfg.Status, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if len(routeStrategyJSON) > 0 {
			json.Unmarshal(routeStrategyJSON, &cfg.RouteStrategy)
		}
		if len(endpointsJSON) > 0 {
			json.Unmarshal(endpointsJSON, &cfg.Endpoints)
		}

		if cfg.RouteStrategy == nil {
			cfg.RouteStrategy = make(map[string]interface{})
		}
		if cfg.Endpoints == nil {
			cfg.Endpoints = make(map[string]interface{})
		}

		configs = append(configs, cfg)
	}

	if configs == nil {
		configs = []ProviderRouteConfig{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    configs,
	})
}

func GetProviderRouteConfig(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	code := c.Param("code")
	if code == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Provider code is required",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var cfg ProviderRouteConfig
	var routeStrategyJSON, endpointsJSON []byte

	err := db.QueryRow(
		`SELECT id, code, name, COALESCE(provider_region, 'domestic'),
		COALESCE(route_strategy, '{}'::jsonb), COALESCE(endpoints, '{}'::jsonb),
		status, created_at, updated_at
		FROM model_providers WHERE code = $1`,
		code,
	).Scan(
		&cfg.ID, &cfg.Code, &cfg.Name, &cfg.ProviderRegion,
		&routeStrategyJSON, &endpointsJSON,
		&cfg.Status, &cfg.CreatedAt, &cfg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NOT_FOUND",
			"Provider not found",
			http.StatusNotFound,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if len(routeStrategyJSON) > 0 {
		json.Unmarshal(routeStrategyJSON, &cfg.RouteStrategy)
	}
	if len(endpointsJSON) > 0 {
		json.Unmarshal(endpointsJSON, &cfg.Endpoints)
	}

	if cfg.RouteStrategy == nil {
		cfg.RouteStrategy = make(map[string]interface{})
	}
	if cfg.Endpoints == nil {
		cfg.Endpoints = make(map[string]interface{})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    cfg,
	})
}

func UpdateProviderRouteConfig(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	code := c.Param("code")
	if code == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Provider code is required",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	var req UpdateProviderRouteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.ProviderRegion == "" {
		req.ProviderRegion = "domestic"
	}
	if req.RouteStrategy == nil {
		req.RouteStrategy = make(map[string]interface{})
	}
	if req.Endpoints == nil {
		req.Endpoints = make(map[string]interface{})
	}

	routeStrategyJSON, err := json.Marshal(req.RouteStrategy)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Invalid route_strategy format",
			http.StatusBadRequest,
			err,
		))
		return
	}

	endpointsJSON, err := json.Marshal(req.Endpoints)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Invalid endpoints format",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	result, err := db.Exec(
		`UPDATE model_providers 
		SET provider_region = $1, route_strategy = $2, endpoints = $3, updated_at = CURRENT_TIMESTAMP
		WHERE code = $4`,
		req.ProviderRegion, routeStrategyJSON, endpointsJSON, code,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NOT_FOUND",
			"Provider not found",
			http.StatusNotFound,
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Provider route config updated successfully",
	})
}

type MerchantRouteConfig struct {
	ID              int                    `json:"id"`
	Name            string                 `json:"name"`
	MerchantType    string                 `json:"merchant_type"`
	Region          string                 `json:"region"`
	RoutePreference map[string]interface{} `json:"route_preference"`
	Status          string                 `json:"status"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

type UpdateMerchantRouteConfigRequest struct {
	MerchantType    string                 `json:"merchant_type"`
	Region          string                 `json:"region"`
	RoutePreference map[string]interface{} `json:"route_preference"`
}

func GetMerchantRouteConfigs(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantType := c.Query("type")
	region := c.Query("region")
	status := c.Query("status")

	query := `SELECT id, name, COALESCE(merchant_type, 'standard'), COALESCE(region, 'domestic'),
		COALESCE(route_preference, '{}'::jsonb), status, created_at, updated_at
		FROM merchants WHERE 1=1`
	var args []interface{}
	argPos := 1

	if merchantType != "" {
		query += " AND merchant_type = $" + strconv.Itoa(argPos)
		args = append(args, merchantType)
		argPos++
	}

	if region != "" {
		query += " AND region = $" + strconv.Itoa(argPos)
		args = append(args, region)
		argPos++
	}

	if status != "" {
		query += " AND status = $" + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}

	query += " ORDER BY id DESC LIMIT 100"

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var configs []MerchantRouteConfig
	for rows.Next() {
		var cfg MerchantRouteConfig
		var routePrefJSON []byte

		err := rows.Scan(
			&cfg.ID, &cfg.Name, &cfg.MerchantType, &cfg.Region,
			&routePrefJSON, &cfg.Status, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if len(routePrefJSON) > 0 {
			json.Unmarshal(routePrefJSON, &cfg.RoutePreference)
		}

		if cfg.RoutePreference == nil {
			cfg.RoutePreference = make(map[string]interface{})
		}

		configs = append(configs, cfg)
	}

	if configs == nil {
		configs = []MerchantRouteConfig{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    configs,
	})
}

func UpdateMerchantRouteConfig(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Merchant ID is required",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Invalid merchant ID",
			http.StatusBadRequest,
			err,
		))
		return
	}

	var req UpdateMerchantRouteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.MerchantType == "" {
		req.MerchantType = "standard"
	}
	if req.Region == "" {
		req.Region = "domestic"
	}
	if req.RoutePreference == nil {
		req.RoutePreference = make(map[string]interface{})
	}

	routePrefJSON, err := json.Marshal(req.RoutePreference)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Invalid route_preference format",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	result, err := db.Exec(
		`UPDATE merchants 
		SET merchant_type = $1, region = $2, route_preference = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4`,
		req.MerchantType, req.Region, routePrefJSON, id,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Merchant route config updated successfully",
	})
}

type RouteTestRequest struct {
	ProviderCode string `json:"provider_code" binding:"required"`
	MerchantID   int    `json:"merchant_id" binding:"required"`
}

type RouteTestResult struct {
	Mode             string                 `json:"mode"`
	Endpoint         string                 `json:"endpoint"`
	FallbackMode     string                 `json:"fallback_mode"`
	FallbackEndpoint string                 `json:"fallback_endpoint"`
	Reason           string                 `json:"reason"`
	ProviderConfig   map[string]interface{} `json:"provider_config"`
	MerchantConfig   map[string]interface{} `json:"merchant_config"`
}

func TestRouteDecision(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	var req RouteTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var providerRegion string
	var routeStrategyJSON, endpointsJSON []byte
	err := db.QueryRow(
		`SELECT COALESCE(provider_region, 'domestic'),
		COALESCE(route_strategy, '{}'::jsonb), COALESCE(endpoints, '{}'::jsonb)
		FROM model_providers WHERE code = $1`,
		req.ProviderCode,
	).Scan(&providerRegion, &routeStrategyJSON, &endpointsJSON)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NOT_FOUND",
			"Provider not found",
			http.StatusNotFound,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var routeStrategy, endpoints map[string]interface{}
	json.Unmarshal(routeStrategyJSON, &routeStrategy)
	json.Unmarshal(endpointsJSON, &endpoints)

	var merchantType, region string
	var routePrefJSON []byte
	err = db.QueryRow(
		`SELECT COALESCE(merchant_type, 'standard'), COALESCE(region, 'domestic'),
		COALESCE(route_preference, '{}'::jsonb)
		FROM merchants WHERE id = $1`,
		req.MerchantID,
	).Scan(&merchantType, &region, &routePrefJSON)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var routePref map[string]interface{}
	json.Unmarshal(routePrefJSON, &routePref)

	router := services.NewUnifiedRouter(nil)
	decision, err := router.DecideRoute(nil,
		&services.ProviderConfig{
			Code:           req.ProviderCode,
			ProviderRegion: providerRegion,
			RouteStrategy:  routeStrategy,
			Endpoints:      endpoints,
		},
		&services.MerchantConfig{
			ID:              req.MerchantID,
			Type:            merchantType,
			Region:          region,
			RoutePreference: routePref,
		},
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ROUTE_ERROR",
			"Failed to calculate route decision",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	result := RouteTestResult{
		Mode:             decision.Mode,
		Endpoint:         decision.Endpoint,
		FallbackMode:     decision.FallbackMode,
		FallbackEndpoint: decision.FallbackEndpoint,
		Reason:           decision.Reason,
		ProviderConfig: map[string]interface{}{
			"code":            req.ProviderCode,
			"provider_region": providerRegion,
			"route_strategy":  routeStrategy,
			"endpoints":       endpoints,
		},
		MerchantConfig: map[string]interface{}{
			"id":               req.MerchantID,
			"merchant_type":    merchantType,
			"region":           region,
			"route_preference": routePref,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}
