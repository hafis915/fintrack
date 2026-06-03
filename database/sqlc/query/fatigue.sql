-- name: GetBudgetWithSpend :many
-- Returns the plan + per-item spent_amount for a (user, year, month). One row
-- per budget_item, so the handler can map straight into a list response.
-- Spent is summed from transactions that fall in the same calendar month;
-- we filter by EXTRACT(year/month) rather than budget_plan_id so transactions
-- logged before the plan existed still count (they have NULL budget_plan_id).
select
    bp.id            as plan_id,
    bp.user_id       as user_id,
    bp.period_year   as period_year,
    bp.period_month  as period_month,
    bp.total_income  as total_income,
    bp.program       as program,
    bi.id            as item_id,
    bi.category_id   as category_id,
    c.name           as category_name,
    c.icon           as category_icon,
    c.type           as category_type,
    bi.allocated_amount as allocated_amount,
    bi.percentage    as allocation_percentage,
    bi.is_debt_focus as is_debt_focus,
    coalesce(sum(t.amount), 0)::bigint as spent_amount
from budget_plans bp
join budget_items bi on bi.budget_plan_id = bp.id
join expense_categories c on c.id = bi.category_id
left join transactions t
       on t.user_id = bp.user_id
      and t.category_id = bi.category_id
      and t.deleted_at is null
      and extract(year  from t.transacted_at)::int = bp.period_year
      and extract(month from t.transacted_at)::int = bp.period_month
where bp.user_id = $1
  and bp.period_year = $2
  and bp.period_month = $3
group by bp.id, bi.id, c.id
order by c.sort_order, c.name;

-- name: GetUnallocatedSpendForPeriod :one
-- Catch-all bucket: total spent on categories the user does NOT have a
-- budget item for. Surfaced in the summary so users see "Lainnya" if they
-- log a transaction outside their plan.
select coalesce(sum(t.amount), 0)::bigint as unallocated_spent
from transactions t
where t.user_id = $1
  and t.deleted_at is null
  and extract(year  from t.transacted_at)::int = $2::int
  and extract(month from t.transacted_at)::int = $3::int
  and not exists (
      select 1
      from budget_items bi
      join budget_plans bp on bp.id = bi.budget_plan_id
      where bp.user_id = $1
        and bp.period_year = $2::smallint
        and bp.period_month = $3::smallint
        and bi.category_id = t.category_id
  );
