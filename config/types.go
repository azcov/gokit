package config

import (
	"reflect"
	"strings"
)

type fieldInfo struct {
	path  []string
	value reflect.Value
}

func isStructType(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	if t.PkgPath() == "time" && (t.Name() == "Duration" || t.Name() == "Time") {
		return false
	}
	return true
}

func resolveType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func envNameFromPath(path []string) string {
	var b strings.Builder
	for i, seg := range path {
		if i > 0 {
			b.WriteByte('_')
		}
		b.WriteString(strings.ToUpper(seg))
	}
	return b.String()
}

func pathKey(path []string) string {
	return strings.Join(path, ".")
}
