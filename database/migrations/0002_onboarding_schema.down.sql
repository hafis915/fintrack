drop table if exists budget_items;
drop table if exists budget_plans;
drop table if exists expense_categories;
drop table if exists user_profiles;

drop type if exists housing_type;
drop type if exists lifestyle_style;
drop type if exists fatigue_status;
drop type if exists financial_program;
drop type if exists expense_category_type;

-- Restore the placeholder income columns on users so 0001 stays reversible.
alter table users
    add column if not exists income_cipher bytea,
    add column if not exists income_nonce  bytea;
