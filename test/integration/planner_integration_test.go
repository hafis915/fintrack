// Integration tests for the deterministic financial-planner endpoints:
//
//   - POST /v1/onboarding/suggest — given the intake answers + the user's FIXED
//     expenses, deterministically suggests amounts for the FLEXIBLE categories.
//     Pure money math, no LLM.
//   - POST /v1/planner/chat       — multi-turn natural-language refinement. The
//     LLM (a deterministic stub in tests) only interprets intent + narrates; the
//     actual number changes are re-balanced deterministically by app code.
//
// The testhelper injects llm.NewStubClient(), whose NLU is deterministic:
// "naikin makan jadi 1500000" → set the "makan" flexible category to 1,500,000
// and let the app rebalance the other wants. A message with no category+amount
// (e.g. "halo") yields no adjustments → changed=false.
//
// These tests reuse the seeded default categories from migration 0002 (same as
// the onboarding tests) and reset the DB-mutating tables so they're
// order-independent. They run against the real fintrack_test Postgres.
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

// --- POST /v1/onboarding/suggest fixtures ---------------------------------

type suggestFixedItemReq struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Icon       string `json:"icon,omitempty"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
}

type suggestReq struct {
	Income          int64                 `json:"income"`
	HousingType     string                `json:"housing_type"`
	Goal            string                `json:"goal"`
	DebtTypes       []string              `json:"debt_types"`
	EmergencyMonths int                   `json:"emergency_months"`
	LifestyleStyle  string                `json:"lifestyle_style"`
	FixedItems      []suggestFixedItemReq `json:"fixed_items"`
}

// defaultSuggestReq is a sane 9jt-income intake: a kosan + a utility bill as the
// only fixed expenses. The flexible set comes from the seeded catalog.
func defaultSuggestReq(cats map[string]uuid.UUID) suggestReq {
	return suggestReq{
		Income:          9_000_000,
		HousingType:     "kosan",
		Goal:            "balance",
		DebtTypes:       []string{"none"},
		EmergencyMonths: 3,
		LifestyleStyle:  "balanced",
		FixedItems: []suggestFixedItemReq{
			{CategoryID: cats["Sewa kosan"].String(), Name: "Sewa kosan", Type: "fixed", Amount: 1_500_000},
			{CategoryID: cats["Listrik & air"].String(), Name: "Listrik & air", Type: "fixed", Amount: 300_000},
		},
	}
}

func postSuggest(t *testing.T, ts *testhelper.TestServer, userID uuid.UUID, body any) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshalling suggest body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/onboarding/suggest", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, userID))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	return rr, decode(t, rr)
}

// --- POST /v1/planner/chat fixtures ---------------------------------------

type chatItemReq struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Amount     int64  `json:"amount"`
}

type chatMessageReq struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Income         int64            `json:"income"`
	Goal           string           `json:"goal"`
	LifestyleStyle string           `json:"lifestyle_style"`
	SavingsTarget  int64            `json:"savings_target"`
	FixedItems     []chatItemReq    `json:"fixed_items"`
	Flexible       []chatItemReq    `json:"flexible"`
	Messages       []chatMessageReq `json:"messages"`
	UserMessage    string           `json:"user_message"`
}

func postChat(t *testing.T, ts *testhelper.TestServer, userID uuid.UUID, body any) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshalling chat body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/planner/chat", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, userID))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	return rr, decode(t, rr)
}

// flexFromEnvelope pulls the flexible[] array out of a chat/suggest envelope as
// a name→amount map and the raw slice, so tests can assert per-category amounts
// and total sums regardless of ordering.
func flexFromEnvelope(t *testing.T, env envelope, amountKey string) (map[string]int64, int64) {
	t.Helper()
	raw, ok := env.Data["flexible"].([]any)
	if !ok {
		t.Fatalf("flexible missing or wrong type: %T", env.Data["flexible"])
	}
	byName := map[string]int64{}
	var total int64
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("flexible item wrong type: %T", item)
		}
		name, _ := m["name"].(string)
		amt, _ := m[amountKey].(float64)
		byName[name] = int64(amt)
		total += int64(amt)
	}
	return byName, total
}

// --- POST /v1/onboarding/suggest tests ------------------------------------

func TestIntegration_Suggest_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)

	req := httptest.NewRequest(http.MethodPost, "/v1/onboarding/suggest", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 for missing token, got %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestIntegration_Suggest_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultSuggestReq(cats)
	rr, env := postSuggest(t, ts, uid, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	// Program is present and non-empty.
	if prog, _ := env.Data["program"].(string); prog == "" {
		t.Errorf("program: want non-empty, got %v", env.Data["program"])
	}

	// fixed_total must match exactly what the user entered (1.5jt + 300k).
	var wantFixed int64
	for _, f := range body.FixedItems {
		wantFixed += f.Amount
	}
	fixedTotal, ok := env.Data["fixed_total"].(float64)
	if !ok {
		t.Fatalf("fixed_total missing or wrong type: %T", env.Data["fixed_total"])
	}
	if int64(fixedTotal) != wantFixed {
		t.Errorf("fixed_total: want %d, got %v", wantFixed, fixedTotal)
	}

	// savings_target must be positive for a comfortable 9jt intake.
	savings, ok := env.Data["savings_target"].(float64)
	if !ok {
		t.Fatalf("savings_target missing or wrong type: %T", env.Data["savings_target"])
	}
	if savings <= 0 {
		t.Errorf("savings_target: want positive, got %v", savings)
	}

	// flexible[] must be non-empty (the seeded want/variable categories).
	flexByName, flexTotal := flexFromEnvelope(t, env, "suggested_amount")
	if len(flexByName) == 0 {
		t.Fatal("flexible: want non-empty suggestions, got none")
	}

	// Invariant: fixed + flexible + savings stays within income (the finalize
	// step persists savings as income - sum(items), so this must not exceed
	// income).
	if sum := wantFixed + flexTotal + int64(savings); sum > body.Income {
		t.Errorf("fixed+flexible+savings = %d exceeds income %d", sum, body.Income)
	}
}

// An over-budget intake (fixed expenses far exceeding income) must still return
// 200 with a coaching warning and must NOT produce negative flexible amounts —
// the planner protects needs and floors the discretionary set at zero.
func TestIntegration_Suggest_OverBudget_WarnsNoNegatives(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := defaultSuggestReq(cats)
	// Fixed expenses well above income, but within the degenerate-input ceiling
	// so we get a coached plan (not an income_too_low 400). 9jt income, ~9.3jt
	// fixed → discretionary is squeezed to zero.
	body.FixedItems = []suggestFixedItemReq{
		{CategoryID: cats["Sewa kosan"].String(), Name: "Sewa kosan", Type: "fixed", Amount: 8_000_000},
		{CategoryID: cats["Listrik & air"].String(), Name: "Listrik & air", Type: "fixed", Amount: 1_300_000},
	}

	rr, env := postSuggest(t, ts, uid, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("over-budget: want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	if w, _ := env.Data["warning"].(string); w == "" {
		t.Error("over-budget: want a coaching warning, got empty")
	}

	flexByName, _ := flexFromEnvelope(t, env, "suggested_amount")
	for name, amt := range flexByName {
		if amt < 0 {
			t.Errorf("flexible %q: negative suggested amount %d", name, amt)
		}
	}
}

// --- POST /v1/planner/chat tests ------------------------------------------

func TestIntegration_PlannerChat_RequiresAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/planner/chat", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 for missing token, got %d, body=%s", rr.Code, rr.Body.String())
	}
}

// chatBaseReq returns a chat request with a current flexible set including Makan,
// Belanja, and Hiburan — enough wants for the rebalancer to pull from when one
// goes up. Income 9jt, savings target 2jt, fixed 1.8jt → discretionary 5.2jt,
// flexible currently sums to 4.2jt (room to grow).
func chatBaseReq(cats map[string]uuid.UUID) chatReq {
	return chatReq{
		Income:         9_000_000,
		Goal:           "balance",
		LifestyleStyle: "balanced",
		SavingsTarget:  2_000_000,
		FixedItems: []chatItemReq{
			{CategoryID: cats["Sewa kosan"].String(), Name: "Sewa kosan", Amount: 1_500_000},
			{CategoryID: cats["Listrik & air"].String(), Name: "Listrik & air", Amount: 300_000},
		},
		Flexible: []chatItemReq{
			{CategoryID: cats["Makan & minum"].String(), Name: "Makan & minum", Amount: 1_200_000},
			{CategoryID: cats["Belanja harian"].String(), Name: "Belanja harian", Amount: 1_500_000},
			{CategoryID: cats["Hiburan"].String(), Name: "Hiburan", Amount: 1_500_000},
		},
		Messages: []chatMessageReq{},
	}
}

// "naikin makan jadi 1500000": the stub resolves "makan" + 1,500,000; the app
// sets Makan & minum to 1.5jt and pulls the +300k from the other wants, keeping
// the total within income. changed=true and a non-empty reply.
func TestIntegration_PlannerChat_RaisesCategory_RebalancesOthers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := chatBaseReq(cats)
	// Sum of fixed + flexible + savings before the change (sanity anchor).
	var beforeFlex int64
	for _, f := range body.Flexible {
		beforeFlex += f.Amount
	}
	body.UserMessage = "naikin makan jadi 1500000"

	rr, env := postChat(t, ts, uid, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	if changed, _ := env.Data["changed"].(bool); !changed {
		t.Errorf("changed: want true after an explicit adjustment, got false; body=%s", rr.Body.String())
	}
	if reply, _ := env.Data["reply"].(string); reply == "" {
		t.Error("reply: want non-empty, got empty")
	}

	flexByName, afterFlex := flexFromEnvelope(t, env, "amount")
	if got := flexByName["Makan & minum"]; got != 1_500_000 {
		t.Errorf("Makan & minum: want 1_500_000, got %d", got)
	}

	// The +300k raise must be absorbed by the OTHER wants (protect needs +
	// savings): at least one of Belanja/Hiburan must have dropped below its
	// starting amount.
	if flexByName["Belanja harian"] >= 1_500_000 && flexByName["Hiburan"] >= 1_500_000 {
		t.Errorf("other wants not reduced: belanja=%d hiburan=%d (want at least one < 1_500_000)",
			flexByName["Belanja harian"], flexByName["Hiburan"])
	}

	// Discretionary envelope is preserved (rebalance is a pure reshuffle within
	// the wants, savings untouched), so the flexible total should not grow.
	if afterFlex > beforeFlex {
		t.Errorf("flexible total grew: before=%d after=%d (raise should pull from other wants)",
			beforeFlex, afterFlex)
	}

	// Savings target must not be reduced (no "ambil dari tabungan" in the
	// message), and the whole plan must stay within income.
	savings, ok := env.Data["savings_target"].(float64)
	if !ok {
		t.Fatalf("savings_target missing or wrong type: %T", env.Data["savings_target"])
	}
	if int64(savings) != body.SavingsTarget {
		t.Errorf("savings_target changed without consent: want %d, got %d", body.SavingsTarget, int64(savings))
	}
	var fixedTotal int64
	for _, f := range body.FixedItems {
		fixedTotal += f.Amount
	}
	if sum := fixedTotal + afterFlex + int64(savings); sum > body.Income {
		t.Errorf("fixed+flexible+savings = %d exceeds income %d", sum, body.Income)
	}
}

// A no-op message ("halo") has no category+amount, so the stub returns no
// adjustments: the plan is unchanged and changed=false.
func TestIntegration_PlannerChat_NoOpMessage_NoChange(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetOnboardingTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)
	uid := uuid.New()

	body := chatBaseReq(cats)
	body.UserMessage = "halo"

	rr, env := postChat(t, ts, uid, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	if changed, _ := env.Data["changed"].(bool); changed {
		t.Errorf("changed: want false for a no-op message, got true; body=%s", rr.Body.String())
	}
	if reply, _ := env.Data["reply"].(string); reply == "" {
		t.Error("reply: want non-empty even on no-op, got empty")
	}

	// Flexible amounts must be untouched.
	flexByName, _ := flexFromEnvelope(t, env, "amount")
	if got := flexByName["Makan & minum"]; got != 1_200_000 {
		t.Errorf("Makan & minum changed on no-op: want 1_200_000, got %d", got)
	}
}
