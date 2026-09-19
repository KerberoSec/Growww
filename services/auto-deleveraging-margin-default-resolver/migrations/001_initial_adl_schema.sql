-- Auto-Deleveraging (ADL) Priority Queue & Insurance Drawdown
CREATE TABLE IF NOT EXISTS adl_events (
    event_id BIGSERIAL PRIMARY KEY,
    bankrupt_user_id VARCHAR(64) NOT NULL,
    counterparty_user_id VARCHAR(64) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    deleveraged_quantity NUMERIC(24, 8) NOT NULL,
    execution_price NUMERIC(18, 4) NOT NULL,
    insurance_fund_delta NUMERIC(18, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
