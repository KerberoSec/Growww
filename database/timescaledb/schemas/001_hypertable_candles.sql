-- ============================================================================
-- Growww / NBSE TimescaleDB Financial Candlestick Hypertable Schema
-- Compliant with SEBI, RBI, IFSCA 7-Year Retention & Real-Time Market Feed
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Base Candlestick Hypertable
CREATE TABLE IF NOT EXISTS market_candles (
    bucket TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    interval_sec INT NOT NULL,
    open_e8 BIGINT NOT NULL,
    high_e8 BIGINT NOT NULL,
    low_e8 BIGINT NOT NULL,
    close_e8 BIGINT NOT NULL,
    volume_e8 BIGINT NOT NULL,
    quote_volume_e8 BIGINT NOT NULL,
    trade_count BIGINT NOT NULL,
    vwap_e8 BIGINT NOT NULL,
    state_hash VARCHAR(64) NOT NULL,
    finalized BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_market_candles PRIMARY KEY (symbol, interval_sec, bucket)
);

-- Convert to TimescaleDB Hypertable
-- Chunk interval set to 1 day for optimal L1 cache hit ratio and compression granularity
SELECT create_hypertable(
    'market_candles',
    'bucket',
    partitioning_column => 'symbol',
    number_partitions => 8,
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- High-performance composite indexes for low-latency query routing
CREATE INDEX IF NOT EXISTS idx_market_candles_symbol_bucket 
ON market_candles (symbol, interval_sec, bucket DESC);

CREATE INDEX IF NOT EXISTS idx_market_candles_state_hash 
ON market_candles (state_hash);
