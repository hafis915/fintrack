package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/test/testhelper"
)

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), out); err != nil {
		t.Fatalf("decoding body %q: %v", rr.Body.String(), err)
	}
}

// Cross-user isolation is the security boundary that keeps one user's financial
// data away from another's. The database has RLS *enabled* but no policies, and
// the app connects as the table owner — so RLS is currently a no-op and the
// isolation is enforced ENTIRELY at the app layer (every query filters
// user_id; the ...ForUser query naming convention). This test is the regression
// guard for that boundary: if a future query drops its user_id scope, this fails
// loudly instead of silently leaking another user's data.
//
// See the CSO audit (2026-06-05) and the `rls-enabled-but-noop` learning.
func TestIntegration_CrossUserIsolation_Transactions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	alice := uuid.New()
	bob := uuid.New()
	seedUser(t, ts.DB, alice, fmt.Sprintf("%s@local", alice))
	seedUser(t, ts.DB, bob, fmt.Sprintf("%s@local", bob))

	// Alice logs a transaction. Bob must never be able to see or touch it.
	txA := makeTx(t, ts, alice, cats["Makan & minum"], 75_000, "alice lunch")

	serve := func(req *http.Request) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		return rr
	}

	// Bob is authenticated as himself — a valid token, just the wrong user. Each
	// by-id route must behave as if Alice's transaction does not exist (404),
	// never 200-with-Alice's-data and never a 403 that confirms it exists.
	t.Run("bob cannot GET alice's transaction", func(t *testing.T) {
		rr := serve(authReq(t, bob, http.MethodGet, "/v1/transactions/"+txA.ID, nil))
		if rr.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d, body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("bob cannot PATCH alice's transaction", func(t *testing.T) {
		body := map[string]any{"amount": 1, "note": "hijacked"}
		rr := serve(authReq(t, bob, http.MethodPatch, "/v1/transactions/"+txA.ID, body))
		if rr.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d, body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("bob cannot DELETE alice's transaction", func(t *testing.T) {
		rr := serve(authReq(t, bob, http.MethodDelete, "/v1/transactions/"+txA.ID, nil))
		if rr.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d, body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("bob's list does not include alice's transaction", func(t *testing.T) {
		rr := serve(authReq(t, bob, http.MethodGet, "/v1/transactions", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
		}
		var env txListEnvelope
		decodeBody(t, rr, &env)
		if env.Data == nil {
			t.Fatal("list returned no data")
		}
		if env.Data.Total != 0 || len(env.Data.Items) != 0 {
			t.Errorf("bob sees alice's data: total=%d, items=%d", env.Data.Total, len(env.Data.Items))
		}
	})

	// Control: Alice still owns an intact transaction — proving the 404s above
	// are isolation, not a broken route, and that Bob's PATCH/DELETE were no-ops.
	t.Run("alice still sees her own intact transaction", func(t *testing.T) {
		rr := serve(authReq(t, alice, http.MethodGet, "/v1/transactions/"+txA.ID, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
		}
		var env txEnvelope
		decodeBody(t, rr, &env)
		if env.Data == nil || env.Data.Amount != 75_000 || env.Data.Note != "alice lunch" {
			t.Errorf("alice's transaction was altered: %+v", env.Data)
		}
	})
}
