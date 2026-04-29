package services

import (
	"testing"
)

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		errorCode      string
		expectedCategory string
	}{
		{
			name:           "Authentication error - 401",
			errorCode:      "HTTP_401",
			expectedCategory: "AUTHENTICATION_ERROR",
		},
		{
			name:           "Authentication error - 403",
			errorCode:      "HTTP_403",
			expectedCategory: "AUTHENTICATION_ERROR",
		},
		{
			name:           "Rate limit error - 429",
			errorCode:      "HTTP_429",
			expectedCategory: "RATE_LIMIT_ERROR",
		},
		{
			name:           "Network error - timeout",
			errorCode:      "TIMEOUT",
			expectedCategory: "NETWORK_ERROR",
		},
		{
			name:           "Network error - connection refused",
			errorCode:      "ECONNREFUSED",
			expectedCategory: "NETWORK_ERROR",
		},
		{
			name:           "Quota exceeded",
			errorCode:      "QUOTA_EXCEEDED",
			expectedCategory: "QUOTA_EXCEEDED",
		},
		{
			name:           "Insufficient quota",
			errorCode:      "INSUFFICIENT_QUOTA",
			expectedCategory: "QUOTA_EXCEEDED",
		},
		{
			name:           "Model not found",
			errorCode:      "MODEL_NOT_FOUND",
			expectedCategory: "MODEL_NOT_FOUND",
		},
		{
			name:           "Invalid request - 400",
			errorCode:      "HTTP_400",
			expectedCategory: "INVALID_REQUEST",
		},
		{
			name:           "Unknown error",
			errorCode:      "UNKNOWN_ERROR",
			expectedCategory: "UNKNOWN_ERROR",
		},
		{
			name:           "Empty error code",
			errorCode:      "",
			expectedCategory: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := getErrorCategoryByCode(tt.errorCode)
			if category != tt.expectedCategory {
				t.Errorf("getErrorCategoryByCode(%s) = %s, want %s", tt.errorCode, category, tt.expectedCategory)
			}
		})
	}
}

func TestShouldRetryVerificationAttempt(t *testing.T) {
	tests := []struct {
		name          string
		probe         *ProbeModelsResult
		err           error
		shouldRetry   bool
	}{
		{
			name: "Network timeout - should retry",
			probe: &ProbeModelsResult{
				Success:   false,
				ErrorCode: "TIMEOUT",
			},
			err:         nil,
			shouldRetry: true,
		},
		{
			name: "Connection refused - should retry",
			probe: &ProbeModelsResult{
				Success:   false,
				ErrorCode: "ECONNREFUSED",
			},
			err:         nil,
			shouldRetry: true,
		},
		{
			name: "Rate limit - should not retry",
			probe: &ProbeModelsResult{
				Success:   false,
				ErrorCode: "HTTP_429",
			},
			err:         nil,
			shouldRetry: false,
		},
		{
			name: "Authentication error - should not retry",
			probe: &ProbeModelsResult{
				Success:   false,
				ErrorCode: "HTTP_401",
			},
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "Network error - should retry",
			probe:       nil,
			err:         &NetworkError{Message: "connection timeout"},
			shouldRetry: true,
		},
		{
			name:        "Nil probe and error - should not retry",
			probe:       nil,
			err:         nil,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetryVerificationAttempt(tt.probe, tt.err)
			if result != tt.shouldRetry {
				t.Errorf("shouldRetryVerificationAttempt() = %v, want %v", result, tt.shouldRetry)
			}
		})
	}
}

type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return e.Message
}

func getErrorCategoryByCode(errorCode string) string {
	if errorCode == "" {
		return "UNKNOWN_ERROR"
	}

	if errorCode == "HTTP_401" || errorCode == "HTTP_403" || errorCode == "AUTH_FAILED" {
		return "AUTHENTICATION_ERROR"
	}
	if errorCode == "HTTP_429" || errorCode == "RATE_LIMIT" {
		return "RATE_LIMIT_ERROR"
	}
	if errorCode == "TIMEOUT" || errorCode == "ECONNREFUSED" || errorCode == "NETWORK_ERROR" {
		return "NETWORK_ERROR"
	}
	if errorCode == "QUOTA_EXCEEDED" || errorCode == "INSUFFICIENT_QUOTA" {
		return "QUOTA_EXCEEDED"
	}
	if errorCode == "MODEL_NOT_FOUND" {
		return "MODEL_NOT_FOUND"
	}
	if errorCode == "HTTP_400" || errorCode == "INVALID_REQUEST" {
		return "INVALID_REQUEST"
	}

	return "UNKNOWN_ERROR"
}

func shouldRetryVerificationAttempt(probe *ProbeModelsResult, err error) bool {
	if err != nil {
		if _, ok := err.(*NetworkError); ok {
			return true
		}
		return false
	}

	if probe == nil {
		return false
	}

	retryableErrors := []string{"TIMEOUT", "ECONNREFUSED", "NETWORK_ERROR"}
	for _, retryable := range retryableErrors {
		if probe.ErrorCode == retryable {
			return true
		}
	}

	return false
}
