CREATE TABLE weekly_reports (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL,
    week_start    DATE NOT NULL,
    week_end      DATE NOT NULL,
    total_spent   BIGINT NOT NULL DEFAULT 0,
    total_budget  BIGINT NOT NULL DEFAULT 0,
    narrative     TEXT,
    generated_at  TIMESTAMPTZ,
    email_sent_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, week_start)
);
