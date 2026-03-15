package handlers

import (
  "context"
  "log"
  "net/http"
  "os"
  "strconv"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/pintuotuo/backend/config"
  apperrors "github.com/pintuotuo/backend/errors"
  "github.com/pintuotuo/backend/middleware"
  "github.com/pintuotuo/backend/services/analytics"
)

// Initialize analytics service
var analyticsService analytics.Service

func initAnalyticsService() {
  if analyticsService == nil {
    logger := log.New(os.Stderr, "[AnalyticsHandler] ", log.LstdFlags)
    analyticsService = analytics.NewService(config.GetDB(), logger)
  }
}

// GetUserConsumption retrieves consumption summary for authenticated user
// GET /v1/analytics/consumption
func GetUserConsumption(c *gin.Context) {
  initAnalyticsService()

  userID, exists := c.Get("user_id")
  if !exists {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  userIDInt, ok := userID.(int)
  if !ok {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  // Parse date parameters
  startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
  endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

  startDate, err := time.Parse("2006-01-02", startDateStr)
  if err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  endDate, err := time.Parse("2006-01-02", endDateStr)
  if err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  ctx := context.Background()
  summary, err := analyticsService.GetUserConsumption(ctx, &analytics.ConsumptionAnalyticsRequest{
    UserID:    userIDInt,
    StartDate: startDate,
    EndDate:   endDate,
  })

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, summary)
}

// GetUserSpendingPattern retrieves spending pattern analysis for user
// GET /v1/analytics/spending-pattern
func GetUserSpendingPattern(c *gin.Context) {
  initAnalyticsService()

  userID, exists := c.Get("user_id")
  if !exists {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  userIDInt, ok := userID.(int)
  if !ok {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  ctx := context.Background()
  pattern, err := analyticsService.GetSpendingPattern(ctx, userIDInt)

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, pattern)
}

// GetRevenueAnalytics retrieves revenue data for merchant or product
// GET /v1/analytics/revenue
func GetRevenueAnalytics(c *gin.Context) {
  initAnalyticsService()

  startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
  endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

  startDate, err := time.Parse("2006-01-02", startDateStr)
  if err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  endDate, err := time.Parse("2006-01-02", endDateStr)
  if err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  // Parse merchant_id or product_id
  var merchantID, productID *int

  if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
    if mID, err := strconv.Atoi(merchantIDStr); err == nil {
      merchantID = &mID
    }
  }

  if productIDStr := c.Query("product_id"); productIDStr != "" {
    if pID, err := strconv.Atoi(productIDStr); err == nil {
      productID = &pID
    }
  }

  if merchantID == nil && productID == nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  ctx := context.Background()
  data, err := analyticsService.GetRevenueData(ctx, &analytics.RevenueAnalyticsRequest{
    MerchantID: merchantID,
    ProductID:  productID,
    StartDate:  startDate,
    EndDate:    endDate,
  })

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, data)
}

// GetTopSpenders retrieves top spending users
// GET /v1/analytics/top-spenders
func GetTopSpenders(c *gin.Context) {
  initAnalyticsService()

  limit := 10
  if limitStr := c.Query("limit"); limitStr != "" {
    if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
      limit = l
    }
  }

  period := c.DefaultQuery("period", "last_30_days")
  minSpend := 0.0

  if minStr := c.Query("min_spend"); minStr != "" {
    if min, err := strconv.ParseFloat(minStr, 64); err == nil && min >= 0 {
      minSpend = min
    }
  }

  ctx := context.Background()
  spenders, err := analyticsService.GetTopSpenders(ctx, &analytics.TopSpendersRequest{
    Limit:    limit,
    Period:   period,
    MinSpend: minSpend,
  })

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "spenders": spenders,
    "count":    len(spenders),
  })
}

// GetPlatformMetrics retrieves overall platform metrics
// GET /v1/analytics/metrics
func GetPlatformMetrics(c *gin.Context) {
  initAnalyticsService()

  ctx := context.Background()
  metrics, err := analyticsService.GetPlatformMetrics(ctx)

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, metrics)
}

// GetConsumptionHistory retrieves detailed consumption records
// GET /v1/analytics/consumption-history
func GetConsumptionHistory(c *gin.Context) {
  initAnalyticsService()

  userID, exists := c.Get("user_id")
  if !exists {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  userIDInt, ok := userID.(int)
  if !ok {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  // Parse pagination
  page := 1
  if pageStr := c.Query("page"); pageStr != "" {
    if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
      page = p
    }
  }

  pageSize := 20
  if sizeStr := c.Query("page_size"); sizeStr != "" {
    if ps, err := strconv.Atoi(sizeStr); err == nil && ps > 0 && ps <= 100 {
      pageSize = ps
    }
  }

  // Parse dates
  startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
  endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

  startDate, err := time.Parse("2006-01-02", startDateStr)
  if err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  endDate, err := time.Parse("2006-01-02", endDateStr)
  if err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  offset := (page - 1) * pageSize
  ctx := context.Background()
  records, err := analyticsService.GetUserConsumptionRecords(ctx, userIDInt, startDate, endDate, pageSize, offset)

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "records":    records,
    "page":       page,
    "page_size":  pageSize,
    "count":      len(records),
  })
}
