-- 002_ledger.sql
-- Ledger Bounded Context: Immutable Double-Entry Financial Invariants

CREATE SCHEMA IF NOT EXISTS ledger;

CREATE TABLE IF NOT EXISTS ledger.accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identity.users(user_id) ON DELETE RESTRICT,
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    account_type VARCHAR(32) NOT NULL CHECK (account_type IN ('USER_WALLET', 'SETTLEMENT_ESCROW', 'FEE_REVENUE', 'CUSTODY_RESERVE', 'TREASURY', 'CORE_SGF', 'IPF')),
    balance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (balance >= 0),
    locked_balance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (locked_balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT chk_locked_le_balance CHECK (locked_balance <= balance)
);

CREATE TABLE IF NOT EXISTS ledger.journal_entries (
    entry_id UUID DEFAULT gen_random_uuid(),
    reference_id UUID NOT NULL, -- Order ID, Deposit ID, or Settlement ID
    reference_type VARCHAR(32) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (entry_id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE IF NOT EXISTS ledger.postings (
    posting_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL,
    entry_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    account_id UUID REFERENCES ledger.accounts(account_id) ON DELETE RESTRICT,
    amount NUMERIC(18, 4) NOT NULL CHECK (amount <> 0), -- Positive: Credit, Negative: Debit
    balance_after NUMERIC(18, 4) NOT NULL CHECK (balance_after >= 0)
);

-- Default initial monthly partition for ledger.journal_entries
CREATE TABLE IF NOT EXISTS ledger.journal_entries_y2026m09 PARTITION OF ledger.journal_entries
    FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');
