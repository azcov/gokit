package gokit

import (
	"time"

	"github.com/azcov/gokit/ai"
	"github.com/azcov/gokit/analytics"
	"github.com/azcov/gokit/config"
	"github.com/azcov/gokit/geo"
	"github.com/azcov/gokit/monitoring"
	"github.com/azcov/gokit/tracing"
	"github.com/azcov/gokit/tracking"
	"github.com/azcov/gokit/vector"
)

// Config aggregates configuration for all gokit subsystems.
//
// Rules:
//   - Domain configs (ai.Config, analytics.Config, …) are reused directly —
//     no duplication of fields already defined in the domain package.
//   - Provider-specific configs that differ from the domain shape use inline
//     structs here, because importing provider packages would pull their SDKs
//     into every consumer of this root package.
//   - Env vars are auto-derived from the nested path:
//     Config.Auth.Basic.Secret → AUTH_BASIC_SECRET
type Config struct {
	AI           AIConfig           `config:"ai"`
	Analytics    AnalyticsConfig    `config:"analytics"`
	Auth         AuthConfig         `config:"auth"`
	Cache        CacheConfig        `config:"cache"`
	Chat         ChatConfig         `config:"chat"`
	Cron         CronConfig         `config:"cron"`
	DB           DBConfig           `config:"db"`
	Docstore     DocstoreConfig     `config:"docstore"`
	Email        EmailConfig        `config:"email"`
	Feature      FeatureConfig      `config:"feature"`
	Geo          GeoConfig          `config:"geo"`
	Lock         LockConfig         `config:"lock"`
	Media        MediaConfig        `config:"media"`
	Monitoring   MonitoringConfig   `config:"monitoring"`
	Notification NotificationConfig `config:"notification"`
	Payment      PaymentConfig      `config:"payment"`
	PubSub       PubSubConfig       `config:"pubsub"`
	Queue        QueueConfig        `config:"queue"`
	Search       SearchConfig       `config:"search"`
	Secret       SecretConfig       `config:"secret"`
	SMS          SMSConfig          `config:"sms"`
	Storage      StorageConfig      `config:"storage"`
	Tracking     TrackingConfig     `config:"tracking"`
	Tracing      TracingConfig      `config:"tracing"`
	Vector       VectorConfig       `config:"vector"`
}

// ── AI ────────────────────────────────────────────────────────────────────────
// All AI providers accept ai.Config — reuse it directly.

type AIConfig struct {
	OpenAI     ai.Config `config:"openai"`
	Anthropic  ai.Config `config:"anthropic"`
	Gemini     ai.Config `config:"gemini"`
	Mistral    ai.Config `config:"mistral"`
	Groq       ai.Config `config:"groq"`
	Ollama     ai.Config `config:"ollama"`
	Cohere     ai.Config `config:"cohere"`
	OpenRouter struct {
		ai.Config        // BaseURL, Model, APIKey, Timeout
		Referer   string `config:"referer"`
		AppTitle  string `config:"app_title"`
	} `config:"openrouter"`
}

// ── Analytics ─────────────────────────────────────────────────────────────────
// analytics.Config is the domain config reused by all analytics providers.

type AnalyticsConfig struct {
	PostHog     analytics.Config `config:"posthog"`
	Mixpanel    analytics.Config `config:"mixpanel"`
	Segment     analytics.Config `config:"segment"`
	Amplitude   analytics.Config `config:"amplitude"`
	RudderStack analytics.Config `config:"rudderstack"`
}

// ── Auth ──────────────────────────────────────────────────────────────────────
// Auth providers each have different fields; inline structs avoid SDK imports.

