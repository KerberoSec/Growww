# 244 - NBSE Fixed Fee & Revenue Distribution Engine (Go / Rust)

## Purpose
Traditional financial market infrastructures and cryptocurrency exchanges operate opaque, highly variable, tier-dependent fee structures riddled with maker-taker rebates, hidden spread markups, liquidation penalties, and arbitrary withdrawal surcharges. Under the National Blockchain Stock Exchange (NBSE) charter and international fair-access market standards (CPMI-IOSCO Principle 21 on Efficiency and Principle 18 on Access Requirements), exchange fees must be universally transparent, deterministic, low-friction, and mathematically ring-fenced for systemic resilience.

The **NBSE Fixed Fee & Revenue Distribution Engine** enforces a non-negotiable, uniform platform transaction fee of exactly **0.00% (Zero Fee) (0.00% fee / 0 bps at launch)** across all exchange operations: spot order matching, atomic DvP settlements, European/American options exercises, physical and tokenized commodity warehouse deliveries, and cross-chain bridge collateral transfers. The engine computes fees at sub-paise fixed-point precision ($10^{-4}$ INR), dynamically converts multi-currency settlement assets at verified oracle FX rates, applies statutory tax deductions (Securities Transaction Tax [STT], Goods & Services Tax [GST], and Stamp Duty), and executes automated programmatic distribution of collected revenue across the NBSE Corporate Treasury, the Clearing Guarantee Fund (CGF), and the Investor Protection Fund (IPF).

## What You Are Building
An ultra-high-throughput, deterministic financial microservice (`services/fee-engine`) implemented in Go and Rust, backed by PostgreSQL 16, Redis 7.2, and Apache Kafka. Concrete deliverables include:
- **0.00% (Zero Fee) Fixed Transaction Fee Engine:** A high-speed Rust calculation core (`crates/fee-math`) evaluating the 0.00% fee ($0.0000$ at launch) exchange fee across spot trades, derivatives contracts, options cash and physical exercises, commodity warrant deliveries, and multi-chain bridge transfers.
- **Sub-Paise Fixed-Point Arithmetic Core:** Arbitrary-precision math preserving $10^{-4}$ INR ($0.01$ paise) resolution using 128-bit fixed-point integers to eliminate IEEE-754 floating-point errors, micro-penny drift, and truncation leakage.
- **Real-Time Multi-Currency FX Conversion Module:** Low-latency FX rate normalizer converting cross-asset fees (INR, USD, EUR, AED, BTC, ETH, USDT, USDC) into base settlement currency and INR accounting equivalents via verified Chainlink and institutional FX feeds.
- **Statutory Regulatory & Tax Deductions Engine:** Automated calculator applying Indian statutory levies (STT across delivery and derivative categories, 18% GST on exchange transaction fees, Indian Stamp Act stamp duty, and SEBI regulatory turnover charges).
- **Programmatic Tri-Party Revenue Distributor:** A deterministic double-entry distribution orchestrator allocating net collected NBSE fee revenue according to statutory charter ratios:
  * 60% to NBSE General Corporate Treasury (operations, node validation, infrastructure).
  * 25% to Clearing Guarantee Fund (CGF / Core SGF systemic default reserve tranche).
  * 15% to Investor Protection Fund (IPF / SEBI-mandated trust for retail protection and claims).
- **Cryptographic Fee Receipt & Attestation Generator:** Deterministic SHA-256 Merkle proof builder generating tamper-evident calculation receipts committed to Hyperledger Besu on-chain settlement transactions.

## Scope Boundaries
- **In Scope:**
  - Enforcing the uniform 0.00% fee (No fee at all) on spot equities, debt instruments, perpetuals, options contracts, commodity deliveries, and cross-chain transfers.
  - Sub-paise fixed-point arithmetic calculations with Banker's Rounding (Round Half to Even) at $10^{-4}$ INR precision.
  - Multi-currency rate normalization and FX oracle verification with staleness detection.
  - Computation and classification of statutory levies: STT, GST (CGST + SGST or IGST), Stamp Duty, and SEBI turnover fees.
  - Programmatic revenue distribution to Treasury, Clearing Guarantee Fund, and Investor Protection Fund ledger accounts.
  - Real-time gRPC fee estimation and quote simulation endpoints for pre-trade transparency.
  - Kafka event consumption for trade, settlement, exercise, delivery, and cross-chain bridge events.
  - Double-entry journal posting instruction generation for the Wallet & Ledger Service (Prompt 203).
