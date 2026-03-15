package token

import (
  "net/http"

  apperrors "github.com/pintuotuo/backend/errors"
)

// Token service errors
var (
  ErrInsufficientBalance = &apperrors.AppError{
    Code:    "INSUFFICIENT_BALANCE",
    Message: "Insufficient token balance for this operation",
    Status:  http.StatusConflict,
  }

  ErrTokenNotFound = &apperrors.AppError{
    Code:    "TOKEN_NOT_FOUND",
    Message: "Token record not found for user",
    Status:  http.StatusNotFound,
  }

  ErrInvalidAmount = &apperrors.AppError{
    Code:    "INVALID_AMOUNT",
    Message: "Token amount must be greater than zero",
    Status:  http.StatusBadRequest,
  }

  ErrInvalidUserID = &apperrors.AppError{
    Code:    "INVALID_USER_ID",
    Message: "Invalid user ID format",
    Status:  http.StatusBadRequest,
  }

  ErrTransferToSelf = &apperrors.AppError{
    Code:    "TRANSFER_TO_SELF",
    Message: "Cannot transfer tokens to yourself",
    Status:  http.StatusBadRequest,
  }

  ErrRecipientNotFound = &apperrors.AppError{
    Code:    "RECIPIENT_NOT_FOUND",
    Message: "Recipient user not found",
    Status:  http.StatusNotFound,
  }

  ErrTokenExpired = &apperrors.AppError{
    Code:    "TOKEN_EXPIRED",
    Message: "Token record has expired or been deleted",
    Status:  http.StatusGone,
  }

  ErrTransactionFailed = &apperrors.AppError{
    Code:    "TRANSACTION_FAILED",
    Message: "Database transaction failed",
    Status:  http.StatusInternalServerError,
  }

  ErrConcurrentOperation = &apperrors.AppError{
    Code:    "CONCURRENT_OPERATION",
    Message: "Concurrent operation detected, please retry",
    Status:  http.StatusConflict,
  }

  ErrInvalidReason = &apperrors.AppError{
    Code:    "INVALID_REASON",
    Message: "Transaction reason cannot be empty",
    Status:  http.StatusBadRequest,
  }

  ErrDuplicateTransaction = &apperrors.AppError{
    Code:    "DUPLICATE_TRANSACTION",
    Message: "This transaction has already been processed (idempotency)",
    Status:  http.StatusConflict,
  }

  ErrUserNotFound = &apperrors.AppError{
    Code:    "USER_NOT_FOUND",
    Message: "User not found in database",
    Status:  http.StatusNotFound,
  }

  ErrOrderNotFound = &apperrors.AppError{
    Code:    "ORDER_NOT_FOUND",
    Message: "Order not found in database",
    Status:  http.StatusNotFound,
  }

  ErrInvalidRangeParams = &apperrors.AppError{
    Code:    "INVALID_RANGE_PARAMS",
    Message: "Invalid date range parameters",
    Status:  http.StatusBadRequest,
  }
)

// wrapError wraps service errors as AppErrors
func wrapError(serviceName, operation string, err error) *apperrors.AppError {
  if appErr, ok := err.(*apperrors.AppError); ok {
    return appErr
  }

  return apperrors.NewAppError(
    "TOKEN_SERVICE_ERROR",
    "Token service operation failed: "+serviceName+"."+operation,
    http.StatusInternalServerError,
    err,
  )
}
