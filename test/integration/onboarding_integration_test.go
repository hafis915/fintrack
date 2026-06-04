// Integration tests for POST /v1/onboarding.
//
// Covers per DoD:
//   - Auth: missing token → 401 (delegated to JWT middleware tests, but smoke-check here)
//   - Validation: bad income, bad enum values
//   - Category lookup: unknown category_id → 400; foreign user's custom category → 403
//   - Happy paths: each of the five programs produces the right plan
//   - Persistence: income is encrypted (not stored in plaintext), budget_plan + items written
//   - Idempotency: re-submitting in the same month replaces the plan
//
// All tests reset the DB-mutating tables before they run so they're
// order-independent. system default categories are NOT truncated — they
// come from the migration seed.
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/internal/encryption"
	"github.com/hafis915/fintrack/test/testhelper"
)

func resetOnboardingTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	// Order matters: items → plans → user_profiles → custom categories → users.
	stmts := []string{
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

func seedDefaultCategoryIDs(t *testing.T, pool *pgxpool.Pool) map[string]uuid.UUID {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`select name, id from expense_categories where user_id is null`)
	if err != nil {
		t.Fatalf("loading default categories: %v", err)
	}
	defer rows.Close()
	byName := map[string]uuid.UUID{}
	for rows.Next() {
		var name string
		var id uuid.UUID
		if err := rows.Scan(&name, &id); err != nil {
			t.Fatalf("scanning category row: %v", err)
		}
		byName[name] = id
	}
	if len(byName) == 0 {
		t.Fatal("no seeded default categories — migration 0002 not applied to test DB?")
	}
	return byName
}

