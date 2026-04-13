package services

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(5, 60*time.Second)

	if cb == nil {
		t.Fatal("Expected circuit breaker, got nil")
	}

	if cb.threshold != 5 {
		t.Errorf("Expected threshold=5, got %d", cb.threshold)
	}

	if cb.timeout != 60*time.Second {
		t.Errorf("Expected timeout=60s, got %v", cb.timeout)
	}

	if cb.state != CircuitStateOpen {
		t.Errorf("Expected initial state=open, got %s", cb.state)
	}
}

func TestCircuitBreakerStates(t *testing.T) {
	tests := []struct {
		name     string
		state    CircuitState
		expected string
	}{
		{"closed state", CircuitStateClosed, "closed"},
		{"open state", CircuitStateOpen, "open"},
		{"half-open state", CircuitStateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.state) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.state)
			}
		})
	}
}

func TestCircuitBreakerAllowRequest(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	if !cb.AllowRequest() {
		t.Error("Should allow request in open state")
	}

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitStateClosed {
		t.Error("Should be closed after 3 failures")
	}

	if cb.AllowRequest() {
		t.Error("Should not allow request in closed state")
	}

	time.Sleep(1100 * time.Millisecond)

	if !cb.AllowRequest() {
		t.Error("Should allow request after timeout")
	}

	if cb.GetState() != CircuitStateHalfOpen {
		t.Error("Should be in half-open state after timeout")
	}
}

func TestCircuitBreakerRecordSuccess(t *testing.T) {
	cb := NewCircuitBreaker(2, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitStateClosed {
		t.Error("Should be closed after failures")
	}

	time.Sleep(100 * time.Millisecond)

	cb.state = CircuitStateHalfOpen
	cb.halfOpenCount = 0

	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.GetState() != CircuitStateOpen {
		t.Errorf("Should be open after 3 successes in half-open, got %s", cb.GetState())
	}
}

func TestCircuitBreakerRecordFailure(t *testing.T) {
	cb := NewCircuitBreaker(3, 60*time.Second)

	cb.RecordFailure()
	if cb.GetState() != CircuitStateOpen {
		t.Error("Should still be open after 1 failure")
	}

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitStateClosed {
		t.Error("Should be closed after 3 failures")
	}

	cb.state = CircuitStateHalfOpen
	cb.RecordFailure()

	if cb.GetState() != CircuitStateClosed {
		t.Error("Should be closed after failure in half-open state")
	}
}

func TestGetCircuitBreaker(t *testing.T) {
	cb1 := GetCircuitBreaker(1)
	cb2 := GetCircuitBreaker(1)

	if cb1 != cb2 {
		t.Error("Expected same instance for same API key ID")
	}

	cb3 := GetCircuitBreaker(2)
	if cb1 == cb3 {
		t.Error("Expected different instances for different API key IDs")
	}
}

func TestCircuitBreakerConcurrent(t *testing.T) {
	cb := NewCircuitBreaker(100, 60*time.Second)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cb.AllowRequest()
				cb.RecordSuccess()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if cb.GetState() != CircuitStateOpen {
		t.Error("Should remain open after concurrent successes")
	}
}
