CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at DESC);

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);
CREATE INDEX IF NOT EXISTS idx_games_user_id ON games (user_id);

CREATE TABLE IF NOT EXISTS replays (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT,
    original_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    compression TEXT NOT NULL,
    compressed BOOLEAN NOT NULL DEFAULT TRUE,
    comment TEXT,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_replays_uploaded_at ON replays (uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_replays_game_id ON replays (game_id);
CREATE INDEX IF NOT EXISTS idx_replays_user_id ON replays (user_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON users TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON games TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON replays TO PUBLIC;

-- Create default test user
INSERT INTO users (id, login, password_hash) 
VALUES ('00000000-0000-0000-0000-000000000001', 'test_user', 'test_hash')
ON CONFLICT (id) DO NOTHING;