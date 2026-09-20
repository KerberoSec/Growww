-- ============================================================================
-- Growww / NBSE TimescaleDB Continuous Aggregates for Multi-Resolution Candles
-- Standard Intervals: 1-minute, 5-minute, 15-minute, 1-hour, 1-day
-- ============================================================================

-- 1-Minute Continuous Aggregate View
CREATE MATERIALIZED VIEW IF NOT EXISTS candles_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', bucket) AS bucket,
    symbol,
    FIRST(open_e8, bucket) AS open_e8,
    MAX(high_e8) AS high_e8,
    MIN(low_e8) AS low_e8,
    LAST(close_e8, bucket) AS close_e8,
    SUM(volume_e8) AS volume_e8,
    SUM(quote_volume_e8) AS quote_volume_e8,
    SUM(trade_count) AS trade_count,
    CASE 
        WHEN SUM(volume_e8) > 0 THEN (SUM(vwap_e8 * volume_e8) / SUM(volume_e8))
        ELSE LAST(close_e8, bucket)
    END AS vwap_e8
FROM market_candles
WHERE interval_sec = 1
GROUP BY time_bucket('1 minute', bucket), symbol
WITH NO DATA;

-- 5-Minute Continuous Aggregate View
CREATE MATERIALIZED VIEW IF NOT EXISTS candles_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', bucket) AS bucket,
    symbol,
    FIRST(open_e8, bucket) AS open_e8,
    MAX(high_e8) AS high_e8,
    MIN(low_e8) AS low_e8,
    LAST(close_e8, bucket) AS close_e8,
    SUM(volume_e8) AS volume_e8,
    SUM(quote_volume_e8) AS quote_volume_e8,
    SUM(trade_count) AS trade_count,
    CASE 
        WHEN SUM(volume_e8) > 0 THEN (SUM(vwap_e8 * volume_e8) / SUM(volume_e8))
        ELSE LAST(close_e8, bucket)
    END AS vwap_e8
FROM market_candles
WHERE interval_sec = 60
GROUP BY time_bucket('5 minutes', bucket), symbol
WITH NO DATA;

-- 1-Hour Continuous Aggregate View
CREATE MATERIALIZED VIEW IF NOT EXISTS candles_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', bucket) AS bucket,
    symbol,
    FIRST(open_e8, bucket) AS open_e8,
    MAX(high_e8) AS high_e8,
    MIN(low_e8) AS low_e8,
    LAST(close_e8, bucket) AS close_e8,
    SUM(volume_e8) AS volume_e8,
    SUM(quote_volume_e8) AS quote_volume_e8,
    SUM(trade_count) AS trade_count,
    CASE 
        WHEN SUM(volume_e8) > 0 THEN (SUM(vwap_e8 * volume_e8) / SUM(volume_e8))
        ELSE LAST(close_e8, bucket)
    END AS vwap_e8
FROM market_candles
WHERE interval_sec = 60
GROUP BY time_bucket('1 hour', bucket), symbol
WITH NO DATA;

-- Continuous Aggregate Refresh Policies
SELECT add_continuous_aggregate_policy('candles_1m',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '30 seconds',
    if_not_exists => TRUE);

SELECT add_continuous_aggregate_policy('candles_5m',
    start_offset => INTERVAL '12 hours',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => TRUE);

SELECT add_continuous_aggregate_policy('candles_1h',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '10 minutes',
    if_not_exists => TRUE);
