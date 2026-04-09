package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	core "github.com/ashrafAli23/nestgo/core"
)

// TracingConfig holds distributed tracing configuration.
type TracingConfig struct {
	// TraceHeader is the incoming header to extract the trace ID from.
	// Default: "traceparent" (W3C Trace Context).
	TraceHeader string
	// TraceIDContextKey is the key used to store trace_id in the request context.
	// Default: "trace_id".
	TraceIDContextKey string
	// SpanIDContextKey is the key used to store span_id in the request context.
	// Default: "span_id".
	SpanIDContextKey string
	// ResponseHeader sets the trace ID on the response for client correlation.
	// Default: "X-Trace-ID". Set to "" to disable.
	ResponseHeader string
	// RecordDuration stores the handler duration (in ms) in the context under "trace_duration_ms".
	// Useful for metrics collection. Default: true.
	RecordDuration bool
}

// DefaultTracingConfig returns sensible tracing defaults.
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		TraceHeader:       "traceparent",
		TraceIDContextKey: "trace_id",
		SpanIDContextKey:  "span_id",
		ResponseHeader:    "X-Trace-ID",
		RecordDuration:    true,
	}
}

// traceIDPool pools byte slices for trace/span ID generation.
var traceIDPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 16)
		return &b
	},
}

var spanIDPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 8)
		return &b
	},
}

func generateTraceID() string {
	bp := traceIDPool.Get().(*[]byte)
	b := *bp
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)
	traceIDPool.Put(bp)
	return id
}

func generateSpanID() string {
	bp := spanIDPool.Get().(*[]byte)
	b := *bp
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)
	spanIDPool.Put(bp)
	return id
}

// parseTraceparent extracts trace_id from a W3C traceparent header.
// Format: version-trace_id-parent_id-trace_flags (e.g. "00-abc123...-def456...-01")
func parseTraceparent(header string) (traceID, parentSpanID string) {
	parts := strings.SplitN(header, "-", 4)
	if len(parts) < 3 {
		return "", ""
	}
	return parts[1], parts[2]
}

// Tracing returns a middleware that propagates distributed trace context.
//
// It extracts trace_id from the incoming W3C traceparent header (or generates one),
// creates a new span_id for this request, and stores both in the request context.
// Downstream code and the Logger middleware can read these for correlated logging.
//
// This works standalone for trace propagation. For full OpenTelemetry integration
// (exporters, sampling, span hierarchies), use the OTel SDK and inject its middleware
// via GlobalMiddlewares — the trace_id from this middleware will match.
//
// Usage:
//
//	server.Use(middleware.Tracing())
//
//	// In a handler or interceptor:
//	traceID, _ := c.Get("trace_id").(string)
func Tracing(config ...TracingConfig) core.MiddlewareFunc {
	cfg := DefaultTracingConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.TraceHeader == "" {
		cfg.TraceHeader = "traceparent"
	}
	if cfg.TraceIDContextKey == "" {
		cfg.TraceIDContextKey = "trace_id"
	}
	if cfg.SpanIDContextKey == "" {
		cfg.SpanIDContextKey = "span_id"
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			// Extract or generate trace ID
			traceID := ""
			if header := c.GetHeader(cfg.TraceHeader); header != "" {
				traceID, _ = parseTraceparent(header)
			}
			if traceID == "" {
				traceID = generateTraceID()
			}

			// Always generate a new span ID for this request
			spanID := generateSpanID()

			c.Set(cfg.TraceIDContextKey, traceID)
			c.Set(cfg.SpanIDContextKey, spanID)

			if cfg.ResponseHeader != "" {
				c.SetHeader(cfg.ResponseHeader, traceID)
			}

			if cfg.RecordDuration {
				start := time.Now()
				err := next(c)
				c.Set("trace_duration_ms", time.Since(start).Milliseconds())
				return err
			}

			return next(c)
		}
	}
}
