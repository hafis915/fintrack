package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	Create(ctx context.Context, in CreateProfileInput) (*Profile, error)
	UpdateLifestyle(ctx context.Context, userID uuid.UUID, lifestyle *string, emergencyMonths *int) (*Profile, error)
	UpdateIncome(ctx context.Context, userID uuid.UUID, encrypted, hint string) (string, error)
}
