package product

import (
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

// Product service errors
var (
	ErrInvalidProductName = &apperrors.AppError{
		Code:    "INVALID_PRODUCT_NAME",
		Message: "Product name is required",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidPrice = &apperrors.AppError{
		Code:    "INVALID_PRICE",
		Message: "Product price must be greater than 0",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidStock = &apperrors.AppError{
		Code:    "INVALID_STOCK",
		Message: "Product stock cannot be negative",
		Status:  http.StatusBadRequest,
	}

	ErrProductNotFound = &apperrors.AppError{
		Code:    "PRODUCT_NOT_FOUND",
		Message: "Product not found",
		Status:  http.StatusNotFound,
	}

	ErrProductInactive = &apperrors.AppError{
		Code:    "PRODUCT_INACTIVE",
		Message: "Product is not available for purchase",
		Status:  http.StatusConflict,
	}

	ErrInsufficientStock = &apperrors.AppError{
		Code:    "INSUFFICIENT_STOCK",
		Message: "Insufficient product stock",
		Status:  http.StatusConflict,
	}

	ErrMerchantOnly = &apperrors.AppError{
		Code:    "MERCHANT_ONLY",
		Message: "This action is only available to merchants",
		Status:  http.StatusForbidden,
	}

	ErrNotProductOwner = &apperrors.AppError{
		Code:    "NOT_PRODUCT_OWNER",
		Message: "You do not own this product",
		Status:  http.StatusForbidden,
	}

	ErrInvalidSearchQuery = &apperrors.AppError{
		Code:    "INVALID_SEARCH_QUERY",
		Message: "Search query is required",
		Status:  http.StatusBadRequest,
	}

	ErrDatabaseError = &apperrors.AppError{
		Code:    "DATABASE_ERROR",
		Message: "Database operation failed",
		Status:  http.StatusInternalServerError,
	}
)

// wrapError wraps service errors as AppErrors
func wrapError(serviceName, operation string, err error) *apperrors.AppError {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return appErr
	}

	return apperrors.NewAppError(
		"PRODUCT_SERVICE_ERROR",
		"Product service operation failed: "+serviceName+"."+operation,
		http.StatusInternalServerError,
		err,
	)
}
