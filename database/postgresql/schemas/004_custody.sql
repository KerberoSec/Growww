-- 004_custody.sql
-- Custody Bounded Context: 1:1 Physical Demat Backing Allocation & Proof-of-Reserve

CREATE SCHEMA IF NOT EXISTS custody;

CREATE TABLE IF NOT EXISTS custody.depository_participants (
    dp_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    depository VARCHAR(16) NOT NULL CHECK (depository IN ('NSDL', 'CDSL')),
    dp_name VARCHAR(128) NOT NULL,
    dp_code VARCHAR(32) NOT NULL UNIQUE,
    sebi_registration_no VARCHAR(32) NOT NULL UNIQUE,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS custody.demat_accounts (
    demat_account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(user_id) ON DELETE RESTRICT,
    dp_id UUID NOT NULL REFERENCES custody.depository_participants(dp_id),
    client_id VARCHAR(16) NOT NULL,
    beneficiary_id VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'DORMANT', 'CLOSED')),
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_dp_client UNIQUE (dp_id, client_id)
);

CREATE TABLE IF NOT EXISTS custody.allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin CHAR(12) NOT NULL,
    depository VARCHAR(16) NOT NULL CHECK (depository IN ('NSDL', 'CDSL')),
    demat_account_no VARCHAR(32) NOT NULL,
    physical_shares_held NUMERIC(18, 6) NOT NULL CHECK (physical_shares_held >= 0),
    tokens_minted NUMERIC(18, 6) NOT NULL CHECK (tokens_minted >= 0),
    reserve_status VARCHAR(16) NOT NULL DEFAULT 'BALANCED' CHECK (reserve_status IN ('BALANCED', 'DISCREPANCY', 'REBALANCING')),
    merkle_root_hash CHAR(66),
    on_chain_attestation_tx CHAR(66),
    last_reconciled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT chk_custody_1_to_1 CHECK (tokens_minted <= physical_shares_held)
);
