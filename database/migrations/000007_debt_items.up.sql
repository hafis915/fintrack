CREATE TABLE debt_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    category_id     UUID REFERENCES expense_categories(id),
    name            VARCHAR(100) NOT NULL,
    total_amount    BIGINT NOT NULL,
    current_balance BIGINT NOT NULL,
    interest_rate   NUMERIC(5,2) NOT NULL,
    min_payment     BIGINT NOT NULL,
    method          debt_method NOT NULL DEFAULT 'snowball',
    priority        SMALLINT NOT NULL DEFAULT 1,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    started_at      DATE NOT NULL,
    target_paid_at  DATE,
    paid_at         DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_debt_items_user_active ON debt_items (user_id, is_active, priority);
