package transaction

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	BudgetPlanID  *uuid.UUID
	CategoryID    uuid.UUID
	CategoryName  string
	CategoryIcon  string
	CategoryType  string
	Amount        int64
	Note          *string
	ReceiptURL    *string
	AICategorized bool
	AIConfidence  float64
	TransactedAt  time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateInput struct {
	UserID        uuid.UUID
	BudgetPlanID  *uuid.UUID
	CategoryID    uuid.UUID
	Amount        int64
	Note          *string
	ReceiptURL    *string
	AICategorized bool
	AIConfidence  float64
	TransactedAt  time.Time
}

type UpdateInput struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Amount       *int64
	Note         *string
	CategoryID   *uuid.UUID
	TransactedAt *time.Time
}

type ListFilter struct {
	UserID     uuid.UUID
	CategoryID *uuid.UUID
	StartAt    *time.Time
	EndAt      *time.Time
	Limit      int
	Offset     int
}

type ListResult struct {
	Items []Transaction
	Total int
}

type CategorySpent struct {
	CategoryID uuid.UUID
	Total      int64
}
