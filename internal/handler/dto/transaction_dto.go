package dto

import "time"

type TransactionResponse struct {
	ID            string    `json:"id"`
	CategoryID    string    `json:"category_id"`
	CategoryName  string    `json:"category_name,omitempty"`
	CategoryIcon  string    `json:"category_icon,omitempty"`
	CategoryType  string    `json:"category_type,omitempty"`
	Amount        int64     `json:"amount"`
	Note          string    `json:"note,omitempty"`
	ReceiptURL    string    `json:"receipt_url,omitempty"`
	AICategorized bool      `json:"ai_categorized"`
	AIConfidence  float64   `json:"ai_confidence,omitempty"`
	TransactedAt  time.Time `json:"transacted_at"`
}

type CreateTransactionRequest struct {
	CategoryID   string     `json:"category_id"   validate:"required,uuid"`
	Amount       int64      `json:"amount"        validate:"required,gt=0"`
	Note         string     `json:"note"`
	TransactedAt *time.Time `json:"transacted_at"`
}

type UpdateTransactionRequest struct {
	CategoryID   *string    `json:"category_id"   validate:"omitempty,uuid"`
	Amount       *int64     `json:"amount"        validate:"omitempty,gt=0"`
	Note         *string    `json:"note"`
	TransactedAt *time.Time `json:"transacted_at"`
}

type ListTransactionsResponse struct {
	Items   []TransactionResponse `json:"items"`
	Total   int                   `json:"total"`
	Page    int                   `json:"page"`
	PerPage int                   `json:"per_page"`
}

type ReceiptScanAlternative struct {
	CategoryName string  `json:"category_name"`
	Confidence   float64 `json:"confidence"`
}

type ReceiptScanResponse struct {
	Amount                int64                    `json:"amount"`
	SuggestedCategoryID   string                   `json:"suggested_category_id,omitempty"`
	SuggestedCategoryName string                   `json:"suggested_category_name"`
	Note                  string                   `json:"note,omitempty"`
	Confidence            float64                  `json:"confidence"`
	ReceiptURL            string                   `json:"receipt_url,omitempty"`
	Alternatives          []ReceiptScanAlternative `json:"alternatives,omitempty"`
}
