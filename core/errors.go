package core

import (
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP error with a status code and message.
type HTTPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
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

// Convenience constructors.
func ErrBadRequest(msg string) *HTTPError    { return NewHTTPError(http.StatusBadRequest, msg) }
func ErrUnauthorized(msg string) *HTTPError  { return NewHTTPError(http.StatusUnauthorized, msg) }
func ErrForbidden(msg string) *HTTPError     { return NewHTTPError(http.StatusForbidden, msg) }
func ErrNotFound(msg string) *HTTPError      { return NewHTTPError(http.StatusNotFound, msg) }
func ErrConflict(msg string) *HTTPError      { return NewHTTPError(http.StatusConflict, msg) }
func ErrInternalServer(msg string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, msg)
}
func ErrUnprocessable(msg string) *HTTPError {
	return NewHTTPError(http.StatusUnprocessableEntity, msg)
}

// ErrorHandler is the global error handler signature.
type ErrorHandler func(c Context, err error)

// DefaultErrorHandler checks for HTTPError and responds with JSON.
func DefaultErrorHandler(c Context, err error) {
	if err == nil {
		return
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		httpErr = ErrInternalServer(err.Error())
	}
	_ = c.JSON(httpErr.Code, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    httpErr.Code,
			"message": httpErr.Message,
			"details": httpErr.Details,
		},
	})
}
