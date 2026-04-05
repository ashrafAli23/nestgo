package middleware

import (
	"strconv"
	"sync"
	"time"

	core "github.com/ashrafAli23/nestgo/core"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	// Max number of requests allowed within the Window.
	Max int
	// Window is the time window for rate limiting.
	Window time.Duration
	// KeyFunc extracts the rate limit key from the request (e.g. IP, user ID).
	// Default: ClientIP.
	KeyFunc func(c core.Context) string
	// Message is the error message when rate limited. Default: "too many requests".
	Message string
	// StatusCode is the HTTP status when rate limited. Default: 429.
	StatusCode int
	// SkipFunc optionally skips rate limiting for certain requests.
	// Return true to skip.
	SkipFunc func(c core.Context) bool
	// Headers controls whether to set X-RateLimit-* headers. Default: true.
	Headers bool
	// Stop is an optional channel that, when closed, stops the background
	// cleanup goroutine. If nil, the goroutine runs for the process lifetime.
	// Useful for tests and hot-reload scenarios to prevent goroutine leaks.
	Stop <-chan struct{}
}

// DefaultRateLimitConfig returns a sensible default rate limit config.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:    100,
		Window: time.Minute,
		KeyFunc: func(c core.Context) string {
			return c.ClientIP()
		},
		Message:    "too many requests",
		StatusCode: 429,
		Headers:    true,
	}
}

type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

const rateLimitShards = 32

// rateLimitShard is a single shard of the rate limit store.
// Sharding reduces lock contention under high concurrency —
// requests with different IPs hit different shards.
type rateLimitShard struct {
	mu    sync.Mutex
	store map[string]*rateLimitEntry
}

func newShardedStore() []*rateLimitShard {
	shards := make([]*rateLimitShard, rateLimitShards)
	for i := range shards {
		shards[i] = &rateLimitShard{store: make(map[string]*rateLimitEntry)}
	}
	return shards
}

// getShard selects a shard using inline FNV-1a to avoid allocating a hash.Hash per request.
func getShard(shards []*rateLimitShard, key string) *rateLimitShard {
	// FNV-1a inline — same algorithm as fnv.New32a() but zero allocations.
	var h uint32 = 2166136261 // FNV offset basis
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619 // FNV prime
	}
	return shards[h%rateLimitShards]
}

// RateLimit returns a rate limiting middleware.
func RateLimit(config ...RateLimitConfig) core.MiddlewareFunc {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Max <= 0 {
		cfg.Max = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c core.Context) string { return c.ClientIP() }
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = 429
	}
	if cfg.Message == "" {
		cfg.Message = "too many requests"
	}

	shards := newShardedStore()

	// Pre-format static header value to avoid repeated Sprintf
	maxHeader := strconv.Itoa(cfg.Max)

	// Background cleanup goroutine — evicts expired entries to prevent memory leak.
	// Stops when cfg.Stop is closed (if provided), preventing goroutine leaks
	// in tests or hot-reload scenarios.
	stop := cfg.Stop
	if stop == nil {
		// No stop signal — goroutine runs for process lifetime.
		stop = make(chan struct{})
	}
	go func() {
		ticker := time.NewTicker(cfg.Window)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				for _, shard := range shards {
					shard.mu.Lock()
					for key, entry := range shard.store {
						if now.After(entry.expiresAt) {
							delete(shard.store, key)
						}
					}
					shard.mu.Unlock()
				}
			case <-stop:
				return
			}
		}
	}()

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
				return next(c)
			}

			key := cfg.KeyFunc(c)
			now := time.Now()

			shard := getShard(shards, key)
			shard.mu.Lock()
			entry, exists := shard.store[key]
			if !exists || now.After(entry.expiresAt) {
				entry = &rateLimitEntry{
					count:     0,
					expiresAt: now.Add(cfg.Window),
				}
				shard.store[key] = entry
			}
			entry.count++
			count := entry.count
			remaining := cfg.Max - count
			expiresAt := entry.expiresAt
			shard.mu.Unlock()

			if remaining < 0 {
				remaining = 0
			}

			if cfg.Headers {
				c.SetHeader("X-RateLimit-Limit", maxHeader)
				c.SetHeader("X-RateLimit-Remaining", strconv.Itoa(remaining))
				c.SetHeader("X-RateLimit-Reset", strconv.FormatInt(expiresAt.Unix(), 10))
			}

			if count > cfg.Max {
				return core.NewHTTPErrorWithDetails(cfg.StatusCode, cfg.Message, nil)
			}

			return next(c)
		}
	}
}
