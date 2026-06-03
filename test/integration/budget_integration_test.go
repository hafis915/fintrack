package integration_test

import (
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

func seedPlanAndItems(
	t *testing.T,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	year, month int,
	totalIncome int64,
	program string,
	items []seededItem,
) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	var planID uuid.UUID
	err := pool.QueryRow(ctx, `
		insert into budget_plans (user_id, period_year, period_month, total_income, program)
		values ($1, $2, $3, $4, $5)
		returning id
	`, userID, int16(year), int16(month), totalIncome, program).Scan(&planID)
	if err != nil {
		t.Fatalf("seeding plan: %v", err)
	}

	for _, it := range items {
		_, err := pool.Exec(ctx, `
			insert into budget_items (budget_plan_id, category_id, allocated_amount, percentage)
			values ($1, $2, $3, $4)
		`, planID, it.CategoryID, it.AllocatedAmount, it.Percentage)
		if err != nil {
			t.Fatalf("seeding item: %v", err)
		}
	}
	return planID
}

type seededItem struct {
	CategoryID      uuid.UUID
	AllocatedAmount int64
	Percentage      float64
}

func seedTransaction(t *testing.T, pool *pgxpool.Pool, userID, catID uuid.UUID, amount int64, transactedAt time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into transactions (user_id, category_id, amount, transacted_at)
		values ($1, $2, $3, $4)
	`, userID, catID, amount, transactedAt)
	if err != nil {
		t.Fatalf("seeding transaction: %v", err)
	}
}

type budgetItemResp struct {
	ID              string  `json:"id"`
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	AllocatedAmount int64   `json:"allocated_amount"`
	SpentAmount     int64   `json:"spent_amount"`
	Remaining       int64   `json:"remaining"`
	PercentageUsed  float64 `json:"percentage_used"`
	Status          string  `json:"status"`
	Coaching        string  `json:"coaching"`
	IsDebtFocus     bool    `json:"is_debt_focus"`
}

type budgetSummaryResp struct {
	TotalAllocated    int64   `json:"total_allocated"`
	TotalSpent        int64   `json:"total_spent"`
	UnallocatedSpent  int64   `json:"unallocated_spent"`
	OverallPercentage float64 `json:"overall_percentage"`
}

type budgetResp struct {
	ID          string             `json:"id"`
	Period      string             `json:"period"`
	Program     string             `json:"program"`
	TotalIncome int64              `json:"total_income"`
	Items       []budgetItemResp   `json:"items"`
	Summary     budgetSummaryResp  `json:"summary"`
}

type budgetEnvelope struct {
	Data  *budgetResp `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func getBudgetCurrent(t *testing.T, ts *testhelper.TestServer, userID uuid.UUID) (*httptest.ResponseRecorder, budgetEnvelope) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v1/budget/current", nil)
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, userID))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	var env budgetEnvelope
	if rr.Body.Len() > 0 {
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatalf("decoding body %q: %v", rr.Body.String(), err)
		}
	}
	return rr, env
}

func findItem(t *testing.T, items []budgetItemResp, name string) budgetItemResp {
	t.Helper()
	for _, it := range items {
		if it.CategoryName == name {
			return it
		}
	}
	t.Fatalf("item %q not found in response", name)
	return budgetItemResp{}
}

// --- tests --------------------------------------------------------------

func TestIntegration_Budget_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/budget/current", nil)
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestIntegration_Budget_NoPlanReturns404(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, fmt.Sprintf("%s@local", uid))

	rr, env := getBudgetCurrent(t, ts, uid)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Error == nil || env.Error.Code != "no_active_plan" {
		t.Errorf("want code 'no_active_plan', got %+v", env.Error)
	}
}

