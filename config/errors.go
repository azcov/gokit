package config

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, fe := range e.Errors {
		msgs[i] = fe.Error()
	}
	return "validation: " + strings.Join(msgs, "; ")
}

type FieldError struct {
	Path  string
	Tag   string
	Value any
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s (current value: %v)", e.Path, e.Tag, e.Value)
}

type SourceError struct {
	Source string
	Path   string
	Err    error
}

func (e *SourceError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("source %s: %v", e.Source, e.Err)
	}
	return fmt.Sprintf("source %s (%s): %v", e.Source, e.Path, e.Err)
}

func (e *SourceError) Unwrap() error {
	return e.Err
}
