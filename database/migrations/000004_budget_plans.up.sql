CREATE TABLE budget_plans (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL,
    period_year  SMALLINT NOT NULL,
    period_month SMALLINT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    total_income BIGINT NOT NULL,
    program      financial_program NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, period_year, period_month)
);
CREATE INDEX idx_budget_plans_user_period ON budget_plans (user_id, period_year, period_month);
