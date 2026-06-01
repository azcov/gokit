package config

import (
	"testing"
)

func TestLookupNested(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "value",
			},
		},
		"d": "flat",
		"e": map[string]any{
			"f": "nested",
		},
	}

	t.Run("three levels deep", func(t *testing.T) {
		got := lookupNested(data, []string{"a", "b", "c"})
		if got != "value" {
			t.Fatalf("got %v, want 'value'", got)
		}
	})

	t.Run("flat key", func(t *testing.T) {
		got := lookupNested(data, []string{"d"})
		if got != "flat" {
			t.Fatalf("got %v, want 'flat'", got)
		}
	})

	t.Run("missing key returns nil", func(t *testing.T) {
		got := lookupNested(data, []string{"nonexistent"})
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("deep missing returns nil", func(t *testing.T) {
		got := lookupNested(data, []string{"a", "b", "missing"})
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("non-map intermediate returns nil", func(t *testing.T) {
		got := lookupNested(data, []string{"d", "subkey"})
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

func TestNormalizeMap(t *testing.T) {
	t.Run("map[string]any passes through", func(t *testing.T) {
		input := map[string]any{"key": "val"}
		got := normalizeMap(input)
		if got["key"] != "val" {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("map[any]any converts to string keys", func(t *testing.T) {
		input := map[any]any{"key": "val", 42: "number"}
		got := normalizeMap(input)
		if got["key"] != "val" {
			t.Fatalf("key = %v", got["key"])
		}
		// 42 as int key becomes "" string key since toString only handles string
	})

	t.Run("non-map returns nil", func(t *testing.T) {
		got := normalizeMap("not a map")
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("nested yaml-style map normalizes recursively", func(t *testing.T) {
		input := map[any]any{
			"outer": map[any]any{
				"inner": "val",
			},
		}
		got := normalizeMap(input)
		outer, ok := got["outer"].(map[string]any)
		if !ok {
			t.Fatal("outer is not map[string]any")
		}
		if outer["inner"] != "val" {
			t.Fatalf("inner = %v", outer["inner"])
		}
	})

	t.Run("slices with maps normalize their elements", func(t *testing.T) {
		input := map[string]any{
			"items": []any{
				map[any]any{"id": "1"},
				map[any]any{"id": "2"},
			},
		}
		got := normalizeMap(input)
		items, ok := got["items"].([]any)
		if !ok {
			t.Fatal("items not a slice")
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		first, ok := items[0].(map[string]any)
		if !ok || first["id"] != "1" {
			t.Fatalf("first item = %v", items[0])
		}
	})
}
