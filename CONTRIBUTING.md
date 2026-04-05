# Contributing to NestGo

Thank you for your interest in contributing to NestGo! This guide will help you get started.

## Development Setup

```bash
git clone https://github.com/ashrafAli23/nestgo.git
cd nestgo
go mod tidy
go build ./...
go test ./...
```

## Project Structure

```
nestgo/
├── core/           # Interfaces & types (ZERO external deps)
├── di/             # DI container (depends on uber/fx)
├── middleware/     # HTTP middleware (depends on core/ only)
├── go.mod
└── README.md
```

Adapters live in separate repositories:
- `nestgo-gin-adapter` — Gin implementation
- `nestgo-fiber-adapter` — Fiber implementation
- `nestgo-validator` — Validation (wraps go-playground/validator)

## Design Principles

1. **core/ has zero external dependencies.** It contains only interfaces, types, and pure Go functions. Never add third-party imports to `core/`.

2. **Adapters implement core interfaces.** `Server`, `Router`, and `Context` are interfaces in core — adapters provide the concrete implementation. The core package never imports an adapter.

3. **Middleware depends only on core.** Every middleware receives a `core.Context` and returns a `core.MiddlewareFunc`. No adapter-specific code.

4. **Config struct with defaults pattern.** Every configurable component follows:
   ```go
   type XxxConfig struct { /* fields */ }
   func DefaultXxxConfig() XxxConfig { /* sensible defaults */ }
   func Xxx(config ...XxxConfig) core.MiddlewareFunc { /* ... */ }
   ```

5. **Guards, Interceptors, Filters all compile to MiddlewareFunc.** This keeps the router interface simple.

## How to Add Things

### Adding a New Middleware

1. Create `middleware/yourname.go`
2. Define `YournameConfig` struct with exported fields
3. Define `DefaultYournameConfig()` with sensible defaults
4. Define `Yourname(config ...YournameConfig) core.MiddlewareFunc`
5. Add a doc comment with usage example
6. Update `middleware/doc.go` list

### Adding a New Extractor

1. Add to `core/handle.go`
2. Return an `Extractor[T]`:
   ```go
   func Cookie(name string) Extractor[string] {
       return Extractor[string]{
           Extract: func(c Context) (string, error) {
               val := c.Cookie(name)
               if val == "" {
                   return "", ErrBadRequest("cookie required")
               }
               return val, nil
           },
       }
   }
   ```
3. Add an example to `core/example_test.go`

### Adding a New Pipe

1. Add to `core/pipe.go`
2. Return a `PipeFunc[T]`:
   ```go
   var ToLowerPipe = PipeFunc[string](func(val string, c Context) (string, error) {
       return strings.ToLower(val), nil
   })
   ```

### Adding a New Guard

1. Implement `Guard` or use `GuardFunc`:
   ```go
   var IPWhitelistGuard = core.GuardFunc(func(c core.Context) (bool, error) {
       return allowedIPs[c.ClientIP()], nil
   })
   ```

### Adding a New Adapter

1. Create a separate module (e.g., `nestgo-echo-adapter`)
2. Implement `core.Server` (which embeds `core.Router`)
3. Implement `core.Context`
4. Export `func New(config *core.Config) core.Server`
5. Users swap by changing one line: `di.NewApp(config, echoadapter.New, ...)`

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Every exported type and function must have a doc comment
- Use `interface{}` (not `any`) in core interfaces for Go 1.18 compatibility context, but `any` is fine in implementation
- Avoid `panic` except for programming errors (like use-after-release)
- Prefer `sync.Pool` for hot-path allocations
- Use `sync.Map` for read-heavy caches, `sync.Mutex` for write-heavy stores

## Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `go test ./...`
5. Run build: `go build ./...`
6. Commit with a clear message
7. Open a PR against `main`

### PR Guidelines

- Keep PRs focused — one feature or fix per PR
- Add tests for new functionality
- Update doc comments for changed public API
- Don't break existing interfaces (this is a public API)

## Reporting Issues

- Use GitHub Issues
- Include Go version (`go version`)
- Include adapter (Gin/Fiber) and version
- Include minimal reproduction code

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
