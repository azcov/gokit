package r2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	gokits3 "github.com/azcov/gokit/storage/s3"
)

type Config struct {
	AccountID       string `config:"account_id"`
	AccessKeyID     string `config:"access_key_id"`
	AccessKeySecret string `config:"access_key_secret"`
	Bucket          string `config:"bucket"`
}

// New returns an S3-compatible client pointed at Cloudflare R2.
func New(ctx context.Context, cfg Config) (*gokits3.S3, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID, cfg.AccessKeySecret, "",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}
	return gokits3.New(gokits3.Config{
		AWSConfig: awsCfg,
		Bucket:    cfg.Bucket,
		Endpoint:  fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID),
	}), nil
}
