package core

// Guard decides whether a request is allowed to proceed.
// Return true to allow, false to deny.
// Returning an error short-circuits with that error.
type Guard interface {
	CanActivate(c Context) (bool, error)
}

// GuardFunc is a functional adapter for Guard.
type GuardFunc func(Context) (bool, error)

func (f GuardFunc) CanActivate(c Context) (bool, error) { return f(c) }

// UseGuards converts guards into a MiddlewareFunc.
// Guards run in order; first failure stops the chain.
func UseGuards(guards ...Guard) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			for _, g := range guards {
				allowed, err := g.CanActivate(c)
				if err != nil {
					return err
				}
				if !allowed {
					return ErrForbidden("access denied")
				}
			}
			return next(c)
		}
	}
}
