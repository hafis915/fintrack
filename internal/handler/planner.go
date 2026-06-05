package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/internal/llm"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/responses"
)

// Planner wires the deterministic financial-planner endpoints:
//
//   - POST /v1/onboarding/suggest — given the intake answers + the user's FIXED
//     expenses, deterministically suggest amounts for the FLEXIBLE categories.
//   - POST /v1/planner/chat       — multi-turn natural-language refinement. The
//     LLM only interprets intent + narrates; the actual numbers are re-balanced
//     deterministically by budget.Rebalance. The LLM never invents budget numbers.
//
// The handler holds the categories repo (to resolve the user's flexible category
// set), the budget allocator (pure functions, called directly), and an llm.Client
// (real OpenRouter when an API key is set, deterministic stub otherwise).
type Planner struct {
	categories repository.CategoriesRepo
	llm        llm.Client
}

// NewPlanner constructs the planner handler. The llm.Client is selected in
// internal/server (stub when OPEN_ROUTER_API_KEY is empty).
func NewPlanner(categories repository.CategoriesRepo, client llm.Client) *Planner {
	return &Planner{categories: categories, llm: client}
}

// --- shared request shapes ------------------------------------------------

// plannerFixedItem mirrors one fixed (type fixed/debt) expense the user entered.
type plannerFixedItem struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
}

// --- POST /v1/onboarding/suggest ------------------------------------------

type suggestRequest struct {
	Income          int64              `json:"income"`
	HousingType     string             `json:"housing_type"`
	Goal            string             `json:"goal"`
	DebtTypes       []string           `json:"debt_types"`
	EmergencyMonths int                `json:"emergency_months"`
	LifestyleStyle  string             `json:"lifestyle_style"`
	FixedItems      []plannerFixedItem `json:"fixed_items"`
}

