package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// TracingResponseHeaders sets X-Trace-ID from the active OpenTelemetry span (W3C trace id hex).
func TracingResponseHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		sc := trace.SpanFromContext(c.Request.Context()).SpanContext()
		if sc.IsValid() {
			// Set header before downstream handlers write the response body.
			c.Writer.Header().Set("X-Trace-ID", sc.TraceID().String())
		}

		c.Next()

		// Best-effort fallback in case downstream middleware reset headers.
		if c.Writer.Header().Get("X-Trace-ID") == "" {
			sc = trace.SpanFromContext(c.Request.Context()).SpanContext()
		}
		if sc.IsValid() {
			c.Writer.Header().Set("X-Trace-ID", sc.TraceID().String())
		}
	}
}
