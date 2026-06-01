# gokit

A production-grade Go toolkit providing clean abstraction interfaces and concrete provider implementations for common backend infrastructure patterns.

**One interface, many providers. Swap implementations without changing your application code.**

```bash
go get github.com/azcov/gokit@latest
```

---

## Why gokit?

- **Interface-first** — every category defines a minimal Go interface; providers implement it
- **Zero lock-in** — switch from Redis to Dragonfly, Sentry to Rollbar, OpenAI to Gemini with one line
- **No magic** — no global state, no init() side effects, explicit construction
- **Production-ready** — structured errors, context propagation, graceful shutdown, retry logic

---

## Categories

| Category | Interface | Providers |
|----------|-----------|-----------|
| [AI](#ai) | `Provider` | OpenAI, Anthropic, OpenRouter, Ollama |
| [Analytics](#analytics) | `Tracker` | PostHog |
| [Auth](#auth) | `Provider` | JWT, Clerk, Supabase |
| [Cache](#cache) | `Cache` | Memory, Redis, Ristretto, Dragonfly |
| [Config](#config) | `Source` | Env, .env, YAML, JSON |
| [Cron](#cron) | `Scheduler` | robfig/cron, gocron |
| [DB](#db) | `DB` | PostgreSQL, MySQL, MongoDB, CockroachDB |
| [Docstore](#docstore) | `Store` | DynamoDB, Firestore, MongoDB, Redis |
| [Email](#email) | `Mailer` | SMTP, Resend, SendGrid |
| [Errorz](#errorz) | `error` | Structured errors with Is/As/Wrap/Log |
| [Geo](#geo) | `Geocoder` | Nominatim (OpenStreetMap) |
| [HTTP](#http) | `Client` | Retry client, graceful-shutdown server |
| [Lock](#lock) | `Lock` | Memory, Redis |
| [Logger](#logger) | `Logger` | slog, Uber Zap |
| [Middleware](#middleware) | `Middleware` | CORS, rate limit, recovery, auth, request-id, timeout, logger, gzip |
| [Monitoring](#monitoring) | `Metrics` | OpenTelemetry, Prometheus, StatsD |
| [Payment](#payment) | `Gateway` | Stripe, PayPal, Midtrans, Xendit, Razorpay, DOKU |
| [PubSub](#pubsub) | `PubSub` | Memory, Kafka, NSQ |
| [Queue](#queue) | `Enqueuer` / `Worker` | Asynq |
| [Response](#response) | — | JSON envelope helpers |
| [Storage](#storage) | `Storage` | S3, GCS, R2, Local |
| [Tracking](#tracking) | `Tracker` | Sentry |
| [Tracing](#tracing) | `Tracer` | OpenTelemetry, Datadog, Jaeger, Zipkin |
| [Utils](#utils) | — | Generics, pagination, date/time/UUID |
| [Vector](#vector) | `Store` | pgvector |

---

## AI

```go
import (
    "github.com/azcov/gokit/ai"
    "github.com/azcov/gokit/ai/openai"
    "github.com/azcov/gokit/ai/antrophic"
    "github.com/azcov/gokit/ai/ollama"
)

// OpenAI
client := openai.New(ai.Config{APIKey: "sk-...", Model: "gpt-4o"})

// Anthropic Claude
client := antrophic.New(ai.Config{APIKey: "sk-ant-...", Model: "claude-3-5-sonnet-20241022"})

// Ollama (local, no API key)
client := ollama.New(ai.Config{Model: "llama3"})

// Same interface for all
resp, err := client.Chat(ctx, []ai.Message{
    {Role: ai.RoleSystem, Content: "You are a helpful assistant."},
    {Role: ai.RoleUser, Content: "Hello!"},
})
embedding, err := client.Embed(ctx, "some text to embed")
```

---

## Analytics

```go
import (
    "github.com/azcov/gokit/analytics"
    "github.com/azcov/gokit/analytics/posthog"
)

tracker := posthog.New(analytics.Config{APIKey: "phc_..."})

tracker.Track(ctx, analytics.Event{
    Name:       "user_signed_up",
    UserID:     "user_123",
    Properties: analytics.Properties{"plan": "pro"},
})
tracker.Identify(ctx, analytics.Trait{
    UserID:     "user_123",
    Properties: analytics.Properties{"email": "user@example.com"},
})
```

---

## Auth

```go
import (
    "github.com/azcov/gokit/auth"
    "github.com/azcov/gokit/auth/basic"
    "github.com/azcov/gokit/auth/clerk"
)

// JWT (self-issued)
provider := basic.New(basic.Config{Secret: "my-secret", Expiry: 24 * time.Hour})
token, err := provider.CreateToken(ctx, auth.Claims{UserID: "123", Email: "user@example.com"})
claims, err := provider.Verify(ctx, token)

// Clerk
provider := clerk.New(clerk.Config{SecretKey: "sk_live_..."})
claims, err := provider.Verify(ctx, jwtToken)
```

---

## Cache

```go
import (
    "github.com/azcov/gokit/cache"
    "github.com/azcov/gokit/cache/redis"
    "github.com/azcov/gokit/cache/memory"
)

// Redis
c := redis.New(redis.Config{Addr: "localhost:6379"})

// In-memory (development / tests)
c := memory.New()

// Same interface
err := c.Set(ctx, "key", []byte("value"), 5*time.Minute)
val, err := c.Get(ctx, "key")
exists, err := c.Exists(ctx, "key")
```

---

## Config

```go
import "github.com/azcov/gokit/config"

type AppConfig struct {
    Port     int    `config:"port"`
    Database struct {
        DSN string `config:"dsn" validate:"required"`
    } `config:"database"`
}

cfg, err := config.New[AppConfig](
    config.WithEnv(),
    config.WithYAML("config.yaml"),
    config.WithValidation(),
)
```

---

## DB

```go
import (
    "github.com/azcov/gokit/db"
    "github.com/azcov/gokit/db/postgres"
)

db, err := postgres.New(ctx, postgres.Config{DSN: "postgres://..."})

rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
defer rows.Close()
```

---

## Email

```go
import (
    "github.com/azcov/gokit/email"
    "github.com/azcov/gokit/email/resend"
    "github.com/azcov/gokit/email/smtp"
)

// Resend
mailer := resend.New("re_...")

// SMTP
mailer := smtp.New(smtp.Config{Host: "smtp.gmail.com", Port: 587, ...})

// Same interface
err := mailer.Send(ctx, email.Message{
    From:    email.Address{Name: "Acme", Email: "noreply@acme.com"},
    To:      []email.Address{{Email: "user@example.com"}},
    Subject: "Welcome!",
    HTML:    "<h1>Hello</h1>",
})
```

---

## Errorz

```go
import "github.com/azcov/gokit/errorz"

// Define sentinels
var (
    ErrNotFound   = errorz.New(404, 5, "NOT_FOUND", "resource not found")
    ErrUnauthorized = errorz.New(401, 16, "UNAUTHORIZED", "unauthorized")
)

// Wrap with a cause
err := ErrNotFound.Wrap(sql.ErrNoRows).WithDetail("id", userID)

// Compare
errorz.Is(err, ErrNotFound)       // true
errorz.HasCode(err, "NOT_FOUND")  // true
errors.Is(err, sql.ErrNoRows)     // true — walks the cause chain

// Log
err.Log()      // code=NOT_FOUND http=404 rpc=5 message="resource not found" cause="sql: no rows"
err.LogValue() // map[string]any{...}

// Convert to error interface
return err.Err() // nil-safe, returns error
```

---

## HTTP

```go
import (
    "github.com/azcov/gokit/http/client"
    "github.com/azcov/gokit/http/server"
    gkhttp "github.com/azcov/gokit/http"
)

// Client with retry
c := client.New(gkhttp.ClientConfig{
    Timeout:    10 * time.Second,
    MaxRetries: 3,
})
resp, err := c.Get("https://api.example.com/data")

// Server with graceful shutdown
srv := server.New(server.Config{Addr: ":8080"}, mux)
go srv.Start()
// on signal:
srv.Shutdown()
```

---

## Middleware

```go
import (
    "github.com/azcov/gokit/middleware"
    "github.com/azcov/gokit/middleware/cors"
    "github.com/azcov/gokit/middleware/recovery"
    "github.com/azcov/gokit/middleware/requestid"
    "github.com/azcov/gokit/middleware/logger"
    "github.com/azcov/gokit/middleware/ratelimit"
    mwauth "github.com/azcov/gokit/middleware/auth"
)

chain := middleware.Chain(
    recovery.New(),
    requestid.New(),
    logger.New(log),
    cors.Default(),
    ratelimit.New(ratelimit.Config{RequestsPerSecond: 100}),
    mwauth.New(mwauth.Config{Provider: jwtProvider}),
)

http.Handle("/", chain(mux))

// Read request ID in handler
id := requestid.FromContext(r.Context())

// Read auth claims in handler
claims, ok := mwauth.ClaimsFromContext(r.Context())
```

---

## Monitoring

```go
import (
    "github.com/azcov/gokit/monitoring"
    "github.com/azcov/gokit/monitoring/prometheus"
)

prom := prometheus.New(monitoring.Config{ServiceName: "my-service"})

requests := prom.Counter("http_requests_total", "Total HTTP requests")
latency  := prom.Histogram("http_request_duration_seconds", "Request latency", nil)

requests.Add(ctx, 1, monitoring.Label{Key: "method", Value: "GET"})
latency.Record(ctx, 0.042)

// Expose /metrics
http.Handle("/metrics", prom)
```

---

## Payment

```go
import (
    "github.com/azcov/gokit/payment"
    "github.com/azcov/gokit/payment/stripe"
)

gw := stripe.New(stripe.Config{SecretKey: "sk_live_..."})

charge, err := gw.CreateCharge(ctx, payment.ChargeRequest{
    Amount:      5000, // cents
    Currency:    payment.USD,
    Method:      payment.MethodCard,
    Description: "Order #123",
    Token:       "tok_visa",
})

event, err := gw.VerifyWebhook(ctx, body, headers)
```

---

## PubSub

```go
import (
    "github.com/azcov/gokit/pubsub"
    "github.com/azcov/gokit/pubsub/kafka"
)

ps := kafka.New(kafka.Config{Brokers: []string{"localhost:9092"}})

// Publisher
ps.Publish(ctx, "orders", pubsub.Message{Payload: []byte(`{"id":"123"}`)})

// Subscriber
ps.Subscribe(ctx, "orders", func(ctx context.Context, msg pubsub.Message) error {
    fmt.Println(string(msg.Payload))
    return nil
})
```

---

## Storage

```go
import (
    "github.com/azcov/gokit/storage"
    "github.com/azcov/gokit/storage/s3"
    "github.com/azcov/gokit/storage/local"
)

// AWS S3
store, err := s3.New(ctx, s3.Config{Bucket: "my-bucket", Region: "us-east-1"})

// Local filesystem (dev/test)
store, err := local.New("./uploads")

// Same interface
err = store.Put(ctx, "images/avatar.png", r, storage.PutOptions{ContentType: "image/png"})
rc, err := store.Get(ctx, "images/avatar.png")
url, err := store.URL(ctx, "images/avatar.png")
```

---

## Tracing

```go
import (
    "github.com/azcov/gokit/tracing"
    "github.com/azcov/gokit/tracing/datadog"
    "github.com/azcov/gokit/tracing/jaeger"
)

tracer := datadog.New(tracing.Config{ServiceName: "api", Endpoint: "http://dd-agent:8126"})
// or
tracer := jaeger.New(tracing.Config{ServiceName: "api", Endpoint: "http://jaeger:9411"})

ctx, span := tracer.Start(ctx, "db.query",
    tracing.WithAttribute("db.table", "users"),
)
defer span.End()

span.SetAttribute("rows", 42)
span.RecordError(err)
```

---

## Utils

```go
import (
    "github.com/azcov/gokit/utils"
    "github.com/azcov/gokit/utils/page"
)

// Generics
evens  := utils.Filter([]int{1,2,3,4}, func(n int) bool { return n%2 == 0 })
doubled := utils.Map(evens, func(n int) int { return n * 2 })
unique  := utils.Unique([]string{"a","b","a","c"})

// Retry with backoff
err := utils.Retry(3, 100*time.Millisecond, func() error {
    return callExternalAPI()
})

// Pagination
p := page.New(2, 20, 150) // page 2, 20 per page, 150 total
rows, _ := db.QueryContext(ctx, "SELECT * FROM items LIMIT $1 OFFSET $2", p.Limit(), p.Offset())
resp := page.NewResponse(items, p) // {items, page, size, total, total_pages, has_next, has_prev}
```

---

## Vector (pgvector)

```go
import (
    "github.com/azcov/gokit/vector"
    "github.com/azcov/gokit/vector/pgvector"
)

store := pgvector.New(pool, vector.Config{Dimension: 1536, Namespace: "embeddings"})
store.CreateTable(ctx) // CREATE TABLE IF NOT EXISTS embeddings (id TEXT, embedding vector(1536), ...)

// Upsert
store.Upsert(ctx, []vector.Record{
    {ID: "doc_1", Vector: embedding, Metadata: map[string]any{"title": "Hello"}},
})

// Semantic search
results, err := store.Query(ctx, queryVector, vector.QueryOptions{TopK: 5})
for _, r := range results {
    fmt.Printf("id=%s score=%.4f\n", r.ID, r.Score)
}
```

---

## Architecture

Every category follows the same pattern:

```
domain/domain.go          ← interface + shared types (zero provider imports)
domain/provider/          ← concrete implementation
    provider.go           ← var _ domain.Interface = (*Provider)(nil)
```

Provider config uses `config` tags for automatic env/YAML/JSON loading:

```go
type Config struct {
    APIKey  string        `config:"api_key"`   // env: MYAPP_EMAIL_API_KEY
    Timeout time.Duration `config:"timeout"`   // env: MYAPP_EMAIL_TIMEOUT
}
```

---

## Requirements

- Go 1.21+
- Import only the providers you need — unused providers add zero overhead

---

## License

MIT
