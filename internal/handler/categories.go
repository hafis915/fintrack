package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/responses"
)

// customCategorySortOrder places user-created categories after every system
// default (defaults seed in the low double digits).
const customCategorySortOrder = 900

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
		out = append(out, toCategoryResponse(cat))
	}
	return responses.OK(c, out)
}

type categoryCreateRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	Type string `json:"type"`
}

var validCategoryTypes = map[string]bool{"fixed": true, "variable": true, "debt": true, "want": true}

// Create handles POST /v1/categories — a user-scoped custom expense category
// for anything the default seed doesn't cover (used by the onboarding wizard).
func (h *Categories) Create(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	var req categoryCreateRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "name is required")
	}
	if !validCategoryTypes[req.Type] {
		return responses.Err(c, http.StatusBadRequest, "invalid_type",
			"type must be one of fixed, variable, debt, want")
	}

	cat, err := h.repo.Create(c.Request().Context(), repository.CreateCategoryParams{
		UserID:    uid,
		Name:      req.Name,
		Icon:      strings.TrimSpace(req.Icon),
		Type:      req.Type,
		SortOrder: customCategorySortOrder,
	})
	if err != nil {
		c.Logger().Error(err)
		return responses.Err(c, http.StatusInternalServerError, "create_failed", "could not create category")
	}
	return responses.Created(c, toCategoryResponse(cat))
}

func toCategoryResponse(cat repository.ExpenseCategory) categoryResponse {
	return categoryResponse{
		ID:        cat.ID.String(),
		Name:      cat.Name,
		Icon:      cat.Icon,
		Type:      cat.Type,
		IsDefault: cat.IsDefault,
		SortOrder: cat.SortOrder,
	}
}
