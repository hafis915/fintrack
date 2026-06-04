-- Reverse 0005_user_auth.

-- Restore the narrower percentage type.
alter table budget_items
    alter column percentage type numeric(5,2);

-- Restore the original non-unique lower(email) index.
drop index if exists users_email_idx;
create index users_email_idx on users (lower(email));

-- Drop the display name column.
alter table users
    drop column if exists name;
