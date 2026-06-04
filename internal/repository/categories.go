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

type ExpenseCategory struct {
	ID        uuid.UUID
	UserID    *uuid.UUID // nil for system defaults
	Name      string
	Icon      string
	Type      string // fixed | variable | debt | want
	IsDefault bool
	IsActive  bool
	SortOrder int16
	CreatedAt time.Time
}

type CategoriesRepo interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]ExpenseCategory, error)
	GetByID(ctx context.Context, id uuid.UUID) (ExpenseCategory, error)
	Create(ctx context.Context, p CreateCategoryParams) (ExpenseCategory, error)
}

// CreateCategoryParams is a user-scoped custom expense category — for items
// the default seed doesn't cover. SortOrder is set by the caller so customs
// sort after the system defaults.
type CreateCategoryParams struct {
	UserID    uuid.UUID
	Name      string
	Icon      string
	Type      string // fixed | variable | debt | want
	SortOrder int16
}

type categoriesRepo struct {
	q *generated.Queries
}

func NewCategoriesRepo(pool *pgxpool.Pool) CategoriesRepo {
	return &categoriesRepo{q: generated.New(pool)}
}

func (r *categoriesRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]ExpenseCategory, error) {
	rows, err := r.q.ListExpenseCategoriesForUser(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	out := make([]ExpenseCategory, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCategory(row))
	}
	return out, nil
}

func (r *categoriesRepo) GetByID(ctx context.Context, id uuid.UUID) (ExpenseCategory, error) {
	row, err := r.q.GetExpenseCategoryByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExpenseCategory{}, apperr.ErrNotFound
		}
		return ExpenseCategory{}, fmt.Errorf("getting category: %w", err)
	}
	return toCategory(row), nil
}

func (r *categoriesRepo) Create(ctx context.Context, p CreateCategoryParams) (ExpenseCategory, error) {
	row, err := r.q.CreateCustomExpenseCategory(ctx, generated.CreateCustomExpenseCategoryParams{
		UserID:    toPgUUID(p.UserID),
		Name:      p.Name,
		Icon:      nilIfEmpty(p.Icon),
		Type:      generated.ExpenseCategoryType(p.Type),
		SortOrder: p.SortOrder,
	})
	if err != nil {
		return ExpenseCategory{}, fmt.Errorf("creating category: %w", err)
	}
	return toCategory(row), nil
}

func toCategory(row generated.ExpenseCategory) ExpenseCategory {
	c := ExpenseCategory{
		ID:        fromPgUUID(row.ID),
		Name:      row.Name,
		Type:      string(row.Type),
		IsDefault: row.IsDefault,
		IsActive:  row.IsActive,
		SortOrder: row.SortOrder,
		CreatedAt: fromPgTime(row.CreatedAt),
	}
	if row.UserID.Valid {
		id := fromPgUUID(row.UserID)
		c.UserID = &id
	}
	if row.Icon != nil {
		c.Icon = *row.Icon
	}
	return c
}
