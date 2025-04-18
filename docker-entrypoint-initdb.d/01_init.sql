CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS auth_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    user_id UUID REFERENCES auth_users(id) ON DELETE CASCADE,
    jti TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    client_ip TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, jti)
);

CREATE INDEX idx_refresh_tokens_user_jti_expires 
ON refresh_tokens (user_id, jti, expires_at);