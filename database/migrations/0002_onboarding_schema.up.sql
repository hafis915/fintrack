-- Phase 1: schema for the goal-first onboarding flow.
-- Aligned with the PRD (docs/prd.html § Database) and ADR-014: we keep the
-- `users` table as the local stand-in for Supabase `auth.users`. Financial
-- data lives on `user_profiles` (one row per user) so the auth table stays
-- slim; income is encrypted at rest.

-- --- enums ---------------------------------------------------------------
create type expense_category_type as enum ('fixed', 'variable', 'debt', 'want');
create type financial_program     as enum ('pondasi', 'bebas_utang', 'goal_chaser', 'tumbuh', 'seimbang');
create type fatigue_status        as enum ('fresh', 'warning', 'fatigued');
create type lifestyle_style       as enum ('easy', 'balanced', 'strict');
create type housing_type          as enum ('kosan', 'kpr', 'keluarga');

-- --- users: drop the placeholder income columns -------------------------
-- 0001 stored income on `users`. PRD puts it on `user_profiles`, which is
-- the right shape (auth row stays slim). The fields were never read by
-- any code in Phase 0, so dropping is safe.
alter table users
    drop column if exists income_cipher,
    drop column if exists income_nonce;

-- --- user_profiles -------------------------------------------------------
create table user_profiles (
    id                  uuid primary key default uuid_generate_v4(),
    user_id             uuid not null unique references users(id) on delete cascade,
    income_encrypted    text,                   -- base64( nonce || ciphertext+tag ), AES-256-GCM
    income_hint         varchar(20),            -- "Rp 8jt" — safe-to-display digest
    housing_type        housing_type,
    lifestyle_style     lifestyle_style,
    emergency_months    smallint not null default 0,
    active_program      financial_program,
    onboarding_done     boolean not null default false,
    created_at          timestamptz not null default now(),
    updated_at          timestamptz not null default now()
);
alter table user_profiles enable row level security;

-- --- expense_categories --------------------------------------------------
-- user_id NULL → system default, available to every user.
-- user_id NOT NULL → custom row created via POST /categories.
create table expense_categories (
    id          uuid primary key default uuid_generate_v4(),
    user_id     uuid references users(id) on delete cascade,
    name        varchar(100) not null,
    icon        varchar(10),
    type        expense_category_type not null,
    is_default  boolean not null default false,
    is_active   boolean not null default true,
    sort_order  smallint not null default 0,
    created_at  timestamptz not null default now()
);
create index expense_categories_user_idx on expense_categories (user_id) where user_id is not null;
alter table expense_categories enable row level security;

-- --- budget_plans (one per user per month) -------------------------------
create table budget_plans (
    id           uuid primary key default uuid_generate_v4(),
    user_id      uuid not null references users(id) on delete cascade,
    period_year  smallint not null,
    period_month smallint not null,
    total_income bigint not null,                -- rupiah, plaintext (rls-protected; encrypted form lives on user_profiles)
    program      financial_program not null,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    unique (user_id, period_year, period_month),
    check (period_month between 1 and 12),
    check (total_income > 0)
);
alter table budget_plans enable row level security;

-- --- budget_items (per-category allocations within a plan) --------------
create table budget_items (
    id               uuid primary key default uuid_generate_v4(),
    budget_plan_id   uuid not null references budget_plans(id) on delete cascade,
    category_id      uuid not null references expense_categories(id),
    allocated_amount bigint not null check (allocated_amount >= 0),
    percentage       numeric(5,2),
    is_debt_focus    boolean not null default false,
    created_at       timestamptz not null default now(),
    updated_at       timestamptz not null default now(),
    unique (budget_plan_id, category_id)
);
create index budget_items_plan_idx on budget_items (budget_plan_id);
alter table budget_items enable row level security;

-- --- system default expense categories -----------------------------------
-- Seeded once so every new user has these as picklist items in onboarding.
-- IDs are deterministic (gen_random_uuid not used) so test fixtures can rely on
-- them, but we let Postgres generate UUIDs to keep migrations simple.
insert into expense_categories (user_id, name, icon, type, is_default, sort_order) values
    -- fixed
    (null, 'Sewa kosan',       '🏠', 'fixed',    true, 10),
    (null, 'Cicilan KPR',      '🏗',  'fixed',    true, 11),
    (null, 'Listrik & air',    '💡', 'fixed',    true, 12),
    (null, 'Internet & pulsa', '📶', 'fixed',    true, 13),
    (null, 'Transportasi',     '🛵', 'fixed',    true, 14),
    -- variable
    (null, 'Makan & minum',    '🍱', 'variable', true, 20),
    (null, 'Belanja harian',   '🛒', 'variable', true, 21),
    (null, 'Kesehatan',        '💊', 'variable', true, 22),
    -- want
    (null, 'Hiburan',          '🎬', 'want',     true, 30),
    (null, 'Nongkrong',        '☕', 'want',     true, 31),
    (null, 'Self-care',        '💆', 'want',     true, 32),
    -- debt
    (null, 'Kartu kredit',     '💳', 'debt',     true, 40),
    (null, 'Paylater',         '📱', 'debt',     true, 41),
    (null, 'Pinjaman lain',    '💸', 'debt',     true, 42);
