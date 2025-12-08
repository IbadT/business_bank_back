CREATE TABLE IF NOT EXISTS financial_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES generation_requests(id),
    initial_balance DECIMAL(15,2) NOT NULL,
    final_balance DECIMAL(15,2) NOT NULL,
    total_revenue DECIMAL(15,2) NOT NULL,
    total_expenses DECIMAL(15,2) NOT NULL,
    net_profit DECIMAL(15,2) NOT NULL
);