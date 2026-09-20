-- ============================================================================
-- Growww / NBSE Redis Keyspace Notifications & Expiry Manager Schema & Config
-- ============================================================================

-- Redis Configuration for Keyspace Notifications:
-- K: Keyspace events, published with __keyspace@<db>__ prefix.
-- E: Keyevent events, published with __keyevent@<db>__ prefix.
-- A: Alias for g$lshzxe, representing all events (including expired 'x' and evictions 'e').
-- CONFIG SET notify-keyspace-events "KEA"

-- Fallback Persistent Expiry Index Table for At-Least-Once Active Reconciliation
CREATE TABLE IF NOT EXISTS redis_active_timer_index (
    timer_key VARCHAR(512) PRIMARY KEY,
    event_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(128) NOT NULL,
    metadata JSONB,
    expires_at_ms BIGINT NOT NULL,
    dispatched BOOLEAN NOT NULL DEFAULT FALSE,
    dispatched_at TIMESTAMPTZ,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_redis_timer_pending 
ON redis_active_timer_index (expires_at_ms) 
WHERE dispatched = FALSE;
