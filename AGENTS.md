# AGENTS.md

This file provides guidance to coding agents (Claude Code, etc.) when working with code in this repository.

## What this is

`github.com/azcov/gokit` — an interface-first Go toolkit (Go 1.26, `GOTOOLCHAIN=auto`). Each backend concern (cache, db, auth, payment, …) is a *category* that defines one minimal Go interface plus several swappable provider implementations. There is no application/binary here; it's a library consumed via `go get`.

## Commands

```bash
go build ./...                       # build everything
go vet ./...
go test -race ./...                  # full test suite (CI runs with -coverprofile)
go test -race ./cache/...            # one category
go test -race -run TestName ./cache/memory   # one test
gofmt -l .                           # must print nothing (CI fails otherwise)
go mod tidy                          # must leave go.mod/go.sum unchanged (CI checks)
```

CI (`.github/workflows`) runs `go mod tidy` diff check, `go vet`, `go build`, `go test -race`, `gofmt -l`, and `staticcheck` on push/PR to `main`. Run all of these before considering a change done.

## Architecture

Every category is the same shape — learn one, you know all of them:

```
domain/domain.go        ← the interface + shared types; imports ONLY stdlib + other gokit domain packages (zero third-party SDKs)
domain/doc.go           ← package doc
domain/provider/        ← one sub-package per concrete provider; the ONLY place SDKs are imported
    provider.go         ← var _ domain.Interface = (*Provider)(nil)  (compile-time conformance assertion)
```

Example: `cache/cache.go` declares `Cache`; `cache/redis/`, `cache/memory/`, `cache/ristretto/`, `cache/dragonfly/` each implement it. Consumers import only the providers they use, so unused providers (and their SDKs) add zero overhead. The category list and interface names live in `README.md`.

### Repo map

Top-level dirs are categories (each `category/category.go` = interface, sub-dirs = providers). Root files: `config.go` (aggregate config, package `gokit`), `go.mod`, `README.md` (per-category usage examples), `CONTRIBUTING.md` (conventions + provider template), `.planning/plan.md` (provider tracking).

```
ai/        Provider — Chat/Embed     openai, antrophic, ollama, openrouter
analytics/ Tracker                   posthog
auth/      Provider — tokens/verify  basic(JWT), clerk, supabase, oauth
cache/     Cache                     memory, redis, ristretto, dragonfly
config/    Source — typed loader     env/.env/YAML/JSON, generic config.New[T]
cron/      Scheduler                 robfig/cron, gocron
db/        DB                        postgres, mysql, mongodb, cockroachdb
docstore/  Store                     dynamodb, firestore, mongodb, redis
email/     Mailer                    smtp, resend, sendgrid
errorz/    structured error          Is/As/Wrap/HasCode/Log (used everywhere)
geo/       Geocoder                  nominatim
http/      Client / server           retry client, graceful-shutdown server
lock/      Lock                      memory, redis
logger/    Logger                    slog, zap
middleware/Middleware                cors, ratelimit, recovery, auth, requestid, timeout, logger, gzip
monitoring/Metrics                   otel, prometheus, statsd
payment/   Gateway                   stripe, paypal, midtrans, xendit, razorpay, doku
pubsub/    PubSub                    memory, kafka, nsq
queue/     Enqueuer/Worker           asynq
response/  JSON envelope helpers     (no interface)
storage/   Storage                   s3, gcs, r2, local
tracking/  Tracker                   sentry
tracing/   Tracer                    otel, datadog, jaeger, zipkin
utils/     generics/page/date        (no interface)
vector/    Store                     pgvector
```

(`README.md` has the authoritative table + code examples per category.)

### Root config (`config.go`)

`config.go` (package `gokit`) aggregates every subsystem's config into one `Config` struct using only the `config` tag. Env var names are auto-derived from the nested field path (`Config.Auth.Basic.Secret` → `AUTH_BASIC_SECRET`). Key rule: it **reuses domain configs directly** (e.g. `ai.Config`) and uses **inline structs** for provider-specific fields — it must never import a provider sub-package, since that would pull every SDK into every consumer of the root package. The `config` package itself (`config/`) implements typed loading from env/.env/YAML/JSON with validation and generic `config.New[T]`.

## Conventions (enforced by CONTRIBUTING.md and CI)

- **Interface first**: domain root file imports zero third-party SDKs.
- **One provider per sub-package**; only sub-packages import SDKs.
- **Single `config` tag** on Config structs — never `env`/`json`/`yaml`.
- **Context first**: every I/O method takes `context.Context` first and honors cancellation.
- **Wrap errors with package prefix**: `fmt.Errorf("redis: get %s: %w", key, err)`.
- **No new dependencies** without discussion — prefer `net/http` for REST API providers over vendor SDKs.
- Table-driven tests with `httptest` for HTTP providers; target ≥40% coverage, ≥80% for core packages.
- Errors use the `errorz` package (structured, `Is`/`As`/`Wrap`/code-based); see README "Errorz".

### Adding a provider

1. Create `domain/provider/provider.go` from the template in `CONTRIBUTING.md`.
2. Add a config block to the root `config.go` (reuse the domain `Config` if shapes match, else inline struct).
3. Add a row to `.planning/plan.md`.
4. Write the test; run the local checks above.

## Commits

Conventional commits, one logical change per commit: `feat(auth): add Apple OAuth provider`, `fix(config): normalize provider config tags`, `test(payment): cover stripe webhook verification`.
