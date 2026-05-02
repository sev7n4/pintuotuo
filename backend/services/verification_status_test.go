package services

import "testing"

func TestMapErrorCategoryToVerificationStatus(t *testing.T) {
	tests := []struct {
		name          string
		errorCategory string
		currentStatus string
		want          string
	}{
		{
			name:          "AUTH_INVALID_KEY maps to invalid",
			errorCategory: "AUTH_INVALID_KEY",
			currentStatus: "verified",
			want:          "invalid",
		},
		{
			name:          "AUTH_PERMISSION_DENIED maps to invalid",
			errorCategory: "AUTH_PERMISSION_DENIED",
			currentStatus: "verified",
			want:          "invalid",
		},
		{
			name:          "QUOTA_INSUFFICIENT maps to suspend",
			errorCategory: "QUOTA_INSUFFICIENT",
			currentStatus: "verified",
			want:          "suspend",
		},
		{
			name:          "NETWORK_TIMEOUT maps to unreachable",
			errorCategory: "NETWORK_TIMEOUT",
			currentStatus: "verified",
			want:          "unreachable",
		},
		{
			name:          "NETWORK_DNS maps to unreachable",
			errorCategory: "NETWORK_DNS",
			currentStatus: "verified",
			want:          "unreachable",
		},
		{
			name:          "SERVICE_UNAVAILABLE maps to unreachable",
			errorCategory: "SERVICE_UNAVAILABLE",
			currentStatus: "verified",
			want:          "unreachable",
		},
		{
			name:          "RATE_LIMITED keeps current status",
			errorCategory: "RATE_LIMITED",
			currentStatus: "verified",
			want:          "verified",
		},
		{
			name:          "RATE_LIMITED keeps pending status",
			errorCategory: "RATE_LIMITED",
			currentStatus: "pending",
			want:          "pending",
		},
		{
			name:          "MODEL_NOT_FOUND maps to failed",
			errorCategory: "MODEL_NOT_FOUND",
			currentStatus: "verified",
			want:          "failed",
		},
		{
			name:          "CONTEXT_WINDOW_EXCEEDED maps to failed",
			errorCategory: "CONTEXT_WINDOW_EXCEEDED",
			currentStatus: "verified",
			want:          "failed",
		},
		{
			name:          "UPSTREAM_BAD_REQUEST maps to failed",
			errorCategory: "UPSTREAM_BAD_REQUEST",
			currentStatus: "verified",
			want:          "failed",
		},
		{
			name:          "UNKNOWN maps to failed",
			errorCategory: "UNKNOWN",
			currentStatus: "verified",
			want:          "failed",
		},
		{
			name:          "empty category maps to failed",
			errorCategory: "",
			currentStatus: "verified",
			want:          "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapErrorCategoryToVerificationStatus(tt.errorCategory, tt.currentStatus)
			if got != tt.want {
				t.Errorf("MapErrorCategoryToVerificationStatus(%q, %q) = %q, want %q",
					tt.errorCategory, tt.currentStatus, got, tt.want)
			}
		})
	}
}

func TestIsValidVerificationStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"pending", false},
		{"verified", true},
		{"suspend", false},
		{"unreachable", false},
		{"invalid", false},
		{"failed", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := IsValidVerificationStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsValidVerificationStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
