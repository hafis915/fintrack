package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/category"
	"github.com/hafis915/fintrack/internal/handler/dto"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
	v "github.com/hafis915/fintrack/pkg/validator"
)

type CategoryHandler struct{ Svc category.Service }

func (h *CategoryHandler) List(c echo.Context) error {
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}
	cats, err := h.Svc.ListForUser(c.Request().Context(), uid)
	if err != nil {
		return err
	}
	out := make([]dto.CategoryResponse, 0, len(cats))
	for _, cat := range cats {
		out = append(out, toCategoryResponse(cat))
	}
	return response.OK(c, out)
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req dto.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	cat, err := h.Svc.Create(c.Request().Context(), category.CreateInput{
		UserID: uid,
		Name:   req.Name,
		Icon:   req.Icon,
		Type:   req.Type,
	})
	if err != nil {
		return err
	}
	return response.Created(c, toCategoryResponse(*cat))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.Validation("invalid id", nil)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	if err := h.Svc.Delete(c.Request().Context(), id, uid); err != nil {
		return err
	}
	return c.NoContent(204)
}

func toCategoryResponse(cat category.Category) dto.CategoryResponse {
	r := dto.CategoryResponse{
		ID:        cat.ID.String(),
		Name:      cat.Name,
		Type:      cat.Type,
		IsDefault: cat.IsDefault,
		IsActive:  cat.IsActive,
		SortOrder: cat.SortOrder,
	}
	if cat.Icon != nil {
		r.Icon = *cat.Icon
	}
	return r
}