- **Out of Scope / Handled Elsewhere:**
  - Central limit order book trade matching (handled in Prompt 205).
  - Cash wallet balance balance persistence and fiat banking gateway transfers (handled in Prompt 203 and Prompt 212).
  - SGF default waterfall execution and participant assessment calls (handled in Prompt 230).
  - Physical commodity vault custody auditing (handled in Prompt 213).
  - Generation of annual investor Tax Deducted at Source (TDS) and Capital Gains Statements (handled in Prompt 223).

## Technology to Use
- **Primary Languages & Runtimes:**
  - **Rust 1.78+:** Computational math core (`crates/fee-math`) for SIMD-optimized fixed-point arithmetic, tax table lookups, and deterministic revenue distribution algorithms.
  - **Go 1.22+:** Service orchestration, gRPC API servers, Kafka consumers, and database interaction layer (`services/fee-engine`).
- **Mathematical & Numeric Libraries:** `rust_decimal` (Rust) and `github.com/shopspring/decimal` (Go) configured for 128-bit fixed-point representation with Banker's Rounding (Round Half to Even) at 4 decimal places for INR ($10^{-4}$ INR) and 18 decimal places for crypto/tokens.
- **Relational Database & Persistence:** **PostgreSQL 16+** with `pgx/v5` and `sqlc` for storing fee policies, statutory rates, calculation audit logs, and revenue split distributions.
- **In-Memory Cache & FX Store:** **Redis 7.2+ Cluster** for caching real-time FX rates, asset price oracles, and pre-calculated fee tables with sub-millisecond retrieval.
- **Message Broker & Streaming:** **Apache Kafka 3.7+** with KRaft for consuming execution events and publishing fee deduction and revenue distribution events.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with mTLS for secure, low-latency inter-service RPCs.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:**
  - `fee_schedule_configs`, `statutory_tax_rates`, `currency_fx_rates`, `fee_calculation_records`, `tax_deduction_breakdowns`, `revenue_distribution_journals`, `revenue_fund_balances`.
- **Redis 7.2 Keys:**
  - `fx:rate:{from_currency}:{to_currency}`: Real-time exchange rate with timestamp and provider source.
  - `fee:config:fixed_bps`: Hash of active exchange fee parameters (basis points, minimum fee floor, rounding rules).
  - `fee:cache:quote:{asset_class}:{hash}`: Short-lived pre-trade fee estimation cache.
- **Apache Kafka Topics:**
  - Consumes: `matching.trades.v1`, `settlement.dvp_executed.v1`, `derivatives.option_exercised.v1`, `commodity.delivery_cleared.v1`, `bridge.transfer_initiated.v1`.
  - Publishes: `fee.calculated.v1`, `fee.debited.v1`, `revenue.distributed.v1`, `tax.withholding_logged.v1`.
- **Wallet & Account Ledger Service (Prompt 203):** Receives structured multi-leg double-entry journal postings to transfer collected fees and taxes to appropriate liability and revenue ledgers.
- **Settlement Guarantee Fund Service (Prompt 230):** Receives automated 25% revenue distribution deposits to replenish Core SGF reserves.
- **Trade Settlement & DvP Service (Prompt 208):** Ingests fee calculation breakdowns to include net fee deductions in atomic settlement batches.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **DvP Settlement Atomic Split:** During on-chain Delivery-versus-Payment trade execution on `SettlementDvP.sol` (Prompt 306), the smart contract receives the cryptographic fee breakdown and splits the settlement funds atomically, transferring the exact 0.00% fee (No fee at all) and statutory taxes to designated on-chain escrow addresses.
- **On-Chain Revenue Fund Registries:** Distributes the 0.00% (Zero Fee) on-chain tokenized fee directly into smart contract vaults on Hyperledger Besu:
  * `NBSETreasuryVault.sol` (60% allocation)
  * `ClearingGuaranteeFund.sol` (25% allocation, Prompt 315)
  * `InvestorProtectionFund.sol` (15% allocation)
