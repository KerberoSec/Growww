-- 006_compliance.sql
-- Compliance Bounded Context: Investor Whitelists, Freezes, and Sanctions Screening

CREATE SCHEMA IF NOT EXISTS compliance;

CREATE TABLE IF NOT EXISTS compliance.investor_whitelists (
    whitelist_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blockchain_address CHAR(42) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES identity.users(user_id),
    jurisdiction VARCHAR(16) NOT NULL CHECK (jurisdiction IN ('IN_DOMESTIC', 'IN_GIFT_CITY', 'FOREIGN_FPI')),
    kyc_hash CHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'REVOKED')),
    on_chain_synced BOOLEAN NOT NULL DEFAULT FALSE,
    synced_tx_hash CHAR(66),
    approved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS compliance.regulatory_freezes (
    freeze_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(user_id),
    regulator VARCHAR(32) NOT NULL CHECK (regulator IN ('SEBI', 'RBI', 'IFSCA', 'FIU_IND', 'ED', 'COURT_ORDER')),
    order_reference_no VARCHAR(128) NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'LIFTED')),
    imposed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    lifted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS compliance.sanctions_screening_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(user_id),
    screening_vendor VARCHAR(32) NOT NULL,
    hit_score NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    is_pep BOOLEAN NOT NULL DEFAULT FALSE, -- Politically Exposed Person
    is_sanctioned BOOLEAN NOT NULL DEFAULT FALSE,
    match_details JSONB,
    screened_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
