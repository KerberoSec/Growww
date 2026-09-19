-- SEBI Margin Pledge / Re-Pledge Tracking
CREATE TABLE IF NOT EXISTS demat_margin_pledges (
    pledge_request_no VARCHAR(32) PRIMARY KEY,
    demat_account_no VARCHAR(16) NOT NULL,
    depository VARCHAR(4) NOT NULL CHECK (depository IN ('NSDL', 'CDSL')),
    isin VARCHAR(12) NOT NULL,
    pledged_quantity BIGINT NOT NULL,
    pledgee_demat_no VARCHAR(16) NOT NULL,
    otp_authenticated BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'INITIATED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
