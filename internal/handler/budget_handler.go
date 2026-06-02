package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/internal/handler/dto"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
)

type BudgetHandler struct{ Svc budget.Service }

func (h *BudgetHandler) Current(c echo.Context) error {
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}
	now := time.Now()
	plan, err := h.Svc.GetCurrent(c.Request().Context(), uid, now.Year(), int(now.Month()))
	if err != nil {
		return err
	}
	items := make([]dto.BudgetItemResponse, 0, len(plan.Items))
	for _, it := range plan.Items {
		items = append(items, dto.BudgetItemResponse{
			ID:              it.ID.String(),
			CategoryID:      it.CategoryID.String(),
			CategoryName:    it.CategoryName,
			CategoryIcon:    it.CategoryIcon,
			CategoryType:    it.CategoryType,
			AllocatedAmount: it.AllocatedAmount,
			Percentage:      it.Percentage,
			IsDebtFocus:     it.IsDebtFocus,
		})
	}
	return response.OK(c, dto.CurrentBudgetResponse{
		BudgetPlanID: plan.Plan.ID.String(),
		Year:         plan.Plan.PeriodYear,
		Month:        plan.Plan.PeriodMonth,
		Program:      plan.Plan.Program,
		TotalIncome:  plan.Plan.TotalIncome,
		Items:        items,
	})
}
