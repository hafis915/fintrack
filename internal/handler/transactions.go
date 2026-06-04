package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/apperr"
	"github.com/hafis915/fintrack/pkg/responses"
)

// Transactions wires POST/GET/PATCH/DELETE /v1/transactions and friends.
type Transactions struct {
	repo       repository.TransactionsRepo
	categories repository.CategoriesRepo
}

func NewTransactions(repo repository.TransactionsRepo, categories repository.CategoriesRepo) *Transactions {
	return &Transactions{repo: repo, categories: categories}
}

// --- request / response shapes ------------------------------------------

type transactionCreateRequest struct {
	CategoryID   string `json:"category_id"`
	Amount       int64  `json:"amount"`
	Note         string `json:"note,omitempty"`
	TransactedAt string `json:"transacted_at"` // RFC3339
}

type transactionUpdateRequest struct {
	CategoryID   *string `json:"category_id,omitempty"`
	Amount       *int64  `json:"amount,omitempty"`
	Note         *string `json:"note,omitempty"`
	TransactedAt *string `json:"transacted_at,omitempty"`
}

type transactionResponse struct {
	ID            string  `json:"id"`
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name,omitempty"`
	CategoryIcon  string  `json:"category_icon,omitempty"`
	CategoryType  string  `json:"category_type,omitempty"`
	Amount        int64   `json:"amount"`
	Note          string  `json:"note,omitempty"`
	Merchant      string  `json:"merchant,omitempty"`
	ReceiptURL    string  `json:"receipt_url,omitempty"`
	AICategorized bool    `json:"ai_categorized"`
	AIConfidence  float64 `json:"ai_confidence,omitempty"`
	TransactedAt  string  `json:"transacted_at"`
	BudgetPlanID  string  `json:"budget_plan_id,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type listTransactionsResponse struct {
	Items []transactionResponse `json:"items"`
	Total int64                 `json:"total"`
}

// --- handlers -----------------------------------------------------------

// Create handles POST /v1/transactions.
func (h *Transactions) Create(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	var req transactionCreateRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}

	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "category_id must be a UUID")
	}
	if req.Amount <= 0 {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "amount must be > 0")
	}
	t, err := time.Parse(time.RFC3339, req.TransactedAt)
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "transacted_at must be RFC3339")
	}

	cat, err := h.categories.GetByID(c.Request().Context(), catID)
	if errors.Is(err, apperr.ErrNotFound) {
		return responses.Err(c, http.StatusBadRequest, "invalid_category", "category_id not found")
	}
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "category_lookup_failed", err.Error())
	}
	if cat.UserID != nil && *cat.UserID != uid {
		return responses.Err(c, http.StatusForbidden, "category_not_owned", "category_id belongs to another user")
	}

	tx, err := h.repo.Create(c.Request().Context(), repository.CreateTransactionParams{
		UserID:       uid,
		CategoryID:   catID,
		Amount:       req.Amount,
		Note:         req.Note,
		TransactedAt: t,
	})
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "create_failed", err.Error())
	}
	return responses.Created(c, transactionToResponse(tx, cat))
}

// Get handles GET /v1/transactions/:id.
func (h *Transactions) Get(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_id", "id must be a UUID")
	}

	tx, err := h.repo.Get(c.Request().Context(), id, uid)
	if errors.Is(err, apperr.ErrNotFound) {
		return responses.Err(c, http.StatusNotFound, "not_found", "transaction not found")
	}
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "get_failed", err.Error())
	}
	return responses.OK(c, transactionToResponse(tx, repository.ExpenseCategory{}))
}

// List handles GET /v1/transactions?category_id=&from=&to=&limit=&offset=.
func (h *Transactions) List(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	params := repository.ListTransactionsParams{UserID: uid}

	if raw := c.QueryParam("category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_filter", "category_id must be a UUID")
		}
		params.CategoryID = &id
	}
	if raw := c.QueryParam("from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_filter", "from must be RFC3339")
		}
		params.From = &t
	}
	if raw := c.QueryParam("to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_filter", "to must be RFC3339")
		}
		params.To = &t
	}
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return responses.Err(c, http.StatusBadRequest, "invalid_filter", "limit must be a positive integer")
		}
		params.Limit = n
	}
	if raw := c.QueryParam("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return responses.Err(c, http.StatusBadRequest, "invalid_filter", "offset must be a non-negative integer")
		}
		params.Offset = n
	}

	txs, total, err := h.repo.List(c.Request().Context(), params)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "list_failed", err.Error())
	}
	items := make([]transactionResponse, 0, len(txs))
	for _, tx := range txs {
		items = append(items, transactionToResponse(tx, repository.ExpenseCategory{}))
	}
	return responses.OK(c, listTransactionsResponse{Items: items, Total: total})
}

// Update handles PATCH /v1/transactions/:id.
func (h *Transactions) Update(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_id", "id must be a UUID")
	}

	var req transactionUpdateRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}

	params := repository.UpdateTransactionParams{ID: id, UserID: uid}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			return responses.Err(c, http.StatusBadRequest, "invalid_payload", "amount must be > 0")
		}
		params.Amount = req.Amount
	}
	if req.Note != nil {
		params.Note = req.Note
	}
	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_payload", "category_id must be a UUID")
		}
		cat, err := h.categories.GetByID(c.Request().Context(), catID)
		if errors.Is(err, apperr.ErrNotFound) {
			return responses.Err(c, http.StatusBadRequest, "invalid_category", "category_id not found")
		}
		if err != nil {
			return responses.Err(c, http.StatusInternalServerError, "category_lookup_failed", err.Error())
		}
		if cat.UserID != nil && *cat.UserID != uid {
			return responses.Err(c, http.StatusForbidden, "category_not_owned", "category_id belongs to another user")
		}
		params.CategoryID = &catID
	}
	if req.TransactedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.TransactedAt)
		if err != nil {
			return responses.Err(c, http.StatusBadRequest, "invalid_payload", "transacted_at must be RFC3339")
		}
		params.TransactedAt = &t
	}

	tx, err := h.repo.Update(c.Request().Context(), params)
	if errors.Is(err, apperr.ErrNotFound) {
		return responses.Err(c, http.StatusNotFound, "not_found", "transaction not found")
	}
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "update_failed", err.Error())
	}
	return responses.OK(c, transactionToResponse(tx, repository.ExpenseCategory{}))
}

// Delete handles DELETE /v1/transactions/:id (soft delete).
func (h *Transactions) Delete(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_id", "id must be a UUID")
	}

	if err := h.repo.SoftDelete(c.Request().Context(), id, uid); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return responses.Err(c, http.StatusNotFound, "not_found", "transaction not found")
		}
		return responses.Err(c, http.StatusInternalServerError, "delete_failed", err.Error())
	}
	return responses.NoContent(c)
}

// --- helpers ------------------------------------------------------------

func transactionToResponse(tx repository.Transaction, _ repository.ExpenseCategory) transactionResponse {
	r := transactionResponse{
		ID:            tx.ID.String(),
		CategoryID:    tx.CategoryID.String(),
		CategoryName:  tx.CategoryName,
		CategoryIcon:  tx.CategoryIcon,
		CategoryType:  tx.CategoryType,
		Amount:        tx.Amount,
		Note:          tx.Note,
		Merchant:      tx.Merchant,
		ReceiptURL:    tx.ReceiptURL,
		AICategorized: tx.AICategorized,
		AIConfidence:  tx.AIConfidence,
		TransactedAt:  tx.TransactedAt.Format(time.RFC3339),
		CreatedAt:     tx.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     tx.UpdatedAt.Format(time.RFC3339),
	}
	if tx.BudgetPlanID != nil {
		r.BudgetPlanID = tx.BudgetPlanID.String()
	}
	_ = fmt.Sprintf // keep import alive in case formatting helpers are inlined
	return r
}
