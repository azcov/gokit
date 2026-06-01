package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func validate(target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	var errs []FieldError
	validateRecursive(v, nil, &errs)

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

func validateRecursive(v reflect.Value, prefix []string, errs *[]FieldError) {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("config")
		if tag == "-" {
			continue
		}

		validateTag := f.Tag.Get("validate")

		name := tag
		if name == "" && !f.Anonymous {
			name = strings.ToLower(f.Name)
		}

		path := append(append([]string{}, prefix...), name)

		if f.Anonymous {
			validateRecursive(fv, prefix, errs)
			continue
		}

		ft := resolveType(f.Type)

		if ft.Kind() == reflect.Struct && isStructType(ft) {
			actual := fv
			if f.Type.Kind() == reflect.Ptr && !fv.IsNil() {
				actual = fv.Elem()
			}
			if f.Type.Kind() != reflect.Ptr || !fv.IsNil() {
				validateRecursive(actual, path, errs)
			}
			continue
		}

		if validateTag == "" {
			continue
		}

		p := pathKey(path)

		for _, rule := range strings.Split(validateTag, ",") {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			switch {
			case rule == "required":
				if fv.IsZero() {
					*errs = append(*errs, FieldError{Path: p, Tag: "required", Value: fv.Interface()})
				}

			case strings.HasPrefix(rule, "min="):
				val := strings.TrimPrefix(rule, "min=")
				validateMin(fv, val, p, errs)

			case strings.HasPrefix(rule, "max="):
				val := strings.TrimPrefix(rule, "max=")
				validateMax(fv, val, p, errs)

			case strings.HasPrefix(rule, "len="):
				val := strings.TrimPrefix(rule, "len=")
				validateLen(fv, val, p, errs)

			case strings.HasPrefix(rule, "oneof="):
				opts := strings.TrimPrefix(rule, "oneof=")
				validateOneOf(fv, opts, p, errs)
			}
		}
	}
}

func validateMin(fv reflect.Value, val string, path string, errs *[]FieldError) {
	minVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return
	}
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fv.Int()) < minVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("min=%s", val), Value: fv.Interface()})
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(fv.Uint()) < minVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("min=%s", val), Value: fv.Interface()})
		}
	case reflect.Float32, reflect.Float64:
		if fv.Float() < minVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("min=%s", val), Value: fv.Interface()})
		}
	case reflect.String:
		if float64(len(fv.String())) < minVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("min=%s", val), Value: fv.Interface()})
		}
	}
}

func validateMax(fv reflect.Value, val string, path string, errs *[]FieldError) {
	maxVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return
	}
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fv.Int()) > maxVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("max=%s", val), Value: fv.Interface()})
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(fv.Uint()) > maxVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("max=%s", val), Value: fv.Interface()})
		}
	case reflect.Float32, reflect.Float64:
		if fv.Float() > maxVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("max=%s", val), Value: fv.Interface()})
		}
	case reflect.String:
		if float64(len(fv.String())) > maxVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("max=%s", val), Value: fv.Interface()})
		}
	}
}

func validateLen(fv reflect.Value, val string, path string, errs *[]FieldError) {
	lenVal, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	switch fv.Kind() {
	case reflect.String:
		if len(fv.String()) != lenVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("len=%s", val), Value: fv.Interface()})
		}
	case reflect.Slice, reflect.Array:
		if fv.Len() != lenVal {
			*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("len=%s", val), Value: fv.Interface()})
		}
	}
}

func validateOneOf(fv reflect.Value, opts string, path string, errs *[]FieldError) {
	if fv.Kind() != reflect.String {
		return
	}
	vals := strings.Fields(opts)
	s := fv.String()
	for _, v := range vals {
		if s == v {
			return
		}
	}
	*errs = append(*errs, FieldError{Path: path, Tag: fmt.Sprintf("oneof=%s", opts), Value: s})
}
