# 264 - Dynamic Collateral Haircut & Margin Call Notification Engine (Go / Rust)

## Purpose
In a 24/7 tokenized equity and financial derivatives ecosystem, market participants post diverse collateral assets-including cash, central bank digital currencies (e₹ CBDC), Government of India Sovereign Securities (G-Secs), Treasury Bills (T-Bills), Sovereign Gold Bonds (SGBs), and Group I approved equities-to satisfy initial and maintenance margin obligations. Asset values, market liquidity, and systemic volatility fluctuate continuously. Static haircuts fail to protect the clearinghouse against sudden liquidity freezes or gap risk, while excessively conservative static haircuts unnecessarily lock up trader capital and depress market efficiency.

The **Dynamic Collateral Haircut & Margin Call Notification Engine** continuously evaluates asset-specific collateral haircuts in real time. Grounded in the SEBI Comprehensive Risk Management Framework and the J.R. Varma Committee guidelines, the engine dynamically derives asset haircuts from real-time Value at Risk (VaR) rates, Extreme Loss Margin (ELM) rates, Exponentially Weighted Moving Average (EWMA) historical volatility, bid-ask impact cost liquidity ratios, and portfolio concentration factors. 

When dynamic haircut recalculations or adverse market price shifts erode client collateral valuation, the engine executes deterministic multi-stage margin call workflows. It provides legally defensible, verifiable fair notices to investors, manages automated wallet top-ups when pre-authorized, and coordinates with liquidation engines before account insolvency can jeopardize the exchange clearinghouse.

## What You Are Building
A mission-critical, ultra-low-latency distributed microservice (`services/collateral-haircut-engine`) implemented in Rust and Go. It combines an AVX-512/NEON SIMD-accelerated numerical evaluation core with a high-throughput event processing orchestration layer. Concrete deliverables include:

- **Dynamic Haircut Calculation Core (Rust):** A vectorized mathematical library computing dynamic asset-specific haircuts ($H_t$) combining base regulatory floors, 1-day $99\%$ VaR, scrip-level ELM, rolling EWMA volatility, 15-minute order book bid-ask impact costs, and client/exchange-wide concentration penalties.
- **Multi-Asset Collateral Valuation Engine:** An in-memory valuation pipeline tracking pledged user assets across cash, G-Secs, T-Bills, SGBs, and dematerialized equity tokens, computing Total Effective Collateral ($C_{\text{eff}} = \sum Q_i \cdot P_i \cdot (1 - H_{i, t})$) at sub-millisecond latencies.
- **Redis Sorted Set Margin Monitoring Daemon:** An asynchronous scanner utilizing Redis Sorted Sets (`collateral:cure_deadline:zset`, `collateral:margin_utilization:zset`) to index accounts by collateral coverage ratio, enabling instant detection of margin breaches without linear database scans.
- **Multi-Stage Margin Call Workflow Manager:** A state machine coordinating multi-tier escalations across four distinct stages:
  1. *Warning Alert* (Margin Utilization $\ge 75\%$)
  2. *Formal Margin Call & Top-Up Window* (Margin Utilization $\ge 90\%$)
  3. *Restricted Trading / Reduce-Only Mode* (Margin Utilization $\ge 100\%$)
  4. *Liquidation Trigger* (Margin Utilization $\ge 105\%$ or cure window expiry)
- **Automated Collateral Auto-Topup Controller:** An automated sweep mechanism that interfaces with the Cash Wallet Service (Prompt 203) to automatically transfer pre-authorized unencumbered fiat or e₹ CBDC funds into margin collateral to avert liquidation.
- **On-Chain Oracle Relayer Bridge:** A secure cryptographic bridge updating collateral valuation parameters and tier-based haircut floors in the on-chain `CollateralValuationOracle.sol` smart contract on Hyperledger Besu.
- **Audit & Compliance Telemetry Store:** A PostgreSQL repository logging timestamped haircut rate changes, collateral snapshot states, fair notice delivery receipts, and margin call dispute evidence.