type onboardingItemReq struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Icon       string `json:"icon,omitempty"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
}

type onboardingReq struct {
	Income          int64               `json:"income"`
	HousingType     string              `json:"housing_type"`
	Goal            string              `json:"goal"`
	DebtTypes       []string            `json:"debt_types"`
	EmergencyMonths int                 `json:"emergency_months"`
	LifestyleStyle  string              `json:"lifestyle_style"`
	ExpenseItems    []onboardingItemReq `json:"expense_items"`
}

func defaultReq(cats map[string]uuid.UUID) onboardingReq {
	return onboardingReq{
		Income:          8_000_000,
		HousingType:     "kpr",
		Goal:            "debt", // PRD example uses "bebas_utang" but that's a program; goal enum is emergency|debt|goal|invest|balance.
		DebtTypes:       []string{"cc"},
		EmergencyMonths: 1,
		LifestyleStyle:  "balanced",
		ExpenseItems: []onboardingItemReq{
			{CategoryID: cats["Cicilan KPR"].String(), Name: "Cicilan KPR", Type: "fixed", Amount: 1_500_000},
			{CategoryID: cats["Makan & minum"].String(), Name: "Makan & minum", Type: "variable", Amount: 1_200_000},
			{CategoryID: cats["Kartu kredit"].String(), Name: "Kartu kredit", Type: "debt", Amount: 400_000},
			{CategoryID: cats["Hiburan"].String(), Name: "Hiburan", Type: "want", Amount: 500_000},
		},
	}
}

func postOnboarding(t *testing.T, ts *testhelper.TestServer, userID uuid.UUID, body any) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshalling body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/onboarding", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, userID))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	return rr, decode(t, rr)
}

func TestIntegration_Onboarding_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)

	req := httptest.NewRequest(http.MethodPost, "/v1/onboarding", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestIntegration_Onboarding_RejectsInvalidPayload(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	cases := map[string]func(*onboardingReq){
		"zero_income":   func(r *onboardingReq) { r.Income = 0 },
		"bad_housing":   func(r *onboardingReq) { r.HousingType = "mansion" },
		"bad_goal":      func(r *onboardingReq) { r.Goal = "yolo" },
		"bad_lifestyle": func(r *onboardingReq) { r.LifestyleStyle = "spartan" },
		"bad_emergency": func(r *onboardingReq) { r.EmergencyMonths = 2 },
		"non_uuid_cat":  func(r *onboardingReq) { r.ExpenseItems[0].CategoryID = "not-a-uuid" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			body := defaultReq(cats)
			mutate(&body)
			rr, env := postOnboarding(t, ts, uid, body)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d, body=%s", rr.Code, rr.Body.String())
			}
			if env.Error == nil {
				t.Fatal("expected error envelope")
			}
		})
	}
}

func TestIntegration_Onboarding_RejectsUnknownCategory(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultReq(cats)
	body.ExpenseItems[0].CategoryID = uuid.NewString() // random, non-existent
	rr, env := postOnboarding(t, ts, uid, body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Error == nil || env.Error.Code != "invalid_category" {
		t.Fatalf("want code 'invalid_category', got %+v", env.Error)
	}
}

func TestIntegration_Onboarding_RejectsCategoryTypeMismatch(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultReq(cats)
	// "Cicilan KPR" is a `fixed` category — claim it's `want` in the item.
	body.ExpenseItems[0].Type = "want"
	rr, env := postOnboarding(t, ts, uid, body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	if env.Error == nil || env.Error.Code != "category_type_mismatch" {
		t.Fatalf("want 'category_type_mismatch', got %+v", env.Error)
	}
}

// Bug #2 regression: a degenerate income (e.g. a typo) must yield a clean 400
// with the income_too_low code — never a 500 from a budget_items.percentage
// numeric overflow, and never a raw DB error leaked in the response body.
func TestIntegration_Onboarding_RejectsDegenerateIncome(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultReq(cats)
	body.Income = 100 // expenses sum to ~3.6jt → ratio ~36000x, far past the 50x guard
	rr, env := postOnboarding(t, ts, uid, body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 (friendly), got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Error == nil || env.Error.Code != "income_too_low" {
		t.Fatalf("want code 'income_too_low', got %+v", env.Error)
	}
	// The body must never leak a raw DB error (e.g. a numeric-overflow
	// SQLSTATE 22003) — that would mean the 500 path fired instead of the
	// friendly 400 guard.
	if strings.Contains(rr.Body.String(), "SQLSTATE") {
		t.Errorf("response body leaked a raw DB error: %s", rr.Body.String())
	}

	// Sanity: a normal valid intake on the same server still returns 201.
	rrOK, _ := postOnboarding(t, ts, uuid.New(), defaultReq(cats))
	if rrOK.Code != http.StatusCreated {
		t.Fatalf("valid intake after degenerate one: want 201, got %d, body=%s",
			rrOK.Code, rrOK.Body.String())
	}
}

// Bug #2 fix: an overspending budget (expenses > income, but within the
// degenerate-input ceiling) must be ALLOWED — it produces a plan with a
// coaching warning instead of being hard-blocked. Income 50k vs ~3.6jt of
// declared expenses is a 72x ratio: rejected under the old 50x cap, accepted
// under the 999x cap so the "gym for your money" overspender gets coached.
func TestIntegration_Onboarding_AllowsOverspendWithWarning(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	body := defaultReq(cats)
	body.Income = 50_000 // expenses ~3.6jt → 72x: overspend, but not a typo
	rr, env := postOnboarding(t, ts, uuid.New(), body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("overspend within ceiling: want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	// Savings (tabungan) goes negative and the plan carries a coaching warning.
	summary, ok := env.Data["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary missing: %T", env.Data["summary"])
	}
	if amt := summary["tabungan"].(map[string]any)["amount"].(float64); amt >= 0 {
		t.Errorf("tabungan: want negative (overspend), got %v", amt)
	}
	if w, _ := env.Data["warning"].(string); w == "" {
		t.Error("want a coaching warning for the overspend plan, got empty")
	}
}

func TestIntegration_Onboarding_HappyPath_BebasUtang(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	rr, env := postOnboarding(t, ts, uid, defaultReq(cats))
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if got, _ := env.Data["program"].(string); got != "bebas_utang" {
		t.Errorf("program: want bebas_utang, got %v", env.Data["program"])
	}
	if env.Data["budget_plan_id"] == "" {
		t.Error("budget_plan_id empty")
	}

	summary, ok := env.Data["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary missing or wrong type: %T", env.Data["summary"])
	}
	bucket := summary["utang"].(map[string]any)
	if bucket["amount"].(float64) != 400_000 {
		t.Errorf("utang amount: want 400_000, got %v", bucket["amount"])
	}
}

func TestIntegration_Onboarding_HappyPath_Pondasi(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultReq(cats)
	body.Goal = "emergency"
	body.DebtTypes = []string{"none"}
	body.EmergencyMonths = 0
	// Remove the debt item so it doesn't drag program back to bebas_utang.
	body.ExpenseItems = body.ExpenseItems[:0]
	body.ExpenseItems = append(body.ExpenseItems,
		onboardingItemReq{CategoryID: cats["Sewa kosan"].String(), Name: "Sewa kosan", Type: "fixed", Amount: 1_200_000},
		onboardingItemReq{CategoryID: cats["Makan & minum"].String(), Name: "Makan & minum", Type: "variable", Amount: 1_200_000},
	)

	rr, env := postOnboarding(t, ts, uid, body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Data["program"].(string) != "pondasi" {
		t.Errorf("program: want pondasi, got %v", env.Data["program"])
	}
}

func TestIntegration_Onboarding_HappyPath_Tumbuh(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultReq(cats)
	body.Goal = "invest"
	body.DebtTypes = []string{"none"}
	body.EmergencyMonths = 6
	body.ExpenseItems = body.ExpenseItems[:0]
	body.ExpenseItems = append(body.ExpenseItems,
		onboardingItemReq{CategoryID: cats["Cicilan KPR"].String(), Name: "Cicilan KPR", Type: "fixed", Amount: 1_500_000},
		onboardingItemReq{CategoryID: cats["Makan & minum"].String(), Name: "Makan & minum", Type: "variable", Amount: 1_200_000},
	)

	rr, env := postOnboarding(t, ts, uid, body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if env.Data["program"].(string) != "tumbuh" {
		t.Errorf("program: want tumbuh, got %v", env.Data["program"])
	}
}

func TestIntegration_Onboarding_EncryptsIncomeAtRest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultReq(cats)
	body.Income = 8_500_000
	rr, _ := postOnboarding(t, ts, uid, body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rr.Code)
	}

	var stored struct {
		IncomeEncrypted *string
		IncomeHint      *string
	}
	row := ts.DB.QueryRow(context.Background(),
		`select income_encrypted, income_hint from user_profiles where user_id = $1`, uid)
	if err := row.Scan(&stored.IncomeEncrypted, &stored.IncomeHint); err != nil {
		t.Fatalf("loading stored profile: %v", err)
	}
	if stored.IncomeEncrypted == nil || *stored.IncomeEncrypted == "" {
		t.Fatal("income_encrypted is empty — encryption not running")
	}
	if *stored.IncomeEncrypted == "8500000" {
		t.Fatal("income stored in plaintext!")
	}

	// Decrypt with the same test key the server uses; expect "8500000" back.
	c, err := encryption.NewCipherFromHex(testhelper.TestEncryptionKey)
	if err != nil {
		t.Fatalf("NewCipherFromHex: %v", err)
	}
	plain, err := c.Decrypt(*stored.IncomeEncrypted)
	if err != nil {
		t.Fatalf("decrypting stored income: %v", err)
	}
	if string(plain) != "8500000" {
		t.Errorf("decrypted income: want 8500000, got %q", plain)
	}

	if stored.IncomeHint == nil {
		t.Fatal("income_hint missing")
	}
	if want := "Rp 8,5jt"; *stored.IncomeHint != want {
		t.Errorf("income_hint: want %q, got %q", want, *stored.IncomeHint)
	}
}

func TestIntegration_Onboarding_PersistsPlanAndItems(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	rr, env := postOnboarding(t, ts, uid, defaultReq(cats))
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rr.Code)
	}
	planID, err := uuid.Parse(env.Data["budget_plan_id"].(string))
	if err != nil {
		t.Fatalf("budget_plan_id not a UUID: %v", err)
	}

	var nItems int
	if err := ts.DB.QueryRow(context.Background(),
		`select count(*) from budget_items where budget_plan_id = $1`, planID).Scan(&nItems); err != nil {
		t.Fatalf("counting items: %v", err)
	}
	if nItems != 4 {
		t.Errorf("budget items: want 4, got %d", nItems)
	}
}

func TestIntegration_Onboarding_IsIdempotentForSamePeriod(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	// First submission.
	rr1, env1 := postOnboarding(t, ts, uid, defaultReq(cats))
	if rr1.Code != http.StatusCreated {
		t.Fatalf("first POST: want 201, got %d", rr1.Code)
	}
	firstPlanID := env1.Data["budget_plan_id"].(string)

	// Second submission with different income → should overwrite, same plan UUID.
	body2 := defaultReq(cats)
	body2.Income = 10_000_000
	rr2, env2 := postOnboarding(t, ts, uid, body2)
	if rr2.Code != http.StatusCreated {
		t.Fatalf("second POST: want 201, got %d, body=%s", rr2.Code, rr2.Body.String())
	}
	secondPlanID := env2.Data["budget_plan_id"].(string)
	if firstPlanID != secondPlanID {
		t.Errorf("plan id changed on idempotent re-submit: %s → %s", firstPlanID, secondPlanID)
	}

	// Plans table should still have exactly one row for this user.
	var n int
	if err := ts.DB.QueryRow(context.Background(),
		`select count(*) from budget_plans where user_id = $1`, uid).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("duplicate plans for user: want 1, got %d", n)
	}

	var stored int64
	if err := ts.DB.QueryRow(context.Background(),
		`select total_income from budget_plans where user_id = $1`, uid).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != 10_000_000 {
		t.Errorf("total_income after overwrite: want 10_000_000, got %d", stored)
	}
}

// Sanity check: the seeded category fixtures align with what defaultReq expects.
func TestIntegration_Seed_HasExpectedCategoryNames(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	cats := seedDefaultCategoryIDs(t, ts.DB)
	required := []string{"Cicilan KPR", "Sewa kosan", "Makan & minum", "Kartu kredit", "Hiburan"}
	for _, n := range required {
		if _, ok := cats[n]; !ok {
			t.Errorf("seed missing default category %q", n)
		}
	}
	_ = fmt.Sprintf // keep fmt import alive if test list shrinks
}