type AuthConfig struct {
	Basic struct {
		Secret string        `config:"secret"`
		TTL    time.Duration `config:"ttl"`
	} `config:"basic"`

	Clerk struct {
		SecretKey string `config:"secret_key"`
	} `config:"clerk"`

	Supabase struct {
		JWTSecret      string `config:"jwt_secret"`
		ProjectURL     string `config:"project_url"`
		ServiceRoleKey string `config:"service_role_key"`
	} `config:"supabase"`

	BetterAuth struct {
		BaseURL string        `config:"base_url"`
		Secret  string        `config:"secret"`
		Timeout time.Duration `config:"timeout"`
	} `config:"betterauth"`

	Firebase struct {
		ProjectID string `config:"project_id"`
	} `config:"firebase"`

	Auth0 struct {
		Domain       string `config:"domain"`
		ClientID     string `config:"client_id"`
		ClientSecret string `config:"client_secret"`
	} `config:"auth0"`

	Cognito struct {
		Region     string `config:"region"`
		UserPoolID string `config:"user_pool_id"`
		ClientID   string `config:"client_id"`
	} `config:"cognito"`

	// OAuth holds credentials for each social login provider.
	// ClientID + ClientSecret are the only loadable fields (redirect URIs and
	// state are request-scoped, not config-scoped).
	OAuth struct {
		Google struct {
			ClientID     string `config:"client_id"`
			ClientSecret string `config:"client_secret"`
		} `config:"google"`

		GitHub struct {
			ClientID     string `config:"client_id"`
			ClientSecret string `config:"client_secret"`
		} `config:"github"`

		Facebook struct {
			ClientID     string `config:"client_id"`
			ClientSecret string `config:"client_secret"`
		} `config:"facebook"`

		Discord struct {
			ClientID     string `config:"client_id"`
			ClientSecret string `config:"client_secret"`
		} `config:"discord"`

		Microsoft struct {
			ClientID     string `config:"client_id"`
			ClientSecret string `config:"client_secret"`
			TenantID     string `config:"tenant_id"`
		} `config:"microsoft"`

		Twitter struct {
			ClientID     string `config:"client_id"`
			ClientSecret string `config:"client_secret"`
		} `config:"twitter"`

		Apple struct {
			ClientID string `config:"client_id"`
			TeamID   string `config:"team_id"`
			KeyID    string `config:"key_id"`
			KeyPEM   string `config:"key_pem"`
		} `config:"apple"`
	} `config:"oauth"`
}

// ── Cache ─────────────────────────────────────────────────────────────────────

type CacheConfig struct {
	Redis struct {
		Address  string        `config:"address"`
		Password string        `config:"password"`
		DB       int           `config:"db"`
		Timeout  time.Duration `config:"timeout"`
	} `config:"redis"`

	Dragonfly struct {
		Address  string        `config:"address"`
		Password string        `config:"password"`
		DB       int           `config:"db"`
		Timeout  time.Duration `config:"timeout"`
	} `config:"dragonfly"`
}

// ── Chat ──────────────────────────────────────────────────────────────────────

type ChatConfig struct {
	Slack struct {
		BotToken   string `config:"bot_token"`
		WebhookURL string `config:"webhook_url"`
	} `config:"slack"`

	Telegram struct {
		BotToken string `config:"bot_token"`
	} `config:"telegram"`

	Discord struct {
		BotToken   string `config:"bot_token"`
		WebhookURL string `config:"webhook_url"`
	} `config:"discord"`

	GoogleChat struct {
		WebhookURL string `config:"webhook_url"`
	} `config:"googlechat"`

	Lark struct {
		WebhookURL string `config:"webhook_url"`
		AppID      string `config:"app_id"`
		AppSecret  string `config:"app_secret"`
	} `config:"lark"`

	MSTeams struct {
		WebhookURL string `config:"webhook_url"`
	} `config:"msteams"`

	WhatsApp struct {
		AccessToken   string `config:"access_token"`
		PhoneNumberID string `config:"phone_number_id"`
	} `config:"whatsapp"`
}

// ── Cron ──────────────────────────────────────────────────────────────────────

type CronConfig struct {
	Location string `config:"location"` // IANA timezone, default UTC
}

// ── DB ────────────────────────────────────────────────────────────────────────

type DBConfig struct {
	Postgres struct {
		DSN          string `config:"dsn"`
		MaxOpenConns int    `config:"max_open_conns"`
		MaxIdleConns int    `config:"max_idle_conns"`
	} `config:"postgres"`

	MySQL struct {
		DSN          string `config:"dsn"`
		MaxOpenConns int    `config:"max_open_conns"`
		MaxIdleConns int    `config:"max_idle_conns"`
	} `config:"mysql"`

	Mongo struct {
		URI      string `config:"uri"`
		Database string `config:"database"`
	} `config:"mongo"`

	CockroachDB struct {
		DSN          string `config:"dsn"`
		MaxOpenConns int    `config:"max_open_conns"`
		MaxIdleConns int    `config:"max_idle_conns"`
	} `config:"cockroachdb"`
}

