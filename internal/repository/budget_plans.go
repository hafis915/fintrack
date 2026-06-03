package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/pkg/apperr"
)

type BudgetPlan struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	PeriodYear  int16
	PeriodMonth int16
	TotalIncome int64
	Program     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BudgetItem struct {
	ID              uuid.UUID
	BudgetPlanID    uuid.UUID
	CategoryID      uuid.UUID
	CategoryName    string
	CategoryIcon    string
	CategoryType    string
	AllocatedAmount int64
	Percentage      float64
	IsDebtFocus     bool
}

type CreateBudgetItemParams struct {
	BudgetPlanID    uuid.UUID
	CategoryID      uuid.UUID
	AllocatedAmount int64
	Percentage      float64
	IsDebtFocus     bool
}

type UpsertPlanParams struct {
	UserID      uuid.UUID
	PeriodYear  int16
	PeriodMonth int16
	TotalIncome int64
	Program     string
}

// BudgetPlansRepo owns plans + items. ReplaceItemsForPlan is the onboarding
// operation: nuke any existing items for the plan, insert the new ones.
// Runs inside a single transaction so an interrupted onboarding never
// leaves a half-built plan.
type BudgetPlansRepo interface {
	UpsertPlanWithItems(ctx context.Context, plan UpsertPlanParams, items []CreateBudgetItemParams) (BudgetPlan, []BudgetItem, error)
	GetCurrentForUser(ctx context.Context, userID uuid.UUID, year, month int16) (BudgetPlan, []BudgetItem, error)
}

type budgetPlansRepo struct {
	pool *pgxpool.Pool
}

func NewBudgetPlansRepo(pool *pgxpool.Pool) BudgetPlansRepo {
	return &budgetPlansRepo{pool: pool}
}

func (r *budgetPlansRepo) UpsertPlanWithItems(ctx context.Context, plan UpsertPlanParams, items []CreateBudgetItemParams) (BudgetPlan, []BudgetItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return BudgetPlan{}, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if Commit succeeds

	q := generated.New(tx)

	pgPlan, err := q.UpsertBudgetPlan(ctx, generated.UpsertBudgetPlanParams{
		UserID:      toPgUUID(plan.UserID),
		PeriodYear:  plan.PeriodYear,
		PeriodMonth: plan.PeriodMonth,
		TotalIncome: plan.TotalIncome,
		Program:     generated.FinancialProgram(plan.Program),
	})
	if err != nil {
		return BudgetPlan{}, nil, fmt.Errorf("upserting budget plan: %w", err)
	}

	planID := fromPgUUID(pgPlan.ID)
	if err := q.DeleteBudgetItemsForPlan(ctx, pgPlan.ID); err != nil {
		return BudgetPlan{}, nil, fmt.Errorf("clearing prior items: %w", err)
	}

	created := make([]BudgetItem, 0, len(items))
	for _, it := range items {
		row, err := q.CreateBudgetItem(ctx, generated.CreateBudgetItemParams{
			BudgetPlanID:    toPgUUID(planID),
			CategoryID:      toPgUUID(it.CategoryID),
			AllocatedAmount: it.AllocatedAmount,
			Percentage:      pgNumeric(it.Percentage),
			IsDebtFocus:     it.IsDebtFocus,
		})
		if err != nil {
			return BudgetPlan{}, nil, fmt.Errorf("creating budget item: %w", err)
		}
		created = append(created, BudgetItem{
			ID:              fromPgUUID(row.ID),
			BudgetPlanID:    fromPgUUID(row.BudgetPlanID),
			CategoryID:      fromPgUUID(row.CategoryID),
			AllocatedAmount: row.AllocatedAmount,
			Percentage:      it.Percentage,
			IsDebtFocus:     row.IsDebtFocus,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return BudgetPlan{}, nil, fmt.Errorf("committing tx: %w", err)
	}

	return toBudgetPlan(pgPlan), created, nil
}

func (r *budgetPlansRepo) GetCurrentForUser(ctx context.Context, userID uuid.UUID, year, month int16) (BudgetPlan, []BudgetItem, error) {
	q := generated.New(r.pool)
	row, err := q.GetBudgetPlanForPeriod(ctx, generated.GetBudgetPlanForPeriodParams{
		UserID:      toPgUUID(userID),
		PeriodYear:  year,
		PeriodMonth: month,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BudgetPlan{}, nil, apperr.ErrNotFound
		}
		return BudgetPlan{}, nil, fmt.Errorf("getting plan: %w", err)
	}
	items, err := q.ListBudgetItemsForPlan(ctx, row.ID)
	if err != nil {
		return BudgetPlan{}, nil, fmt.Errorf("listing items: %w", err)
	}
	out := make([]BudgetItem, 0, len(items))
	for _, it := range items {
		icon := ""
		if it.CategoryIcon != nil {
			icon = *it.CategoryIcon
		}
		out = append(out, BudgetItem{
			ID:              fromPgUUID(it.ID),
			BudgetPlanID:    fromPgUUID(it.BudgetPlanID),
			CategoryID:      fromPgUUID(it.CategoryID),
			CategoryName:    it.CategoryName,
			CategoryIcon:    icon,
			CategoryType:    string(it.CategoryType),
			AllocatedAmount: it.AllocatedAmount,
			Percentage:      fromPgNumeric(it.Percentage),
			IsDebtFocus:     it.IsDebtFocus,
		})
	}
	return toBudgetPlan(row), out, nil
}

func toBudgetPlan(row generated.BudgetPlan) BudgetPlan {
	return BudgetPlan{
		ID:          fromPgUUID(row.ID),
		UserID:      fromPgUUID(row.UserID),
		PeriodYear:  row.PeriodYear,
		PeriodMonth: row.PeriodMonth,
		TotalIncome: row.TotalIncome,
		Program:     string(row.Program),
		CreatedAt:   fromPgTime(row.CreatedAt),
		UpdatedAt:   fromPgTime(row.UpdatedAt),
	}
}
