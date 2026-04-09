package core

import (
	"fmt"
	"net/http"
)

// ─── HTTPError ──────────────────────────────────────────────────────────────

// HTTPError represents an HTTP error with a status code and message.
type HTTPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	// Errors holds a list of individual field-level errors (e.g. validation).
	Errors []FieldError `json:"errors,omitempty"`
}

// FieldError represents a single validation or field-level error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`   // validation tag that failed (e.g. "required", "min")
	Value   string `json:"value,omitempty"` // the rejected value
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

func NewHTTPErrorWithDetails(code int, message string, details interface{}) *HTTPError {
	return &HTTPError{Code: code, Message: message, Details: details}
}

// NewValidationError creates a 422 error with structured field errors.
//
//	return core.NewValidationError(
//	    core.FieldError{Field: "email", Message: "invalid format", Tag: "email"},
//	    core.FieldError{Field: "name", Message: "is required", Tag: "required"},
//	)
func NewValidationError(errors ...FieldError) *HTTPError {
	return &HTTPError{
		Code:    http.StatusUnprocessableEntity,
		Message: "validation failed",
		Errors:  errors,
	}
}

// Convenience constructors.
func ErrBadRequest(msg string) *HTTPError   { return NewHTTPError(http.StatusBadRequest, msg) }
func ErrUnauthorized(msg string) *HTTPError { return NewHTTPError(http.StatusUnauthorized, msg) }
func ErrForbidden(msg string) *HTTPError    { return NewHTTPError(http.StatusForbidden, msg) }
func ErrNotFound(msg string) *HTTPError     { return NewHTTPError(http.StatusNotFound, msg) }
func ErrConflict(msg string) *HTTPError     { return NewHTTPError(http.StatusConflict, msg) }
func ErrInternalServer(msg string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, msg)
}
func ErrUnprocessable(msg string) *HTTPError {
	return NewHTTPError(http.StatusUnprocessableEntity, msg)
}

// ─── Error Handlers ─────────────────────────────────────────────────────────

// ErrorHandler is the global error handler signature.
type ErrorHandler func(c Context, err error)

// DefaultErrorHandler checks for HTTPError and responds with JSON.
// Output: {"error": {"code": 422, "message": "...", "errors": [...]}}
func DefaultErrorHandler(c Context, err error) {
	if err == nil {
		return
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		httpErr = ErrInternalServer(err.Error())
	}

	body := map[string]interface{}{
		"code":    httpErr.Code,
		"message": httpErr.Message,
	}
	if httpErr.Details != nil {
		body["details"] = httpErr.Details
	}
	if len(httpErr.Errors) > 0 {
		body["errors"] = httpErr.Errors
	}

	_ = c.JSON(httpErr.Code, map[string]interface{}{"error": body})
}

// ─── RFC 7807 Problem Details ───────────────────────────────────────────────

// ProblemDetail represents an RFC 7807 (application/problem+json) error response.
//
//	config.ErrorHandler = core.ProblemDetailErrorHandler
type ProblemDetail struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// statusTitles maps HTTP status codes to their standard titles.
var statusTitles = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	409: "Conflict",
	413: "Payload Too Large",
	422: "Unprocessable Entity",
	429: "Too Many Requests",
	500: "Internal Server Error",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
}

// ProblemDetailErrorHandler outputs errors as RFC 7807 application/problem+json.
//
// Usage:
//
//	config.ErrorHandler = core.ProblemDetailErrorHandler
func ProblemDetailErrorHandler(c Context, err error) {
	if err == nil {
		return
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		httpErr = ErrInternalServer(err.Error())
	}

	title := statusTitles[httpErr.Code]
	if title == "" {
		title = http.StatusText(httpErr.Code)
	}

	problem := ProblemDetail{
		Type:   "about:blank",
		Title:  title,
		Status: httpErr.Code,
		Detail: httpErr.Message,
		Errors: httpErr.Errors,
	}

	c.SetHeader("Content-Type", "application/problem+json")
	_ = c.JSON(httpErr.Code, problem)
}