## Scope Boundaries
- **In Scope:**
  - Dynamic haircut formula computation incorporating SEBI VaR, ELM, EWMA volatility, order book impact cost, and portfolio concentration.
  - Valuation of approved collateral classes: Cash (0% haircut), e₹ CBDC (0% haircut), Sovereign G-Secs & T-Bills (2-5% dynamic haircut), Sovereign Gold Bonds (10-15% dynamic haircut), and Group I Blue-Chip Equities (20-50% dynamic haircut).
  - Multi-channel margin call alert generation and tracking via Notification Service (Prompt 211).
  - Execution of pre-authorized automated wallet balance top-ups to restore margin health.
  - Tracking of statutory cure period countdown timers and issuance of immutable fair notices before liquidation initiation.
  - Streaming dynamic haircut parameter updates to the VaR Margin Engine (Prompt 229) and Pre-Trade Risk Engine (Prompt 206).
- **Out of Scope / Handled Elsewhere:**
  - Primary portfolio-level parametric VaR and ELM covariance matrix calculation (handled in Prompt 229).
  - Order placement, limit book matching, and trade execution (handled in Prompt 204 and Prompt 205).
  - Physical custodian pledging/demat lock-in via CDSL/NSDL (handled in Prompt 213).
  - Direct execution of forced liquidation order matching on the lit order book (handled in Prompt 208 and Prompt 230).
  - Cash ledger double-entry accounting and banking payment gateway debits (handled in Prompt 203 and Prompt 212).

## Technology to Use
- **Core Languages:** 
  - **Rust 1.78+:** Dynamic haircut mathematical engine, vectorized SIMD numerical calculations, and high-frequency Kafka consumers for market ticks.
  - **Go 1.22+:** gRPC gateway services, Redis sorted set scheduler, margin call lifecycle orchestrator, and external API integrations.
- **Vectorized Math Acceleration:** Rust `core::arch::x86_64` (AVX-512 / AVX2) and `core::arch::aarch64` (NEON) intrinsic vector instructions for fast floating-point array operations across thousands of collateral asset definitions.
- **In-Memory Cache & State Store:** **Redis 7.2+ Cluster** utilizing Redis Sorted Sets (ZSETs) for real-time margin utilization ranking and cure timer scheduling, paired with Redis Hashes for active user collateral asset allocations.
- **Event Streaming & Messaging:** **Apache Kafka** (`librdkafka` / `segmentio/kafka-go`) processing real-time feeds from `market.ticker.v1`, `market.depth.v1`, `rms.var_updates.v1`, and publishing to `collateral.margin_calls.v1` and `collateral.haircut_updates.v1`.
- **Database & Persistence:** **PostgreSQL 16+** with `pgx/v5` and `sqlc` for Go / `sqlx` for Rust, storing asset master haircut configurations, user pledge balances, and legally binding margin call audit trails.
- **Inter-Service Communication:** **gRPC (HTTP/2 with Protobuf)** for sub-millisecond synchronous haircut evaluation queries and status updates.

