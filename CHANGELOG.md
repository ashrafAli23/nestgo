# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2026-04-09

### Added

#### Core (`core/`)

- `ResponseBody() []byte` on `Context` — read the written response body after handler execution; enables interceptors to log, transform, or cache response content
- `Get(key) interface{}` — simplified single-return context storage lookup (was `(interface{}, bool)`); returns `nil` when key is absent
- `ANY(path, handler, mw...)` on `Router` — registers a handler for all HTTP methods; implemented via Gin's `Any()` and Fiber's `All()`
- `StartTLS(addr, certFile, keyFile)` on `Server` — start HTTPS server; DI container auto-selects TLS vs plain based on `Config.TLSCertFile`/`TLSKeyFile`
- `FieldError` — structured per-field error (`field`, `message`, `tag`, `value`)
- `NewValidationError(errors ...FieldError)` — construct 422 with structured field errors
- `ProblemDetail` — RFC 7807 `application/problem+json` response type
- `ProblemDetailErrorHandler` — drop-in `ErrorHandler` that outputs RFC 7807 format
- `RouteOptions.Pipes` — global and per-route pipes now slot between Guards and Interceptors, matching NestJS execution order

#### Config (`core/Config`)

- `TLSCertFile` / `TLSKeyFile` — enable HTTPS; both must be set to activate TLS
- `ShutdownTimeout` — drain in-flight requests before forceful stop (default: 10s)
- `HealthCheck bool` — auto-registers `GET /health` returning `{"status":"ok"}`
- `ReadinessCheck func() error` — auto-registers `GET /ready`; 503 on error, 200 on nil
- `GlobalMiddlewares []MiddlewareFunc` — applied first (outermost) on all routes
- `GlobalPipes []MiddlewareFunc` — applied between Guards and Interceptors on all routes
- `ValidateFunc func(v interface{}) error` — equivalent to NestJS `useGlobalPipes(ValidationPipe)`; called by `Body[T]()` and `QueryDTO[T]()` automatically

#### DI (`di/`)

- **Graceful OS signal handling** — `StartServer` now spawns a goroutine that listens for `SIGINT`/`SIGTERM` and triggers `fx.Shutdown`, ensuring `OnShutdown` hooks and in-flight request draining run on Kubernetes/container termination
- `RegisterHealthEndpoints` — wired into `CoreModule`; auto-registers health/readiness routes when enabled in config

#### Middleware (`middleware/`)

- `Tracing()` — W3C `traceparent` header propagation; extracts/generates `trace_id` and `span_id`, stores in context, sets `X-Trace-ID` response header. Works standalone or alongside the OTel SDK
- `ETag()` — generates SHA-256 ETags from response bodies; handles `If-None-Match` → 304 Not Modified. Configurable weak/strong. Zero allocations for non-matching requests
- `Idempotency()` — caches responses by `Idempotency-Key` header; replays cached response on retry. Pluggable `IdempotencyStore` interface for Redis in multi-replica deployments. Detects concurrent duplicates with 409 Conflict
- `BodyLimit()` — per-route request body size limit; overrides global `Config.BodyLimit`. Checks `Content-Length` header fast-path before body read
- `Upload()` — validates file uploads: max size, allowed MIME types (via content sniffing, not spoofable header), allowed extensions. Stores `*UploadedFile` in context. Includes `SaveFile()` helper

### Changed

#### Core (`core/`)

- `DefaultErrorHandler` now includes an `errors` field in the response body when `HTTPError.Errors` is non-empty — no breaking change to existing responses
- `RouteOptions` gains `Pipes []MiddlewareFunc` field; `ApplyRouteOptions` slots them between Guards and Interceptors

#### Config (`core/Config`)

- `GlobalMiddlewares` added before `GlobalGuards` in DI router setup, establishing the correct outermost position in the execution chain

#### Middleware (`middleware/`)

- `CORS()` — new `ExposeHeaders []string` field; sets `Access-Control-Expose-Headers` when non-empty (e.g. `X-Total-Count` for pagination)
- `Logger()` — auto-includes `request_id` and `trace_id` fields from context when present (correlates logs with `RequestID()` and `Tracing()` middleware)
- `Recovery()` — default `LogFunc` now uses `core.Log()` instead of `fmt.Printf`

#### DI / Adapters

- All `fmt.Printf` / `fmt.Println` calls in framework internals replaced with `core.Log()` — the pluggable logger is now used throughout, including server start/stop messages and lifecycle hook notifications
- **Fiber adapter:** `Debug` config flag now controls `EnablePrintRoutes` and `DisableStartupMessage` in `fiber.Config`

---

## [1.2.0] - 2026-04-06

### Added

#### Core (`core/`)

