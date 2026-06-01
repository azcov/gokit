# Gokit — Go Toolkit Implementation Prompt

## Mission

Build **gokit** (`github.com/azcov/gokit`) into a production-grade Go toolkit providing clean abstraction interfaces and concrete provider implementations for common backend infrastructure patterns, with **≥80% test coverage across all packages**.

---

## Project Structure

```
gokit/
├── config.go            # Root Config struct aggregating all subsystem configs
├── config/              # Config library (env/YAML/JSON loading via reflection)
│   ├── config.go        # Public API: Source, Load, New[T], Option, With*
│   ├── walk.go          # Struct reflection walker
│   ├── set.go           # Field setters (string→value, any→value)
│   ├── env.go           # EnvSource (OS env + .env file parsing)
│   ├── yaml.go          # YamlSource (YAML file decode)
│   ├── json.go          # JsonSource (JSON file decode)
│   ├── types.go         # Type helpers (isStructType, resolveType, envNameFromPath)
│   ├── maputil.go       # Nested map operations (lookupNested, normalizeMap)
│   ├── validate.go      # Struct validation (required, min, max, len, oneof)
│   ├── errors.go        # Typed errors (ValidationError, FieldError, SourceError)
│   ├── source/          # Option constructors by category
│   │   ├── env.go       # WithEnv, WithDotenv, WithPrefix
│   │   ├── file.go      # WithYAML, WithJSON, WithReader, WithBytes, WithDir
│   │   └── vault.go     # [Stub] WithVault remote source
│   └── format/          # Format decoders
│       ├── yaml.go      # YAML decode → normalizeMap → setFromNestedMap
│       └── json.go      # JSON decode → setFromNestedMap
├── ai/                  # AI provider integration
│   ├── ai.go            # [TODO] Provider interface (Generate, Chat, Stream, Embed)
│   ├── antrophic/       # [TODO] Anthropic Claude
│   ├── openai/          # [TODO] OpenAI
│   ├── opencode/        # [TODO] OpenAI Codex/OpenCode
│   └── openrouter/      # [TODO] OpenRouter (multi-model gateway)
├── analytics/           # [TODO] Analytics event tracking
│   ├── analytics.go     # [TODO] Tracker interface (Track, Identify, Page)
│   ├── posthog/         # [TODO] PostHog
│   └── mixpanel/        # [TODO] Mixpanel
├── auth/                # Authentication
│   ├── auth.go          # Provider interface: Verify, CreateToken, RevokeToken
│   ├── basic/           # JWT HMAC-SHA256
│   ├── clerk/           # Clerk.com
│   └── supabase/        # Supabase JWT validation
├── cache/               # Caching
│   ├── cache.go         # Cache interface: Get, Set, Delete, Exists, Close
│   ├── memory/          # In-memory map with TTL
│   ├── redis/           # Redis-backed
│   └── ristretto/       # DGraph Ristretto (high-performance LRU)
├── cron/                # Scheduled jobs
│   ├── cron.go          # Scheduler interface: Add, Remove, Start, Stop
│   ├── robfig/          # robfig/cron v3
│   └── gocron/          # go-co-op/gocron v2
├── db/                  # SQL database access
│   ├── db.go            # DB interface: Query, Exec, BeginTx, Ping, Close
│   ├── mongo/           # MongoDB
│   ├── postgres/        # PostgreSQL (pgx v5)
│   └── mysql/           # MySQL (go-sql-driver)
├── docstore/            # Document/NoSQL store
│   ├── docstore.go      # Store interface: Get, Set, Update, Delete, List
│   ├── mongo/           # MongoDB
│   ├── dynamodb/        # AWS DynamoDB
│   └── firestore/       # GCP Firestore
├── email/               # Email sending
│   ├── email.go         # Mailer interface: Send, Close
│   ├── smtp/            # Raw SMTP
│   ├── resend/          # Resend.com
│   └── sendgrid/        # SendGrid
├── errorz/              # Structured errors
│   └── errorz.go        # Error struct with status, code, details
├── http/                # [TODO] HTTP utilities
│   ├── http.go          # [TODO] Client/Server helpers
│   ├── client/          # [TODO] HTTP client with retry, circuit breaker
│   └── server/          # [TODO] HTTP server utilities, CORS, rate limit
├── lock/                # Distributed locks
│   ├── lock.go          # Lock interface: Acquire, Release, Extend
│   ├── memory/          # In-memory (sync.Mutex)
│   └── redis/           # Redis-based distributed lock
├── logger/              # Logging
│   ├── logger.go        # Logger interface: Debug/Info/Warn/Error/Fatal, With, Fields
│   ├── slog/            # Go stdlib log/slog
│   └── zap/             # Uber Zap
├── middleware/           # [TODO] HTTP middleware
│   ├── middleware.go    # [TODO] Middleware types, chain
│   ├── cors/            # [TODO] CORS
│   ├── ratelimit/       # [TODO] Rate limiting
│   └── auth/            # [TODO] Auth middleware
├── monitoring/          # [TODO] Metrics & monitoring
│   ├── monitoring.go    # [TODO] Metrics interface (Counter, Gauge, Histogram)
│   ├── prometheus/      # [TODO] Prometheus
│   └── otel/            # [TODO] OpenTelemetry metrics
├── payment/             # Payment gateway integration (largest domain)
│   ├── payment.go       # Gateway interface: CreateCharge, GetCharge, Refund, VerifyWebhook
│   ├── stripe/          # Stripe
│   ├── paypal/          # PayPal
│   ├── midtrans/        # Midtrans (Indonesia)
│   ├── xendit/          # Xendit (SE Asia)
│   ├── razorpay/        # Razorpay (India)
│   └── doku/            # DOKU (Indonesia)
├── pubsub/              # Publish/Subscribe messaging
│   ├── pubsub.go        # Publisher, Subscriber, PubSub interfaces
│   ├── memory/          # In-memory goroutine-based
│   ├── kafka/           # Apache Kafka (Sarama)
│   └── nsq/             # NSQ
├── queue/               # Background job queue
│   ├── queue.go         # Enqueuer, Worker interfaces
│   └── asynq/           # Redis-backed (Asynq)
├── response/            # [TODO] HTTP response helpers
│   ├── response.go      # [TODO] JSON/XML/HTML response builders
│   └── validator/       # [TODO] Response validation helpers
├── storage/             # Object storage
│   ├── storage.go       # Storage interface: Put, Get, Delete, Exists, URL, List
│   ├── s3/              # AWS S3 (+ S3-compatible like R2)
│   ├── gcs/             # GCP Cloud Storage
│   └── r2/              # Cloudflare R2
├── tracing/             # Distributed tracing
│   ├── tracing.go       # [TODO] Tracer interface: StartSpan, Inject, Extract
│   ├── opentelemetry/   # [RENAME FROM opentelemery] OpenTelemetry
│   └── datadog/         # [TODO] Datadog APM
├── tracking/            # Error/performance tracking
│   ├── tracking.go      # [TODO] Tracker interface: CaptureError, CaptureMessage
│   ├── sentry/          # [TODO] Sentry
│   └── rollbar/         # [TODO] Rollbar
└── utils/               # Shared utilities
    ├── utils.go         # [TODO] Shared helpers
    └── page/            # [TODO] Pagination utilities
```