func TestIntegration_Budget_PlanWithNoTransactionsIsAllFresh(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "u@local")
	now := time.Now().UTC()
	seedPlanAndItems(t, ts.DB, uid, now.Year(), int(now.Month()), 8_000_000, "seimbang", []seededItem{
		{CategoryID: cats["Makan & minum"], AllocatedAmount: 1_200_000, Percentage: 15},
		{CategoryID: cats["Hiburan"], AllocatedAmount: 500_000, Percentage: 6.25},
	})

	rr, env := getBudgetCurrent(t, ts, uid)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Data == nil {
		t.Fatal("data missing")
	}
	if len(env.Data.Items) != 2 {
		t.Fatalf("items: want 2, got %d", len(env.Data.Items))
	}
	for _, it := range env.Data.Items {
		if it.Status != "fresh" {
			t.Errorf("item %q: want fresh, got %q", it.CategoryName, it.Status)
		}
		if it.SpentAmount != 0 {
			t.Errorf("item %q: want spent=0, got %d", it.CategoryName, it.SpentAmount)
		}
		if it.Remaining != it.AllocatedAmount {
			t.Errorf("item %q: remaining should equal allocated when nothing spent", it.CategoryName)
		}
	}
	if env.Data.Summary.TotalSpent != 0 {
		t.Errorf("summary total_spent: want 0, got %d", env.Data.Summary.TotalSpent)
	}
	if env.Data.Summary.TotalAllocated != 1_700_000 {
		t.Errorf("summary total_allocated: want 1_700_000, got %d", env.Data.Summary.TotalAllocated)
	}
}

func TestIntegration_Budget_MixedStatusesPerCategory(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "u@local")
	now := time.Now().UTC()
	seedPlanAndItems(t, ts.DB, uid, now.Year(), int(now.Month()), 8_000_000, "seimbang", []seededItem{
		{CategoryID: cats["Makan & minum"], AllocatedAmount: 1_000_000, Percentage: 12.5}, // will warn
		{CategoryID: cats["Hiburan"], AllocatedAmount: 500_000, Percentage: 6.25},         // stays fresh
		{CategoryID: cats["Kartu kredit"], AllocatedAmount: 400_000, Percentage: 5},       // will fatigue
	})

	// Makan: 800_000 / 1_000_000 = 80% → warning
	seedTransaction(t, ts.DB, uid, cats["Makan & minum"], 500_000, now)
	seedTransaction(t, ts.DB, uid, cats["Makan & minum"], 300_000, now.Add(-1*time.Hour))
	// Hiburan: 100_000 / 500_000 = 20% → fresh
	seedTransaction(t, ts.DB, uid, cats["Hiburan"], 100_000, now)
	// Kartu kredit: 500_000 / 400_000 = 125% → fatigued
	seedTransaction(t, ts.DB, uid, cats["Kartu kredit"], 500_000, now)

	rr, env := getBudgetCurrent(t, ts, uid)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	makan := findItem(t, env.Data.Items, "Makan & minum")
	if makan.Status != "warning" {
		t.Errorf("makan: want warning, got %q (used=%.2f)", makan.Status, makan.PercentageUsed)
	}
	if makan.SpentAmount != 800_000 || makan.Remaining != 200_000 {
		t.Errorf("makan totals off: spent=%d remaining=%d", makan.SpentAmount, makan.Remaining)
	}

	hiburan := findItem(t, env.Data.Items, "Hiburan")
	if hiburan.Status != "fresh" {
		t.Errorf("hiburan: want fresh, got %q", hiburan.Status)
	}

	cc := findItem(t, env.Data.Items, "Kartu kredit")
	if cc.Status != "fatigued" {
		t.Errorf("cc: want fatigued, got %q", cc.Status)
	}
	if cc.Remaining >= 0 {
		t.Errorf("cc remaining should be negative when fatigued, got %d", cc.Remaining)
	}

	// summary totals
	wantTotalSpent := int64(800_000 + 100_000 + 500_000)
	if env.Data.Summary.TotalSpent != wantTotalSpent {
		t.Errorf("summary total_spent: want %d, got %d", wantTotalSpent, env.Data.Summary.TotalSpent)
	}
}

