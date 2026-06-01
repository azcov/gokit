package gokit

import (
	"time"

	"github.com/azcov/gokit/config"
)

// Config aggregates configuration for all gokit subsystems.
// Env var names are flat and explicit — no prefix concatenation.
// JSON and YAML keys mirror the field names directly.
type Config struct {
	// Auth
	AuthBasicSecret        string        `env:"AUTH_BASIC_SECRET" json:"auth_basic_secret" yaml:"auth_basic_secret"`
	AuthBasicTTL           time.Duration `env:"AUTH_BASIC_TTL" json:"auth_basic_ttl" yaml:"auth_basic_ttl"`
	AuthClerkSecretKey     string        `env:"AUTH_CLERK_SECRET_KEY" json:"auth_clerk_secret_key" yaml:"auth_clerk_secret_key"`
	AuthSupabaseJWTSecret  string        `env:"AUTH_SUPABASE_JWT_SECRET" json:"auth_supabase_jwt_secret" yaml:"auth_supabase_jwt_secret"`

	// Cache
	CacheRedisAddress  string `env:"CACHE_REDIS_ADDRESS" json:"cache_redis_address" yaml:"cache_redis_address"`
	CacheRedisPassword string `env:"CACHE_REDIS_PASSWORD" json:"cache_redis_password" yaml:"cache_redis_password"`
	CacheRedisDB       int    `env:"CACHE_REDIS_DB" json:"cache_redis_db" yaml:"cache_redis_db"`

	// Database
	DBMongoURI      string `env:"DB_MONGO_URI" json:"db_mongo_uri" yaml:"db_mongo_uri"`
	DBMongoDatabase string `env:"DB_MONGO_DATABASE" json:"db_mongo_database" yaml:"db_mongo_database"`
	DBPostgresDSN   string `env:"DB_POSTGRES_DSN" json:"db_postgres_dsn" yaml:"db_postgres_dsn"`
	DBMySQLDSN      string `env:"DB_MYSQL_DSN" json:"db_mysql_dsn" yaml:"db_mysql_dsn"`

	// Docstore
	DocstoreMongoURI           string `env:"DOCSTORE_MONGO_URI" json:"docstore_mongo_uri" yaml:"docstore_mongo_uri"`
	DocstoreMongoDatabase      string `env:"DOCSTORE_MONGO_DATABASE" json:"docstore_mongo_database" yaml:"docstore_mongo_database"`
	DocstoreDynamoPartitionKey string `env:"DOCSTORE_DYNAMO_PARTITION_KEY" json:"docstore_dynamo_partition_key" yaml:"docstore_dynamo_partition_key"`
	DocstoreFirestoreProjectID string `env:"DOCSTORE_FIRESTORE_PROJECT_ID" json:"docstore_firestore_project_id" yaml:"docstore_firestore_project_id"`

	// Email
	EmailSMTPHost     string `env:"EMAIL_SMTP_HOST" json:"email_smtp_host" yaml:"email_smtp_host"`
	EmailSMTPPort     int    `env:"EMAIL_SMTP_PORT" json:"email_smtp_port" yaml:"email_smtp_port"`
	EmailSMTPUsername string `env:"EMAIL_SMTP_USERNAME" json:"email_smtp_username" yaml:"email_smtp_username"`
	EmailSMTPPassword string `env:"EMAIL_SMTP_PASSWORD" json:"email_smtp_password" yaml:"email_smtp_password"`
	EmailResendAPIKey string `env:"EMAIL_RESEND_API_KEY" json:"email_resend_api_key" yaml:"email_resend_api_key"`
	EmailSendGridAPIKey string `env:"EMAIL_SENDGRID_API_KEY" json:"email_sendgrid_api_key" yaml:"email_sendgrid_api_key"`

	// Payment
	PaymentStripeSecretKey       string `env:"PAYMENT_STRIPE_SECRET_KEY" json:"payment_stripe_secret_key" yaml:"payment_stripe_secret_key"`
	PaymentStripeWebhookSecret   string `env:"PAYMENT_STRIPE_WEBHOOK_SECRET" json:"payment_stripe_webhook_secret" yaml:"payment_stripe_webhook_secret"`
	PaymentPayPalClientID        string `env:"PAYMENT_PAYPAL_CLIENT_ID" json:"payment_paypal_client_id" yaml:"payment_paypal_client_id"`
	PaymentPayPalClientSecret    string `env:"PAYMENT_PAYPAL_CLIENT_SECRET" json:"payment_paypal_client_secret" yaml:"payment_paypal_client_secret"`
	PaymentPayPalWebhookID       string `env:"PAYMENT_PAYPAL_WEBHOOK_ID" json:"payment_paypal_webhook_id" yaml:"payment_paypal_webhook_id"`
	PaymentPayPalIsProduction    bool   `env:"PAYMENT_PAYPAL_IS_PRODUCTION" json:"payment_paypal_is_production" yaml:"payment_paypal_is_production"`
	PaymentMidtransServerKey     string `env:"PAYMENT_MIDTRANS_SERVER_KEY" json:"payment_midtrans_server_key" yaml:"payment_midtrans_server_key"`
	PaymentMidtransIsProduction  bool   `env:"PAYMENT_MIDTRANS_IS_PRODUCTION" json:"payment_midtrans_is_production" yaml:"payment_midtrans_is_production"`
	PaymentDokuClientID          string `env:"PAYMENT_DOKU_CLIENT_ID" json:"payment_doku_client_id" yaml:"payment_doku_client_id"`
	PaymentDokuSecretKey         string `env:"PAYMENT_DOKU_SECRET_KEY" json:"payment_doku_secret_key" yaml:"payment_doku_secret_key"`
	PaymentDokuIsProduction      bool   `env:"PAYMENT_DOKU_IS_PRODUCTION" json:"payment_doku_is_production" yaml:"payment_doku_is_production"`
	PaymentRazorpayKeyID         string `env:"PAYMENT_RAZORPAY_KEY_ID" json:"payment_razorpay_key_id" yaml:"payment_razorpay_key_id"`
	PaymentRazorpayKeySecret     string `env:"PAYMENT_RAZORPAY_KEY_SECRET" json:"payment_razorpay_key_secret" yaml:"payment_razorpay_key_secret"`
	PaymentRazorpayWebhookSecret string `env:"PAYMENT_RAZORPAY_WEBHOOK_SECRET" json:"payment_razorpay_webhook_secret" yaml:"payment_razorpay_webhook_secret"`
	PaymentXenditSecretKey       string `env:"PAYMENT_XENDIT_SECRET_KEY" json:"payment_xendit_secret_key" yaml:"payment_xendit_secret_key"`
	PaymentXenditCallbackToken   string `env:"PAYMENT_XENDIT_CALLBACK_TOKEN" json:"payment_xendit_callback_token" yaml:"payment_xendit_callback_token"`

	// PubSub
	PubSubKafkaBrokers   []string `env:"PUBSUB_KAFKA_BROKERS" json:"pubsub_kafka_brokers" yaml:"pubsub_kafka_brokers"`
	PubSubKafkaGroupID   string   `env:"PUBSUB_KAFKA_GROUP_ID" json:"pubsub_kafka_group_id" yaml:"pubsub_kafka_group_id"`
	PubSubNSQNSQDAddress    string `env:"PUBSUB_NSQ_NSQD_ADDRESS" json:"pubsub_nsq_nsqd_address" yaml:"pubsub_nsq_nsqd_address"`
	PubSubNSQLookupdAddress string `env:"PUBSUB_NSQ_LOOKUPD_ADDRESS" json:"pubsub_nsq_lookupd_address" yaml:"pubsub_nsq_lookupd_address"`

	// Storage
	StorageS3Bucket          string `env:"STORAGE_S3_BUCKET" json:"storage_s3_bucket" yaml:"storage_s3_bucket"`
	StorageS3Endpoint        string `env:"STORAGE_S3_ENDPOINT" json:"storage_s3_endpoint" yaml:"storage_s3_endpoint"`
	StorageS3Region          string `env:"STORAGE_S3_REGION" json:"storage_s3_region" yaml:"storage_s3_region"`
	StorageS3AccessKeyID     string `env:"STORAGE_S3_ACCESS_KEY_ID" json:"storage_s3_access_key_id" yaml:"storage_s3_access_key_id"`
	StorageS3AccessKeySecret string `env:"STORAGE_S3_ACCESS_KEY_SECRET" json:"storage_s3_access_key_secret" yaml:"storage_s3_access_key_secret"`
	StorageS3Token           string `env:"STORAGE_S3_TOKEN" json:"storage_s3_token" yaml:"storage_s3_token"`
	StorageGCSBucket         string `env:"STORAGE_GCS_BUCKET" json:"storage_gcs_bucket" yaml:"storage_gcs_bucket"`
	StorageR2AccountID       string `env:"STORAGE_R2_ACCOUNT_ID" json:"storage_r2_account_id" yaml:"storage_r2_account_id"`
	StorageR2AccessKeyID     string `env:"STORAGE_R2_ACCESS_KEY_ID" json:"storage_r2_access_key_id" yaml:"storage_r2_access_key_id"`
	StorageR2AccessKeySecret string `env:"STORAGE_R2_ACCESS_KEY_SECRET" json:"storage_r2_access_key_secret" yaml:"storage_r2_access_key_secret"`
	StorageR2Bucket          string `env:"STORAGE_R2_BUCKET" json:"storage_r2_bucket" yaml:"storage_r2_bucket"`

	// Queue
	QueueAsynqRedisAddress string         `env:"QUEUE_ASYNQ_REDIS_ADDRESS" json:"queue_asynq_redis_address" yaml:"queue_asynq_redis_address"`
	QueueAsynqConcurrency  int            `env:"QUEUE_ASYNQ_CONCURRENCY" json:"queue_asynq_concurrency" yaml:"queue_asynq_concurrency"`
	QueueAsynqQueues       map[string]int `json:"queue_asynq_queues" yaml:"queue_asynq_queues"`
}

// LoadConfig loads cfg from env vars, then yaml, then json — last wins.
// Pass config file paths as arguments (.yaml, .yml, or .json).
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
