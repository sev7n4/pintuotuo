package group

import (
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

// Group service errors
var (
	ErrGroupNotFound = &apperrors.AppError{
		Code:    "GROUP_NOT_FOUND",
		Message: "Group not found",
		Status:  http.StatusNotFound,
	}

	ErrGroupFull = &apperrors.AppError{
		Code:    "GROUP_FULL",
		Message: "Group has reached target member count",
		Status:  http.StatusConflict,
	}

	ErrGroupExpired = &apperrors.AppError{
		Code:    "GROUP_EXPIRED",
		Message: "Group deadline has passed",
		Status:  http.StatusConflict,
	}

	ErrGroupInactive = &apperrors.AppError{
		Code:    "GROUP_INACTIVE",
		Message: "Group is not active",
		Status:  http.StatusConflict,
	}

	ErrAlreadyInGroup = &apperrors.AppError{
		Code:    "ALREADY_IN_GROUP",
		Message: "User is already in this group",
		Status:  http.StatusConflict,
	}

	ErrInvalidTargetCount = &apperrors.AppError{
		Code:    "INVALID_TARGET_COUNT",
		Message: "Target count must be greater than 0",
		Status:  http.StatusBadRequest,
	}

	ErrInvalidDeadline = &apperrors.AppError{
		Code:    "INVALID_DEADLINE",
		Message: "Deadline must be in the future",
		Status:  http.StatusBadRequest,
	}

	ErrProductNotFound = &apperrors.AppError{
		Code:    "PRODUCT_NOT_FOUND",
		Message: "Product not found",
		Status:  http.StatusNotFound,
	}

	ErrCannotCancelGroup = &apperrors.AppError{
		Code:    "CANNOT_CANCEL_GROUP",
		Message: "Cannot cancel group in its current status",
		Status:  http.StatusConflict,
	}

	ErrNotGroupCreator = &apperrors.AppError{
		Code:    "NOT_GROUP_CREATOR",
		Message: "Only group creator can cancel the group",
		Status:  http.StatusForbidden,
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
		"GROUP_SERVICE_ERROR",
		"Group service operation failed: "+serviceName+"."+operation,
		http.StatusInternalServerError,
		err,
	)
}
