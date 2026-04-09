package core

import "context"

// Server is the main application server interface.
type Server interface {
	Router
	Start(addr string) error
	// StartTLS starts the server with TLS using the given cert and key files.
	StartTLS(addr, certFile, keyFile string) error
	Shutdown(ctx context.Context) error
	Name() string
	Underlying() interface{}
}
