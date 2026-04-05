package core

// Controller is the interface that all feature controllers implement.
// The DI system collects all Controllers and calls RegisterRoutes on each.
type Controller interface {
	RegisterRoutes(router Router)
}

// PrefixedController is an optional interface that controllers can implement
// to declare a route prefix. The DI system auto-wraps the router in a Group
// with this prefix before calling RegisterRoutes.
//
//	func (c *UserController) Prefix() string { return "/users" }
//	func (c *UserController) RegisterRoutes(r core.Router) {
//	    r.GET("/", c.List)       // GET /users/
//	    r.GET("/:id", ...)       // GET /users/:id
//	}
type PrefixedController interface {
	Prefix() string
}

// VersionedController is an optional interface that controllers can implement
// to declare an API version. Combined with Config.Versioning, this controls
// how the version is applied (URI prefix, header check, or media type).
//
//	func (c *UserControllerV1) Version() string { return "1" }
type VersionedController interface {
	Version() string
}
