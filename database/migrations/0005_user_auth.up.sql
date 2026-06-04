-- Phase 0: lightweight LOCAL email register/login (ADR-014 keeps real Supabase
-- Auth deferred to v2). No passwords yet — we mint the same HS256 JWT the
-- middleware already validates. This migration:
--   1. adds users.name so a registered user has a display name
--   2. enforces a case-insensitive UNIQUE on users.email so register can 409
--   3. widens budget_items.percentage so realistic budgets can't overflow

-- --- users.name ----------------------------------------------------------
alter table users
    add column if not exists name text;

-- --- case-insensitive unique email --------------------------------------
-- 0001 already has a plain UNIQUE(email) plus a non-unique lower(email) index.
-- The existing <uuid>@local bootstrap rows are already unique under lower(),
-- so promoting the lower(email) index to UNIQUE is safe and blocks
-- case-variant duplicates (Foo@x.com vs foo@x.com) at the DB layer.
drop index if exists users_email_idx;
create unique index users_email_idx on users (lower(email));

-- --- widen budget_items.percentage --------------------------------------
-- numeric(5,2) maxes at 999.99 — a fat-fingered expense (e.g. income typo)
-- could push an allocation percentage past that and 500 on a DB overflow.
-- numeric(7,2) (max 99999.99) gives generous headroom while the domain layer
-- rejects truly degenerate input with a clean 400 (see ErrIncomeTooLow).
alter table budget_items
    alter column percentage type numeric(7,2);
