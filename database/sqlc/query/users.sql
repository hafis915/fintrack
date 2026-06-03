-- name: GetUserByID :one
select id, email, created_at, updated_at
from users
where id = $1;

-- name: GetUserByEmail :one
select id, email, created_at, updated_at
from users
where lower(email) = lower($1);

-- name: UpsertUser :one
-- Idempotent: used to bootstrap a user row from a JWT subject when the
-- caller is the authenticated subject for the first time.
insert into users (id, email)
values ($1, $2)
on conflict (id) do update set updated_at = now()
returning id, email, created_at, updated_at;