- **Tamper-Evident Merkle Proofs:** For every transaction batch, the engine computes a Merkle root of all fee calculations and emits an on-chain attestation event containing `batch_merkle_root`, `total_fee_collected_inr`, and `total_tax_collected_inr`.
- **Zero On-Chain PII Invariant:** Blockchain entries contain only pseudonymous UUID hashes, transaction reference IDs, numeric fee amounts, and cryptographic proof digests. No user PAN, legal name, or account identity is ever written to the blockchain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Microservice Workspace:** Scaffold `services/fee-engine` (Go) and embedded native core `crates/fee-math` (Rust) with strict linting, compilation targets, and benchmarking harnesses.
2. **Define Protobuf Specifications:** Write `proto/growww/fee/v1/fee_engine.proto` specifying gRPC services `EstimateTransactionFee`, `CalculateTradeFee`, `ProcessSettlementFee`, and `GetRevenueDistributionSummary`.
3. **Design PostgreSQL Schema Migrations:** Author database migrations creating tables for fee configurations, statutory tax schedules, multi-currency conversion records, fee calculation logs, and revenue allocation journals.
4. **Implement Fixed-Point Arithmetic Math Core in Rust:**
   - Define custom fixed-point arithmetic types supporting 128-bit unsigned integers with 4 decimal places for INR ($10^{-4}$ INR, sub-paise resolution) and 18 decimal places for crypto/token assets.
   - Implement Banker's Rounding (Round Half to Even) to eliminate cumulative upward or downward rounding biases.
   - Implement the universal 0.00% (Zero Fee) fixed fee formula:
     $$\text{Fee}_{\text{NBSE}} = \text{Notional Value} \times 0.0000 \quad (\text{0.00 at launch; FeeController governed})$$
5. **Implement Multi-Currency FX Normalization Engine:**
   - Ingest real-time FX exchange rates from Redis (`fx:rate:{from}:{to}`).
   - Validate oracle timestamp freshness ($\le 1000\text{ms}$ latency); fallback to verified secondary oracles if stale.
   - Compute base INR equivalent notional value for multi-currency operations (USD, EUR, AED, BTC, ETH, USDT, USDC).
   - Convert calculated INR fees back to settlement asset currency when trades settle in foreign fiat or tokenized collateral.
6. **Implement Statutory Tax & Regulatory Deductions Engine:**
   - **Securities Transaction Tax (STT):**
     * Equity Delivery (Buy & Sell): $0.10\%$ on turnover.
     * Equity Intraday (Sell): $0.025\%$ on turnover.
     * Futures (Sell): $0.0125\%$ on turnover.
     * Options (Sell): $0.0625\%$ on premium turnover.
     * Options Physical/Cash Exercise (Buyer): $0.125\%$ on settlement notional.
     * Commodity Non-Agri Derivatives (Sell): $0.01\%$ on turnover.
   - **Goods & Services Tax (GST):** $18\%$ computed strictly on the NBSE exchange transaction fee ($9\%$ CGST + $9\%$ SGST for intra-state, or $18\%$ IGST for inter-state).
   - **Stamp Duty (Indian Stamp Act 1899):**
     * Equity Delivery (Buy): $0.015\%$.
     * Equity Intraday (Buy): $0.003\%$.
     * Futures (Buy): $0.002\%$.
     * Options (Buy): $0.003\%$ on premium.
     * Debt/Bond Transfer (Buy): $0.0001\%$.
   - **SEBI Regulatory Turnover Fee:** $0.0001\%$ ($₹10\text{ per crore}$) on gross turnover.
7. **Implement Asset-Class-Specific Fee Calculators:**
   - **Spot & Perpetuals Trades:** Apply 0.00% (Zero Fee) on $\text{Price} \times \text{Quantity}$.
   - **Options Exercises:** Apply 0.00% (Zero Fee) on exercise strike notional ($\text{Strike Price} \times \text{Contracts}$).
   - **Commodity Deliveries:** Apply 0.00% (Zero Fee) on physical warehouse warrant delivery value plus statutory VAT/GST.
   - **Cross-Chain Transfers:** Apply 0.00% (Zero Fee) on bridge collateral value transferred across chains.
