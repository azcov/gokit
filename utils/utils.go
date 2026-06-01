package utils

import (
	"math/rand"
	"time"
)

// Contains reports whether v is present in s.
func Contains[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// Map applies fn to each element of s and returns the results.
func Map[T, U any](s []T, fn func(T) U) []U {
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = fn(v)
	}
	return out
}

// Filter returns elements of s for which fn returns true.
func Filter[T any](s []T, fn func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if fn(v) {
			out = append(out, v)
		}
	}
	return out
}

// Unique returns a deduplicated copy of s preserving order.
func Unique[T comparable](s []T) []T {
	seen := make(map[T]struct{}, len(s))
	out := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

// Chunk splits s into slices of at most size n.
func Chunk[T any](s []T, n int) [][]T {
	if n <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(s)+n-1)/n)
	for len(s) > 0 {
		if len(s) < n {
			n = len(s)
		}
		chunks = append(chunks, s[:n])
		s = s[n:]
	}
	return chunks
}

// Keys returns the keys of m in undefined order.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns the values of m in undefined order.
func Values[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

// Retry calls fn up to attempts times, backing off by delay * 2^i between tries.
// It stops early if fn returns nil. A zero delay skips sleeping.
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := range attempts {
		if err = fn(); err == nil {
			return nil
		}
		if i < attempts-1 && delay > 0 {
			half := int64(delay) / 2
			var jitter time.Duration
			if half > 0 {
				jitter = time.Duration(rand.Int63n(half))
			}
			time.Sleep(delay + jitter)
			delay *= 2
		}
	}
	return err
}

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T { return &v }

// Deref dereferences p, returning zero if p is nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
