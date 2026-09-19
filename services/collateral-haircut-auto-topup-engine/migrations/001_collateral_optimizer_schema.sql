-- Collateral Optimization & Auto Top-up Trigger Schema
CREATE TABLE IF NOT EXISTS collateral_allocations (
    allocation_id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    asset_symbol VARCHAR(20) NOT NULL,
    allocated_quantity NUMERIC(24, 8) NOT NULL,
    haircut_applied_pct NUMERIC(6, 2) NOT NULL,
    effective_value_inr NUMERIC(18, 2) NOT NULL,
    margin_pool_id VARCHAR(32) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'COMMITTED',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
