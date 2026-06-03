-- name: ListExpenseCategoriesForUser :many
-- Returns system defaults + the user's custom categories, sorted by sort_order.
select id, user_id, name, icon, type, is_default, is_active, sort_order, created_at
from expense_categories
where (user_id is null or user_id = $1)
  and is_active = true
order by sort_order, name;

-- name: GetExpenseCategoryByID :one
select id, user_id, name, icon, type, is_default, is_active, sort_order, created_at
from expense_categories
where id = $1;

-- name: CreateCustomExpenseCategory :one
insert into expense_categories (user_id, name, icon, type, sort_order)
values ($1, $2, $3, $4, $5)
returning id, user_id, name, icon, type, is_default, is_active, sort_order, created_at;
