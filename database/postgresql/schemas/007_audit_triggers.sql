-- 007_audit_triggers.sql
-- Financial Immutability Triggers (SEBI 8-Year Audit Mandate)

CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.system_events (
    event_id UUID DEFAULT gen_random_uuid(),
    table_name VARCHAR(64) NOT NULL,
    operation VARCHAR(8) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    record_id VARCHAR(64) NOT NULL,
    actor_id VARCHAR(64),
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (event_id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE IF NOT EXISTS audit.system_events_y2026m09 PARTITION OF audit.system_events
    FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');

-- Trigger function enforcing immutability on financial records
CREATE OR REPLACE FUNCTION audit.prevent_mutation_trigger()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Financial immutability violation: % operations are strictly forbidden on table %', TG_OP, TG_TABLE_NAME
        USING ERRCODE = '55000'; -- object_not_in_prerequisite_state
END;
$$ LANGUAGE plpgsql;

-- Apply immutability triggers to double-entry journal and trade execution tables
DROP TRIGGER IF EXISTS trg_immutable_journal_entries ON ledger.journal_entries;
CREATE TRIGGER trg_immutable_journal_entries
    BEFORE UPDATE OR DELETE ON ledger.journal_entries
    FOR EACH ROW EXECUTE FUNCTION audit.prevent_mutation_trigger();

DROP TRIGGER IF EXISTS trg_immutable_postings ON ledger.postings;
CREATE TRIGGER trg_immutable_postings
    BEFORE UPDATE OR DELETE ON ledger.postings
    FOR EACH ROW EXECUTE FUNCTION audit.prevent_mutation_trigger();

DROP TRIGGER IF EXISTS trg_immutable_trades ON trading.trades;
CREATE TRIGGER trg_immutable_trades
    BEFORE UPDATE OR DELETE ON trading.trades
    FOR EACH ROW EXECUTE FUNCTION audit.prevent_mutation_trigger();
