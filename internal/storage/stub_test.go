package storage

import (
	"context"
	"testing"
	"time"
)

func TestStubStorage_UploadThenSignedURL(t *testing.T) {
	ctx := context.Background()
	s := NewStubStorage()

	const key = "receipts/user-1/txn-1.jpg"
	url, err := s.Upload(ctx, key, "image/jpeg", []byte("fake-bytes"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if want := "stub://" + key; url != want {
		t.Fatalf("Upload url = %q, want %q", url, want)
	}

	signed, err := s.SignedURL(ctx, key, 15*time.Minute)
	if err != nil {
		t.Fatalf("SignedURL: %v", err)
	}
	if want := "stub://" + key; signed != want {
		t.Fatalf("SignedURL = %q, want %q", signed, want)
	}
}

func TestStubStorage_Delete(t *testing.T) {
	ctx := context.Background()
	s := NewStubStorage().(*StubStorage)

	const key = "receipts/user-1/txn-2.jpg"
	if _, err := s.Upload(ctx, key, "image/jpeg", []byte("data")); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if _, ok := s.objects[key]; !ok {
		t.Fatal("object not stored after Upload")
	}

	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := s.objects[key]; ok {
		t.Fatal("object still present after Delete")
	}

	// Deleting a missing key must be a no-op, not an error.
	if err := s.Delete(ctx, "receipts/user-1/missing.jpg"); err != nil {
		t.Fatalf("Delete missing key: %v", err)
	}
}

func TestStubStorage_UploadCopiesData(t *testing.T) {
	ctx := context.Background()
	s := NewStubStorage().(*StubStorage)

	const key = "receipts/user-1/txn-3.jpg"
	data := []byte{1, 2, 3}
	if _, err := s.Upload(ctx, key, "image/jpeg", data); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	// Mutating the caller's slice must not affect stored bytes.
	data[0] = 99
	if got := s.objects[key][0]; got != 1 {
		t.Fatalf("stored bytes mutated through caller slice: got %d, want 1", got)
	}
}
