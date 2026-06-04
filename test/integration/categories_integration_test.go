package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/test/testhelper"
)

type categoryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"is_default"`
}

type categoryListEnvelope struct {
	Data  []categoryItem `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func TestIntegration_Categories_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

type categoryOneEnvelope struct {
	Data  *categoryItem `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func postCategory(t *testing.T, ts *testhelper.TestServer, uid uuid.UUID, body any) (*httptest.ResponseRecorder, categoryOneEnvelope) {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	var env categoryOneEnvelope
	_ = json.Unmarshal(rr.Body.Bytes(), &env)
	return rr, env
}

func TestIntegration_Categories_Create_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/categories",
		bytes.NewReader([]byte(`{"name":"X","type":"variable"}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestIntegration_Categories_Create_AddsCustomAndListsIt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	uid := uuid.New()

	rr, env := postCategory(t, ts, uid, map[string]any{"name": "Kursus online", "type": "want"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Data == nil || env.Data.Name != "Kursus online" || env.Data.Type != "want" {
		t.Fatalf("unexpected created category: %+v", env.Data)
	}
	if env.Data.IsDefault {
		t.Error("custom category must not be is_default")
	}
	newID := env.Data.ID

	// It now appears in this user's category list.
	listReq := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	listReq.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	listRR := httptest.NewRecorder()
	ts.Echo.ServeHTTP(listRR, listReq)
	var list categoryListEnvelope
	_ = json.Unmarshal(listRR.Body.Bytes(), &list)
	found := false
	for _, c := range list.Data {
		if c.ID == newID {
			found = true
		}
	}
	if !found {
		t.Error("created custom category not returned by GET /v1/categories")
	}
}

func TestIntegration_Categories_Create_RejectsBadInput(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	uid := uuid.New()

	cases := map[string]map[string]any{
		"empty_name":   {"name": "  ", "type": "want"},
		"bad_type":     {"name": "X", "type": "luxury"},
		"missing_type": {"name": "X"},
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rr, env := postCategory(t, ts, uid, body)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d, body=%s", rr.Code, rr.Body.String())
			}
			if env.Error == nil {
				t.Error("want an error body, got none")
			}
		})
	}
}

func TestIntegration_Categories_ListsSystemDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	uid := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	var env categoryListEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if len(env.Data) == 0 {
		t.Fatal("expected default categories from migration seed, got empty list")
	}

	// Sanity: every returned row is a system default for this fresh user
	// (we deleted custom ones in reset). At least one of each type appears.
	types := map[string]bool{}
	for _, c := range env.Data {
		if !c.IsDefault {
			t.Errorf("unexpected non-default category in fresh-user list: %s", c.Name)
		}
		types[c.Type] = true
	}
	for _, want := range []string{"fixed", "variable", "debt", "want"} {
		if !types[want] {
			t.Errorf("missing category of type %q in default seed", want)
		}
	}
}
