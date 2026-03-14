package logger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel represents the severity level of a log
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// RequestLog represents a structured request log
type RequestLog struct {
	Timestamp   time.Time `json:"timestamp"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Status      int       `json:"status"`
	Duration    int64     `json:"duration_ms"`
	UserID      interface{} `json:"user_id,omitempty"`
	RequestID   string    `json:"request_id,omitempty"`
	Error       string    `json:"error,omitempty"`
	ClientIP    string    `json:"client_ip"`
	RequestBody interface{} `json:"request_body,omitempty"`
}

// AppLog represents a structured application log
type AppLog struct {
	Timestamp time.Time   `json:"timestamp"`
	Level     LogLevel    `json:"level"`
	Message   string      `json:"message"`
	Component string      `json:"component,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

var (
	jsonLogger = log.New(os.Stdout, "", 0)
	plainLogger = log.New(os.Stdout, "", log.LstdFlags)
	useJSON    = os.Getenv("LOG_FORMAT") == "json"
)

// LogRequest logs an HTTP request
func LogRequest(c *gin.Context, duration time.Duration, status int, requestID string) {
	userID, _ := c.Get("user_id")

	rl := RequestLog{
		Timestamp:   time.Now(),
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		Status:      status,
		Duration:    duration.Milliseconds(),
		UserID:      userID,
		RequestID:   requestID,
		ClientIP:    c.ClientIP(),
	}

	// Log error if present
	if len(c.Errors) > 0 {
		rl.Error = c.Errors.Last().Error()
	}

	logJSON(rl)
}

// LogDatabase logs database operations
func LogDatabase(component string, operation string, duration time.Duration, err error) {
	message := "database:" + operation
	level := INFO

	if err != nil {
		level = ERROR
		message += ":error"
	}

	al := AppLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: component,
		Data: map[string]interface{}{
			"operation": operation,
			"duration_ms": duration.Milliseconds(),
		},
	}

	if err != nil {
		al.Error = err.Error()
	}

	logJSON(al)
}

// LogCache logs cache operations
func LogCache(operation string, key string, hit bool, duration time.Duration, err error) {
	level := INFO
	if !hit {
		level = DEBUG
	}
	if err != nil {
		level = ERROR
	}

	message := "cache:" + operation
	if hit {
		message += ":hit"
	} else {
		message += ":miss"
	}

	al := AppLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: "cache",
		Data: map[string]interface{}{
			"key": key,
			"hit": hit,
			"duration_ms": duration.Milliseconds(),
		},
	}

	if err != nil {
		al.Error = err.Error()
	}

	logJSON(al)
}

// LogPayment logs payment operations
func LogPayment(operation string, orderID int, amount float64, method string, data interface{}, err error) {
	level := INFO
	if err != nil {
		level = ERROR
	}

	al := AppLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   "payment:" + operation,
		Component: "payment",
		Data: map[string]interface{}{
			"order_id": orderID,
			"amount": amount,
			"method": method,
			"details": data,
		},
	}

	if err != nil {
		al.Error = err.Error()
	}

	logJSON(al)
}

// LogAuth logs authentication operations
func LogAuth(operation string, email string, userID interface{}, err error) {
	level := INFO
	if err != nil {
		level = WARN
	}

	al := AppLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   "auth:" + operation,
		Component: "auth",
		Data: map[string]interface{}{
			"email": email,
			"user_id": userID,
		},
	}

	if err != nil {
		al.Error = err.Error()
	}

	logJSON(al)
}

// LogInfo logs an information message
func LogInfo(ctx context.Context, component string, message string, data interface{}) {
	al := AppLog{
		Timestamp: time.Now(),
		Level:     INFO,
		Message:   message,
		Component: component,
		Data:      data,
	}
	logJSON(al)
}

// LogError logs an error message
func LogError(ctx context.Context, component string, message string, err error, data interface{}) {
	al := AppLog{
		Timestamp: time.Now(),
		Level:     ERROR,
		Message:   message,
		Component: component,
		Data:      data,
	}
	if err != nil {
		al.Error = err.Error()
	}
	logJSON(al)
}

// LogWarn logs a warning message
func LogWarn(ctx context.Context, component string, message string, data interface{}) {
	al := AppLog{
		Timestamp: time.Now(),
		Level:     WARN,
		Message:   message,
		Component: component,
		Data:      data,
	}
	logJSON(al)
}

// LogDebug logs a debug message
func LogDebug(ctx context.Context, component string, message string, data interface{}) {
	al := AppLog{
		Timestamp: time.Now(),
		Level:     DEBUG,
		Message:   message,
		Component: component,
		Data:      data,
	}
	logJSON(al)
}

// logJSON logs structured data as JSON
func logJSON(data interface{}) {
	if useJSON {
		jsonBytes, _ := json.Marshal(data)
		jsonLogger.Println(string(jsonBytes))
	} else {
		plainLogger.Printf("%+v\n", data)
	}
}
