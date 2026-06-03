-- Phase 0: bare minimum schema so the API has something to talk to.
-- A users table + the uuid extension. Domain tables (transactions, budgets,
-- categories, fatigue) come in Phase 1+ as separate numbered migrations.

create extension if not exists "uuid-ossp";

create table users (
    id              uuid primary key default uuid_generate_v4(),
    email           text not null unique,
    income_cipher   bytea,                  -- AES-256-GCM ciphertext, nullable until onboarding
    income_nonce    bytea,                  -- per-row nonce, populated together with cipher
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);

create index users_email_idx on users (lower(email));

-- RLS scaffolding: enable now so we don't forget. Policies are added in Phase 1
-- once we have JWT->session wiring; until then connect as the owner role to bypass.
alter table users enable row level security;
