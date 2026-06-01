package memory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/azcov/gokit/cache"
	"github.com/azcov/gokit/cache/memory"
)

func TestInterfaceCompliance(t *testing.T) {
	var _ cache.Cache = memory.New()
}

func TestSetGet(t *testing.T) {
	m := memory.New()
	ctx := context.Background()

	if err := m.Set(ctx, "k", []byte("v"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := m.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "v" {
		t.Errorf("got %q, want v", got)
	}
}

func TestGet_Missing(t *testing.T) {
	m := memory.New()
	got, err := m.Get(context.Background(), "nope")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing key, got %q", got)
	}
}

func TestTTL_Expiry(t *testing.T) {
	m := memory.New()
	ctx := context.Background()

	_ = m.Set(ctx, "k", []byte("v"), 20*time.Millisecond)
	if v, _ := m.Get(ctx, "k"); string(v) != "v" {
		t.Fatal("value should be present before expiry")
	}
	time.Sleep(40 * time.Millisecond)
	if v, _ := m.Get(ctx, "k"); v != nil {
		t.Errorf("value should have expired, got %q", v)
	}
}

func TestExists(t *testing.T) {
	m := memory.New()
	ctx := context.Background()

	if ok, _ := m.Exists(ctx, "k"); ok {
		t.Error("should not exist yet")
	}
	_ = m.Set(ctx, "k", []byte("v"), 0)
	if ok, _ := m.Exists(ctx, "k"); !ok {
		t.Error("should exist after set")
	}
}

func TestDelete(t *testing.T) {
	m := memory.New()
	ctx := context.Background()
	_ = m.Set(ctx, "k", []byte("v"), 0)
	if err := m.Delete(ctx, "k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if v, _ := m.Get(ctx, "k"); v != nil {
		t.Error("value should be gone after delete")
	}
}

func TestClose(t *testing.T) {
	if err := memory.New().Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := memory.New()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "k"
			_ = m.Set(ctx, key, []byte("v"), time.Minute)
			_, _ = m.Get(ctx, key)
			_, _ = m.Exists(ctx, key)
		}(i)
	}
	wg.Wait()
}
