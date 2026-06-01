package lock

import (
	"context"
	"time"
)

// Lock is a distributed mutual exclusion lock.
type Lock interface {
	// Acquire tries to obtain the lock. Returns true if acquired.
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	// Release releases the lock. Only succeeds if the caller still owns it.
	Release(ctx context.Context, key string) error
	// Extend refreshes the TTL of an already-held lock.
	Extend(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