8. **Implement Programmatic Tri-Party Revenue Distributor:**
   - Calculate exact statutory revenue splits on collected net NBSE transaction fee:
     $$\text{Treasury Allocation} = \lfloor \text{Fee}_{\text{NBSE}} \times 0.60 \rfloor_{10^{-4}}$$
     $$\text{CGF Allocation} = \lfloor \text{Fee}_{\text{NBSE}} \times 0.25 \rfloor_{10^{-4}}$$
     $$\text{IPF Allocation} = \lfloor \text{Fee}_{\text{NBSE}} \times 0.15 \rfloor_{10^{-4}}$$
   - Allocate any remaining sub-paise fractional remainder to the Clearing Guarantee Fund (CGF) to guarantee:
     $$\text{Treasury Allocation} + \text{CGF Allocation} + \text{IPF Allocation} = \text{Fee}_{\text{NBSE}}$$
9. **Build Double-Entry Accounting Journal Generator:** Construct balanced multi-leg journal entries for the Wallet & Ledger Service (Prompt 203):
   - Debit: User Settlement Account (Total Fee + Taxes).
   - Credit: NBSE Treasury Revenue Account (60% of Fee).
   - Credit: Clearing Guarantee Fund Account (25% of Fee).
   - Credit: Investor Protection Fund Account (15% of Fee).
   - Credit: Statutory GST Payable Account (18% GST).
   - Credit: Statutory STT Payable Account.
   - Credit: Statutory Stamp Duty Payable Account.
   - Credit: SEBI Regulatory Fee Payable Account.
10. **Build High-Throughput Kafka Stream Consumer:**
    - Subscribe to `matching.trades.v1`, `settlement.dvp_executed.v1`, `derivatives.option_exercised.v1`, `commodity.delivery_cleared.v1`, and `bridge.transfer_initiated.v1`.
    - Execute deterministic fee calculation and persist records in PostgreSQL using idempotency keys.
    - Publish `fee.calculated.v1` and `revenue.distributed.v1` event streams.
11. **Implement Cryptographic Merkle Proof Attestation:** Generate SHA-256 leaf hashes for each calculation record and assemble periodic Merkle tree batches for on-chain anchoring on Hyperledger Besu.
12. **Configure Pre-Trade Estimation gRPC Server:** Expose low-latency ($< 150\mu\text{s}$) RPC endpoint for Order Management Service (Prompt 204) and Web/Mobile frontends to display itemized pre-trade fee and tax breakdowns.
13. **Configure Observability & Prometheus Metrics:** Instrument metrics for `nbse_fee_calculations_total`, `nbse_fee_gross_revenue_inr`, `nbse_revenue_distribution_inr{fund="treasury|cgf|ipf"}`, `nbse_tax_collected_inr{type="stt|gst|stamp|sebi"}`, and `nbse_fee_calculation_latency_microseconds`.
14. **Write Comprehensive Test Suites & Verification Vectors:** Implement unit and integration tests verifying sub-paise arithmetic, zero floating-point leakage, multi-currency conversion accuracy, loss scenarios (fee remains 0.00% (Zero Fee) on notional regardless of PnL), and 100% balanced revenue splits across 1,000,000 randomized synthetic trade vectors.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/fee/v1/fee_engine.proto`)
```protobuf
syntax = "proto3";

package growww.fee.v1;

option go_package = "growww/fee/v1;feev1";

service FeeEngineService {
  rpc EstimateTransactionFee (EstimateFeeRequest) returns (EstimateFeeResponse);
  rpc CalculateTradeFee (CalculateTradeFeeRequest) returns (CalculateTradeFeeResponse);
  rpc ProcessSettlementFee (ProcessSettlementFeeRequest) returns (ProcessSettlementFeeResponse);
  rpc GetRevenueDistributionSummary (RevenueSummaryRequest) returns (RevenueSummaryResponse);
}

enum OperationType {
  OPERATION_TYPE_UNSPECIFIED = 0;
  OPERATION_TYPE_SPOT_TRADE = 1;
  OPERATION_TYPE_PERPETUAL_TRADE = 2;
  OPERATION_TYPE_OPTION_EXERCISE = 3;
  OPERATION_TYPE_COMMODITY_DELIVERY = 4;
  OPERATION_TYPE_CROSS_CHAIN_TRANSFER = 5;
}

