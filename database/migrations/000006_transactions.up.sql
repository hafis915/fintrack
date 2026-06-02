CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    budget_plan_id  UUID REFERENCES budget_plans(id),
    category_id     UUID NOT NULL REFERENCES expense_categories(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    note            TEXT,
    receipt_url     TEXT,
    ai_categorized  BOOLEAN NOT NULL DEFAULT FALSE,
    ai_confidence   NUMERIC(3,2),
    transacted_at   TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transactions_user_period   ON transactions (user_id, transacted_at DESC);
CREATE INDEX idx_transactions_user_category ON transactions (user_id, category_id, transacted_at DESC);
