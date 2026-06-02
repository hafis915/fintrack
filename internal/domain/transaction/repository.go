package transaction

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*Transaction, error)
	Update(ctx context.Context, in UpdateInput) (*Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	SumSpentByCategoryForPlan(ctx context.Context, userID, planID uuid.UUID) ([]CategorySpent, error)
}
