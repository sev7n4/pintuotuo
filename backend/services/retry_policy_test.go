package services

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryPolicy(t *testing.T) {
	if DefaultRetryPolicy.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", DefaultRetryPolicy.MaxRetries)
	}

	if DefaultRetryPolicy.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay=100ms, got %v", DefaultRetryPolicy.InitialDelay)
	}

	if DefaultRetryPolicy.MaxDelay != 5*time.Second {
		t.Errorf("Expected MaxDelay=5s, got %v", DefaultRetryPolicy.MaxDelay)
	}

	if DefaultRetryPolicy.BackoffFactor != 2.0 {
		t.Errorf("Expected BackoffFactor=2.0, got %f", DefaultRetryPolicy.BackoffFactor)
	}
}

func TestRetryPolicyShouldRetry(t *testing.T) {
	policy := DefaultRetryPolicy

	tests := []struct {
		name        string
		err         error
		attempt     int
		shouldRetry bool
	}{
		{
			name:        "timeout error",
			err:         errors.New("timeout"),
			attempt:     0,
			shouldRetry: true,
		},
		{
			name:        "connection refused error",
			err:         errors.New("connection refused"),
			attempt:     1,
			shouldRetry: true,
		},
		{
			name:        "non-retryable error",
			err:         errors.New("invalid request"),
			attempt:     0,
			shouldRetry: false,
		},
		{
			name:        "max retries exceeded",
			err:         errors.New("timeout"),
			attempt:     3,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRetry, _ := policy.ShouldRetry(tt.err, tt.attempt)
			if shouldRetry != tt.shouldRetry {
				t.Errorf("Expected shouldRetry=%v, got %v", tt.shouldRetry, shouldRetry)
			}
		})
	}
}

func TestRetryPolicyCalculateDelay(t *testing.T) {
	policy := DefaultRetryPolicy

	tests := []struct {
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{0, 100 * time.Millisecond, 100 * time.Millisecond},
		{1, 200 * time.Millisecond, 200 * time.Millisecond},
		{2, 400 * time.Millisecond, 400 * time.Millisecond},
		{3, 800 * time.Millisecond, 800 * time.Millisecond},
		{10, 5 * time.Second, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			delay := policy.calculateDelay(tt.attempt)
			if delay < tt.expectedMin || delay > tt.expectedMax {
				t.Errorf("Attempt %d: expected delay between %v and %v, got %v",
					tt.attempt, tt.expectedMin, tt.expectedMax, delay)
			}
		})
	}
}

func TestExecuteWithRetry(t *testing.T) {
	policy := &RetryPolicy{
		MaxRetries:      3,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"timeout"},
	}

	tests := []struct {
		name       string
		failures   int
		shouldFail bool
	}{
		{
			name:       "success on first try",
			failures:   0,
			shouldFail: false,
		},
		{
			name:       "success after retry",
			failures:   2,
			shouldFail: false,
		},
		{
			name:       "fail after max retries",
			failures:   4,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			err := ExecuteWithRetry(nil, func() error {
				callCount++
				if callCount <= tt.failures {
					return errors.New("timeout")
				}
				return nil
			}, policy)

			if tt.shouldFail && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"timeout error", "timeout", true},
		{"connection refused", "connection", true},
		{"invalid request", "valid", true},
		{"TIMEOUT while waiting", "timeout", true},
		{"", "test", false},
		{"test", "", false},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%s, %s): expected %v, got %v", tt.s, tt.substr, tt.expected, result)
		}
	}
}
