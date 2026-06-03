package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/fatigue"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/apperr"
	"github.com/hafis915/fintrack/pkg/responses"
)

// Budget wires GET /v1/budget/current — the Category Fatigue Dashboard.
type Budget struct {
	repo repository.FatigueRepo
	Now  func() time.Time // injectable so tests don't depend on the wall clock
}

func NewBudget(repo repository.FatigueRepo) *Budget {
	return &Budget{repo: repo, Now: time.Now}
}

// --- response shapes ----------------------------------------------------

type budgetItemResponse struct {
	ID              string  `json:"id"`
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	CategoryIcon    string  `json:"category_icon,omitempty"`
	CategoryType    string  `json:"category_type"`
	AllocatedAmount int64   `json:"allocated_amount"`
	SpentAmount     int64   `json:"spent_amount"`
	Remaining       int64   `json:"remaining"`
	PercentageUsed  float64 `json:"percentage_used"`
	Status          string  `json:"status"`
	Coaching        string  `json:"coaching,omitempty"`
	IsDebtFocus     bool    `json:"is_debt_focus,omitempty"`
}

type budgetSummaryResponse struct {
	TotalAllocated    int64   `json:"total_allocated"`
	TotalSpent        int64   `json:"total_spent"`
	UnallocatedSpent  int64   `json:"unallocated_spent,omitempty"`
	OverallPercentage float64 `json:"overall_percentage"`
}

type budgetResponse struct {
	ID          string                `json:"id"`
	Period      string                `json:"period"`
	Program     string                `json:"program"`
	TotalIncome int64                 `json:"total_income"`
	Items       []budgetItemResponse  `json:"items"`
	Summary     budgetSummaryResponse `json:"summary"`
}

// --- handlers -----------------------------------------------------------

// Current handles GET /v1/budget/current.
func (h *Budget) Current(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	now := h.Now()
	year, month := int16(now.Year()), int16(now.Month())

	rows, err := h.repo.GetForPeriod(c.Request().Context(), uid, year, month)
	if errors.Is(err, apperr.ErrNotFound) {
		return responses.Err(c, http.StatusNotFound, "no_active_plan",
			"belum ada budget plan untuk bulan ini — selesaikan onboarding dulu")
	}
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "fetch_failed", err.Error())
	}

	unallocated, err := h.repo.GetUnallocatedSpend(c.Request().Context(), uid, int(year), int(month))
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "unallocated_failed", err.Error())
	}

	plan := rows[0]
	items := make([]budgetItemResponse, 0, len(rows))
	var totalAllocated, totalSpent int64

	for _, r := range rows {
		res := fatigue.Classify(r.AllocatedAmount, r.SpentAmount)
		items = append(items, budgetItemResponse{
			ID:              r.ItemID.String(),
			CategoryID:      r.CategoryID.String(),
			CategoryName:    r.CategoryName,
			CategoryIcon:    r.CategoryIcon,
			CategoryType:    r.CategoryType,
			AllocatedAmount: r.AllocatedAmount,
			SpentAmount:     r.SpentAmount,
			Remaining:       res.Remaining,
			PercentageUsed:  res.PercentageUsed,
			Status:          string(res.Status),
			Coaching:        res.Coaching,
			IsDebtFocus:     r.IsDebtFocus,
		})
		totalAllocated += r.AllocatedAmount
		totalSpent += r.SpentAmount
	}

	totalSpent += unallocated
	summary := fatigue.BuildSummary(totalAllocated, totalSpent, unallocated, plan.TotalIncome)

	return responses.OK(c, budgetResponse{
		ID:          plan.PlanID.String(),
		Period:      fmt.Sprintf("%04d-%02d", plan.PeriodYear, plan.PeriodMonth),
		Program:     plan.Program,
		TotalIncome: plan.TotalIncome,
		Items:       items,
		Summary: budgetSummaryResponse{
			TotalAllocated:    summary.TotalAllocated,
			TotalSpent:        summary.TotalSpent,
			UnallocatedSpent:  summary.UnallocatedSpent,
			OverallPercentage: summary.OverallPercentage,
		},
	})
}
