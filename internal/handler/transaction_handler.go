package handler

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/internal/domain/fatigue"
	"github.com/hafis915/fintrack/internal/domain/transaction"
	"github.com/hafis915/fintrack/internal/handler/dto"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
	v "github.com/hafis915/fintrack/pkg/validator"
)

type TransactionHandler struct {
	Svc     transaction.Service
	Budget  budget.Service
	Fatigue fatigue.Service
}

func (h *TransactionHandler) List(c echo.Context) error {
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}
	filter := transaction.ListFilter{
		UserID: uid,
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}
	if cat := c.QueryParam("category_id"); cat != "" {
		cid, err := uuid.Parse(cat)
		if err != nil {
			return apperror.Validation("invalid category_id", nil)
		}
		filter.CategoryID = &cid
	}
	res, err := h.Svc.List(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	items := make([]dto.TransactionResponse, 0, len(res.Items))
	for _, tx := range res.Items {
		items = append(items, toTxResponse(tx))
	}
	return response.OK(c, dto.ListTransactionsResponse{
		Items: items, Total: res.Total, Page: page, PerPage: perPage,
	})
}

func (h *TransactionHandler) Create(c echo.Context) error {
	var req dto.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	catID, _ := uuid.Parse(req.CategoryID)
	when := time.Now()
	if req.TransactedAt != nil {
		when = *req.TransactedAt
	}
	in := transaction.CreateInput{
		UserID:       uid,
		CategoryID:   catID,
		Amount:       req.Amount,
		TransactedAt: when,
	}
	if req.Note != "" {
		in.Note = &req.Note
	}
	// Attach to current budget plan if one exists for the transaction's period
	if h.Budget != nil {
		if plan, perr := h.Budget.GetCurrent(c.Request().Context(), uid, when.Year(), int(when.Month())); perr == nil && plan != nil {
			planID := plan.Plan.ID
			in.BudgetPlanID = &planID
		}
	}
	tx, err := h.Svc.Create(c.Request().Context(), in)
	if err != nil {
		return err
	}
	resp := toTxResponse(*tx)
	if h.Fatigue != nil {
		if alert, _ := h.Fatigue.AlertForCategory(c.Request().Context(), uid, tx.CategoryID, time.Now()); alert != nil && alert.Status != "fresh" {
			return response.Created(c, map[string]any{
				"transaction":   resp,
				"fatigue_alert": dto.FatigueAlertResponse{Status: alert.Status, CategoryName: alert.CategoryName, PercentageUsed: alert.PercentageUsed, Message: alert.Message},
			})
		}
	}
	return response.Created(c, resp)
}

func (h *TransactionHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.Validation("invalid id", nil)
	}
	var req dto.UpdateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	in := transaction.UpdateInput{ID: id, UserID: uid, Amount: req.Amount, Note: req.Note, TransactedAt: req.TransactedAt}
	if req.CategoryID != nil {
		cid, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return apperror.Validation("invalid category_id", nil)
		}
		in.CategoryID = &cid
	}
	tx, err := h.Svc.Update(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return response.OK(c, toTxResponse(*tx))
}

func (h *TransactionHandler) Delete(c echo.Context) error {
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

func toTxResponse(tx transaction.Transaction) dto.TransactionResponse {
	r := dto.TransactionResponse{
		ID:            tx.ID.String(),
		CategoryID:    tx.CategoryID.String(),
		CategoryName:  tx.CategoryName,
		CategoryIcon:  tx.CategoryIcon,
		CategoryType:  tx.CategoryType,
		Amount:        tx.Amount,
		AICategorized: tx.AICategorized,
		AIConfidence:  tx.AIConfidence,
		TransactedAt:  tx.TransactedAt,
	}
	if tx.Note != nil {
		r.Note = *tx.Note
	}
	if tx.ReceiptURL != nil {
		r.ReceiptURL = *tx.ReceiptURL
	}
	return r
}
