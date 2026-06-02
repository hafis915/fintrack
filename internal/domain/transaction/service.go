package transaction

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, in CreateInput) (*Transaction, error)
	Update(ctx context.Context, in UpdateInput) (*Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	SumSpentByCategoryForPlan(ctx context.Context, userID, planID uuid.UUID) ([]CategorySpent, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Create(ctx context.Context, in CreateInput) (*Transaction, error) {
	return s.repo.Create(ctx, in)
}
func (s *service) Update(ctx context.Context, in UpdateInput) (*Transaction, error) {
	return s.repo.Update(ctx, in)
}
func (s *service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}
func (s *service) List(ctx context.Context, filter ListFilter) (*ListResult, error) {
	return s.repo.List(ctx, filter)
}
func (s *service) SumSpentByCategoryForPlan(ctx context.Context, userID, planID uuid.UUID) ([]CategorySpent, error) {
	return s.repo.SumSpentByCategoryForPlan(ctx, userID, planID)
}
