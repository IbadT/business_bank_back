CREATE TABLE IF NOT EXISTS default_customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    min_percent DECIMAL(5,4) NOT NULL,
    max_percent DECIMAL(5,4) NOT NULL,
    min_transactions INTEGER NOT NULL,
    max_transactions INTEGER NOT NULL
);