type suggestBucket struct {
	Amount     int64   `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type suggestFlexItem struct {
	CategoryID      string `json:"category_id"`
	Name            string `json:"name"`
	Icon            string `json:"icon,omitempty"`
	Type            string `json:"type"`
	SuggestedAmount int64  `json:"suggested_amount"`
}

type suggestResponse struct {
	Program       string                   `json:"program"`
	SavingsTarget int64                    `json:"savings_target"`
	FixedTotal    int64                    `json:"fixed_total"`
	Discretionary int64                    `json:"discretionary"`
	Summary       map[string]suggestBucket `json:"summary"`
	Flexible      []suggestFlexItem        `json:"flexible"`
	Warning       string                   `json:"warning,omitempty"`
}

// Suggest is POST /v1/onboarding/suggest. It validates the intake answers, loads
// the user's category catalog, splits it into the FIXED set (from the request)
// and the FLEXIBLE set (variable/want categories from the catalog), and runs the
// deterministic budget.SuggestFlexible to produce per-category suggestions. No
// LLM is involved here — this is pure money math.
func (h *Planner) Suggest(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	var req suggestRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}

	answers := suggestAnswers(req)
	if err := answers.Validate(); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", err.Error())
	}

	ctx := c.Request().Context()

	// Resolve the user's category catalog so we can (a) validate the fixed items
	// the user entered and (b) derive the flexible set the planner will suggest.
	cats, err := h.categories.ListForUser(ctx, uid)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "category_lookup_failed", err.Error())
	}
	catByID := make(map[uuid.UUID]repository.ExpenseCategory, len(cats))
	for _, cat := range cats {
		catByID[cat.ID] = cat
	}

	// Fixed (needs/debt) come from the request. Validate each against the catalog:
	// must exist, belong to the user (or be a system default), and carry a fixed/
	// variable/debt type — a 'want' is never a fixed expense.
	fixed := make([]budget.IntakeItem, 0, len(req.FixedItems))
	for i, raw := range req.FixedItems {
		id, err := uuid.Parse(raw.CategoryID)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_category",
				fmt.Sprintf("fixed_items[%d].category_id is not a UUID", i))
		}
		cat, ok := catByID[id]
		if !ok {
			return responses.Err(c, http.StatusBadRequest, "invalid_category",
				fmt.Sprintf("fixed_items[%d].category_id not found for this user", i))
		}
		if budget.ExpenseType(cat.Type) == budget.ExpenseWant {
			return responses.Err(c, http.StatusBadRequest, "invalid_fixed_category",
				fmt.Sprintf("fixed_items[%d] is a 'want' category — only fixed/variable/debt are fixed expenses", i))
		}
		fixed = append(fixed, budget.IntakeItem{
			CategoryID: id,
			Name:       cat.Name,
			Icon:       cat.Icon,
			Type:       budget.ExpenseType(cat.Type),
			Amount:     raw.Amount,
		})
	}

	// Flexible (variable/want) come from the user's catalog, with zero amounts —
	// SuggestFlexible fills in the suggested amounts. We exclude any category the
	// user already pinned as a fixed expense so it isn't double-counted.
	fixedSet := make(map[uuid.UUID]bool, len(fixed))
	for _, it := range fixed {
		fixedSet[it.CategoryID] = true
	}
	flexibleCats := flexibleFromCatalog(cats, fixedSet)

	sug, err := budget.SuggestFlexible(answers, fixed, flexibleCats)
	if err != nil {
		if errors.Is(err, budget.ErrIncomeTooLow) {
			return responses.Err(c, http.StatusBadRequest, "income_too_low",
				"Angka pengeluaran kelihatan tidak wajar dibanding pemasukan — cek lagi angkanya")
		}
		return responses.Err(c, http.StatusBadRequest, "suggest_failed", err.Error())
	}

	return responses.OK(c, suggestionToResponse(sug))
}

// --- POST /v1/planner/chat ------------------------------------------------

type chatItem struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Amount     int64  `json:"amount"`
}

type chatRequest struct {
	Income         int64         `json:"income"`
	Goal           string        `json:"goal"`
	LifestyleStyle string        `json:"lifestyle_style"`
	SavingsTarget  int64         `json:"savings_target"`
	FixedItems     []chatItem    `json:"fixed_items"`
	Flexible       []chatItem    `json:"flexible"`
	Messages       []llm.Message `json:"messages"`
	UserMessage    string        `json:"user_message"`
}

type chatFlexItem struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Amount     int64  `json:"amount"`
}

type chatResponse struct {
	Reply         string         `json:"reply"`
	Flexible      []chatFlexItem `json:"flexible"`
	SavingsTarget int64          `json:"savings_target"`
	Changed       bool           `json:"changed"`
}

// llmAdjustment is one entry the model proposes: WHICH flexible category to
// change and to WHAT target. The app — not the model — computes how that delta
// is absorbed across the other categories (budget.Rebalance).
type llmAdjustment struct {
	CategoryName string `json:"category_name"`
	TargetAmount int64  `json:"target_amount"`
}

// llmChatReply is the STRICT JSON contract our system prompt instructs the model
// to emit. take_from_savings gates whether Rebalance is allowed to reduce the
// savings target (only when the user explicitly says "ambil dari tabungan").
type llmChatReply struct {
	Reply           string          `json:"reply"`
	Adjustments     []llmAdjustment `json:"adjustments"`
	TakeFromSavings bool            `json:"take_from_savings"`
}

// Chat is POST /v1/planner/chat. It is stateless — the frontend holds the thread
// and replays it each turn. The LLM interprets intent and narrates; every number
// change is applied deterministically by budget.Rebalance. On any LLM or parse
// failure we return a friendly reply with changed=false (never a 500 to the
// user) and log the real cause.
func (h *Planner) Chat(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	var req chatRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}
	if req.Income <= 0 {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "income must be > 0")
	}
	if strings.TrimSpace(req.UserMessage) == "" {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "user_message is required")
	}
	// The client owns the working set, but reject negative money outright so a bad
	// payload returns a clean 400 instead of feeding nonsensical numbers into
	// Rebalance. (No persistence happens here, so this only protects the caller.)
	if req.SavingsTarget < 0 {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "savings_target cannot be negative")
	}
	for _, f := range req.Flexible {
		if f.Amount < 0 {
			return responses.Err(c, http.StatusBadRequest, "invalid_payload", "flexible[].amount cannot be negative")
		}
	}

	// Build the live flexible working set the rebalancer mutates. Names are kept
	// so we can resolve the model's category_name (which is text, not an ID) back
	// to a category_id.
	flexible := make([]budget.FlexItem, 0, len(req.Flexible))
	for _, f := range req.Flexible {
		id, err := uuid.Parse(f.CategoryID)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_category",
				"flexible[].category_id is not a UUID")
		}
		// The chat payload carries no per-item type. We tag every flexible item as
		// 'want' so budget.Rebalance treats the whole discretionary set uniformly
		// (all candidates absorb the delta in the first pass), which is correct
		// here: these are exactly the keinginan categories the user is nudging.
		flexible = append(flexible, budget.FlexItem{
			CategoryID: id,
			Name:       f.Name,
			Type:       budget.ExpenseWant,
			Amount:     f.Amount,
		})
	}

	// Build the system prompt + a context message describing the current plan,
	// then the prior thread + the new user message.
	system := plannerSystemPrompt()
	messages := buildChatMessages(req)

	ctx := c.Request().Context()
	raw, err := h.llm.Complete(ctx, system, messages)
	if err != nil {
		c.Logger().Errorf("planner chat: llm complete failed: %v", err)
		return responses.OK(c, chatResponse{
			Reply:         "Maaf, lagi ada gangguan di sisi planner. Coba lagi sebentar ya — atau ubah angkanya langsung.",
			Flexible:      flexItemsToChat(flexible),
			SavingsTarget: req.SavingsTarget,
			Changed:       false,
		})
	}

	parsed, err := parseLLMReply(raw)
	if err != nil {
		c.Logger().Errorf("planner chat: parsing llm reply failed: %v (raw=%q)", err, raw)
		return responses.OK(c, chatResponse{
			Reply:         "Hmm, aku belum nangkap maksudnya. Coba sebutkan kategori + angkanya, misal \"naikin makan jadi 1.5jt\".",
			Flexible:      flexItemsToChat(flexible),
			SavingsTarget: req.SavingsTarget,
			Changed:       false,
		})
	}

	// Apply each proposed adjustment deterministically. The model only chose
	// WHICH category and target; budget.Rebalance computes the absorption.
	savings := req.SavingsTarget
	changed := false
	for _, adj := range parsed.Adjustments {
		idx := resolveFlexByName(flexible, adj.CategoryName)
		if idx == -1 {
			continue // model named a category we don't have in the flexible set; skip
		}
		if adj.TargetAmount < 0 {
			continue
		}
		updated, newSavings, rErr := budget.Rebalance(
			req.Income, savings, flexible,
			flexible[idx].CategoryID, adj.TargetAmount, parsed.TakeFromSavings,
		)
		if rErr != nil {
			c.Logger().Errorf("planner chat: rebalance failed for %q: %v", adj.CategoryName, rErr)
			continue
		}
		flexible = updated
		savings = newSavings
		changed = true
	}

	return responses.OK(c, chatResponse{
		Reply:         parsed.Reply,
		Flexible:      flexItemsToChat(flexible),
		SavingsTarget: savings,
		Changed:       changed,
	})
}

// --- helpers --------------------------------------------------------------

func suggestAnswers(req suggestRequest) budget.IntakeAnswers {
	answers := budget.IntakeAnswers{
		Income:          req.Income,
		HousingType:     budget.HousingType(req.HousingType),
		Goal:            budget.Goal(req.Goal),
		EmergencyMonths: req.EmergencyMonths,
		LifestyleStyle:  budget.LifestyleStyle(req.LifestyleStyle),
	}
	for _, d := range req.DebtTypes {
		answers.DebtTypes = append(answers.DebtTypes, budget.DebtType(d))
	}
	return answers
}

// flexibleFromCatalog returns the user's variable/want categories (the flexible
// set the planner suggests amounts for), skipping any the user already pinned as
// a fixed expense and any inactive categories. Amounts are left at zero —
// SuggestFlexible fills them in.
func flexibleFromCatalog(cats []repository.ExpenseCategory, skip map[uuid.UUID]bool) []budget.IntakeItem {
	out := make([]budget.IntakeItem, 0, len(cats))
	for _, cat := range cats {
		if skip[cat.ID] || !cat.IsActive {
			continue
		}
		t := budget.ExpenseType(cat.Type)
		if t != budget.ExpenseVariable && t != budget.ExpenseWant {
			continue
		}
		out = append(out, budget.IntakeItem{
			CategoryID: cat.ID,
			Name:       cat.Name,
			Icon:       cat.Icon,
			Type:       t,
			Amount:     0,
		})
	}
	return out
}

func suggestionToResponse(sug *budget.Suggestion) suggestResponse {
	resp := suggestResponse{
		Program:       string(sug.Program),
		SavingsTarget: sug.SavingsTarget,
		FixedTotal:    sug.FixedTotal,
		Discretionary: sug.Discretionary,
		Summary: map[string]suggestBucket{
			"kebutuhan": {Amount: sug.Summary.Kebutuhan.Amount, Percentage: sug.Summary.Kebutuhan.Percentage},
			"utang":     {Amount: sug.Summary.Utang.Amount, Percentage: sug.Summary.Utang.Percentage},
			"keinginan": {Amount: sug.Summary.Keinginan.Amount, Percentage: sug.Summary.Keinginan.Percentage},
			"tabungan":  {Amount: sug.Summary.Tabungan.Amount, Percentage: sug.Summary.Tabungan.Percentage},
		},
		Flexible: make([]suggestFlexItem, 0, len(sug.Flexible)),
		Warning:  sug.Warning,
	}
	for _, it := range sug.Flexible {
		resp.Flexible = append(resp.Flexible, suggestFlexItem{
			CategoryID:      it.CategoryID.String(),
			Name:            it.CategoryName,
			Icon:            it.Icon,
			Type:            string(it.Type),
			SuggestedAmount: it.AllocatedAmount,
		})
	}
	return resp
}

func flexItemsToChat(items []budget.FlexItem) []chatFlexItem {
	out := make([]chatFlexItem, 0, len(items))
	for _, it := range items {
		out = append(out, chatFlexItem{
			CategoryID: it.CategoryID.String(),
			Name:       it.Name,
			Amount:     it.Amount,
		})
	}
	return out
}

// resolveFlexByName matches the model's free-text category_name to a flexible
// category by case-insensitive substring (either direction), so "makan",
// "Makan & Minum", and "makan & minum" all resolve. Returns the index or -1.
func resolveFlexByName(items []budget.FlexItem, name string) int {
	target := strings.ToLower(strings.TrimSpace(name))
	if target == "" {
		return -1
	}
	// Exact match first.
	for i := range items {
		if strings.ToLower(items[i].Name) == target {
			return i
		}
	}
	// Then substring either way.
	for i := range items {
		n := strings.ToLower(items[i].Name)
		if strings.Contains(n, target) || strings.Contains(target, n) {
			return i
		}
	}
	return -1
}

// plannerSystemPrompt instructs the model to act as a concise Indonesian
// financial planner and to reply with STRICT JSON. The model decides WHICH
// category to change and to what target; the app re-balances deterministically.
func plannerSystemPrompt() string {
	return `Kamu adalah perencana keuangan pribadi berbahasa Indonesia yang ringkas dan suportif untuk aplikasi Fintrack.

TUGASMU: membaca maksud user yang ingin menyesuaikan budget kategori "keinginan" (flexible), lalu menarasikan trade-off-nya dengan singkat dan ramah. Kamu TIDAK menghitung ulang angka budget — aplikasi yang melakukan itu secara deterministik. Kamu hanya menentukan KATEGORI mana yang diubah dan TARGET angkanya.

ATURAN:
- Hanya boleh mengubah kategori flexible (keinginan), bukan pengeluaran tetap atau target tabungan.
- JANGAN mengurangi target tabungan KECUALI user secara eksplisit minta (mis. "ambil dari tabungan"). Jika ya, set "take_from_savings": true.
- Jika user tidak menyebut kategori atau angka yang jelas, kembalikan "adjustments": [] dan ajukan pertanyaan singkat di "reply".

WAJIB: balas HANYA dengan JSON valid, tanpa teks lain, tanpa code fence. Bentuknya persis:
{"reply": "<kalimat singkat bahasa Indonesia>", "adjustments": [{"category_name": "<nama kategori>", "target_amount": <angka rupiah integer>}], "take_from_savings": <true|false>}`
}

// buildChatMessages assembles the thread sent to the model: the prior turns the
// frontend replayed, a context message describing the current plan numbers, and
// the new user message last (so the stub's last-user-message NLU sees it).
func buildChatMessages(req chatRequest) []llm.Message {
	msgs := make([]llm.Message, 0, len(req.Messages)+2)
	msgs = append(msgs, req.Messages...)

	var b strings.Builder
	b.WriteString("Konteks rencana saat ini (jangan diubah kecuali diminta):\n")
	fmt.Fprintf(&b, "- Pemasukan: Rp %d\n", req.Income)
	fmt.Fprintf(&b, "- Target tabungan: Rp %d\n", req.SavingsTarget)
	if req.Goal != "" {
		fmt.Fprintf(&b, "- Tujuan: %s\n", req.Goal)
	}
	if req.LifestyleStyle != "" {
		fmt.Fprintf(&b, "- Gaya hidup: %s\n", req.LifestyleStyle)
	}
	var fixedTotal int64
	for _, it := range req.FixedItems {
		fixedTotal += it.Amount
	}
	fmt.Fprintf(&b, "- Total pengeluaran tetap: Rp %d\n", fixedTotal)
	b.WriteString("- Kategori flexible saat ini:\n")
	for _, f := range req.Flexible {
		fmt.Fprintf(&b, "  - %s: Rp %d\n", f.Name, f.Amount)
	}
	msgs = append(msgs, llm.Message{Role: "system", Content: b.String()})

	msgs = append(msgs, llm.Message{Role: "user", Content: req.UserMessage})
	return msgs
}

// parseLLMReply robustly parses the model's raw text into the STRICT JSON
// contract. It trims markdown code fences and any prose around the JSON object
// (some models wrap JSON despite instructions). Returns an error if no valid
// JSON object can be recovered.
func parseLLMReply(raw string) (llmChatReply, error) {
	s := strings.TrimSpace(stripCodeFences(raw))

	// Narrow to the outermost JSON object if the model added surrounding prose.
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start == -1 || end == -1 || end < start {
		return llmChatReply{}, fmt.Errorf("no JSON object found in llm reply")
	}
	s = s[start : end+1]

	var out llmChatReply
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return llmChatReply{}, fmt.Errorf("unmarshaling llm reply: %w", err)
	}
	if strings.TrimSpace(out.Reply) == "" {
		out.Reply = "Oke, sudah aku sesuaikan."
	}
	return out, nil
}

// stripCodeFences removes a leading ```json / ``` fence and a trailing ``` if
// the model wrapped its JSON in a markdown block.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	// Drop an optional language tag on the first line (e.g. "json").
	if i := strings.IndexByte(s, '\n'); i != -1 {
		first := strings.TrimSpace(s[:i])
		if first == "" || !strings.ContainsAny(first, "{}") {
			s = s[i+1:]
		}
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}
