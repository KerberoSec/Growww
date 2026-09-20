-- ============================================================================
-- Growww / NBSE Parquet Columnar Data Lake Export Schema
-- High-Throughput Analytics & Regulatory Data Lake Partitioning
-- ============================================================================

CREATE TABLE IF NOT EXISTS parquet_export_manifest (
    export_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(64) NOT NULL,
    partition_path VARCHAR(512) NOT NULL,
    file_name VARCHAR(256) NOT NULL,
    row_count BIGINT NOT NULL,
    uncompressed_bytes BIGINT NOT NULL,
    compressed_bytes BIGINT NOT NULL,
    compression_codec VARCHAR(32) NOT NULL DEFAULT 'SNAPPY',
    sha256_checksum VARCHAR(64) NOT NULL,
    start_timestamp_ns BIGINT NOT NULL,
    end_timestamp_ns BIGINT NOT NULL,
    exported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parquet_export_lookup 
ON parquet_export_manifest (table_name, partition_path, exported_at DESC);
