-- After-Hours Trading Sessions & Orders
CREATE TABLE IF NOT EXISTS afterhours_sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    session_type VARCHAR(20) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    spread_multiplier NUMERIC(4, 2) NOT NULL DEFAULT 1.50,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN'
);
