package di

import (
	"context"

	core "github.com/ashrafAli23/nestgo/core"
	"go.uber.org/fx"
)

// LifecycleRegistration collects all lifecycle-aware services.
type LifecycleRegistration struct {
	fx.In
	InitHooks    []core.OnModuleInit    `group:"lifecycle_init"`
	DestroyHooks []core.OnModuleDestroy `group:"lifecycle_destroy"`
}

// AsInitHook tags a constructor so its result is collected for OnModuleInit.
func AsInitHook(f interface{}) interface{} {
	return fx.Annotate(
		f,
		fx.As(new(core.OnModuleInit)),
		fx.ResultTags(`group:"lifecycle_init"`),
	)
}

// AsDestroyHook tags a constructor so its result is collected for OnModuleDestroy.
func AsDestroyHook(f interface{}) interface{} {
	return fx.Annotate(
		f,
		fx.As(new(core.OnModuleDestroy)),
		fx.ResultTags(`group:"lifecycle_destroy"`),
	)
}

// RegisterLifecycleHooks wires all collected hooks into fx lifecycle.
func RegisterLifecycleHooks(lc fx.Lifecycle, reg LifecycleRegistration) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, h := range reg.InitHooks {
				if err := h.OnModuleInit(ctx); err != nil {
					core.Log().Error("OnModuleInit failed", core.F("error", err))
					return err
				}
			}
			if len(reg.InitHooks) > 0 {
				core.Log().Info("lifecycle hooks executed", core.F("OnModuleInit", len(reg.InitHooks)))
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			for _, h := range reg.DestroyHooks {
				if err := h.OnModuleDestroy(ctx); err != nil {
					core.Log().Error("OnModuleDestroy error", core.F("error", err))
				}
			}
			if len(reg.DestroyHooks) > 0 {
				core.Log().Info("lifecycle hooks executed", core.F("OnModuleDestroy", len(reg.DestroyHooks)))
			}
			return nil
		},
	})
}
