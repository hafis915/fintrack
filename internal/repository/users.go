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

// User is the domain shape of a user row. Email is the only meaningful
// secondary field today; richer profile data lives on UserProfile.
type User struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UsersRepo is the contract handlers depend on. Implementations live below.
type UsersRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Upsert(ctx context.Context, id uuid.UUID, email string) (User, error)
}

type usersRepo struct {
	q *generated.Queries
}

// NewUsersRepo wraps the sqlc-generated queries against the supplied pool.
func NewUsersRepo(pool *pgxpool.Pool) UsersRepo {
	return &usersRepo{q: generated.New(pool)}
}

func (r *usersRepo) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := r.q.GetUserByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, apperr.ErrNotFound
		}
		return User{}, fmt.Errorf("getting user: %w", err)
	}
	return toUser(row), nil
}

func (r *usersRepo) Upsert(ctx context.Context, id uuid.UUID, email string) (User, error) {
	row, err := r.q.UpsertUser(ctx, generated.UpsertUserParams{
		ID:    toPgUUID(id),
		Email: email,
	})
	if err != nil {
		return User{}, fmt.Errorf("upserting user: %w", err)
	}
	return toUser(row), nil
}

func toUser(row generated.User) User {
	return User{
		ID:        fromPgUUID(row.ID),
		Email:     row.Email,
		CreatedAt: fromPgTime(row.CreatedAt),
		UpdatedAt: fromPgTime(row.UpdatedAt),
	}
}
