package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"sync"

	core "github.com/ashrafAli23/nestgo/core"
)

// CSRFConfig holds CSRF protection configuration.
type CSRFConfig struct {
	// TokenLength is the byte length of the generated token.
	// The actual header/cookie value will be hex-encoded (2x this length).
	// Default: 32 (produces 64-char hex string).
	TokenLength int
	// CookieName is the name of the cookie that stores the CSRF token.
	// Default: "_csrf".
	CookieName string
	// HeaderName is the HTTP header the client must send the token in.
	// Default: "X-CSRF-Token".
	HeaderName string
	// FormField is the form field name to check as a fallback if the header is missing.
	// Default: "_csrf".
	FormField string
	// CookiePath is the path for the CSRF cookie. Default: "/".
	CookiePath string
	// CookieDomain is the domain for the CSRF cookie. Default: "" (current domain).
	CookieDomain string
	// CookieSecure sets the Secure flag on the cookie. Default: false.
	CookieSecure bool
	// CookieHTTPOnly sets the HttpOnly flag on the cookie. Default: true.
	CookieHTTPOnly bool
	// CookieMaxAge is the max age for the cookie in seconds. Default: 86400 (24h).
	CookieMaxAge int
	// SkipFunc optionally skips CSRF check for certain requests.
	// Return true to skip. Default: nil (check all non-safe requests).
	SkipFunc func(c core.Context) bool
	// ErrorHandler is called when CSRF validation fails.
	// Default: returns 403 with "CSRF token mismatch".
	ErrorHandler func(c core.Context) error
	// TokenGenerator generates a new CSRF token string.
	// Default: crypto/rand hex-encoded bytes.
	TokenGenerator func() (string, error)
}

// DefaultCSRFConfig returns sensible CSRF defaults.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    32,
		CookieName:     "_csrf",
		HeaderName:     "X-CSRF-Token",
		FormField:      "_csrf",
		CookiePath:     "/",
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieMaxAge:   86400,
	}
}

// CSRF returns a middleware that provides Cross-Site Request Forgery protection.
//
// How it works:
//   - On every request, if no CSRF cookie exists, a new token is generated
//     and set as a cookie.
//   - For unsafe methods (POST, PUT, PATCH, DELETE), the middleware checks
//     that the token in the header (or form field) matches the cookie.
//   - Safe methods (GET, HEAD, OPTIONS, TRACE) are always allowed through.
//
// Usage:
//
//	// Global: default config
//	server.Use(middleware.CSRF())
//
//	// Custom config
//	server.Use(middleware.CSRF(middleware.CSRFConfig{
//	    CookieSecure: true,
//	    HeaderName:   "X-XSRF-Token",
//	    CookieName:   "XSRF-TOKEN",
//	}))
//
//	// Skip API routes (e.g. JWT-protected)
//	server.Use(middleware.CSRF(middleware.CSRFConfig{
//	    SkipFunc: func(c core.Context) bool {
//	        return strings.HasPrefix(c.Path(), "/api/")
//	    },
//	}))
func CSRF(config ...CSRFConfig) core.MiddlewareFunc {
	cfg := DefaultCSRFConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.TokenLength <= 0 {
		cfg.TokenLength = 32
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "_csrf"
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-CSRF-Token"
	}
	if cfg.FormField == "" {
		cfg.FormField = "_csrf"
	}
	if cfg.CookiePath == "" {
		cfg.CookiePath = "/"
	}
	if cfg.CookieMaxAge == 0 {
		cfg.CookieMaxAge = 86400
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c core.Context) error {
			return core.NewHTTPError(403, "CSRF token mismatch")
		}
	}
	if cfg.TokenGenerator == nil {
		// Pool byte slices to avoid allocation per token generation.
		tokenPool := sync.Pool{
			New: func() interface{} {
				b := make([]byte, cfg.TokenLength)
				return &b
			},
		}
		cfg.TokenGenerator = func() (string, error) {
			bp := tokenPool.Get().(*[]byte)
			b := *bp
			if _, err := rand.Read(b); err != nil {
				tokenPool.Put(bp)
				return "", err
			}
			token := hex.EncodeToString(b)
			tokenPool.Put(bp)
			return token, nil
		}
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			// Skip if user-defined skip function says so
			if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
				return next(c)
			}

			// Get existing token from cookie, or generate a new one
			token := c.Cookie(cfg.CookieName)
			if token == "" {
				var err error
				token, err = cfg.TokenGenerator()
				if err != nil {
					return core.NewHTTPError(500, "failed to generate CSRF token")
				}
				c.SetCookie(
					cfg.CookieName, token, cfg.CookieMaxAge,
					cfg.CookiePath, cfg.CookieDomain,
					cfg.CookieSecure, cfg.CookieHTTPOnly,
				)
			}

			// Store token in context so handlers can access it (e.g. for templates)
			c.Set("csrf_token", token)

			// Safe methods — skip validation.
			// HTTP methods from the framework are already uppercase; no ToUpper needed.
			method := c.Method()
			if method == "GET" || method == "HEAD" || method == "OPTIONS" || method == "TRACE" {
				return next(c)
			}

			// Unsafe methods — validate token.
			clientToken := c.GetHeader(cfg.HeaderName)
			if clientToken == "" {
				clientToken = c.FormValue(cfg.FormField)
			}

			// Use constant-time comparison to prevent timing side-channel attacks.
			if clientToken == "" || subtle.ConstantTimeCompare([]byte(clientToken), []byte(token)) != 1 {
				return cfg.ErrorHandler(c)
			}

			return next(c)
		}
	}
}
