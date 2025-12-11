CREATE TABLE IF NOT EXISTS user_gateways (
    user_id UUID PRIMARY KEY,
    gateway_id VARCHAR(50) NOT NULL,
    gateway_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_gateways_user_id ON user_gateways(user_id);
