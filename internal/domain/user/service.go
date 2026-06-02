package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/internal/encryption"
)

type IncomeEncryptor interface {
	EncryptIncome(amount int64) (string, error)
}

type Service interface {
	Get(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateLifestyle(ctx context.Context, userID uuid.UUID, lifestyle *string, emergencyMonths *int) (*Profile, error)
	UpdateIncome(ctx context.Context, userID uuid.UUID, amount int64) (string, error)
}

type service struct {
	repo Repository
	enc  IncomeEncryptor
}

func NewService(repo Repository, enc IncomeEncryptor) Service {
	return &service{repo: repo, enc: enc}
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*Profile, error) {
	return s.repo.GetByUserID(ctx, id)
}

func (s *service) UpdateLifestyle(ctx context.Context, id uuid.UUID, ls *string, em *int) (*Profile, error) {
	return s.repo.UpdateLifestyle(ctx, id, ls, em)
}

func (s *service) UpdateIncome(ctx context.Context, id uuid.UUID, amount int64) (string, error) {
	cipher, err := s.enc.EncryptIncome(amount)
	if err != nil {
		return "", err
	}
	hint := encryption.MaskIncome(amount)
	return s.repo.UpdateIncome(ctx, id, cipher, hint)
}
