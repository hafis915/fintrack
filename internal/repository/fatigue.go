package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/pkg/apperr"
)

// BudgetWithSpend is one row returned by GET /v1/budget/current. The plan
// header data is duplicated across all rows (denormalised on purpose so
// the handler can read it from any row).
type BudgetWithSpend struct {
	PlanID               uuid.UUID
	UserID               uuid.UUID
	PeriodYear           int16
	PeriodMonth          int16
	TotalIncome          int64
	Program              string
	ItemID               uuid.UUID
	CategoryID           uuid.UUID
	CategoryName         string
	CategoryIcon         string
	CategoryType         string
	AllocatedAmount      int64
	AllocationPercentage float64
	IsDebtFocus          bool
	SpentAmount          int64
}

type FatigueRepo interface {
	GetForPeriod(ctx context.Context, userID uuid.UUID, year, month int16) ([]BudgetWithSpend, error)
	GetUnallocatedSpend(ctx context.Context, userID uuid.UUID, year, month int) (int64, error)
}

type fatigueRepo struct {
	pool *pgxpool.Pool
}

func NewFatigueRepo(pool *pgxpool.Pool) FatigueRepo {
	return &fatigueRepo{pool: pool}
}

func (r *fatigueRepo) GetForPeriod(ctx context.Context, userID uuid.UUID, year, month int16) ([]BudgetWithSpend, error) {
	q := generated.New(r.pool)
	rows, err := q.GetBudgetWithSpend(ctx, generated.GetBudgetWithSpendParams{
		UserID:      toPgUUID(userID),
		PeriodYear:  year,
		PeriodMonth: month,
	})
	if err != nil {
		return nil, fmt.Errorf("listing budget+spend: %w", err)
	}
	if len(rows) == 0 {
		return nil, apperr.ErrNotFound
	}

	out := make([]BudgetWithSpend, 0, len(rows))
	for _, row := range rows {
		icon := ""
		if row.CategoryIcon != nil {
			icon = *row.CategoryIcon
		}
		out = append(out, BudgetWithSpend{
			PlanID:               fromPgUUID(row.PlanID),
			UserID:               fromPgUUID(row.UserID),
			PeriodYear:           row.PeriodYear,
			PeriodMonth:          row.PeriodMonth,
			TotalIncome:          row.TotalIncome,
			Program:              string(row.Program),
			ItemID:               fromPgUUID(row.ItemID),
			CategoryID:           fromPgUUID(row.CategoryID),
			CategoryName:         row.CategoryName,
			CategoryIcon:         icon,
			CategoryType:         string(row.CategoryType),
			AllocatedAmount:      row.AllocatedAmount,
			AllocationPercentage: fromPgNumeric(row.AllocationPercentage),
			IsDebtFocus:          row.IsDebtFocus,
			SpentAmount:          row.SpentAmount,
		})
	}
	return out, nil
}

func (r *fatigueRepo) GetUnallocatedSpend(ctx context.Context, userID uuid.UUID, year, month int) (int64, error) {
	q := generated.New(r.pool)
	v, err := q.GetUnallocatedSpendForPeriod(ctx, generated.GetUnallocatedSpendForPeriodParams{
		UserID:  toPgUUID(userID),
		Column2: int32(year),
		Column3: int32(month),
	})
	if err != nil {
		return 0, fmt.Errorf("counting unallocated spend: %w", err)
	}
	return v, nil
}
