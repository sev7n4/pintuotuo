package order

import (
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

// Order service errors
var (
	ErrOrderNotFound = &apperrors.AppError{
		Code:    "ORDER_NOT_FOUND",
		Message: "Order not found",
		Status:  http.StatusNotFound,
	}

	ErrProductNotFound = &apperrors.AppError{
		Code:    "PRODUCT_NOT_FOUND",
		Message: "Product not found",
		Status:  http.StatusNotFound,
	}

	ErrInsufficientStock = &apperrors.AppError{
		Code:    "INSUFFICIENT_STOCK",
		Message: "Insufficient product stock",
		Status:  http.StatusConflict,
	}

	ErrInvalidQuantity = &apperrors.AppError{
		Code:    "INVALID_QUANTITY",
		Message: "Quantity must be greater than 0",
		Status:  http.StatusBadRequest,
	}

	ErrCannotCancelOrder = &apperrors.AppError{
		Code:    "CANNOT_CANCEL_ORDER",
		Message: "Cannot cancel order in its current status",
		Status:  http.StatusConflict,
	}

	ErrOrderAlreadyPaid = &apperrors.AppError{
		Code:    "ORDER_ALREADY_PAID",
		Message: "Order has already been paid",
		Status:  http.StatusConflict,
	}

	ErrNotOrderOwner = &apperrors.AppError{
		Code:    "NOT_ORDER_OWNER",
		Message: "You do not own this order",
		Status:  http.StatusForbidden,
	}

	ErrInvalidStatus = &apperrors.AppError{
		Code:    "INVALID_STATUS",
		Message: "Invalid order status",
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
		"ORDER_SERVICE_ERROR",
		"Order service operation failed: "+serviceName+"."+operation,
		http.StatusInternalServerError,
		err,
	)
}
