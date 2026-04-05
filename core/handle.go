package core

// ─── Handler Builder ────────────────────────────────────────────────────────
//
// This gives you NestJS-style parameter injection. Instead of writing:
//
//   func (ctrl *Controller) Create(c core.Context) error {
//       dto, err := core.Body[CreateUserDTO](c)
//       if err != nil { return err }
//       id, err := core.ParamInt(c, "id")
//       if err != nil { return err }
//       return ctrl.service.Create(id, dto)
//   }
//
// You write:
//
//   func (ctrl *Controller) Create(id int, dto *CreateUserDTO) error {
//       return ctrl.service.Create(id, dto)
//   }
//
// And register it as:
//
//   r.POST("/:id", core.Handle(
//       core.P[int]("id"),           // extract path param "id" as int
//       core.B[CreateUserDTO](),     // extract & validate body as *CreateUserDTO
//       ctrl.Create,                 // your clean function
//   ))
//
// The Handle function builds a core.HandlerFunc that:
//   1. Runs each extractor in order
//   2. If any fails, returns the error (already an HTTPError)
//   3. Calls your function with the extracted values
//
// ─── Extractors ─────────────────────────────────────────────────────────────
//
// These are the "decorators" — each one defines HOW to extract a value
// from the request context.

// Extractor pulls a typed value from the request context.
type Extractor[T any] struct {
	Extract func(Context) (T, error)
}

// B extracts and validates the request body into a DTO.
// Equivalent to NestJS's @Body()
func B[T any]() Extractor[*T] {
	return Extractor[*T]{
		Extract: func(c Context) (*T, error) {
			return Body[T](c)
		},
	}
}

// P extracts a path parameter as string.
// Equivalent to NestJS's @Param('key')
func P(key string) Extractor[string] {
	return Extractor[string]{
		Extract: func(c Context) (string, error) {
			return Param(c, key)
		},
	}
}

// PInt extracts a path parameter as int.
// Equivalent to NestJS's @Param('key', ParseIntPipe)
func PInt(key string) Extractor[int] {
	return Extractor[int]{
		Extract: func(c Context) (int, error) {
			return ParamInt(c, key)
		},
	}
}

// Q extracts a single query parameter as string with optional default.
// Equivalent to NestJS's @Query('key')
func Q(key string, defaultValue ...string) Extractor[string] {
	return Extractor[string]{
		Extract: func(c Context) (string, error) {
			return Query(c, key, defaultValue...), nil
		},
	}
}

// QInt extracts a single query parameter as int with optional default.
// Equivalent to NestJS's @Query('key', ParseIntPipe)
func QInt(key string, defaultValue ...int) Extractor[int] {
	return Extractor[int]{
		Extract: func(c Context) (int, error) {
			return QueryInt(c, key, defaultValue...)
		},
	}
}

// QDto extracts and validates query parameters into a DTO struct.
// Equivalent to NestJS's @Query() with a DTO class.
func QDto[T any]() Extractor[*T] {
	return Extractor[*T]{
		Extract: func(c Context) (*T, error) {
			return QueryDTO[T](c)
		},
	}
}

// H extracts a header value.
// Equivalent to NestJS's @Headers('key')
func H(key string, required bool) Extractor[string] {
	return Extractor[string]{
		Extract: func(c Context) (string, error) {
			return Header(c, key, required)
		},
	}
}

// Ctx passes the raw core.Context through.
// Use when your handler also needs direct context access.
func Ctx() Extractor[Context] {
	return Extractor[Context]{
		Extract: func(c Context) (Context, error) {
			return c, nil
		},
	}
}

// ─── Handle: 1 extractor ────────────────────────────────────────────────────

// Handle1 builds a HandlerFunc from 1 extractor + a typed handler function.
//
//	r.GET("/:id", core.Handle1(core.PInt("id"), ctrl.GetByID))
//
// Where ctrl.GetByID is:
//
//	func (ctrl *Ctrl) GetByID(id int) (any, error)
func Handle1[A any](
	e1 Extractor[A],
	fn func(A) (any, error),
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		result, err := fn(a)
		if err != nil {
			return err
		}
		if result == nil {
			return c.NoContent(204)
		}
		return c.JSON(200, result)
	}
}

