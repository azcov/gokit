# Contributing to gokit

Thanks for helping build gokit. This guide covers the conventions that keep the
toolkit consistent across 100+ packages.

## Ground rules

1. **Interface first.** Every category defines a minimal interface in
   `domain/domain.go` (e.g. `cache/cache.go`). The root file imports **zero**
   third-party SDKs — only stdlib and other gokit domain packages.
2. **One provider per sub-package.** Concrete implementations live in
   `domain/provider/` (e.g. `cache/redis/`). Only sub-packages import SDKs.
3. **Compile-time interface check.** Every provider asserts conformance:
   ```go
   var _ cache.Cache = (*Redis)(nil)
   ```
4. **Single `config` tag.** Provider `Config` structs use only the `config` tag
   (never `env`/`json`/`yaml`). Env var names derive from the field path.
5. **Context first.** Every method that does I/O takes `context.Context` as its
   first argument and honors cancellation.
6. **Wrap errors with package context.** `fmt.Errorf("redis: get %s: %w", key, err)`.
7. **No new dependencies** without discussion — prefer `net/http` for REST APIs.

## Adding a new provider

1. Create `domain/provider/provider.go` from the template below.
2. Add a `Config` block to the root `config.go` (reuse the domain `Config` if the
   shape matches; otherwise an inline struct).
3. Add a row to `.planning/plan.md`.
4. Write a table-driven test with `httptest` for HTTP providers — target ≥40%
   coverage, ≥80% for core packages.
5. Run the checks below; open a PR.

## Provider template

```go
// Package foo implements the bar.Interface backed by the Foo service.
package foo

import (
	"context"
	"fmt"

	"github.com/azcov/gokit/bar"
)

var _ bar.Interface = (*Foo)(nil)

type Config struct {
	APIKey  string `config:"api_key"`
	BaseURL string `config:"base_url"`
}

type Foo struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) *Foo {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.foo.com"
	}
	return &Foo{cfg: cfg, client: &http.Client{Timeout: 30 * time.Second}}
}

func (f *Foo) Do(ctx context.Context, req bar.Request) (*bar.Response, error) {
	// ... use net/http, wrap errors with "foo: " prefix ...
	return nil, fmt.Errorf("foo: not implemented")
}

func (f *Foo) Close() error { return nil }
```

## Local checks (must pass before PR)

```bash
gofmt -l .            # must print nothing
go vet ./...
go build ./...
go test -race ./...
go mod tidy           # must leave go.mod/go.sum unchanged
```

CI runs all of these plus `staticcheck` on every push and PR.

## Commit style

Conventional commits, one logical change per commit:

```
feat(auth): add Apple OAuth provider
fix(config): normalize provider config tags
docs: add package doc.go to every category
test(payment): cover stripe webhook verification
```

End commit messages with the `Co-Authored-By` trailer if pairing with an agent.