// ── Docstore ──────────────────────────────────────────────────────────────────

type DocstoreConfig struct {
	Mongo struct {
		URI      string `config:"uri"`
		Database string `config:"database"`
	} `config:"mongo"`

	DynamoDB struct {
		Region          string `config:"region"`
		Table           string `config:"table"`
		PartitionKey    string `config:"partition_key"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
	} `config:"dynamodb"`

	Firestore struct {
		ProjectID string `config:"project_id"`
	} `config:"firestore"`

	Redis struct {
		Address   string `config:"address"`
		Password  string `config:"password"`
		DB        int    `config:"db"`
		Namespace string `config:"namespace"`
	} `config:"redis"`
}

// ── Email ─────────────────────────────────────────────────────────────────────

type EmailConfig struct {
	SMTP struct {
		Host     string `config:"host"`
		Port     int    `config:"port"`
		Username string `config:"username"`
		Password string `config:"password"`
		From     string `config:"from"`
	} `config:"smtp"`

	Resend struct {
		APIKey string `config:"api_key"`
	} `config:"resend"`

	SendGrid struct {
		APIKey string `config:"api_key"`
	} `config:"sendgrid"`

	Mailgun struct {
		Domain string `config:"domain"`
		APIKey string `config:"api_key"`
	} `config:"mailgun"`

	Postmark struct {
		ServerToken string `config:"server_token"`
	} `config:"postmark"`

	Brevo struct {
		APIKey string `config:"api_key"`
	} `config:"brevo"`
}

// ── Feature Flags ─────────────────────────────────────────────────────────────

type FeatureConfig struct {
	LaunchDarkly struct {
		SDKKey string `config:"sdk_key"`
	} `config:"launchdarkly"`

	Unleash struct {
		URL     string `config:"url"`
		APIKey  string `config:"api_key"`
		AppName string `config:"app_name"`
	} `config:"unleash"`

	Flagsmith struct {
		EnvironmentKey string `config:"environment_key"`
		BaseURL        string `config:"base_url"`
	} `config:"flagsmith"`

	// PostHog feature flags reuse the same APIKey as analytics/posthog
	// but target the /decide endpoint — kept separate for clarity.
	PostHog struct {
		APIKey   string `config:"api_key"`
		Endpoint string `config:"endpoint"`
	} `config:"posthog"`

	GrowthBook struct {
		APIKey   string `config:"api_key"`
		Endpoint string `config:"endpoint"`
	} `config:"growthbook"`
}

// ── Geo ───────────────────────────────────────────────────────────────────────
// geo.Config is the domain config reused by all geo providers.

type GeoConfig struct {
	GoogleMaps geo.Config `config:"googlemaps"`
	Mapbox     geo.Config `config:"mapbox"`
	Nominatim  geo.Config `config:"nominatim"`
	HERE       geo.Config `config:"here"`
}

// ── Lock ──────────────────────────────────────────────────────────────────────

type LockConfig struct {
	Redis struct {
		Address  string `config:"address"`
		Password string `config:"password"`
		DB       int    `config:"db"`
	} `config:"redis"`
}

// ── Media ─────────────────────────────────────────────────────────────────────

type MediaConfig struct {
	Cloudinary struct {
		CloudName string `config:"cloud_name"`
		APIKey    string `config:"api_key"`
		APISecret string `config:"api_secret"`
	} `config:"cloudinary"`

	ImageKit struct {
		PublicKey   string `config:"public_key"`
		PrivateKey  string `config:"private_key"`
		URLEndpoint string `config:"url_endpoint"`
	} `config:"imagekit"`

	Uploadcare struct {
		PublicKey string `config:"public_key"`
		SecretKey string `config:"secret_key"`
	} `config:"uploadcare"`
}

// ── Monitoring ────────────────────────────────────────────────────────────────
// monitoring.Config is the domain config reused by otel and prometheus.

type MonitoringConfig struct {
	OTel       monitoring.Config `config:"otel"`
	Prometheus monitoring.Config `config:"prometheus"`
	StatsD     struct {
		monitoring.Config        // ServiceName, Endpoint
		Addr              string `config:"addr"`
		Prefix            string `config:"prefix"`
	} `config:"statsd"`
	Datadog struct {
		monitoring.Config
		Addr   string `config:"addr"`
		Prefix string `config:"prefix"`
	} `config:"datadog"`
}

// ── Notification ──────────────────────────────────────────────────────────────

type NotificationConfig struct {
	FCM struct {
		ServiceAccountJSON string `config:"service_account_json"`
		ProjectID          string `config:"project_id"`
	} `config:"fcm"`

	APNs struct {
		KeyPEM   string `config:"key_pem"`
		KeyID    string `config:"key_id"`
		TeamID   string `config:"team_id"`
		BundleID string `config:"bundle_id"`
		Sandbox  bool   `config:"sandbox"`
	} `config:"apns"`

	OneSignal struct {
		AppID  string `config:"app_id"`
		APIKey string `config:"api_key"`
	} `config:"onesignal"`
}

// ── Payment ───────────────────────────────────────────────────────────────────

type PaymentConfig struct {
	Stripe struct {
		SecretKey     string `config:"secret_key"`
		WebhookSecret string `config:"webhook_secret"`
	} `config:"stripe"`

	PayPal struct {
		ClientID     string `config:"client_id"`
		ClientSecret string `config:"client_secret"`
		WebhookID    string `config:"webhook_id"`
		Production   bool   `config:"production"`
	} `config:"paypal"`

	Midtrans struct {
		ServerKey  string `config:"server_key"`
		Production bool   `config:"production"`
	} `config:"midtrans"`

	Xendit struct {
		SecretKey     string `config:"secret_key"`
		CallbackToken string `config:"callback_token"`
	} `config:"xendit"`

	Razorpay struct {
		KeyID         string `config:"key_id"`
		KeySecret     string `config:"key_secret"`
		WebhookSecret string `config:"webhook_secret"`
	} `config:"razorpay"`

	DOKU struct {
		ClientID   string `config:"client_id"`
		SecretKey  string `config:"secret_key"`
		Production bool   `config:"production"`
	} `config:"doku"`
}

// ── PubSub ────────────────────────────────────────────────────────────────────

type PubSubConfig struct {
	Kafka struct {
		Brokers []string `config:"brokers"`
		GroupID string   `config:"group_id"`
	} `config:"kafka"`

	NSQ struct {
		NSQDAddress    string `config:"nsqd_address"`
		LookupdAddress string `config:"lookupd_address"`
	} `config:"nsq"`

	Redis struct {
		Address  string `config:"address"`
		Password string `config:"password"`
		DB       int    `config:"db"`
	} `config:"redis"`

	RabbitMQ struct {
		URL string `config:"url"`
	} `config:"rabbitmq"`

	Google struct {
		ProjectID string `config:"project_id"`
	} `config:"google"`
}

// ── Queue ─────────────────────────────────────────────────────────────────────

type QueueConfig struct {
	Asynq struct {
		RedisAddress string         `config:"redis_address"`
		Concurrency  int            `config:"concurrency"`
		Queues       map[string]int `config:"queues"`
	} `config:"asynq"`

	SQS struct {
		Region          string `config:"region"`
		QueueURL        string `config:"queue_url"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
	} `config:"sqs"`

	RabbitMQ struct {
		URL   string `config:"url"`
		Queue string `config:"queue"`
	} `config:"rabbitmq"`
}

