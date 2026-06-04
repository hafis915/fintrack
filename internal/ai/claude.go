package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	anthropicEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicVersion  = "2023-06-01"
	maxTokens         = 512

	// extractPrompt is intentionally strict: we want ONLY a JSON object so the
	// parser stays simple. The model is reminded the receipt is Indonesian and
	// that the amount must be an integer Rupiah value (matches BIGINT money).
	extractPrompt = `This is an Indonesian receipt. Extract the total amount in Rupiah as an integer with no decimals or separators, the merchant name, and a short spending category. Respond with ONLY a JSON object, no prose: {"amount": <int>, "merchant": <string>, "category": <string>, "confidence": <float 0..1>}`
)

// claudeAnalyzer calls the Anthropic Messages API over plain net/http (no SDK,
// per project constraints). It is safe for concurrent use — all per-call state
// lives in AnalyzeReceipt.
type claudeAnalyzer struct {
	apiKey string
	model  string
	hc     *http.Client
}

// NewClaudeAnalyzer builds a ReceiptAnalyzer backed by Claude Vision. If hc is
// nil, a client with a 30s timeout is used — http.DefaultClient has NO timeout,
// so a hung Anthropic connection would block the request goroutine forever.
func NewClaudeAnalyzer(apiKey, model string, hc *http.Client) ReceiptAnalyzer {
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	return &claudeAnalyzer{apiKey: apiKey, model: model, hc: hc}
}

// --- request shapes ---

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type   string           `json:"type"`
	Text   string           `json:"text,omitempty"`
	Source *anthropicSource `json:"source,omitempty"`
}

type anthropicSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// --- response shapes ---

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// draftJSON mirrors the JSON object the prompt asks the model to emit.
type draftJSON struct {
	Amount     int64   `json:"amount"`
	Merchant   string  `json:"merchant"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

func (a *claudeAnalyzer) AnalyzeReceipt(ctx context.Context, image []byte, mimeType string) (ReceiptDraft, error) {
	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: maxTokens,
		Messages: []anthropicMessage{
			{
				Role: "user",
				Content: []anthropicContent{
					{
						Type: "image",
						Source: &anthropicSource{
							Type:      "base64",
							MediaType: mimeType,
							Data:      base64.StdEncoding.EncodeToString(image),
						},
					},
					{Type: "text", Text: extractPrompt},
				},
			},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return ReceiptDraft{}, fmt.Errorf("marshaling request: %w: %w", err, ErrAnalyzeFailed)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(payload))
	if err != nil {
		return ReceiptDraft{}, fmt.Errorf("building request: %w: %w", err, ErrAnalyzeFailed)
	}
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := a.hc.Do(req)
	if err != nil {
		return ReceiptDraft{}, fmt.Errorf("calling anthropic: %w: %w", err, ErrAnalyzeFailed)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReceiptDraft{}, fmt.Errorf("reading response: %w: %w", err, ErrAnalyzeFailed)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ReceiptDraft{}, fmt.Errorf("anthropic status %d: %s: %w", resp.StatusCode, strings.TrimSpace(string(body)), ErrAnalyzeFailed)
	}

	var ar anthropicResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return ReceiptDraft{}, fmt.Errorf("decoding anthropic envelope: %w: %w", err, ErrAnalyzeFailed)
	}
	if len(ar.Content) == 0 {
		return ReceiptDraft{}, fmt.Errorf("anthropic returned no content: %w", ErrAnalyzeFailed)
	}

	raw := extractJSONObject(ar.Content[0].Text)
	if raw == "" {
		return ReceiptDraft{}, fmt.Errorf("no JSON object in model output: %w", ErrAnalyzeFailed)
	}

	var d draftJSON
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return ReceiptDraft{}, fmt.Errorf("parsing model JSON: %w: %w", err, ErrAnalyzeFailed)
	}

	return ReceiptDraft{
		Amount:       d.Amount,
		Merchant:     d.Merchant,
		CategoryHint: d.Category,
		Confidence:   d.Confidence,
	}, nil
}

// extractJSONObject pulls the first balanced {...} object out of model output.
// It tolerates Markdown code fences and surrounding prose by scanning for the
// first '{' and the matching closing '}' (string-aware so braces inside string
// literals don't throw off the depth counter). Returns "" if none found.
func extractJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}

	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
