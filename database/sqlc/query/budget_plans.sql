-- name: CreateBudgetPlan :one
insert into budget_plans (user_id, period_year, period_month, total_income, program)
values ($1, $2, $3, $4, $5)
returning id, user_id, period_year, period_month, total_income, program, created_at, updated_at;

-- name: UpsertBudgetPlan :one
-- Used by onboarding so re-running it for the same month replaces the plan
-- rather than 409-ing the user. PRD's onboarding API is idempotent in spirit.
insert into budget_plans (user_id, period_year, period_month, total_income, program)
values ($1, $2, $3, $4, $5)
on conflict (user_id, period_year, period_month) do update set
    total_income = excluded.total_income,
    program      = excluded.program,
    updated_at   = now()
returning id, user_id, period_year, period_month, total_income, program, created_at, updated_at;

-- name: GetBudgetPlanForPeriod :one
select id, user_id, period_year, period_month, total_income, program, created_at, updated_at
from budget_plans
where user_id = $1 and period_year = $2 and period_month = $3;

-- name: DeleteBudgetItemsForPlan :exec
delete from budget_items where budget_plan_id = $1;

-- name: CreateBudgetItem :one
insert into budget_items (budget_plan_id, category_id, allocated_amount, percentage, is_debt_focus)
values ($1, $2, $3, $4, $5)
returning id, budget_plan_id, category_id, allocated_amount, percentage, is_debt_focus, created_at, updated_at;

-- name: ListBudgetItemsForPlan :many
select bi.id, bi.budget_plan_id, bi.category_id, bi.allocated_amount, bi.percentage,
       bi.is_debt_focus, bi.created_at, bi.updated_at,
       c.name as category_name, c.icon as category_icon, c.type as category_type
from budget_items bi
join expense_categories c on c.id = bi.category_id
where bi.budget_plan_id = $1
order by c.sort_order, c.name;
