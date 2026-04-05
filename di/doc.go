// Package di provides dependency injection and application bootstrapping for NestGo.
//
// It uses [go.uber.org/fx] under the hood but exposes a simple API:
//
//   - [NewApp] — creates the application from a config + adapter + modules
//   - [AsController] — tags a constructor for auto-registration
//   - [AsInitHook] / [AsDestroyHook] — tags constructors for lifecycle hooks
//
// # Quick Start
//
//	func main() {
//	    config := core.DefaultConfig()
//	    config.Addr = ":3000"
//	    config.GlobalPrefix = "/api"
//
//	    app := di.NewApp(config, ginadapter.New,
//	        users.Module,
//	        products.Module,
//	    )
//	    app.Run()
//	}
//
// # Modules
//
// A module is an [fx.Option] that groups related providers:
//
//	var Module = fx.Module("users",
//	    fx.Provide(di.AsController(NewUserController)),
//	    fx.Provide(NewUserService),
//	    fx.Provide(NewUserRepository),
//	)
//
// # What NewApp Does
//
//  1. Provides [core.Config] to DI
//  2. Creates the [core.Server] using the adapter
//  3. Creates a [core.Router] (with global prefix + guards/interceptors/filters)
//  4. Registers all controllers (auto-detects [core.PrefixedController] and [core.VersionedController])
//  5. Starts the server with lifecycle hooks
//
// # Adapter Agnostic
//
// The [ServerProvider] type makes the DI package adapter-agnostic. The user passes
// the adapter's New function — the DI package never imports Gin or Fiber:
//
//	// Swap adapters by changing one line:
//	di.NewApp(config, ginadapter.New, ...)   // Gin
//	di.NewApp(config, fiberadapter.New, ...) // Fiber
package di