- `ResponseStatus() int` on `Context` interface — read the HTTP response status code after handler execution, enabling logging middleware and metrics interceptors to observe response outcomes
- `ParamInt64(c, key)` — extract path parameters as `int64` for database primary keys
- `PInt64(key)` extractor — use with Handle functions: `core.Handle1(core.PInt64("id"), ctrl.GetByID)`
- `RCtx()` extractor — extract `context.Context` from request for passing to DB queries, gRPC clients, and HTTP calls with automatic cancellation propagation
- `SetValidateFunc(fn)` — register a global validation function that `Body[T]()` and `QueryDTO[T]()` call automatically after binding, bridging `nestgo-validator` or any validation library without implementing `Validatable` on every DTO

#### Middleware (`middleware/`)

- `Logger()` — request logging middleware with method, path, status code, duration, and client IP. Supports `SkipFunc` for excluding routes (e.g. health checks) and `LogFunc` for custom structured output (slog, zap, zerolog)

### Changed

#### Middleware (`middleware/`)

- **Breaking fix:** `Timeout()` middleware rewritten to use `context.WithTimeout` instead of goroutine + `Clone()`. The previous implementation ran the handler on a cloned context whose response methods were no-ops (Fiber adapter), silently discarding responses. The new implementation sets a deadline on the request context — all downstream I/O (database queries, HTTP clients, gRPC calls) cancel automatically when the deadline fires. No goroutine races, no clone issues, works identically on both Gin and Fiber adapters.

### Fixed

- Timeout middleware no longer silently drops handler responses on the Fiber adapter

---

## [1.1.0] - 2026-04-05

### Added

#### Core (`core/`)

- `Server`, `Router`, `Context` interfaces — adapter-agnostic HTTP abstraction
- `HandlerFunc` and `MiddlewareFunc` types
- `Controller`, `PrefixedController`, `VersionedController` interfaces
- `Config` with global prefix, guards, interceptors, filters, versioning, shutdown hooks
- `Guard` interface + `GuardFunc` + `UseGuards()` — request authorization
- `Interceptor` interface + `InterceptorFunc` + `UseInterceptors()` — before/after handler logic
- `Pipe[T]` interface + `PipeFunc[T]` + `WithPipes()` — value transformation/validation
- `ExceptionFilter` interface + `UseFilters()` + `HTTPErrorFilter` — custom error handling
- `Extractor[T]` with `B`, `P`, `PInt`, `Q`, `QInt`, `QDto`, `H`, `Ctx` extractors
- `Handle1` through `Handle4` + `HandleC1` through `HandleC3` — type-safe handler builders
- `Body[T]`, `QueryDTO[T]`, `Param`, `ParamInt`, `Query`, `QueryInt`, `Header` — binding helpers
- `Validatable` interface — auto-validation on extraction
- `Required`, `MinLength`, `MaxLength`, `InRange` validation helpers
- `RouteOptions` + `ApplyRouteOptions()` — correct NestJS execution order
- `WithMeta()` / `Meta()` — route metadata
- `VersioningConfig` with URI, Header, MediaType strategies + `VersionGuard`
- `SSEEvent`, `SSEStream`, `NewSSEStream()`, `SSE()` — Server-Sent Events
- `IsWebSocketRequest()` — WebSocket detection helper
- `HTTPError` + constructors (`ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrConflict`, `ErrUnprocessable`, `ErrInternalServer`)
- `DefaultErrorHandler` — JSON error responses
- `Logger` interface + `Field` + `F()` + `SetLogger()` + `Log()` + `LoggerFunc`
- `OnModuleInit` / `OnModuleDestroy` lifecycle interfaces
- `Context.Clone()` — goroutine-safe context copy

#### DI (`di/`)

- `NewApp()` — application builder (config + adapter + modules)
- `ServerProvider` type — adapter-agnostic server creation
- `AsController()` — auto-register controllers via DI
- `AsInitHook()` / `AsDestroyHook()` — lifecycle hook registration
- `CoreModule` — bundles controller registration + server lifecycle
- `ConfigModule()` — provides config to DI container
- Global prefix, guards, interceptors, filters applied via DI router provider
- Graceful shutdown with `OnShutdown` hooks

#### Middleware (`middleware/`)

- `Recovery()` — panic recovery with pooled stack trace buffers
- `CORS()` — Cross-Origin Resource Sharing with O(1) origin lookup
- `Helmet()` — OWASP security headers
- `RateLimit()` — per-key rate limiting with 32-shard map
- `RequestID()` — unique request ID with pooled generators
- `Timeout()` — request deadline with stoppable timers
- `CSRF()` — CSRF protection with double-submit cookie pattern
- `Compress()` / `GzipJSON()` — gzip response compression
