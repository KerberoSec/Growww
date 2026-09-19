-- NBBO Consolidated Tape Quotes
CREATE TABLE IF NOT EXISTS nbbo_quotes (
    quote_id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    best_bid_price NUMERIC(18, 4) NOT NULL,
    best_bid_venue VARCHAR(20) NOT NULL,
    best_ask_price NUMERIC(18, 4) NOT NULL,
    best_ask_venue VARCHAR(20) NOT NULL,
    spread NUMERIC(18, 4) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
