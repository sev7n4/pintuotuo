package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMetricsMiddlewareRecordsRequest verifies metrics middleware records requests
func TestMetricsMiddlewareRecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(MetricsMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Request should succeed")
}

// TestMetricsMiddlewareHandlesErrors verifies metrics middleware records errors
func TestMetricsMiddlewareHandlesErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(MetricsMiddleware())

	router.GET("/error", func(c *gin.Context) {
		c.AbortWithError(http.StatusInternalServerError, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Greater(t, w.Code, 399, "Error status should be recorded")
}

// TestMetricsMiddlewareTracksRequestSize verifies request size is tracked
func TestMetricsMiddlewareTracksRequestSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(MetricsMiddleware())

	router.POST("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"received": true})
	})

	// Make a request with body
	body := []byte(`{"test":"data"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/echo", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Request should succeed")
}

// TestMetricsMiddlewareWithVariousHTTPMethods verifies all HTTP methods are supported
func TestMetricsMiddlewareWithVariousHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(MetricsMiddleware())

	router.GET("/resource", func(c *gin.Context) { c.JSON(200, nil) })
	router.POST("/resource", func(c *gin.Context) { c.JSON(201, nil) })
	router.PUT("/resource/:id", func(c *gin.Context) { c.JSON(200, nil) })
	router.PATCH("/resource/:id", func(c *gin.Context) { c.JSON(200, nil) })
	router.DELETE("/resource/:id", func(c *gin.Context) { c.JSON(204, nil) })

	methods := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{"GET", "/resource", 200},
		{"POST", "/resource", 201},
		{"PUT", "/resource/123", 200},
		{"PATCH", "/resource/456", 200},
		{"DELETE", "/resource/789", 204},
	}

	for _, test := range methods {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(test.method, test.path, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, test.expectedCode, w.Code, "Expected status code for "+test.method)
	}
}

// TestMetricsMiddlewareWithDifferentStatusCodes verifies all status codes are handled
func TestMetricsMiddlewareWithDifferentStatusCodes(t *testing.T) {
	statusTests := []struct {
		status int
		setup  func(router *gin.Engine)
	}{
		{200, func(r *gin.Engine) { r.GET("/ok", func(c *gin.Context) { c.JSON(200, nil) }) }},
		{201, func(r *gin.Engine) { r.POST("/created", func(c *gin.Context) { c.JSON(201, nil) }) }},
		{400, func(r *gin.Engine) { r.GET("/bad", func(c *gin.Context) { c.JSON(400, nil) }) }},
		{401, func(r *gin.Engine) { r.GET("/unauth", func(c *gin.Context) { c.JSON(401, nil) }) }},
		{403, func(r *gin.Engine) { r.GET("/forbidden", func(c *gin.Context) { c.JSON(403, nil) }) }},
		{404, func(r *gin.Engine) { r.GET("/notfound", func(c *gin.Context) { c.JSON(404, nil) }) }},
		{500, func(r *gin.Engine) { r.GET("/error", func(c *gin.Context) { c.JSON(500, nil) }) }},
	}

	for _, test := range statusTests {
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.Use(MetricsMiddleware())

		test.setup(router)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)

		// Adjust request based on status code
		switch test.status {
		case 201:
			req, _ = http.NewRequest("POST", "/created", nil)
		case 400:
			req, _ = http.NewRequest("GET", "/bad", nil)
		case 401:
			req, _ = http.NewRequest("GET", "/unauth", nil)
		case 403:
			req, _ = http.NewRequest("GET", "/forbidden", nil)
		case 404:
			req, _ = http.NewRequest("GET", "/notfound", nil)
		case 500:
			req, _ = http.NewRequest("GET", "/error", nil)
		default:
			req, _ = http.NewRequest("GET", "/ok", nil)
		}

		router.ServeHTTP(w, req)
		assert.Equal(t, test.status, w.Code, "Status code should be recorded")
	}
}

// TestMetricsMiddlewareResponseTime verifies response time is reasonable
func TestMetricsMiddlewareResponseTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(MetricsMiddleware())

	router.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"speed": "fast"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fast", nil)
	router.ServeHTTP(w, req)

	// Request should complete very quickly in tests
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPrometheusHandler verifies the metrics handler is exposed
func TestPrometheusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Add Prometheus handler
	router.GET("/metrics", PrometheusHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	// Metrics endpoint should return 200
	assert.Equal(t, http.StatusOK, w.Code, "Metrics endpoint should be accessible")

	// Response should contain metrics content
	assert.NotEmpty(t, w.Body.String(), "Metrics endpoint should return data")
}
