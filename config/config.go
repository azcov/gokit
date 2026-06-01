// Package config provides a production-grade configuration library with
// support for environment variables (.env + OS), YAML, and JSON sources.
//
// The `config` struct tag is the source of truth for all mapping.
// Env var names are generated automatically from nested struct paths.
//
// # Generic API
//
//	cfg, err := config.New[Config](
//	    config.WithEnv(),
//	    config.WithYAML("config.yaml"),
//	)
//
// # Existing Struct API
//
//	var cfg Config
//	err := config.Load(&cfg,
//	    config.WithEnv(),
//	    config.WithYAML("config.yaml"),
//	)
//
// # Priority (later overrides earlier)
//
//	Default values → YAML/JSON → Environment variables
package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Source interface {
	Load(target any) error
}

type SourceFunc func(target any) error

func (f SourceFunc) Load(target any) error { return f(target) }

type loader struct {
	sources  []Source
	validate bool
	prefix   string
	hooks    []func(any) error
}

type Option func(*loader)

func Load(target any, opts ...Option) error {
	l := &loader{}
	for _, opt := range opts {
		opt(l)
	}
	for _, src := range l.sources {
		if err := src.Load(target); err != nil {
			return fmt.Errorf("config: %w", err)
		}
	}
	for _, hook := range l.hooks {
		if err := hook(target); err != nil {
			return err
		}
	}
	if l.validate {
		if err := validate(target); err != nil {
			return err
		}
	}
	return nil
}

func New[T any](opts ...Option) (*T, error) {
	var cfg T
	if err := Load(&cfg, opts...); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// --- Environment options ---

func WithEnv() Option {
	return func(l *loader) {
		l.sources = append(l.sources, &envSource{})
	}
}

func WithDotenv(paths ...string) Option {
	return func(l *loader) {
		l.sources = append(l.sources, &envSource{dotenv: paths})
	}
}

func WithPrefix(prefix string) Option {
	return func(l *loader) {
		l.prefix = prefix
		l.sources = append(l.sources, &prefixedEnvSource{prefix: prefix})
	}
}

// --- File source options ---

func WithYAML(path string) Option {
	return func(l *loader) {
		l.sources = append(l.sources, &yamlSource{path: path})
	}
}

func WithJSON(path string) Option {
	return func(l *loader) {
		l.sources = append(l.sources, &jsonSource{path: path})
	}
}

func WithReader(r io.Reader, format string) Option {
	return func(l *loader) {
		l.sources = append(l.sources, &readerSource{r: r, format: format})
	}
}

func WithBytes(data []byte, format string) Option {
	return func(l *loader) {
		l.sources = append(l.sources, &readerSource{
			r:      strings.NewReader(string(data)),
			format: format,
		})
	}
}

func WithDir(path string) Option {
	return func(l *loader) {
		l.sources = append(l.sources, &dirSource{path: path})
	}
}

// --- Validation & transform ---

func WithValidation() Option {
	return func(l *loader) {
		l.validate = true
	}
}

func WithHook(fn func(target any) error) Option {
	return func(l *loader) {
		l.hooks = append(l.hooks, fn)
	}
}

// --- Backward-compatible aliases ---

type Loader = Source

func FromEnv() Source {
	return &envSource{}
}

func FromDotenv(paths ...string) Source {
	return &envSource{dotenv: paths}
}

func FromYAML(path string) Source {
	return &yamlSource{path: path}
}

func FromJSON(path string) Source {
	return &jsonSource{path: path}
}

func Chain(sources ...Source) Source {
	return SourceFunc(func(target any) error {
		for _, s := range sources {
			if err := s.Load(target); err != nil {
				return err
			}
		}
		return nil
	})
}

// --- Internal source types ---

type prefixedEnvSource struct {
	prefix string
}

func (s *prefixedEnvSource) Load(target any) error {
	fields, err := collectFields(target)
	if err != nil {
		return err
	}
	for _, f := range fields {
		envName := s.prefix + "_" + envNameFromPath(f.path)
		val := os.Getenv(envName)
		if val == "" {
			val = os.Getenv(envNameFromPath(f.path))
		}
		if val == "" {
			continue
		}
		if err := setField(f.value, val); err != nil {
			return err
		}
	}
	return nil
}

type readerSource struct {
	r      io.Reader
	format string
}

func (s *readerSource) Load(target any) error {
	switch s.format {
	case "yaml", "yml":
		return decodeYAML(s.r, target)
	case "json":
		return decodeJSON(s.r, target)
	default:
		return fmt.Errorf("unsupported format: %s", s.format)
	}
}

type dirSource struct {
	path string
}

func (s *dirSource) Load(target any) error {
	entries, err := os.ReadDir(s.path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		fpath := filepath.Join(s.path, entry.Name())
		switch ext {
		case ".yaml", ".yml":
			if err := (&yamlSource{path: fpath}).Load(target); err != nil {
				return err
			}
		case ".json":
			if err := (&jsonSource{path: fpath}).Load(target); err != nil {
				return err
			}
		}
	}
	return nil
}
