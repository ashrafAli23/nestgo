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

	// TLS enables HTTPS. Both CertFile and KeyFile must be set.
	TLSCertFile string `json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file" yaml:"tls_key_file"`

	// Global features — applied to all routes across all controllers.
	// Execution order mirrors NestJS:
	//   Middleware → Filters → Guards → Pipes → Interceptors → Handler
	GlobalMiddlewares  []MiddlewareFunc  `json:"-" yaml:"-"`
	GlobalGuards       []Guard           `json:"-" yaml:"-"`
	GlobalPipes        []MiddlewareFunc  `json:"-" yaml:"-"`
	GlobalInterceptors []Interceptor     `json:"-" yaml:"-"`
	GlobalFilters      []ExceptionFilter `json:"-" yaml:"-"`

	// ValidateFunc is the global DTO validation function.
	// Called automatically by Body[T]() and QueryDTO[T]() after binding.
	// This is equivalent to NestJS's app.useGlobalPipes(new ValidationPipe()).
	//
	//   config.ValidateFunc = validator.Validate
	ValidateFunc func(v interface{}) error `json:"-" yaml:"-"`

	// RequestTimeout is the max duration for handler execution.
	// If set, a timeout middleware is applied globally.
	// Default: 0 (no timeout). Set to e.g. 30*time.Second for production.
	RequestTimeout time.Duration `json:"-" yaml:"-"`

	// Versioning configures API versioning strategy.
	Versioning *VersioningConfig `json:"-" yaml:"-"`

	// HealthCheck enables the built-in GET /health endpoint.
	// Returns 200 {"status":"ok"} when the server is alive.
	HealthCheck bool `json:"health_check" yaml:"health_check"`

	// ReadinessCheck is called by the GET /ready endpoint.
	// Return nil when the app is ready to serve traffic (DB connected, caches warm, etc.).
	// Return an error to respond with 503.
	// If nil and HealthCheck is true, /ready is not registered.
	ReadinessCheck func() error `json:"-" yaml:"-"`

	// ShutdownTimeout is how long to wait for in-flight requests to drain
	// before forcefully stopping. Default: 10 seconds.
	ShutdownTimeout time.Duration `json:"-" yaml:"-"`

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
