package config

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// setField sets a field from a string value (env var).
func setField(field reflect.Value, val string) error {
	if !field.CanSet() {
		return nil
	}

	// Handle time.Duration specifically (it's an int64 alias that needs ParseDuration).
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("cannot parse %q as time.Duration: %w", val, err)
		}
		field.Set(reflect.ValueOf(d))
		return nil
	}

	// Handle time.Time specifically (it's a struct that can unmarshal from text).
	if field.Type() == reflect.TypeOf(time.Time{}) {
		var t time.Time
		if err := t.UnmarshalText([]byte(val)); err != nil {
			return fmt.Errorf("cannot parse %q as time.Time: %w", val, err)
		}
		field.Set(reflect.ValueOf(t))
		return nil
	}

	// Handle pointers by allocating and delegating.
	if field.Kind() == reflect.Ptr {
		ptr := reflect.New(field.Type().Elem())
		if err := setField(ptr.Elem(), val); err != nil {
			return err
		}
		field.Set(ptr)
		return nil
	}

	// Try TextUnmarshaler first (covers time.Duration, time.Time, custom types).
	if tu, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
		return tu.UnmarshalText([]byte(val))
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("cannot parse %q as bool: %w", val, err)
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bitSize := field.Type().Bits()
		n, err := strconv.ParseInt(val, 10, bitSize)
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", val, field.Type(), err)
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", val, field.Type(), err)
		}
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", val, field.Type(), err)
		}
		field.SetFloat(n)

	case reflect.Slice:
		return setSliceFromEnv(field, val)

	case reflect.Map:
		return setMapFromEnv(field, val)

	default:
		return fmt.Errorf("unsupported type %s for env var", field.Type())
	}
	return nil
}

// setFieldFromAny sets a field from a decoded value (JSON/YAML).
func setFieldFromAny(field reflect.Value, val any) error {
	if !field.CanSet() {
		return nil
	}

	// Nil means "not set" — skip.
	if val == nil {
		return nil
	}

	// Direct assignment if types match.
	rv := reflect.ValueOf(val)
	if rv.Type().AssignableTo(field.Type()) {
		field.Set(rv)
		return nil
	}

	// Handle time.Duration specifically.
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("cannot convert %T to time.Duration", val)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("cannot parse %q as time.Duration: %w", s, err)
		}
		field.Set(reflect.ValueOf(d))
		return nil
	}

	// Handle pointers.
	if field.Kind() == reflect.Ptr {
		ptr := reflect.New(field.Type().Elem())
		if err := setFieldFromAny(ptr.Elem(), val); err != nil {
			return err
		}
		field.Set(ptr)
		return nil
	}

	// Try TextUnmarshaler first.
	if tu, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
		switch s := val.(type) {
		case string:
			return tu.UnmarshalText([]byte(s))
		case []byte:
			return tu.UnmarshalText(s)
		}
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", val))

	case reflect.Bool:
		switch v := val.(type) {
		case bool:
			field.SetBool(v)
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("cannot parse %q as bool: %w", v, err)
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("cannot convert %T to bool", val)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(val)
		if err != nil {
			return err
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch v := val.(type) {
		case uint64:
			field.SetUint(v)
		case int64:
			if v < 0 {
				return fmt.Errorf("negative value %d for %s", v, field.Type())
			}
			field.SetUint(uint64(v))
		case float64:
			if v < 0 {
				return fmt.Errorf("negative value %f for %s", v, field.Type())
			}
			field.SetUint(uint64(v))
		case string:
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as %s: %w", v, field.Type(), err)
			}
			field.SetUint(n)
		default:
			return fmt.Errorf("cannot convert %T to %s", val, field.Type())
		}

	case reflect.Float32, reflect.Float64:
		f, err := toFloat64(val)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case reflect.Slice:
		return setSliceFromAny(field, val)

	case reflect.Map:
		return setMapFromAny(field, val)

	default:
		return fmt.Errorf("unsupported type %s", field.Type())
	}
	return nil
}

func toInt64(val any) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case uint64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}

func toFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", val)
	}
}

func setSliceFromEnv(field reflect.Value, val string) error {
	parts := strings.Split(val, ",")
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		elem := slice.Index(i)
		if err := setField(elem, p); err != nil {
			return fmt.Errorf("slice element %d: %w", i, err)
		}
	}
	field.Set(slice)
	return nil
}

func setSliceFromAny(field reflect.Value, val any) error {
	slice, ok := val.([]any)
	if !ok {
		return fmt.Errorf("expected array, got %T", val)
	}
	elemType := field.Type().Elem()
	result := reflect.MakeSlice(field.Type(), len(slice), len(slice))
	for i, item := range slice {
		tmp := reflect.New(elemType).Elem()
		if err := setFieldFromAny(tmp, item); err != nil {
			return fmt.Errorf("slice element %d: %w", i, err)
		}
		result.Index(i).Set(tmp)
	}
	field.Set(result)
	return nil
}

func setMapFromEnv(field reflect.Value, val string) error {
	elemType := field.Type().Elem()
	if elemType.Kind() != reflect.String {
		return fmt.Errorf("env var map only supports map[string]string, got map[string]%s", elemType)
	}

	m := reflect.MakeMap(field.Type())
	for _, pair := range strings.Split(val, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return fmt.Errorf("invalid map entry %q, expected key=value", pair)
		}
		m.SetMapIndex(reflect.ValueOf(strings.TrimSpace(k)), reflect.ValueOf(strings.TrimSpace(v)))
	}
	field.Set(m)
	return nil
}

func setMapFromAny(field reflect.Value, val any) error {
	m, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("expected object, got %T", val)
	}

	elemType := field.Type().Elem()
	result := reflect.MakeMapWithSize(field.Type(), len(m))

	for k, v := range m {
		kv := reflect.ValueOf(k)
		ev := reflect.New(elemType).Elem()
		if err := setFieldFromAny(ev, v); err != nil {
			return fmt.Errorf("map key %q: %w", k, err)
		}
		result.SetMapIndex(kv, ev)
	}
	field.Set(result)
	return nil
}
