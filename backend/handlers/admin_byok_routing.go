package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

type BYOKRoutingListItem struct {
	ID                  int                    `json:"id"`
	MerchantID          int                    `json:"merchant_id"`
	CompanyName         string                 `json:"company_name"`
	BYOKType            string                 `json:"byok_type"`
	Provider            string                 `json:"provider"`
	Name                string                 `json:"name"`
	Region              string                 `json:"region"`
	RouteMode           string                 `json:"route_mode"`
	EndpointURL         string                 `json:"endpoint_url"`
	FallbackEndpointURL string                 `json:"fallback_endpoint_url"`
	RouteConfig         map[string]interface{} `json:"route_config"`
	HealthStatus        string                 `json:"health_status"`
	HealthErrorMessage  string                 `json:"health_error_message"`
	HealthErrorCategory string                 `json:"health_error_category"`
	HealthErrorCode     string                 `json:"health_error_code"`
	LastHealthCheckAt   *string                `json:"last_health_check_at"`
	VerificationResult  string                 `json:"verification_result"`
	VerificationMessage string                 `json:"verification_message"`
	ModelsSupported     []string               `json:"models_supported"`
	VerifiedAt          *string                `json:"verified_at"`
	Status              string                 `json:"status"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

type BYOKRoutingListResponse struct {
	Data  []BYOKRoutingListItem `json:"data"`
	Total int                   `json:"total"`
}

func GetBYOKRoutingList(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantIDStr := c.Query("merchant_id")
	byokType := strings.ToLower(strings.TrimSpace(c.Query("byok_type")))
	provider := strings.ToLower(strings.TrimSpace(c.Query("provider")))
	region := strings.ToLower(strings.TrimSpace(c.Query("region")))
	routeMode := strings.ToLower(strings.TrimSpace(c.Query("route_mode")))
	healthStatus := strings.ToLower(strings.TrimSpace(c.Query("health_status")))

	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "1=1")

	if merchantIDStr != "" {
		if merchantID, err := strconv.Atoi(merchantIDStr); err == nil && merchantID > 0 {
			conditions = append(conditions, "mak.merchant_id = $"+strconv.Itoa(argIndex))
			args = append(args, merchantID)
			argIndex++
		}
	}

	if byokType != "" {
		switch byokType {
		case byokTypeOfficial, byokTypeReseller, byokTypeSelfHosted:
			conditions = append(conditions, "mak.byok_type = $"+strconv.Itoa(argIndex))
			args = append(args, byokType)
			argIndex++
		}
	}

	if provider != "" {
		conditions = append(conditions, "LOWER(mak.provider) = $"+strconv.Itoa(argIndex))
		args = append(args, provider)
		argIndex++
	}

	if region != "" {
		switch region {
		case regionDomestic, regionOverseas:
			conditions = append(conditions, "mak.region = $"+strconv.Itoa(argIndex))
			args = append(args, region)
			argIndex++
		}
	}

	if routeMode != "" {
		conditions = append(conditions, "mak.route_mode = $"+strconv.Itoa(argIndex))
		args = append(args, routeMode)
		argIndex++
	}

	if healthStatus != "" {
		conditions = append(conditions, "LOWER(mak.health_status) = $"+strconv.Itoa(argIndex))
		args = append(args, healthStatus)
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := `SELECT COUNT(*) FROM merchant_api_keys mak WHERE ` + whereClause
	var total int
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	query := `SELECT mak.id, mak.merchant_id, COALESCE(NULLIF(TRIM(m.company_name), ''), '—') as company_name,
		COALESCE(mak.byok_type, 'official'), mak.provider, mak.name,
		COALESCE(mak.region, 'domestic'), COALESCE(mak.route_mode, 'auto'),
		COALESCE(mak.endpoint_url, ''), COALESCE(mak.fallback_endpoint_url, ''),
		COALESCE(mak.route_config, '{}'::jsonb),
		COALESCE(NULLIF(TRIM(mak.health_status), ''), 'unknown'),
		COALESCE((
			SELECT h.error_message
			FROM api_key_health_history h
			WHERE h.api_key_id = mak.id
			ORDER BY h.created_at DESC
			LIMIT 1
		), ''),
		COALESCE((
			SELECT h.error_category
			FROM api_key_health_history h
			WHERE h.api_key_id = mak.id
			ORDER BY h.created_at DESC
			LIMIT 1
		), ''),
		COALESCE((
			SELECT h.provider_error_code
			FROM api_key_health_history h
			WHERE h.api_key_id = mak.id
			ORDER BY h.created_at DESC
			LIMIT 1
		), ''),
		mak.last_health_check_at,
		COALESCE(NULLIF(TRIM(mak.verification_result), ''), 'pending'),
		COALESCE(mak.verification_message, ''),
		mak.models_supported,
		mak.verified_at,
		mak.status, mak.created_at, mak.updated_at
		FROM merchant_api_keys mak
		LEFT JOIN merchants m ON m.id = mak.merchant_id
		WHERE ` + whereClause + ` ORDER BY mak.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var items []BYOKRoutingListItem
	for rows.Next() {
		var item BYOKRoutingListItem
		var routeConfigBytes []byte
		var createdAt, updatedAt sql.NullTime
		var lastHealthCheckAt sql.NullTime
		var verifiedAt sql.NullTime
		var modelsJSON []byte
		if err := rows.Scan(
			&item.ID, &item.MerchantID, &item.CompanyName,
			&item.BYOKType, &item.Provider, &item.Name,
			&item.Region, &item.RouteMode,
			&item.EndpointURL, &item.FallbackEndpointURL,
			&routeConfigBytes,
			&item.HealthStatus,
			&item.HealthErrorMessage, &item.HealthErrorCategory, &item.HealthErrorCode,
			&lastHealthCheckAt,
			&item.VerificationResult, &item.VerificationMessage,
			&modelsJSON,
			&verifiedAt,
			&item.Status, &createdAt, &updatedAt,
		); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if len(routeConfigBytes) > 0 {
			_ = json.Unmarshal(routeConfigBytes, &item.RouteConfig)
		}
		if len(modelsJSON) > 0 {
			_ = json.Unmarshal(modelsJSON, &item.ModelsSupported)
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if lastHealthCheckAt.Valid {
			t := lastHealthCheckAt.Time.Format("2006-01-02 15:04:05")
			item.LastHealthCheckAt = &t
		}
		if verifiedAt.Valid {
			t := verifiedAt.Time.Format("2006-01-02 15:04:05")
			item.VerifiedAt = &t
		}
		items = append(items, item)
	}

	if items == nil {
		items = []BYOKRoutingListItem{}
	}

	c.JSON(http.StatusOK, BYOKRoutingListResponse{
		Data:  items,
		Total: total,
	})
}

type UpdateBYOKRouteConfigRequest struct {
	RouteMode           string                 `json:"route_mode"`
	EndpointURL         string                 `json:"endpoint_url"`
	FallbackEndpointURL string                 `json:"fallback_endpoint_url"`
	RouteConfig         map[string]interface{} `json:"route_config"`
}

func UpdateBYOKRouteConfig(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req UpdateBYOKRouteConfigRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	routeMode := strings.ToLower(strings.TrimSpace(req.RouteMode))
	if routeMode != "" {
		switch routeMode {
		case routeModeAuto, routeModeDirect, routeModeLitellm, routeModeProxy:
		default:
			middleware.RespondWithError(c, apperrors.NewAppError(
				"INVALID_ROUTE_MODE",
				"route_mode must be auto, direct, litellm, or proxy",
				http.StatusBadRequest,
				nil,
			))
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var routeConfigJSON []byte
	if req.RouteConfig != nil {
		routeConfigJSON, _ = json.Marshal(req.RouteConfig)
	} else {
		routeConfigJSON = []byte("{}")
	}

	var merchantID int
	err = db.QueryRow(
		`UPDATE merchant_api_keys SET 
		 route_mode = CASE WHEN $1::text = '' THEN route_mode ELSE $1::varchar(20) END,
		 endpoint_url = CASE WHEN $2::text = '' THEN endpoint_url ELSE NULLIF(TRIM($2::text), '')::varchar(500) END,
		 fallback_endpoint_url = CASE WHEN $3::text = '' THEN fallback_endpoint_url ELSE NULLIF(TRIM($3::text), '')::varchar(500) END,
		 route_config = $4::jsonb,
		 updated_at = CURRENT_TIMESTAMP
		 WHERE id = $5
		 RETURNING merchant_id`,
		routeMode, req.EndpointURL, req.FallbackEndpointURL, routeConfigJSON, keyID,
	).Scan(&merchantID)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantAPIKeysKey(merchantID))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Route config updated successfully",
		"api_key_id": keyID,
	})
}

