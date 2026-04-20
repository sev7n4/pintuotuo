package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// TracingResponseHeaders sets X-Trace-ID from the active OpenTelemetry span (W3C trace id hex).
func TracingResponseHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		sc := trace.SpanFromContext(c.Request.Context()).SpanContext()
		if sc.IsValid() {
			c.Writer.Header().Set("X-Trace-ID", sc.TraceID().String())
		}
	}
}
