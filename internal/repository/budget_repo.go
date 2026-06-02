package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/internal/domain/budget"
	"github.com/hafis915/fintrack/pkg/apperror"
)

type budgetRepo struct{ q *db.Queries }

func NewBudgetRepo(pool *pgxpool.Pool) budget.Repository {
	return &budgetRepo{q: db.New(pool)}
}

func (r *budgetRepo) UpsertPlan(ctx context.Context, userID uuid.UUID, year, month int, income int64, program string) (*budget.Plan, error) {
	row, err := r.q.CreateBudgetPlan(ctx, db.CreateBudgetPlanParams{
		UserID:      toPgUUID(userID),
		PeriodYear:  int16(year),
		PeriodMonth: int16(month),
		TotalIncome: income,
		Program:     program,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &budget.Plan{
		ID:          fromPgUUID(row.ID),
		UserID:      fromPgUUID(row.UserID),
		PeriodYear:  int(row.PeriodYear),
		PeriodMonth: int(row.PeriodMonth),
		TotalIncome: row.TotalIncome,
		Program:     row.Program,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *budgetRepo) UpsertItem(ctx context.Context, planID, categoryID uuid.UUID, amount int64, percentage float64, isDebtFocus bool) error {
	pct := pgtype.Numeric{}
	if err := pct.Scan(formatPct(percentage)); err != nil {
		return apperror.Internal(err)
	}
	_, err := r.q.CreateBudgetItem(ctx, db.CreateBudgetItemParams{
		BudgetPlanID:    toPgUUID(planID),
		CategoryID:      toPgUUID(categoryID),
		AllocatedAmount: amount,
		Percentage:      pct,
		IsDebtFocus:     isDebtFocus,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *budgetRepo) GetCurrentPlan(ctx context.Context, userID uuid.UUID, year, month int) (*budget.Plan, error) {
	row, err := r.q.GetCurrentBudgetPlan(ctx, db.GetCurrentBudgetPlanParams{
		UserID:      toPgUUID(userID),
		PeriodYear:  int16(year),
		PeriodMonth: int16(month),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("budget_plan", "current")
		}
		return nil, apperror.Internal(err)
	}
	return &budget.Plan{
		ID:          fromPgUUID(row.ID),
		UserID:      fromPgUUID(row.UserID),
		PeriodYear:  int(row.PeriodYear),
		PeriodMonth: int(row.PeriodMonth),
		TotalIncome: row.TotalIncome,
		Program:     row.Program,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *budgetRepo) ListItems(ctx context.Context, planID uuid.UUID) ([]budget.Item, error) {
	rows, err := r.q.ListBudgetItemsWithCategory(ctx, toPgUUID(planID))
	if err != nil {
		return nil, apperror.Internal(err)
	}
	items := make([]budget.Item, 0, len(rows))
	for _, row := range rows {
		pctF, _ := pgNumericToFloat(row.Percentage)
		icon := ""
		if row.CategoryIcon != nil {
			icon = *row.CategoryIcon
		}
		items = append(items, budget.Item{
			ID:              fromPgUUID(row.ID),
			BudgetPlanID:    fromPgUUID(row.BudgetPlanID),
			CategoryID:      fromPgUUID(row.CategoryID),
			CategoryName:    row.CategoryName,
			CategoryIcon:    icon,
			CategoryType:    row.CategoryType,
			AllocatedAmount: row.AllocatedAmount,
			Percentage:      pctF,
			IsDebtFocus:     row.IsDebtFocus,
		})
	}
	return items, nil
}

func formatPct(p float64) string {
	return strconv.FormatFloat(p, 'f', 2, 64)
}

func pgNumericToFloat(n pgtype.Numeric) (float64, error) {
	if !n.Valid {
		return 0, nil
	}
	f, err := n.Float64Value()
	if err != nil {
		return 0, err
	}
	return f.Float64, nil
}
