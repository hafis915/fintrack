-- name: CreateTransaction :one
insert into transactions (
    user_id, budget_plan_id, category_id, amount, note,
    receipt_url, ai_categorized, ai_confidence, transacted_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
returning id, user_id, budget_plan_id, category_id, amount, note,
          receipt_url, ai_categorized, ai_confidence, transacted_at,
          deleted_at, created_at, updated_at;

-- name: GetTransactionForUser :one
select t.id, t.user_id, t.budget_plan_id, t.category_id, t.amount, t.note,
       t.receipt_url, t.ai_categorized, t.ai_confidence, t.transacted_at,
       t.deleted_at, t.created_at, t.updated_at,
       c.name as category_name, c.icon as category_icon, c.type as category_type
from transactions t
join expense_categories c on c.id = t.category_id
where t.id = $1 and t.user_id = $2 and t.deleted_at is null;

-- name: ListTransactionsForUser :many
-- Filters are nullable — null on a param means "no filter on that column".
-- We rely on sqlc's NULL-handling: passing a NULL value to a coalesce-style
-- predicate matches every row.
select t.id, t.user_id, t.budget_plan_id, t.category_id, t.amount, t.note,
       t.receipt_url, t.ai_categorized, t.ai_confidence, t.transacted_at,
       t.deleted_at, t.created_at, t.updated_at,
       c.name as category_name, c.icon as category_icon, c.type as category_type
from transactions t
join expense_categories c on c.id = t.category_id
where t.user_id = $1
  and t.deleted_at is null
  and ($2::uuid is null or t.category_id = $2::uuid)
  and ($3::timestamptz is null or t.transacted_at >= $3::timestamptz)
  and ($4::timestamptz is null or t.transacted_at <  $4::timestamptz)
order by t.transacted_at desc, t.id desc
limit $5 offset $6;

-- name: CountTransactionsForUser :one
select count(*)
from transactions
where user_id = $1
  and deleted_at is null
  and ($2::uuid is null or category_id = $2::uuid)
  and ($3::timestamptz is null or transacted_at >= $3::timestamptz)
  and ($4::timestamptz is null or transacted_at <  $4::timestamptz);

-- name: UpdateTransactionForUser :one
-- Partial PATCH: sqlc.narg generates pointer params so a nil leaves the
-- existing column value untouched.
update transactions
set amount         = coalesce(sqlc.narg('amount'),         amount),
    note           = coalesce(sqlc.narg('note'),           note),
    category_id    = coalesce(sqlc.narg('category_id'),    category_id),
    transacted_at  = coalesce(sqlc.narg('transacted_at'),  transacted_at),
    budget_plan_id = coalesce(sqlc.narg('budget_plan_id'), budget_plan_id),
    updated_at     = now()
where id = @id and user_id = @user_id and deleted_at is null
returning id, user_id, budget_plan_id, category_id, amount, note,
          receipt_url, ai_categorized, ai_confidence, transacted_at,
          deleted_at, created_at, updated_at;

-- name: SoftDeleteTransactionForUser :execrows
update transactions
set deleted_at = now(), updated_at = now()
where id = $1 and user_id = $2 and deleted_at is null;

-- name: GetActiveBudgetPlanForDate :one
-- Resolves the user's plan for the year+month implicit in the given
-- transacted_at timestamp. Used by the create handler to auto-link.
select id
from budget_plans
where user_id = $1
  and period_year  = extract(year  from $2::timestamptz)::smallint
  and period_month = extract(month from $2::timestamptz)::smallint;
