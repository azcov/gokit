package config

import (
	"testing"
)

type validateRequired struct {
	Name string `config:"name" validate:"required"`
}

type validateNested struct {
	Inner validateRequired `config:"inner"`
	Opt   string           `config:"opt"`
}

type validateMinMax struct {
	Count int     `config:"count" validate:"min=1,max=100"`
	Score float64 `config:"score" validate:"min=0,max=10"`
	Name  string  `config:"name" validate:"min=2"`
}

type valLenTest struct {
	Code string `config:"code" validate:"len=5"`
}

type valOneOfTest struct {
	Env string `config:"env" validate:"oneof=prod staging dev"`
}

type validateMultiple struct {
	A string `config:"a" validate:"required"`
	B string `config:"b" validate:"required"`
	C string `config:"c"`
}

func TestValidateRequired(t *testing.T) {
	t.Run("present - no error", func(t *testing.T) {
		cfg := &validateRequired{Name: "hello"}
		if err := validate(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing - error", func(t *testing.T) {
		cfg := &validateRequired{}
		err := validate(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Fatalf("expected *ValidationError, got %T", err)
		}
		if len(ve.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(ve.Errors))
		}
		if ve.Errors[0].Tag != "required" {
			t.Fatalf("tag = %q, want 'required'", ve.Errors[0].Tag)
		}
	})
}

func TestValidateNestedRequired(t *testing.T) {
	cfg := &validateNested{Opt: "yes"}
	err := validate(cfg)
	if err == nil {
		t.Fatal("expected error for missing nested required field")
	}
}

func TestValidateMinMax(t *testing.T) {
	t.Run("within range - ok", func(t *testing.T) {
		cfg := &validateMinMax{Count: 50, Score: 5.0, Name: "hello"}
		if err := validate(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("below min - error", func(t *testing.T) {
		cfg := &validateMinMax{Count: 0, Score: 5.0, Name: "hello"}
		err := validate(cfg)
		if err == nil {
			t.Fatal("expected error for count below min")
		}
	})

	t.Run("above max - error", func(t *testing.T) {
		cfg := &validateMinMax{Count: 50, Score: 11.0, Name: "hello"}
		err := validate(cfg)
		if err == nil {
			t.Fatal("expected error for score above max")
		}
	})

	t.Run("string min length - error", func(t *testing.T) {
		cfg := &validateMinMax{Count: 50, Score: 5.0, Name: "x"}
		err := validate(cfg)
		if err == nil {
			t.Fatal("expected error for name below min length")
		}
	})
}

func TestValidateLen(t *testing.T) {
	t.Run("correct length - ok", func(t *testing.T) {
		cfg := &valLenTest{Code: "abcde"}
		if err := validate(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wrong length - error", func(t *testing.T) {
		cfg := &valLenTest{Code: "abcd"}
		err := validate(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestValidateOneOf(t *testing.T) {
	t.Run("valid option - ok", func(t *testing.T) {
		cfg := &valOneOfTest{Env: "prod"}
		if err := validate(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid option - error", func(t *testing.T) {
		cfg := &valOneOfTest{Env: "invalid"}
		err := validate(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestValidateMultipleErrors(t *testing.T) {
	cfg := &validateMultiple{}
	err := validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidateErrorFormat(t *testing.T) {
	cfg := &validateRequired{}
	err := validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if msg == "" {
		t.Fatal("empty error message")
	}
}

func TestValidateNonStruct(t *testing.T) {
	var s string
	if err := validate(&s); err != nil {
		t.Fatalf("unexpected error for non-struct: %v", err)
	}
}
