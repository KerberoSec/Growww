-- Bond Yield & Convexity Analytics
CREATE TABLE IF NOT EXISTS bond_analytics_cache (
    isin VARCHAR(12) PRIMARY KEY,
    face_value NUMERIC(18, 2) NOT NULL,
    coupon_rate NUMERIC(8, 4) NOT NULL,
    ytm NUMERIC(8, 4) NOT NULL,
    modified_duration NUMERIC(8, 4) NOT NULL,
    convexity NUMERIC(12, 6) NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