---

## Architecture Pattern

Every domain follows the same pattern:

### 1. Root Interface (`domain/domain.go`)

```go
package domain

type Interface interface {
    Method(ctx context.Context, req Request) (Response, error)
    Close() error
}

// Types, options, errors shared across providers.
type Config struct {
    Timeout time.Duration `config:"timeout" validate:"required"`
}
```

### 2. Provider Sub-package (`domain/provider/provider.go`)

```go
package provider

import "github.com/azcov/gokit/domain"

var _ domain.Interface = (*Provider)(nil)

type Provider struct {
    client *sdk.Client
}

func New(cfg domain.Config) (*Provider, error) { ... }
func (p *Provider) Method(ctx context.Context, req domain.Request) (domain.Response, error) { ... }
func (p *Provider) Close() error { ... }
```

### 3. Config Integration

All provider config structs use `config` tags for env/YAML/JSON mapping:

```go
type Config struct {
    URL     string `config:"url"`
    Timeout int    `config:"timeout"`
}
```

Env var names auto-derived: `PAYMENT_STRIPE_URL`, `PAYMENT_STRIPE_TIMEOUT`.

### 4. Zero Dependencies on Providers in Root Interface

Root `domain.go` files import ZERO provider SDKs. Sub-packages import SDKs.

---

## Implementation Priority by Category

### Tier 1 — Core Infrastructure (test + stabilize)
| Package | Action |
|---------|--------|
| `config/` | Full test suite (≥80% coverage), bench tests, source/format sub-packages |
| `errorz/` | Unit tests, JSON marshal/unmarshal, error wrapping with `errors.As`/`Is` |
| `logger/` | Tests for slog + zap implementations, interface compliance tests |

### Tier 2 — Storage & Messaging (test + harden)
| Package | Action |
|---------|--------|
| `cache/` | Tests for all 3 providers, TTL edge cases, concurrent access |
| `lock/` | Tests for memory + redis, timeout/expiry edge cases |
| `pubsub/` | Tests for memory + kafka + nsq, message ordering, delivery guarantees |
| `queue/` | Tests for asynq, job retry, schedule, unique options |
| `storage/` | Tests for s3 + gcs + r2, upload/download, streaming |