## Backend / Infra Touchpoints
- **VaR Margin Engine (Prompt 229):** Subscribes to dynamic asset haircuts to establish margin requirements; reciprocally consumes real-time portfolio VaR and ELM rates to adjust asset baseline haircut floors.
- **Market Data Service (Prompt 207):** Ingests Level 2 order book depth and top-of-book trades to evaluate continuous 15-minute VWAP, bid-ask spreads, and market impact costs.
- **Notification Service (Prompt 211):** Emits priority multi-channel dispatch payloads (SMS, Email, Push Notification, WhatsApp, Webhooks) for statutory margin call fair notices.
- **Wallet & Account Service (Prompt 203):** Requests pre-authorized funds reservation and transfers for auto-topup execution; receives alerts on unpledge/withdrawal requests.
- **Custodian & Depository Integration (Prompt 213):** Reconciles pledged asset quantities against CDSL/NSDL depository pledge confirmations.
- **Trade Settlement & Liquidation Service (Prompt 208 / 230):** Hand-off trigger receiving signed liquidation instructions upon cure deadline expiration or severe margin depletion ($\ge 105\%$).

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Collateral Valuation Oracle:** Relays updated dynamic haircut percentages and asset eligibility lists to `CollateralValuationOracle.sol` deployed on Hyperledger Besu. Smart contracts executing atomic Delivery-versus-Payment (DvP) and Settlement Guarantee Fund (SGF, Prompt 315) allocations evaluate solvency using these on-chain haircut parameters.
- **Signed Valuation Attestation:** The engine periodically publishes a cryptographic root (Merkle root or Ed25519 oracle state signature) containing asset haircut parameters, providing deterministic on-chain proof that off-chain margin calculations honor verified market parameters.
- **Zero On-Chain PII Guarantee:** All blockchain transactions refer strictly to pseudonymous `user_id` UUID hashes, public smart contract addresses (`0x...`), and ISO standard `isin` identifiers. Investor PAN, full legal names, and contact coordinates remain strictly secured within off-chain encrypted vaults.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Workspace & Service Layout:** Scaffold `services/collateral-haircut-engine` with a dual-crate / dual-module architecture: a high-performance Rust core library (`crates/haircut-math`) and a robust Go service orchestrator (`cmd/haircut-service`).
2. **Define Protobuf Specifications:** Author `proto/growww/collateral/v1/collateral_haircut.proto` specifying `EvaluateCollateralHaircut`, `GetEffectiveCollateral`, `TriggerAutoTopup`, and streaming Kafka message events `MarginCallEvent` and `HaircutUpdateEvent`.
3. **Generate Go & Rust Stubs:** Compile Protobuf contracts using `protoc` with `prost` / `tonic-build` for Rust and `protoc-gen-go` / `protoc-gen-go-grpc` for Go.
4. **Author PostgreSQL DDL Migrations:** Create migration scripts for `collateral_asset_definitions`, `user_collateral_pledges`, `dynamic_haircut_snapshots`, `margin_call_ledger`, and `auto_topup_authorizations`.
5. **Implement SIMD Haircut Calculation Core (Rust):** Develop vectorized numerical routines in `crates/haircut-math`:
   - Compute base haircut: $H_{\text{base}} = \max(\text{RegulatoryFloor}, \text{VaR}_{99\%} + \text{ELM})$.
   - Compute volatility add-on: $H_{\text{vol}} = \alpha \cdot \left(\frac{\sigma_{\text{30d, ann}}}{\sigma_{\text{benchmark}}} - 1\right)^+$.
   - Compute liquidity/impact cost add-on: $H_{\text{liq}} = \min(25\%, \gamma \cdot \text{ImpactCost}_{\text{15m}})$.
   - Compute final haircut: $H_{\text{final}} = \min(100\%, H_{\text{base}} + H_{\text{vol}} + H_{\text{liq}} + H_{\text{conc}})$.
6. **Implement Market Data Ingestion Pipeline:** Build real-time Kafka consumer ingesting `market.ticker.v1` and `market.depth.v1` to maintain rolling 15-minute order book depth matrices and calculate dynamic bid-ask impact costs for standard institutional lot sizes.
7. **Implement In-Memory Collateral Portfolio State:** Construct a high-performance Redis cache structure maintaining real-time pledged asset balances for each user (`collateral:user:{user_id}:pledges`), with sub-millisecond local caching using Go's `sync.Map` or Rust's `moka`.
8. **Build Collateral Valuation Aggregator:** Implement valuation logic computing Total Effective Collateral ($C_{\text{eff}} = \sum Q_i \cdot P_i \cdot (1 - H_{i, t})$) and Collateral Coverage Ratio ($CCR = \frac{C_{\text{eff}}}{M_{\text{req}}}$) whenever asset prices or haircut rates update.
9. **Implement Redis Sorted Set Margin Monitor:** Maintain a real-time index of accounts in Redis Sorted Sets:
   - Score = Margin Utilization ($U = \frac{M_{\text{req}}}{C_{\text{eff}}} \times 100\%$).
   - A concurrent worker polls the ZSET for accounts where $U \ge 75\%$ to trigger state transitions without full-table database scans.
10. **Implement Multi-Tier Margin Call State Machine:**
    - **Tier 1 (75% <= U < 90%):** Dispatch `WARNING_ALERT` notification via Notification Service.
    - **Tier 2 (90% <= U < 100%):** Issue `FORMAL_MARGIN_CALL`, establish a statutory cure countdown timer (e.g., 2 hours intraday or $T+1$ 11:00 AM IST for EOD), and log notice to `margin_call_ledger`.
    - **Tier 3 (100% <= U < 105%):** Place account in `REDUCE_ONLY` trading mode, blocking opening orders.
    - **Tier 4 (U >= 105% or Cure Timer Expired):** Issue `LIQUIDATION_TRIGGER` event to `collateral.liquidation_notices.v1` for Trade Settlement / Liquidation Engine processing.
