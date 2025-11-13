package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracing returns a middleware that adds OpenTelemetry tracing to HTTP requests
func Tracing(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		// Start a new span
		ctx, span := tracer.Start(
			ctx,
			c.Request.Method+" "+c.FullPath(),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// Add session ID if present
		if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
			span.SetAttributes(attribute.String("session.id", sessionID))
		}

		// Store context in gin context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Record response status
		span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
	}
}

