package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	core "github.com/ashrafAli23/nestgo/core"
)

// RequestIDConfig holds request ID configuration.
type RequestIDConfig struct {
	// Header is the header name to read/write the request ID.
	// Default: "X-Request-ID".
	Header string
	// Generator creates a new request ID when none is provided.
	// Default: random 16-byte hex string.
	Generator func() string
	// ContextKey is the key used to store the request ID in the context.
	// Default: "request_id".
	ContextKey string
}

// DefaultRequestIDConfig returns the default request ID config.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:     "X-Request-ID",
		Generator:  pooledIDGenerator(),
		ContextKey: "request_id",
	}
}

// pooledIDGenerator returns an ID generator that pools byte slices
// to avoid allocation per request.
func pooledIDGenerator() func() string {
	pool := sync.Pool{
		New: func() interface{} {
			b := make([]byte, 16)
			return &b
		},
	}
	return func() string {
		bp := pool.Get().(*[]byte)
		b := *bp
		_, _ = rand.Read(b)
		id := hex.EncodeToString(b)
		pool.Put(bp)
		return id
	}
}

// RequestID returns a middleware that ensures every request has a unique ID.
// If the request already has the header, it's reused; otherwise a new one is generated.
// The ID is stored in the context (via ContextKey) and set on the response header.
func RequestID(config ...RequestIDConfig) core.MiddlewareFunc {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Header == "" {
		cfg.Header = "X-Request-ID"
	}
	if cfg.Generator == nil {
		cfg.Generator = pooledIDGenerator()
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = "request_id"
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			id := c.GetHeader(cfg.Header)
			if id == "" {
				id = cfg.Generator()
			}
			c.Set(cfg.ContextKey, id)
			c.SetHeader(cfg.Header, id)
			return next(c)
		}
	}
}
