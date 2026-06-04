package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/pkg/apperr"
)

// User is the domain shape of a user row. Email is the primary secondary
// field; Name is set during local-first register (empty for bootstrap rows).
// Richer profile data lives on UserProfile.
type User struct {
	ID           uuid.UUID
	Email        string
	Name         string
	PasswordHash string // bcrypt hash; empty for legacy/bootstrap rows. Never serialized.
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UsersRepo is the contract handlers depend on. Implementations live below.
type UsersRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, email, name, passwordHash string) (User, error)
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
	return User{
		ID:        fromPgUUID(row.ID),
		Email:     row.Email,
		Name:      derefString(row.Name),
		CreatedAt: fromPgTime(row.CreatedAt),
		UpdatedAt: fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *usersRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, apperr.ErrNotFound
		}
		return User{}, fmt.Errorf("getting user by email: %w", err)
	}
	return User{
		ID:           fromPgUUID(row.ID),
		Email:        row.Email,
		Name:         derefString(row.Name),
		PasswordHash: derefString(row.PasswordHash),
		CreatedAt:    fromPgTime(row.CreatedAt),
		UpdatedAt:    fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *usersRepo) Create(ctx context.Context, email, name, passwordHash string) (User, error) {
	row, err := r.q.CreateUserWithEmail(ctx, generated.CreateUserWithEmailParams{
		Email:        email,
		Name:         ptrOrNil(name),
		PasswordHash: ptrOrNil(passwordHash),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, apperr.ErrAlreadyExists
		}
		return User{}, fmt.Errorf("creating user: %w", err)
	}
	return User{
		ID:        fromPgUUID(row.ID),
		Email:     row.Email,
		Name:      derefString(row.Name),
		CreatedAt: fromPgTime(row.CreatedAt),
		UpdatedAt: fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *usersRepo) Upsert(ctx context.Context, id uuid.UUID, email string) (User, error) {
	row, err := r.q.UpsertUser(ctx, generated.UpsertUserParams{
		ID:    toPgUUID(id),
		Email: email,
	})
	if err != nil {
		return User{}, fmt.Errorf("upserting user: %w", err)
	}
	return User{
		ID:        fromPgUUID(row.ID),
		Email:     row.Email,
		Name:      derefString(row.Name),
		CreatedAt: fromPgTime(row.CreatedAt),
		UpdatedAt: fromPgTime(row.UpdatedAt),
	}, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ptrOrNil returns nil for an empty string so the column is stored as NULL
// rather than an empty string — keeps "no name" and "" indistinguishable at
// the DB layer, which matches the nullable column.
func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505) — used to translate a duplicate-email insert into
// apperr.ErrAlreadyExists instead of a generic 500.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
