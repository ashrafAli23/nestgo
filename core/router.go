package core

// Router defines route registration.
type Router interface {
	GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
	Group(prefix string, middleware ...MiddlewareFunc) Router
	Use(middleware ...MiddlewareFunc)
	Static(path string, root string, middleware ...MiddlewareFunc)
	StaticFile(path string, filePath string, middleware ...MiddlewareFunc)
}
