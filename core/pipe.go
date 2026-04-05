package core

import (
	"fmt"
	"strings"
)

// Pipe transforms or validates an extracted value.
type Pipe[T any] interface {
	Transform(value T, c Context) (T, error)
}

// PipeFunc is a functional adapter for Pipe.
type PipeFunc[T any] func(value T, c Context) (T, error)

func (f PipeFunc[T]) Transform(value T, c Context) (T, error) { return f(value, c) }

// WithPipes wraps an Extractor so that after extraction, each pipe
// transforms/validates the value in order.
func WithPipes[T any](extractor Extractor[T], pipes ...Pipe[T]) Extractor[T] {
	return Extractor[T]{
		Extract: func(c Context) (T, error) {
			val, err := extractor.Extract(c)
			if err != nil {
				return val, err
			}
			for _, p := range pipes {
				val, err = p.Transform(val, c)
				if err != nil {
					return val, err
				}
			}
			return val, nil
		},
	}
}

// ─── Built-in Pipes ─────────────────────────────────────────────────────────

// TrimPipe trims whitespace from string values.
var TrimPipe = PipeFunc[string](func(val string, c Context) (string, error) {
	return strings.TrimSpace(val), nil
})

// NonEmptyPipe ensures a string is not empty after trimming.
func NonEmptyPipe(paramName string) PipeFunc[string] {
	return func(val string, c Context) (string, error) {
		if strings.TrimSpace(val) == "" {
			return "", ErrBadRequest(fmt.Sprintf("'%s' must not be empty", paramName))
		}
		return val, nil
	}
}

// IntRangePipe validates that an int is within bounds.
func IntRangePipe(paramName string, min, max int) PipeFunc[int] {
	return func(val int, c Context) (int, error) {
		if val < min || val > max {
			return 0, ErrBadRequest(fmt.Sprintf("'%s' must be between %d and %d", paramName, min, max))
		}
		return val, nil
	}
}
