package core

// Interceptor runs logic before and after a handler.
// Call next(c) to proceed to the handler. The error from next
// indicates whether the handler succeeded.
type Interceptor interface {
	Intercept(c Context, next HandlerFunc) error
}

// InterceptorFunc is a functional adapter for Interceptor.
type InterceptorFunc func(c Context, next HandlerFunc) error

func (f InterceptorFunc) Intercept(c Context, next HandlerFunc) error { return f(c, next) }

// UseInterceptors converts interceptors into a MiddlewareFunc.
// First interceptor is outermost.
func UseInterceptors(interceptors ...Interceptor) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		handler := next
		for i := len(interceptors) - 1; i >= 0; i-- {
			ic := interceptors[i]
			inner := handler
			handler = func(c Context) error {
				return ic.Intercept(c, inner)
			}
		}
		return handler
	}
}
