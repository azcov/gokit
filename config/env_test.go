package config

import (
	"os"
	"testing"
)

func TestParseDotenv(t *testing.T) {
	cleanup := setEnvCleaner(t, "TEST_KEY", "TEST_EXPORT", "TEST_QUOTED", "TEST_SINGLE", "TEST_QUOTE_INNER")
	defer cleanup()

	t.Run("basic key=value", func(t *testing.T) {
		parseDotenv("TEST_KEY=hello")
		if os.Getenv("TEST_KEY") != "hello" {
			t.Fatalf("got %q", os.Getenv("TEST_KEY"))
		}
	})

	t.Run("export prefix", func(t *testing.T) {
		parseDotenv("export TEST_EXPORT=value")
		if os.Getenv("TEST_EXPORT") != "value" {
			t.Fatalf("got %q", os.Getenv("TEST_EXPORT"))
		}
	})

	t.Run("double quoted value", func(t *testing.T) {
		parseDotenv(`TEST_QUOTED="hello world"`)
		if os.Getenv("TEST_QUOTED") != "hello world" {
			t.Fatalf("got %q", os.Getenv("TEST_QUOTED"))
		}
	})

	t.Run("single quoted value", func(t *testing.T) {
		parseDotenv(`TEST_SINGLE='single quotes'`)
		if os.Getenv("TEST_SINGLE") != "single quotes" {
			t.Fatalf("got %q", os.Getenv("TEST_SINGLE"))
		}
	})

	t.Run("comments and empty lines skipped", func(t *testing.T) {
		parseDotenv("# this is a comment\n\n\nTEST_QUOTE_INNER=val")
		if os.Getenv("TEST_QUOTE_INNER") != "val" {
			t.Fatalf("got %q", os.Getenv("TEST_QUOTE_INNER"))
		}
	})
}

type envLoadCfg struct {
	String string `config:"string"`
	Int    int    `config:"int"`
	Bool   bool   `config:"bool"`
}

func TestEnvSourceLoad(t *testing.T) {
	cleanup := setEnvCleaner(t, "STRING", "INT", "BOOL")
	defer cleanup()

	os.Setenv("STRING", "hello")
	os.Setenv("INT", "42")
	os.Setenv("BOOL", "true")

	cfg := &envLoadCfg{}
	src := &envSource{}
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.String != "hello" {
		t.Fatalf("String = %q, want %q", cfg.String, "hello")
	}
	if cfg.Int != 42 {
		t.Fatalf("Int = %d, want %d", cfg.Int, 42)
	}
	if cfg.Bool != true {
		t.Fatalf("Bool = %v, want %v", cfg.Bool, true)
	}
}

type envNestedOuter struct {
	Name  string      `config:"name"`
	Inner envNestedIn `config:"inner"`
}

type envNestedIn struct {
	Key string `config:"key"`
}

func TestEnvSourceNested(t *testing.T) {
	cleanup := setEnvCleaner(t, "NAME", "INNER_KEY")
	defer cleanup()

	os.Setenv("NAME", "parent")
	os.Setenv("INNER_KEY", "child")

	cfg := &envNestedOuter{}
	src := &envSource{}
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "parent" {
		t.Fatalf("Name = %q", cfg.Name)
	}
	if cfg.Inner.Key != "child" {
		t.Fatalf("Inner.Key = %q", cfg.Inner.Key)
	}
}

type envEmptyCfg struct {
	Val string `config:"val"`
}

func TestEnvSourceEmptyEnvVar(t *testing.T) {
	cleanup := setEnvCleaner(t, "VAL")
	defer cleanup()

	os.Unsetenv("VAL")
	cfg := &envEmptyCfg{Val: "default"}

	src := &envSource{}
	if err := src.Load(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Val != "default" {
		t.Fatalf("Val = %q, want 'default'", cfg.Val)
	}
}

func setEnvCleaner(t *testing.T, keys ...string) func() {
	for _, k := range keys {
		os.Unsetenv(k)
	}
	return func() {
		for _, k := range keys {
			os.Unsetenv(k)
		}
	}
}
