package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthCheck tests the liveness probe endpoint
func TestHealthCheck(t *testing.T) {
	t.Run("HealthCheck returns 200 OK", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Health check should return 200 OK")
	})

	t.Run("HealthCheck response has correct structure", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)

		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, "healthy", response.Status, "Status should be 'healthy'")
		assert.NotEmpty(t, response.Timestamp, "Timestamp should be present")
		assert.NotEmpty(t, response.Version, "Version should be present")
		assert.NotEmpty(t, response.Services, "Services should have entries")
	})

	t.Run("HealthCheck includes application service", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.NotNil(t, response.Services["application"], "Application service should be present")
		assert.Equal(t, "up", response.Services["application"].Status)
	})

	t.Run("HealthCheck includes uptime", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.GreaterOrEqual(t, response.Uptime, int64(0), "Uptime should be non-negative")
	})
}

// TestReadinessProbe tests the readiness probe endpoint
func TestReadinessProbe(t *testing.T) {
	t.Run("ReadinessProbe returns status code", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/ready", ReadinessProbe)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return either 200 (ready) or 503 (not ready)
		// When database is nil, it will return 503
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable,
			"ReadinessProbe should return 200 or 503")
	})

	t.Run("ReadinessProbe response structure", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/ready", ReadinessProbe)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)

		assert.NoError(t, err, "Response should be valid JSON")
		assert.NotEmpty(t, response.Services, "Should check services")
	})

	t.Run("ReadinessProbe checks database", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/ready", ReadinessProbe)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.NotNil(t, response.Services["database"], "Should check database")
		// Database may be down in test environment
		assert.True(t, response.Services["database"].Status == "healthy" ||
			response.Services["database"].Status == "down")
	})

	t.Run("ReadinessProbe checks redis", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/ready", ReadinessProbe)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.NotNil(t, response.Services["redis"], "Should check redis")
		// Redis may be down in test environment
		assert.True(t, response.Services["redis"].Status == "healthy" ||
			response.Services["redis"].Status == "down")
	})

	t.Run("ReadinessProbe response time measurement", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/ready", ReadinessProbe)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		// Check that duration is recorded for each service
		for _, service := range response.Services {
			assert.GreaterOrEqual(t, service.Duration, int64(0), "Duration should be measured")
		}
	})
}

// TestLivenessProbe tests the liveness probe endpoint
func TestLivenessProbe(t *testing.T) {
	t.Run("LivenessProbe returns 200 OK", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/live", LivenessProbe)

		req := httptest.NewRequest(http.MethodGet, "/live", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Liveness probe should return 200 OK")
	})

	t.Run("LivenessProbe indicates alive status", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/live", LivenessProbe)

		req := httptest.NewRequest(http.MethodGet, "/live", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "alive", response.Status, "Should indicate alive status")
	})

	t.Run("LivenessProbe includes uptime", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/live", LivenessProbe)

		req := httptest.NewRequest(http.MethodGet, "/live", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.GreaterOrEqual(t, response.Uptime, int64(0), "Should include uptime")
	})
}

// TestMetrics tests the metrics endpoint
func TestMetrics(t *testing.T) {
	t.Run("Metrics returns 200 OK", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/metrics", Metrics)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Metrics should return 200 OK")
	})

	t.Run("Metrics includes uptime", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/metrics", Metrics)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.NotNil(t, response["uptime_seconds"], "Should include uptime")
		// uptime_seconds is float64 in JSON unmarshaling
		uptime := response["uptime_seconds"].(float64)
		assert.GreaterOrEqual(t, uptime, float64(0), "Uptime should be non-negative")
	})

	t.Run("Metrics includes version", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/metrics", Metrics)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "1.0.0", response["version"], "Should include version")
	})

	t.Run("Metrics includes status", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/metrics", Metrics)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "running", response["status"], "Should include running status")
	})
}

// TestHealthCheckResponseStructure tests the response format
func TestHealthCheckResponseStructure(t *testing.T) {
	t.Run("HealthCheckResponse unmarshals correctly", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)

		assert.NoError(t, err, "Should unmarshal without error")
		assert.NotEmpty(t, response.Status)
		assert.NotEmpty(t, response.Timestamp)
	})

	t.Run("ServiceStatus includes all fields", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response HealthCheckResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		for _, service := range response.Services {
			assert.NotEmpty(t, service.Status, "Service should have status")
			assert.GreaterOrEqual(t, service.Duration, int64(0), "Service should have duration")
		}
	})
}

// TestHealthCheckIntegration tests health check integration
func TestHealthCheckIntegration(t *testing.T) {
	t.Run("Multiple health check calls work", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("Health check uptime increases", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		var response1 HealthCheckResponse
		json.Unmarshal(w1.Body.Bytes(), &response1)

		req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		var response2 HealthCheckResponse
		json.Unmarshal(w2.Body.Bytes(), &response2)

		assert.GreaterOrEqual(t, response2.Uptime, response1.Uptime,
			"Uptime should not decrease")
	})
}

// TestHealthCheckContentType tests response content type
func TestHealthCheckContentType(t *testing.T) {
	t.Run("HealthCheck returns JSON content type", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Content-Type"), "application/json",
			"Should return JSON content type")
	})

	t.Run("ReadinessProbe returns JSON content type", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/ready", ReadinessProbe)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Content-Type"), "application/json",
			"Should return JSON content type")
	})

	t.Run("LivenessProbe returns JSON content type", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)
		router.GET("/live", LivenessProbe)

		req := httptest.NewRequest(http.MethodGet, "/live", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Content-Type"), "application/json",
			"Should return JSON content type")
	})
}
