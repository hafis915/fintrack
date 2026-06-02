-- name: ListTransactions :many
SELECT t.id, t.user_id, t.budget_plan_id, t.category_id, t.amount, t.note, t.receipt_url,
       t.ai_categorized, t.ai_confidence, t.transacted_at, t.created_at, t.updated_at,
       ec.name AS category_name, ec.icon AS category_icon, ec.type AS category_type
FROM transactions t JOIN expense_categories ec ON ec.id = t.category_id
WHERE t.user_id = $1
  AND (sqlc.narg('category_id')::uuid IS NULL OR t.category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('start_at')::timestamptz IS NULL OR t.transacted_at >= sqlc.narg('start_at'))
  AND (sqlc.narg('end_at')::timestamptz   IS NULL OR t.transacted_at <  sqlc.narg('end_at'))
ORDER BY t.transacted_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1
  AND (sqlc.narg('category_id')::uuid IS NULL OR category_id = sqlc.narg('category_id'));

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, budget_plan_id, category_id, amount, note, receipt_url,
                          ai_categorized, ai_confidence, transacted_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions SET
    amount        = COALESCE(sqlc.narg('amount'),        amount),
    note          = COALESCE(sqlc.narg('note'),          note),
    category_id   = COALESCE(sqlc.narg('category_id'),   category_id),
    transacted_at = COALESCE(sqlc.narg('transacted_at'), transacted_at),
    updated_at    = NOW()
WHERE id = $1 AND user_id = $2 RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: SumSpentByCategoryForPlan :many
SELECT category_id, SUM(amount)::bigint AS total
FROM transactions
WHERE user_id = $1 AND budget_plan_id = $2
GROUP BY category_id;
