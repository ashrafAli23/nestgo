package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	core "github.com/ashrafAli23/nestgo/core"
)

// ETagConfig holds ETag middleware configuration.
type ETagConfig struct {
	// Weak generates weak ETags (W/"...") instead of strong ETags.
	// Weak ETags are appropriate when the response is semantically equivalent
	// but not byte-for-byte identical (e.g. different whitespace in JSON).
	// Default: true (weak ETags — safest default for JSON APIs).
	Weak bool
	// SkipFunc optionally skips ETag generation for certain requests.
	// Return true to skip. Default: nil.
	SkipFunc func(c core.Context) bool
}

// DefaultETagConfig returns sensible ETag defaults.
func DefaultETagConfig() ETagConfig {
	return ETagConfig{
		Weak: true,
	}
}

// ETag returns a middleware that generates ETag headers from response bodies
// and handles conditional requests (If-None-Match → 304 Not Modified).
//
// How it works:
//   - After the handler runs, the middleware reads the response body.
//   - It computes a SHA-256 hash of the body and sets it as the ETag header.
//   - If the request includes an If-None-Match header that matches the ETag,
//     a 304 Not Modified response is returned with no body.
//
// This saves bandwidth on read-heavy APIs — clients cache responses and only
// download the body when it has actually changed.
//
// Usage:
//
//	server.Use(middleware.ETag())
//
//	// Strong ETags for byte-exact responses
//	server.Use(middleware.ETag(middleware.ETagConfig{Weak: false}))
//
//	// Skip for streaming endpoints
//	server.Use(middleware.ETag(middleware.ETagConfig{
//	    SkipFunc: func(c core.Context) bool {
//	        return c.GetHeader("Accept") == "text/event-stream"
//	    },
//	}))
func ETag(config ...ETagConfig) core.MiddlewareFunc {
	cfg := DefaultETagConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
				return next(c)
			}

			// Only apply to safe (cacheable) methods.
			method := c.Method()
			if method != http.MethodGet && method != http.MethodHead {
				return next(c)
			}

			// Run the handler
			err := next(c)
			if err != nil {
				return err
			}

			// Only generate ETags for successful responses.
			status := c.ResponseStatus()
			if status < 200 || status >= 300 {
				return nil
			}

			body := c.ResponseBody()
			if len(body) == 0 {
				return nil
			}

			// Compute hash
			hash := sha256.Sum256(body)
			etag := `"` + hex.EncodeToString(hash[:16]) + `"`
			if cfg.Weak {
				etag = "W/" + etag
			}

			c.SetHeader("ETag", etag)

			// Check If-None-Match
			ifNoneMatch := c.GetHeader("If-None-Match")
			if ifNoneMatch != "" && ifNoneMatch == etag {
				return c.NoContent(http.StatusNotModified)
			}

			return nil
		}
	}
}
