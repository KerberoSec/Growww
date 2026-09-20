-- ============================================================================
-- Growww / NBSE TimescaleDB Compression and Retention Policies
-- Native Columnar Compression & 7-Year WORM Compliance Lifecycle
-- ============================================================================

-- Enable native columnar compression on market_candles hypertable
-- Segment by symbol and interval_sec to maximize run-length & dictionary encoding
-- Order by bucket DESC for delta-of-delta timestamp compression & fast reverse range scans
ALTER TABLE market_candles SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, interval_sec',
    timescaledb.compress_orderby = 'bucket DESC'
);

-- Automated Compression Policy: Compress chunks older than 7 days
SELECT add_compression_policy('market_candles', INTERVAL '7 days', if_not_exists => TRUE);

-- Automated Data Retention Policy: Retain 2555 days (7 years) per SEBI / RBI compliance
SELECT add_retention_policy('market_candles', INTERVAL '2555 days', if_not_exists => TRUE);

-- ============================================================================
-- Safe Historical Backfill and Decompression Procedures
-- ============================================================================

-- Procedure to safely decompress chunks before backfilling historical records
CREATE OR REPLACE PROCEDURE decompress_chunks_for_backfill(
    p_symbol VARCHAR(32),
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ
)
LANGUAGE plpgsql
AS $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (
        SELECT show_chunks('market_candles', older_than => p_end_time, newer_than => p_start_time) AS chunk_name
    ) LOOP
        BEGIN
            EXECUTE format('SELECT decompress_chunk(%L, if_compressed => true)', r.chunk_name);
            RAISE NOTICE 'Decompressed chunk % for historical backfill', r.chunk_name;
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING 'Failed to decompress chunk %: %', r.chunk_name, SQLERRM;
        END;
    END LOOP;
END;
$$;

-- Procedure to recompress chunks after backfill completion
CREATE OR REPLACE PROCEDURE recompress_chunks_after_backfill(
    p_symbol VARCHAR(32),
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ
)
LANGUAGE plpgsql
AS $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (
        SELECT show_chunks('market_candles', older_than => p_end_time, newer_than => p_start_time) AS chunk_name
    ) LOOP
        BEGIN
            EXECUTE format('SELECT compress_chunk(%L, if_not_compressed => true)', r.chunk_name);
            RAISE NOTICE 'Recompressed chunk % after backfill', r.chunk_name;
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING 'Failed to recompress chunk %: %', r.chunk_name, SQLERRM;
        END;
    END LOOP;
END;
$$;
