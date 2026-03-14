package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppErrorImplementsError(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Status:  http.StatusBadRequest,
	}

	assert.Implements(t, (*error)(nil), err)
	assert.Equal(t, "Test error message", err.Error())
}

func TestNewAppError(t *testing.T) {
	internalErr := assert.AnError
	appErr := NewAppError("TEST_CODE", "Test message", http.StatusBadRequest, internalErr)

	assert.NotNil(t, appErr)
	assert.Equal(t, "TEST_CODE", appErr.Code)
	assert.Equal(t, "Test message", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.Status)
	assert.Equal(t, internalErr, appErr.Internal)
}

func TestNewAppErrorWithDetails(t *testing.T) {
	details := map[string]interface{}{"field": "value"}
	appErr := NewAppErrorWithDetails("TEST_CODE", "Test message", http.StatusBadRequest, nil, details)

	assert.NotNil(t, appErr)
	assert.Equal(t, details, appErr.Details)
}

func TestIsAppError(t *testing.T) {
	appErr := &AppError{Code: "TEST"}
	regularErr := assert.AnError

	assert.True(t, IsAppError(appErr))
	assert.False(t, IsAppError(regularErr))
}

func TestGetAppError(t *testing.T) {
	appErr := &AppError{Code: "TEST", Status: http.StatusBadRequest}
	retrievedErr := GetAppError(appErr)

	assert.Equal(t, appErr, retrievedErr)
}

func TestGetAppErrorWithRegularError(t *testing.T) {
	regularErr := assert.AnError
	retrievedErr := GetAppError(regularErr)

	assert.Equal(t, ErrInternalServer.Code, retrievedErr.Code)
	assert.Equal(t, ErrInternalServer.Status, retrievedErr.Status)
	assert.Equal(t, regularErr, retrievedErr.Internal)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    *AppError
		code   string
		status int
	}{
		{"InvalidCredentials", ErrInvalidCredentials, "INVALID_CREDENTIALS", http.StatusUnauthorized},
		{"MissingToken", ErrMissingToken, "MISSING_TOKEN", http.StatusUnauthorized},
		{"UserNotFound", ErrUserNotFound, "USER_NOT_FOUND", http.StatusNotFound},
		{"ProductNotFound", ErrProductNotFound, "PRODUCT_NOT_FOUND", http.StatusNotFound},
		{"InsufficientStock", ErrInsufficientStock, "INSUFFICIENT_STOCK", http.StatusConflict},
		{"OrderNotFound", ErrOrderNotFound, "ORDER_NOT_FOUND", http.StatusNotFound},
		{"GroupNotFound", ErrGroupNotFound, "GROUP_NOT_FOUND", http.StatusNotFound},
		{"PaymentNotFound", ErrPaymentNotFound, "PAYMENT_NOT_FOUND", http.StatusNotFound},
		{"TokenNotFound", ErrTokenNotFound, "TOKEN_NOT_FOUND", http.StatusNotFound},
		{"Forbidden", ErrForbidden, "FORBIDDEN", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, tt.status, tt.err.Status)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

func TestErrorStatusCodes(t *testing.T) {
	// Verify BadRequest errors
	badRequestErrors := []*AppError{
		ErrInvalidEmail,
		ErrPasswordTooShort,
		ErrInvalidProductData,
		ErrInvalidOrderData,
		ErrInvalidGroupData,
		ErrInvalidPaymentMethod,
		ErrInvalidRequest,
	}
	for _, err := range badRequestErrors {
		assert.Equal(t, http.StatusBadRequest, err.Status, "Error %s should be BadRequest", err.Code)
	}

	// Verify Unauthorized errors
	unauthorizedErrors := []*AppError{
		ErrInvalidCredentials,
		ErrMissingToken,
		ErrInvalidToken,
		ErrInvalidAPIKey,
	}
	for _, err := range unauthorizedErrors {
		assert.Equal(t, http.StatusUnauthorized, err.Status, "Error %s should be Unauthorized", err.Code)
	}

	// Verify NotFound errors
	notFoundErrors := []*AppError{
		ErrUserNotFound,
		ErrProductNotFound,
		ErrOrderNotFound,
		ErrGroupNotFound,
		ErrPaymentNotFound,
		ErrTokenNotFound,
		ErrAPIKeyNotFound,
	}
	for _, err := range notFoundErrors {
		assert.Equal(t, http.StatusNotFound, err.Status, "Error %s should be NotFound", err.Code)
	}

	// Verify Conflict errors
	conflictErrors := []*AppError{
		ErrUserAlreadyExists,
		ErrProductInactive,
		ErrInsufficientStock,
		ErrOrderAlreadyPaid,
		ErrCannotCancelOrder,
		ErrGroupFull,
		ErrGroupExpired,
		ErrAlreadyInGroup,
		ErrPaymentAlreadyProcessed,
		ErrInsufficientBalance,
	}
	for _, err := range conflictErrors {
		assert.Equal(t, http.StatusConflict, err.Status, "Error %s should be Conflict", err.Code)
	}
}