11. **Implement Automated Auto-Topup Workflow:** When a Tier 2 Margin Call triggers, check `auto_topup_authorizations`:
    - Calculate required top-up amount: $\Delta C = M_{\text{req}} \cdot 1.20 - C_{\text{eff}}$ (restoring buffer to 80% utilization).
    - Query Wallet Service (Prompt 203) for available cash/e₹ CBDC balance.
    - If funds available up to user-configured maximum limit, execute atomic fund pledge reservation via gRPC and credit collateral ledger.
    - If auto-topup succeeds, transition margin call status to `CURED_VIA_AUTOTOPUP`.
12. **Build Fair Notice Delivery & Telemetry Logger:** Guarantee compliance with SEBI investor protection mandates:
    - Dispatch redundant notifications across SMS, WhatsApp, and email via Notification Service.
    - Record delivery receipts, provider message IDs, timestamps, and investor read acknowledgments in `margin_call_ledger`.
13. **Implement On-Chain Oracle Update Relayer:** Construct an asynchronous worker that batches significant haircut updates ($|\Delta H| \ge 1.0\%$) and submits multi-sig signed transactions to `CollateralValuationOracle.sol` on Hyperledger Besu.
14. **Instrument Prometheus Metrics & Tracing:** Expose operational metrics including `haircut_computation_duration_micros`, `collateral_valuation_total_inr`, `active_margin_calls_by_tier`, `auto_topup_success_total`, and `cure_window_breaches_total`.
15. **Write Unit, SIMD, and Integration Test Suite:** Write comprehensive unit tests for mathematical haircut formulas, SIMD edge cases (zero volume, inverted spreads), Redis ZSET concurrency, and an end-to-end simulation of margin call escalation and automated top-up recovery.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/collateral/v1/collateral_haircut.proto`)
```protobuf
syntax = "proto3";

package growww.collateral.v1;

option go_package = "growww/collateral/v1;collateralv1";

service CollateralHaircutService {
  rpc EvaluateCollateralHaircut (HaircutEvaluationRequest) returns (HaircutEvaluationResponse);
  rpc GetEffectiveCollateral (GetEffectiveCollateralRequest) returns (GetEffectiveCollateralResponse);
  rpc SetAutoTopupPreference (SetAutoTopupPreferenceRequest) returns (SetAutoTopupPreferenceResponse);
  rpc GetMarginCallStatus (GetMarginCallStatusRequest) returns (GetMarginCallStatusResponse);
  rpc StreamHaircutUpdates (StreamHaircutUpdatesRequest) returns (stream HaircutUpdateEvent);
}

enum CollateralAssetClass {
  COLLATERAL_ASSET_CLASS_UNSPECIFIED = 0;
  COLLATERAL_ASSET_CLASS_CASH = 1;
  COLLATERAL_ASSET_CLASS_CBDC = 2;
  COLLATERAL_ASSET_CLASS_GSEC = 3;
  COLLATERAL_ASSET_CLASS_TBILL = 4;
  COLLATERAL_ASSET_CLASS_SGB = 5;
  COLLATERAL_ASSET_CLASS_EQUITY_GROUP_1 = 6;
}

enum MarginCallSeverity {
  MARGIN_CALL_SEVERITY_UNSPECIFIED = 0;
  MARGIN_CALL_SEVERITY_WARNING = 1;      // 75% <= Utilization < 90%
  MARGIN_CALL_SEVERITY_CRITICAL = 2;     // 90% <= Utilization < 100% (Formal Call)
  MARGIN_CALL_SEVERITY_BREACH = 3;       // 100% <= Utilization < 105% (Reduce-Only)
  MARGIN_CALL_SEVERITY_LIQUIDATION = 4;  // Utilization >= 105% or Timer Expired
}

enum MarginCallStatus {
  MARGIN_CALL_STATUS_UNSPECIFIED = 0;
  MARGIN_CALL_STATUS_ACTIVE = 1;
  MARGIN_CALL_STATUS_CURED_VIA_DEPOSIT = 2;
  MARGIN_CALL_STATUS_CURED_VIA_AUTOTOPUP = 3;
  MARGIN_CALL_STATUS_CURED_VIA_MARKET = 4;
  MARGIN_CALL_STATUS_EXPIRED_LIQUIDATED = 5;
}

