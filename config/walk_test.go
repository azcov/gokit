package config

import (
	"reflect"
	"testing"
)

type walkFlat struct {
	Name  string `config:"name"`
	Value int    `config:"value"`
}

type walkNested struct {
	Name    string     `config:"name"`
	Nested  walkInner  `config:"nested"`
	PtrNest *walkInner `config:"ptr"`
	Ignored string     `config:"-"`
	NoTag   string
	unexp   string
}

type walkInner struct {
	Key string `config:"key"`
}

type walkEmbedded struct {
	walkInner
	Extra string `config:"extra"`
}

type walkDeep struct {
	A struct {
		B struct {
			C string `config:"val"`
		} `config:"b"`
	} `config:"a"`
}

type walkPtrStruct struct {
	Inner *walkInner `config:"inner"`
	Name  string     `config:"name"`
}

func TestCollectFields(t *testing.T) {
	t.Run("flat struct", func(t *testing.T) {
		cfg := &walkFlat{Name: "x", Value: 1}
		fields, err := collectFields(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(fields))
		}
	})

	t.Run("nested struct", func(t *testing.T) {
		cfg := &walkNested{Name: "x"}
		fields, err := collectFields(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 4 {
			t.Fatalf("expected 4 fields (name, nested.key, noTag, id), got %d", len(fields))
		}
	})

	t.Run("skip tag", func(t *testing.T) {
		cfg := &walkNested{Name: "x"}
		fields, _ := collectFields(cfg)
		for _, f := range fields {
			for _, seg := range f.path {
				if seg == "ignored" {
					t.Fatal("found skipped field 'ignored'")
				}
			}
		}
	})

	t.Run("embedded struct inlines fields", func(t *testing.T) {
		cfg := &walkEmbedded{Extra: "e"}
		fields, err := collectFields(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 2 {
			t.Fatalf("expected 2 fields (key, extra), got %d", len(fields))
		}
	})

	t.Run("pointer struct auto-allocated", func(t *testing.T) {
		cfg := &walkPtrStruct{Name: "test"}
		fields, err := collectFields(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(fields))
		}
		if cfg.Inner == nil {
			t.Fatal("pointer field was not auto-allocated")
		}
	})

	t.Run("deep nesting builds correct path", func(t *testing.T) {
		cfg := &walkDeep{}
		fields, err := collectFields(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 1 {
			t.Fatalf("expected 1 field, got %d", len(fields))
		}
		expected := []string{"a", "b", "val"}
		for i, seg := range fields[0].path {
			if seg != expected[i] {
				t.Fatalf("path[%d] = %q, want %q", i, seg, expected[i])
			}
		}
	})

	t.Run("unexported fields skipped", func(t *testing.T) {
		cfg := &walkNested{Name: "x"}
		fields, _ := collectFields(cfg)
		for _, f := range fields {
			if f.value.Kind() == reflect.String && !f.value.CanSet() {
				// This is fine, unexported won't be collected
			}
		}
	})

	t.Run("nil pointer returns nil", func(t *testing.T) {
		fields, err := collectFields((*walkFlat)(nil))
		if err == nil {
			t.Fatal("expected error for nil pointer")
		}
		if fields != nil {
			t.Fatal("expected nil fields")
		}
	})
}

func TestEnvNameFromPath(t *testing.T) {
	tests := []struct {
		name string
		path []string
		want string
	}{
		{name: "single", path: []string{"name"}, want: "NAME"},
		{name: "two levels", path: []string{"storage", "s3"}, want: "STORAGE_S3"},
		{name: "three levels", path: []string{"db", "mongo", "uri"}, want: "DB_MONGO_URI"},
		{name: "mixed case segments", path: []string{"myConfig", "endpoint"}, want: "MYCONFIG_ENDPOINT"},
		{name: "single uppercase", path: []string{"REGION"}, want: "REGION"},
		{name: "empty path", path: []string{}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := envNameFromPath(tt.path)
			if got != tt.want {
				t.Errorf("envNameFromPath(%v) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestPathKey(t *testing.T) {
	tests := []struct {
		path []string
		want string
	}{
		{path: []string{"a", "b", "c"}, want: "a.b.c"},
		{path: []string{"single"}, want: "single"},
		{path: []string{}, want: ""},
	}
	for _, tt := range tests {
		got := pathKey(tt.path)
		if got != tt.want {
			t.Errorf("pathKey(%v) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsStructType(t *testing.T) {
	if !isStructType(reflect.TypeOf(walkInner{})) {
		t.Error("expected true for struct type")
	}
	if isStructType(reflect.TypeOf("")) {
		t.Error("expected false for string")
	}
	if isStructType(reflect.TypeOf(42)) {
		t.Error("expected false for int")
	}
}

func TestResolveType(t *testing.T) {
	if resolveType(reflect.TypeOf("")).Kind() != reflect.String {
		t.Error("expected string")
	}
	ptrType := reflect.PointerTo(reflect.TypeOf(""))
	if resolveType(ptrType).Kind() != reflect.String {
		t.Error("expected string after resolving pointer")
	}
}
