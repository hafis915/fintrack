-- name: GetUserProfileByUserID :one
SELECT * FROM user_profiles WHERE user_id = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (
    user_id, income_encrypted, income_hint, housing_type, lifestyle_style,
    emergency_months, active_program, onboarding_done
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE user_profiles SET
    lifestyle_style  = COALESCE(sqlc.narg('lifestyle_style'), lifestyle_style),
    emergency_months = COALESCE(sqlc.narg('emergency_months'), emergency_months),
    active_program   = COALESCE(sqlc.narg('active_program'), active_program),
    onboarding_done  = COALESCE(sqlc.narg('onboarding_done'), onboarding_done),
    updated_at       = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserIncome :one
UPDATE user_profiles SET
    income_encrypted = $2,
    income_hint      = $3,
    updated_at       = NOW()
WHERE user_id = $1
RETURNING income_hint;
