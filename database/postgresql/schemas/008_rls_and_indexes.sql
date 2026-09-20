-- 008_rls_and_indexes.sql
-- Row-Level Security (RLS) and High-Performance Indexing Strategy

-- Enable RLS on identity and financial boundary tables
ALTER TABLE identity.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE ledger.accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE trading.orders ENABLE ROW LEVEL SECURITY;

-- Tenant context parameter: session_replication_role or custom tenant context app.current_entity
-- Policy: Users can only see records matching the active entity boundary
CREATE POLICY tenant_isolation_users ON identity.users
    FOR ALL
    USING (
        current_setting('app.current_entity', true) IS NULL OR
        entity_type = current_setting('app.current_entity', true)
    );

CREATE POLICY tenant_isolation_accounts ON ledger.accounts
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM identity.users u
            WHERE u.user_id = ledger.accounts.user_id
            AND (current_setting('app.current_entity', true) IS NULL OR u.entity_type = current_setting('app.current_entity', true))
        )
    );

CREATE POLICY tenant_isolation_orders ON trading.orders
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM identity.users u
            WHERE u.user_id = trading.orders.user_id
            AND (current_setting('app.current_entity', true) IS NULL OR u.entity_type = current_setting('app.current_entity', true))
        )
    );

-- High-Performance Indexes for Sub-Millisecond Financial Lookups

-- Partial index for active order lookups in the orderbook
CREATE INDEX IF NOT EXISTS idx_orders_user_active
    ON trading.orders (user_id, status)
    WHERE status IN ('NEW', 'ACTIVE', 'PARTIALLY_FILLED');

-- Composite index for matching engine order queueing
CREATE INDEX IF NOT EXISTS idx_orders_book_queue
    ON trading.orders (isin, side, price, created_at)
    WHERE status IN ('NEW', 'ACTIVE', 'PARTIALLY_FILLED');

-- Composite index for user ledger postings
CREATE INDEX IF NOT EXISTS idx_postings_account_created
    ON ledger.postings (account_id, entry_created_at DESC);

-- Index for journal entry reference lookups (DvP, Order, Settlement)
CREATE INDEX IF NOT EXISTS idx_journal_reference
    ON ledger.journal_entries (reference_id, reference_type);

-- BRIN index on high-throughput trades executed_at timestamp
CREATE INDEX IF NOT EXISTS idx_trades_executed_at_brin
    ON trading.trades USING BRIN (executed_at);

-- Composite index for ISIN and executed_at on trades
CREATE INDEX IF NOT EXISTS idx_trades_isin_executed
    ON trading.trades (isin, executed_at DESC);

-- Composite index for custody allocation lookups
CREATE INDEX IF NOT EXISTS idx_custody_isin_depository
    ON custody.allocations (isin, depository);
