-- ERC-4337 Account Abstraction UserOperation Mempool
CREATE TABLE IF NOT EXISTS user_operations (
    user_op_hash VARCHAR(66) PRIMARY KEY,
    sender VARCHAR(42) NOT NULL,
    nonce NUMERIC(38, 0) NOT NULL,
    call_data BYTEA NOT NULL,
    call_gas_limit BIGINT NOT NULL,
    verification_gas_limit BIGINT NOT NULL,
    pre_verification_gas BIGINT NOT NULL,
    max_fee_per_gas NUMERIC(38, 0) NOT NULL,
    max_priority_fee_per_gas NUMERIC(38, 0) NOT NULL,
    paymaster_and_data BYTEA,
    signature BYTEA NOT NULL,
    bundled_tx_hash VARCHAR(66),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
