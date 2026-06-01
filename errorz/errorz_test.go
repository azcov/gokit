package errorz_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/azcov/gokit/errorz"
)

var (
	ErrNotFound     = errorz.New(404, 5, "NOT_FOUND", "resource not found")
	ErrUnauthorized = errorz.New(401, 16, "UNAUTHORIZED", "unauthorized")
)

func TestError_Message(t *testing.T) {
	if ErrNotFound.Error() != "resource not found" {
		t.Errorf("got %q", ErrNotFound.Error())
	}
}

func TestError_SatisfiesErrorInterface(t *testing.T) {
	var err error = errorz.New(500, 13, "X", "boom")
	if err.Error() != "boom" {
		t.Errorf("got %q", err.Error())
	}
}

func TestWrap_AndUnwrap(t *testing.T) {
	wrapped := ErrNotFound.Wrap(sql.ErrNoRows)
	if !strings.Contains(wrapped.Error(), "sql: no rows") {
		t.Errorf("wrapped message should include cause, got %q", wrapped.Error())
	}
	if !errors.Is(wrapped, sql.ErrNoRows) {
		t.Error("errors.Is should find the wrapped cause")
	}
	// original sentinel must be unchanged (Wrap clones)
	if ErrNotFound.Cause() != nil {
		t.Error("Wrap must not mutate the original")
	}
}

func TestIs_ByCode(t *testing.T) {
	wrapped := ErrNotFound.Wrap(sql.ErrNoRows)
	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("errors.Is should match by code even after Wrap")
	}
	if errors.Is(wrapped, ErrUnauthorized) {
		t.Error("should not match a different code")
	}
}

func TestPackageIs(t *testing.T) {
	if !errorz.Is(ErrNotFound, ErrNotFound) {
		t.Error("errorz.Is should match same sentinel")
	}
	if !errorz.Is(ErrNotFound.Wrap(sql.ErrNoRows), sql.ErrNoRows) {
		t.Error("errorz.Is should walk to the cause")
	}
}

func TestHasCode(t *testing.T) {
	wrapped := ErrNotFound.Wrap(sql.ErrNoRows)
	if !errorz.HasCode(wrapped, "NOT_FOUND") {
		t.Error("HasCode should find the code in the chain")
	}
	if errorz.HasCode(wrapped, "OTHER") {
		t.Error("HasCode should not match a missing code")
	}
}

func TestEqualNotEqual(t *testing.T) {
	if !errorz.Equal(ErrNotFound, ErrNotFound) {
		t.Error("Equal should be true for same code")
	}
	if errorz.NotEqual(ErrNotFound, ErrNotFound) {
		t.Error("NotEqual should be false for same code")
	}
	if !errorz.NotEqual(ErrNotFound, ErrUnauthorized) {
		t.Error("NotEqual should be true for different codes")
	}
}

func TestFrom(t *testing.T) {
	if errorz.From(sql.ErrNoRows) != nil {
		t.Error("From should return nil for a non-errorz error")
	}
	e := errorz.From(ErrNotFound.Wrap(sql.ErrNoRows))
	if e == nil || e.Code != "NOT_FOUND" {
		t.Errorf("From should extract the *Error, got %+v", e)
	}
}

func TestErr_NilSafe(t *testing.T) {
	var e *errorz.Error
	if e.Err() != nil {
		t.Error("Err() on nil *Error must return a true nil interface")
	}
}

func TestWithDetail(t *testing.T) {
	e := errorz.New(400, 3, "BAD", "bad").WithDetail("field", "email")
	if e.Details["field"] != "email" {
		t.Errorf("detail not set: %+v", e.Details)
	}
}

func TestLogValue(t *testing.T) {
	e := ErrNotFound.Wrap(sql.ErrNoRows).WithDetail("id", "u1")
	lv := e.LogValue()
	if lv["code"] != "NOT_FOUND" || lv["http_status"] != 404 {
		t.Errorf("unexpected LogValue: %+v", lv)
	}
	if lv["cause"] != "sql: no rows in result set" {
		t.Errorf("cause missing: %+v", lv["cause"])
	}
}

func TestLog(t *testing.T) {
	line := ErrNotFound.Log()
	if !strings.Contains(line, "code=NOT_FOUND") || !strings.Contains(line, "http=404") {
		t.Errorf("unexpected Log: %q", line)
	}
	wrapped := ErrNotFound.Wrap(sql.ErrNoRows).Log()
	if !strings.Contains(wrapped, "cause=") {
		t.Errorf("wrapped Log should include cause: %q", wrapped)
	}
}
