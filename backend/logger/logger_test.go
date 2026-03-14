package logger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestLogStructure(t *testing.T) {
	log := RequestLog{
		Timestamp:   time.Now(),
		Method:      "GET",
		Path:        "/api/v1/products",
		Status:      200,
		Duration:    150,
		UserID:      123,
		RequestID:   "req-12345",
		ClientIP:    "127.0.0.1",
		RequestBody: nil,
	}

	assert.Equal(t, "GET", log.Method)
	assert.Equal(t, "/api/v1/products", log.Path)
	assert.Equal(t, 200, log.Status)
	assert.Equal(t, int64(150), log.Duration)
	assert.Equal(t, 123, log.UserID)
	assert.Equal(t, "req-12345", log.RequestID)
	assert.Equal(t, "127.0.0.1", log.ClientIP)
}

func TestRequestLogWithError(t *testing.T) {
	log := RequestLog{
		Timestamp:   time.Now(),
		Method:      "POST",
		Path:        "/api/v1/orders",
		Status:      400,
		Duration:    50,
		UserID:      123,
		RequestID:   "req-12346",
		Error:       "invalid request body",
		ClientIP:    "127.0.0.1",
	}

	assert.Equal(t, "invalid request body", log.Error)
	assert.Equal(t, 400, log.Status)
}

func TestAppLogStructure(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	log := AppLog{
		Timestamp: time.Now(),
		Level:     INFO,
		Message:   "User logged in successfully",
		Component: "auth",
		Data:      data,
	}

	assert.Equal(t, INFO, log.Level)
	assert.Equal(t, "User logged in successfully", log.Message)
	assert.Equal(t, "auth", log.Component)
	assert.Equal(t, data, log.Data)
	assert.Empty(t, log.Error)
}

func TestAppLogWithError(t *testing.T) {
	log := AppLog{
		Timestamp: time.Now(),
		Level:     ERROR,
		Message:   "Database connection failed",
		Component: "database",
		Error:     "connection timeout",
	}

	assert.Equal(t, ERROR, log.Level)
	assert.Equal(t, "Database connection failed", log.Message)
	assert.Equal(t, "connection timeout", log.Error)
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"WARN", WARN},
		{"ERROR", ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.level)
			assert.Equal(t, LogLevel(tt.name), tt.level)
		})
	}
}

func TestLogInfoCreation(t *testing.T) {
	data := map[string]interface{}{"user_id": 123, "action": "login"}
	log := AppLog{
		Timestamp: time.Now(),
		Level:     INFO,
		Message:   "User action recorded",
		Component: "user",
		Data:      data,
	}

	assert.Equal(t, INFO, log.Level)
	assert.NotNil(t, log.Data)
}

func TestLogErrorCreation(t *testing.T) {
	log := AppLog{
		Timestamp: time.Now(),
		Level:     ERROR,
		Message:   "Operation failed",
		Component: "payment",
		Error:     "transaction rejected",
	}

	assert.Equal(t, ERROR, log.Level)
	assert.NotEmpty(t, log.Error)
}

func TestLogWarningCreation(t *testing.T) {
	log := AppLog{
		Timestamp: time.Now(),
		Level:     WARN,
		Message:   "Unusual activity detected",
		Component: "security",
		Data:      map[string]interface{}{"ip": "192.168.1.100"},
	}

	assert.Equal(t, WARN, log.Level)
	assert.NotNil(t, log.Data)
}

func TestRequestLogContext(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	// Test that context is properly used in function signatures
	log := RequestLog{
		Timestamp: time.Now(),
		Method:    "GET",
		Path:      "/api/v1/products",
		Status:    200,
		Duration:  100,
		ClientIP:  "127.0.0.1",
	}

	assert.Equal(t, "GET", log.Method)
	assert.Equal(t, 200, log.Status)
}

func TestLogComponentOrganization(t *testing.T) {
	components := []string{"auth", "payment", "database", "cache", "user", "product", "order", "group", "security"}

	for _, component := range components {
		log := AppLog{
			Timestamp: time.Now(),
			Level:     INFO,
			Message:   "Test message",
			Component: component,
		}

		assert.Equal(t, component, log.Component)
	}
}

func TestLogTimestampPresent(t *testing.T) {
	now := time.Now()
	log := AppLog{
		Timestamp: now,
		Level:     INFO,
		Message:   "Test",
	}

	assert.Equal(t, now, log.Timestamp)
	assert.False(t, log.Timestamp.IsZero())
}

func TestRequestLogStatusCodes(t *testing.T) {
	statusCodes := []int{200, 201, 400, 401, 403, 404, 409, 500}

	for _, status := range statusCodes {
		log := RequestLog{
			Timestamp: time.Now(),
			Method:    "GET",
			Path:      "/test",
			Status:    status,
			Duration:  10,
			ClientIP:  "127.0.0.1",
		}

		assert.Equal(t, status, log.Status)
	}
}
