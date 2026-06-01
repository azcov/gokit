package errorz

import (
	"errors"
	"fmt"
)

// Error is a structured error carrying an HTTP status, RPC status, machine-readable
// code, human-readable message, optional key-value details, and an optional cause.
type Error struct {
	HttpStatus int            `json:"http_status"`
	RPCStatus  int            `json:"rpc_status"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	cause      error
}

// ── core interface ────────────────────────────────────────────────────────────

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.cause.Error())
	}
	return e.Message
}

// Unwrap lets errors.Is / errors.As walk the cause chain.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Is makes errors.Is(err, target) compare by Code when target is *Error,
// so sentinel variables work even after Wrap creates a clone.
//
//	errors.Is(err, ErrNotFound)          // true when err.Code == "NOT_FOUND"
//	errors.Is(err, errors.New("other"))  // falls through to pointer equality
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	var t *Error
	// Only do code comparison when target is an *Error; for stdlib/other errors
	// return false so errors.Is continues walking via Unwrap.
	if errors.As(target, &t) {
		return t != nil && e.Code == t.Code
	}
	return false
}

// ── constructors ──────────────────────────────────────────────────────────────

func New(httpStatus, rpcStatus int, code, message string) *Error {
	return &Error{
		HttpStatus: httpStatus,
		RPCStatus:  rpcStatus,
		Code:       code,
		Message:    message,
		Details:    make(map[string]any),
	}
}

// Wrap returns a shallow clone with cause attached. The original sentinel is
// unchanged — safe to use as a package-level var.
func (e *Error) Wrap(cause error) *Error {
	if e == nil {
		return nil
	}
	clone := *e
	clone.cause = cause
	clone.Details = make(map[string]any, len(e.Details))
	for k, v := range e.Details {
		clone.Details[k] = v
	}
	return &clone
}

func (e *Error) WithDetail(key string, value any) *Error {
	e.Details[key] = value
	return e
}

func (e *Error) WithDetails(details map[string]any) *Error {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// ── comparison helpers ────────────────────────────────────────────────────────

// Equal reports whether err matches target by Code (when target is *Error)
// or by standard errors.Is semantics otherwise.
//
//	errorz.Equal(err, ErrNotFound)     // code comparison
//	errorz.Equal(err, sql.ErrNoRows)   // pointer/value equality via errors.Is
func Equal(err, target error) bool {
	return errors.Is(err, target)
}

// NotEqual is the inverse of Equal.
func NotEqual(err, target error) bool {
	return !errors.Is(err, target)
}

// HasCode reports whether any error in err's chain is an *Error with code.
//
//	errorz.HasCode(err, "NOT_FOUND")
func HasCode(err error, code string) bool {
	var e *Error
	for err != nil {
		if errors.As(err, &e) && e.Code == code {
			return true
		}
		// Step past the *Error to its cause to continue walking.
		err = errors.Unwrap(err)
	}
	return false
}

// Is is the package-level shorthand for errors.Is.
// It is here so callers only need to import errorz, not errors.
//
//	errorz.Is(err, ErrNotFound)
//	errorz.Is(err, sql.ErrNoRows)
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As is the package-level shorthand for errors.As for *Error targets.
//
//	var e *errorz.Error
//	if errorz.As(err, &e) { ... }
func As(err error, target **Error) bool {
	return errors.As(err, target)
}

// From extracts the first *Error in err's chain, or nil.
func From(err error) *Error {
	var e *Error
	errors.As(err, &e)
	return e
}

// Cause returns the direct cause attached via Wrap, or nil.
func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Err returns e as a plain error interface value.
// Useful when the compiler needs an explicit error, not *Error:
//
//	return ErrNotFound.Wrap(cause).Err()
//	ch <- ErrInternal.Err()
func (e *Error) Err() error {
	if e == nil {
		return nil
	}
	return e
}

// ── equality on the struct itself ─────────────────────────────────────────────

// Equal reports whether e has the same Code as other.
// Prefer the package-level Equal / errors.Is for chain-aware comparison.
func (e *Error) Equal(other error) bool {
	if e == nil {
		return other == nil
	}
	return e.Is(other)
}

// NotEqual reports whether e does not share a Code with other.
func (e *Error) NotEqual(other error) bool {
	return !e.Equal(other)
}

// ── logging ───────────────────────────────────────────────────────────────────

// LogValue returns a structured map for structured loggers.
func (e *Error) LogValue() map[string]any {
	if e == nil {
		return map[string]any{"error": nil}
	}
	m := map[string]any{
		"code":        e.Code,
		"message":     e.Message,
		"http_status": e.HttpStatus,
		"rpc_status":  e.RPCStatus,
	}
	if len(e.Details) > 0 {
		m["details"] = e.Details
	}
	if e.cause != nil {
		m["cause"] = e.cause.Error()
	}
	return m
}

// Log returns a single key=value line for text loggers.
func (e *Error) Log() string {
	if e == nil {
		return "error=nil"
	}
	if e.cause != nil {
		return fmt.Sprintf("code=%s http=%d rpc=%d message=%q cause=%q",
			e.Code, e.HttpStatus, e.RPCStatus, e.Message, e.cause.Error())
	}
	return fmt.Sprintf("code=%s http=%d rpc=%d message=%q",
		e.Code, e.HttpStatus, e.RPCStatus, e.Message)
}