// ─── Handle: 2 extractors ───────────────────────────────────────────────────

// Handle2 builds a HandlerFunc from 2 extractors + a typed handler function.
//
//	r.POST("/:id", core.Handle2(
//	    core.PInt("id"),
//	    core.B[CreateDTO](),
//	    ctrl.Create,
//	))
//
// Where ctrl.Create is:
//
//	func (ctrl *Ctrl) Create(id int, dto *CreateDTO) (any, error)
func Handle2[A any, B any](
	e1 Extractor[A],
	e2 Extractor[B],
	fn func(A, B) (any, error),
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		b, err := e2.Extract(c)
		if err != nil {
			return err
		}
		result, err := fn(a, b)
		if err != nil {
			return err
		}
		if result == nil {
			return c.NoContent(204)
		}
		return c.JSON(200, result)
	}
}

// ─── Handle: 3 extractors ───────────────────────────────────────────────────

// Handle3 builds a HandlerFunc from 3 extractors + a typed handler function.
//
//	r.PUT("/:id", core.Handle3(
//	    core.PInt("id"),
//	    core.B[UpdateDTO](),
//	    core.H("Authorization", true),
//	    ctrl.Update,
//	))
func Handle3[A any, B any, C any](
	e1 Extractor[A],
	e2 Extractor[B],
	e3 Extractor[C],
	fn func(A, B, C) (any, error),
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		b, err := e2.Extract(c)
		if err != nil {
			return err
		}
		cv, err := e3.Extract(c)
		if err != nil {
			return err
		}
		result, err := fn(a, b, cv)
		if err != nil {
			return err
		}
		if result == nil {
			return c.NoContent(204)
		}
		return c.JSON(200, result)
	}
}

// ─── Handle: 4 extractors ───────────────────────────────────────────────────

func Handle4[A any, B any, C any, D any](
	e1 Extractor[A],
	e2 Extractor[B],
	e3 Extractor[C],
	e4 Extractor[D],
	fn func(A, B, C, D) (any, error),
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		b, err := e2.Extract(c)
		if err != nil {
			return err
		}
		cv, err := e3.Extract(c)
		if err != nil {
			return err
		}
		d, err := e4.Extract(c)
		if err != nil {
			return err
		}
		result, err := fn(a, b, cv, d)
		if err != nil {
			return err
		}
		if result == nil {
			return c.NoContent(204)
		}
		return c.JSON(200, result)
	}
}

// ─── HandleC: With custom status / response control ─────────────────────────
//
// When you need to control the response status code or send non-JSON,
// use HandleC variants. These pass core.Context as the FIRST argument
// so you can call c.JSON(201, ...) yourself.

// HandleC1 is like Handle1 but also passes Context for response control.
//
//	r.POST("/", core.HandleC1(core.B[CreateDTO](), ctrl.Create))
//
// Where ctrl.Create is:
//
//	func (ctrl *Ctrl) Create(c core.Context, dto *CreateDTO) error
func HandleC1[A any](
	e1 Extractor[A],
	fn func(Context, A) error,
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		return fn(c, a)
	}
}

// HandleC2 is like Handle2 but also passes Context.
func HandleC2[A any, B any](
	e1 Extractor[A],
	e2 Extractor[B],
	fn func(Context, A, B) error,
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		b, err := e2.Extract(c)
		if err != nil {
			return err
		}
		return fn(c, a, b)
	}
}

// HandleC3 is like Handle3 but also passes Context.
func HandleC3[A any, B any, C any](
	e1 Extractor[A],
	e2 Extractor[B],
	e3 Extractor[C],
	fn func(Context, A, B, C) error,
) HandlerFunc {
	return func(c Context) error {
		a, err := e1.Extract(c)
		if err != nil {
			return err
		}
		b, err := e2.Extract(c)
		if err != nil {
			return err
		}
		cv, err := e3.Extract(c)
		if err != nil {
			return err
		}
		return fn(c, a, b, cv)
	}
}
