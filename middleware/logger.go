package middleware

import (
	"time"

	core "github.com/ashrafAli23/nestgo/core"
)

// LoggerConfig holds request logging configuration.
type LoggerConfig struct {
	// SkipFunc optionally skips logging for certain requests (e.g. health checks).
	// Return true to skip.
	SkipFunc func(c core.Context) bool
	// LogFunc overrides the default log output. Receives the context after the
	// handler has run, the duration, and the handler error (nil on success).
	LogFunc func(c core.Context, status int, duration time.Duration, err error)
}

// Logger returns a middleware that logs every request with method, path,
// status code, and duration.
//
// Usage:
//
//	// Default: logs via core.Log()
//	server.Use(middleware.Logger())
//
//	// Skip health checks
//	server.Use(middleware.Logger(middleware.LoggerConfig{
//	    SkipFunc: func(c core.Context) bool {
//	        return c.Path() == "/health"
//	    },
//	}))
//
//	// Custom log output (e.g. structured JSON logger)
//	server.Use(middleware.Logger(middleware.LoggerConfig{
//	    LogFunc: func(c core.Context, status int, d time.Duration, err error) {
//	        slog.Info("request", "method", c.Method(), "path", c.FullURL(), "status", status, "ms", d.Milliseconds())
//	    },
//	}))
func Logger(config ...LoggerConfig) core.MiddlewareFunc {
	var cfg LoggerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			// Determine status: from the response if written, or from the error.
			status := c.ResponseStatus()
			if err != nil {
				if httpErr, ok := err.(*core.HTTPError); ok {
					status = httpErr.Code
				} else if status == 0 {
					status = 500
				}
			}

			if cfg.LogFunc != nil {
				cfg.LogFunc(c, status, duration, err)
			} else {
				fields := []core.Field{
					core.F("method", c.Method()),
					core.F("path", c.FullURL()),
					core.F("status", status),
					core.F("duration", duration.String()),
					core.F("ip", c.ClientIP()),
				}
				// Auto-include correlation IDs when present
				if rid, _ := c.Get("request_id").(string); rid != "" {
					fields = append(fields, core.F("request_id", rid))
				}
				if tid, _ := c.Get("trace_id").(string); tid != "" {
					fields = append(fields, core.F("trace_id", tid))
				}
				if err != nil {
					fields = append(fields, core.F("error", err.Error()))
					core.Log().Error("request", fields...)
				} else {
					core.Log().Info("request", fields...)
				}
			}

			return err
		}
	}
}
