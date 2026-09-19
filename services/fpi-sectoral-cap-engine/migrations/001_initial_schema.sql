-- FPI Sectoral Cap & Clubbing Schema (SEBI 24% aggregate cap)
CREATE TABLE IF NOT EXISTS fpi_entities (
    fpi_registration_no VARCHAR(32) PRIMARY KEY,
    entity_name VARCHAR(255) NOT NULL,
    investor_group_id VARCHAR(64) NOT NULL,
    jurisdiction VARCHAR(3) NOT NULL,
    category INT NOT NULL CHECK (category IN (1, 2)),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sectoral_holdings (
    id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL,
    sector_name VARCHAR(100) NOT NULL,
    fpi_group_id VARCHAR(64) NOT NULL,
    holding_shares NUMERIC(24, 4) NOT NULL DEFAULT 0,
    holding_percent NUMERIC(8, 4) NOT NULL DEFAULT 0,
    aggregate_fpi_percent NUMERIC(8, 4) NOT NULL DEFAULT 0,
    statutory_cap_percent NUMERIC(8, 4) NOT NULL DEFAULT 24.0000,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
