package integration_test

import (
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
