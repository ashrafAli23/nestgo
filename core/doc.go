// Package core defines the interfaces and types for the NestGo framework.
//
// It has ZERO external dependencies. Your app code imports only this package
// plus one adapter. Adapters (Gin, Fiber) implement [Server], [Router], and
// [Context] so you can swap HTTP engines without changing a single handler.
//
// # Install
//
//	go get github.com/ashrafAli23/nestgo
//
// Then pick ONE adapter (separate modules):
//
//	go get github.com/ashrafAli23/nestgo-gin-adapter    # Gin
//	go get github.com/ashrafAli23/nestgo-fiber-adapter  # Fiber
//
// # Architecture
//
// The core package sits at the center of the framework:
//
//	┌─────────────┐    ┌──────────────┐    ┌──────────────────┐
//	│  core        │    │  di           │    │  middleware       │
//	│  (interfaces │    │  (uber/fx DI) │    │  (CORS, Helmet,  │
//	│   zero deps) │    │              │    │   RateLimit, ...) │
//	└──────┬───────┘    └──────┬───────┘    └────────┬─────────┘
//	       │                   │                     │
//	       └───────────────────┼─────────────────────┘
//	                           │
//	       ┌───────────────────┴───────────────────┐
//	       │                                       │
//	  ┌────┴──────────┐                  ┌─────────┴───────┐
//	  │ gin adapter   │                  │ fiber adapter    │
//	  └───────────────┘                  └─────────────────┘
//
// # Core Concepts
//
// [Server] is the top-level interface (embeds [Router]). Adapters implement it.
//
// [Router] registers routes with [HandlerFunc] and optional [MiddlewareFunc].
//
// [Context] abstracts the HTTP request/response. Every handler and middleware
// receives a Context — never a framework-specific type.
//
// [Controller] groups related routes. The DI system auto-registers controllers.
//
// # NestJS-Inspired Features
//
// [Guard] — decides if a request proceeds (auth, RBAC). See [UseGuards].
//
// [Interceptor] — wraps handler execution (logging, caching). See [UseInterceptors].
//
// [Pipe] — transforms/validates extracted values. See [WithPipes].
//
// [ExceptionFilter] — catches errors and formats responses. See [UseFilters].
//
// [Extractor] + [Handle1] through [Handle4] — type-safe parameter injection,
// the Go equivalent of NestJS decorators (@Body, @Param, @Query).
//
// [RouteOptions] + [ApplyRouteOptions] — bundles Guards, Interceptors, and Filters
// in the correct execution order.
//
// # Real-Time
//
// [SSE] streams Server-Sent Events via a channel-based [SSEStream].
//
// WebSocket support is adapter-specific via [Context.Underlying].
// See [IsWebSocketRequest] for detection.
package core
