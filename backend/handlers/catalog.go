package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

type UsageScenario struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status"`
	SPUCount    int    `json:"spu_count,omitempty"`
}

type ScenarioSPU struct {
	ID                int     `json:"id"`
	SPUCode           string  `json:"spu_code"`
	Name              string  `json:"name"`
	ModelProvider     string  `json:"model_provider"`
	ModelName         string  `json:"model_name"`
	ModelTier         string  `json:"model_tier"`
	ContextWindow     int     `json:"context_window"`
	BaseComputePoints float64 `json:"base_compute_points"`
	Description       string  `json:"description,omitempty"`
	ThumbnailURL      string  `json:"thumbnail_url,omitempty"`
	TotalSalesCount   int64   `json:"total_sales_count"`
	AverageRating     float64 `json:"average_rating,omitempty"`
	AvgLatencyMs      int     `json:"avg_latency_ms,omitempty"`
	AvailabilityRate  float64 `json:"availability_rate,omitempty"`
	IsPrimary         bool    `json:"is_primary"`
}

func GetScenarios(c *gin.Context) {
	ctx := context.Background()
	db := config.GetDB()

	query := `
		SELECT us.id, us.code, us.name, us.description, us.icon_url, us.sort_order, us.status,
			   COUNT(DISTINCT ss.spu_id) as spu_count
		FROM usage_scenarios us
		LEFT JOIN spu_scenarios ss ON us.id = ss.scenario_id
		LEFT JOIN spus s ON ss.spu_id = s.id AND s.status = 'active'
		WHERE us.status = 'active'
		GROUP BY us.id, us.code, us.name, us.description, us.icon_url, us.sort_order, us.status
		ORDER BY us.sort_order ASC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var scenarios []UsageScenario
	for rows.Next() {
		var s UsageScenario
		var description, iconURL sql.NullString
		err := rows.Scan(
			&s.ID, &s.Code, &s.Name, &description, &iconURL,
			&s.SortOrder, &s.Status, &s.SPUCount,
		)
		if err != nil {
			continue
		}
		if description.Valid {
			s.Description = description.String
		}
		if iconURL.Valid {
			s.IconURL = iconURL.String
		}
		scenarios = append(scenarios, s)
	}

	if scenarios == nil {
		scenarios = []UsageScenario{}
	}

	c.JSON(http.StatusOK, gin.H{
		"scenarios": scenarios,
	})
}

func GetSPUsByScenario(c *gin.Context) {
	scenarioCode := c.Param("scenario")
	if scenarioCode == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_SCENARIO",
			"Scenario code is required",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	sortBy := c.DefaultQuery("sort_by", "sales")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	ctx := context.Background()
	db := config.GetDB()

	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT s.id)
		FROM spus s
		JOIN spu_scenarios ss ON s.id = ss.spu_id
		JOIN usage_scenarios us ON ss.scenario_id = us.id
		WHERE us.code = $1 AND s.status = 'active'
	`
	err := db.QueryRowContext(ctx, countQuery, scenarioCode).Scan(&totalCount)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	offset := (pageNum - 1) * perPageNum

	orderClause := "s.total_sales_count DESC"
	switch sortBy {
	case "rating":
		orderClause = "s.average_rating DESC NULLS LAST"
	case "latency":
		orderClause = "s.avg_latency_ms ASC NULLS LAST"
	case "price":
		orderClause = "s.base_compute_points ASC"
	}

	query := `
		SELECT DISTINCT s.id, s.spu_code, s.name, s.model_provider, s.model_name,
			   s.model_tier, s.context_window, s.base_compute_points, s.description,
			   s.thumbnail_url, s.total_sales_count, COALESCE(s.average_rating, 0),
			   COALESCE(s.avg_latency_ms, 0), COALESCE(s.availability_rate, 99.9),
			   ss.is_primary
		FROM spus s
		JOIN spu_scenarios ss ON s.id = ss.spu_id
		JOIN usage_scenarios us ON ss.scenario_id = us.id
		WHERE us.code = $1 AND s.status = 'active'
		ORDER BY ` + orderClause + `
		LIMIT $2 OFFSET $3
	`

	rows, err := db.QueryContext(ctx, query, scenarioCode, perPageNum, offset)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var spus []ScenarioSPU
	for rows.Next() {
		var spu ScenarioSPU
		var description, thumbnailURL sql.NullString
		err := rows.Scan(
			&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName,
			&spu.ModelTier, &spu.ContextWindow, &spu.BaseComputePoints, &description,
			&thumbnailURL, &spu.TotalSalesCount, &spu.AverageRating,
			&spu.AvgLatencyMs, &spu.AvailabilityRate, &spu.IsPrimary,
		)
		if err != nil {
			continue
		}
		if description.Valid {
			spu.Description = description.String
		}
		if thumbnailURL.Valid {
			spu.ThumbnailURL = thumbnailURL.String
		}
		spus = append(spus, spu)
	}

	if spus == nil {
		spus = []ScenarioSPU{}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    totalCount,
		"page":     pageNum,
		"per_page": perPageNum,
		"spus":     spus,
	})
}

func GetSPUPerformance(c *gin.Context) {
	spuIDStr := c.Param("id")
	spuID, err := strconv.Atoi(spuIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_SPU_ID",
			"Invalid SPU ID",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	ctx := context.Background()
	db := config.GetDB()

	var performance struct {
		SPUID            int     `json:"spu_id"`
		AvgLatencyMs     int     `json:"avg_latency_ms"`
		AvailabilityRate float64 `json:"availability_rate"`
		LastHealthCheck  string  `json:"last_health_check_at,omitempty"`
	}

	query := `
		SELECT id, COALESCE(avg_latency_ms, 0), COALESCE(availability_rate, 99.9),
			   COALESCE(to_char(last_health_check_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '')
		FROM spus
		WHERE id = $1
	`

	var lastHealthCheck sql.NullString
	err = db.QueryRowContext(ctx, query, spuID).Scan(
		&performance.SPUID, &performance.AvgLatencyMs,
		&performance.AvailabilityRate, &lastHealthCheck,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"SPU_NOT_FOUND",
				"SPU not found",
				http.StatusNotFound,
				nil,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if lastHealthCheck.Valid {
		performance.LastHealthCheck = lastHealthCheck.String
	}

	c.JSON(http.StatusOK, performance)
}
