package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pintuotuo/backend/metrics"
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		startPos := c.Writer.Size()

		c.Next()

		duration := time.Since(startTime).Seconds()
		endPos := c.Writer.Size()
		requestSize := int64(c.Request.ContentLength)
		responseSize := int64(endPos - startPos)

		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			requestSize,
			responseSize,
		)

		// Record errors if they exist
		if len(c.Errors) > 0 {
			for range c.Errors {
				metrics.RecordApplicationError(
					"HTTP_ERROR",
					"error",
				)
			}
		}
	}
}

// PrometheusHandler returns the Prometheus metrics handler
func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
