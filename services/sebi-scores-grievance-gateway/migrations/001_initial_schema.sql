-- SEBI SCORES 2.0 Grievance & ATR History
CREATE TABLE IF NOT EXISTS scores_complaint_history (
    docket_id VARCHAR(64) PRIMARY KEY,
    scores_reg_no VARCHAR(32) UNIQUE NOT NULL,
    pan_hash VARCHAR(64) NOT NULL,
    current_sla_stage INT NOT NULL DEFAULT 0,
    atr_generated BOOLEAN NOT NULL DEFAULT FALSE,
    statutory_deadline TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
