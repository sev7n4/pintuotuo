package services

import (
	"context"
	"math"
	"time"
)

type RetryPolicy struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

var DefaultRetryPolicy = &RetryPolicy{
	MaxRetries:    3,
	InitialDelay:  100 * time.Millisecond,
	MaxDelay:      5 * time.Second,
	BackoffFactor: 2.0,
	RetryableErrors: []string{
		"timeout",
		"connection refused",
		"service unavailable",
		"too many requests",
		"internal error",
	},
}

func (p *RetryPolicy) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	if attempt >= p.MaxRetries {
		return false, 0 * time.Millisecond
	}

	errMsg := err.Error()
	for _, retryable := range p.RetryableErrors {
		if contains(errMsg, retryable) {
			delay := p.calculateDelay(attempt)
			return true, delay
		}
	}

	return false, 0 * time.Millisecond
}

func (p *RetryPolicy) calculateDelay(attempt int) time.Duration {
	delay := p.InitialDelay * time.Duration(math.Pow(p.BackoffFactor, float64(attempt)))
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	return delay
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) && s[:len(substr)] == substr)
}

func ExecuteWithRetry(ctx context.Context, fn func() error, policy *RetryPolicy) error {
	var lastErr error
	for i := 0; i < policy.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		shouldRetry, delay := policy.ShouldRetry(err, i)
		if !shouldRetry {
			return err
		}

		lastErr = err

		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		} else {
			time.Sleep(delay)
		}
	}
	return lastErr
}