// ── Search ────────────────────────────────────────────────────────────────────

type SearchConfig struct {
	Elasticsearch struct {
		Addresses []string `config:"addresses"`
		Username  string   `config:"username"`
		Password  string   `config:"password"`
		APIKey    string   `config:"api_key"`
	} `config:"elasticsearch"`

	Meilisearch struct {
		Host   string `config:"host"`
		APIKey string `config:"api_key"`
	} `config:"meilisearch"`

	Typesense struct {
		Host   string `config:"host"`
		APIKey string `config:"api_key"`
	} `config:"typesense"`

	Algolia struct {
		AppID  string `config:"app_id"`
		APIKey string `config:"api_key"`
	} `config:"algolia"`
}

// ── Secret ────────────────────────────────────────────────────────────────────

type SecretConfig struct {
	Vault struct {
		Address  string `config:"address"`
		Token    string `config:"token"`
		RoleID   string `config:"role_id"`
		SecretID string `config:"secret_id"`
	} `config:"vault"`

	AWSSM struct {
		Region          string `config:"region"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
	} `config:"awssm"`

	GCPSM struct {
		ProjectID string `config:"project_id"`
	} `config:"gcpsm"`

	Doppler struct {
		Token   string `config:"token"`
		Project string `config:"project"`
		Config  string `config:"config"`
	} `config:"doppler"`
}

// ── SMS ───────────────────────────────────────────────────────────────────────

type SMSConfig struct {
	Twilio struct {
		AccountSID string `config:"account_sid"`
		AuthToken  string `config:"auth_token"`
		From       string `config:"from"`
	} `config:"twilio"`

	Vonage struct {
		APIKey    string `config:"api_key"`
		APISecret string `config:"api_secret"`
		From      string `config:"from"`
	} `config:"vonage"`

	MessageBird struct {
		AccessKey string `config:"access_key"`
		From      string `config:"from"`
	} `config:"messagebird"`

	AWSSNS struct {
		Region          string `config:"region"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
	} `config:"awssns"`
}

