// Package storage abstracts object storage for receipt images behind a small
// interface so handlers and the repository layer never depend on a concrete
// S3/MinIO client. Two implementations exist:
//
//   - MinIOStorage: real S3-compatible storage (MinIO locally, Supabase
//     Storage in prod). Constructed from main.go via NewMinIOStorage.
//   - StubStorage: an in-memory map used by integration tests so they don't
//     require a running MinIO container.
//
// Object keys follow the convention `receipts/{user_id}/{txn_id}.jpg`; the
// storage layer is key-agnostic and stores whatever key it is handed.
package storage

import (
	"context"
	"time"
)

// Storage is the contract every backing store must satisfy. It is safe for
// concurrent use by multiple goroutines.
type Storage interface {
	// Upload writes data under key with the given content type and returns a
	// stable object URL. It overwrites any existing object at key.
	Upload(ctx context.Context, key, contentType string, data []byte) (url string, err error)
	// SignedURL returns a time-limited URL granting read access to key.
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	// Delete removes the object at key. Deleting a missing key is not an error.
	Delete(ctx context.Context, key string) error
}

// MinIOConfig carries the connection settings for a real S3-compatible
// backend. Populated by the config layer from environment variables.
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}
