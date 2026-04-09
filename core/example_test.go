package core_test

import (
	"fmt"

	"github.com/ashrafAli23/nestgo/core"
)

// A Guard that checks for an Authorization header.
func ExampleGuardFunc() {
	authGuard := core.GuardFunc(func(c core.Context) (bool, error) {
		token := c.GetHeader("Authorization")
		if token == "" {
			return false, core.ErrUnauthorized("missing token")
		}
		return true, nil
	})

	// Use as middleware on a route group:
	//   r.Group("/admin", core.UseGuards(authGuard))
	_ = authGuard
}

// UseGuards converts guards into middleware.
// Guards run in order — first failure stops the chain.
func ExampleUseGuards() {
	authGuard := core.GuardFunc(func(c core.Context) (bool, error) {
		return c.Get("user") != nil, nil
	})

	adminGuard := core.GuardFunc(func(c core.Context) (bool, error) {
		return c.Get("role") == "admin", nil
	})

	// Apply at different levels:
	//   config.GlobalGuards = []core.Guard{authGuard}          // global
	//   r.Group("/admin", core.UseGuards(adminGuard))          // per-controller
	//   r.DELETE("/:id", handler, core.UseGuards(adminGuard))  // per-route
	_ = core.UseGuards(authGuard, adminGuard)
}

// InterceptorFunc wraps handler execution for cross-cutting concerns.
func ExampleInterceptorFunc() {
	logInterceptor := core.InterceptorFunc(func(c core.Context, next core.HandlerFunc) error {
		fmt.Printf("→ %s %s\n", c.Method(), c.Path())
		err := next(c)
		fmt.Printf("← %s %s\n", c.Method(), c.Path())
		return err
	})

	// Use as middleware:
	//   r.Group("/api", core.UseInterceptors(logInterceptor))
	_ = logInterceptor
}

// WithPipes wraps an extractor with transform/validate steps.
func ExampleWithPipes() {
	// Extract path param "id" as int, then validate it's between 1 and 999999.
	idExtractor := core.WithPipes(
		core.PInt("id"),
		core.IntRangePipe("id", 1, 999999),
	)

	// Use with Handle1:
	//   r.GET("/:id", core.Handle1(idExtractor, ctrl.GetByID))
	_ = idExtractor
}

// Handle1 builds a HandlerFunc from one extractor and a typed function.
// The framework extracts the value, calls your function, and auto-responds.
func ExampleHandle1() {
	// Your handler receives typed values — no manual parsing:
	getUser := func(id int) (any, error) {
		return map[string]any{"id": id, "name": "John"}, nil
	}

	handler := core.Handle1(core.PInt("id"), getUser)

	// Register:
	//   r.GET("/users/:id", handler)
	//
	// GET /users/42 → {"id": 42, "name": "John"}
	_ = handler
}

// Handle2 builds a HandlerFunc from two extractors.
func ExampleHandle2() {
	type CreateDTO struct {
		Name string `json:"name"`
	}

	create := func(id int, dto *CreateDTO) (any, error) {
		return map[string]any{"id": id, "name": dto.Name}, nil
	}

	handler := core.Handle2(core.PInt("id"), core.B[CreateDTO](), create)

	// Register:
	//   r.POST("/users/:id", handler)
	_ = handler
}

// HandleC1 passes Context as the first argument for full response control.
func ExampleHandleC1() {
	type CreateDTO struct {
		Name string `json:"name"`
	}

	create := func(c core.Context, dto *CreateDTO) error {
		// Full control over the response:
		return c.JSON(201, map[string]any{"name": dto.Name})
	}

	handler := core.HandleC1(core.B[CreateDTO](), create)

	// Register:
	//   r.POST("/users", handler)
	_ = handler
}

// ApplyRouteOptions bundles guards, interceptors, and filters
// in the correct NestJS execution order: Filters -> Guards -> Interceptors.
func ExampleApplyRouteOptions() {
	authGuard := core.GuardFunc(func(c core.Context) (bool, error) {
		return true, nil
	})

	logInterceptor := core.InterceptorFunc(func(c core.Context, next core.HandlerFunc) error {
		return next(c)
	})

	opts := core.ApplyRouteOptions(core.RouteOptions{
		Guards:       []core.Guard{authGuard},
		Interceptors: []core.Interceptor{logInterceptor},
	})

	// Use as middleware on a group:
	//   r.Group("/users", opts...)
	_ = opts
}

// WithMeta attaches metadata to routes for guards/interceptors to read.
func ExampleWithMeta() {
	roleGuard := core.GuardFunc(func(c core.Context) (bool, error) {
		roles := c.Get("roles")
		if roles == nil {
			return false, nil
		}
		_ = roles // check if user's role is in the allowed list
		return true, nil
	})

	// Attach metadata + guard to a route:
	//   r.DELETE("/:id", handler,
	//       core.WithMeta("roles", []string{"admin"}),
	//       core.UseGuards(roleGuard),
	//   )
	_ = roleGuard
}

// HTTPErrorFilter catches *HTTPError and formats a custom response.
func ExampleHTTPErrorFilter() {
	filter := &core.HTTPErrorFilter{
		Formatter: func(c core.Context, httpErr *core.HTTPError) {
			_ = c.JSON(httpErr.Code, map[string]any{
				"status":  "error",
				"code":    httpErr.Code,
				"message": httpErr.Message,
			})
		},
	}

	// Use as middleware:
	//   r.Group("/api", core.UseFilters(filter))
	_ = filter
}

// Validatable DTOs are auto-validated when extracted with B[T]().
func ExampleValidatable() {
	type CreateUserDTO struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Implement Validate() on your DTO:
	//   func (d *CreateUserDTO) Validate() error {
	//       if d.Name == "" {
	//           return core.ErrBadRequest("name is required")
	//       }
	//       return nil
	//   }

	// Then use with extractors — validation runs automatically:
	//   r.POST("/users", core.Handle1(core.B[CreateUserDTO](), ctrl.Create))
	_ = core.B[CreateUserDTO]()
}

// NewSSEStream creates a channel for streaming Server-Sent Events.
func ExampleNewSSEStream() {
	stream := core.NewSSEStream(10)

	// In a goroutine, send events:
	go func() {
		defer close(stream)
		for i := 0; i < 5; i++ {
			stream <- core.SSEEvent{
				Event: "message",
				Data:  fmt.Sprintf(`{"count": %d}`, i),
				ID:    fmt.Sprintf("%d", i),
			}
		}
	}()

	// In a handler:
	//   return core.SSE(c, stream)
	_ = stream
}

// SetLogger replaces the global logger with your own implementation.
func ExampleSetLogger() {
	// Use the functional adapter for simple logging:
	core.SetLogger(core.LoggerFunc(func(level, msg string, fields []core.Field) {
		fmt.Printf("[%s] %s\n", level, msg)
	}))

	core.Log().Info("server started", core.F("addr", ":3000"))

	// Output:
	// [INFO] server started
}

// DefaultConfig returns sensible defaults for development.
func ExampleDefaultConfig() {
	config := core.DefaultConfig()
	config.Addr = ":3000"
	config.GlobalPrefix = "/api"
	config.Debug = true

	fmt.Println(config.AppName)
	fmt.Println(config.Addr)
	fmt.Println(config.GlobalPrefix)

	// Output:
	// NestGo App
	// :3000
	// /api
}
