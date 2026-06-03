-- name: GetUserProfile :one
select id, user_id, income_encrypted, income_hint, housing_type,
       lifestyle_style, emergency_months, active_program, onboarding_done,
       created_at, updated_at
from user_profiles
where user_id = $1;

-- name: UpsertUserProfileForOnboarding :one
insert into user_profiles (
    user_id, income_encrypted, income_hint, housing_type,
    lifestyle_style, emergency_months, active_program, onboarding_done
) values (
    $1, $2, $3, $4, $5, $6, $7, true
)
on conflict (user_id) do update set
    income_encrypted = excluded.income_encrypted,
    income_hint      = excluded.income_hint,
    housing_type     = excluded.housing_type,
    lifestyle_style  = excluded.lifestyle_style,
    emergency_months = excluded.emergency_months,
    active_program   = excluded.active_program,
    onboarding_done  = true,
    updated_at       = now()
returning id, user_id, income_encrypted, income_hint, housing_type,
          lifestyle_style, emergency_months, active_program, onboarding_done,
          created_at, updated_at;

-- name: UpdateUserProfileLifestyle :one
update user_profiles
set lifestyle_style  = coalesce($2, lifestyle_style),
    emergency_months = coalesce($3, emergency_months),
    updated_at       = now()
where user_id = $1
returning id, user_id, income_encrypted, income_hint, housing_type,
          lifestyle_style, emergency_months, active_program, onboarding_done,
          created_at, updated_at;

-- name: UpdateUserProfileIncome :one
update user_profiles
set income_encrypted = $2,
    income_hint      = $3,
    updated_at       = now()
where user_id = $1
returning id, user_id, income_encrypted, income_hint, housing_type,
          lifestyle_style, emergency_months, active_program, onboarding_done,
          created_at, updated_at;
