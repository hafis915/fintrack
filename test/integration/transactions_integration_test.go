package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/test/testhelper"
)

// --- helpers ------------------------------------------------------------

func resetTransactionTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	stmts := []string{
		`delete from transactions`,
		`delete from budget_items`,
		`delete from budget_plans`,
		`delete from user_profiles`,
		`delete from expense_categories where user_id is not null`,
		`delete from users`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(context.Background(), s); err != nil {
			t.Fatalf("resetting %q: %v", s, err)
		}
	}
}

func seedUser(t *testing.T, pool *pgxpool.Pool, id uuid.UUID, email string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`insert into users (id, email) values ($1, $2) on conflict (id) do nothing`, id, email)
	if err != nil {
		t.Fatalf("seeding user: %v", err)
	}
}

type txItem struct {
	ID            string `json:"id"`
	CategoryID    string `json:"category_id"`
	CategoryName  string `json:"category_name"`
	Amount        int64  `json:"amount"`
	Note          string `json:"note"`
	TransactedAt  string `json:"transacted_at"`
	BudgetPlanID  string `json:"budget_plan_id,omitempty"`
}

type txListData struct {
	Items []txItem `json:"items"`
	Total int64    `json:"total"`
}

type txListEnvelope struct {
	Data  *txListData `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type txEnvelope struct {
	Data  *txItem `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func authReq(t *testing.T, userID uuid.UUID, method, path string, body any) *http.Request {
	t.Helper()
	var r *http.Request
	if body == nil {
		r = httptest.NewRequest(method, path, nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshalling body: %v", err)
		}
		r = httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, userID))
	return r
}

func makeTx(t *testing.T, ts *testhelper.TestServer, userID, catID uuid.UUID, amount int64, note string) txItem {
	t.Helper()
	body := map[string]any{
		"category_id":   catID.String(),
		"amount":        amount,
		"note":          note,
		"transacted_at": time.Now().UTC().Format(time.RFC3339),
	}
	req := authReq(t, userID, http.MethodPost, "/v1/transactions", body)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create tx: want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	var env txEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if env.Data == nil {
		t.Fatal("create returned no data")
	}
	return *env.Data
}

// --- tests --------------------------------------------------------------

func TestIntegration_Transactions_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodPost, http.MethodGet} {
		req := httptest.NewRequest(method, "/v1/transactions", nil)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("%s /v1/transactions without auth: want 401, got %d", method, rr.Code)
		}
	}
}

func TestIntegration_Transactions_CreateValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, fmt.Sprintf("%s@local", uid))

	cases := map[string]map[string]any{
		"zero_amount": {
			"category_id": cats["Makan & minum"].String(),
			"amount":      0,
			"transacted_at": time.Now().UTC().Format(time.RFC3339),
		},
		"non_uuid_category": {
			"category_id":   "not-a-uuid",
			"amount":        1000,
			"transacted_at": time.Now().UTC().Format(time.RFC3339),
		},
		"bad_transacted_at": {
			"category_id":   cats["Makan & minum"].String(),
			"amount":        1000,
			"transacted_at": "yesterday",
		},
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			req := authReq(t, uid, http.MethodPost, "/v1/transactions", body)
			rr := httptest.NewRecorder()
			ts.Echo.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("want 400, got %d, body=%s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestIntegration_Transactions_RejectsUnknownCategory(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, "user@local")

	body := map[string]any{
		"category_id":   uuid.NewString(),
		"amount":        15_000,
		"transacted_at": time.Now().UTC().Format(time.RFC3339),
	}
	req := authReq(t, uid, http.MethodPost, "/v1/transactions", body)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}

func TestIntegration_Transactions_FullCRUDFlow(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, "user@local")

	// Create
	created := makeTx(t, ts, uid, cats["Makan & minum"], 25_000, "lunch")

	// Get by ID
	{
		req := authReq(t, uid, http.MethodGet, "/v1/transactions/"+created.ID, nil)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("get: want 200, got %d", rr.Code)
		}
	}

	// List
	{
		req := authReq(t, uid, http.MethodGet, "/v1/transactions", nil)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("list: want 200, got %d", rr.Code)
		}
		var env txListEnvelope
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatalf("decoding list: %v", err)
		}
		if env.Data == nil || len(env.Data.Items) != 1 || env.Data.Total != 1 {
			t.Fatalf("list returned unexpected data: %+v", env.Data)
		}
		if env.Data.Items[0].CategoryName == "" {
			t.Errorf("list item missing category_name")
		}
	}

	// Update amount + note
	{
		body := map[string]any{"amount": 30_000, "note": "lunch updated"}
		req := authReq(t, uid, http.MethodPatch, "/v1/transactions/"+created.ID, body)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("update: want 200, got %d, body=%s", rr.Code, rr.Body.String())
		}
		var env txEnvelope
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatal(err)
		}
		if env.Data.Amount != 30_000 || env.Data.Note != "lunch updated" {
			t.Errorf("update didn't apply: %+v", env.Data)
		}
	}

	// Delete
	{
		req := authReq(t, uid, http.MethodDelete, "/v1/transactions/"+created.ID, nil)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("delete: want 204, got %d", rr.Code)
		}
	}

	// Subsequent get → 404
	{
		req := authReq(t, uid, http.MethodGet, "/v1/transactions/"+created.ID, nil)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("get after delete: want 404, got %d", rr.Code)
		}
	}

	// Second delete → 404
	{
		req := authReq(t, uid, http.MethodDelete, "/v1/transactions/"+created.ID, nil)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("re-delete: want 404, got %d", rr.Code)
		}
	}
}

