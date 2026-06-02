package category

import (
	"context"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/pkg/apperror"
)

type Service interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Category, error)
	Create(ctx context.Context, in CreateInput) (*Category, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ListForUser(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *service) Create(ctx context.Context, in CreateInput) (*Category, error) {
	return s.repo.Create(ctx, in)
}

func (s *service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	cat, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if cat.IsDefault {
		return apperror.Forbidden("cannot delete default system category")
	}
	if cat.UserID == nil || *cat.UserID != userID {
		return apperror.Forbidden("category not owned by user")
	}
	return s.repo.Delete(ctx, id, userID)
}
