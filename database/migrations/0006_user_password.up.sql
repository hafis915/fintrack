-- Adds a bcrypt password hash for local-first auth (Phase 0). Nullable so
-- pre-password registrations and JWT-bootstrap rows (uuid@local) stay valid;
-- the register flow sets it and login requires it. Plaintext is never stored.
alter table users
    add column if not exists password_hash text;