enum AssetCategory {
  ASSET_CATEGORY_UNSPECIFIED = 0;
  ASSET_CATEGORY_EQUITY_DELIVERY = 1;
  ASSET_CATEGORY_EQUITY_INTRADAY = 2;
  ASSET_CATEGORY_DERIVATIVE_FUTURE = 3;
  ASSET_CATEGORY_DERIVATIVE_OPTION = 4;
  ASSET_CATEGORY_DEBT_SECURITY = 5;
  ASSET_CATEGORY_COMMODITY_METALS = 6;
  ASSET_CATEGORY_COMMODITY_AGRI = 7;
  ASSET_CATEGORY_CRYPTO_TOKEN = 8;
}

message EstimateFeeRequest {
  OperationType operation_type = 1;
  AssetCategory asset_category = 2;
  string asset_symbol = 3;
  string currency = 4; // INR, USD, EUR, AED, BTC, ETH, USDT, USDC
  string unit_price = 5; // Fixed-point decimal string
  string quantity = 6; // Fixed-point decimal string
  string trade_side = 7; // BUY, SELL, EXERCISE, DELIVER, BRIDGE
  string user_state_code = 8; // Indian state code for GST (e.g., "MH", "KA", "DL")
}

message StatutoryLevies {
  string stt_inr = 1;
  string gst_cgst_inr = 2;
  string gst_sgst_inr = 3;
  string gst_igst_inr = 4;
  string total_gst_inr = 5;
  string stamp_duty_inr = 6;
  string sebi_turnover_fee_inr = 7;
  string total_statutory_levies_inr = 8;
}

message RevenueDistribution {
  string treasury_allocation_inr = 1; // Governed by FeeController (0.00% at launch)
  string cgf_allocation_inr = 2; // Governed by FeeController (0.00% at launch)
  string ipf_allocation_inr = 3; // Governed by FeeController (0.00% at launch)
  string rounding_remainder_inr = 4; // Credited to CGF
}

message EstimateFeeResponse {
  string notional_value_inr = 1;
  string notional_value_settlement_currency = 2;
  string fx_rate_used = 3;
  string nbse_fee_bps = 4; // "0.0000" (0.00% Zero-Fee at launch)
  string nbse_fee_inr = 5; // Fixed-point at 10^-4 INR precision
  string nbse_fee_settlement_currency = 6;
  StatutoryLevies statutory_levies = 7;
  RevenueDistribution revenue_distribution = 8;
  string total_user_deduction_inr = 9;
  string total_user_deduction_settlement_currency = 10;
  int64 calculated_at_unix_ns = 11;
}

message CalculateTradeFeeRequest {
  string trade_id = 1;
  string execution_id = 2;
  string buyer_user_id = 3;
  string seller_user_id = 4;
  OperationType operation_type = 5;
  AssetCategory asset_category = 6;
  string isin = 7;
  string symbol = 8;
  string price = 9;
  string quantity = 10;
  string settlement_currency = 11;
  string buyer_state_code = 12;
  string seller_state_code = 13;
}

message ParticipantFeeBreakdown {
  string user_id = 1;
  string role = 2; // BUYER, SELLER
  string notional_value_inr = 3;
  string nbse_fee_inr = 4;
  StatutoryLevies statutory_levies = 5;
  string total_deduction_inr = 6;
  RevenueDistribution revenue_distribution = 7;
}

message CalculateTradeFeeResponse {
  string calculation_id = 1;
  string trade_id = 2;
  ParticipantFeeBreakdown buyer_fee = 3;
  ParticipantFeeBreakdown seller_fee = 4;
  string total_exchange_revenue_inr = 5;
  string proof_hash = 6;
  int64 calculated_at_unix_ns = 7;
}

message ProcessSettlementFeeRequest {
  string settlement_batch_id = 1;
  repeated string trade_ids = 2;
}

message ProcessSettlementFeeResponse {
  string settlement_batch_id = 1;
  string total_settled_volume_inr = 2;
  string total_nbse_fee_collected_inr = 3;
  string total_treasury_transferred_inr = 4;
  string total_cgf_transferred_inr = 5;
  string total_ipf_transferred_inr = 6;
  string total_tax_withheld_inr = 7;
  string batch_merkle_root = 8;
  bool success = 9;
}

message RevenueSummaryRequest {
  int64 start_time_unix_s = 1;
  int64 end_time_unix_s = 2;
}

