package config

import (
	"fmt"
	"reflect"
	"strings"
)

func collectFields(target any) ([]fieldInfo, error) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, fmt.Errorf("target must be a non-nil pointer, got %T", target)
	}
	return walkFields(v.Elem(), nil)
}

func walkFields(v reflect.Value, prefix []string) ([]fieldInfo, error) {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, nil
	}

	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		if f.Anonymous {
			sub, err := walkFields(fv, prefix)
			if err != nil {
				return nil, err
			}
			fields = append(fields, sub...)
			continue
		}

		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("config")

		if tag == "-" {
			continue
		}

		name := tag
		if name == "" {
			name = strings.ToLower(f.Name)
		}

		path := append(append([]string{}, prefix...), name)
		ft := resolveType(f.Type)

		if ft.Kind() == reflect.Struct && isStructType(ft) {
			actual := fv
			if f.Type.Kind() == reflect.Ptr && fv.IsNil() {
				actual = reflect.New(ft)
				fv.Set(actual)
			}
			if f.Type.Kind() == reflect.Ptr {
				actual = actual.Elem()
			}
			sub, err := walkFields(actual, path)
			if err != nil {
				return nil, err
			}
			fields = append(fields, sub...)
		} else {
			fields = append(fields, fieldInfo{
				path:  path,
				value: fv,
			})
		}
	}
	return fields, nil
}
