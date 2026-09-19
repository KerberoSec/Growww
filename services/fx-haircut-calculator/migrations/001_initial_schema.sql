-- Cross-Currency Dynamic FX Haircut Schema
CREATE TABLE IF NOT EXISTS fx_rate_ticks (
    id BIGSERIAL PRIMARY KEY,
    currency_pair VARCHAR(10) NOT NULL,
    bid_rate NUMERIC(18, 6) NOT NULL,
    ask_rate NUMERIC(18, 6) NOT NULL,
    volatility_30d NUMERIC(8, 4) NOT NULL,
    haircut_percent NUMERIC(8, 4) NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
