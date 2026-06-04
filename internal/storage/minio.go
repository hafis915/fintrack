package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage is the S3-compatible Storage implementation backed by minio-go.
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOStorage constructs a MinIOStorage and ensures the configured bucket
// exists, creating it on first run. The returned Storage is safe for
// concurrent use — the underlying minio.Client is goroutine-safe.
func NewMinIOStorage(cfg MinIOConfig) (Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	// Ensure the bucket exists. We do this once at init so Upload never has to
	// branch on bucket existence on the hot path. Bounded so an unreachable
	// MinIO fails fast instead of hanging server startup forever.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("checking bucket %q: %w", cfg.Bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("creating bucket %q: %w", cfg.Bucket, err)
		}
	}

	return &MinIOStorage{client: client, bucket: cfg.Bucket}, nil
}

// Upload writes data under key and returns a plain object URL. The URL is not
// pre-signed — callers needing read access for clients should use SignedURL.
func (s *MinIOStorage) Upload(ctx context.Context, key, contentType string, data []byte) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("uploading object %q: %w", key, err)
	}

	endpoint := s.client.EndpointURL()
	url := fmt.Sprintf("%s/%s/%s", endpoint.String(), s.bucket, key)
	return url, nil
}

// SignedURL returns a pre-signed GET URL valid for ttl.
func (s *MinIOStorage) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("signing url for object %q: %w", key, err)
	}
	return u.String(), nil
}

// Delete removes the object at key.
func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("deleting object %q: %w", key, err)
	}
	return nil
}
