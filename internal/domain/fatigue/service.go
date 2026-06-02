package fatigue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/internal/domain/transaction"
)

type Service interface {
	Snapshot(ctx context.Context, userID uuid.UUID, now time.Time) (*Snapshot, error)
	AlertForCategory(ctx context.Context, userID, categoryID uuid.UUID, now time.Time) (*Alert, error)
}

type service struct {
	budget budget.Service
	tx     transaction.Service
}

func NewService(budgetSvc budget.Service, txSvc transaction.Service) Service {
	return &service{budget: budgetSvc, tx: txSvc}
}

func (s *service) Snapshot(ctx context.Context, userID uuid.UUID, now time.Time) (*Snapshot, error) {
	plan, err := s.budget.GetCurrent(ctx, userID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}
	spentMap := map[uuid.UUID]int64{}
	if sums, err := s.tx.SumSpentByCategoryForPlan(ctx, userID, plan.Plan.ID); err == nil {
		for _, s := range sums {
			spentMap[s.CategoryID] = s.Total
		}
	}

	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	daysRemaining := int(endOfMonth.Sub(now).Hours()/24) + 1
	if daysRemaining < 1 {
		daysRemaining = 1
	}

	cats := make([]CategorySnapshot, 0, len(plan.Items))
	var totalAlloc, totalSpent int64
	for _, it := range plan.Items {
		spent := spentMap[it.CategoryID]
		remaining := it.AllocatedAmount - spent
		status := ComputeStatus(spent, it.AllocatedAmount)
		daily := int64(0)
		if remaining > 0 {
			daily = remaining / int64(daysRemaining)
		}
		cats = append(cats, CategorySnapshot{
			CategoryID:           it.CategoryID,
			CategoryName:         it.CategoryName,
			CategoryIcon:         it.CategoryIcon,
			Type:                 it.CategoryType,
			Allocated:            it.AllocatedAmount,
			Spent:                spent,
			Remaining:            remaining,
			Percentage:           pct(spent, it.AllocatedAmount),
			Status:               status,
			DailyBudgetRemaining: daily,
			Tip:                  tip(it.CategoryName, status, remaining, daysRemaining),
		})
		totalAlloc += it.AllocatedAmount
		totalSpent += spent
	}

	return &Snapshot{
		Period:        fmt.Sprintf("%d-%02d", plan.Plan.PeriodYear, plan.Plan.PeriodMonth),
		DayOfMonth:    now.Day(),
		DaysRemaining: daysRemaining,
		Categories:    cats,
		Overall: Overall{
			TotalAllocated: totalAlloc,
			TotalSpent:     totalSpent,
			Percentage:     pct(totalSpent, totalAlloc),
		},
	}, nil
}

func (s *service) AlertForCategory(ctx context.Context, userID, categoryID uuid.UUID, now time.Time) (*Alert, error) {
	snap, err := s.Snapshot(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	for _, c := range snap.Categories {
		if c.CategoryID == categoryID {
			return &Alert{
				Status:         c.Status,
				CategoryName:   c.CategoryName,
				PercentageUsed: c.Percentage,
				Message:        c.Tip,
			}, nil
		}
	}
	return nil, nil
}

func tip(name, status string, remaining int64, days int) string {
	if status == "fresh" {
		return ""
	}
	return fmt.Sprintf("Budget %s %s. Sisa Rp %d untuk %d hari ke depan.", name, status, max64(remaining, 0), days)
}

func pct(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
