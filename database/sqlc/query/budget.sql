-- name: CreateBudgetPlan :one
INSERT INTO budget_plans (user_id, period_year, period_month, total_income, program)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (user_id, period_year, period_month)
DO UPDATE SET total_income = EXCLUDED.total_income, program = EXCLUDED.program, updated_at = NOW()
RETURNING *;

-- name: CreateBudgetItem :one
INSERT INTO budget_items (budget_plan_id, category_id, allocated_amount, percentage, is_debt_focus)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (budget_plan_id, category_id)
DO UPDATE SET allocated_amount = EXCLUDED.allocated_amount, percentage = EXCLUDED.percentage,
              is_debt_focus = EXCLUDED.is_debt_focus, updated_at = NOW()
RETURNING *;

-- name: GetCurrentBudgetPlan :one
SELECT * FROM budget_plans WHERE user_id = $1 AND period_year = $2 AND period_month = $3;

-- name: ListBudgetItemsWithCategory :many
SELECT bi.id, bi.budget_plan_id, bi.category_id, bi.allocated_amount, bi.percentage,
       bi.is_debt_focus, bi.created_at, bi.updated_at,
       ec.name AS category_name, ec.icon AS category_icon, ec.type AS category_type
FROM budget_items bi JOIN expense_categories ec ON ec.id = bi.category_id
WHERE bi.budget_plan_id = $1
ORDER BY ec.sort_order;
