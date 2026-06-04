package ai

import "context"

// stubAnalyzer returns a fixed ReceiptDraft regardless of input. Used by tests
// and local dev runs where no ANTHROPIC_API_KEY is configured. The returned
// values are asserted on by tests across slices, so they must not change.
type stubAnalyzer struct{}

// NewStubAnalyzer returns a ReceiptAnalyzer that deterministically yields the
// same draft on every call. It never errors.
func NewStubAnalyzer() ReceiptAnalyzer {
	return stubAnalyzer{}
}

func (stubAnalyzer) AnalyzeReceipt(_ context.Context, _ []byte, _ string) (ReceiptDraft, error) {
	return ReceiptDraft{
		Amount:       50000,
		Merchant:     "Indomaret",
		CategoryHint: "Belanja",
		Confidence:   0.95,
	}, nil
}
