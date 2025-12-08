CREATE TABLE IF NOT EXISTS daily_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES generation_requests(id),
    balance_date DATE NOT NULL,
    balance DECIMAL(15,2) NOT NULL
);