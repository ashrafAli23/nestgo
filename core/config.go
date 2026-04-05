package core

import (
	"context"
	"time"
)

// Config holds framework-level configuration.
type Config struct {
	AppName       string       `json:"app_name" yaml:"app_name"`
	Addr          string       `json:"addr" yaml:"addr"`
	Adapter       string       `json:"adapter" yaml:"adapter"`
	GlobalPrefix  string       `json:"global_prefix" yaml:"global_prefix"`
	BodyLimit     int          `json:"body_limit" yaml:"body_limit"`
	ReadTimeout   int          `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout  int          `json:"write_timeout" yaml:"write_timeout"`
	Debug         bool         `json:"debug" yaml:"debug"`
	DisableLogger bool         `json:"disable_logger" yaml:"disable_logger"`
	ErrorHandler  ErrorHandler `json:"-" yaml:"-"`

	// Global features — applied to all routes across all controllers.
	GlobalGuards       []Guard           `json:"-" yaml:"-"`
	GlobalInterceptors []Interceptor     `json:"-" yaml:"-"`
	GlobalFilters      []ExceptionFilter `json:"-" yaml:"-"`

	// RequestTimeout is the max duration for handler execution.
	// If set, a timeout middleware is applied globally.
	// Default: 0 (no timeout). Set to e.g. 30*time.Second for production.
	RequestTimeout time.Duration `json:"-" yaml:"-"`

	// Versioning configures API versioning strategy.
	Versioning *VersioningConfig `json:"-" yaml:"-"`

	// OnShutdown hooks run before the server stops.
	// Use for cleanup: close DB pools, flush queues, etc.
	OnShutdown []func(ctx context.Context) error `json:"-" yaml:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		AppName:      "NestGo App",
		Addr:         ":8080",
		Adapter:      "gin",
		BodyLimit:    4 * 1024 * 1024,
		ReadTimeout:  10,
		WriteTimeout: 10,
		Debug:        false,
		ErrorHandler: DefaultErrorHandler,
	}
}
