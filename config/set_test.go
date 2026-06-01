package config

import (
	"reflect"
	"testing"
	"time"
)

type textUnmarshalerCustom struct {
	Value string
}

func (c *textUnmarshalerCustom) UnmarshalText(text []byte) error {
	c.Value = "custom:" + string(text)
	return nil
}

type testStruct struct {
	Name string `config:"name"`
}

func TestSetField(t *testing.T) {
	tests := []struct {
		name    string
		target  func() any
		value   string
		want    any
		wantErr bool
	}{
		{name: "string", target: func() any { return new(string) }, value: "hello", want: "hello"},
		{name: "bool_true", target: func() any { return new(bool) }, value: "true", want: true},
		{name: "bool_false", target: func() any { return new(bool) }, value: "false", want: false},
		{name: "bool_1", target: func() any { return new(bool) }, value: "1", want: true},
		{name: "int", target: func() any { return new(int) }, value: "42", want: int(42)},
		{name: "int8", target: func() any { return new(int8) }, value: "127", want: int8(127)},
		{name: "int16", target: func() any { return new(int16) }, value: "32767", want: int16(32767)},
		{name: "int32", target: func() any { return new(int32) }, value: "2147483647", want: int32(2147483647)},
		{name: "int64", target: func() any { return new(int64) }, value: "9223372036854775807", want: int64(9223372036854775807)},
		{name: "uint", target: func() any { return new(uint) }, value: "100", want: uint(100)},
		{name: "uint8", target: func() any { return new(uint8) }, value: "255", want: uint8(255)},
		{name: "uint16", target: func() any { return new(uint16) }, value: "65535", want: uint16(65535)},
		{name: "uint32", target: func() any { return new(uint32) }, value: "4294967295", want: uint32(4294967295)},
		{name: "uint64", target: func() any { return new(uint64) }, value: "18446744073709551615", want: uint64(18446744073709551615)},
		{name: "float32", target: func() any { return new(float32) }, value: "3.14", want: float32(3.14)},
		{name: "float64", target: func() any { return new(float64) }, value: "2.71828", want: 2.71828},
		{name: "duration", target: func() any { return new(time.Duration) }, value: "10s", want: time.Duration(10 * time.Second)},
		{name: "time", target: func() any { return new(time.Time) }, value: "2024-01-01T00:00:00Z", want: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{name: "text_unmarshaler", target: func() any { return &textUnmarshalerCustom{} }, value: "test", want: textUnmarshalerCustom{Value: "custom:test"}},
		{name: "error_invalid_int", target: func() any { return new(int) }, value: "abc", wantErr: true},
		{name: "error_overflow", target: func() any { return new(int8) }, value: "999", wantErr: true},
		{name: "error_bool", target: func() any { return new(bool) }, value: "notbool", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := tt.target()
			err := setField(reflect.ValueOf(target).Elem(), tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := reflect.ValueOf(target).Elem().Interface()
			if tt.name == "time" {
				gotTime := got.(time.Time)
				wantTime := tt.want.(time.Time)
				if !gotTime.Equal(wantTime) {
					t.Fatalf("got %v, want %v", gotTime, wantTime)
				}
			} else if tt.name == "float32" {
				gotF := got.(float32)
				wantF := tt.want.(float32)
				if gotF != wantF {
					t.Fatalf("got %v, want %v", gotF, wantF)
				}
		} else if tt.name == "text_unmarshaler" {
			gotC := got.(textUnmarshalerCustom)
			wantC := tt.want.(textUnmarshalerCustom)
			if gotC.Value != wantC.Value {
				t.Fatalf("got %v, want %v", gotC.Value, wantC.Value)
			}
			} else if got != tt.want {
				t.Fatalf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestSetFieldFromAny(t *testing.T) {
	tests := []struct {
		name    string
		target  func() any
		value   any
		want    any
		wantErr bool
	}{
		{name: "string_from_string", target: func() any { return new(string) }, value: "hello", want: "hello"},
		{name: "bool_from_bool", target: func() any { return new(bool) }, value: true, want: true},
		{name: "bool_from_string", target: func() any { return new(bool) }, value: "true", want: true},
		{name: "int_from_int", target: func() any { return new(int) }, value: 42, want: int(42)},
		{name: "int_from_float64", target: func() any { return new(int) }, value: float64(42), want: int(42)},
		{name: "float_from_float", target: func() any { return new(float64) }, value: 3.14, want: 3.14},
		{name: "float_from_int", target: func() any { return new(float64) }, value: int64(42), want: 42.0},
		{name: "duration_from_string", target: func() any { return new(time.Duration) }, value: "5m", want: time.Duration(5 * time.Minute)},
		{name: "nil_value_skips", target: func() any { return new(string) }, value: nil, want: ""},
		{name: "uint_from_negative", target: func() any { return new(uint) }, value: int64(-1), wantErr: true},
		{name: "unsupported_type", target: func() any { return new(struct{}) }, value: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := tt.target()
			err := setFieldFromAny(reflect.ValueOf(target).Elem(), tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := reflect.ValueOf(target).Elem().Interface()
			if tt.name == "duration_from_string" {
				gotD := got.(time.Duration)
				wantD := tt.want.(time.Duration)
				if gotD != wantD {
					t.Fatalf("got %v, want %v", gotD, wantD)
				}
			} else if got != tt.want {
				t.Fatalf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestSetSliceFromEnv(t *testing.T) {
	var s []string
	if err := setField(reflect.ValueOf(&s).Elem(), "a,b,c"); err != nil {
		t.Fatal(err)
	}
	if len(s) != 3 || s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Fatalf("got %v", s)
	}
}

func TestSetSliceFromAny(t *testing.T) {
	t.Run("string slice", func(t *testing.T) {
		var s []string
		if err := setFieldFromAny(reflect.ValueOf(&s).Elem(), []any{"a", "b"}); err != nil {
			t.Fatal(err)
		}
		if len(s) != 2 || s[0] != "a" || s[1] != "b" {
			t.Fatalf("got %v", s)
		}
	})

	t.Run("non-array value errors", func(t *testing.T) {
		var s []string
		err := setFieldFromAny(reflect.ValueOf(&s).Elem(), "not an array")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSetMapFromEnv(t *testing.T) {
	var m map[string]string
	if err := setField(reflect.ValueOf(&m).Elem(), "k1=v1,k2=v2"); err != nil {
		t.Fatal(err)
	}
	if m["k1"] != "v1" || m["k2"] != "v2" {
		t.Fatalf("got %v", m)
	}
}

func TestSetMapFromAny(t *testing.T) {
	t.Run("string map", func(t *testing.T) {
		var m map[string]string
		if err := setFieldFromAny(reflect.ValueOf(&m).Elem(), map[string]any{"k1": "v1"}); err != nil {
			t.Fatal(err)
		}
		if m["k1"] != "v1" {
			t.Fatalf("got %v", m)
		}
	})

	t.Run("non-map value errors", func(t *testing.T) {
		var m map[string]string
		err := setFieldFromAny(reflect.ValueOf(&m).Elem(), "not a map")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestToInt64(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		v, err := toInt64(int(42))
		if err != nil || v != 42 {
			t.Fatalf("got %d, %v", v, err)
		}
	})
	t.Run("float64", func(t *testing.T) {
		v, err := toInt64(float64(42))
		if err != nil || v != 42 {
			t.Fatalf("got %d, %v", v, err)
		}
	})
	t.Run("string", func(t *testing.T) {
		v, err := toInt64("42")
		if err != nil || v != 42 {
			t.Fatalf("got %d, %v", v, err)
		}
	})
	t.Run("uint64", func(t *testing.T) {
		v, err := toInt64(uint64(42))
		if err != nil || v != 42 {
			t.Fatalf("got %d, %v", v, err)
		}
	})
	t.Run("unsupported_type", func(t *testing.T) {
		_, err := toInt64(struct{}{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestToFloat64(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		v, err := toFloat64(int64(42))
		if err != nil || v != 42.0 {
			t.Fatalf("got %f, %v", v, err)
		}
	})
	t.Run("int", func(t *testing.T) {
		v, err := toFloat64(int(42))
		if err != nil || v != 42.0 {
			t.Fatalf("got %f, %v", v, err)
		}
	})
	t.Run("string", func(t *testing.T) {
		v, err := toFloat64("3.14")
		if err != nil || v != 3.14 {
			t.Fatalf("got %f, %v", v, err)
		}
	})
	t.Run("unsupported_type", func(t *testing.T) {
		_, err := toFloat64(struct{}{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSetPointerField(t *testing.T) {
	t.Run("env sets pointer string", func(t *testing.T) {
		var s *string
		if err := setField(reflect.ValueOf(&s).Elem(), "hello"); err != nil {
			t.Fatal(err)
		}
		if s == nil {
			t.Fatal("pointer not allocated")
		}
		if *s != "hello" {
			t.Fatalf("got %q", *s)
		}
	})

	t.Run("any sets pointer string", func(t *testing.T) {
		var s *string
		if err := setFieldFromAny(reflect.ValueOf(&s).Elem(), "world"); err != nil {
			t.Fatal(err)
		}
		if s == nil {
			t.Fatal("pointer not allocated")
		}
		if *s != "world" {
			t.Fatalf("got %q", *s)
		}
	})
}
