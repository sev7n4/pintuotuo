package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pintuotuo/backend/apperrors"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"

	"github.com/gin-gonic/gin"
)

func AdminGetBillings(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	merchantIDStr := c.Query("merchant_id")
	provider := c.Query("provider")
	model := c.Query("model")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	filter := &services.BillingFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}

	if merchantIDStr != "" {
		if merchantID, err := strconv.Atoi(merchantIDStr); err == nil {
			filter.MerchantID = &merchantID
		}
	}

	if provider != "" {
		filter.Provider = &provider
	}

	if model != "" {
		filter.Model = &model
	}

	if startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endOfDay
		}
	}

	billingService := services.NewBillingService(db)
	billings, total, err := billingService.GetMerchantBillings(filter)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"BILLINGS_QUERY_FAILED",
			"Failed to query billings",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"billings": billings,
		"total":    total,
		"page":     filter.Page,
		"page_size": filter.PageSize,
	})
}

func AdminGetBillingStats(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	merchantIDStr := c.Query("merchant_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	filter := &services.BillingFilter{}

	if merchantIDStr != "" {
		if merchantID, err := strconv.Atoi(merchantIDStr); err == nil {
			filter.MerchantID = &merchantID
		}
	}

	if startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endOfDay
		}
	}

	billingService := services.NewBillingService(db)
	stats, err := billingService.GetBillingStats(filter)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"BILLING_STATS_FAILED",
			"Failed to get billing stats",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, stats)
}

func AdminGetUserBillings(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	userIDStr := c.Query("user_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	filter := &services.BillingFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}

	if userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}

	if startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endOfDay
		}
	}

	billingService := services.NewBillingService(db)
	billings, total, err := billingService.GetUserBillings(filter)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_BILLINGS_QUERY_FAILED",
			"Failed to query user billings",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"billings":  billings,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

func MerchantGetBillings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrUnauthorized)
		return
	}

	provider := c.Query("provider")
	model := c.Query("model")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userID).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	filter := &services.BillingFilter{
		MerchantID: &merchantID,
		Page:       1,
		PageSize:   20,
	}

	if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}

	if provider != "" {
		filter.Provider = &provider
	}

	if model != "" {
		filter.Model = &model
	}

	if startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endOfDay
		}
	}

	billingService := services.NewBillingService(db)
	billings, total, err := billingService.GetMerchantBillings(filter)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"BILLINGS_QUERY_FAILED",
			"Failed to query billings",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"billings":  billings,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}
