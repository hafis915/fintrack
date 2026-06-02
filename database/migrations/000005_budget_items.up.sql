CREATE TABLE budget_items (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    budget_plan_id   UUID NOT NULL REFERENCES budget_plans(id) ON DELETE CASCADE,
    category_id      UUID NOT NULL REFERENCES expense_categories(id),
    allocated_amount BIGINT NOT NULL,
    percentage       NUMERIC(5,2),
    is_debt_focus    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (budget_plan_id, category_id)
);
