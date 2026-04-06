package middleware

import (
	"context"
	"time"

	core "github.com/ashrafAli23/nestgo/core"
)

// TimeoutConfig holds request timeout configuration.
type TimeoutConfig struct {
	// Timeout is the maximum duration for a handler to complete.
	// Default: 30 seconds.
	Timeout time.Duration
	// Message is the error message when timeout is exceeded. Default: "request timeout".
	Message string
	// StatusCode is the HTTP status when timeout is exceeded. Default: 504.
	StatusCode int
	// SkipFunc optionally skips timeout for certain requests (e.g. WebSocket, SSE).
	// Return true to skip.
	SkipFunc func(c core.Context) bool
}

// DefaultTimeoutConfig returns sensible timeout defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout:    30 * time.Second,
		Message:    "request timeout",
		StatusCode: 504,
	}
}

// Timeout returns a middleware that enforces a maximum handler execution time.
// If the handler does not complete within the timeout, a 504 error is returned.
//
// Uses time.NewTimer instead of time.After to properly stop and reclaim the
// timer when the handler completes before the deadline.
//
// Usage:
//
//	// Global: 30 second timeout for all routes
//	server.Use(middleware.Timeout())
//
//	// Custom: 5 second timeout
//	server.Use(middleware.Timeout(middleware.TimeoutConfig{
//	    Timeout: 5 * time.Second,
//	}))
//
//	// Per-route: 60 second timeout for uploads
//	r.POST("/upload", handler, middleware.Timeout(middleware.TimeoutConfig{
//	    Timeout: 60 * time.Second,
//	}))
//
//	// Skip WebSocket and SSE routes
//	server.Use(middleware.Timeout(middleware.TimeoutConfig{
//	    SkipFunc: func(c core.Context) bool {
//	        return c.IsWebSocket() || c.GetHeader("Accept") == "text/event-stream"
//	    },
//	}))
func Timeout(config ...TimeoutConfig) core.MiddlewareFunc {
	cfg := DefaultTimeoutConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Message == "" {
		cfg.Message = "request timeout"
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = 504
	}

	// Pre-create the error to avoid allocation per timeout.
	timeoutErr := core.NewHTTPError(cfg.StatusCode, cfg.Message)

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
				return next(c)
			}

			// Set a deadline on the request context. All downstream I/O
			// (DB queries, HTTP clients, gRPC calls) that respect
			// context.Context will cancel automatically when the
			// deadline fires — no goroutine or Clone needed.
			ctx, cancel := context.WithTimeout(c.RequestCtx(), cfg.Timeout)
			defer cancel()
			c.SetRequestCtx(ctx)

			err := next(c)

			// If the handler returned an error and the deadline was
			// exceeded, replace the error with a clean timeout response.
			if err != nil && ctx.Err() == context.DeadlineExceeded {
				return timeoutErr
			}
			// If the handler succeeded but the deadline fired during
			// response writing, still report the timeout.
			if err == nil && ctx.Err() == context.DeadlineExceeded {
				return timeoutErr
			}
			return err
		}
	}
}