func TestIntegration_Transactions_FiltersByCategory(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, "user@local")

	makeTx(t, ts, uid, cats["Makan & minum"], 10_000, "")
	makeTx(t, ts, uid, cats["Makan & minum"], 12_000, "")
	makeTx(t, ts, uid, cats["Hiburan"], 50_000, "")

	req := authReq(t, uid, http.MethodGet,
		fmt.Sprintf("/v1/transactions?category_id=%s", cats["Makan & minum"]), nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	var env txListEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Total != 2 || len(env.Data.Items) != 2 {
		t.Fatalf("category filter: want 2 items, got total=%d count=%d", env.Data.Total, len(env.Data.Items))
	}
}

func TestIntegration_Transactions_FiltersByDateRange(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, "user@local")

	now := time.Now().UTC().Truncate(time.Second)
	old := now.Add(-30 * 24 * time.Hour)

	// One old, two recent.
	for _, ts2 := range []time.Time{old, now.Add(-1 * time.Hour), now.Add(-2 * time.Hour)} {
		body := map[string]any{
			"category_id":   cats["Makan & minum"].String(),
			"amount":        10_000,
			"transacted_at": ts2.Format(time.RFC3339),
		}
		req := authReq(t, uid, http.MethodPost, "/v1/transactions", body)
		rr := httptest.NewRecorder()
		ts.Echo.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("seed tx: %d %s", rr.Code, rr.Body.String())
		}
	}

	from := now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	to := now.Add(time.Hour).Format(time.RFC3339)
	req := authReq(t, uid, http.MethodGet,
		fmt.Sprintf("/v1/transactions?from=%s&to=%s", from, to), nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	var env txListEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Total != 2 {
		t.Errorf("date filter: want 2 in range, got %d", env.Data.Total)
	}
}

func TestIntegration_Transactions_CrossUserIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	a := uuid.New()
	b := uuid.New()
	seedUser(t, ts.DB, a, "a@local")
	seedUser(t, ts.DB, b, "b@local")

	created := makeTx(t, ts, a, cats["Makan & minum"], 10_000, "a's tx")

	// User B should NOT be able to see, update, or delete user A's tx.
	req := authReq(t, b, http.MethodGet, "/v1/transactions/"+created.ID, nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("user B GET A's tx: want 404, got %d", rr.Code)
	}

	body := map[string]any{"amount": 999_999}
	req = authReq(t, b, http.MethodPatch, "/v1/transactions/"+created.ID, body)
	rr = httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("user B PATCH A's tx: want 404, got %d", rr.Code)
	}

	req = authReq(t, b, http.MethodDelete, "/v1/transactions/"+created.ID, nil)
	rr = httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("user B DELETE A's tx: want 404, got %d", rr.Code)
	}

	// A's tx is still there.
	req = authReq(t, a, http.MethodGet, "/v1/transactions/"+created.ID, nil)
	rr = httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("A still sees own tx after B's attacks: got %d", rr.Code)
	}
}

func TestIntegration_Transactions_PaginationLimitOffset(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, "user@local")

	for i := 0; i < 5; i++ {
		makeTx(t, ts, uid, cats["Makan & minum"], int64(10_000+i), fmt.Sprintf("tx-%d", i))
	}

	req := authReq(t, uid, http.MethodGet, "/v1/transactions?limit=2&offset=2", nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	var env txListEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Total != 5 {
		t.Errorf("total: want 5, got %d", env.Data.Total)
	}
	if len(env.Data.Items) != 2 {
		t.Errorf("page size: want 2, got %d", len(env.Data.Items))
	}
}

func TestIntegration_Transactions_AutoLinksBudgetPlan(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()
	seedUser(t, ts.DB, uid, "user@local")

	// Manually insert a budget plan for the current month so the tx auto-links.
	now := time.Now().UTC()
	var planID uuid.UUID
	err := ts.DB.QueryRow(context.Background(), `
		insert into budget_plans (user_id, period_year, period_month, total_income, program)
		values ($1, $2, $3, $4, 'seimbang') returning id
	`, uid, int16(now.Year()), int16(now.Month()), int64(8_000_000)).Scan(&planID)
	if err != nil {
		t.Fatalf("seeding plan: %v", err)
	}

	tx := makeTx(t, ts, uid, cats["Makan & minum"], 10_000, "")
	if tx.BudgetPlanID != planID.String() {
		t.Errorf("auto-link: want budget_plan_id=%s, got %q", planID, tx.BudgetPlanID)
	}
}
