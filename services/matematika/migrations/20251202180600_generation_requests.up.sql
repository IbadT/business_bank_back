CREATE TABLE IF NOT EXISTS generation_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    month VARCHAR(7) NOT NULL,
    year INTEGER NOT NULL,
    turnover DECIMAL(15,2) NOT NULL,
    desired_profit_percent DECIMAL(5,2) NOT NULL,
    model VARCHAR(10) NOT NULL CHECK (model IN ('B2C', 'B2B')),
    initial_balance DECIMAL(15,2) NOT NULL,
    scale_factor INTEGER DEFAULT 1,
    custom_data JSONB,
    status VARCHAR(20) DEFAULT 'processing',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);