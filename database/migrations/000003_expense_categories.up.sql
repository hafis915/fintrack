CREATE TABLE expense_categories (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID,
    name        VARCHAR(100) NOT NULL,
    icon        VARCHAR(10),
    type        expense_category_type NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_expense_categories_user ON expense_categories (user_id);