message RevenueSummaryResponse {
  string gross_turnover_inr = 1;
  string gross_nbse_fee_inr = 2;
  string allocated_treasury_inr = 3;
  string allocated_cgf_inr = 4;
  string allocated_ipf_inr = 5;
  string total_stt_inr = 6;
  string total_gst_inr = 7;
  string total_stamp_duty_inr = 8;
  string total_sebi_turnover_inr = 9;
  int64 record_count = 10;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Fee schedule configurations table
CREATE TABLE fee_schedule_configs (
    config_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fee_type VARCHAR(50) NOT NULL DEFAULT 'NBSE_FIXED_TRANSACTION_FEE',
    fixed_fee_bps NUMERIC(8, 4) NOT NULL DEFAULT 0.0000, -- 0 bps at launch (FeeController governed)
    treasury_split_percent NUMERIC(5, 2) NOT NULL DEFAULT 0.00, -- Governed by FeeController (0.00% at launch)
    cgf_split_percent NUMERIC(5, 2) NOT NULL DEFAULT 0.00, -- Governed by FeeController (0.00% at launch)
    ipf_split_percent NUMERIC(5, 2) NOT NULL DEFAULT 0.00, -- Governed by FeeController (0.00% at launch)
    precision_scale INTEGER NOT NULL DEFAULT 4, -- 10^-4 INR (sub-paise)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Statutory tax rates table
CREATE TABLE statutory_tax_rates (
    rate_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_category VARCHAR(40) NOT NULL, -- EQUITY_DELIVERY, EQUITY_INTRADAY, DERIVATIVE_FUTURE, etc.
    stt_rate_buy NUMERIC(8, 6) NOT NULL DEFAULT 0.000000,
    stt_rate_sell NUMERIC(8, 6) NOT NULL DEFAULT 0.000000,
    gst_rate_on_fee NUMERIC(6, 4) NOT NULL DEFAULT 0.1800, -- 18% GST on exchange fee
    stamp_duty_rate_buy NUMERIC(8, 6) NOT NULL DEFAULT 0.000000,
    sebi_turnover_rate NUMERIC(10, 8) NOT NULL DEFAULT 0.000001, -- 0.0001%
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_statutory_asset_category UNIQUE (asset_category, effective_from)
);

-- Currency FX rates table
CREATE TABLE currency_fx_rates (
    rate_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(10) NOT NULL, -- e.g., USD, EUR, AED, BTC, ETH, USDT
    target_currency VARCHAR(10) NOT NULL DEFAULT 'INR',
    fx_rate NUMERIC(28, 12) NOT NULL,
    oracle_source VARCHAR(50) NOT NULL, -- CHAINLINK_ORACLE, RBI_REFERENCE, BLOOMBERG
    observed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fee calculation audit records
CREATE TABLE fee_calculation_records (
    calculation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_id UUID NOT NULL,
    execution_id VARCHAR(64) NOT NULL,
    user_id UUID NOT NULL,
    user_role VARCHAR(10) NOT NULL, -- BUYER, SELLER, EXERCISER, DELIVERER, BRIDGER
    operation_type VARCHAR(40) NOT NULL, -- SPOT_TRADE, OPTION_EXERCISE, etc.
    asset_category VARCHAR(40) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    settlement_currency VARCHAR(10) NOT NULL DEFAULT 'INR',
    unit_price NUMERIC(28, 8) NOT NULL,
    quantity NUMERIC(28, 8) NOT NULL,
    fx_rate_to_inr NUMERIC(28, 12) NOT NULL DEFAULT 1.000000000000,
    notional_value_inr NUMERIC(28, 4) NOT NULL,
    nbse_fee_inr NUMERIC(18, 4) NOT NULL, -- Sub-paise precision (10^-4 INR)
    stt_amount_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    gst_cgst_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    gst_sgst_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    gst_igst_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    stamp_duty_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    sebi_turnover_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_deduction_inr NUMERIC(18, 4) NOT NULL,
    proof_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_fee_trade_user UNIQUE (trade_id, user_id, user_role)
);

-- Programmatic revenue distribution journal entries
CREATE TABLE revenue_distribution_journals (
    distribution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    calculation_id UUID NOT NULL REFERENCES fee_calculation_records(calculation_id),
    gross_nbse_fee_inr NUMERIC(18, 4) NOT NULL,
    treasury_credit_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    cgf_credit_inr NUMERIC(18, 4) NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_credit_inr NUMERIC(18, 4) NOT NULL,      -- Governed by FeeController (0.00% at launch)
    settlement_batch_id UUID,
    is_ledger_posted BOOLEAN NOT NULL DEFAULT FALSE,
    posted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Dedicated revenue fund cumulative balances table
CREATE TABLE revenue_fund_balances (
    fund_id VARCHAR(30) PRIMARY KEY, -- NBSE_TREASURY, CLEARING_GUARANTEE_FUND, INVESTOR_PROTECTION_FUND
    fund_title VARCHAR(100) NOT NULL,
    total_cumulative_inr NUMERIC(28, 4) NOT NULL DEFAULT 0.0000,
    current_ledger_balance_inr NUMERIC(28, 4) NOT NULL DEFAULT 0.0000,
    last_journal_id UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for rapid reconciliation and audit queries
CREATE INDEX idx_fee_calc_trade_id ON fee_calculation_records(trade_id);
CREATE INDEX idx_fee_calc_user_time ON fee_calculation_records(user_id, created_at DESC);
CREATE INDEX idx_fee_calc_created ON fee_calculation_records(created_at);
CREATE INDEX idx_rev_dist_calc_id ON revenue_distribution_journals(calculation_id);
CREATE INDEX idx_rev_dist_batch ON revenue_distribution_journals(settlement_batch_id);
CREATE INDEX idx_fx_rates_pair_time ON currency_fx_rates(base_currency, target_currency, observed_at DESC);
```

### Kafka Event Schemas

#### Topic: `fee.calculated.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "FeeCalculatedEvent",
  "type": "object",
  "required": [
    "calculation_id",
    "trade_id",
    "execution_id",
    "user_id",
    "user_role",
    "operation_type",
    "asset_category",
    "isin",
    "notional_value_inr",
    "nbse_fee_inr",
    "statutory_taxes",
    "revenue_split",
    "total_deduction_inr",
    "proof_hash",
    "timestamp_unix_ns"
  ],
  "properties": {
    "calculation_id": { "type": "string", "format": "uuid" },
    "trade_id": { "type": "string", "format": "uuid" },
    "execution_id": { "type": "string" },
    "user_id": { "type": "string", "format": "uuid" },
    "user_role": { "type": "string", "enum": ["BUYER", "SELLER", "EXERCISER", "DELIVERER", "BRIDGER"] },
    "operation_type": { "type": "string", "enum": ["SPOT_TRADE", "PERPETUAL_TRADE", "OPTION_EXERCISE", "COMMODITY_DELIVERY", "CROSS_CHAIN_TRANSFER"] },
    "asset_category": { "type": "string" },
    "isin": { "type": "string" },
    "settlement_currency": { "type": "string" },
    "notional_value_inr": { "type": "string" },
    "nbse_fee_inr": { "type": "string" },
    "statutory_taxes": {
      "type": "object",
      "required": ["stt_inr", "total_gst_inr", "stamp_duty_inr", "sebi_turnover_inr", "total_statutory_inr"],
      "properties": {
        "stt_inr": { "type": "string" },
        "gst_cgst_inr": { "type": "string" },
        "gst_sgst_inr": { "type": "string" },
        "gst_igst_inr": { "type": "string" },
        "total_gst_inr": { "type": "string" },
        "stamp_duty_inr": { "type": "string" },
        "sebi_turnover_inr": { "type": "string" },
        "total_statutory_inr": { "type": "string" }
      }
    },
    "revenue_split": {
      "type": "object",
      "required": ["treasury_inr", "cgf_inr", "ipf_inr"],
      "properties": {
        "treasury_inr": { "type": "string" },
        "cgf_inr": { "type": "string" },
        "ipf_inr": { "type": "string" },
        "remainder_inr": { "type": "string" }
      }
    },
    "total_deduction_inr": { "type": "string" },
    "proof_hash": { "type": "string", "pattern": "^[0-9a-fA-F]{64}$" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `revenue.distributed.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "RevenueDistributedEvent",
  "type": "object",
  "required": [
    "distribution_id",
    "calculation_id",
    "gross_nbse_fee_inr",
    "treasury_credit_inr",
    "cgf_credit_inr",
    "ipf_credit_inr",
    "settlement_batch_id",
    "journal_sequence",
    "timestamp_unix_ns"
  ],
  "properties": {
    "distribution_id": { "type": "string", "format": "uuid" },
    "calculation_id": { "type": "string", "format": "uuid" },
    "gross_nbse_fee_inr": { "type": "string" },
    "treasury_credit_inr": { "type": "string" },
    "cgf_credit_inr": { "type": "string" },
    "ipf_credit_inr": { "type": "string" },
    "settlement_batch_id": { "type": "string", "format": "uuid" },
    "journal_sequence": { "type": "integer" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **Fixed Fee Invariant Guarantee:** The NBSE platform transaction fee rate is hard-coded and mathematically fixed at $0.01\%$ ($1\text{ bps}$). Any configuration update requires dual-key cryptographic authorization from both the NBSE Board of Governors and the Chief Risk Officer.
- **Sub-Paise Truncation Prevention:** Floating-point rounding drift is completely eradicated by enforcing 128-bit fixed-point arithmetic with Banker's Rounding (Round Half to Even) at $10^{-4}$ INR ($0.01$ paise). Residual sub-paise rounding fractions during three-way revenue splitting are programmatically allocated to the Clearing Guarantee Fund (CGF).
- **Statutory Tax Integrity (Income Tax Act & GST Act):** Statutory levies (STT, Stamp Duty, GST, SEBI turnover fees) are strictly separated from exchange operating revenue and credited to designated government liability accounts to ensure error-free statutory tax withholding and remittance.
- **Fail-Closed Verification Invariant:** In the event of an unresolvable currency FX rate or missing tax rate configuration, the engine must fail-closed, rejecting the fee calculation with an explicit error rather than estimating or defaulting to zero.
- **Zero On-Chain PII Compliance:** All transaction receipts and Merkle root proofs anchored to Hyperledger Besu contain solely cryptographic hashes and monetary integers, ensuring zero leakage of investor Personally Identifiable Information (PII) under DPDP Act 2023 and GDPR.

## Acceptance Criteria
- [ ] Rust calculation core executes 0.00% fee (No fee at all) calculations with sub-paise ($10^{-4}$ INR) fixed-point precision in $< 50\mu\text{s}$.
- [ ] Pre-trade fee estimation gRPC endpoint responds in $< 150\mu\text{s}$ at $p99$.
- [ ] Multi-currency FX conversion accurately converts foreign fiat (USD, EUR, AED) and crypto tokens (BTC, ETH, USDT, USDC) to INR base notional using real-time Redis cache rates.
- [ ] Statutory tax deduction accurately computes STT across equity delivery, intraday, futures, and options exercises in accordance with current Indian tax schedules.
- [ ] GST is calculated at exactly 18% on the NBSE exchange transaction fee, correctly split into CGST/SGST or IGST based on user state jurisdiction.
- [ ] Indian Stamp Act duty and SEBI turnover fees are correctly computed and segregated into respective liability ledgers.
- [ ] Programmatic revenue distribution splits 100.00% of collected NBSE fees into Treasury (60%), Clearing Guarantee Fund (25%), and Investor Protection Fund (15%) with zero arithmetic leakage across 1,000,000 test cases.
- [ ] Residual sub-paise remainder fractions from 3-way distribution are deterministically credited to the Clearing Guarantee Fund.
- [ ] Kafka consumer processes incoming execution events, emits `fee.calculated.v1` and `revenue.distributed.v1` events, and logs immutable records to PostgreSQL.
- [ ] Cryptographic SHA-256 Merkle proofs are generated for settlement batches and validated against Hyperledger Besu smart contracts.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `203` (Wallet & Double-Entry Account Ledger Service), Prompt `401` (PostgreSQL Schema), Prompt `402` (Redis Patterns & Caching).
- **Parallel Tasks:** Prompt `208` (Trade Settlement & DvP Orchestration Service), Prompt `210` (Profit-Only Fee & Realized P&L Engine), Prompt `230` (Settlement Guarantee Fund & Default Waterfall Service).
- **Downstream Blockers:** Prompt `223` (Tax Reporting & Capital Gains Statement Service), Prompt `306` (DvP Settlement Smart Contract).
