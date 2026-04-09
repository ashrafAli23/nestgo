package middleware

import (
	"net/http"
	"sync"
	"time"

	core "github.com/ashrafAli23/nestgo/core"
)

// IdempotencyConfig holds idempotency middleware configuration.
type IdempotencyConfig struct {
	// Header is the header name clients must send with a unique key.
	// Default: "Idempotency-Key".
	Header string
	// TTL is how long cached responses are kept. Default: 24 hours.
	TTL time.Duration
	// Methods is the list of HTTP methods to enforce idempotency on.
	// Default: POST, PUT, PATCH.
	Methods []string
	// Required when true returns 400 if the idempotency header is missing
	// on applicable methods. Default: false (passes through without caching).
	Required bool
	// Store is the backing store for idempotency records.
	// Default: in-memory (suitable for single-instance deployments).
	// For multi-replica deployments, implement IdempotencyStore with Redis.
	Store IdempotencyStore
	// Stop is an optional channel to stop the background cleanup goroutine.
	Stop <-chan struct{}
}

// IdempotencyStore is the interface for pluggable idempotency storage.
// Implement this with Redis for multi-instance deployments.
type IdempotencyStore interface {
	// Get retrieves a cached response. Returns nil if not found or expired.
	Get(key string) *IdempotencyEntry
	// Set stores a response. The store must respect TTL.
	Set(key string, entry *IdempotencyEntry, ttl time.Duration)
	// SetProcessing marks a key as in-flight (to handle concurrent duplicates).
	// Returns false if the key is already being processed.
	SetProcessing(key string, ttl time.Duration) bool
	// Remove deletes a key (called on handler failure so retries work).
	Remove(key string)
}

// IdempotencyEntry holds a cached response.
type IdempotencyEntry struct {
	Status int
	Body   []byte
}

// DefaultIdempotencyConfig returns sensible defaults.
func DefaultIdempotencyConfig() IdempotencyConfig {
	return IdempotencyConfig{
		Header:  "Idempotency-Key",
		TTL:     24 * time.Hour,
		Methods: []string{http.MethodPost, http.MethodPut, http.MethodPatch},
	}
}

// ─── In-Memory Store ────────────────────────────────────────────────────────

type memIdempotencyEntry struct {
	entry     *IdempotencyEntry
	expiresAt time.Time
	pending   bool // true = handler is running, response not yet cached
}

type memIdempotencyStore struct {
	mu    sync.Mutex
	items map[string]*memIdempotencyEntry
}

func newMemIdempotencyStore(ttl time.Duration, stop <-chan struct{}) *memIdempotencyStore {
	s := &memIdempotencyStore{items: make(map[string]*memIdempotencyEntry)}
	if stop == nil {
		stop = make(chan struct{})
	}
	go func() {
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				s.mu.Lock()
				for k, v := range s.items {
					if now.After(v.expiresAt) {
						delete(s.items, k)
					}
				}
				s.mu.Unlock()
			case <-stop:
				return
			}
		}
	}()
	return s
}

func (s *memIdempotencyStore) Get(key string) *IdempotencyEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[key]
	if !ok || time.Now().After(e.expiresAt) || e.pending {
		return nil
	}
	return e.entry
}

func (s *memIdempotencyStore) Set(key string, entry *IdempotencyEntry, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = &memIdempotencyEntry{
		entry:     entry,
		expiresAt: time.Now().Add(ttl),
		pending:   false,
	}
}

func (s *memIdempotencyStore) SetProcessing(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[key]; ok && !time.Now().After(e.expiresAt) {
		return false // already exists (processing or completed)
	}
	s.items[key] = &memIdempotencyEntry{
		expiresAt: time.Now().Add(ttl),
		pending:   true,
	}
	return true
}

func (s *memIdempotencyStore) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// ─── Middleware ──────────────────────────────────────────────────────────────

// Idempotency returns a middleware that caches responses by idempotency key.
// Clients send a unique key via the Idempotency-Key header. If the same key
// is seen again within the TTL, the cached response is returned immediately
// without running the handler again.
//
// This is critical for payment APIs, financial systems, and any mutation
// endpoint where retries must be safe.
//
// Usage:
//
//	// Global: default config (in-memory store, 24h TTL)
//	server.Use(middleware.Idempotency())
//
//	// Per-route with required key
//	r.POST("/payments", handler, middleware.Idempotency(middleware.IdempotencyConfig{
//	    Required: true,
//	    TTL:      1 * time.Hour,
//	}))
//
//	// Multi-replica: plug in a Redis store
//	server.Use(middleware.Idempotency(middleware.IdempotencyConfig{
//	    Store: myRedisIdempotencyStore,
//	}))
func Idempotency(config ...IdempotencyConfig) core.MiddlewareFunc {
	cfg := DefaultIdempotencyConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Header == "" {
		cfg.Header = "Idempotency-Key"
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 24 * time.Hour
	}
	if len(cfg.Methods) == 0 {
		cfg.Methods = []string{http.MethodPost, http.MethodPut, http.MethodPatch}
	}
	if cfg.Store == nil {
		cfg.Store = newMemIdempotencyStore(cfg.TTL, cfg.Stop)
	}

	// Pre-build method lookup set.
	methodSet := make(map[string]struct{}, len(cfg.Methods))
	for _, m := range cfg.Methods {
		methodSet[m] = struct{}{}
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			// Only apply to configured methods.
			if _, ok := methodSet[c.Method()]; !ok {
				return next(c)
			}

			key := c.GetHeader(cfg.Header)
			if key == "" {
				if cfg.Required {
					return core.ErrBadRequest("missing " + cfg.Header + " header")
				}
				return next(c)
			}

			// Check for cached response.
			if entry := cfg.Store.Get(key); entry != nil {
				c.SetHeader("Idempotency-Replayed", "true")
				return c.SendBytes(entry.Status, entry.Body)
			}

			// Mark as processing (handles concurrent duplicate requests).
			if !cfg.Store.SetProcessing(key, cfg.TTL) {
				// Another request with the same key is in-flight.
				return core.NewHTTPError(http.StatusConflict,
					"a request with this idempotency key is already being processed")
			}

			// Run the handler.
			err := next(c)
			if err != nil {
				// On failure, remove the key so retries work.
				cfg.Store.Remove(key)
				return err
			}

			// Cache the successful response.
			cfg.Store.Set(key, &IdempotencyEntry{
				Status: c.ResponseStatus(),
				Body:   c.ResponseBody(),
			}, cfg.TTL)

			return nil
		}
	}
}
