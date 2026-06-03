-- Phase 2: transactions table.
-- Matches PRD § Database. budget_plan_id is nullable: the API auto-links
-- on create when a plan exists for the transaction's (year, month),
-- otherwise leaves NULL — the fatigue dashboard joins on it later.
--
-- amount stays plaintext (per PRD) — RLS + the auth shadow are what
-- protect transaction data, not at-rest encryption. Income is the only
-- field encrypted at rest (ADR-008/014).

create table transactions (
    id              uuid primary key default uuid_generate_v4(),
    user_id         uuid not null references users(id) on delete cascade,
    budget_plan_id  uuid references budget_plans(id) on delete set null,
    category_id     uuid not null references expense_categories(id),
    amount          bigint not null check (amount > 0),
    note            text,
    receipt_url     text,
    ai_categorized  boolean not null default false,
    ai_confidence   numeric(3,2),
    transacted_at   timestamptz not null,
    deleted_at      timestamptz,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);

-- The two queries the MVP actually runs:
--   • per-user listing in reverse-chronological order (fatigue + ledger view)
--   • per-user + per-category filtering (the category-fatigue join)
create index transactions_user_time_idx
    on transactions (user_id, transacted_at desc)
    where deleted_at is null;

create index transactions_user_category_idx
    on transactions (user_id, category_id)
    where deleted_at is null;

alter table transactions enable row level security;
