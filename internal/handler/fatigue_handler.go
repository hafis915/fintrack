package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/fatigue"
	"github.com/hafis915/fintrack/internal/handler/dto"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
)

type FatigueHandler struct{ Svc fatigue.Service }

func (h *FatigueHandler) Dashboard(c echo.Context) error {
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}
	snap, err := h.Svc.Snapshot(c.Request().Context(), uid, time.Now())
	if err != nil {
		return err
	}
	cats := make([]dto.FatigueCategoryResponse, 0, len(snap.Categories))
	for _, cat := range snap.Categories {
		cats = append(cats, dto.FatigueCategoryResponse{
			CategoryID:           cat.CategoryID.String(),
			CategoryName:         cat.CategoryName,
			CategoryIcon:         cat.CategoryIcon,
			Type:                 cat.Type,
			Allocated:            cat.Allocated,
			Spent:                cat.Spent,
			Remaining:            cat.Remaining,
			Percentage:           cat.Percentage,
			Status:               cat.Status,
			DailyBudgetRemaining: cat.DailyBudgetRemaining,
			Tip:                  cat.Tip,
		})
	}
	return response.OK(c, dto.FatigueDashboardResponse{
		Period:        snap.Period,
		DayOfMonth:    snap.DayOfMonth,
		DaysRemaining: snap.DaysRemaining,
		Categories:    cats,
		Overall: dto.FatigueOverallResponse{
			TotalAllocated: snap.Overall.TotalAllocated,
			TotalSpent:     snap.Overall.TotalSpent,
			Percentage:     snap.Overall.Percentage,
		},
	})
}
