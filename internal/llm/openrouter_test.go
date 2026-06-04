package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// capturedReq holds the parts of the outbound request we want to assert on.
type capturedReq struct {
	authorization string
	contentType   string
	body          openRouterRequest
}

// newTestServer returns an httptest.Server that records the request and replies
// with respBody/status, plus a capturedReq pointer populated on each call.
func newTestServer(t *testing.T, status int, respBody string) (*httptest.Server, *capturedReq) {
	t.Helper()
	cap := &capturedReq{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.authorization = r.Header.Get("Authorization")
		cap.contentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &cap.body)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

// rewriteTransport redirects every request to the test server, so we can use
// the real https://openrouter.ai endpoint constant unchanged.
type rewriteTransport struct{ base string }

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := http.NewRequestWithContext(req.Context(), req.Method, rt.base, req.Body)
	if err != nil {
		return nil, err
	}
	target.Header = req.Header
	return http.DefaultClient.Do(target)
}

func TestOpenRouterClient_Complete(t *testing.T) {
	const replyJSON = `{"reply":"Oke","adjustments":[]}`

	tests := []struct {
		name     string
		status   int
		respBody string
		wantErr  bool
		want     string
	}{
		{
			name:     "happy path",
			status:   http.StatusOK,
			respBody: `{"choices":[{"message":{"role":"assistant","content":` + jsonString(replyJSON) + `}}]}`,
			want:     replyJSON,
		},
		{
			name:     "non-2xx -> ErrLLMFailed",
			status:   http.StatusUnauthorized,
			respBody: `{"error":"bad key"}`,
			wantErr:  true,
		},
		{
			name:     "empty choices -> ErrLLMFailed",
			status:   http.StatusOK,
			respBody: `{"choices":[]}`,
			wantErr:  true,
		},
		{
			name:     "garbage envelope -> ErrLLMFailed",
			status:   http.StatusOK,
			respBody: `not json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, cap := newTestServer(t, tc.status, tc.respBody)
			hc := &http.Client{Transport: rewriteTransport{base: srv.URL}}

			c := NewOpenRouterClient("test-key", "test/model", hc)
			msgs := []Message{{Role: "user", Content: "naikin makan jadi 1500000"}}
			got, err := c.Complete(context.Background(), "you are a planner", msgs)

			if tc.wantErr {
				if !errors.Is(err, ErrLLMFailed) {
					t.Fatalf("want ErrLLMFailed, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("content mismatch: got %q want %q", got, tc.want)
			}

			// Assert request shape on the happy path.
			if cap.authorization != "Bearer test-key" {
				t.Errorf("Authorization = %q want Bearer test-key", cap.authorization)
			}
			if cap.contentType != "application/json" {
				t.Errorf("Content-Type = %q want application/json", cap.contentType)
			}
			if cap.body.Model != "test/model" {
				t.Errorf("model = %q want test/model", cap.body.Model)
			}
			// system prompt must be prepended, then the user thread.
			if len(cap.body.Messages) != 2 {
				t.Fatalf("want 2 messages (system + user), got %d: %+v", len(cap.body.Messages), cap.body.Messages)
			}
			if cap.body.Messages[0].Role != "system" || cap.body.Messages[0].Content != "you are a planner" {
				t.Errorf("first message not the system prompt: %+v", cap.body.Messages[0])
			}
			if cap.body.Messages[1].Role != "user" || cap.body.Messages[1].Content != "naikin makan jadi 1500000" {
				t.Errorf("second message not the user turn: %+v", cap.body.Messages[1])
			}
		})
	}
}

func TestOpenRouterClient_NilClientUsesDefault(t *testing.T) {
	c := NewOpenRouterClient("k", "m", nil)
	if c == nil {
		t.Fatal("expected client")
	}
}

func TestStubClient_Complete(t *testing.T) {
	tests := []struct {
		name       string
		messages   []Message
		wantCat    string
		wantAmt    int64
		wantAdjust bool
	}{
		{
			name:       "naikin makan jadi 1500000",
			messages:   []Message{{Role: "user", Content: "naikin makan jadi 1500000"}},
			wantCat:    "makan",
			wantAmt:    1_500_000,
			wantAdjust: true,
		},
		{
			name:       "juta suffix decimal: belanja jadi 1,5jt",
			messages:   []Message{{Role: "user", Content: "set belanja jadi 1,5jt dong"}},
			wantCat:    "belanja",
			wantAmt:    1_500_000,
			wantAdjust: true,
		},
		{
			name:       "ribu suffix: transportasi 750rb",
			messages:   []Message{{Role: "user", Content: "transportasi 750rb cukup"}},
			wantCat:    "transportasi",
			wantAmt:    750_000,
			wantAdjust: true,
		},
		{
			name:       "thousands separators: hiburan jadi 300.000",
			messages:   []Message{{Role: "user", Content: "hiburan jadi 300.000"}},
			wantCat:    "hiburan",
			wantAmt:    300_000,
			wantAdjust: true,
		},
		{
			name:       "longer name wins: makan & minum jadi 2000000",
			messages:   []Message{{Role: "user", Content: "makan & minum jadi 2000000"}},
			wantCat:    "makan & minum",
			wantAmt:    2_000_000,
			wantAdjust: true,
		},
		{
			name:       "no category -> clarifying question, no adjustments",
			messages:   []Message{{Role: "user", Content: "naikin jadi 1500000"}},
			wantAdjust: false,
		},
		{
			name:       "no amount -> clarifying question, no adjustments",
			messages:   []Message{{Role: "user", Content: "ubah makan dong"}},
			wantAdjust: false,
		},
		{
			name:       "uses LAST user message",
			messages:   []Message{{Role: "user", Content: "halo"}, {Role: "assistant", Content: "hai"}, {Role: "user", Content: "naikin kesehatan jadi 500000"}},
			wantCat:    "kesehatan",
			wantAmt:    500_000,
			wantAdjust: true,
		},
	}

	c := NewStubClient()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := c.Complete(context.Background(), "ignored", tc.messages)
			if err != nil {
				t.Fatalf("stub returned error: %v", err)
			}

			var got stubReply
			if err := json.Unmarshal([]byte(raw), &got); err != nil {
				t.Fatalf("stub output is not valid JSON: %v\nraw=%s", err, raw)
			}
			if got.Reply == "" {
				t.Errorf("reply must not be empty")
			}

			if !tc.wantAdjust {
				if len(got.Adjustments) != 0 {
					t.Fatalf("want no adjustments, got %+v", got.Adjustments)
				}
				return
			}

			if len(got.Adjustments) != 1 {
				t.Fatalf("want 1 adjustment, got %+v", got.Adjustments)
			}
			adj := got.Adjustments[0]
			if adj.CategoryName != tc.wantCat {
				t.Errorf("category_name = %q want %q", adj.CategoryName, tc.wantCat)
			}
			if adj.TargetAmount != tc.wantAmt {
				t.Errorf("target_amount = %d want %d", adj.TargetAmount, tc.wantAmt)
			}
		})
	}
}

// jsonString JSON-encodes s as a quoted string literal for embedding in a
// response body template.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
