package core

// WithMeta returns a MiddlewareFunc that attaches metadata to the request context.
// Guards and interceptors can read it via c.Get(key).
//
//	r.DELETE("/:id", handler,
//	    core.WithMeta("roles", []string{"admin"}),
//	    core.UseGuards(roleGuard),
//	)
//
// In the guard:
//
//	roles := c.Get("roles")
func WithMeta(key string, value interface{}) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			c.Set(key, value)
			return next(c)
		}
	}
}

// Meta is a convenience for attaching multiple metadata key-value pairs at once.
//
//	r.GET("/admin", handler, core.Meta(map[string]interface{}{
//	    "roles":      []string{"admin"},
//	    "permission": "users:delete",
//	}))
func Meta(pairs map[string]interface{}) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			for k, v := range pairs {
				c.Set(k, v)
			}
			return next(c)
		}
	}
}
