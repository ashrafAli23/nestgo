package middleware

import (
	"strconv"

	core "github.com/ashrafAli23/nestgo/core"
)

// BodyLimitConfig holds per-route body size limit configuration.
type BodyLimitConfig struct {
	// MaxBytes is the maximum allowed request body size in bytes.
	// Requests exceeding this limit get a 413 Payload Too Large error.
	MaxBytes int64
	// Message is the error message. Default: "request body too large".
	Message string
}

// BodyLimit returns a middleware that enforces a maximum request body size.
// Use this per-route to override the global Config.BodyLimit.
//
// Usage:
//
//	// Allow 100MB for uploads
//	r.POST("/upload", handler, middleware.BodyLimit(middleware.BodyLimitConfig{
//	    MaxBytes: 100 * 1024 * 1024,
//	}))
//
//	// Restrict JSON endpoints to 1KB
//	r.POST("/api/data", handler, middleware.BodyLimit(middleware.BodyLimitConfig{
//	    MaxBytes: 1024,
//	}))
func BodyLimit(config ...BodyLimitConfig) core.MiddlewareFunc {
	var cfg BodyLimitConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 4 * 1024 * 1024 // 4MB default
	}
	if cfg.Message == "" {
		cfg.Message = "request body too large"
	}

	maxStr := strconv.FormatInt(cfg.MaxBytes, 10)

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			// Check Content-Length header first (fast path, avoids reading body).
			if cl := c.GetHeader("Content-Length"); cl != "" {
				length, err := strconv.ParseInt(cl, 10, 64)
				if err == nil && length > cfg.MaxBytes {
					c.SetHeader("X-Body-Limit", maxStr)
					return core.NewHTTPError(413, cfg.Message)
				}
			}
			return next(c)
		}
	}
}
