// Package handler holds the HTTP layer: request decoding, validation,
// calling into the domain via repositories, and shaping responses. Handlers
// never touch sqlc or pgx directly — that's the repository's job.
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/internal/encryption"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/apperr"
	"github.com/hafis915/fintrack/pkg/responses"
)

// OnboardingDeps groups the dependencies the handler needs. Wired in
// internal/server when the route is mounted.
type OnboardingDeps struct {
	Users         repository.UsersRepo
	UserProfiles  repository.UserProfilesRepo
	Categories    repository.CategoriesRepo
	BudgetPlans   repository.BudgetPlansRepo
	Cipher        *encryption.Cipher
	Now           func() time.Time // injectable for deterministic period in tests
}

// Onboarding wires POST /v1/onboarding.
type Onboarding struct {
	d OnboardingDeps
}

func NewOnboarding(d OnboardingDeps) *Onboarding {
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Onboarding{d: d}
}

// --- request / response shapes -------------------------------------------

type onboardingRequest struct {
	Income          int64                   `json:"income"`
	HousingType     string                  `json:"housing_type"`
	Goal            string                  `json:"goal"`
	DebtTypes       []string                `json:"debt_types"`
	EmergencyMonths int                     `json:"emergency_months"`
	LifestyleStyle  string                  `json:"lifestyle_style"`
	ExpenseItems    []onboardingExpenseItem `json:"expense_items"`
}

type onboardingExpenseItem struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
}

