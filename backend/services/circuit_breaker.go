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
		state:         CircuitStateClosed,
		failureCount:  0,
		successCount:  0,
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
		return true
	case CircuitStateHalfOpen:
		if cb.halfOpenCount >= cb.halfOpenMax {
			return false
		}
		cb.halfOpenCount++
		return true
	case CircuitStateOpen:
		if time.Since(cb.lastFailure) >= cb.timeout {
			cb.state = CircuitStateHalfOpen
			cb.halfOpenCount = 0
			cb.successCount = 0
			cb.halfOpenCount++
			return true
		}
		return false
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.lastSuccess = time.Now()

	if cb.state == CircuitStateHalfOpen {
		if cb.successCount >= cb.halfOpenMax {
			cb.state = CircuitStateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.halfOpenCount = 0
		}
		return
	}
	if cb.state == CircuitStateClosed {
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateOpen
		cb.halfOpenCount = 0
		cb.successCount = 0
		return
	}

	if cb.state == CircuitStateClosed && cb.failureCount >= cb.threshold {
		cb.state = CircuitStateOpen
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
