-- Algorithmic Execution Orders (TWAP / VWAP Slicing)
CREATE TABLE IF NOT EXISTS algo_execution_orders (
    algo_order_id VARCHAR(64) PRIMARY KEY,
    parent_order_id VARCHAR(64) NOT NULL,
    strategy VARCHAR(20) NOT NULL,
    total_quantity NUMERIC(24, 8) NOT NULL,
    filled_quantity NUMERIC(24, 8) NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    slice_interval_seconds INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
);
