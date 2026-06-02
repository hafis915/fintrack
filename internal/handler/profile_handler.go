package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/user"
	"github.com/hafis915/fintrack/internal/handler/dto"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
	v "github.com/hafis915/fintrack/pkg/validator"
)

type ProfileHandler struct{ Svc user.Service }

func (h *ProfileHandler) Get(c echo.Context) error {
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}
	p, err := h.Svc.Get(c.Request().Context(), uid)
	if err != nil {
		return err
	}
	return response.OK(c, toProfileResponse(p))
}

func (h *ProfileHandler) Update(c echo.Context) error {
	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	p, err := h.Svc.UpdateLifestyle(c.Request().Context(), uid, req.LifestyleStyle, req.EmergencyMonths)
	if err != nil {
		return err
	}
	return response.OK(c, toProfileResponse(p))
}

func (h *ProfileHandler) UpdateIncome(c echo.Context) error {
	var req dto.UpdateIncomeRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	hint, err := h.Svc.UpdateIncome(c.Request().Context(), uid, req.Income)
	if err != nil {
		return err
	}
	return response.OK(c, map[string]string{"income_hint": hint})
}

func toProfileResponse(p *user.Profile) dto.ProfileResponse {
	r := dto.ProfileResponse{
		ID:              p.ID.String(),
		EmergencyMonths: p.EmergencyMonths,
		OnboardingDone:  p.OnboardingDone,
	}
	if p.IncomeHint != nil {
		r.IncomeHint = *p.IncomeHint
	}
	if p.HousingType != nil {
		r.HousingType = *p.HousingType
	}
	if p.LifestyleStyle != nil {
		r.LifestyleStyle = *p.LifestyleStyle
	}
	if p.ActiveProgram != nil {
		r.ActiveProgram = *p.ActiveProgram
	}
	return r
}