// ── Storage ───────────────────────────────────────────────────────────────────

type StorageConfig struct {
	S3 struct {
		Bucket          string `config:"bucket"`
		Region          string `config:"region"`
		Endpoint        string `config:"endpoint"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
		Token           string `config:"token"`
	} `config:"s3"`

	GCS struct {
		Bucket string `config:"bucket"`
	} `config:"gcs"`

	R2 struct {
		AccountID       string `config:"account_id"`
		Bucket          string `config:"bucket"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
	} `config:"r2"`

	Local struct {
		Root string `config:"root"`
	} `config:"local"`

	MinIO struct {
		Endpoint        string `config:"endpoint"`
		Bucket          string `config:"bucket"`
		AccessKeyID     string `config:"access_key_id"`
		AccessKeySecret string `config:"access_key_secret"`
		UseSSL          bool   `config:"use_ssl"`
	} `config:"minio"`
}

// ── Tracking ──────────────────────────────────────────────────────────────────
// tracking.Config is the domain config (DSN, Environment, Release).

type TrackingConfig struct {
	Sentry   tracking.Config `config:"sentry"`
	Rollbar  tracking.Config `config:"rollbar"`
	Bugsnag  tracking.Config `config:"bugsnag"`
	NewRelic struct {
		tracking.Config        // Environment, Release
		LicenseKey      string `config:"license_key"`
		AppName         string `config:"app_name"`
	} `config:"newrelic"`
}

// ── Tracing ───────────────────────────────────────────────────────────────────
// tracing.Config is the domain config (ServiceName, Endpoint, SampleRate, Timeout).

type TracingConfig struct {
	OTel    tracing.Config `config:"otel"`
	Datadog tracing.Config `config:"datadog"`
	Jaeger  tracing.Config `config:"jaeger"`
	Zipkin  tracing.Config `config:"zipkin"`
	XRay    struct {
		tracing.Config
		Region     string `config:"region"`
		DaemonAddr string `config:"daemon_addr"`
	} `config:"xray"`
}

// ── Vector ────────────────────────────────────────────────────────────────────
// vector.Config is the domain config (APIKey, BaseURL, Dimension, Namespace).

type VectorConfig struct {
	PgVector vector.Config `config:"pgvector"`
	Pinecone vector.Config `config:"pinecone"`
	Qdrant   vector.Config `config:"qdrant"`
	Weaviate vector.Config `config:"weaviate"`
	ChromaDB vector.Config `config:"chromadb"`
}

// ── loader ────────────────────────────────────────────────────────────────────

// LoadConfig loads cfg from env vars, then from any YAML/JSON files provided.
// Later sources override earlier ones.
func LoadConfig(cfg *Config, paths ...string) error {
	loaders := []config.Loader{config.FromEnv()}
	for _, p := range paths {
		switch {
		case hasExt(p, ".yaml"), hasExt(p, ".yml"):
			loaders = append(loaders, config.FromYAML(p))
		case hasExt(p, ".json"):
			loaders = append(loaders, config.FromJSON(p))
		}
	}
	return config.Chain(loaders...).Load(cfg)
}

func hasExt(path, ext string) bool {
	return len(path) >= len(ext) && path[len(path)-len(ext):] == ext
}
