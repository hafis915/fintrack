package integration_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/test/testhelper"
)

// --- helpers ------------------------------------------------------------

// tinyJPEG is a minimal in-memory byte slice used as the image payload. The
// integration server injects ai.NewStubAnalyzer(), which ignores the bytes and
// returns a fixed draft, so the content only needs to be non-empty.
var tinyJPEG = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0xFF, 0xD9}

// multipartImage builds a multipart/form-data body with an "image" file part
// (using the given content type) plus any extra string form fields. It returns
// the encoded body and the matching Content-Type header value.
func multipartImage(t *testing.T, contentType string, data []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", `form-data; name="image"; filename="receipt.jpg"`)
	hdr.Set("Content-Type", contentType)
	part, err := w.CreatePart(hdr)
	if err != nil {
		t.Fatalf("creating image part: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("writing image part: %v", err)
	}

	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("writing field %q: %v", k, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("closing multipart writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}

// receiptDraftItem mirrors the data shape returned by POST /v1/receipts/analyze.
type receiptDraftItem struct {
	Amount       int64   `json:"amount"`
	Merchant     string  `json:"merchant"`
	CategoryID   string  `json:"category_id"`
	CategoryHint string  `json:"category_hint"`
	Confidence   float64 `json:"confidence"`
}

type receiptDraftEnvelope struct {
	Data  *receiptDraftItem `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// receiptTxItem mirrors the transaction shape returned by POST /v1/receipts/confirm.
type receiptTxItem struct {
	ID            string  `json:"id"`
	CategoryID    string  `json:"category_id"`
	Amount        int64   `json:"amount"`
	Note          string  `json:"note"`
	Merchant      string  `json:"merchant"`
	ReceiptURL    string  `json:"receipt_url"`
	AICategorized bool    `json:"ai_categorized"`
	AIConfidence  float64 `json:"ai_confidence"`
	TransactedAt  string  `json:"transacted_at"`
}

type receiptTxEnvelope struct {
	Data  *receiptTxItem `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// --- auth matrix --------------------------------------------------------

func TestIntegration_Receipts_AuthMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "auth@local")

	endpoints := []struct {
		name string
		path string
		// fields included alongside the image part. confirm needs the full set
		// so that a valid token reaches a 2xx rather than a 400 validation error.
		fields map[string]string
		// success is the status code expected for a well-formed authorized call.
		success int
	}{
		{
			name:    "analyze",
			path:    "/v1/receipts/analyze",
			fields:  nil,
			success: http.StatusOK,
		},
		{
			name: "confirm",
			path: "/v1/receipts/confirm",
			fields: map[string]string{
				"amount":        "50000",
				"merchant":      "Indomaret",
				"category_id":   cats["Belanja harian"].String(),
				"transacted_at": time.Now().UTC().Format(time.RFC3339),
				"ai_confidence": "0.95",
			},
			success: http.StatusCreated,
		},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			authCases := []struct {
				name       string
				authHeader string // "" means omit the header entirely
				wantStatus int
			}{
				{"missing_token", "", http.StatusUnauthorized},
				{"malformed_header", "Bearer", http.StatusUnauthorized},
				{"garbage_header", "NotBearer xyz", http.StatusUnauthorized},
				{"invalid_signature", "Bearer " + testhelper.MintTokenWithSecret(t, uid, "wrong-secret-wrong-secret-wrong-secret"), http.StatusUnauthorized},
				{"expired_token", "Bearer " + testhelper.MintExpiredToken(t, uid), http.StatusUnauthorized},
				{"valid_token", "Bearer " + testhelper.MintTokenForUser(t, uid), ep.success},
			}

			for _, ac := range authCases {
				t.Run(ac.name, func(t *testing.T) {
					body, ctype := multipartImage(t, "image/jpeg", tinyJPEG, ep.fields)
					req := httptest.NewRequest(http.MethodPost, ep.path, body)
					req.Header.Set("Content-Type", ctype)
					if ac.authHeader != "" {
						req.Header.Set("Authorization", ac.authHeader)
					}
					rr := httptest.NewRecorder()
					ts.Echo.ServeHTTP(rr, req)
					if rr.Code != ac.wantStatus {
						t.Errorf("POST %s [%s]: want %d, got %d, body=%s",
							ep.path, ac.name, ac.wantStatus, rr.Code, rr.Body.String())
					}
				})
			}
		})
	}
}

// --- analyze ------------------------------------------------------------

func TestIntegration_Receipts_AnalyzeHappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "analyze@local")

	body, ctype := multipartImage(t, "image/jpeg", tinyJPEG, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/receipts/analyze", body)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("analyze: want 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	var env receiptDraftEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decoding analyze response: %v", err)
	}
	if env.Data == nil {
		t.Fatalf("analyze returned no data: %s", rr.Body.String())
	}
	// Stub analyzer returns Amount:50000, Merchant:"Indomaret", Confidence:0.95.
	if env.Data.Amount != 50000 {
		t.Errorf("amount: want 50000, got %d", env.Data.Amount)
	}
	if env.Data.Merchant != "Indomaret" {
		t.Errorf("merchant: want Indomaret, got %q", env.Data.Merchant)
	}
	if env.Data.Confidence != 0.95 {
		t.Errorf("confidence: want 0.95, got %v", env.Data.Confidence)
	}
}

