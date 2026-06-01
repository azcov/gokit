package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/azcov/gokit/storage"
)

var _ storage.Storage = (*Local)(nil)

// Local implements storage.Storage using the local filesystem.
// Useful for development, testing, and single-node deployments.
type Local struct {
	root string
}

func New(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("local storage: create root %s: %w", root, err)
	}
	return &Local{root: root}, nil
}

func (l *Local) Put(_ context.Context, key string, r io.Reader, _ storage.PutOptions) error {
	path := l.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("local storage: create %s: %w", key, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("local storage: write %s: %w", key, err)
	}
	return nil
}

func (l *Local) Get(_ context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(l.path(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local storage: %s: not found", key)
		}
		return nil, fmt.Errorf("local storage: open %s: %w", key, err)
	}
	return f, nil
}

func (l *Local) Delete(_ context.Context, key string) error {
	err := os.Remove(l.path(key))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local storage: delete %s: %w", key, err)
	}
	return nil
}

func (l *Local) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(l.path(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("local storage: stat %s: %w", key, err)
}

func (l *Local) URL(_ context.Context, key string) (string, error) {
	return "file://" + l.path(key), nil
}

func (l *Local) List(_ context.Context, prefix string) ([]storage.Object, error) {
	var objects []storage.Object
	err := filepath.WalkDir(l.root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(l.root, path)
		key := filepath.ToSlash(rel)
		if !strings.HasPrefix(key, prefix) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		objects = append(objects, storage.Object{Key: key, Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("local storage: list: %w", err)
	}
	return objects, nil
}

func (l *Local) Close() error { return nil }

func (l *Local) path(key string) string {
	return filepath.Join(l.root, filepath.FromSlash(key))
}
