// Phase 0 API integration tests.
//
// What's covered here:
//   - /health      → 200, returns DB status
//   - /v1/me       → 401 missing / malformed / wrong-secret / expired,
//                    200 valid token returns the sub from the JWT
//   - Request ID middleware echoes the X-Request-ID header
//
// Tests run against fintrack_test (real Postgres). Run via:
//   make test           (all)
//   make test-integration
//
// The "_integration_test.go" suffix is convention; the package is also
// gated by the standard testing.Short() check so `go test -short ./...`
// skips us if the DB isn't available.
package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/test/testhelper"
)

type envelope struct {
	Data  map[string]any `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func decode(t *testing.T, rr *httptest.ResponseRecorder) envelope {
	t.Helper()
	var e envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &e); err != nil {
		t.Fatalf("decoding response body %q: %v", rr.Body.String(), err)
	}
	return e
}

func TestIntegration_Health_ReturnsOK(t *testing.T) {
	if testing.Short() {
		t.Skip("-short flag; skipping integration test")
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	e := decode(t, rr)
	if got := e.Data["status"]; got != "ok" {
		t.Errorf("status field: want 'ok', got %v", got)
	}
	if got := e.Data["db"]; got != "ok" {
		t.Errorf("db field: want 'ok', got %v (db unreachable?)", got)
	}
	if rr.Header().Get("X-Request-Id") == "" {
		t.Error("X-Request-Id header missing — RequestID middleware not wired")
	}
}

func TestIntegration_Health_RequestIDPassthrough(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	const incoming = "trace-12345"
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-Id", incoming)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Request-Id"); got != incoming {
		t.Errorf("X-Request-Id passthrough: want %q, got %q", incoming, got)
	}
}

func TestIntegration_Me_RejectsMissingAuthHeader(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", rr.Code)
	}
	e := decode(t, rr)
	if e.Error == nil || e.Error.Code != "missing_token" {
		t.Errorf("error code: want 'missing_token', got %+v", e.Error)
	}
}

func TestIntegration_Me_RejectsMalformedAuthHeader(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	cases := []struct {
		name   string
		header string
	}{
		{"no_bearer_prefix", "abc.def.ghi"},
		{"wrong_scheme", "Basic abc.def.ghi"},
		{"too_few_parts", "Bearer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
			req.Header.Set("Authorization", tc.header)
			rr := httptest.NewRecorder()
			ts.Echo.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("want 401, got %d, body=%s", rr.Code, rr.Body.String())
			}
			e := decode(t, rr)
			if e.Error == nil || !strings.HasSuffix(e.Error.Code, "_token") {
				t.Errorf("expected an *_token error code, got %+v", e.Error)
			}
		})
	}
}

func TestIntegration_Me_RejectsTokenSignedByOtherSecret(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	tok := testhelper.MintTokenWithSecret(t, uuid.New(), "definitely-not-the-server-secret")

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", rr.Code, rr.Body.String())
	}
	e := decode(t, rr)
	if e.Error == nil || e.Error.Code != "invalid_token" {
		t.Errorf("want code 'invalid_token', got %+v", e.Error)
	}
}

func TestIntegration_Me_RejectsExpiredToken(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	tok := testhelper.MintExpiredToken(t, uuid.New())

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", rr.Code, rr.Body.String())
	}
	e := decode(t, rr)
	if e.Error == nil || e.Error.Code != "invalid_token" {
		t.Errorf("want code 'invalid_token', got %+v", e.Error)
	}
}

func TestIntegration_Me_RejectsNonUUIDSubject(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	// Hand-craft a token whose `sub` claim is not a UUID. We sign with the
	// real test secret so signature passes; the failure must come from the
	// subject-parse step in the JWT middleware.
	rawTok := mintNonUUIDSubToken(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+rawTok)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
	e := decode(t, rr)
	if e.Error == nil || e.Error.Code != "invalid_subject" {
		t.Errorf("want 'invalid_subject', got %+v", e.Error)
	}
}

func TestIntegration_Me_AcceptsValidTokenAndReturnsSub(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	userID := uuid.New()
	tok := testhelper.MintTokenForUser(t, userID)

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	e := decode(t, rr)
	if got := e.Data["user_id"]; got != userID.String() {
		t.Errorf("user_id: want %q, got %v", userID.String(), got)
	}
}
