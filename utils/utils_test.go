package utils_test

import (
	"fmt"
	"testing"

	"github.com/azcov/gokit/utils"
)

func TestContains(t *testing.T) {
	s := []int{1, 2, 3, 4}
	if !utils.Contains(s, 3) {
		t.Error("expected Contains(3)=true")
	}
	if utils.Contains(s, 99) {
		t.Error("expected Contains(99)=false")
	}
}

func TestMap(t *testing.T) {
	result := utils.Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
	want := []int{2, 4, 6}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("index %d: expected %d, got %d", i, want[i], v)
		}
	}
}

func TestFilter(t *testing.T) {
	evens := utils.Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
	if len(evens) != 2 {
		t.Errorf("expected 2 even numbers, got %d", len(evens))
	}
}

func TestUnique(t *testing.T) {
	result := utils.Unique([]string{"a", "b", "a", "c", "b"})
	if len(result) != 3 {
		t.Errorf("expected 3 unique, got %d", len(result))
	}
}

func TestChunk(t *testing.T) {
	chunks := utils.Chunk([]int{1, 2, 3, 4, 5}, 2)
	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != 2 || len(chunks[2]) != 1 {
		t.Errorf("unexpected chunk sizes: %v", chunks)
	}
}

func TestChunk_InvalidSize(t *testing.T) {
	if utils.Chunk([]int{1, 2, 3}, 0) != nil {
		t.Error("expected nil for invalid chunk size")
	}
}

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := utils.Keys(m)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	vals := utils.Values(m)
	if len(vals) != 2 {
		t.Errorf("expected 2 values, got %d", len(vals))
	}
}

func TestRetry_Success(t *testing.T) {
	calls := 0
	err := utils.Retry(3, 0, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_Exhausted(t *testing.T) {
	calls := 0
	err := utils.Retry(3, 0, func() error {
		calls++
		return fmt.Errorf("fail")
	})
	if err == nil {
		t.Error("expected error after exhausted retries")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestPtr(t *testing.T) {
	v := 42
	p := utils.Ptr(v)
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}
}

func TestDeref(t *testing.T) {
	v := 99
	if utils.Deref(&v) != 99 {
		t.Errorf("expected 99")
	}
	var nilPtr *int
	if utils.Deref(nilPtr) != 0 {
		t.Errorf("expected 0 for nil ptr")
	}
}
