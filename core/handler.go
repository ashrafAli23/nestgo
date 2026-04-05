package core

// HandlerFunc is the universal handler signature.
type HandlerFunc func(Context) error

// MiddlewareFunc wraps a handler and returns a new handler.
type MiddlewareFunc func(HandlerFunc) HandlerFunc
