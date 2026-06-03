package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/responses"
)

// Categories exposes the read paths the onboarding wizard needs to render
// the expense-item picker. Custom-category create/delete (PRD spec) lands
// in a later slice.
type Categories struct {
	repo repository.CategoriesRepo
}

func NewCategories(repo repository.CategoriesRepo) *Categories {
	return &Categories{repo: repo}
}

type categoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	Type      string `json:"type"`
	IsDefault bool   `json:"is_default"`
	SortOrder int16  `json:"sort_order"`
}

// List handles GET /v1/categories.
func (h *Categories) List(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	cats, err := h.repo.ListForUser(c.Request().Context(), uid)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "list_failed", err.Error())
	}

	out := make([]categoryResponse, 0, len(cats))
	for _, cat := range cats {
		out = append(out, categoryResponse{
			ID:        cat.ID.String(),
			Name:      cat.Name,
			Icon:      cat.Icon,
			Type:      cat.Type,
			IsDefault: cat.IsDefault,
			SortOrder: cat.SortOrder,
		})
	}
	return responses.OK(c, out)
}
