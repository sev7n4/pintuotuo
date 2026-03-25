package analytics

import (
  "net/http"

  apperrors "github.com/pintuotuo/backend/errors"
)

// Analytics service errors
var (
  ErrUserNotFound = &apperrors.AppError{
    Code:    "USER_NOT_FOUND",
    Message: "User not found for analytics",
    Status:  http.StatusNotFound,
  }

  ErrInvalidDateRange = &apperrors.AppError{
    Code:    "INVALID_DATE_RANGE",
    Message: "Invalid date range for analytics query",
    Status:  http.StatusBadRequest,
  }

  ErrInvalidPeriod = &apperrors.AppError{
    Code:    "INVALID_PERIOD",
    Message: "Invalid period type (must be: daily, weekly, monthly)",
    Status:  http.StatusBadRequest,
  }

  ErrInvalidLimit = &apperrors.AppError{
    Code:    "INVALID_LIMIT",
    Message: "Invalid limit parameter (must be > 0 and <= 500)",
    Status:  http.StatusBadRequest,
  }

  ErrMerchantNotFound = &apperrors.AppError{
    Code:    "MERCHANT_NOT_FOUND",
    Message: "Merchant not found for analytics",
    Status:  http.StatusNotFound,
  }

  ErrProductNotFound = &apperrors.AppError{
    Code:    "PRODUCT_NOT_FOUND",
    Message: "Product not found for analytics",
    Status:  http.StatusNotFound,
  }

  ErrNoDataAvailable = &apperrors.AppError{
    Code:    "NO_DATA_AVAILABLE",
    Message: "No analytics data available for the specified period",
    Status:  http.StatusNotFound,
  }

  ErrAnalyticsQueryFailed = &apperrors.AppError{
    Code:    "ANALYTICS_QUERY_FAILED",
    Message: "Failed to retrieve analytics data",
    Status:  http.StatusInternalServerError,
  }

  ErrInvalidMetricsQuery = &apperrors.AppError{
    Code:    "INVALID_METRICS_QUERY",
    Message: "Invalid metrics query parameters",
    Status:  http.StatusBadRequest,
  }
)

// wrapError wraps service errors as AppErrors
func wrapError(serviceName, operation string, err error) *apperrors.AppError {
  if appErr, ok := err.(*apperrors.AppError); ok {
    return appErr
  }

  return apperrors.NewAppError(
    "ANALYTICS_SERVICE_ERROR",
    "Analytics service operation failed: "+serviceName+"."+operation,
    http.StatusInternalServerError,
    err,
  )
}
