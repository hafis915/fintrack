package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// captured holds the parts of the outbound request we want to assert on.
type captured struct {
	apiKey      string
	version     string
	contentType string
	body        anthropicRequest
}

// newTestServer returns an httptest.Server that records the request and replies
// with respBody/status, plus a captured pointer populated on each call.
func newTestServer(t *testing.T, status int, respBody string) (*httptest.Server, *captured) {
	t.Helper()
	cap := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.apiKey = r.Header.Get("x-api-key")
		cap.version = r.Header.Get("anthropic-version")
		cap.contentType = r.Header.Get("content-type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &cap.body)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

// rewriteTransport redirects every request to the test server, so we can use
// the real https://api.anthropic.com endpoint constant unchanged.
type rewriteTransport struct{ base string }

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := http.NewRequestWithContext(req.Context(), req.Method, rt.base, req.Body)
	if err != nil {
		return nil, err
	}
	target.Header = req.Header
	return http.DefaultClient.Do(target)
}

func TestClaudeAnalyzer_AnalyzeReceipt(t *testing.T) {
	const goodInner = `{"amount": 73500, "merchant": "Alfamart", "category": "Belanja", "confidence": 0.88}`

	tests := []struct {
		name      string
		status    int
		respBody  string
		wantErr   bool
		wantDraft ReceiptDraft
	}{
		{
			name:     "happy path bare json",
			status:   http.StatusOK,
			respBody: `{"content":[{"type":"text","text":` + jsonString(goodInner) + `}]}`,
			wantDraft: ReceiptDraft{
				Amount: 73500, Merchant: "Alfamart", CategoryHint: "Belanja", Confidence: 0.88,
			},
		},
		{
			name:     "happy path with code fences and prose",
			status:   http.StatusOK,
			respBody: `{"content":[{"type":"text","text":` + jsonString("Here you go:\n```json\n"+goodInner+"\n```\nthanks") + `}]}`,
			wantDraft: ReceiptDraft{
				Amount: 73500, Merchant: "Alfamart", CategoryHint: "Belanja", Confidence: 0.88,
			},
		},
		{
			name:     "malformed inner json -> ErrAnalyzeFailed",
			status:   http.StatusOK,
			respBody: `{"content":[{"type":"text","text":"not json at all"}]}`,
			wantErr:  true,
		},
		{
			name:     "non-2xx -> ErrAnalyzeFailed",
			status:   http.StatusUnauthorized,
			respBody: `{"error":"bad key"}`,
			wantErr:  true,
		},
		{
			name:     "empty content -> ErrAnalyzeFailed",
			status:   http.StatusOK,
			respBody: `{"content":[]}`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, cap := newTestServer(t, tc.status, tc.respBody)
			hc := &http.Client{Transport: rewriteTransport{base: srv.URL}}

			a := NewClaudeAnalyzer("test-key", "claude-haiku-4-5-20251001", hc)
			img := []byte{0x01, 0x02, 0x03}
			got, err := a.AnalyzeReceipt(context.Background(), img, "image/jpeg")

			if tc.wantErr {
				if !errors.Is(err, ErrAnalyzeFailed) {
					t.Fatalf("want ErrAnalyzeFailed, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantDraft {
				t.Fatalf("draft mismatch: got %+v want %+v", got, tc.wantDraft)
			}

			// Assert request shape on the happy path.
			if cap.apiKey != "test-key" {
				t.Errorf("x-api-key = %q want test-key", cap.apiKey)
			}
			if cap.version != "2023-06-01" {
				t.Errorf("anthropic-version = %q want 2023-06-01", cap.version)
			}
			if cap.contentType != "application/json" {
				t.Errorf("content-type = %q want application/json", cap.contentType)
			}
			if cap.body.Model != "claude-haiku-4-5-20251001" {
				t.Errorf("model = %q", cap.body.Model)
			}
			if cap.body.MaxTokens != 512 {
				t.Errorf("max_tokens = %d want 512", cap.body.MaxTokens)
			}
			if len(cap.body.Messages) != 1 || len(cap.body.Messages[0].Content) != 2 {
				t.Fatalf("unexpected message structure: %+v", cap.body.Messages)
			}
			imgPart := cap.body.Messages[0].Content[0]
			if imgPart.Type != "image" || imgPart.Source == nil {
				t.Fatalf("first content part is not an image: %+v", imgPart)
			}
			if imgPart.Source.Type != "base64" || imgPart.Source.MediaType != "image/jpeg" {
				t.Errorf("bad image source: %+v", imgPart.Source)
			}
			if imgPart.Source.Data != base64.StdEncoding.EncodeToString(img) {
				t.Errorf("image data not base64 of input")
			}
			textPart := cap.body.Messages[0].Content[1]
			if textPart.Type != "text" || textPart.Text == "" {
				t.Errorf("second content part is not text prompt: %+v", textPart)
			}
		})
	}
}

func TestClaudeAnalyzer_NilClientUsesDefault(t *testing.T) {
	a := NewClaudeAnalyzer("k", "m", nil)
	if a == nil {
		t.Fatal("expected analyzer")
	}
}

func TestStubAnalyzer(t *testing.T) {
	a := NewStubAnalyzer()
	got, err := a.AnalyzeReceipt(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("stub returned error: %v", err)
	}
	want := ReceiptDraft{Amount: 50000, Merchant: "Indomaret", CategoryHint: "Belanja", Confidence: 0.95}
	if got != want {
		t.Fatalf("stub draft mismatch: got %+v want %+v", got, want)
	}
}

// jsonString JSON-encodes s as a quoted string literal for embedding in a
// response body template.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
