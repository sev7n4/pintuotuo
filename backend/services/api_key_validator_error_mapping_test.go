package services

import (
	"testing"
)

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name             string
		errorCode        string
		expectedCategory string
	}{
		{
			name:             "Authentication error - 401",
			errorCode:        "HTTP_401",
			expectedCategory: "AUTHENTICATION_ERROR",
		},
		{
			name:             "Authentication error - 403",
			errorCode:        "HTTP_403",
			expectedCategory: "AUTHENTICATION_ERROR",
		},
		{
			name:             "Rate limit error - 429",
			errorCode:        "HTTP_429",
			expectedCategory: "RATE_LIMIT_ERROR",
		},
		{
			name:             "Network error - timeout",
			errorCode:        "TIMEOUT",
			expectedCategory: "NETWORK_ERROR",
		},
		{
			name:             "Network error - connection refused",
			errorCode:        "ECONNREFUSED",
			expectedCategory: "NETWORK_ERROR",
		},
		{
			name:             "Quota exceeded",
			errorCode:        "QUOTA_EXCEEDED",
			expectedCategory: "QUOTA_EXCEEDED",
		},
		{
			name:             "Insufficient quota",
			errorCode:        "INSUFFICIENT_QUOTA",
			expectedCategory: "QUOTA_EXCEEDED",
		},
		{
			name:             "Model not found",
			errorCode:        "MODEL_NOT_FOUND",
			expectedCategory: "MODEL_NOT_FOUND",
		},
		{
			name:             "Invalid request - 400",
			errorCode:        "HTTP_400",
			expectedCategory: "INVALID_REQUEST",
		},
		{
			name:             "Unknown error",
			errorCode:        "UNKNOWN_ERROR",
			expectedCategory: "UNKNOWN_ERROR",
		},
		{
			name:             "Empty error code",
			errorCode:        "",
			expectedCategory: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
