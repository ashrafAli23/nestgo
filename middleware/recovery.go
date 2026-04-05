package middleware

import (
	"fmt"
	"runtime"
	"sync"

	core "github.com/ashrafAli23/nestgo/core"
)

// RecoveryConfig holds panic recovery configuration.
type RecoveryConfig struct {
	// StackSize is the max stack trace size in bytes. Default: 8192.
	// Go stack traces in real apps often exceed 4KB. 8KB covers most cases.
	StackSize int
	// EnableStackTrace includes stack trace in the error details. Default: true in debug, false in production.
	EnableStackTrace bool
	// LogFunc is called when a panic is recovered. Default: fmt.Printf.
	LogFunc func(c core.Context, err interface{}, stack string)
	// ErrorHandler is called to produce the error response. If set, overrides
	// the default 500 response behavior. This allows users to customize the
	// error format (e.g. JSON API errors, localized messages).
	// Receives the recovered value and the stack trace string.
	ErrorHandler func(c core.Context, recovered interface{}, stack string) error
}

// DefaultRecoveryConfig returns the default recovery config.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		StackSize:        8192,
		EnableStackTrace: false,
		LogFunc: func(c core.Context, err interface{}, stack string) {
			fmt.Printf("[NestGo] PANIC RECOVERED: %v\n%s\n", err, stack)
		},
	}
}

// Recovery returns a middleware that recovers from panics in handlers.
// Without this, a panic in any handler crashes the entire server.
//
// Usage:
//
//	server.Use(middleware.Recovery())
//
// Always register this as the FIRST middleware so it wraps everything.
func Recovery(config ...RecoveryConfig) core.MiddlewareFunc {
	cfg := DefaultRecoveryConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.StackSize <= 0 {
		cfg.StackSize = 8192
	}
	if cfg.LogFunc == nil {
		cfg.LogFunc = func(c core.Context, err interface{}, stack string) {
			fmt.Printf("[NestGo] PANIC RECOVERED: %v\n%s\n", err, stack)
		}
	}

	// Pool stack trace buffers to avoid allocation per panic recovery.
	stackPool := sync.Pool{
		New: func() interface{} {
			b := make([]byte, cfg.StackSize)
			return &b
		},
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		// Pre-check: does the panic path need the stack as a string?
		// If no stack trace in response and no custom error handler, we only
		// need the string for LogFunc — we can build it once and avoid
		// redundant conversions.
		needsStack := cfg.EnableStackTrace || cfg.ErrorHandler != nil

		return func(c core.Context) (returnErr error) {
			defer func() {
				if r := recover(); r != nil {
					// Capture stack trace using pooled buffer
					bp := stackPool.Get().(*[]byte)
					buf := *bp
					n := runtime.Stack(buf, false)

					// Convert to string once — only when needed beyond logging.
					// In production (EnableStackTrace=false, no ErrorHandler),
					// this is the only allocation on the panic path.
					stack := string(buf[:n])
					stackPool.Put(bp)

					// Log the panic
					cfg.LogFunc(c, r, stack)

					// Custom error handler takes full control if set
					if cfg.ErrorHandler != nil {
						returnErr = cfg.ErrorHandler(c, r, stack)
						return
					}

					// Preserve typed errors: if user panicked with an error or
					// *HTTPError, use its status code and message instead of
					// always returning a generic 500.
					if !needsStack {
						// Fast path: no stack in response — avoid map allocation
						switch v := r.(type) {
						case *core.HTTPError:
							returnErr = v
						case error:
							returnErr = core.ErrInternalServer(v.Error())
						default:
							returnErr = core.ErrInternalServer("internal server error")
						}
						return
					}

					switch v := r.(type) {
					case *core.HTTPError:
						returnErr = core.NewHTTPErrorWithDetails(v.Code, v.Message, map[string]interface{}{
							"stack": stack,
						})
					case error:
						returnErr = core.NewHTTPErrorWithDetails(500, v.Error(), map[string]interface{}{
							"stack": stack,
						})
					default:
						returnErr = core.NewHTTPErrorWithDetails(500, "internal server error", map[string]interface{}{
							"panic": fmt.Sprintf("%v", r),
							"stack": stack,
						})
					}
				}
			}()

			return next(c)
		}
	}
}
