package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hafis915/fintrack/database/sqlc/generated"
	"github.com/hafis915/fintrack/pkg/apperr"
)

type Transaction struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	BudgetPlanID  *uuid.UUID
	CategoryID    uuid.UUID
	CategoryName  string
	CategoryIcon  string
	CategoryType  string
	Amount        int64
	Note          string
	Merchant      string
	ReceiptURL    string
	AICategorized bool
	AIConfidence  float64
	TransactedAt  time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateTransactionParams struct {
	UserID        uuid.UUID
	CategoryID    uuid.UUID
	Amount        int64
	Note          string
	Merchant      string
	TransactedAt  time.Time
	ReceiptURL    string
	AICategorized bool
	AIConfidence  float64
}

type SetReceiptParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ReceiptURL   string
	AIConfidence float64
}

type ListTransactionsParams struct {
	UserID     uuid.UUID
	CategoryID *uuid.UUID
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

type UpdateTransactionParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Amount       *int64
	Note         *string
	CategoryID   *uuid.UUID
	TransactedAt *time.Time
}

type TransactionsRepo interface {
	Create(ctx context.Context, p CreateTransactionParams) (Transaction, error)
	Get(ctx context.Context, id, userID uuid.UUID) (Transaction, error)
	List(ctx context.Context, p ListTransactionsParams) ([]Transaction, int64, error)
	Update(ctx context.Context, p UpdateTransactionParams) (Transaction, error)
	SetReceipt(ctx context.Context, p SetReceiptParams) (Transaction, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
}

type transactionsRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionsRepo(pool *pgxpool.Pool) TransactionsRepo {
	return &transactionsRepo{pool: pool}
}

func (r *transactionsRepo) Create(ctx context.Context, p CreateTransactionParams) (Transaction, error) {
	q := generated.New(r.pool)

	// Auto-link to the user's plan for the same calendar month, if one exists.
	planID, err := q.GetActiveBudgetPlanForDate(ctx, generated.GetActiveBudgetPlanForDateParams{
		UserID:  toPgUUID(p.UserID),
		Column2: toPgTime(p.TransactedAt),
	})
	var pgPlanID pgtype.UUID
	if err == nil {
		pgPlanID = planID
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, fmt.Errorf("looking up plan: %w", err)
	}

	createParams := generated.CreateTransactionParams{
		UserID:        toPgUUID(p.UserID),
		BudgetPlanID:  pgPlanID,
		CategoryID:    toPgUUID(p.CategoryID),
		Amount:        p.Amount,
		Note:          nilIfEmpty(p.Note),
		Merchant:      nilIfEmpty(p.Merchant),
		ReceiptUrl:    nilIfEmpty(p.ReceiptURL),
		AiCategorized: p.AICategorized,
		TransactedAt:  toPgTime(p.TransactedAt),
	}
	if p.AIConfidence > 0 {
		createParams.AiConfidence = pgNumeric(p.AIConfidence)
	}
	row, err := q.CreateTransaction(ctx, createParams)
	if err != nil {
		return Transaction{}, fmt.Errorf("creating transaction: %w", err)
	}
	return Transaction{
		ID:            fromPgUUID(row.ID),
		UserID:        fromPgUUID(row.UserID),
		BudgetPlanID:  optUUID(row.BudgetPlanID),
		CategoryID:    fromPgUUID(row.CategoryID),
		Amount:        row.Amount,
		Note:          strOrEmpty(row.Note),
		Merchant:      strOrEmpty(row.Merchant),
		ReceiptURL:    strOrEmpty(row.ReceiptUrl),
		AICategorized: row.AiCategorized,
		AIConfidence:  fromPgNumeric(row.AiConfidence),
		TransactedAt:  fromPgTime(row.TransactedAt),
		CreatedAt:     fromPgTime(row.CreatedAt),
		UpdatedAt:     fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *transactionsRepo) Get(ctx context.Context, id, userID uuid.UUID) (Transaction, error) {
	q := generated.New(r.pool)
	row, err := q.GetTransactionForUser(ctx, generated.GetTransactionForUserParams{
		ID:     toPgUUID(id),
		UserID: toPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, apperr.ErrNotFound
		}
		return Transaction{}, fmt.Errorf("getting transaction: %w", err)
	}
	return Transaction{
		ID:            fromPgUUID(row.ID),
		UserID:        fromPgUUID(row.UserID),
		BudgetPlanID:  optUUID(row.BudgetPlanID),
		CategoryID:    fromPgUUID(row.CategoryID),
		CategoryName:  row.CategoryName,
		CategoryIcon:  strOrEmpty(row.CategoryIcon),
		CategoryType:  string(row.CategoryType),
		Amount:        row.Amount,
		Note:          strOrEmpty(row.Note),
		Merchant:      strOrEmpty(row.Merchant),
		ReceiptURL:    strOrEmpty(row.ReceiptUrl),
		AICategorized: row.AiCategorized,
		AIConfidence:  fromPgNumeric(row.AiConfidence),
		TransactedAt:  fromPgTime(row.TransactedAt),
		CreatedAt:     fromPgTime(row.CreatedAt),
		UpdatedAt:     fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *transactionsRepo) List(ctx context.Context, p ListTransactionsParams) ([]Transaction, int64, error) {
	q := generated.New(r.pool)
	if p.Limit <= 0 {
		p.Limit = 50
	}
	if p.Limit > 200 {
		p.Limit = 200
	}

	listParams := generated.ListTransactionsForUserParams{
		UserID: toPgUUID(p.UserID),
		Limit:  int32(p.Limit),
		Offset: int32(p.Offset),
	}
	if p.CategoryID != nil {
		listParams.Column2 = toPgUUID(*p.CategoryID)
	}
	if p.From != nil {
		listParams.Column3 = toPgTime(*p.From)
	}
	if p.To != nil {
		listParams.Column4 = toPgTime(*p.To)
	}

	rows, err := q.ListTransactionsForUser(ctx, listParams)
	if err != nil {
		return nil, 0, fmt.Errorf("listing transactions: %w", err)
	}

	countParams := generated.CountTransactionsForUserParams{
		UserID:  listParams.UserID,
		Column2: listParams.Column2,
		Column3: listParams.Column3,
		Column4: listParams.Column4,
	}
	total, err := q.CountTransactionsForUser(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	out := make([]Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, Transaction{
			ID:            fromPgUUID(row.ID),
			UserID:        fromPgUUID(row.UserID),
			BudgetPlanID:  optUUID(row.BudgetPlanID),
			CategoryID:    fromPgUUID(row.CategoryID),
			CategoryName:  row.CategoryName,
			CategoryIcon:  strOrEmpty(row.CategoryIcon),
			CategoryType:  string(row.CategoryType),
			Amount:        row.Amount,
			Note:          strOrEmpty(row.Note),
			Merchant:      strOrEmpty(row.Merchant),
			ReceiptURL:    strOrEmpty(row.ReceiptUrl),
			AICategorized: row.AiCategorized,
			AIConfidence:  fromPgNumeric(row.AiConfidence),
			TransactedAt:  fromPgTime(row.TransactedAt),
			CreatedAt:     fromPgTime(row.CreatedAt),
			UpdatedAt:     fromPgTime(row.UpdatedAt),
		})
	}
	return out, total, nil
}

func (r *transactionsRepo) Update(ctx context.Context, p UpdateTransactionParams) (Transaction, error) {
	q := generated.New(r.pool)
	params := generated.UpdateTransactionForUserParams{
		ID:     toPgUUID(p.ID),
		UserID: toPgUUID(p.UserID),
		Amount: p.Amount,
		Note:   p.Note,
	}
	if p.CategoryID != nil {
		params.CategoryID = toPgUUID(*p.CategoryID)
	}
	if p.TransactedAt != nil {
		params.TransactedAt = toPgTime(*p.TransactedAt)
	}

	row, err := q.UpdateTransactionForUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, apperr.ErrNotFound
		}
		return Transaction{}, fmt.Errorf("updating transaction: %w", err)
	}
	return Transaction{
		ID:            fromPgUUID(row.ID),
		UserID:        fromPgUUID(row.UserID),
		BudgetPlanID:  optUUID(row.BudgetPlanID),
		CategoryID:    fromPgUUID(row.CategoryID),
		Amount:        row.Amount,
		Note:          strOrEmpty(row.Note),
		Merchant:      strOrEmpty(row.Merchant),
		ReceiptURL:    strOrEmpty(row.ReceiptUrl),
		AICategorized: row.AiCategorized,
		AIConfidence:  fromPgNumeric(row.AiConfidence),
		TransactedAt:  fromPgTime(row.TransactedAt),
		CreatedAt:     fromPgTime(row.CreatedAt),
		UpdatedAt:     fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *transactionsRepo) SetReceipt(ctx context.Context, p SetReceiptParams) (Transaction, error) {
	q := generated.New(r.pool)
	params := generated.SetTransactionReceiptParams{
		ID:         toPgUUID(p.ID),
		UserID:     toPgUUID(p.UserID),
		ReceiptUrl: nilIfEmpty(p.ReceiptURL),
	}
	if p.AIConfidence > 0 {
		params.AiConfidence = pgNumeric(p.AIConfidence)
	}

	row, err := q.SetTransactionReceipt(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, apperr.ErrNotFound
		}
		return Transaction{}, fmt.Errorf("setting transaction receipt: %w", err)
	}
	return Transaction{
		ID:            fromPgUUID(row.ID),
		UserID:        fromPgUUID(row.UserID),
		BudgetPlanID:  optUUID(row.BudgetPlanID),
		CategoryID:    fromPgUUID(row.CategoryID),
		Amount:        row.Amount,
		Note:          strOrEmpty(row.Note),
		Merchant:      strOrEmpty(row.Merchant),
		ReceiptURL:    strOrEmpty(row.ReceiptUrl),
		AICategorized: row.AiCategorized,
		AIConfidence:  fromPgNumeric(row.AiConfidence),
		TransactedAt:  fromPgTime(row.TransactedAt),
		CreatedAt:     fromPgTime(row.CreatedAt),
		UpdatedAt:     fromPgTime(row.UpdatedAt),
	}, nil
}

func (r *transactionsRepo) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	q := generated.New(r.pool)
	n, err := q.SoftDeleteTransactionForUser(ctx, generated.SoftDeleteTransactionForUserParams{
		ID:     toPgUUID(id),
		UserID: toPgUUID(userID),
	})
	if err != nil {
		return fmt.Errorf("soft-deleting transaction: %w", err)
	}
	if n == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

// --- small helpers, all in-file because they're trivial -----------------

func toPgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func optUUID(p pgtype.UUID) *uuid.UUID {
	if !p.Valid {
		return nil
	}
	u := fromPgUUID(p)
	return &u
}
