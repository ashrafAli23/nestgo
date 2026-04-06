package core

import (
	"fmt"
	"strconv"
	"strings"
)

// ─── Pluggable Validator ────────────────────────────────────────────────────

// globalValidateFunc is the pluggable validation function used by Body() and
// QueryDTO(). Set it once at startup via SetValidateFunc.
var globalValidateFunc func(v interface{}) error

// SetValidateFunc registers a global validation function that Body() and
// QueryDTO() call automatically after binding. This bridges nestgo-validator
// (or any validation library) into the extraction pipeline.
//
// Call this once in main() before starting the server:
//
//	core.SetValidateFunc(validator.Validate)
//
// The function receives the bound DTO pointer and should return an error
// (typically *validator.ValidationError or *core.HTTPError) on failure.
func SetValidateFunc(fn func(v interface{}) error) {
	globalValidateFunc = fn
}

// ─── Validator Interface ────────────────────────────────────────────────────

// Validatable is an interface that DTOs can implement to self-validate.
// If your struct implements this, Body/QueryDTO will call it automatically.
//
// Example:
//
//	type CreateUserDTO struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	func (d *CreateUserDTO) Validate() error {
//	    if d.Name == "" {
//	        return core.ErrBadRequest("name is required")
//	    }
//	    if !strings.Contains(d.Email, "@") {
//	        return core.ErrBadRequest("invalid email format")
//	    }
//	    return nil
//	}
type Validatable interface {
	Validate() error
}

// ─── Body Extractor ─────────────────────────────────────────────────────────

// Body binds the request body to a DTO struct and validates it.
// This is the Go equivalent of NestJS's @Body() decorator.
//
// Usage:
//
//	func (ctrl *Controller) Create(c core.Context) error {
//	    dto, err := core.Body[CreateUserDTO](c)
//	    if err != nil {
//	        return err // already an HTTPError
//	    }
//	    // dto is *CreateUserDTO, fully bound and validated
//	    user, err := ctrl.service.Create(dto)
//	    ...
//	}
func Body[T any](c Context) (*T, error) {
	dto := new(T)
	if err := c.Bind(dto); err != nil {
		return nil, ErrBadRequest("invalid request body: " + err.Error())
	}
	// 1. Run global struct-tag validator (e.g. nestgo-validator) if registered.
	if globalValidateFunc != nil {
		if err := globalValidateFunc(dto); err != nil {
			return nil, err
		}
	}
	// 2. Run per-DTO custom validation if the DTO implements Validatable.
	if v, ok := any(dto).(Validatable); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}
	return dto, nil
}

// ─── Query DTO Extractor ────────────────────────────────────────────────────

// QueryDTO binds query string parameters to a DTO struct and validates it.
// The struct should use `query:"fieldname"` tags for mapping.
// This is the Go equivalent of NestJS's @Query() decorator with a DTO.
//
// Note: This does manual field extraction from query params. For full
// struct binding from query strings, use c.Bind() with the appropriate
// struct tags for your adapter (form: for gin, query: for fiber).
//
// Usage:
//
//	func (ctrl *Controller) List(c core.Context) error {
//	    q, err := core.QueryDTO[ListUsersQuery](c)
//	    if err != nil {
//	        return err
//	    }
//	    users, err := ctrl.service.List(q.Page, q.Limit)
//	    ...
//	}
func QueryDTO[T any](c Context) (*T, error) {
	dto := new(T)
	// Use the underlying Bind which supports query params via struct tags
	if err := c.Bind(dto); err != nil {
		return nil, ErrBadRequest("invalid query parameters: " + err.Error())
	}
	if globalValidateFunc != nil {
		if err := globalValidateFunc(dto); err != nil {
			return nil, err
		}
	}
	if v, ok := any(dto).(Validatable); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}
	return dto, nil
}

// ─── Single Param Extractors ────────────────────────────────────────────────

// Param extracts a path parameter as string.
// Returns an error if the parameter is empty.
//
//	id, err := core.Param(c, "id")
func Param(c Context, key string) (string, error) {
	val := c.Param(key)
	if val == "" {
		return "", ErrBadRequest(fmt.Sprintf("path parameter '%s' is required", key))
	}
	return val, nil
}

// ParamInt extracts a path parameter as int.
//
//	id, err := core.ParamInt(c, "id")
func ParamInt(c Context, key string) (int, error) {
	val := c.Param(key)
	if val == "" {
		return 0, ErrBadRequest(fmt.Sprintf("path parameter '%s' is required", key))
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, ErrBadRequest(fmt.Sprintf("path parameter '%s' must be an integer", key))
	}
	return n, nil
}

// ParamInt64 extracts a path parameter as int64.
//
//	id, err := core.ParamInt64(c, "id")
func ParamInt64(c Context, key string) (int64, error) {
	val := c.Param(key)
	if val == "" {
		return 0, ErrBadRequest(fmt.Sprintf("path parameter '%s' is required", key))
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, ErrBadRequest(fmt.Sprintf("path parameter '%s' must be an integer", key))
	}
	return n, nil
}

// ─── Single Query Extractors ────────────────────────────────────────────────

// Query extracts a single query parameter as string.
// Returns the default if the parameter is empty.
//
//	sort := core.Query(c, "sort", "created_at")
func Query(c Context, key string, defaultValue ...string) string {
	def := ""
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	return c.QueryDefault(key, def)
}

// QueryInt extracts a single query parameter as int.
//
//	page, err := core.QueryInt(c, "page", 1)
func QueryInt(c Context, key string, defaultValue ...int) (int, error) {
	def := 0
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	raw := c.Query(key)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, ErrBadRequest(fmt.Sprintf("query parameter '%s' must be an integer", key))
	}
	return n, nil
}

// ─── Header Extractor ───────────────────────────────────────────────────────

// Header extracts a request header. Returns error if required and missing.
//
//	token, err := core.Header(c, "Authorization", true)
func Header(c Context, key string, required bool) (string, error) {
	val := c.GetHeader(key)
	if val == "" && required {
		return "", ErrBadRequest(fmt.Sprintf("header '%s' is required", key))
	}
	return val, nil
}

// ─── Composite Extractor (multiple sources at once) ─────────────────────────

// RequestData holds extracted data from multiple sources.
// Use ExtractRequest to populate it.
type RequestData struct {
	Params  map[string]string
	Queries map[string]string
	Headers map[string]string
}

// ─── Validation Helpers ─────────────────────────────────────────────────────

// Required checks that a string field is not empty.
func Required(field, name string) error {
	if strings.TrimSpace(field) == "" {
		return ErrBadRequest(fmt.Sprintf("'%s' is required", name))
	}
	return nil
}

// MinLength checks that a string field meets minimum length.
func MinLength(field, name string, min int) error {
	if len(field) < min {
		return ErrBadRequest(fmt.Sprintf("'%s' must be at least %d characters", name, min))
	}
	return nil
}

// MaxLength checks that a string field doesn't exceed maximum length.
func MaxLength(field, name string, max int) error {
	if len(field) > max {
		return ErrBadRequest(fmt.Sprintf("'%s' must be at most %d characters", name, max))
	}
	return nil
}

// InRange checks that an int is within a range.
func InRange(value int, name string, min, max int) error {
	if value < min || value > max {
		return ErrBadRequest(fmt.Sprintf("'%s' must be between %d and %d", name, min, max))
	}
	return nil
}
