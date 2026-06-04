package storage

import (
	"context"
	"sync"
	"time"
)

// StubStorage is an in-memory Storage implementation for integration tests, so
// they can exercise the full receipt flow without a running MinIO container.
// It is safe for concurrent use.
type StubStorage struct {
	mu      sync.Mutex
	objects map[string][]byte
}

// NewStubStorage returns an empty in-memory Storage.
func NewStubStorage() Storage {
	return &StubStorage{objects: make(map[string][]byte)}
}

// Upload copies data into the in-memory map and returns a stub:// URL.
func (s *StubStorage) Upload(_ context.Context, key, _ string, data []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Copy so the caller can't mutate stored bytes through the original slice.
	stored := make([]byte, len(data))
	copy(stored, data)
	s.objects[key] = stored
	return "stub://" + key, nil
}

// SignedURL returns the same stub:// URL Upload produced; there is nothing to
// sign in memory.
func (s *StubStorage) SignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "stub://" + key, nil
}

// Delete removes the object at key. Deleting a missing key is a no-op.
func (s *StubStorage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}