func TestIntegration_Receipts_AnalyzeRejectsBadContentType(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "badtype@local")

	// A text/plain part should be rejected by readImage's content-type check.
	body, ctype := multipartImage(t, "text/plain", []byte("not an image"), nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/receipts/analyze", body)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("analyze bad content-type: want 400, got %d, body=%s", rr.Code, rr.Body.String())
	}
}

// --- confirm ------------------------------------------------------------

func TestIntegration_Receipts_ConfirmHappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	cats := seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "confirm@local")

	transactedAt := time.Now().UTC().Truncate(time.Second)
	fields := map[string]string{
		"amount":        "50000",
		"merchant":      "Indomaret",
		"category_id":   cats["Belanja harian"].String(),
		"note":          "weekly groceries",
		"transacted_at": transactedAt.Format(time.RFC3339),
		"ai_confidence": "0.95",
	}
	body, ctype := multipartImage(t, "image/jpeg", tinyJPEG, fields)
	req := httptest.NewRequest(http.MethodPost, "/v1/receipts/confirm", body)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("confirm: want 201, got %d, body=%s", rr.Code, rr.Body.String())
	}
	var env receiptTxEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decoding confirm response: %v", err)
	}
	if env.Data == nil {
		t.Fatalf("confirm returned no data: %s", rr.Body.String())
	}
	if env.Data.Merchant != "Indomaret" {
		t.Errorf("merchant: want Indomaret, got %q", env.Data.Merchant)
	}
	if env.Data.ReceiptURL == "" {
		t.Errorf("receipt_url should be non-empty, got %q", env.Data.ReceiptURL)
	}
	if !env.Data.AICategorized {
		t.Errorf("ai_categorized should be true")
	}
	if env.Data.Amount != 50000 {
		t.Errorf("amount: want 50000, got %d", env.Data.Amount)
	}

	// The confirmed transaction must surface in the transactions list with its
	// merchant attached.
	listReq := authReq(t, uid, http.MethodGet, "/v1/transactions", nil)
	listRR := httptest.NewRecorder()
	ts.Echo.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list after confirm: want 200, got %d", listRR.Code)
	}

	var raw struct {
		Data *struct {
			Items []struct {
				ID       string `json:"id"`
				Merchant string `json:"merchant"`
				Amount   int64  `json:"amount"`
			} `json:"items"`
			Total int64 `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decoding list: %v", err)
	}
	if raw.Data == nil || raw.Data.Total != 1 || len(raw.Data.Items) != 1 {
		t.Fatalf("list after confirm: want 1 item, got %+v", raw.Data)
	}
	got := raw.Data.Items[0]
	if got.ID != env.Data.ID {
		t.Errorf("list item id: want %s, got %s", env.Data.ID, got.ID)
	}
	if got.Merchant != "Indomaret" {
		t.Errorf("list item merchant: want Indomaret, got %q", got.Merchant)
	}
}

func TestIntegration_Receipts_ConfirmRejectsUnknownCategory(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ts := testhelper.NewTestServer(t)
	defer ts.Close()
	resetTransactionTables(t, ts.DB)
	seedDefaultCategoryIDs(t, ts.DB)

	uid := uuid.New()
	seedUser(t, ts.DB, uid, "confirmbadcat@local")

	fields := map[string]string{
		"amount":        "50000",
		"merchant":      "Indomaret",
		"category_id":   uuid.NewString(), // not a real category
		"transacted_at": time.Now().UTC().Format(time.RFC3339),
		"ai_confidence": "0.95",
	}
	body, ctype := multipartImage(t, "image/jpeg", tinyJPEG, fields)
	req := httptest.NewRequest(http.MethodPost, "/v1/receipts/confirm", body)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("Authorization", "Bearer "+testhelper.MintTokenForUser(t, uid))
	rr := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("confirm unknown category: want 400, got %d, body=%s", rr.Code, rr.Body.String())
	}
}
