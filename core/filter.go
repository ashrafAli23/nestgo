package core

// ExceptionFilter handles errors for specific routes or controllers.
// CanHandle returns true if this filter should handle the error.
// Handle formats and sends the error response.
type ExceptionFilter interface {
	CanHandle(err error) bool
	Handle(c Context, err error)
}

// UseFilters converts exception filters into a MiddlewareFunc.
// The first filter whose CanHandle returns true handles the error.
// If none match, the error propagates to the global ErrorHandler.
func UseFilters(filters ...ExceptionFilter) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			err := next(c)
			if err == nil {
				return nil
			}
			for _, f := range filters {
				if f.CanHandle(err) {
					f.Handle(c, err)
					return nil
				}
			}
			return err
		}
	}
}

// HTTPErrorFilter handles all *HTTPError with a custom formatter.
type HTTPErrorFilter struct {
	Formatter func(c Context, httpErr *HTTPError)
}

func (f *HTTPErrorFilter) CanHandle(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

func (f *HTTPErrorFilter) Handle(c Context, err error) {
	f.Formatter(c, err.(*HTTPError))
}
