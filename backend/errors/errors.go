package errors

import (
	"net/http"
)

// AppError represents a standardized application error
type AppError struct {
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	Status   int         `json:"-"`
	Internal error       `json:"-"`
	Details  interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(code, message string, status int, internal error) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Status:   status,
		Internal: internal,
	}
}

// NewAppErrorWithDetails creates a new AppError with details
func NewAppErrorWithDetails(code, message string, status int, internal error, details interface{}) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Status:   status,
		Internal: internal,
		Details:  details,
	}
}

// Common error definitions
var (
	// Authentication errors
	ErrInvalidCredentials = &AppError{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid email or password",
		Status:  http.StatusUnauthorized,
	}

	ErrMissingToken = &AppError{
		Code:    "MISSING_TOKEN",
		Message: "Missing authorization token",
		Status:  http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:    "INVALID_TOKEN",
		Message: "Invalid or expired token",
		Status:  http.StatusUnauthorized,
	}

	// User errors
	ErrUserNotFound = &AppError{
		Code:    "USER_NOT_FOUND",
		Message: "User not found",
		Status:  http.StatusNotFound,
	}

	ErrUserAlreadyExists = &AppError{
		Code:    "USER_ALREADY_EXISTS",
		Message: "User with this email already exists",
		Status:  http.StatusConflict,
	}

	ErrEmailRequired = &AppError{
		Code:    "EMAIL_REQUIRED",
		Message: "Email is required",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidEmail = &AppError{
		Code:    "INVALID_EMAIL",
		Message: "Invalid email format",
		Status:  http.StatusBadRequest,
	}

	ErrPasswordTooShort = &AppError{
		Code:    "PASSWORD_TOO_SHORT",
		Message: "Password must be at least 6 characters",
		Status:  http.StatusBadRequest,
	}

	// Product errors
	ErrProductNotFound = &AppError{
		Code:    "PRODUCT_NOT_FOUND",
		Message: "Product not found",
		Status:  http.StatusNotFound,
	}

	ErrProductInactive = &AppError{
		Code:    "PRODUCT_INACTIVE",
		Message: "Product is not available for purchase",
		Status:  http.StatusConflict,
	}

	ErrInsufficientStock = &AppError{
		Code:    "INSUFFICIENT_STOCK",
		Message: "Insufficient product stock",
		Status:  http.StatusConflict,
	}

	ErrInvalidProductData = &AppError{
		Code:    "INVALID_PRODUCT_DATA",
		Message: "Invalid product data",
		Status:  http.StatusBadRequest,
	}

	// Order errors
	ErrOrderNotFound = &AppError{
		Code:    "ORDER_NOT_FOUND",
		Message: "Order not found",
		Status:  http.StatusNotFound,
	}

	ErrOrderAlreadyPaid = &AppError{
		Code:    "ORDER_ALREADY_PAID",
		Message: "Order has already been paid",
		Status:  http.StatusConflict,
	}

	ErrCannotCancelOrder = &AppError{
		Code:    "CANNOT_CANCEL_ORDER",
		Message: "Cannot cancel order in its current status",
		Status:  http.StatusConflict,
	}

	ErrInvalidOrderData = &AppError{
		Code:    "INVALID_ORDER_DATA",
		Message: "Invalid order data",
		Status:  http.StatusBadRequest,
	}

	// Group errors
	ErrGroupNotFound = &AppError{
		Code:    "GROUP_NOT_FOUND",
		Message: "Group not found",
		Status:  http.StatusNotFound,
	}

	ErrGroupFull = &AppError{
		Code:    "GROUP_FULL",
		Message: "Group has reached target member count",
		Status:  http.StatusConflict,
	}

	ErrGroupExpired = &AppError{
		Code:    "GROUP_EXPIRED",
		Message: "Group deadline has passed",
		Status:  http.StatusConflict,
	}

	ErrAlreadyInGroup = &AppError{
		Code:    "ALREADY_IN_GROUP",
		Message: "User is already in this group",
		Status:  http.StatusConflict,
	}

	ErrInvalidGroupData = &AppError{
		Code:    "INVALID_GROUP_DATA",
		Message: "Invalid group data",
		Status:  http.StatusBadRequest,
	}

	// Payment errors
	ErrPaymentNotFound = &AppError{
		Code:    "PAYMENT_NOT_FOUND",
		Message: "Payment not found",
		Status:  http.StatusNotFound,
	}

	ErrPaymentAlreadyProcessed = &AppError{
		Code:    "PAYMENT_ALREADY_PROCESSED",
		Message: "Payment has already been processed",
		Status:  http.StatusConflict,
	}

	ErrPaymentFailed = &AppError{
		Code:    "PAYMENT_FAILED",
		Message: "Payment processing failed",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidPaymentMethod = &AppError{
		Code:    "INVALID_PAYMENT_METHOD",
		Message: "Invalid payment method",
		Status:  http.StatusBadRequest,
	}

	// Token errors
	ErrInsufficientBalance = &AppError{
		Code:    "INSUFFICIENT_BALANCE",
		Message: "Insufficient token balance",
		Status:  http.StatusConflict,
	}

	ErrTokenNotFound = &AppError{
		Code:    "TOKEN_NOT_FOUND",
		Message: "Token balance not found",
		Status:  http.StatusNotFound,
	}

	// API Key errors
	ErrAPIKeyNotFound = &AppError{
		Code:    "API_KEY_NOT_FOUND",
		Message: "API key not found",
		Status:  http.StatusNotFound,
	}

	ErrInvalidAPIKey = &AppError{
		Code:    "INVALID_API_KEY",
		Message: "Invalid API key",
		Status:  http.StatusUnauthorized,
	}

	// Settlement errors
	ErrSettlementNotFound = &AppError{
		Code:    "SETTLEMENT_NOT_FOUND",
		Message: "Settlement not found",
		Status:  http.StatusNotFound,
	}

	ErrSettlementAlreadyExists = &AppError{
		Code:    "SETTLEMENT_ALREADY_EXISTS",
		Message: "Settlement already exists for this period",
		Status:  http.StatusConflict,
	}

	ErrMerchantConfirmationRequired = &AppError{
		Code:    "MERCHANT_CONFIRMATION_REQUIRED",
		Message: "Merchant confirmation required before finance approval",
		Status:  http.StatusBadRequest,
	}

	ErrFinanceApprovalRequired = &AppError{
		Code:    "FINANCE_APPROVAL_REQUIRED",
		Message: "Finance approval required before marking as paid",
		Status:  http.StatusBadRequest,
	}

	ErrSettlementAlreadyApproved = &AppError{
		Code:    "SETTLEMENT_ALREADY_APPROVED",
		Message: "Settlement already approved by finance",
		Status:  http.StatusConflict,
	}

	ErrDisputeNotFound = &AppError{
		Code:    "DISPUTE_NOT_FOUND",
		Message: "Dispute not found",
		Status:  http.StatusNotFound,
	}

	ErrDisputeAlreadyResolved = &AppError{
		Code:    "DISPUTE_ALREADY_RESOLVED",
		Message: "Dispute has already been resolved",
		Status:  http.StatusConflict,
	}

	ErrUnauthorizedSettlementAccess = &AppError{
		Code:    "UNAUTHORIZED_SETTLEMENT_ACCESS",
		Message: "You do not have permission to access this settlement",
		Status:  http.StatusForbidden,
	}

	// Permission errors
	ErrForbidden = &AppError{
		Code:    "FORBIDDEN",
		Message: "You do not have permission to perform this action",
		Status:  http.StatusForbidden,
	}

	ErrMerchantOnly = &AppError{
		Code:    "MERCHANT_ONLY",
		Message: "This action is only available to merchants",
		Status:  http.StatusForbidden,
	}

	// Server errors
	ErrInternalServer = &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}

	ErrDatabaseError = &AppError{
		Code:    "DATABASE_ERROR",
		Message: "Database operation failed",
		Status:  http.StatusInternalServerError,
	}

	ErrInvalidRequest = &AppError{
		Code:    "INVALID_REQUEST",
		Message: "Invalid request",
		Status:  http.StatusBadRequest,
	}
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError returns an AppError if the error is one, otherwise returns ErrInternalServer
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return &AppError{
		Code:     ErrInternalServer.Code,
		Message:  ErrInternalServer.Message,
		Status:   ErrInternalServer.Status,
		Internal: err,
	}
}
