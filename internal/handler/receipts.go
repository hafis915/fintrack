package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/ai"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/internal/storage"
	"github.com/hafis915/fintrack/pkg/apperr"
	"github.com/hafis915/fintrack/pkg/responses"
)

// maxReceiptBytes caps the decoded image size accepted by the receipt
// endpoints. The route-level BodyLimit is looser (it counts multipart
// overhead); this is the hard limit on the actual image payload.
const maxReceiptBytes = 2 << 20 // 2 MiB

// Receipts wires POST /v1/receipts/analyze and POST /v1/receipts/confirm.
// Analyze runs the image through the AI analyzer and returns a draft; Confirm
// persists a transaction, uploads the image, and records the receipt URL.
type Receipts struct {
	repo       repository.TransactionsRepo
	categories repository.CategoriesRepo
	analyzer   ai.ReceiptAnalyzer
	storage    storage.Storage
}

func NewReceipts(
	repo repository.TransactionsRepo,
	categories repository.CategoriesRepo,
	analyzer ai.ReceiptAnalyzer,
	store storage.Storage,
) *Receipts {
	return &Receipts{repo: repo, categories: categories, analyzer: analyzer, storage: store}
}

// --- response shapes ----------------------------------------------------

type receiptDraftResponse struct {
	Amount       int64   `json:"amount"`
	Merchant     string  `json:"merchant"`
	CategoryID   string  `json:"category_id,omitempty"`
	CategoryHint string  `json:"category_hint"`
	Confidence   float64 `json:"confidence"`
}

// --- handlers -----------------------------------------------------------

// Analyze handles POST /v1/receipts/analyze.
// multipart/form-data field "image" → AI draft + best-effort category match.
func (h *Receipts) Analyze(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	data, mime, errResp := readImage(c)
	if errResp != nil {
		return errResp
	}

	draft, err := h.analyzer.AnalyzeReceipt(c.Request().Context(), data, mime)
	if err != nil {
		if errors.Is(err, ai.ErrAnalyzeFailed) {
			return responses.Err(c, http.StatusBadGateway, "analyze_failed", "could not analyze receipt")
		}
		c.Logger().Error(err)
		return responses.Err(c, http.StatusInternalServerError, "analyze_failed", "could not analyze receipt")
	}

	resp := receiptDraftResponse{
		Amount:       draft.Amount,
		Merchant:     draft.Merchant,
		CategoryHint: draft.CategoryHint,
		Confidence:   draft.Confidence,
	}
	if catID, err := h.matchCategory(c, uid, draft.CategoryHint); err != nil {
		c.Logger().Error(err)
		return responses.Err(c, http.StatusInternalServerError, "category_lookup_failed", "could not match category")
	} else if catID != uuid.Nil {
		resp.CategoryID = catID.String()
	}

	return responses.OK(c, resp)
}

