CREATE TABLE api_tokens (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL,
    name         VARCHAR(100) NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    token_hint   VARCHAR(10),
    can_read     BOOLEAN NOT NULL DEFAULT TRUE,
    can_write    BOOLEAN NOT NULL DEFAULT FALSE,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
