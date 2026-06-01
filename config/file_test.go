package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

type fileTestConfig struct {
	Name    string        `config:"name"`
	Count   int           `config:"count"`
	Enabled bool          `config:"enabled"`
	Nested  fileNestedCfg `config:"nested"`
}

type fileNestedCfg struct {
	Key   string  `config:"key"`
	Value float64 `config:"value"`
}

func TestYamlBasic(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.yaml")
	writeFile(t, f, `
name: test
count: 10
enabled: true
nested:
  key: yaml-val
  value: 3.14
`)

	cfg := &fileTestConfig{}
	src := &yamlSource{path: f}
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "test" || cfg.Count != 10 || !cfg.Enabled || cfg.Nested.Key != "yaml-val" || cfg.Nested.Value != 3.14 {
		t.Fatalf("got %+v", cfg)
	}
}

func TestJsonBasic(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.json")
	writeFile(t, f, `{"name":"json-test","count":20,"enabled":false,"nested":{"key":"json-val","value":2.71}}`)

	cfg := &fileTestConfig{}
	src := &jsonSource{path: f}
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "json-test" || cfg.Count != 20 || cfg.Enabled != false || cfg.Nested.Key != "json-val" || cfg.Nested.Value != 2.71 {
		t.Fatalf("got %+v", cfg)
	}
}

func TestFileNotFound(t *testing.T) {
	cfg := &fileTestConfig{}
	if err := (&yamlSource{path: "/nonexistent/file.yaml"}).Load(cfg); err == nil {
		t.Fatal("expected error for missing file")
	}
	if err := (&jsonSource{path: "/nonexistent/file.json"}).Load(cfg); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestYamlCustomTags(t *testing.T) {
	type customCfg struct {
		Address string `config:"endpoint"`
		Port    int    `config:"port"`
	}

	dir := t.TempDir()
	f := filepath.Join(dir, "custom.yaml")
	writeFile(t, f, `endpoint: http://localhost:8080
port: 9000`)

	cfg := &customCfg{}
	if err := (&yamlSource{path: f}).Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "http://localhost:8080" || cfg.Port != 9000 {
		t.Fatalf("got %+v", cfg)
	}
}

func TestJsonArraysAndMaps(t *testing.T) {
	type jsonComplex struct {
		Tags []string          `config:"tags"`
		Meta map[string]string `config:"meta"`
	}

	dir := t.TempDir()
	f := filepath.Join(dir, "complex.json")
	writeFile(t, f, `{"tags":["a","b","c"],"meta":{"key1":"val1","key2":"val2"}}`)

	cfg := &jsonComplex{}
	if err := (&jsonSource{path: f}).Load(cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Tags) != 3 || cfg.Tags[0] != "a" || cfg.Tags[2] != "c" {
		t.Fatalf("Tags = %v", cfg.Tags)
	}
	if cfg.Meta["key1"] != "val1" || cfg.Meta["key2"] != "val2" {
		t.Fatalf("Meta = %v", cfg.Meta)
	}
}

func TestWithReader(t *testing.T) {
	t.Run("yaml from reader", func(t *testing.T) {
		cfg := &struct {
			Val string `config:"val"`
		}{}
		if err := Load(cfg, WithReader(
			strings.NewReader("val: hello\n"),
			"yaml",
		)); err != nil {
			t.Fatal(err)
		}
		if cfg.Val != "hello" {
			t.Fatalf("got %q", cfg.Val)
		}
	})

	t.Run("json from reader", func(t *testing.T) {
		cfg := &struct {
			Val string `config:"val"`
		}{}
		if err := Load(cfg, WithReader(
			strings.NewReader(`{"val":"world"}`),
			"json",
		)); err != nil {
			t.Fatal(err)
		}
		if cfg.Val != "world" {
			t.Fatalf("got %q", cfg.Val)
		}
	})
}

func TestWithBytes(t *testing.T) {
	cfg := &struct {
		Val string `config:"val"`
	}{}
	if err := Load(cfg, WithBytes([]byte("val: from-bytes\n"), "yaml")); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "from-bytes" {
		t.Fatalf("got %q", cfg.Val)
	}
}

func TestWithDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.yaml"), `name: from-a`)
	writeFile(t, filepath.Join(dir, "b.json"), `{"name":"from-b"}`)

	type dirCfg struct {
		Name string `config:"name"`
	}

	// a.yaml loads first, b.json overrides
	cfg := &dirCfg{}
	if err := Load(cfg, WithDir(dir)); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "from-b" {
		t.Fatalf("got %q, want 'from-b'", cfg.Name)
	}
}
