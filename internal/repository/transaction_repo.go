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
	"github.com/hafis915/fintrack/internal/domain/transaction"
	"github.com/hafis915/fintrack/pkg/apperror"
)

type transactionRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) transaction.Repository {
	return &transactionRepo{q: db.New(pool), pool: pool}
}

func (r *transactionRepo) Create(ctx context.Context, in transaction.CreateInput) (*transaction.Transaction, error) {
	var planID pgtype.UUID
	if in.BudgetPlanID != nil {
		planID = toPgUUID(*in.BudgetPlanID)
	}
	conf := pgtype.Numeric{}
	if in.AICategorized {
		_ = conf.Scan(strconv.FormatFloat(in.AIConfidence, 'f', 2, 64))
	}
	row, err := r.q.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:        toPgUUID(in.UserID),
		BudgetPlanID:  planID,
		CategoryID:    toPgUUID(in.CategoryID),
		Amount:        in.Amount,
		Note:          in.Note,
		ReceiptUrl:    in.ReceiptURL,
		AiCategorized: in.AICategorized,
		AiConfidence:  conf,
		TransactedAt:  pgtype.Timestamptz{Time: in.TransactedAt, Valid: true},
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	tx := txRowToDomain(row)
	return &tx, nil
}

func (r *transactionRepo) Update(ctx context.Context, in transaction.UpdateInput) (*transaction.Transaction, error) {
	var (
		catID pgtype.UUID
		ts    pgtype.Timestamptz
	)
	if in.CategoryID != nil {
		catID = toPgUUID(*in.CategoryID)
	}
	if in.TransactedAt != nil {
		ts = pgtype.Timestamptz{Time: *in.TransactedAt, Valid: true}
	}
	row, err := r.q.UpdateTransaction(ctx, db.UpdateTransactionParams{
		ID:           toPgUUID(in.ID),
		UserID:       toPgUUID(in.UserID),
		Amount:       in.Amount,
		Note:         in.Note,
		CategoryID:   catID,
		TransactedAt: ts,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("transaction", in.ID.String())
		}
		return nil, apperror.Internal(err)
	}
	tx := txRowToDomain(row)
	return &tx, nil
}

func (r *transactionRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.DeleteTransaction(ctx, db.DeleteTransactionParams{
		ID:     toPgUUID(id),
		UserID: toPgUUID(userID),
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *transactionRepo) List(ctx context.Context, f transaction.ListFilter) (*transaction.ListResult, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	var (
		catID pgtype.UUID
		start pgtype.Timestamptz
		end   pgtype.Timestamptz
	)
	if f.CategoryID != nil {
		catID = toPgUUID(*f.CategoryID)
	}
	if f.StartAt != nil {
		start = pgtype.Timestamptz{Time: *f.StartAt, Valid: true}
	}
	if f.EndAt != nil {
		end = pgtype.Timestamptz{Time: *f.EndAt, Valid: true}
	}

	rows, err := r.q.ListTransactions(ctx, db.ListTransactionsParams{
		UserID:     toPgUUID(f.UserID),
		Limit:      int32(limit),
		Offset:     int32(f.Offset),
		CategoryID: catID,
		StartAt:    start,
		EndAt:      end,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}

	total, err := r.q.CountTransactions(ctx, db.CountTransactionsParams{
		UserID:     toPgUUID(f.UserID),
		CategoryID: catID,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}

	items := make([]transaction.Transaction, 0, len(rows))
	for _, row := range rows {
		items = append(items, txListRowToDomain(row))
	}
	return &transaction.ListResult{Items: items, Total: int(total)}, nil
}

func (r *transactionRepo) SumSpentByCategoryForPlan(ctx context.Context, userID, planID uuid.UUID) ([]transaction.CategorySpent, error) {
	rows, err := r.q.SumSpentByCategoryForPlan(ctx, db.SumSpentByCategoryForPlanParams{
		UserID:       toPgUUID(userID),
		BudgetPlanID: toPgUUID(planID),
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	out := make([]transaction.CategorySpent, 0, len(rows))
	for _, row := range rows {
		out = append(out, transaction.CategorySpent{
			CategoryID: fromPgUUID(row.CategoryID),
			Total:      row.Total,
		})
	}
	return out, nil
}

// --- conversion helpers ---

func txRowToDomain(r db.Transaction) transaction.Transaction {
	tx := transaction.Transaction{
		ID:            fromPgUUID(r.ID),
		UserID:        fromPgUUID(r.UserID),
		CategoryID:    fromPgUUID(r.CategoryID),
		Amount:        r.Amount,
		Note:          r.Note,
		ReceiptURL:    r.ReceiptUrl,
		AICategorized: r.AiCategorized,
		TransactedAt:  r.TransactedAt.Time,
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
	}
	if r.BudgetPlanID.Valid {
		id := fromPgUUID(r.BudgetPlanID)
		tx.BudgetPlanID = &id
	}
	if conf, err := pgNumericToFloat(r.AiConfidence); err == nil {
		tx.AIConfidence = conf
	}
	return tx
}

func txListRowToDomain(r db.ListTransactionsRow) transaction.Transaction {
	tx := transaction.Transaction{
		ID:            fromPgUUID(r.ID),
		UserID:        fromPgUUID(r.UserID),
		CategoryID:    fromPgUUID(r.CategoryID),
		CategoryName:  r.CategoryName,
		CategoryType:  r.CategoryType,
		Amount:        r.Amount,
		Note:          r.Note,
		ReceiptURL:    r.ReceiptUrl,
		AICategorized: r.AiCategorized,
		TransactedAt:  r.TransactedAt.Time,
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
	}
	if r.CategoryIcon != nil {
		tx.CategoryIcon = *r.CategoryIcon
	}
	if r.BudgetPlanID.Valid {
		id := fromPgUUID(r.BudgetPlanID)
		tx.BudgetPlanID = &id
	}
	if conf, err := pgNumericToFloat(r.AiConfidence); err == nil {
		tx.AIConfidence = conf
	}
	return tx
}
