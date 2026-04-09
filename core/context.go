package core

import (
	"context"
	"io"
	"mime/multipart"
)

// Context is the core abstraction over HTTP request/response.
// Both GinContext and FiberContext implement this interface.
// Your controllers and middleware ONLY use this — never gin.Context or fiber.Ctx directly.
type Context interface {
	// ─── Request Data ───────────────────────────────────────────────

	// Method returns the HTTP method (GET, POST, PUT, DELETE, etc.)
	Method() string

	// Path returns the matched route path
	Path() string

	// Param returns a URL path parameter by name.
	//   e.g., for route "/users/:id" → Param("id") returns the value
	Param(key string) string

	// Query returns a query string parameter by name.
	//   e.g., /users?page=2 → Query("page") returns "2"
	Query(key string) string

	// QueryDefault returns a query string parameter or the provided default if empty.
	QueryDefault(key string, defaultValue string) string

	// GetHeader returns a request header value by key.
	GetHeader(key string) string

	// Cookie returns a cookie value by name.
	//   Returns empty string if the cookie does not exist.
	Cookie(name string) string

	// Body reads the raw request body as bytes.
	Body() ([]byte, error)

	// Bind parses the request body (JSON, XML, form) into the given struct.
	//   Gin uses ShouldBind (auto-detects Content-Type),
	//   Fiber uses Bind().Body()
	Bind(v interface{}) error

	// FormValue returns a form field value by key (from POST/PUT form data).
	FormValue(key string) string

	// FormFile returns a file from a multipart form upload.
	FormFile(key string) (*multipart.FileHeader, error)

	// ContentType returns the Content-Type header of the request.
	ContentType() string

	// IsWebSocket returns true if the request is a WebSocket upgrade.
	IsWebSocket() bool

	// ─── Response ───────────────────────────────────────────────────

	// Status sets the HTTP response status code and returns Context for chaining.
	Status(code int) Context

	// JSON sends a JSON response with the given status code.
	JSON(status int, data interface{}) error

	// XML sends an XML response with the given status code.
	XML(status int, data interface{}) error

	// String sends a plain text response with the given status code.
	String(status int, format string, values ...interface{}) error

	// SendBytes sends raw bytes as the response body.
	SendBytes(status int, data []byte) error

	// SendStream sends a streaming response.
	SendStream(stream io.Reader) error

	// SendFile sends a file to the client.
	SendFile(filePath string) error

	// Download prompts the client to download the file.
	Download(filePath string, filename string) error

	// NoContent sends a response with no body and the given status code.
	NoContent(status int) error

	// ResponseStatus returns the HTTP response status code that has been set.
	// Returns 0 if no response has been written yet.
	ResponseStatus() int

	// ResponseBody returns a copy of the response body that has been written.
	// Useful in interceptors to inspect/log the response after the handler runs.
	// Returns nil if no body has been written yet.
	ResponseBody() []byte

	// SetHeader sets a response header.
	SetHeader(key, value string)

	// SetCookie sets a response cookie.
	//   For advanced cookie options (SameSite, Partitioned, etc.),
	//   use the Underlying() escape hatch.
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)

	// Redirect sends an HTTP redirect to the specified URL.
	Redirect(status int, url string) error

	// ─── Request Metadata ───────────────────────────────────────────

	// ClientIP returns the client's IP address.
	ClientIP() string

	// FullURL returns the full request URL.
	FullURL() string

	// ─── Context Storage (per-request key/value store) ──────────────

	// Set stores a key/value pair in the request context.
	//   Useful for passing data between middleware and handlers.
	Set(key string, value interface{})

	// Get retrieves a value from the request context.
	// Returns nil if the key does not exist.
	Get(key string) interface{}

	// ─── Flow Control ───────────────────────────────────────────────

	// Next calls the next handler/middleware in the chain.
	Next() error

	// ─── Safe Concurrency ──────────────────────────────────────────

	// Clone returns a copy of the Context that is safe to use in goroutines.
	// The original Context uses pooling and is recycled after the handler returns.
	// If you pass Context to a goroutine WITHOUT cloning, you will get data corruption.
	//
	//   go func(ctx core.Context) {
	//       // WRONG: ctx is recycled after handler returns
	//   }(c)
	//
	//   go func(ctx core.Context) {
	//       // CORRECT: cloned ctx is safe
	//   }(c.Clone())
	Clone() Context

	// ─── Underlying Access (escape hatch) ───────────────────────────

	// Underlying returns the original framework context (gin.Context or fiber.Ctx).
	// Use this ONLY when you absolutely need framework-specific features.
	// This breaks the abstraction — use sparingly.
	Underlying() interface{}

	// ─── Native Standard Context ────────────────────────────────────────────

	// RequestCtx returns the standard library context.Context.
	// This is useful for passing to database queries, gRPC clients, etc.
	RequestCtx() context.Context

	// SetRequestCtx sets the standard library context.Context.
	// This is useful for middleware that needs to inject values (like tracing).
	SetRequestCtx(ctx context.Context)
}