type onboardingResponseBucket struct {
	Amount     int64   `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type onboardingResponseItem struct {
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	Type            string  `json:"type"`
	Icon            string  `json:"icon,omitempty"`
	AllocatedAmount int64   `json:"allocated_amount"`
	Percentage      float64 `json:"percentage"`
	IsDebtFocus     bool    `json:"is_debt_focus,omitempty"`
}

type onboardingResponse struct {
	Program      string                              `json:"program"`
	BudgetPlanID string                              `json:"budget_plan_id"`
	Period       string                              `json:"period"`
	TotalIncome  int64                               `json:"total_income"`
	Summary      map[string]onboardingResponseBucket `json:"summary"`
	Items        []onboardingResponseItem            `json:"items"`
	Warning      string                              `json:"warning,omitempty"`
}

// --- handler ------------------------------------------------------------

// Handle is POST /v1/onboarding. Steps:
//  1. Decode + validate the payload (PRD contract).
//  2. Confirm every expense_item category exists and belongs to the user (or is a system default).
//  3. Run the program engine + allocator (pure domain).
//  4. Encrypt income, write user_profile + budget_plan + items in one tx.
//  5. Respond with the PRD-shaped envelope.
func (h *Onboarding) Handle(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	var req onboardingRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}

	answers, items, err := req.toDomain()
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", err.Error())
	}
	if err := answers.Validate(); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", err.Error())
	}

	ctx := c.Request().Context()

	// Ensure the user row exists — for local-only dev the JWT has the sub
	// but we may not have the user inserted yet. Email is best-effort
	// derived from sub; Supabase will populate it for real later.
	if _, err := h.d.Users.Upsert(ctx, uid, uid.String()+"@local"); err != nil {
		return responses.Err(c, http.StatusInternalServerError, "user_upsert_failed", err.Error())
	}

	// Validate each item's category exists. System defaults have user_id=null;
	// custom categories must belong to the calling user.
	for i, it := range items {
		cat, err := h.d.Categories.GetByID(ctx, it.CategoryID)
		if errors.Is(err, apperr.ErrNotFound) {
			return responses.Err(c, http.StatusBadRequest, "invalid_category",
				fmt.Sprintf("expense_items[%d].category_id not found", i))
		}
		if err != nil {
			return responses.Err(c, http.StatusInternalServerError, "category_lookup_failed", err.Error())
		}
		if cat.UserID != nil && *cat.UserID != uid {
			return responses.Err(c, http.StatusForbidden, "category_not_owned",
				fmt.Sprintf("expense_items[%d].category_id belongs to another user", i))
		}
		if string(it.Type) != cat.Type {
			return responses.Err(c, http.StatusBadRequest, "category_type_mismatch",
				fmt.Sprintf("expense_items[%d].type=%q does not match category.type=%q", i, it.Type, cat.Type))
		}
	}

	plan, err := budget.Allocate(answers, items)
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "allocation_failed", err.Error())
	}

	// Encrypt income before any DB write touches it.
	cipherText, err := h.d.Cipher.Encrypt([]byte(strconv.FormatInt(answers.Income, 10)))
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "encrypt_failed", err.Error())
	}
	hint := formatIncomeHint(answers.Income)

	if _, err := h.d.UserProfiles.UpsertOnboarding(ctx, repository.UpsertOnboardingParams{
		UserID:          uid,
		IncomeEncrypted: cipherText,
		IncomeHint:      hint,
		HousingType:     string(answers.HousingType),
		LifestyleStyle:  string(answers.LifestyleStyle),
		EmergencyMonths: int16(answers.EmergencyMonths),
		ActiveProgram:   string(plan.Program),
	}); err != nil {
		return responses.Err(c, http.StatusInternalServerError, "profile_upsert_failed", err.Error())
	}

	now := h.d.Now()
	year, month := int16(now.Year()), int16(now.Month())

	itemParams := make([]repository.CreateBudgetItemParams, 0, len(plan.Items))
	for _, it := range plan.Items {
		itemParams = append(itemParams, repository.CreateBudgetItemParams{
			BudgetPlanID:    uuid.Nil, // ignored by repo — set inside tx
			CategoryID:      it.CategoryID,
			AllocatedAmount: it.AllocatedAmount,
			Percentage:      it.Percentage,
			IsDebtFocus:     it.IsDebtFocus,
		})
	}

	persistedPlan, _, err := h.d.BudgetPlans.UpsertPlanWithItems(ctx,
		repository.UpsertPlanParams{
			UserID:      uid,
			PeriodYear:  year,
			PeriodMonth: month,
			TotalIncome: answers.Income,
			Program:     string(plan.Program),
		}, itemParams)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "plan_persist_failed", err.Error())
	}

	return responses.Created(c, planToResponse(persistedPlan, plan))
}

// --- helpers ------------------------------------------------------------

func (req onboardingRequest) toDomain() (budget.IntakeAnswers, []budget.IntakeItem, error) {
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

	items := make([]budget.IntakeItem, 0, len(req.ExpenseItems))
	for i, raw := range req.ExpenseItems {
		id, err := uuid.Parse(raw.CategoryID)
		if err != nil {
			return budget.IntakeAnswers{}, nil, fmt.Errorf("expense_items[%d].category_id is not a UUID", i)
		}
		items = append(items, budget.IntakeItem{
			CategoryID: id,
			Name:       raw.Name,
			Icon:       raw.Icon,
			Type:       budget.ExpenseType(raw.Type),
			Amount:     raw.Amount,
		})
	}
	return answers, items, nil
}

// formatIncomeHint produces the safe "Rp 8jt" / "Rp 8,5jt" digest shown
// in the UI. Buckets: < 1jt → "<1jt", < 10jt → "Rp Njt" / "Rp N,5jt",
// >= 10jt → "Rp Njt". This is intentionally low-resolution so the hint
// can't be reverse-engineered into exact income.
func formatIncomeHint(rupiah int64) string {
	switch {
	case rupiah < 1_000_000:
		return "<Rp 1jt"
	case rupiah < 10_000_000:
		whole := rupiah / 1_000_000
		half := (rupiah % 1_000_000) >= 500_000
		if half {
			return "Rp " + strconv.FormatInt(whole, 10) + ",5jt"
		}
		return "Rp " + strconv.FormatInt(whole, 10) + "jt"
	default:
		whole := rupiah / 1_000_000
		return "Rp " + strconv.FormatInt(whole, 10) + "jt"
	}
}

func planToResponse(persisted repository.BudgetPlan, computed *budget.Plan) onboardingResponse {
	resp := onboardingResponse{
		Program:      string(computed.Program),
		BudgetPlanID: persisted.ID.String(),
		Period:       fmt.Sprintf("%04d-%02d", persisted.PeriodYear, persisted.PeriodMonth),
		TotalIncome:  computed.TotalIncome,
		Summary: map[string]onboardingResponseBucket{
			"kebutuhan": {Amount: computed.Summary.Kebutuhan.Amount, Percentage: computed.Summary.Kebutuhan.Percentage},
			"utang":     {Amount: computed.Summary.Utang.Amount, Percentage: computed.Summary.Utang.Percentage},
			"keinginan": {Amount: computed.Summary.Keinginan.Amount, Percentage: computed.Summary.Keinginan.Percentage},
			"tabungan":  {Amount: computed.Summary.Tabungan.Amount, Percentage: computed.Summary.Tabungan.Percentage},
		},
		Items:   make([]onboardingResponseItem, 0, len(computed.Items)),
		Warning: computed.Warning,
	}
	for _, it := range computed.Items {
		resp.Items = append(resp.Items, onboardingResponseItem{
			CategoryID:      it.CategoryID.String(),
			CategoryName:    it.CategoryName,
			Type:            string(it.Type),
			Icon:            it.Icon,
			AllocatedAmount: it.AllocatedAmount,
			Percentage:      it.Percentage,
			IsDebtFocus:     it.IsDebtFocus,
		})
	}
	return resp
}
