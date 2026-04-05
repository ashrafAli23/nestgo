package core

import "context"

// Server is the main application server interface.
type Server interface {
	Router
	Start(addr string) error
	Shutdown(ctx context.Context) error
	Name() string
	Underlying() interface{}
}
