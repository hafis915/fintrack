CREATE TABLE user_profiles (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL UNIQUE,
    income_encrypted    TEXT,
    income_hint         VARCHAR(20),
    housing_type        housing_type,
    lifestyle_style     lifestyle_style,
    emergency_months    SMALLINT NOT NULL DEFAULT 0,
    active_program      financial_program,
    onboarding_done     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
