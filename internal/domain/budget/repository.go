package budget

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	UpsertPlan(ctx context.Context, userID uuid.UUID, year, month int, income int64, program string) (*Plan, error)
	UpsertItem(ctx context.Context, planID, categoryID uuid.UUID, amount int64, percentage float64, isDebtFocus bool) error
	GetCurrentPlan(ctx context.Context, userID uuid.UUID, year, month int) (*Plan, error)
	ListItems(ctx context.Context, planID uuid.UUID) ([]Item, error)
}
