package middleware

import (
	core "github.com/ashrafAli23/nestgo/core"
)

// HelmetConfig holds security headers configuration.
// Each field controls a specific security header.
// Set any string field to "" to disable that header.
type HelmetConfig struct {
	// X-Content-Type-Options. Default: "nosniff".
	ContentTypeNoSniff string
	// X-Frame-Options. Default: "SAMEORIGIN". Options: "DENY", "SAMEORIGIN".
	XFrameOptions string
	// Strict-Transport-Security. Default: "max-age=63072000; includeSubDomains".
	HSTS string
	// X-XSS-Protection. Default: "0" (disabled, modern browsers use CSP instead).
	XXSSProtection string
	// Referrer-Policy. Default: "strict-origin-when-cross-origin".
	ReferrerPolicy string
	// Content-Security-Policy. Default: "" (not set — highly app-specific).
	ContentSecurityPolicy string
	// X-DNS-Prefetch-Control. Default: "off".
	XDNSPrefetchControl string
	// X-Permitted-Cross-Domain-Policies. Default: "none".
	CrossDomainPolicies string
	// X-Download-Options. Default: "noopen".
	XDownloadOptions string
	// Permissions-Policy. Default: "" (not set — highly app-specific).
	PermissionsPolicy string
}

// DefaultHelmetConfig returns secure defaults following OWASP recommendations.
func DefaultHelmetConfig() HelmetConfig {
	return HelmetConfig{
		ContentTypeNoSniff:  "nosniff",
		XFrameOptions:       "SAMEORIGIN",
		HSTS:                "max-age=63072000; includeSubDomains",
		XXSSProtection:      "0",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		XDNSPrefetchControl: "off",
		CrossDomainPolicies: "none",
		XDownloadOptions:    "noopen",
	}
}

// Helmet returns a middleware that sets security headers.
func Helmet(config ...HelmetConfig) core.MiddlewareFunc {
	cfg := DefaultHelmetConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	headers := map[string]string{}
	if cfg.ContentTypeNoSniff != "" {
		headers["X-Content-Type-Options"] = cfg.ContentTypeNoSniff
	}
	if cfg.XFrameOptions != "" {
		headers["X-Frame-Options"] = cfg.XFrameOptions
	}
	if cfg.HSTS != "" {
		headers["Strict-Transport-Security"] = cfg.HSTS
	}
	if cfg.XXSSProtection != "" {
		headers["X-XSS-Protection"] = cfg.XXSSProtection
	}
	if cfg.ReferrerPolicy != "" {
		headers["Referrer-Policy"] = cfg.ReferrerPolicy
	}
	if cfg.ContentSecurityPolicy != "" {
		headers["Content-Security-Policy"] = cfg.ContentSecurityPolicy
	}
	if cfg.XDNSPrefetchControl != "" {
		headers["X-DNS-Prefetch-Control"] = cfg.XDNSPrefetchControl
	}
	if cfg.CrossDomainPolicies != "" {
		headers["X-Permitted-Cross-Domain-Policies"] = cfg.CrossDomainPolicies
	}
	if cfg.XDownloadOptions != "" {
		headers["X-Download-Options"] = cfg.XDownloadOptions
	}
	if cfg.PermissionsPolicy != "" {
		headers["Permissions-Policy"] = cfg.PermissionsPolicy
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			for key, value := range headers {
				c.SetHeader(key, value)
			}
			return next(c)
		}
	}
}
