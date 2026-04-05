package middleware

import (
	"sync"
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

			// Clone the context so the goroutine doesn't race with the
			// timeout path writing on the original response.
			cloned := c.Clone()

			// finished guards against double-write: only the first path
			// (handler completion OR timeout) gets to touch the original context.
			var finished sync.Once
			done := make(chan error, 1)

			go func() {
				err := next(cloned)
				done <- err
			}()

			timer := time.NewTimer(cfg.Timeout)
			defer timer.Stop()

			select {
			case err := <-done:
				// Handler finished in time — no data race because the
				// goroutine wrote to cloned, not c.
				return err
			case <-timer.C:
				// Timeout fired. Drain the goroutine in background so it
				// doesn't leak.
				go func() {
					<-done // wait for handler goroutine to finish
				}()
				var retErr error
				finished.Do(func() {
					retErr = timeoutErr
				})
				return retErr
			}
		}
	}
}
