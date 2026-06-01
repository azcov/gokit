package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type benchConfig struct {
	Name    string      `config:"name"`
	Count   int         `config:"count"`
	Enabled bool        `config:"enabled"`
	Nested  benchNested `config:"nested"`
	Deep    benchDeep   `config:"deep"`
}

type benchNested struct {
	Key   string  `config:"key"`
	Value float64 `config:"value"`
}

type benchDeep struct {
	Level1 benchLevel1 `config:"level1"`
}

type benchLevel1 struct {
	Level2 benchLevel2 `config:"level2"`
}

type benchLevel2 struct {
	Value string `config:"value"`
}

func BenchmarkCollectFields(b *testing.B) {
	cfg := &benchConfig{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectFields(cfg)
	}
}

func BenchmarkSetFieldString(b *testing.B) {
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setField(reflect.ValueOf(&s).Elem(), "hello")
	}
}

func BenchmarkSetFieldInt(b *testing.B) {
	var n int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setField(reflect.ValueOf(&n).Elem(), "42")
	}
}

func BenchmarkEnvLoad(b *testing.B) {
	os.Setenv("NAME", "bench")
	os.Setenv("COUNT", "100")
	os.Setenv("ENABLED", "true")
	os.Setenv("NESTED_KEY", "k")
	os.Setenv("NESTED_VALUE", "1.5")
	os.Setenv("DEEP_LEVEL1_LEVEL2_VALUE", "deepval")
	defer os.Unsetenv("NAME")
	defer os.Unsetenv("COUNT")
	defer os.Unsetenv("ENABLED")
	defer os.Unsetenv("NESTED_KEY")
	defer os.Unsetenv("NESTED_VALUE")
	defer os.Unsetenv("DEEP_LEVEL1_LEVEL2_VALUE")

	src := &envSource{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := &benchConfig{}
		src.Load(cfg)
	}
}

func BenchmarkYamlLoad(b *testing.B) {
	dir := b.TempDir()
	f := filepath.Join(dir, "bench.yaml")
	os.WriteFile(f, []byte(`
name: bench
count: 100
enabled: true
nested:
  key: k
  value: 1.5
deep:
  level1:
    level2:
      value: deepval
`), 0644)

	src := &yamlSource{path: f}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := &benchConfig{}
		src.Load(cfg)
	}
}

func BenchmarkJsonLoad(b *testing.B) {
	dir := b.TempDir()
	f := filepath.Join(dir, "bench.json")
	os.WriteFile(f, []byte(`{"name":"bench","count":100,"enabled":true,"nested":{"key":"k","value":1.5},"deep":{"level1":{"level2":{"value":"deepval"}}}}`), 0644)

	src := &jsonSource{path: f}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := &benchConfig{}
		src.Load(cfg)
	}
}

func BenchmarkNestedMapLookup(b *testing.B) {
	data := map[string]any{
		"deep": map[string]any{
			"level1": map[string]any{
				"level2": map[string]any{
					"value": "found",
				},
			},
		},
	}
	path := []string{"deep", "level1", "level2", "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lookupNested(data, path)
	}
}
