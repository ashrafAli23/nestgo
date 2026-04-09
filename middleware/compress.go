package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"strings"

	core "github.com/ashrafAli23/nestgo/core"
)

// CompressConfig holds compression configuration.
type CompressConfig struct {
	// Level is the gzip compression level (1-9). Default: gzip.DefaultCompression.
	// Use gzip.BestSpeed (1) for fastest, gzip.BestCompression (9) for smallest.
	Level int
	// MinLength is the minimum response body size to trigger compression.
	// Default: 1024 bytes. Responses smaller than this are sent uncompressed.
	MinLength int
	// SkipFunc optionally skips compression for certain requests.
	// Return true to skip.
	SkipFunc func(c core.Context) bool
}

// DefaultCompressConfig returns a sensible default compression config.
func DefaultCompressConfig() CompressConfig {
	return CompressConfig{
		Level:     gzip.DefaultCompression,
		MinLength: 1024,
	}
}

// Compress returns a middleware that gzip-compresses JSON responses.
//
// How it works: intercepts the handler's result by replacing the context's
// response path. The handler stores its response data via c.Set("__compress_data"),
// then this middleware compresses and sends it.
//
// For this to work, use CompressJSON() as the response helper instead of c.JSON()
// in handlers that should be compressed. Or use the adapter-native compression:
//
//   - Gin: github.com/gin-contrib/gzip
//   - Fiber: fiber's built-in compress middleware
//
// For a simpler approach that works with any handler, this middleware sets
// Accept-Encoding awareness on the response. Handlers that want compression
// should use the GzipJSON helper function.
func Compress(config ...CompressConfig) core.MiddlewareFunc {
	cfg := DefaultCompressConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Level == 0 {
		cfg.Level = gzip.DefaultCompression
	}
	if cfg.MinLength <= 0 {
		cfg.MinLength = 1024
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
				return next(c)
			}

			// Check if client accepts gzip
			acceptEncoding := c.GetHeader("Accept-Encoding")
			if !strings.Contains(acceptEncoding, "gzip") {
				return next(c)
			}

			// Store compression config in context for GzipJSON to use
			c.Set("__compress_level", cfg.Level)
			c.Set("__compress_min", cfg.MinLength)

			return next(c)
		}
	}
}

// GzipJSON is a helper that sends a JSON response with gzip compression
// if the Compress middleware is active and the client accepts gzip.
// Use this instead of c.JSON() in handlers that should compress responses.
//
//	func (ctrl *Controller) List(c core.Context) error {
//	    data := ctrl.service.GetLargeDataset()
//	    return middleware.GzipJSON(c, 200, data)
//	}
func GzipJSON(c core.Context, status int, data interface{}) error {
	levelRaw := c.Get("__compress_level")
	if levelRaw == nil {
		return c.JSON(status, data)
	}

	level := levelRaw.(int)
	minLength := 1024
	if minRaw := c.Get("__compress_min"); minRaw != nil {
		minLength = minRaw.(int)
	}

	// Marshal JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return core.ErrInternalServer("failed to marshal JSON: " + err.Error())
	}

	// If below minimum length, send uncompressed
	if len(jsonBytes) < minLength {
		c.SetHeader("Content-Type", "application/json")
		return c.SendBytes(status, jsonBytes)
	}

	// Compress
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		// Fallback to uncompressed
		c.SetHeader("Content-Type", "application/json")
		return c.SendBytes(status, jsonBytes)
	}
	if _, err = writer.Write(jsonBytes); err != nil {
		c.SetHeader("Content-Type", "application/json")
		return c.SendBytes(status, jsonBytes)
	}
	if err = writer.Close(); err != nil {
		c.SetHeader("Content-Type", "application/json")
		return c.SendBytes(status, jsonBytes)
	}

	c.SetHeader("Content-Encoding", "gzip")
	c.SetHeader("Content-Type", "application/json")
	c.SetHeader("Vary", "Accept-Encoding")
	return c.SendBytes(status, buf.Bytes())
}
