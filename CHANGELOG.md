# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-04-05

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