// Confirm handles POST /v1/receipts/confirm.
// multipart/form-data: image + amount + merchant + category_id + note? +
// transacted_at + ai_confidence. Persists the transaction, uploads the image,
// then records the receipt URL. Returns 201 with the full transaction.
func (h *Receipts) Confirm(c echo.Context) error {
	uid := middleware.UserID(c)
	if uid == uuid.Nil {
		return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
	}

	data, contentType, errResp := readImage(c)
	if errResp != nil {
		return errResp
	}

	amount, err := strconv.ParseInt(strings.TrimSpace(c.FormValue("amount")), 10, 64)
	if err != nil || amount <= 0 {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "amount must be a positive integer")
	}
	merchant := strings.TrimSpace(c.FormValue("merchant"))
	if merchant == "" {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "merchant is required")
	}
	catID, err := uuid.Parse(c.FormValue("category_id"))
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "category_id must be a UUID")
	}
	note := strings.TrimSpace(c.FormValue("note"))
	transactedAt, err := time.Parse(time.RFC3339, c.FormValue("transacted_at"))
	if err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "transacted_at must be RFC3339")
	}
	var aiConfidence float64
	if raw := strings.TrimSpace(c.FormValue("ai_confidence")); raw != "" {
		aiConfidence, err = strconv.ParseFloat(raw, 64)
		if err != nil || aiConfidence < 0 || aiConfidence > 1 {
			return responses.Err(c, http.StatusBadRequest, "invalid_payload", "ai_confidence must be a number in [0,1]")
		}
	}

	cat, err := h.categories.GetByID(c.Request().Context(), catID)
	if errors.Is(err, apperr.ErrNotFound) {
		return responses.Err(c, http.StatusBadRequest, "invalid_category", "category_id not found")
	}
	if err != nil {
		c.Logger().Error(err)
		return responses.Err(c, http.StatusInternalServerError, "category_lookup_failed", "could not load category")
	}
	if cat.UserID != nil && *cat.UserID != uid {
		return responses.Err(c, http.StatusForbidden, "category_not_owned", "category_id belongs to another user")
	}

	tx, err := h.repo.Create(c.Request().Context(), repository.CreateTransactionParams{
		UserID:       uid,
		CategoryID:   catID,
		Amount:       amount,
		Note:         note,
		Merchant:     merchant,
		TransactedAt: transactedAt,
	})
	if err != nil {
		c.Logger().Error(err)
		return responses.Err(c, http.StatusInternalServerError, "create_failed", "could not save transaction")
	}

	key := fmt.Sprintf("receipts/%s/%s.jpg", uid.String(), tx.ID.String())
	url, err := h.storage.Upload(c.Request().Context(), key, contentType, data)
	if err != nil {
		// Roll back the just-created transaction so a failed image upload
		// doesn't leave a money-logged-but-receiptless orphan in the ledger.
		c.Logger().Error(err)
		_ = h.repo.SoftDelete(c.Request().Context(), tx.ID, uid)
		return responses.Err(c, http.StatusBadGateway, "upload_failed", "could not store receipt image")
	}

	tx, err = h.repo.SetReceipt(c.Request().Context(), repository.SetReceiptParams{
		ID:           tx.ID,
		UserID:       uid,
		ReceiptURL:   url,
		AIConfidence: aiConfidence,
	})
	if err != nil {
		c.Logger().Error(err)
		_ = h.repo.SoftDelete(c.Request().Context(), tx.ID, uid)
		return responses.Err(c, http.StatusInternalServerError, "set_receipt_failed", "could not finalize receipt")
	}

	return responses.Created(c, transactionToResponse(tx, cat))
}

// --- helpers ------------------------------------------------------------

// readImage extracts the "image" multipart file, validates its content type
// (jpeg/png) and decoded size, and returns its bytes + content type. On
// failure it returns a ready-to-send error response (non-nil third value).
func readImage(c echo.Context) ([]byte, string, error) {
	fh, err := c.FormFile("image")
	if err != nil {
		return nil, "", responses.Err(c, http.StatusBadRequest, "invalid_payload", "image file is required")
	}

	contentType := fh.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, "", responses.Err(c, http.StatusBadRequest, "invalid_image", "image must be JPEG or PNG")
	}

	src, err := fh.Open()
	if err != nil {
		c.Logger().Error(err)
		return nil, "", responses.Err(c, http.StatusInternalServerError, "image_read_failed", "could not read image")
	}
	defer func(f multipart.File) { _ = f.Close() }(src)

	// Read up to the cap + 1 byte so we can detect oversized payloads without
	// buffering an unbounded amount of memory.
	data, err := io.ReadAll(io.LimitReader(src, maxReceiptBytes+1))
	if err != nil {
		c.Logger().Error(err)
		return nil, "", responses.Err(c, http.StatusInternalServerError, "image_read_failed", "could not read image")
	}
	if len(data) > maxReceiptBytes {
		return nil, "", responses.Err(c, http.StatusBadRequest, "image_too_large", "image must be 2MB or smaller")
	}
	if len(data) == 0 {
		return nil, "", responses.Err(c, http.StatusBadRequest, "invalid_image", "image is empty")
	}
	// Sniff the actual bytes — the multipart Content-Type header is
	// client-controlled and can lie. Reject anything that isn't really an image.
	if sniffed := http.DetectContentType(data); !strings.HasPrefix(sniffed, "image/") {
		return nil, "", responses.Err(c, http.StatusBadRequest, "invalid_image", "image must be JPEG or PNG")
	}

	return data, contentType, nil
}

// matchCategory maps an AI category hint to one of the user's expense
// categories via case-insensitive substring matching. Returns uuid.Nil when no
// category matches (callers omit category_id in that case).
func (h *Receipts) matchCategory(c echo.Context, userID uuid.UUID, hint string) (uuid.UUID, error) {
	hint = strings.ToLower(strings.TrimSpace(hint))
	if hint == "" {
		return uuid.Nil, nil
	}
	cats, err := h.categories.ListForUser(c.Request().Context(), userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("listing categories for hint match: %w", err)
	}
	for _, cat := range cats {
		name := strings.ToLower(cat.Name)
		if strings.Contains(name, hint) || strings.Contains(hint, name) {
			return cat.ID, nil
		}
	}
	return uuid.Nil, nil
}