func TriggerBYOKProbe(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var existsKey bool
	err = db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM merchant_api_keys WHERE id = $1 AND status = 'active')`,
		keyID,
	).Scan(&existsKey)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !existsKey {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found or inactive",
			http.StatusNotFound,
			nil,
		))
		return
	}

	go func() {
		if checkErr := services.GetHealthScheduler().TriggerImmediateCheck(keyID); checkErr != nil {
			logger.LogError(context.Background(), "admin_byok_routing", "Immediate health check failed", checkErr, map[string]interface{}{
				"api_key_id": keyID,
			})
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Probe triggered successfully",
		"api_key_id": keyID,
	})
}

func LightVerifyBYOK(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var apiKey models.MerchantAPIKey
	err = db.QueryRow(
		`SELECT id, merchant_id, provider, api_key_encrypted, route_mode, route_config, region
		 FROM merchant_api_keys
		 WHERE id = $1 AND status = 'active'`,
		keyID,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.RouteMode, &apiKey.RouteConfig, &apiKey.Region)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found or inactive",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	validator := services.GetAPIKeyValidator()
	err = validator.ValidateAsyncWithRouteMode(
		apiKey.ID, apiKey.Provider, apiKey.APIKeyEncrypted, "admin_light",
		apiKey.RouteMode, apiKey.RouteConfig, apiKey.Region,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"VERIFICATION_FAILED",
			"Failed to start verification",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	cache.Delete(context.Background(), cache.MerchantAPIKeysKey(apiKey.MerchantID))

	c.JSON(http.StatusAccepted, gin.H{
		"message":           "Light verification started",
		"api_key_id":        apiKey.ID,
		"verification_type": "admin_light",
	})
}

func DeepVerifyBYOK(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var apiKey models.MerchantAPIKey
	err = db.QueryRow(
		`SELECT id, merchant_id, provider, api_key_encrypted, route_mode, route_config, region
		 FROM merchant_api_keys
		 WHERE id = $1 AND status = 'active'`,
		keyID,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.RouteMode, &apiKey.RouteConfig, &apiKey.Region)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found or inactive",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	validator := services.GetAPIKeyValidator()
	err = validator.ValidateAsyncWithRouteMode(
		apiKey.ID, apiKey.Provider, apiKey.APIKeyEncrypted, "admin_deep",
		apiKey.RouteMode, apiKey.RouteConfig, apiKey.Region,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"VERIFICATION_FAILED",
			"Failed to start verification",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	cache.Delete(context.Background(), cache.MerchantAPIKeysKey(apiKey.MerchantID))

	c.JSON(http.StatusAccepted, gin.H{
		"message":           "Deep verification started",
		"api_key_id":        apiKey.ID,
		"verification_type": "admin_deep",
	})
}

func GetBYOKVerificationDetails(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var apiKey models.MerchantAPIKey
	var verificationResult, verificationMsg sql.NullString
	var verifiedAt sql.NullTime
	var modelsJSON []byte
	err = db.QueryRow(
		`SELECT id, merchant_id, provider, verification_result, verified_at, models_supported, verification_message
		 FROM merchant_api_keys
		 WHERE id = $1`,
		keyID,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &verificationResult, &verifiedAt, &modelsJSON, &verificationMsg)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if verificationResult.Valid {
		apiKey.VerificationResult = verificationResult.String
	}
	if verificationMsg.Valid {
		apiKey.VerificationMsg = verificationMsg.String
	}
	if verifiedAt.Valid {
		t := verifiedAt.Time
		apiKey.VerifiedAt = &t
	}
	if len(modelsJSON) > 0 {
		_ = json.Unmarshal(modelsJSON, &apiKey.ModelsSupported)
	}

	validator := services.GetAPIKeyValidator()
	history, err := validator.GetVerificationHistory(keyID, 10)
	if err != nil {
		logger.LogError(context.Background(), "admin_byok_routing", "Failed to get verification history", err, map[string]interface{}{
			"api_key_id": keyID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
		"history": history,
	})
}