func TestIntegration_Budget_IgnoresSoftDeletedTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "u@local")
	now := time.Now().UTC()
	seedPlanAndItems(t, ts.DB, uid, now.Year(), int(now.Month()), 8_000_000, "seimbang", []seededItem{
		{CategoryID: cats["Makan & minum"], AllocatedAmount: 1_000_000, Percentage: 12.5},
	})
	seedTransaction(t, ts.DB, uid, cats["Makan & minum"], 800_000, now)

	// Soft-delete that tx — should drop out of spent.
	_, err := ts.DB.Exec(context.Background(),
		`update transactions set deleted_at = now() where user_id = $1`, uid)
	if err != nil {
		t.Fatal(err)
	}

	_, env := getBudgetCurrent(t, ts, uid)
	if env.Data == nil {
		t.Fatal("nil data")
	}
	it := findItem(t, env.Data.Items, "Makan & minum")
	if it.SpentAmount != 0 {
		t.Errorf("soft-deleted tx still counted: spent=%d", it.SpentAmount)
	}
}

func TestIntegration_Budget_IgnoresOutOfPeriodTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "u@local")
	now := time.Now().UTC()
	seedPlanAndItems(t, ts.DB, uid, now.Year(), int(now.Month()), 8_000_000, "seimbang", []seededItem{
		{CategoryID: cats["Makan & minum"], AllocatedAmount: 1_000_000, Percentage: 12.5},
	})
	// 45 days ago — definitely a different calendar month.
	seedTransaction(t, ts.DB, uid, cats["Makan & minum"], 500_000, now.Add(-45*24*time.Hour))

	_, env := getBudgetCurrent(t, ts, uid)
	if env.Data == nil {
		t.Fatal("nil data")
	}
	it := findItem(t, env.Data.Items, "Makan & minum")
	if it.SpentAmount != 0 {
		t.Errorf("out-of-period tx counted: spent=%d", it.SpentAmount)
	}
}

func TestIntegration_Budget_UnallocatedSpendInSummary(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "u@local")
	now := time.Now().UTC()
	// Plan only covers Makan; Nongkrong is not in the plan.
	seedPlanAndItems(t, ts.DB, uid, now.Year(), int(now.Month()), 8_000_000, "seimbang", []seededItem{
		{CategoryID: cats["Makan & minum"], AllocatedAmount: 1_000_000, Percentage: 12.5},
	})
	seedTransaction(t, ts.DB, uid, cats["Makan & minum"], 200_000, now)
	seedTransaction(t, ts.DB, uid, cats["Nongkrong"], 150_000, now) // unallocated

	_, env := getBudgetCurrent(t, ts, uid)
	if env.Data.Summary.UnallocatedSpent != 150_000 {
		t.Errorf("unallocated_spent: want 150_000, got %d", env.Data.Summary.UnallocatedSpent)
	}
	if env.Data.Summary.TotalSpent != 350_000 {
		t.Errorf("total_spent should include unallocated: want 350_000, got %d",
			env.Data.Summary.TotalSpent)
	}
}

func TestIntegration_Budget_CrossUserIsolation(t *testing.T) {
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
	now := time.Now().UTC()

	// A has a plan. B does not.
	seedPlanAndItems(t, ts.DB, a, now.Year(), int(now.Month()), 8_000_000, "seimbang", []seededItem{
		{CategoryID: cats["Makan & minum"], AllocatedAmount: 1_000_000, Percentage: 12.5},
	})
	seedTransaction(t, ts.DB, a, cats["Makan & minum"], 500_000, now)
	// B has a transaction but no plan — should still 404 for B.
	seedTransaction(t, ts.DB, b, cats["Makan & minum"], 100_000, now)

	rr, env := getBudgetCurrent(t, ts, b)
	if rr.Code != http.StatusNotFound {
		t.Errorf("B expected 404 (no plan), got %d", rr.Code)
	}
	if env.Error == nil || env.Error.Code != "no_active_plan" {
		t.Errorf("B expected no_active_plan, got %+v", env.Error)
	}

	// A should still see only their own spend (500_000), not B's.
	_, envA := getBudgetCurrent(t, ts, a)
	if envA.Data == nil {
		t.Fatal("A response missing data")
	}
	it := findItem(t, envA.Data.Items, "Makan & minum")
	if it.SpentAmount != 500_000 {
		t.Errorf("A's spent leaked from B: want 500_000, got %d", it.SpentAmount)
	}
}
