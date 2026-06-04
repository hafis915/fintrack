// Package ai provides receipt analysis backed by Claude Vision. The package
// exposes a small interface so callers (handlers/repository wiring) depend on
// behaviour, not on the HTTP transport. A deterministic stub implementation is
// provided for tests and local dev without an API key.
package ai

import (
	"context"
	"errors"
)

// ReceiptDraft is the structured result of analysing a receipt image. Amount is
// Rupiah as a whole integer (no decimals/separators) to match the BIGINT money
// convention used across the codebase. Confidence is the model's self-reported
// confidence in the range [0,1].
type ReceiptDraft struct {
	Amount       int64
	Merchant     string
	CategoryHint string
	Confidence   float64
}

// ReceiptAnalyzer turns a raw image into a ReceiptDraft. Implementations must
// respect ctx cancellation and return a wrapped ErrAnalyzeFailed on any failure
// so callers can branch with errors.Is.
type ReceiptAnalyzer interface {
	AnalyzeReceipt(ctx context.Context, image []byte, mimeType string) (ReceiptDraft, error)
}

// ErrAnalyzeFailed is the sentinel wrapped by every analyzer failure path
// (network error, non-2xx response, unparseable JSON, etc.). Callers use
// errors.Is(err, ai.ErrAnalyzeFailed) rather than matching on transport detail.
var ErrAnalyzeFailed = errors.New("receipt analysis failed")
