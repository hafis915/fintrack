// Integration tests for the Phase 0 local-first auth routes:
//
//	POST /v1/auth/register
//	POST /v1/auth/login
//
// These are PUBLIC (they mint the JWT) so there are no missing/invalid-token
// cases here — the JWT middleware is covered by api_integration_test.go. We do
// assert that a token minted by register/login is actually accepted by a
// protected route (/v1/me and /v1/categories), closing the loop end-to-end.
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/test/testhelper"
)

func resetUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	// Child rows FK to users AND to expense_categories — clear them before the
	// custom categories and the users themselves, or the deletes trip a foreign
	// key. transactions/budget_items reference expense_categories, so they MUST
	// come before `delete from expense_categories` (a transaction logged against
	// a custom category would otherwise block it — see the inline-category flow).
	stmts := []string{
		`delete from budget_items`,
		`delete from budget_plans`,
		`delete from user_profiles`,
		`delete from transactions`,
		`delete from expense_categories where user_id is not null`,
		`delete from users`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(context.Background(), s); err != nil {
			t.Fatalf("resetting %q: %v", s, err)
		}
	}
}

func postJSON(t *testing.T, ts *testhelper.TestServer, path string, body any) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshalling body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	return rr, decode(t, rr)
}

func TestIntegration_AuthRegister_CreatesUserAndMintsWorkingToken(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	rr, env := postJSON(t, ts, "/v1/auth/register", map[string]any{
		"name":     "Hafis",
		"email":    "Hafis@Example.com",
		"password": "supersecret123",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Data == nil {
		t.Fatal("expected data envelope")
	}
	token, _ := env.Data["token"].(string)
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	// Email is normalized to lower-case on store.
	if got, _ := env.Data["email"].(string); got != "hafis@example.com" {
		t.Fatalf("want normalized email, got %q", got)
	}
	if got, _ := env.Data["name"].(string); got != "Hafis" {
		t.Fatalf("want name 'Hafis', got %q", got)
	}

	// The minted token must be accepted by a protected route.
	meReq := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRR := httptest.NewRecorder()
	ts.Echo.ServeHTTP(meRR, meReq)
	if meRR.Code != http.StatusOK {
		t.Fatalf("minted token rejected by /v1/me: %d, body=%s", meRR.Code, meRR.Body.String())
	}
	meEnv := decode(t, meRR)
	if got, _ := meEnv.Data["user_id"].(string); got != env.Data["user_id"].(string) {
		t.Fatalf("/v1/me user_id %q != register user_id %q", got, env.Data["user_id"])
	}

	// And the token must authenticate a data route guarded by JWTAuth +
	// EnsureUser — GET /v1/categories returns 200 with the seeded defaults.
	catReq := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	catReq.Header.Set("Authorization", "Bearer "+token)
	catRR := httptest.NewRecorder()
	ts.Echo.ServeHTTP(catRR, catReq)
	if catRR.Code != http.StatusOK {
		t.Fatalf("minted token rejected by /v1/categories: %d, body=%s", catRR.Code, catRR.Body.String())
	}
}

// Regression: EnsureUser upserts a <uuid>@local placeholder on the first /v1
// request for a JWT subject. The UpsertUser conflict path must only bump
// updated_at — it must NOT overwrite the real email set at register time.
func TestIntegration_Auth_EnsureUserDoesNotClobberRegisteredEmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	const email = "keepme@example.com"
	rr, env := postJSON(t, ts, "/v1/auth/register", map[string]any{
		"name":     "Hafis",
		"email":    email,
		"password": "supersecret123",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("register: want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	token, _ := env.Data["token"].(string)
	uid, err := uuid.Parse(env.Data["user_id"].(string))
	if err != nil {
		t.Fatalf("register user_id not a UUID: %v", err)
	}

	// Make a /v1 call as this user — this runs EnsureUser, which upserts the
	// <uuid>@local placeholder against the existing row.
	catReq := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	catReq.Header.Set("Authorization", "Bearer "+token)
	catRR := httptest.NewRecorder()
	ts.Echo.ServeHTTP(catRR, catReq)
	if catRR.Code != http.StatusOK {
		t.Fatalf("GET /v1/categories: want 200, got %d, body=%s", catRR.Code, catRR.Body.String())
	}

	// The stored email must still be the registered one, not <uuid>@local.
	var stored string
	if err := ts.DB.QueryRow(context.Background(),
		`select email from users where id = $1`, uid).Scan(&stored); err != nil {
		t.Fatalf("loading stored email: %v", err)
	}
	if stored != email {
		t.Errorf("EnsureUser clobbered registered email: want %q, got %q", email, stored)
	}
	if stored == uid.String()+"@local" {
		t.Errorf("email overwritten with <uuid>@local placeholder")
	}
}

func TestIntegration_AuthRegister_DuplicateEmailConflicts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	body := map[string]any{"name": "Hafis", "email": "dup@example.com", "password": "supersecret123"}
	if rr, _ := postJSON(t, ts, "/v1/auth/register", body); rr.Code != http.StatusCreated {
		t.Fatalf("first register: want 201, got %d", rr.Code)
	}

	// Same email, different case → still a conflict (case-insensitive unique).
	rr, env := postJSON(t, ts, "/v1/auth/register", map[string]any{
		"name": "Other", "email": "DUP@example.com", "password": "supersecret123",
	})
	if rr.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Error == nil || env.Error.Code != "email_taken" {
		t.Fatalf("want code 'email_taken', got %+v", env.Error)
	}
}

func TestIntegration_AuthRegister_RejectsInvalidInput(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	cases := map[string]map[string]any{
		"missing_name":  {"email": "x@example.com", "password": "supersecret123"},
		"blank_name":    {"name": "   ", "email": "x@example.com", "password": "supersecret123"},
		"missing_email": {"name": "Hafis", "password": "supersecret123"},
		"bad_email":     {"name": "Hafis", "email": "not-an-email", "password": "supersecret123"},
		"weak_password": {"name": "Hafis", "email": "weak@example.com", "password": "short"},
		"no_password":   {"name": "Hafis", "email": "nopw@example.com"},
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rr, env := postJSON(t, ts, "/v1/auth/register", body)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d, body=%s", rr.Code, rr.Body.String())
			}
			if env.Error == nil {
				t.Fatal("expected error envelope")
			}
		})
	}
}

