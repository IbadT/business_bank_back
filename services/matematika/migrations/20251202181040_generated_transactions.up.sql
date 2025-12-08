CREATE TABLE IF NOT EXISTS generated_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES generation_requests(id),
    transaction_id VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    posting_date DATE NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    category VARCHAR(100) NOT NULL,
    method VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2),
    is_manual BOOLEAN DEFAULT FALSE,
    calculation_details JSONB,
    sort_order INTEGER
);