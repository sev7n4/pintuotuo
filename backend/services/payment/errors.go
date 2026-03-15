package payment

import (
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

// Payment service errors
var (
	ErrPaymentNotFound = &apperrors.AppError{
		Code:    "PAYMENT_NOT_FOUND",
		Message: "Payment not found",
		Status:  http.StatusNotFound,
	}

	ErrOrderNotFound = &apperrors.AppError{
		Code:    "ORDER_NOT_FOUND",
		Message: "Order not found",
		Status:  http.StatusNotFound,
	}

	ErrPaymentAlreadyProcessed = &apperrors.AppError{
		Code:    "PAYMENT_ALREADY_PROCESSED",
		Message: "Payment has already been processed",
		Status:  http.StatusConflict,
	}

	ErrPaymentFailed = &apperrors.AppError{
		Code:    "PAYMENT_FAILED",
		Message: "Payment processing failed",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidPaymentMethod = &apperrors.AppError{
		Code:    "INVALID_PAYMENT_METHOD",
		Message: "Invalid payment method",
		Status:  http.StatusBadRequest,
	}

	ErrOrderAlreadyPaid = &apperrors.AppError{
		Code:    "ORDER_ALREADY_PAID",
		Message: "Order has already been paid",
		Status:  http.StatusConflict,
	}

	ErrOrderCancelled = &apperrors.AppError{
		Code:    "ORDER_CANCELLED",
		Message: "Order has been cancelled",
		Status:  http.StatusConflict,
	}

	ErrInsufficientAmount = &apperrors.AppError{
		Code:    "INSUFFICIENT_AMOUNT",
		Message: "Payment amount does not match order total",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidSignature = &apperrors.AppError{
		Code:    "INVALID_SIGNATURE",
		Message: "Payment webhook signature verification failed",
		Status:  http.StatusBadRequest,
	}

	ErrCannotRefundPendingPayment = &apperrors.AppError{
		Code:    "CANNOT_REFUND_PENDING_PAYMENT",
		Message: "Cannot refund a payment that is still pending",
		Status:  http.StatusConflict,
	}
)

// wrapError wraps service errors as AppErrors
func wrapError(serviceName, operation string, err error) *apperrors.AppError {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return appErr
	}

	return apperrors.NewAppError(
		"PAYMENT_SERVICE_ERROR",
		"Payment service operation failed: "+serviceName+"."+operation,
		http.StatusInternalServerError,
		err,
	)
}