func TestIntegration_AuthLogin_ReturnsTokenForKnownUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	reg := map[string]any{"name": "Hafis", "email": "login@example.com", "password": "supersecret123"}
	if rr, _ := postJSON(t, ts, "/v1/auth/register", reg); rr.Code != http.StatusCreated {
		t.Fatalf("register setup: want 201, got %d", rr.Code)
	}

	// Login with a different email case to prove lookup is case-insensitive.
	rr, env := postJSON(t, ts, "/v1/auth/login", map[string]any{
		"email": "Login@Example.com", "password": "supersecret123",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if token, _ := env.Data["token"].(string); token == "" {
		t.Fatal("expected non-empty token")
	}
	if got, _ := env.Data["name"].(string); got != "Hafis" {
		t.Fatalf("want name 'Hafis', got %q", got)
	}
}

func TestIntegration_AuthLogin_WrongPasswordUnauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	reg := map[string]any{"name": "Hafis", "email": "pw@example.com", "password": "supersecret123"}
	if rr, _ := postJSON(t, ts, "/v1/auth/register", reg); rr.Code != http.StatusCreated {
		t.Fatalf("register setup: want 201, got %d", rr.Code)
	}

	rr, env := postJSON(t, ts, "/v1/auth/login", map[string]any{
		"email": "pw@example.com", "password": "wrongpassword",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: want 401, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Error == nil || env.Error.Code != "invalid_credentials" {
		t.Fatalf("want code 'invalid_credentials', got %+v", env.Error)
	}
}

// An unknown email returns the SAME generic 401 as a wrong password — the
// endpoint must not reveal which emails are registered.
func TestIntegration_AuthLogin_UnknownEmailUnauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetUsers(t, ts.DB)

	rr, env := postJSON(t, ts, "/v1/auth/login", map[string]any{
		"email": "ghost@example.com", "password": "supersecret123",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Error == nil || env.Error.Code != "invalid_credentials" {
		t.Fatalf("want code 'invalid_credentials', got %+v", env.Error)
	}
}

func TestIntegration_Auth_RejectsMalformedJSON(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	for _, path := range []string{"/v1/auth/register", "/v1/auth/login"} {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(`{bad`)))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("%s: want 400 on malformed json, got %d", path, rr.Code)
		}
	}
}
