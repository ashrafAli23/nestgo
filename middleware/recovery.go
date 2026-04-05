package middleware

import (
	"fmt"
	"runtime"
	"sync"

	core "github.com/ashrafAli23/nestgo/core"
)

// RecoveryConfig holds panic recovery configuration.
type RecoveryConfig struct {
	// StackSize is the max stack trace size in bytes. Default: 4096.
	StackSize int
	// EnableStackTrace includes stack trace in the error details. Default: true in debug, false in production.
	EnableStackTrace bool
	// LogFunc is called when a panic is recovered. Default: fmt.Printf.
	LogFunc func(c core.Context, err interface{}, stack string)
}

// DefaultRecoveryConfig returns the default recovery config.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		StackSize:        4096,
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
		cfg.StackSize = 4096
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
		return func(c core.Context) (returnErr error) {
			defer func() {
				if r := recover(); r != nil {
					// Capture stack trace using pooled buffer
					bp := stackPool.Get().(*[]byte)
					buf := *bp
					n := runtime.Stack(buf, false)
					stack := string(buf[:n])
					stackPool.Put(bp)

					// Log the panic
					cfg.LogFunc(c, r, stack)

					// Return 500 error
					if cfg.EnableStackTrace {
						returnErr = core.NewHTTPErrorWithDetails(
							500,
							"internal server error",
							map[string]interface{}{
								"panic": fmt.Sprintf("%v", r),
								"stack": stack,
							},
						)
					} else {
						returnErr = core.ErrInternalServer("internal server error")
					}
				}
			}()

			return next(c)
		}
	}
}
