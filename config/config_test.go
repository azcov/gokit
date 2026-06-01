package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testEnvCfg struct {
	Name string `config:"name"`
	Port int    `config:"port"`
}

func TestLoadWithEnv(t *testing.T) {
	cleanup := setEnvCleaner(t, "NAME", "PORT")
	defer cleanup()
	os.Setenv("NAME", "myapp")
	os.Setenv("PORT", "8080")

	cfg := &testEnvCfg{}
	if err := Load(cfg, WithEnv()); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "myapp" || cfg.Port != 8080 {
		t.Fatalf("got %+v", cfg)
	}
}

type yamlEnvCfg struct {
	Name  string `config:"name"`
	Count int    `config:"count"`
}

func TestLoadWithYamlAndEnv(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	writeFile(t, f, "name: from-yaml\ncount: 10\n")

	cleanup := setEnvCleaner(t, "COUNT")
	defer cleanup()
	os.Setenv("COUNT", "99")

	cfg := &yamlEnvCfg{Count: -1}
	if err := Load(cfg, WithYAML(f), WithEnv()); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "from-yaml" {
		t.Fatalf("Name = %q, want 'from-yaml'", cfg.Name)
	}
	if cfg.Count != 99 {
		t.Fatalf("Count = %d, want 99 (env overrides yaml)", cfg.Count)
	}
}

type newCfg struct {
	Name string `config:"name"`
}

func TestNewGeneric(t *testing.T) {
	cleanup := setEnvCleaner(t, "NAME")
	defer cleanup()
	os.Setenv("NAME", "generic-test")

	cfg, err := New[newCfg](WithEnv())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "generic-test" {
		t.Fatalf("Name = %q, want 'generic-test'", cfg.Name)
	}
}

type chainCfg struct {
	Name string `config:"name"`
}

func TestChainBackwardCompat(t *testing.T) {
	cleanup := setEnvCleaner(t, "NAME")
	defer cleanup()
	os.Setenv("NAME", "chain-test")

	cfg := &chainCfg{}
	loader := Chain(FromEnv())
	if err := loader.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "chain-test" {
		t.Fatalf("Name = %q", cfg.Name)
	}
}

func TestLoaderAlias(t *testing.T) {
	var l Loader = FromEnv()
	if l == nil {
		t.Fatal("Loader alias should not be nil")
	}
}

type errCfg struct {
	Val string `config:"val"`
}

func TestErrorPropagation(t *testing.T) {
	cfg := &errCfg{}
	err := Load(cfg, WithYAML("/nonexistent/file.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

type emptyCfg struct {
	Val string `config:"val"`
}

func TestEmptyOptions(t *testing.T) {
	cfg := &emptyCfg{Val: "default"}
	if err := Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "default" {
		t.Fatalf("Val = %q, want 'default'", cfg.Val)
	}
}

type reqCfg struct {
	Name string `config:"name" validate:"required"`
}

func TestWithValidation(t *testing.T) {
	t.Run("missing required - error", func(t *testing.T) {
		cfg := &reqCfg{}
		err := Load(cfg, WithValidation())
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("present required - ok", func(t *testing.T) {
		cfg := &reqCfg{Name: "hello"}
		err := Load(cfg, WithValidation())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

type hookCfg struct {
	Val string `config:"val"`
}

func TestWithHook(t *testing.T) {
	cfg := &hookCfg{}
	called := false
	if err := Load(cfg, WithHook(func(target any) error {
		called = true
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("hook was not called")
	}
}

type prefixCfg struct {
	Name string `config:"name"`
}

func TestWithPrefix(t *testing.T) {
	cleanup := setEnvCleaner(t, "MYAPP_NAME")
	defer cleanup()
	os.Setenv("MYAPP_NAME", "prefixed")

	cfg := &prefixCfg{}
	if err := Load(cfg, WithPrefix("MYAPP")); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "prefixed" {
		t.Fatalf("Name = %q, want 'prefixed'", cfg.Name)
	}
}

type sourceFuncCfg struct {
	Val string `config:"val"`
}

func TestSourceFunc(t *testing.T) {
	cfg := &sourceFuncCfg{}
	fn := SourceFunc(func(target any) error {
		targetVal := target.(*sourceFuncCfg)
		targetVal.Val = "from-sourcefunc"
		return nil
	})
	if err := fn.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "from-sourcefunc" {
		t.Fatalf("Val = %q", cfg.Val)
	}
}

func TestFromDotenvBackwardCompat(t *testing.T) {
	cleanup := setEnvCleaner(t, "VAL")
	defer cleanup()
	os.Setenv("VAL", "dotenv-test")

	cfg := &emptyCfg{}
	src := FromDotenv()
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "dotenv-test" {
		t.Fatalf("Val = %q, want 'dotenv-test'", cfg.Val)
	}
}

func TestFromYAMLBackwardCompat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.yaml")
	writeFile(t, f, "val: from-yaml\n")

	cfg := &emptyCfg{}
	src := FromYAML(f)
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "from-yaml" {
		t.Fatalf("Val = %q", cfg.Val)
	}
}

func TestFromJSONBackwardCompat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.json")
	writeFile(t, f, `{"val":"from-json"}`)

	cfg := &emptyCfg{}
	src := FromJSON(f)
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "from-json" {
		t.Fatalf("Val = %q", cfg.Val)
	}
}

func TestWithDotenv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	writeFile(t, envFile, "VAL=from-dotenv\n")

	cfg := &emptyCfg{}
	if err := Load(cfg, WithDotenv(envFile), WithEnv()); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "from-dotenv" {
		t.Fatalf("Val = %q, want 'from-dotenv'", cfg.Val)
	}
}

func TestSourceError(t *testing.T) {
	err := &SourceError{Err: io.EOF, Source: "test"}
	if err.Error() != "source test: EOF" {
		t.Fatalf("got %q", err.Error())
	}
	if err.Unwrap() != io.EOF {
		t.Fatal("Unwrap should return the wrapped error")
	}
}

func TestReaderSourceUnsupportedFormat(t *testing.T) {
	cfg := &emptyCfg{}
	err := Load(cfg, WithReader(strings.NewReader("unexpected value"), "txt"))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestLoadWithAllSources(t *testing.T) {
	dir := t.TempDir()
	yamlFile := filepath.Join(dir, "all.yaml")
	jsonFile := filepath.Join(dir, "all.json")
	writeFile(t, yamlFile, "name: yaml-only\n")
	writeFile(t, jsonFile, `{"count": 50}`)

	cleanup := setEnvCleaner(t, "ENABLED")
	defer cleanup()
	os.Setenv("ENABLED", "true")

	type allCfg struct {
		Name    string `config:"name"`
		Count   int    `config:"count"`
		Enabled bool   `config:"enabled"`
	}

	cfg := &allCfg{}
	if err := Load(cfg, WithYAML(yamlFile), WithJSON(jsonFile), WithEnv()); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "yaml-only" {
		t.Fatalf("Name = %q", cfg.Name)
	}
	if cfg.Count != 50 {
		t.Fatalf("Count = %d", cfg.Count)
	}
	if cfg.Enabled != true {
		t.Fatalf("Enabled = %v", cfg.Enabled)
	}
}
