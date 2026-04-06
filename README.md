<p align="center">
  <h1 align="center">NestGo</h1>
  <p align="center">NestGo is a powerful and flexible web framework for Go (Golang) designed for building scalable and maintainable server-side applications. Inspired by <a href="https://nestjs.com" title="Visit NestJS official website">NestJS</a>
, NestGo stays true to Go’s philosophy of simplicity, performance, and explicitness while offering a modular architecture, dependency injection, and an adapter-based design. It is ideal for building REST APIs, microservices, and production-ready backend systems.</p>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/ashrafAli23/nestgo"><img src="https://pkg.go.dev/badge/github.com/ashrafAli23/nestgo.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/ashrafAli23/nestgo"><img src="https://goreportcard.com/badge/github.com/ashrafAli23/nestgo" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

---

NestGo brings the best ideas from [NestJS](https://nestjs.com) to Go: **Guards**, **Interceptors**, **Pipes**, **Exception Filters**, **Dependency Injection**, and a clean **Controller** pattern. Write your handlers once, swap HTTP engines (Gin or Fiber) by changing one line.

## Features

- **Zero-dep core** — interfaces only, no framework lock-in
- **Adapter pattern** — swap Gin/Fiber without changing handlers
- **Dependency injection** — uber/fx powered, module-based
- **Guards** — auth/RBAC checks before handlers run
- **Interceptors** — before/after logic (logging, caching, timing)
- **Pipes** — transform/validate extracted values
- **Exception filters** — custom error formatting per route or controller
- **Type-safe extractors** — `@Body()`, `@Param()`, `@Query()` equivalents via generics
- **API versioning** — URI, Header, or Media Type strategies
- **SSE support** — channel-based Server-Sent Events
- **Middleware ecosystem** — CORS, Helmet, Rate Limit, CSRF, Timeout, Recovery, Request ID, Compress
- **Lifecycle hooks** — OnModuleInit / OnModuleDestroy + graceful shutdown
- **Structured logger** — plug in zerolog, slog, zap, or any logger

## Install

```bash
go get github.com/ashrafAli23/nestgo
```

Pick one adapter (separate modules, installed independently):

```bash
go get github.com/ashrafAli23/nestgo-gin-adapter    # Gin
go get github.com/ashrafAli23/nestgo-fiber-adapter   # Fiber
```

Optional validation package:

```bash
go get github.com/ashrafAli23/nestgo-validator
```

## Quick Start

```go
package main

import (
    "github.com/ashrafAli23/nestgo/core"
    "github.com/ashrafAli23/nestgo/di"
    "github.com/ashrafAli23/nestgo/middleware"
    ginadapter "github.com/ashrafAli23/nestgo-gin-adapter"
    "go.uber.org/fx"
)

// ─── DTO ────────────────────────────────────────────────────────────────────

type CreateUserDTO struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (d *CreateUserDTO) Validate() error {
    if d.Name == "" {
        return core.ErrBadRequest("name is required")
    }
    return nil
}

// ─── Controller ─────────────────────────────────────────────────────────────

type UserController struct{}

func NewUserController() *UserController { return &UserController{} }

func (c *UserController) Prefix() string { return "/users" }

func (c *UserController) RegisterRoutes(r core.Router) {
    r.GET("/", core.Handle1(core.Q("search", ""), c.List))
    r.GET("/:id", core.Handle1(core.PInt("id"), c.GetByID))
    r.POST("/", core.HandleC1(core.B[CreateUserDTO](), c.Create))
}

func (c *UserController) List(search string) (any, error) {
    return []map[string]any{{"name": "John"}, {"name": "Jane"}}, nil
}

func (c *UserController) GetByID(id int) (any, error) {
    return map[string]any{"id": id, "name": "John"}, nil
}

func (c *UserController) Create(ctx core.Context, dto *CreateUserDTO) error {
    return ctx.JSON(201, map[string]any{"name": dto.Name, "email": dto.Email})
}

// ─── Main ───────────────────────────────────────────────────────────────────

func main() {
    config := core.DefaultConfig()
    config.Addr = ":3000"
    config.GlobalPrefix = "/api"

    app := di.NewApp(config, ginadapter.New,
        fx.Invoke(func(server core.Server) {
            server.Use(middleware.Recovery())
            server.Use(middleware.RequestID())
            server.Use(middleware.Helmet())
            server.Use(middleware.CORS())
        }),
        fx.Module("users",
            fx.Provide(di.AsController(NewUserController)),
        ),
    )
    app.Run()
}
```

```
GET  /api/users          → List users
GET  /api/users/:id      → Get user by ID
POST /api/users          → Create user (201)
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Your App (main.go)                      │
│  config + adapter + modules → di.NewApp() → app.Run()       │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ nestgo/core   │  │ nestgo/di    │  │ nestgo/middleware │  │
│  │ Interfaces    │  │ DI Container │  │ CORS, Helmet,    │  │
│  │ only.         │  │ (uber/fx)    │  │ RateLimit, etc.  │  │
│  │ Zero deps.    │  │              │  │                  │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
│         └──────────────────┼───────────────────┘            │
│                            │                                │
│         ┌──────────────────┴──────────────────┐             │
│         │                                     │             │
│   ┌─────┴────────────┐            ┌───────────┴──────────┐  │
│   │ nestgo-gin-adapter│            │ nestgo-fiber-adapter  │  │
│   │ (separate module) │            │ (separate module)     │  │
│   └──────────────────┘            └──────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Key principle:** The `core` package has zero external dependencies. Adapters implement `Server`, `Router`, and `Context`. Swap adapters by changing one line in `main.go`.

## Core Concepts

### Server, Router, Context

```go
// Server is the top-level app — embeds Router
type Server interface {
    Router
    Start(addr string) error
    Shutdown(ctx context.Context) error
}

// Router registers routes
type Router interface {
    GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
    POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
    // PUT, DELETE, PATCH, OPTIONS, HEAD...
    Group(prefix string, middleware ...MiddlewareFunc) Router
    Use(middleware ...MiddlewareFunc)
}

// Context abstracts HTTP request/response — both adapters implement this
type Context interface {
    // Request: Method, Path, Param, Query, Body, Bind, GetHeader, Cookie, ...
    // Response: JSON, XML, String, SendBytes, SendStream, SetHeader, Redirect, ...
    // Storage: Set, Get
    // Safety: Clone
}
```

### Controller

```go
type Controller interface {
    RegisterRoutes(router Router)
}

// Optional interfaces:
type PrefixedController interface {
    Prefix() string  // auto-prefixes all routes
}

type VersionedController interface {
    Version() string  // auto-applies versioning
}
```

### Handler & Middleware

```go
type HandlerFunc func(Context) error
type MiddlewareFunc func(HandlerFunc) HandlerFunc
```

## Type-Safe Extractors

NestGo's equivalent of NestJS decorators. Extract typed values from requests with compile-time safety:

| Extractor | NestJS Equivalent | Returns | Description |
|-----------|-------------------|---------|-------------|
| `B[T]()` | `@Body()` | `*T` | Parse & validate body |
| `P(key)` | `@Param('key')` | `string` | Path parameter |
| `PInt(key)` | `@Param('key', ParseIntPipe)` | `int` | Path param as int |
| `Q(key, default...)` | `@Query('key')` | `string` | Query parameter |
| `QInt(key, default...)` | `@Query('key', ParseIntPipe)` | `int` | Query param as int |
| `QDto[T]()` | `@Query() dto` | `*T` | Query params into struct |
| `H(key, required)` | `@Headers('key')` | `string` | Request header |
| `Ctx()` | &mdash; | `Context` | Raw context |

### Handle Builders

```go
// Framework extracts, calls your function, auto-responds with JSON 200:
r.GET("/:id", core.Handle1(core.PInt("id"), ctrl.GetByID))

// Two extractors:
r.POST("/:id", core.Handle2(core.PInt("id"), core.B[UpdateDTO](), ctrl.Update))

// Need custom status code? Use HandleC variants (Context as first arg):
r.POST("/", core.HandleC1(core.B[CreateDTO](), ctrl.Create))
// ctrl.Create = func(c core.Context, dto *CreateDTO) error { return c.JSON(201, ...) }
```

## Guards

Decide whether a request proceeds. Runs **before** the handler.

```go
authGuard := core.GuardFunc(func(c core.Context) (bool, error) {
    token := c.GetHeader("Authorization")
    if token == "" {
        return false, core.ErrUnauthorized("missing token")
    }
    return true, nil
})

// Three levels:
config.GlobalGuards = []core.Guard{authGuard}                    // global
group := r.Group("/admin", core.UseGuards(adminGuard))           // per-controller
r.DELETE("/:id", handler, core.UseGuards(ownerGuard))            // per-route
```

## Interceptors

Wrap handler execution — run code **before and after**.

```go
timing := core.InterceptorFunc(func(c core.Context, next core.HandlerFunc) error {
    start := time.Now()
    err := next(c)
    core.Log().Info("request", core.F("duration", time.Since(start)))
    return err
})

r.Group("/api", core.UseInterceptors(timing))
```

## Pipes

Transform or validate values **after extraction, before the handler**.

```go
r.GET("/:id", core.Handle1(
    core.WithPipes(core.PInt("id"), core.IntRangePipe("id", 1, 999999)),
    ctrl.GetByID,
))
```

Built-in: `TrimPipe`, `NonEmptyPipe(name)`, `IntRangePipe(name, min, max)`.

## Exception Filters

Catch errors and format custom responses.

```go
type APIErrorFilter struct{}

func (f *APIErrorFilter) CanHandle(err error) bool {
    _, ok := err.(*core.HTTPError)
    return ok
}

func (f *APIErrorFilter) Handle(c core.Context, err error) {
    httpErr := err.(*core.HTTPError)
    c.JSON(httpErr.Code, map[string]any{"error": httpErr.Message})
}

r.Group("/api", core.UseFilters(&APIErrorFilter{}))
```

## RouteOptions

Bundle guards + interceptors + filters in the correct execution order:

```go
opts := core.ApplyRouteOptions(core.RouteOptions{
    Guards:       []core.Guard{authGuard},
    Interceptors: []core.Interceptor{logInterceptor},
    Filters:      []core.ExceptionFilter{apiFilter},
})
r.Group("/users", opts...)
```

Execution order: **Filters (outermost) -> Guards -> Interceptors -> Handler**.

## Route Metadata

Attach data for guards/interceptors to read:

```go
r.DELETE("/:id", handler,
    core.WithMeta("roles", []string{"admin"}),
    core.UseGuards(roleGuard),
)

// In the guard:
roles, _ := c.Get("roles")
```

## API Versioning

Three strategies:

```go
// URI: /v1/users, /v2/users
config.Versioning = &core.VersioningConfig{Strategy: core.URIVersioning}

// Header: Accept-Version: 1
config.Versioning = &core.VersioningConfig{Strategy: core.HeaderVersioning}

// Media Type: Accept: application/vnd.api.v1+json
config.Versioning = &core.VersioningConfig{Strategy: core.MediaTypeVersioning}
```

Controller declares its version:

```go
func (c *UserControllerV2) Version() string { return "2" }
```

## Server-Sent Events

```go
func (ctrl *Controller) Events(c core.Context) error {
    stream := core.NewSSEStream(10)
    go func() {
        defer close(stream)
        for i := 0; i < 5; i++ {
            stream <- core.SSEEvent{Event: "tick", Data: fmt.Sprintf("%d", i)}
            time.Sleep(time.Second)
        }
    }()
    return core.SSE(c, stream)
}
```

## Middleware

All middleware follows the same config pattern: defaults + optional customization.

| Middleware | Description |
|-----------|-------------|
| `Recovery()` | Panic recovery (register first) |
| `CORS()` | Cross-Origin Resource Sharing |
| `Helmet()` | OWASP security headers |
| `RateLimit()` | Per-key rate limiting (sharded, auto-cleanup) |
| `RequestID()` | Unique request ID per request |
| `Timeout()` | Request execution deadline |
| `CSRF()` | Cross-Site Request Forgery protection |
| `Compress()` / `GzipJSON()` | Gzip response compression |

```go
// Defaults:
server.Use(middleware.Recovery())
server.Use(middleware.CORS())

// Custom:
server.Use(middleware.RateLimit(middleware.RateLimitConfig{
    Max:    200,
    Window: 5 * time.Minute,
}))

// Per-route:
r.POST("/upload", handler, middleware.Timeout(middleware.TimeoutConfig{
    Timeout: 60 * time.Second,
}))
```

## Lifecycle Hooks

```go
// Service with startup/shutdown hooks:
type DBService struct{ pool *pgxpool.Pool }

func (s *DBService) OnModuleInit(ctx context.Context) error {
    // Connect to DB
    return nil
}

func (s *DBService) OnModuleDestroy(ctx context.Context) error {
    s.pool.Close()
    return nil
}

// Register:
fx.Provide(di.AsInitHook(NewDBService))
fx.Provide(di.AsDestroyHook(NewDBService))
```

Graceful shutdown hooks:

```go
config.OnShutdown = []func(ctx context.Context) error{
    func(ctx context.Context) error { db.Close(); return nil },
    func(ctx context.Context) error { cache.Flush(); return nil },
}
```

## Logger

Pluggable structured logging:

```go
// Use the default (stderr):
core.Log().Info("started", core.F("addr", ":3000"))

// Plug in your own (zerolog, slog, zap):
core.SetLogger(myZerologAdapter)
```

## Error Handling

```go
// Built-in HTTP errors:
return core.ErrBadRequest("name is required")     // 400
return core.ErrUnauthorized("invalid token")       // 401
return core.ErrForbidden("access denied")          // 403
return core.ErrNotFound("user not found")          // 404

// Custom:
return core.NewHTTPError(429, "too many requests")
return core.NewHTTPErrorWithDetails(422, "validation failed", details)

// Custom global error handler:
config.ErrorHandler = func(c core.Context, err error) {
    // your logic
}
```

## Swapping Adapters

Change one line:

```go
// Gin:
app := di.NewApp(config, ginadapter.New, ...)

// Fiber:
app := di.NewApp(config, fiberadapter.New, ...)
```

All handlers, middleware, guards, interceptors, pipes, and filters work identically on both adapters. Use `c.Underlying()` only when you need adapter-specific features (WebSocket upgrade, etc.).

## Context Safety

Both adapters pool contexts. If passing to a goroutine, **always clone**:

```go
// WRONG — context is recycled after handler returns:
go func() { doWork(c) }()

// CORRECT:
go func(ctx core.Context) { doWork(ctx) }(c.Clone())
```

## Performance

NestGo applies optimizations automatically:

- **Context pooling** — zero alloc per request for context structs
- **Sharded rate limiter** — 32 shards reduce lock contention
- **Pre-computed headers** — CORS, Helmet strings built once at init
- **Pooled buffers** — token generation, stack traces, body cache
- **Atomic guards** — Fiber released-flag uses atomic.Bool, not mutex
- **Reusable timers** — Timeout middleware uses stoppable timers

## Project Structure

```
nestgo/
├── core/           # Interfaces & types (zero deps)
│   ├── server.go, router.go, context.go
│   ├── guard.go, interceptor.go, pipe.go, filter.go
│   ├── handle.go, bind.go, config.go, errors.go
│   ├── sse.go, websocket.go, versioning.go
│   ├── lifecycle.go, logger.go, meta.go, route.go
│   └── example_test.go
├── di/             # DI container (uber/fx)
│   ├── container.go
│   └── lifecycle.go
├── middleware/     # HTTP middleware (stdlib only)
│   ├── cors.go, helmet.go, ratelimit.go
│   ├── requestid.go, timeout.go, csrf.go
│   ├── recovery.go, compress.go
│   └── doc.go
├── go.mod
└── README.md
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
