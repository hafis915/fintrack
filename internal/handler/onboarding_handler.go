package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/internal/handler/dto"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
	v "github.com/hafis915/fintrack/pkg/validator"
)

type OnboardingHandler struct{ Svc budget.Service }

func (h *OnboardingHandler) Submit(c echo.Context) error {
	var req dto.OnboardingRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}

	items := make([]budget.OnboardingItem, 0, len(req.ExpenseItems))
	for _, it := range req.ExpenseItems {
		catID, perr := uuid.Parse(it.CategoryID)
		if perr != nil {
			return apperror.Validation("invalid category_id: "+it.CategoryID, nil)
		}
		items = append(items, budget.OnboardingItem{
			Name: it.Name, Icon: it.Icon, Type: it.Type,
			Amount: it.Amount, CategoryID: catID,
		})
	}

	in := budget.OnboardingInput{
		Income:          req.Income,
		Goal:            req.Goal,
		HousingType:     req.HousingType,
		LifestyleStyle:  req.LifestyleStyle,
		EmergencyMonths: req.EmergencyMonths,
		DebtTypes:       req.DebtTypes,
		ExpenseItems:    items,
	}

	result, err := h.Svc.GenerateFromOnboarding(c.Request().Context(), uid, in)
	if err != nil {
		return err
	}
	return response.Created(c, toOnboardingResponse(result))
}

func toOnboardingResponse(r *budget.OnboardingResult) dto.OnboardingResponse {
	items := make([]dto.BudgetItemResponse, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, dto.BudgetItemResponse{
			CategoryID:      it.CategoryID.String(),
			CategoryName:    it.CategoryName,
			CategoryIcon:    it.CategoryIcon,
			CategoryType:    it.CategoryType,
			AllocatedAmount: it.AllocatedAmount,
			Percentage:      it.Percentage,
			IsDebtFocus:     it.IsDebtFocus,
		})
	}
	return dto.OnboardingResponse{
		BudgetPlanID: r.Plan.ID.String(),
		Program:      r.Program,
		IncomeHint:   r.IncomeHint,
		Warning:      r.Warning,
		Summary: dto.AllocationSummaryResponse{
			Kebutuhan: dto.SummaryGroupResponse{Amount: r.Summary.Kebutuhan.Amount, Percentage: r.Summary.Kebutuhan.Percentage},
			Utang:     dto.SummaryGroupResponse{Amount: r.Summary.Utang.Amount, Percentage: r.Summary.Utang.Percentage},
			Keinginan: dto.SummaryGroupResponse{Amount: r.Summary.Keinginan.Amount, Percentage: r.Summary.Keinginan.Percentage},
			Tabungan:  dto.SummaryGroupResponse{Amount: r.Summary.Tabungan.Amount, Percentage: r.Summary.Tabungan.Percentage},
			Total:     r.Summary.Total,
		},
		Items: items,
	}
}
