package category

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Category, error)
	Get(ctx context.Context, id uuid.UUID) (*Category, error)
	Create(ctx context.Context, in CreateInput) (*Category, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
