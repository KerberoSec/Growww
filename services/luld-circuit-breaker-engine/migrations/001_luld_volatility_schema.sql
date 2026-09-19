-- Limit Up / Limit Down State Machine History
CREATE TABLE IF NOT EXISTS luld_price_band_events (
    event_id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL,
    reference_price NUMERIC(18, 4) NOT NULL,
    upper_band NUMERIC(18, 4) NOT NULL,
    lower_band NUMERIC(18, 4) NOT NULL,
    triggered_state VARCHAR(32) NOT NULL,
    halt_duration_seconds INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
