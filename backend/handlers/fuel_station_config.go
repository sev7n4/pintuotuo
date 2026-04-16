package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type fuelStationConfigRequest struct {
	PageTitle    string `json:"page_title"`
	PageSubtitle string `json:"page_subtitle"`
	RuleText     string `json:"rule_text"`
	Sections     []struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Badge       string `json:"badge"`
		SortOrder   int    `json:"sort_order"`
		Status      string `json:"status"`
		Tiers       []struct {
			Label string `json:"label"`
			SKUID int    `json:"sku_id"`
		} `json:"tiers"`
	} `json:"sections"`
}

type fuelStationTemplateLibraryRequest struct {
	Templates []struct {
		Key         string                   `json:"key"`
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Payload     fuelStationConfigRequest `json:"payload"`
	} `json:"templates"`
}

func GetFuelStationConfig(c *gin.Context) {
	cfg, err := services.GetFuelStationConfig(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cfg})
}

func GetAdminFuelStationConfig(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	GetFuelStationConfig(c)
}

func UpdateAdminFuelStationConfig(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	var req fuelStationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	cfg := services.FuelStationConfig{
		PageTitle:    req.PageTitle,
		PageSubtitle: req.PageSubtitle,
		RuleText:     req.RuleText,
		Sections:     make([]services.FuelStationSectionConfig, 0, len(req.Sections)),
	}
	for _, s := range req.Sections {
		sec := services.FuelStationSectionConfig{
			Code:        s.Code,
			Name:        s.Name,
			Description: s.Description,
			Badge:       s.Badge,
			SortOrder:   s.SortOrder,
			Status:      s.Status,
			Tiers:       make([]services.FuelStationTierConfig, 0, len(s.Tiers)),
		}
		for _, t := range s.Tiers {
			sec.Tiers = append(sec.Tiers, services.FuelStationTierConfig{
				Label: t.Label,
				SKUID: t.SKUID,
			})
		}
		cfg.Sections = append(cfg.Sections, sec)
	}
	out, err := services.UpdateFuelStationConfig(c.Request.Context(), cfg)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func GetAdminFuelStationTemplateLibrary(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	out, err := services.GetFuelStationTemplateLibrary(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func UpdateAdminFuelStationTemplateLibrary(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	var req fuelStationTemplateLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	templates := make([]services.FuelStationTemplate, 0, len(req.Templates))
	for _, t := range req.Templates {
		cfg := services.FuelStationConfig{
			PageTitle:    t.Payload.PageTitle,
			PageSubtitle: t.Payload.PageSubtitle,
			RuleText:     t.Payload.RuleText,
			Sections:     make([]services.FuelStationSectionConfig, 0, len(t.Payload.Sections)),
		}
		for _, s := range t.Payload.Sections {
			sec := services.FuelStationSectionConfig{
				Code:        s.Code,
				Name:        s.Name,
				Description: s.Description,
				Badge:       s.Badge,
				SortOrder:   s.SortOrder,
				Status:      s.Status,
				Tiers:       make([]services.FuelStationTierConfig, 0, len(s.Tiers)),
			}
			for _, tier := range s.Tiers {
				sec.Tiers = append(sec.Tiers, services.FuelStationTierConfig{
					Label: tier.Label,
					SKUID: tier.SKUID,
				})
			}
			cfg.Sections = append(cfg.Sections, sec)
		}
		templates = append(templates, services.FuelStationTemplate{
			Key:         t.Key,
			Name:        t.Name,
			Description: t.Description,
			Payload:     cfg,
		})
	}
	out, err := services.UpdateFuelStationTemplateLibrary(c.Request.Context(), templates)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
