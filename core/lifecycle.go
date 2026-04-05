package core

import "context"

// OnModuleInit is called when the module is initialized (server starting).
// Use for DB connections, cache warming, etc.
type OnModuleInit interface {
	OnModuleInit(ctx context.Context) error
}

// OnModuleDestroy is called when the module is shutting down.
// Use for closing connections, flushing buffers, etc.
type OnModuleDestroy interface {
	OnModuleDestroy(ctx context.Context) error
}
