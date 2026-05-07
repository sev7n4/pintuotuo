package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	models "github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

func SyncProviderModels(c *gin.Context) {
	providerCode := c.Param("code")
	if providerCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider code is required"})
		return
	}

	syncService := services.NewProviderModelSyncService()
	syncedCount, err := syncService.SyncProviderModels(c.Request.Context(), providerCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Sync completed",
		"provider_code": providerCode,
		"synced_count":  syncedCount,
	})
}

func ListProviderModels(c *gin.Context) {
	providerCode := c.Query("provider")
	if providerCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider query parameter is required"})
		return
	}

	activeOnly := true
	if a := c.Query("active_only"); a != "" {
		activeOnly, _ = strconv.ParseBool(a)
	}

	syncService := services.NewProviderModelSyncService()
	items, err := syncService.ListProviderModels(c.Request.Context(), providerCode, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if items == nil {
		items = []models.ProviderModel{}
	}

	c.JSON(http.StatusOK, gin.H{
		"provider_code": providerCode,
		"models":        items,
		"count":         len(items),
	})
}
