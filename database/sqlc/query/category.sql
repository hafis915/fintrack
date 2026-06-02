-- name: ListCategoriesForUser :many
SELECT * FROM expense_categories
WHERE (user_id IS NULL OR user_id = $1) AND is_active = TRUE
ORDER BY sort_order, name;

-- name: CreateCategory :one
INSERT INTO expense_categories (user_id, name, icon, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM expense_categories WHERE id = $1 AND user_id = $2;

-- name: GetCategory :one
SELECT * FROM expense_categories WHERE id = $1;
