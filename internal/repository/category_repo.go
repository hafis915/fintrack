package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/internal/domain/category"
	"github.com/hafis915/fintrack/pkg/apperror"
)

type categoryRepo struct{ q *db.Queries }

func NewCategoryRepo(pool *pgxpool.Pool) category.Repository {
	return &categoryRepo{q: db.New(pool)}
}

func (r *categoryRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]category.Category, error) {
	rows, err := r.q.ListCategoriesForUser(ctx, toPgUUID(userID))
	if err != nil {
		return nil, apperror.Internal(err)
	}
	out := make([]category.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCategoryDomain(row))
	}
	return out, nil
}

func (r *categoryRepo) Get(ctx context.Context, id uuid.UUID) (*category.Category, error) {
	row, err := r.q.GetCategory(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("category", id.String())
		}
		return nil, apperror.Internal(err)
	}
	c := toCategoryDomain(row)
	return &c, nil
}

func (r *categoryRepo) Create(ctx context.Context, in category.CreateInput) (*category.Category, error) {
	row, err := r.q.CreateCategory(ctx, db.CreateCategoryParams{
		UserID: toPgUUID(in.UserID),
		Name:   in.Name,
		Icon:   in.Icon,
		Type:   in.Type,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	c := toCategoryDomain(row)
	return &c, nil
}

func (r *categoryRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.DeleteCategory(ctx, db.DeleteCategoryParams{
		ID:     toPgUUID(id),
		UserID: toPgUUID(userID),
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func toCategoryDomain(r db.ExpenseCategory) category.Category {
	c := category.Category{
		ID:        fromPgUUID(r.ID),
		Name:      r.Name,
		Icon:      r.Icon,
		Type:      r.Type,
		IsDefault: r.IsDefault,
		IsActive:  r.IsActive,
		SortOrder: int(r.SortOrder),
		CreatedAt: r.CreatedAt.Time,
	}
	if r.UserID.Valid {
		uid := fromPgUUID(r.UserID)
		c.UserID = &uid
	}
	return c
}
