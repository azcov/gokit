package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/azcov/gokit/lock"
	"github.com/azcov/gokit/lock/memory"
)

func TestInterfaceCompliance(t *testing.T) {
	var _ lock.Lock = memory.New()
}

func TestAcquire_Release(t *testing.T) {
	m := memory.New()
	ctx := context.Background()

	ok, err := m.Acquire(ctx, "resource", time.Minute)
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}

	// second acquire on the same key must fail while held
	if ok, _ := m.Acquire(ctx, "resource", time.Minute); ok {
		t.Error("expected second Acquire to fail while lock held")
	}

	if err := m.Release(ctx, "resource"); err != nil {
		t.Fatalf("Release: %v", err)
	}

	// after release, acquire succeeds again
	if ok, _ := m.Acquire(ctx, "resource", time.Minute); !ok {
		t.Error("expected Acquire to succeed after Release")
	}
}

func TestAcquire_DifferentKeys(t *testing.T) {
	m := memory.New()
	ctx := context.Background()
	ok1, _ := m.Acquire(ctx, "a", time.Minute)
	ok2, _ := m.Acquire(ctx, "b", time.Minute)
	if !ok1 || !ok2 {
		t.Error("locks on different keys should not conflict")
	}
}

func TestAcquire_ExpiredLock(t *testing.T) {
	m := memory.New()
	ctx := context.Background()

	_, _ = m.Acquire(ctx, "k", 20*time.Millisecond)
	time.Sleep(40 * time.Millisecond)
	// expired — a fresh acquire should succeed
	if ok, _ := m.Acquire(ctx, "k", time.Minute); !ok {
		t.Error("expected Acquire to succeed after lock expiry")
	}
}

func TestExtend(t *testing.T) {
	m := memory.New()
	ctx := context.Background()

	_, _ = m.Acquire(ctx, "k", 30*time.Millisecond)
	ok, err := m.Extend(ctx, "k", time.Minute)
	if err != nil || !ok {
		t.Fatalf("Extend: ok=%v err=%v", ok, err)
	}
	// after extending well past the original TTL, the lock is still held
	time.Sleep(40 * time.Millisecond)
	if ok, _ := m.Acquire(ctx, "k", time.Minute); ok {
		t.Error("lock should still be held after Extend")
	}
}

func TestExtend_NotHeld(t *testing.T) {
	m := memory.New()
	if ok, _ := m.Extend(context.Background(), "never-acquired", time.Minute); ok {
		t.Error("Extend on an unheld lock should return false")
	}
}

func TestRelease_NotHeld(t *testing.T) {
	m := memory.New()
	if err := m.Release(context.Background(), "never-acquired"); err != nil {
		t.Errorf("Release on unheld lock should be a no-op, got %v", err)
	}
}
