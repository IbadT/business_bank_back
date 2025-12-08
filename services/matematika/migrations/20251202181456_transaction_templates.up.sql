CREATE TABLE IF NOT EXISTS transaction_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_key VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    is_percentage BOOLEAN NOT NULL DEFAULT FALSE,
    percentage_min DECIMAL(5,4),
    percentage_max DECIMAL(5,2),
    fixed_amount DECIMAL(15,2),
    frequency VARCHAR(50),
    preferred_day VARCHAR(20),
    week_of_month INTEGER[],
    business_hours JSONB,
    is_optional BOOLEAN DEFAULT FALSE,
    priority INTEGER DEFAULT 100,
    method VARCHAR(50),
    min_transactions INTEGER DEFAULT 1,
    max_transactions INTEGER DEFAULT 1
);