// Package middleware provides production-ready HTTP middleware for NestGo.
//
// Every middleware follows the same pattern:
//   - Config struct with sensible defaults
//   - DefaultXxxConfig() returns the defaults
//   - Xxx(config ...XxxConfig) returns a [core.MiddlewareFunc]
//   - Pass zero args for defaults, or one config to customize
//
// # Available Middleware
//
//   - [Recovery] — panic recovery (register first)
//   - [CORS] — Cross-Origin Resource Sharing
//   - [Helmet] — OWASP security headers
//   - [RateLimit] — per-key rate limiting with sharded storage
//   - [RequestID] — unique request ID per request
//   - [Timeout] — request execution deadline
//   - [CSRF] — Cross-Site Request Forgery protection
//   - [Compress] / [GzipJSON] — gzip response compression
//
// # Usage
//
// Apply globally:
//
//	server.Use(middleware.Recovery())
//	server.Use(middleware.Helmet())
//	server.Use(middleware.CORS())
//	server.Use(middleware.RateLimit())
//
// Apply per-group:
//
//	api := r.Group("/api", middleware.RateLimit(middleware.RateLimitConfig{
//	    Max:    200,
//	    Window: time.Minute,
//	}))
//
// Apply per-route:
//
//	r.POST("/upload", handler, middleware.Timeout(middleware.TimeoutConfig{
//	    Timeout: 60 * time.Second,
//	}))
//
// All middleware depends only on [github.com/ashrafAli23/nestgo/core] — no
// adapter-specific imports.
package middleware