message HaircutEvaluationRequest {
  string isin = 1;
  CollateralAssetClass asset_class = 2;
  string market_price = 3;             // e.g. "2450.50"
  string var_rate = 4;                 // e.g. "0.1250" (12.5% from VaR Engine)
  string elm_rate = 5;                 // e.g. "0.0350" (3.5% Extreme Loss Margin)
  string rolling_30d_volatility = 6;   // Annualized volatility
  string bid_ask_impact_cost = 7;      // 15m order book impact cost percentage
  string concentration_ratio = 8;      // Ratio of asset to total collateral pool
}

message HaircutEvaluationResponse {
  string isin = 1;
  string dynamic_haircut_rate = 2;     // e.g. "0.2250" (22.5%)
  string base_haircut = 3;
  string volatility_addon = 4;
  string liquidity_addon = 5;
  string concentration_penalty = 6;
  int64 computed_at_unix_ms = 7;
}

message PledgedAssetItem {
  string isin = 1;
  CollateralAssetClass asset_class = 2;
  string quantity = 3;
  string last_traded_price = 4;
  string dynamic_haircut_rate = 5;
  string gross_valuation = 6;
  string effective_valuation = 7;      // Gross * (1 - Haircut)
}

message GetEffectiveCollateralRequest {
  string user_id = 1;
}

message GetEffectiveCollateralResponse {
  string user_id = 1;
  string total_gross_collateral_inr = 2;
  string total_effective_collateral_inr = 3;
  string total_margin_requirement_inr = 4;
  string margin_utilization_percentage = 5; // (Req / Effective) * 100
  MarginCallSeverity current_severity = 6;
  repeated PledgedAssetItem pledged_assets = 7;
  int64 updated_at_unix_ms = 8;
}

message MarginCallEvent {
  string margin_call_id = 1;
  string user_id = 2;
  MarginCallSeverity severity = 3;
  string total_effective_collateral_inr = 4;
  string total_margin_requirement_inr = 5;
  string margin_shortfall_inr = 6;
  string margin_utilization_percentage = 7;
  int64 cure_deadline_unix_ms = 8;
  string notice_reference_number = 9;
  int64 triggered_at_unix_ms = 10;
}

message HaircutUpdateEvent {
  string isin = 1;
  CollateralAssetClass asset_class = 2;
  string old_haircut_rate = 3;
  string new_haircut_rate = 4;
  string trigger_reason = 5;
  int64 timestamp_unix_ms = 6;
}

message SetAutoTopupPreferenceRequest {
  string user_id = 1;
  bool auto_topup_enabled = 2;
  string max_topup_per_call_inr = 3;
  string daily_cumulative_cap_inr = 4;
  string preferred_source_wallet = 5; // "CASH" or "CBDC"
}

message SetAutoTopupPreferenceResponse {
  string user_id = 1;
  bool auto_topup_enabled = 2;
  string status = 3;
  int64 updated_at_unix_ms = 4;
}

message GetMarginCallStatusRequest {
  string user_id = 1;
}

message GetMarginCallStatusResponse {
  string user_id = 1;
  bool has_active_margin_call = 2;
  string margin_call_id = 3;
  MarginCallSeverity severity = 4;
  MarginCallStatus status = 5;
  string shortfall_inr = 6;
  int64 cure_deadline_unix_ms = 7;
  int64 remaining_cure_seconds = 8;
  bool auto_topup_attempted = 9;
  string auto_topup_result = 10;
}

message StreamHaircutUpdatesRequest {
  repeated string isins = 1;
}
```

### PostgreSQL Database Schema DDL (`migrations/000264_collateral_haircut_schema.up.sql`)
```sql
CREATE TYPE collateral_asset_class_enum AS ENUM (
    'CASH',
    'CBDC',
    'GSEC',
    'TBILL',
    'SGB',
    'EQUITY_GROUP_1'
);

CREATE TYPE margin_call_severity_enum AS ENUM (
    'WARNING',
    'CRITICAL',
    'BREACH',
    'LIQUIDATION'
);

CREATE TYPE margin_call_status_enum AS ENUM (
    'ACTIVE',
    'CURED_VIA_DEPOSIT',
    'CURED_VIA_AUTOTOPUP',
    'CURED_VIA_MARKET',
    'EXPIRED_LIQUIDATED'
);

