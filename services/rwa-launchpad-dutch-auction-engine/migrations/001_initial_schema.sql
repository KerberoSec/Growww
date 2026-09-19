-- RWA Launchpad & Dutch Auction Bids
CREATE TABLE IF NOT EXISTS dutch_auctions (
    auction_id VARCHAR(64) PRIMARY KEY,
    token_address VARCHAR(42) NOT NULL,
    start_price_inr NUMERIC(18, 2) NOT NULL,
    reserve_price_inr NUMERIC(18, 2) NOT NULL,
    clearing_price_inr NUMERIC(18, 2),
    total_tokens_offered NUMERIC(24, 8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL
);
