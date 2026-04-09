package middleware

import (
	"net/http"
	"strconv"
	"strings"

	core "github.com/ashrafAli23/nestgo/core"
)

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string // headers the browser is allowed to read (e.g. X-Total-Count)
	AllowCredentials bool
	MaxAge           int // seconds
}

// DefaultCORSConfig returns a permissive CORS config for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a CORS middleware with the given config.
func CORS(config ...CORSConfig) core.MiddlewareFunc {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Pre-compute all header strings at init time — zero allocations per request.
	methods := strings.Join(cfg.AllowMethods, ", ")
	headers := strings.Join(cfg.AllowHeaders, ", ")
	expose := strings.Join(cfg.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(cfg.MaxAge)

	// Build origin lookup map for O(1) checks instead of O(n) slice iteration.
	allowAll := false
	originMap := make(map[string]struct{}, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		if o == "*" {
			allowAll = true
			break
		}
		originMap[o] = struct{}{}
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			origin := c.GetHeader("Origin")

			// Determine allowed origin
			if allowAll {
				c.SetHeader("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if _, ok := originMap[origin]; ok {
					c.SetHeader("Access-Control-Allow-Origin", origin)
					c.SetHeader("Vary", "Origin")
				}
			}

			c.SetHeader("Access-Control-Allow-Methods", methods)
			c.SetHeader("Access-Control-Allow-Headers", headers)
			if expose != "" {
				c.SetHeader("Access-Control-Expose-Headers", expose)
			}

			if cfg.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			}

			if cfg.MaxAge > 0 {
				c.SetHeader("Access-Control-Max-Age", maxAge)
			}

			// Handle preflight requests — return empty body, no JSON.
			if c.Method() == http.MethodOptions {
				return c.String(http.StatusNoContent, "")
			}

			return next(c)
		}
	}
}
