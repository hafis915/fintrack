-- name: GetUserByID :one
select id, email, name, created_at, updated_at
from users
where id = $1;

-- name: GetUserByEmail :one
-- Includes password_hash for login verification (nil for legacy/bootstrap rows).
select id, email, name, password_hash, created_at, updated_at
from users
where lower(email) = lower($1);

-- name: CreateUserWithEmail :one
-- Local-first register (Phase 0): creates a user row with a real email + name
-- + bcrypt password hash. id defaults to uuid_generate_v4(). A duplicate email
-- surfaces as a unique violation (pg 23505) which the repository maps to
-- apperr.ErrAlreadyExists. password_hash is never returned.
insert into users (email, name, password_hash)
values ($1, $2, $3)
returning id, email, name, created_at, updated_at;

-- name: UpsertUser :one
-- Idempotent bootstrap of a user row from a JWT subject the first time the
-- authenticated subject hits a /v1 route. IMPORTANT: the conflict path only
-- bumps updated_at — it must NOT touch email, otherwise EnsureUser's
-- <uuid>@local placeholder would clobber a real email set during register.
insert into users (id, email)
values ($1, $2)
on conflict (id) do update set updated_at = now()
returning id, email, name, created_at, updated_at;
