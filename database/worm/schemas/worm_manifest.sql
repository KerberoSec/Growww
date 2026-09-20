-- ============================================================================
-- Growww / NBSE WORM (Write Once Read Many) Audit & Storage Manifest Schema
-- Compliant with SEBI, RBI, FIU-IND 7-Year Retention & SEC Rule 17a-4
-- ============================================================================

CREATE TABLE IF NOT EXISTS worm_objects (
    object_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bucket VARCHAR(128) NOT NULL,
    object_key VARCHAR(1024) NOT NULL,
    version_id VARCHAR(64) NOT NULL,
    sha256_checksum VARCHAR(64) NOT NULL,
    content_length BIGINT NOT NULL,
    retention_mode VARCHAR(32) NOT NULL CHECK (retention_mode IN ('COMPLIANCE', 'GOVERNANCE')),
    retain_until_date TIMESTAMPTZ NOT NULL,
    legal_hold BOOLEAN NOT NULL DEFAULT FALSE,
    merkle_root VARCHAR(64),
    besu_tx_hash VARCHAR(66),
    besu_block_number BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_worm_bucket_key_version UNIQUE (bucket, object_key, version_id)
);

CREATE INDEX IF NOT EXISTS idx_worm_objects_lookup 
ON worm_objects (bucket, object_key, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_worm_objects_retention 
ON worm_objects (retention_mode, retain_until_date, legal_hold);

CREATE INDEX IF NOT EXISTS idx_worm_objects_besu 
ON worm_objects (besu_tx_hash, besu_block_number);