### Tier 3 — Business Domains (test + harden)
| Package | Action |
|---------|--------|
| `auth/` | Tests for all 3 providers, token lifecycle, webhook verification |
| `email/` | Tests for smtp + resend + sendgrid, template rendering |
| `payment/` | Tests for all 6 providers, webhook signature verification, idempotency |

### Tier 4 — Complete Empty Packages
| Package | Action |
|---------|--------|
| `ai/` | Define interface, implement OpenAI + Anthropic providers |
| `analytics/` | Define interface, implement PostHog provider |
| `http/` | Client with retry/circuit-breaker, server utilities |
| `middleware/` | CORS, rate-limit, auth, request-ID, recovery middleware |
| `monitoring/` | Prometheus + OpenTelemetry metrics collectors |
| `response/` | JSON/XML/HTML response builders, pagination helpers |
| `tracing/` | **Fix typo**: `opentelemery/` → `opentelemetry/`, implement OpenTelemetry tracing |
| `tracking/` | Define interface, implement Sentry + Rollbar providers |
| `utils/` | Pagination, slice/map helpers, retry utilities |

---

## Config Library (config/) — Detailed Spec

### Core API

```go
type Source interface { Load(target any) error }
type Option func(*loader)
func Load(target any, opts ...Option) error
func New[T any](opts ...Option) (*T, error)
```

### Options by Category

```go
// Environment
func WithEnv() Option
func WithDotenv(paths ...string) Option
func WithPrefix(prefix string) Option

// File sources
func WithYAML(path string) Option
func WithJSON(path string) Option
func WithReader(r io.Reader, format string) Option
func WithBytes(data []byte, format string) Option
func WithDir(path string) Option

// Validation & transform
func WithValidation() Option
func WithHook(fn func(target any) error) Option
```

### Backward-Compatible Aliases (MUST keep)

```go
type Loader = Source
func FromEnv() Source
func FromDotenv(paths ...string) Source
func FromYAML(path string) Source
func FromJSON(path string) Source
func Chain(sources ...Source) Source
```

### Struct Tag Rules

```go
type S3Config struct {
    Address string `config:"endpoint" validate:"required"`
    Bucket  string `config:"bucket"`
    Region  string `config:"region"`   // env: S3_REGION
}
```

- No config tag → `strings.ToLower(FieldName)`
- `config:"-"` → skip
- Anonymous structs → inline fields
- Pointer structs → auto-allocate
- Env names: `Storage.S3.Address` → `STORAGE_S3_ENDPOINT`

### Validation Tags

| Tag | Behavior |
|-----|----------|
| `validate:"required"` | Must be non-zero |
| `validate:"min=1"` | Numeric ≥ min; String len ≥ min |
| `validate:"max=100"` | Numeric ≤ max; String len ≤ max |
| `validate:"len=10"` | String/Slice len == N |
| `validate:"oneof=a b c"` | String must match |

### Supported Types

string, bool, int8–int64, uint8–uint64, float32/64, time.Duration, time.Time, slices, maps, nested structs, pointers, embedded structs, TextUnmarshaler, json.Unmarshaler.

### Test Requirements (config/)

| Test File | What it covers |
|-----------|---------------|
| `walk_test.go` | Field collection, nested structs, embedded, pointers, skip, unexported, env name generation |
| `set_test.go` | Every type for setField (env) + setFieldFromAny (JSON/YAML), error cases |
| `env_test.go` | .env parsing (quotes, export, comments, empty), OS env roundtrip, multi-file |
| `file_test.go` | YAML + JSON decode, nested, arrays, maps, custom tags, file-not-found |
| `validate_test.go` | required, min, max, len, oneof, nested validation, multiple errors |
| `maputil_test.go` | lookupNested, normalizeMap, flattenMap, edge cases |
| `config_test.go` | Integration: Load with all sources, priority, New[T], Chain, error propagation |

---

## Cross-Cutting Concerns

### Test Requirements (every package)

1. **Interface compliance**: `var _ Interface = (*Provider)(nil)` compile-time check
2. **Config roundtrip**: Provider config structs tested with `config.Load` (env + yaml + json)
3. **Table-driven tests**: All test files use `t.Run` with table-driven cases
4. **Coverage target**: ≥80% per package
5. **Benchmark tests**: For `config/`, `cache/`, `lock/`, `pubsub/`, `queue/`

### Directory Fixes Needed

| Current | Fix | Reason |
|---------|-----|--------|
| `tracing/opentelemery/` | `tracing/opentelemetry/` | Misspelled "opentelemetry" |

### Config Tag Consistency

All provider config structs must use `config` tags (NOT separate `env/json/yaml` tags):

```go
// BEFORE (inconsistent — uses env/json/yaml tags)
type Config struct {
    URL string `env:"URL" json:"url" yaml:"url"`
}

// AFTER (single source of truth)
type Config struct {
    URL string `config:"url"`
}
```