-- Master registry of approved collateral assets and regulatory floors
CREATE TABLE collateral_asset_definitions (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    asset_class collateral_asset_class_enum NOT NULL,
    regulatory_floor_rate NUMERIC(6, 4) NOT NULL DEFAULT 0.2000,
    current_dynamic_haircut NUMERIC(6, 4) NOT NULL DEFAULT 0.2000,
    current_market_price NUMERIC(18, 4) NOT NULL,
    current_var_rate NUMERIC(6, 4) NOT NULL,
    current_elm_rate NUMERIC(6, 4) NOT NULL,
    rolling_30d_volatility NUMERIC(8, 4) NOT NULL,
    impact_cost_15m NUMERIC(6, 4) NOT NULL,
    is_eligible_for_pledge BOOLEAN NOT NULL DEFAULT TRUE,
    max_exchange_pool_quantity NUMERIC(18, 4) NOT NULL,
    current_pledged_quantity NUMERIC(18, 4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User pledged assets repository
CREATE TABLE user_collateral_pledges (
    pledge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    isin VARCHAR(12) NOT NULL REFERENCES collateral_asset_definitions(isin),
    depository_reference VARCHAR(64) NOT NULL UNIQUE, -- CDSL/NSDL pledge confirmation ID
    pledged_quantity NUMERIC(18, 4) NOT NULL CHECK (pledged_quantity >= 0),
    is_locked BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Real-time audit log of computed dynamic haircuts
CREATE TABLE dynamic_haircut_snapshots (
    snapshot_id BIGSERIAL PRIMARY KEY,
    isin VARCHAR(12) NOT NULL REFERENCES collateral_asset_definitions(isin),
    dynamic_haircut_rate NUMERIC(6, 4) NOT NULL,
    base_haircut NUMERIC(6, 4) NOT NULL,
    volatility_addon NUMERIC(6, 4) NOT NULL,
    liquidity_addon NUMERIC(6, 4) NOT NULL,
    concentration_penalty NUMERIC(6, 4) NOT NULL,
    market_price NUMERIC(18, 4) NOT NULL,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Formal margin call lifecycle and legal notice tracking
CREATE TABLE margin_call_ledger (
    margin_call_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notice_reference_number VARCHAR(64) NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    severity margin_call_severity_enum NOT NULL,
    status margin_call_status_enum NOT NULL DEFAULT 'ACTIVE',
    effective_collateral_inr NUMERIC(18, 4) NOT NULL,
    margin_requirement_inr NUMERIC(18, 4) NOT NULL,
    shortfall_amount_inr NUMERIC(18, 4) NOT NULL,
    utilization_percentage NUMERIC(6, 2) NOT NULL,
    cure_deadline TIMESTAMPTZ NOT NULL,
    
    -- Fair notice multi-channel dispatch audit
    sms_notification_id VARCHAR(64),
    sms_delivered_at TIMESTAMPTZ,
    email_notification_id VARCHAR(64),
    email_delivered_at TIMESTAMPTZ,
    whatsapp_notification_id VARCHAR(64),
    whatsapp_delivered_at TIMESTAMPTZ,
    in_app_acknowledged_at TIMESTAMPTZ,
    
    -- Resolution telemetry
    cured_at TIMESTAMPTZ,
    resolution_amount_inr NUMERIC(18, 4),
    liquidation_handover_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User configuration for automated collateral replenishment
CREATE TABLE auto_topup_authorizations (
    user_id UUID PRIMARY KEY,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    max_per_call_limit_inr NUMERIC(18, 4) NOT NULL CHECK (max_per_call_limit_inr > 0),
    daily_cumulative_cap_inr NUMERIC(18, 4) NOT NULL CHECK (daily_cumulative_cap_inr >= max_per_call_limit_inr),
    current_day_utilized_inr NUMERIC(18, 4) NOT NULL DEFAULT 0,
    preferred_source_wallet VARCHAR(16) NOT NULL DEFAULT 'CASH', -- 'CASH' or 'CBDC'
    last_reset_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for ultra-fast query execution
CREATE INDEX idx_user_pledges_user_id ON user_collateral_pledges(user_id);
CREATE INDEX idx_margin_call_active ON margin_call_ledger(user_id, status) WHERE status = 'ACTIVE';
CREATE INDEX idx_margin_call_deadline ON margin_call_ledger(cure_deadline) WHERE status = 'ACTIVE';
CREATE INDEX idx_haircut_snapshots_isin ON dynamic_haircut_snapshots(isin, evaluated_at DESC);
```

## Security & Compliance Notes
- **SEBI Comprehensive Risk Management Framework Compliance:** Asset haircuts strictly abide by SEBI Master Circular guidelines. Liquid blue-chip equities under Group I possess an absolute regulatory floor ($H_{\text{floor}} \ge 20\%$), while G-Secs enforce modified-duration-linked floors ($2\%$ to $5\%$). At no point can algorithmic calculations reduce haircuts below SEBI-prescribed floors.
- **Fair Notice Before Forced Liquidation:** Indian securities jurisprudence and SEBI investor protection mandates require brokers to provide demonstrable, verifiable fair notice with a reasonable cure window prior to liquidating client collateral. The engine creates immutable audit records (`margin_call_ledger`) capturing dispatch timestamps, telecommunication provider message IDs, and multi-channel delivery receipts (SMS, WhatsApp, registered email).
- **Auto-Topup Dual-Authorization & Daily Limits:** To prevent runaway wallet draining during extreme flash crashes, automated top-ups require prior explicit investor consent, are bounded by both single-call caps and daily cumulative limits, and can only draw from unencumbered cash or e₹ CBDC wallets.
- **Anti-Manipulation Impact Cost Filtering:** To avoid malicious manipulation of dynamic haircuts via spoofed thin order books, bid-ask impact cost calculations use 15-minute volume-weighted averages and filter out transient quote anomalies exceeding predefined deviation thresholds.
- **Zero On-Chain PII Guarantee:** Oracle updates pushed to Hyperledger Besu publish strictly pseudonymized ISIN parameters and aggregated pool stats. Personal investor identities, PANs, phone numbers, and notification traces are isolated strictly within off-chain encrypted databases.

## Acceptance Criteria
- [ ] SIMD dynamic haircut routine computes haircuts for 2,000 active collateral assets in $< 5\text{ms}$ on standard CPU architecture.
- [ ] Calculated haircut rate never falls below regulatory baseline floors (Cash: 0%, G-Secs: 2%, Blue-chip Equities: 20%).
- [ ] Real-time order book impact cost surges immediately elevate dynamic haircut within 250ms of tick feed ingestion.
- [ ] Collateral valuation accurately updates Total Effective Collateral across mixed portfolios of cash, G-Secs, SGBs, and tokenized equities.
- [ ] Redis Sorted Set scanner identifies accounts crossing $75\%$, $90\%$, and $100\%$ margin utilization without degraded throughput.
- [ ] Tier 2 Margin Call triggers multi-channel fair notices (SMS, email, push, WhatsApp) via Notification Service within 1 second.
- [ ] Pre-authorized auto-topup successfully draws funds from Cash Wallet Service and restores collateral coverage to $\ge 80\%$ utilization.
- [ ] Cure countdown expiration deterministically emits a signed liquidation handover event to Kafka topic `collateral.liquidation_notices.v1`.
- [ ] On-chain `CollateralValuationOracle.sol` smart contract parameters update reliably via relayer whenever haircut parameters diverge by $\ge 1\%$.
- [ ] Abrupt process crashes recover state cleanly from Redis and PostgreSQL without duplicate margin call alerts or missed cure expirations.

## Suggested Order / Dependencies
- **Prerequisites:**
  - 103 (API Design Standards)
  - 104 (Event Schema and Kafka Topic Standards)
  - 203 (Wallet & Account Service)
  - 207 (Market Data Service)
  - 211 (Notification Service)
  - 213 (Custodian & Depository Integration Service)
  - 229 (Real-Time VaR & ELM Engine)
- **Parallel Tasks:**
  - 206 (Pre-Trade Risk Engine)
  - 230 (Settlement Guarantee Fund & Default Waterfall Engine)
  - 315 (Settlement Guarantee Fund Smart Contract)
- **Downstream Blockers:**
  - 208 (Trade Settlement & Liquidation Service)
  - 509 (Mobile Margin Call & Auto-Topup Management UI)
  - 604 (Web Institutional Risk & Collateral Desk)
