package core

// RouteOptions configures guards, interceptors, pipes, and filters for a route or group.
// This enforces the correct NestJS execution order automatically:
// Filters (outermost) → Guards → Pipes → Interceptors (innermost, closest to handler)
type RouteOptions struct {
	Guards       []Guard
	Pipes        []MiddlewareFunc // transform/validate the request before interceptors
	Interceptors []Interceptor
	Filters      []ExceptionFilter
}

// ApplyRouteOptions returns middleware in the correct execution order.
func ApplyRouteOptions(opts RouteOptions) []MiddlewareFunc {
	var mws []MiddlewareFunc
	if len(opts.Filters) > 0 {
		mws = append(mws, UseFilters(opts.Filters...))
	}
	if len(opts.Guards) > 0 {
		mws = append(mws, UseGuards(opts.Guards...))
	}
	if len(opts.Pipes) > 0 {
		mws = append(mws, opts.Pipes...)
	}
	if len(opts.Interceptors) > 0 {
		mws = append(mws, UseInterceptors(opts.Interceptors...))
	}
	return mws
}
