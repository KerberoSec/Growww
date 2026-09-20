-- 001_identity.sql
-- Identity Bounded Context: User Profiles, KYC Records, and Entity Affiliations

CREATE SCHEMA IF NOT EXISTS identity;

CREATE TABLE IF NOT EXISTS identity.users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(16) NOT NULL CHECK (entity_type IN ('DOMESTIC_RE', 'GIFT_CITY_GW')),
    blockchain_address CHAR(42) NOT NULL UNIQUE, -- Zero-PII pseudonymous address (0x...)
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_KYC' CHECK (status IN ('PENDING_KYC', 'ACTIVE', 'FROZEN', 'CLOSED')),
    kyc_level VARCHAR(16) NOT NULL DEFAULT 'NONE' CHECK (kyc_level IN ('NONE', 'BASIC', 'FULL_SEBI', 'FULL_IFSCA')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS identity.kyc_records (
    kyc_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(user_id) ON DELETE RESTRICT,
    pan_hash CHAR(64) NOT NULL, -- SHA-256 hash of PAN, raw PAN isolated in KMS vault
    aadhaar_vault_ref VARCHAR(128), -- Reference key in isolated KMS DPDP vault
    ckyc_reference_no VARCHAR(64),
    verification_agency VARCHAR(32) NOT NULL CHECK (verification_agency IN ('NSDL_CRA', 'CAMS_KRA', 'KARVY_KRA', 'CVL_KRA')),
    verified_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(32) NOT NULL DEFAULT 'IN_PROGRESS' CHECK (status IN ('IN_PROGRESS', 'VERIFIED', 'REJECTED', 'EXPIRED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS identity.bank_accounts (
    bank_account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(user_id) ON DELETE RESTRICT,
    account_number_vault_ref VARCHAR(128) NOT NULL, -- KMS encrypted pointer
    ifsc_code VARCHAR(11) NOT NULL,
    bank_name VARCHAR(128) NOT NULL,
    account_type VARCHAR(16) NOT NULL CHECK (account_type IN ('SAVINGS', 'CURRENT', 'NRE', 'NRO')),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    penny_drop_status VARCHAR(16) NOT NULL DEFAULT 'PENDING' CHECK (penny_drop_status IN ('PENDING', 'SUCCESS', 'FAILED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
