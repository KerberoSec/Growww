-- 005_fees.sql
-- Fees & Tax Bounded Context: Transaction Fee Breakdown & FIFO Tax Lot Disposals

CREATE SCHEMA IF NOT EXISTS fees;

CREATE TABLE IF NOT EXISTS fees.transaction_fees (
    fee_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identity.users(user_id),
    trade_id UUID NOT NULL,
    isin CHAR(12) NOT NULL,
    turnover_amount NUMERIC(18, 4) NOT NULL CHECK (turnover_amount >= 0),
    fee_rate NUMERIC(6, 6) NOT NULL DEFAULT 0.000000, -- 0.00% (Zero Fee) at launch (FeeController governed)
    total_fee_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_fee_inr >= 0),
    treasury_portion_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (treasury_portion_inr >= 0),
    core_sgf_portion_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (core_sgf_portion_inr >= 0),
    ipf_portion_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (ipf_portion_inr >= 0),
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT chk_fee_split_sum CHECK (total_fee_inr = treasury_portion_inr + core_sgf_portion_inr + ipf_portion_inr)
);

CREATE TABLE IF NOT EXISTS fees.tax_lot_disposals (
    disposal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identity.users(user_id),
    trade_id UUID NOT NULL,
    isin CHAR(12) NOT NULL,
    cost_basis NUMERIC(18, 4) NOT NULL CHECK (cost_basis >= 0),
    sale_proceeds NUMERIC(18, 4) NOT NULL CHECK (sale_proceeds >= 0),
    realized_capital_gain NUMERIC(18, 4) NOT NULL,
    holding_period_days INTEGER NOT NULL CHECK (holding_period_days >= 0),
    gain_type VARCHAR(8) NOT NULL CHECK (gain_type IN ('STCG', 'LTCG')), -- Section 111A / 112A compliance
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
