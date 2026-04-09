package di

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core "github.com/ashrafAli23/nestgo/core"
	"go.uber.org/fx"
)

// ─── Controller Auto-Registration ───────────────────────────────────────────

// ControllerRegistration collects all controllers tagged with AsController().
type ControllerRegistration struct {
	fx.In
	Controllers []core.Controller `group:"controllers"`
}

// AsController annotates a controller constructor for auto-registration.
//
//	fx.Provide(di.AsController(user.NewController))
func AsController(f interface{}) interface{} {
	return fx.Annotate(
		f,
		fx.As(new(core.Controller)),
		fx.ResultTags(`group:"controllers"`),
	)
}

// RegisterControllers wires all collected controllers to the server router.
// It auto-detects PrefixedController and VersionedController interfaces.
func RegisterControllers(router core.Router, config *core.Config, reg ControllerRegistration) {
	for _, ctrl := range reg.Controllers {
		r := router

		// Apply version prefix/guard if controller implements VersionedController
		if vc, ok := ctrl.(core.VersionedController); ok && config.Versioning != nil {
			switch config.Versioning.Strategy {
			case core.URIVersioning:
				r = r.Group(fmt.Sprintf("/v%s", vc.Version()))
			case core.HeaderVersioning, core.MediaTypeVersioning:
				r = r.Group("", core.UseGuards(core.VersionGuard(*config.Versioning, vc.Version())))
			}
		}

		// Apply controller prefix if controller implements PrefixedController
		if pc, ok := ctrl.(core.PrefixedController); ok {
			r = r.Group(pc.Prefix())
		}

		ctrl.RegisterRoutes(r)
	}
	core.Log().Info("controllers registered", core.F("count", len(reg.Controllers)))
}

// ─── Server Lifecycle ───────────────────────────────────────────────────────

// StartServer hooks the server into fx lifecycle (start on OnStart, stop on OnStop).
// It also listens for OS signals (SIGINT, SIGTERM) so that graceful shutdown
// runs automatically when the process is killed (e.g. Kubernetes SIGTERM).
func StartServer(lc fx.Lifecycle, server core.Server, config *core.Config, shutdowner fx.Shutdowner) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start the HTTP(S) server in the background.
			go func() {
				var err error
				if config.TLSCertFile != "" && config.TLSKeyFile != "" {
					err = server.StartTLS(config.Addr, config.TLSCertFile, config.TLSKeyFile)
				} else {
					err = server.Start(config.Addr)
				}
				if err != nil {
					core.Log().Error("server error", core.F("error", err))
				}
			}()

			// Listen for OS termination signals and trigger fx shutdown
			// so that OnStop hooks (including OnShutdown) run cleanly.
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
				sig := <-sigCh
				core.Log().Info("received signal, shutting down", core.F("signal", sig.String()))
				_ = shutdowner.Shutdown()
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			timeout := config.ShutdownTimeout
			if timeout <= 0 {
				timeout = 10 * time.Second
			}
			shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			// Run user shutdown hooks before stopping the server
			for _, hook := range config.OnShutdown {
				if err := hook(shutdownCtx); err != nil {
					core.Log().Error("OnShutdown hook error", core.F("error", err))
				}
			}
			return server.Shutdown(shutdownCtx)
		},
	})
}

// ─── Config Provider ────────────────────────────────────────────────────────

// ConfigModule provides a core.Config to the DI container.
func ConfigModule(config *core.Config) fx.Option {
	if config == nil {
		config = core.DefaultConfig()
	}
	return fx.Provide(func() *core.Config { return config })
}

// ─── Core Module ────────────────────────────────────────────────────────────

// CoreModule bundles controller registration + server lifecycle.
// Every app should include this.
var CoreModule = fx.Module("nestgo",
	fx.Invoke(RegisterHealthEndpoints),
	fx.Invoke(RegisterControllers),
	fx.Invoke(StartServer),
	fx.Invoke(RegisterLifecycleHooks),
)

// ─── App Builder ────────────────────────────────────────────────────────────

// NewApp creates an fx.App from a config, a ServerProvider, and any extra modules.
//
// The ServerProvider is what makes this adapter-agnostic. The user passes in a
// function that creates a core.Server from the adapter they installed:
//
//	// Gin user:
//	di.NewApp(config, ginadapter.New, user.Module)
//
//	// Fiber user:
//	di.NewApp(config, fiberadapter.New, user.Module)
//
// The DI package never imports either adapter.
type ServerProvider func(config *core.Config) core.Server

func NewApp(config *core.Config, provider ServerProvider, opts ...fx.Option) *fx.App {
	if config == nil {
		config = core.DefaultConfig()
	}

	baseOpts := []fx.Option{
		ConfigModule(config),
		fx.Provide(func(cfg *core.Config) core.Server {
			return provider(cfg)
		}),
		fx.Provide(func(server core.Server, cfg *core.Config) core.Router {
			var router core.Router = server
			if cfg.GlobalPrefix != "" {
				router = server.Group(cfg.GlobalPrefix)
			}
			// Apply global middleware first (outermost)
			if len(cfg.GlobalMiddlewares) > 0 {
				router.Use(cfg.GlobalMiddlewares...)
			}
			// Register global ValidateFunc if set via config
			if cfg.ValidateFunc != nil {
				core.SetValidateFunc(cfg.ValidateFunc)
			}
			// Apply global guards, pipes, interceptors, and filters
			globalMws := core.ApplyRouteOptions(core.RouteOptions{
				Guards:       cfg.GlobalGuards,
				Pipes:        cfg.GlobalPipes,
				Interceptors: cfg.GlobalInterceptors,
				Filters:      cfg.GlobalFilters,
			})
			if len(globalMws) > 0 {
				router.Use(globalMws...)
			}
			return router
		}),
	}

	// User options (including middleware setup) must come before CoreModule
	// so that middleware is registered before RegisterControllers adds routes.
	// Fiber executes handlers in registration order, so middleware registered
	// after routes won't wrap those route handlers.
	allOpts := append(baseOpts, opts...)
	allOpts = append(allOpts, CoreModule)
	return fx.New(allOpts...)
}