### Backward Compatibility

- Root `config.go` LoadConfig must continue working unchanged
- `config.Chain(config.FromEnv(), config.FromYAML("app.yaml"))` must compile

---

## Go Module Dependencies

| Package | Key Dependencies |
|---------|-----------------|
| `config/` | `gopkg.in/yaml.v3` (stdlib for rest) |
| `auth/` | `golang-jwt/jwt/v5`, `clerk/clerk-sdk-go/v2` |
| `cache/` | `redis/go-redis/v9`, `dgraph-io/ristretto` |
| `cron/` | `robfig/cron/v3`, `go-co-op/gocron/v2` |
| `db/` | `jackc/pgx/v5`, `go-sql-driver/mysql`, `mongo-driver` |
| `docstore/` | `mongo-driver`, `aws-sdk-go-v2/dynamodb`, `cloud.google.com/go/firestore` |
| `email/` | `resend-go/v2`, `sendgrid-go` |
| `logger/` | `go.uber.org/zap` |
| `payment/` | `stripe-go/v81`, `plutov/paypal/v4`, `midtrans-go`, `xendit-go`, `razorpay-go` |
| `pubsub/` | `IBM/sarama`, `go-nsq` |
| `queue/` | `hibiken/asynq` |
| `storage/` | `aws-sdk-go-v2/s3`, `cloud.google.com/go/storage` |

No new dependencies. Use existing versions in `go.mod`.

---

## Implementation Order

### Phase 1: Foundation & Quality (Week 1)
1. Fix `tracing/opentelemery/` → `tracing/opentelemetry/`
2. Write full test suite for `config/` (8 test files, ≥80% coverage)
3. Create `config/source/` and `config/format/` sub-packages
4. Normalize all provider config structs to single `config` tag
5. Write tests for `errorz/` and `logger/`

### Phase 2: Storage & Messaging (Week 2)
6. Write tests for `cache/`, `lock/`, `pubsub/`, `queue/`, `storage/`
7. Add benchmark tests for storage/messaging packages

### Phase 3: Business Domains (Week 3)
8. Write tests for `auth/`, `email/`, `payment/`
9. Stress-test webhook verification, idempotency, error handling

### Phase 4: Complete Empty Packages (Week 4)
10. Define interfaces + implement providers for:
    - `ai/` (OpenAI, Anthropic)
    - `analytics/` (PostHog)
    - `http/` (client + server)
    - `middleware/` (CORS, rate-limit, auth, recovery)
    - `monitoring/` (Prometheus, OpenTelemetry)
    - `response/` (JSON/XML response builders)
    - `tracing/` (OpenTelemetry)
    - `tracking/` (Sentry, Rollbar)
    - `utils/` (pagination, retry, slice/map helpers)

### Phase 5: Polish (Week 5)
11. `go vet ./...` passes across entire module
12. `go test -cover ./...` shows ≥80% per package
13. Benchmarks documented and regression-tested

---

## Deliverable Checklist

- [ ] Root `config.go` LoadConfig unchanged and working
- [ ] `config/` — 8+ test files, ≥80% coverage, benchmarks
- [ ] `config/source/` — env.go, file.go, vault.go (stub)
- [ ] `config/format/` — yaml.go, json.go
- [ ] `ai/` — interface + OpenAI + Anthropic providers
- [ ] `analytics/` — interface + PostHog provider
- [ ] `auth/` — tests for all 3 providers
- [ ] `cache/` — tests for all 3 providers + benchmarks
- [ ] `cron/` — tests for both providers
- [ ] `db/` — tests for all 3 providers
- [ ] `docstore/` — tests for all 3 providers
- [ ] `email/` — tests for all 3 providers
- [ ] `errorz/` — unit tests
- [ ] `http/` — client + server utilities
- [ ] `lock/` — tests for both providers + benchmarks
- [ ] `logger/` — tests for both implementations
- [ ] `middleware/` — CORS, rate-limit, auth, recovery
- [ ] `monitoring/` — Prometheus + OpenTelemetry
- [ ] `payment/` — tests for all 6 providers
- [ ] `pubsub/` — tests for all 3 providers + benchmarks
- [ ] `queue/` — tests for asynq + benchmarks
- [ ] `response/` — JSON/XML response builders
- [ ] `storage/` — tests for all 3 providers
- [ ] `tracing/` — fix typo, implement OpenTelemetry
- [ ] `tracking/` — interface + Sentry + Rollbar
- [ ] `utils/` — pagination + helpers
- [ ] All provider config structs use `config` tags (no `env/json/yaml`)
- [ ] `go vet ./...` passes
- [ ] All tests pass
- [ ] ≥80% test coverage per package
- [ ] Benchmarks documented
