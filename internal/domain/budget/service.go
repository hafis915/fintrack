package budget

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/internal/domain/user"
	"github.com/hafis915/fintrack/pkg/apperror"
)

type Service interface {
	GenerateFromOnboarding(ctx context.Context, userID uuid.UUID, in OnboardingInput) (*OnboardingResult, error)
	GetCurrent(ctx context.Context, userID uuid.UUID, year, month int) (*PlanWithItems, error)
}

type OnboardingResult struct {
	Plan       Plan
	Items      []Item
	Summary    AllocationSummary
	Program    string
	Warning    string
	IncomeHint string
}

type service struct {
	repo     Repository
	userSvc  user.Service
	userRepo user.Repository
}

func NewService(repo Repository, userSvc user.Service, userRepo user.Repository) Service {
	return &service{repo: repo, userSvc: userSvc, userRepo: userRepo}
}

func (s *service) GenerateFromOnboarding(ctx context.Context, userID uuid.UUID, in OnboardingInput) (*OnboardingResult, error) {
	alloc, err := GenerateAllocation(in)
	if err != nil {
		return nil, apperror.Validation(err.Error(), nil)
	}

	hint, err := s.userSvc.UpdateIncome(ctx, userID, in.Income)
	if err != nil {
		// no profile exists yet — create one
		if errors.As(err, new(*apperror.Error)) {
			_, cerr := s.userRepo.Create(ctx, user.CreateProfileInput{
				UserID:          userID,
				HousingType:     in.HousingType,
				LifestyleStyle:  in.LifestyleStyle,
				EmergencyMonths: in.EmergencyMonths,
				ActiveProgram:   alloc.Program,
				OnboardingDone:  true,
			})
			if cerr != nil {
				return nil, cerr
			}
			hint, err = s.userSvc.UpdateIncome(ctx, userID, in.Income)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	now := time.Now()
	year, month := now.Year(), int(now.Month())

	plan, err := s.repo.UpsertPlan(ctx, userID, year, month, in.Income, alloc.Program)
	if err != nil {
		return nil, err
	}

	items := make([]Item, 0, len(alloc.Items))
	for _, ai := range alloc.Items {
		if err := s.repo.UpsertItem(ctx, plan.ID, ai.CategoryID, ai.AllocatedAmount, ai.Percentage, ai.IsDebtFocus); err != nil {
			return nil, err
		}
		items = append(items, Item{
			BudgetPlanID:    plan.ID,
			CategoryID:      ai.CategoryID,
			CategoryName:    ai.CategoryName,
			CategoryIcon:    ai.Icon,
			CategoryType:    ai.Type,
			AllocatedAmount: ai.AllocatedAmount,
			Percentage:      ai.Percentage,
			IsDebtFocus:     ai.IsDebtFocus,
		})
	}

	return &OnboardingResult{
		Plan:       *plan,
		Items:      items,
		Summary:    alloc.Summary,
		Program:    alloc.Program,
		Warning:    alloc.Warning,
		IncomeHint: hint,
	}, nil
}

func (s *service) GetCurrent(ctx context.Context, userID uuid.UUID, year, month int) (*PlanWithItems, error) {
	plan, err := s.repo.GetCurrentPlan(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListItems(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	return &PlanWithItems{Plan: *plan, Items: items}, nil
}
