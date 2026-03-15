package user

import (
	"errors"
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

// UserService errors
var (
	ErrInvalidEmail = &apperrors.AppError{
		Code:    "INVALID_EMAIL",
		Message: "Invalid email format",
		Status:  http.StatusBadRequest,
	}

	ErrPasswordTooShort = &apperrors.AppError{
		Code:    "PASSWORD_TOO_SHORT",
		Message: "Password must be at least 6 characters",
		Status:  http.StatusBadRequest,
	}

	ErrNameTooShort = &apperrors.AppError{
		Code:    "NAME_TOO_SHORT",
		Message: "Name must be at least 2 characters",
		Status:  http.StatusBadRequest,
	}

	ErrUserAlreadyExists = &apperrors.AppError{
		Code:    "USER_ALREADY_EXISTS",
		Message: "User with this email already exists",
		Status:  http.StatusConflict,
	}

	ErrInvalidCredentials = &apperrors.AppError{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid email or password",
		Status:  http.StatusUnauthorized,
	}

	ErrUserNotFound = &apperrors.AppError{
		Code:    "USER_NOT_FOUND",
		Message: "User not found",
		Status:  http.StatusNotFound,
	}

	ErrInvalidToken = &apperrors.AppError{
		Code:    "INVALID_TOKEN",
		Message: "Invalid or expired token",
		Status:  http.StatusUnauthorized,
	}

	ErrTokenExpired = &apperrors.AppError{
		Code:    "TOKEN_EXPIRED",
		Message: "Token has expired",
		Status:  http.StatusUnauthorized,
	}

	ErrUserBanned = &apperrors.AppError{
		Code:    "USER_BANNED",
		Message: "User account has been banned",
		Status:  http.StatusForbidden,
	}

	ErrInvalidResetToken = &apperrors.AppError{
		Code:    "INVALID_RESET_TOKEN",
		Message: "Password reset token is invalid or expired",
		Status:  http.StatusBadRequest,
	}

	ErrCannotDeleteUser = &apperrors.AppError{
		Code:    "CANNOT_DELETE_USER",
		Message: "Cannot delete user",
		Status:  http.StatusInternalServerError,
	}
)

// WrapError wraps service errors as AppErrors
func wrapError(serviceName, operation string, err error) *apperrors.AppError {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return appErr
	}

	return apperrors.NewAppError(
		"USER_SERVICE_ERROR",
		"User service operation failed: "+serviceName+"."+operation,
		http.StatusInternalServerError,
		err,
	)
}

// IsUserNotFoundError checks if error is a user not found error
func IsUserNotFoundError(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}

// IsUserAlreadyExistsError checks if error is a user already exists error
func IsUserAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists)
}

// IsInvalidCredentialsError checks if error is an invalid credentials error
func IsInvalidCredentialsError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}
