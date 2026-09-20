-- 003_trading.sql
-- Trading Bounded Context: Orders, Transitions, and Partitioned Trades

CREATE SCHEMA IF NOT EXISTS trading;

CREATE TABLE IF NOT EXISTS trading.orders (
    order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(user_id) ON DELETE RESTRICT,
    isin CHAR(12) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    order_type VARCHAR(16) NOT NULL CHECK (order_type IN ('LIMIT', 'MARKET', 'STOP_LIMIT')),
    price NUMERIC(18, 4) CHECK (price > 0),
    quantity NUMERIC(18, 6) NOT NULL CHECK (quantity > 0),
    filled_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000 CHECK (filled_quantity >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'ACTIVE', 'PARTIALLY_FILLED', 'FILLED', 'CANCELLED', 'REJECTED')),
    time_in_force VARCHAR(8) NOT NULL DEFAULT 'DAY' CHECK (time_in_force IN ('DAY', 'IOC', 'GTC', 'FOK')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT chk_filled_le_qty CHECK (filled_quantity <= quantity)
);

CREATE TABLE IF NOT EXISTS trading.order_state_transitions (
    transition_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES trading.orders(order_id) ON DELETE CASCADE,
    from_status VARCHAR(32) NOT NULL,
    to_status VARCHAR(32) NOT NULL,
    reason TEXT,
    transitioned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS trading.trades (
    trade_id UUID DEFAULT gen_random_uuid(),
    buyer_order_id UUID NOT NULL REFERENCES trading.orders(order_id),
    seller_order_id UUID NOT NULL REFERENCES trading.orders(order_id),
    isin CHAR(12) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    token_address CHAR(42) NOT NULL,
    fractional_units NUMERIC(18, 6) NOT NULL CHECK (fractional_units > 0),
    price_per_unit NUMERIC(18, 4) NOT NULL CHECK (price_per_unit > 0),
    gross_amount NUMERIC(18, 4) NOT NULL CHECK (gross_amount > 0),
    fee_amount NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (fee_amount >= 0),
    settlement_status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (settlement_status IN ('PENDING', 'SETTLED_DVP', 'FAILED')),
    on_chain_tx_hash CHAR(66), -- 0x... Besu transaction hash
    on_chain_block_num BIGINT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (trade_id, executed_at)
) PARTITION BY RANGE (executed_at);

-- Default initial monthly partition for trading.trades
CREATE TABLE IF NOT EXISTS trading.trades_y2026m09 PARTITION OF trading.trades
    FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');
