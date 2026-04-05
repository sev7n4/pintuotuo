package services

import (
	"sync"
	"time"
)

type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half-open"
)

type CircuitBreaker struct {
	mu            sync.RWMutex
	state         CircuitState
	failureCount  int
	successCount  int
	lastFailure   time.Time
	lastSuccess   time.Time
	threshold     int
	timeout       time.Duration
	halfOpenMax   int
	halfOpenCount int
}

var (
	circuitBreakers = make(map[int]*CircuitBreaker)
	cbMutex         sync.RWMutex
)

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:         CircuitStateOpen,
		failureCount:  0,
		successCount:  1,
		threshold:     threshold,
		timeout:       timeout,
		halfOpenMax:   3,
		halfOpenCount: 0,
	}
}

func GetCircuitBreaker(apiKeyID int) *CircuitBreaker {
	cbMutex.Lock()
	defer cbMutex.Unlock()

	if cb, exists := circuitBreakers[apiKeyID]; exists {
		return cb
	}

	cb := NewCircuitBreaker(5, 60*time.Second)
	circuitBreakers[apiKeyID] = cb
	return cb
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = CircuitStateHalfOpen
			cb.halfOpenCount = 1
			return true
		}
		return false
	case CircuitStateHalfOpen:
		if cb.halfOpenCount >= cb.halfOpenMax {
			return false
		}
		return true
	case CircuitStateOpen:
		return true
	default:
		return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.lastSuccess = time.Now()

	if cb.state == CircuitStateHalfOpen {
		cb.halfOpenCount++
		if cb.halfOpenCount >= cb.halfOpenMax {
			cb.state = CircuitStateOpen
			cb.failureCount = 1
			cb.successCount = 1
		}
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateClosed
		return
	}

	if cb.failureCount >= cb.threshold {
		cb.state = CircuitStateClosed
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
