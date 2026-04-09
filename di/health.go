package di

import (
	core "github.com/ashrafAli23/nestgo/core"
	"go.uber.org/fx"
)

// RegisterHealthEndpoints registers /health and /ready based on config.
// Invoked automatically by CoreModule.
func RegisterHealthEndpoints(server core.Server, config *core.Config) {
	if !config.HealthCheck {
		return
	}

	server.GET("/health", func(c core.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	if config.ReadinessCheck != nil {
		check := config.ReadinessCheck
		server.GET("/ready", func(c core.Context) error {
			if err := check(); err != nil {
				return c.JSON(503, map[string]string{
					"status": "unavailable",
					"reason": err.Error(),
				})
			}
			return c.JSON(200, map[string]string{"status": "ready"})
		})
	}
}

// HealthModule registers health endpoints. Included in CoreModule.
var HealthModule = fx.Invoke(RegisterHealthEndpoints)
