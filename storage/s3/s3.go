package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/azcov/gokit/storage"
)

var _ storage.Storage = (*S3)(nil)

type S3 struct {
	client *awss3.Client
	bucket string
}

type Config struct {
	AWSConfig aws.Config `json:"-" yaml:"-"`
	Bucket    string     `env:"BUCKET" json:"bucket" yaml:"bucket"`
	// Endpoint overrides the default endpoint; used for S3-compatible APIs (e.g. R2).
	Endpoint string `env:"ENDPOINT" json:"endpoint" yaml:"endpoint"`
}

func New(cfg Config) *S3 {
	opts := []func(*awss3.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	}
	return &S3{
		client: awss3.NewFromConfig(cfg.AWSConfig, opts...),
		bucket: cfg.Bucket,
	}
}

func (s *S3) Put(ctx context.Context, key string, r io.Reader, opts storage.PutOptions) error {
	_, err := s.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(opts.ContentType),
	})
	return err
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3) URL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key), nil
}

func (s *S3) List(ctx context.Context, prefix string) ([]storage.Object, error) {
	out, err := s.client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	objects := make([]storage.Object, 0, len(out.Contents))
	for _, obj := range out.Contents {
		objects = append(objects, storage.Object{
			Key:  aws.ToString(obj.Key),
			Size: aws.ToInt64(obj.Size),
		})
	}
	return objects, nil
}

func (s *S3) Close() error { return nil }
