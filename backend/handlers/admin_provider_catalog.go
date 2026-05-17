package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/services"
)

// GetProviderCatalogGap compares provider_models with spus for onboarding / stale alerts.
func GetProviderCatalogGap(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	providerCode := c.Param("code")
	svc := services.NewProviderCatalogGapService()
	gap, err := svc.CompareProviderCatalog(c.Request.Context(), providerCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gap)
}

// CreateProviderSPUDrafts creates inactive SPU drafts from provider model ids.
func CreateProviderSPUDrafts(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	providerCode := c.Param("code")
	var req struct {
		ModelIDs []string `json:"model_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_ids is required"})
		return
	}
	svc := services.NewProviderCatalogGapService()
	created, err := svc.CreateSPUDraftsFromProviderModels(c.Request.Context(), providerCode, req.ModelIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cache.InvalidatePatterns(context.Background(), "spus:list:*")
	c.JSON(http.StatusOK, gin.H{
		"provider_code": providerCode,
		"results":       created,
	})
}

// SyncAllProviderModels triggers upstream sync for all active providers (manual).
func SyncAllProviderModels(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	svc := services.NewProviderCatalogGapService()
	results, err := svc.SyncAllActiveProviderModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "sync completed",
		"results": results,
	})
}
