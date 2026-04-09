package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type healthSchedulerSettingsRequest struct {
	Enabled         bool `json:"health_scheduler_enabled"`
	IntervalSeconds int  `json:"health_scheduler_interval_seconds"`
	Batch           int  `json:"health_scheduler_batch"`
}

// GetAdminPlatformSettings GET /admin/platform-settings
func GetAdminPlatformSettings(c *gin.Context) {
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

	ctx := c.Request.Context()
	if err := services.ReloadPlatformSettingsCache(ctx); err != nil {
		log.Printf("admin platform-settings GET: reload: %v", err)
	}

	cfg := services.GetHealthSchedulerPlatformConfig()
	c.JSON(http.StatusOK, gin.H{
		"code":                              0,
		"message":                           "success",
		"health_scheduler_enabled":          cfg.Enabled,
		"health_scheduler_interval_seconds": cfg.IntervalSeconds,
		"health_scheduler_batch":            cfg.Batch,
		"limits":                            services.HealthSchedulerPlatformLimits(),
	})
}

// UpdateAdminPlatformSettings PUT /admin/platform-settings
func UpdateAdminPlatformSettings(c *gin.Context) {
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

	var req healthSchedulerSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	cfg := services.HealthSchedulerPlatformConfig{
		Enabled:         req.Enabled,
		IntervalSeconds: req.IntervalSeconds,
		Batch:           req.Batch,
	}

	ctx := c.Request.Context()
	if err := services.UpdateHealthSchedulerPlatformSettings(ctx, cfg); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SETTINGS_UPDATE_FAILED",
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	services.GetHealthScheduler().SignalReload()

	c.JSON(http.StatusOK, gin.H{
		"code":                              0,
		"message":                           "success",
		"health_scheduler_enabled":          services.GetHealthSchedulerPlatformConfig().Enabled,
		"health_scheduler_interval_seconds": services.GetHealthSchedulerPlatformConfig().IntervalSeconds,
		"health_scheduler_batch":            services.GetHealthSchedulerPlatformConfig().Batch,
		"limits":                            services.HealthSchedulerPlatformLimits(),
	})
}
