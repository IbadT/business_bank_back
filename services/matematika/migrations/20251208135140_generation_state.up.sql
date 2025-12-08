CREATE TABLE IF NOT EXISTS generation_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    state_key VARCHAR(100) NOT NULL,
    state_value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, state_key)
);

CREATE INDEX IF NOT EXISTS idx_generation_state_user_key ON generation_state(user_id, state_key);
