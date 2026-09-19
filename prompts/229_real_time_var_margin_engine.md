# 229 - Real-Time Value at Risk (VaR) & Extreme Loss Margin (ELM) Engine (Rust / Redis / SIMD)

## Purpose
In financial market infrastructure, pre-trade and intra-day risk management must guarantee that counterparty exposure is continuously capitalized to prevent cascading settlement defaults. Under SEBI Master Circulars for Clearing Corporations (SECC Regulations) and the J.R. Varma Committee Risk Management Framework, clearing entities must enforce real-time multi-tier margining across all trading participants.

The **Real-Time Value at Risk (VaR) & Extreme Loss Margin (ELM) Engine** continuously computes and enforces portfolio-level margin requirements at sub-millisecond latencies. By maintaining dynamic Exponentially Weighted Moving Average (EWMA) volatility matrices, calculating 99% confidence 1-day Value at Risk, layering asset-specific Extreme Loss Margins, Mark-to-Market (MTM) margins, and client-level exposure caps, the engine prevents under-collateralized order placement, initiates intra-day margin calls, and triggers automated orderly liquidations before member insolvency threatens the clearinghouse.

## What You Are Building
A ultra-low-latency, high-throughput risk calculation microservice (`services/var-margin-engine`) implemented in Rust with AVX-512/NEON SIMD acceleration, backed by Redis in-memory portfolio state caches and Apache Kafka event streams. Concrete deliverables include:
- **High-Performance SIMD VaR Matrix Engine:** A Rust computation core calculating portfolio parametric VaR and historical simulation VaR ($99\%$ confidence interval, $1$-day horizon) across thousands of positions in $< 250\mu\text{s}$ using vector CPU instructions.
- **Dynamic EWMA Volatility Calculator:** Real-time volatility updater utilizing the standard RiskMetrics EWMA model ($\lambda = 0.94$ for intra-day, $\lambda = 0.99$ for inter-day) recalculating price variance upon every incoming market trade tick.
- **Extreme Loss Margin (ELM) & MTM Processor:** Automated rule engine applying statutory ELM (minimum $3.5\%$ for standard equities, scaled dynamically for volatile scrips) and real-time MTM loss accounting against available margin collateral.
- **Sub-Millisecond gRPC Evaluation Gateway:** Exposing `EvaluateMarginRequirement` and `CheckPreOrderMargin` RPCs for pre-trade verification and real-time intra-day portfolio margin health checks.
- **Automated Intra-Day Margin Call & Liquidation Trigger:** Event-driven monitor issuing margin replenishment alerts at $70\%$ and $85\%$ margin utilization, and signaling automated position square-off at $\ge 100\%$ margin breach.
- **PostgreSQL Persistence Repository:** Audit repository storing snapshot volatility matrices, member margin account ledger records, and immutable risk breach telemetry.

## Scope Boundaries
- **In Scope:**
 - Portfolio-level parametric and historical simulation VaR calculations.
 - Intra-day EWMA price volatility tracking for all listed equity ISINs.
 - Extreme Loss Margin (ELM), Base Minimum Capital (BMC), and Mark-to-Market (MTM) margin calculation.
 - Multi-asset collateral haircut evaluation (cash, e₹ CBDC, approved sovereign bonds, tokenized blue-chip equities).
 - Intra-day margin utilization tracking, margin call event emission, and liquidation triggers.
 - Pre-order incremental margin requirement checks.
- **Out of Scope / Handled Elsewhere:**
 - Basic single-order fat-finger and static circuit-breaker price band checks (handled in Prompt 206).
 - Cash wallet balance reservation and fiat banking balance hold (handled in Prompt 203).
 - In-memory order book matching (handled in Prompt 205).
 - Post-trade atomic Delivery-versus-Payment (DvP) execution (handled in Prompt 208 / Prompt 306).
 - Settlement Guarantee Fund default waterfall execution (handled in Prompt 230 / Prompt 315).

## Technology to Use
- **Primary Language & Core Framework:** **Rust 1.78+** utilizing `tokio` asynchronous runtime, `tonic` gRPC framework, and `rayon` data-parallel execution library for parallel portfolio simulations.
- **Vector Acceleration:** Rust `core::arch::x86_64` (AVX-512 / AVX2) and `core::arch::aarch64` (NEON) intrinsic vector instructions for vectorized covariance matrix multiplications.
- **In-Memory Cache & State Store:** **Redis 7.2+ Cluster** with pipelined reads and custom Lua scripts maintaining active user portfolio holdings, deposited collateral valuations, and real-time margin usage percentages.
- **Event Streaming:** **Apache Kafka** via `rdkafka` (Rust bindings to `librdkafka`) ingesting L2 market ticker ticks and trade fills, and emitting margin breach events.
- **Database & Storage:** **PostgreSQL 16+** with `sqlx` async connection pooling for persistent parameter definitions, historical volatility logs, and margin call audit trails.

## Backend / Infra Touchpoints
- **Redis 7.2 Keys:** `rms:portfolio:{user_id}:positions`, `rms:collateral:{user_id}:valuation`, `rms:volatility:{isin}:ewma`, `rms:margin:{user_id}:utilization`, `rms:config:elm_rates`.
- **PostgreSQL 16 Tables:** `var_parameters`, `asset_volatilities`, `margin_collateral_allocations`, `intra_day_margin_snapshots`, `margin_breach_events`.
- **Apache Kafka Topics:** Consumes from `market.ticker.v1`, `matching.trades.v1`, `wallet.collateral.v1`; publishes to `rms.margin_calls.v1`, `rms.liquidation_triggers.v1`, `rms.risk_alerts.v1`.
- **Order Service (Prompt 204) & Risk Service (Prompt 206):** Ingests pre-order margin validation requests before order commitment.
- **Trade Settlement Service (Prompt 208):** Receives settled trade position deltas to adjust portfolio composition.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Collateral Attestation Verification:** Verifies tokenized collateral balances committed to the on-chain Settlement Guarantee Fund contract (`SettlementGuaranteeFund.sol`, Prompt 315) on Hyperledger Besu.
- **Zero On-Chain PII Guarantee:** Operates strictly with `user_id` UUIDs, `ledger_address` (0x...), and `isin` identifiers. All personal identification details, PANs, and bank accounts are isolated off-chain.
- **On-Chain Liquidation Signatures:** When a catastrophic margin breach occurs ($\ge 100\%$ margin exhaustion with unfulfilled margin call), the RMS engine emits a cryptographically signed liquidation authorization payload consumed by the Settlement Relayer to execute forced on-chain DvP liquidation.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Rust Workspace:** Scaffold `services/var-margin-engine` using Cargo with optimized release profiles (`lto = "fat"`, `codegen-units = 1`, `target-cpu = "native"`).
2. **Define Protobuf Schema:** Author `proto/growww/rms/v1/margin_engine.proto` specifying `EvaluateMarginRequirement`, `CheckPreOrderMargin`, `UpdateVolatilityParameters`, and `GetPortfolioRiskMetrics`.
3. **Generate gRPC Stubs:** Compile Protobuf definitions using `tonic-build` into idiomatic Rust server and client traits.
4. **Design PostgreSQL Schema:** Write Flyway/SQL migrations establishing tables for `var_parameters`, `asset_volatilities`, `collateral_haircuts`, and `margin_breach_events`.
5. **Implement SIMD Covariance & VaR Engine:** Write vectorized mathematical routines in `src/math/var.rs`:
 - Compute variance-covariance matrix: $\Sigma = X^T X / (N-1)$.
 - Parametric VaR formula: $\text{VaR}_{\alpha} = Z_{\alpha} \times \sqrt{w^T \Sigma w} \times \sqrt{t}$, where $Z_{0.99} = 2.3263$, $t = 1\text{ day}$.
 - Historical simulation module sorting simulated portfolio P&L vectors to extract the 99th percentile empirical loss.
6. **Implement Real-Time EWMA Volatility Service:** Implement streaming volatility calculator updating scrip variance:
   $$\sigma_{t}^2 = \lambda \sigma_{t-1}^2 + (1 - \lambda) r_t^2$$
   where $r_t = \ln(P_t / P_{t-1})$ and $\lambda = 0.94$.
7. **Implement Extreme Loss Margin (ELM) Module:** Calculate ELM as:
   $$\text{ELM} = \sum_{i=1}^n |Q_i \times P_i| \times \max(\text{ELM\_Rate}_i, 3.5\%)$$
   incorporating scrip-specific volatility multipliers for High Beta securities.
8. **Build Collateral Valuation & Haircut Engine:** Calculate Total Effective Collateral ($C_{\text{eff}}$) across deposited assets:
   $$C_{\text{eff}} = \sum \text{AssetValue}_j \times (1 - \text{Haircut}_j)$$
   applying $0\%$ haircut for cash/e₹ CBDC, $2\%$ for G-Secs, and $20\%-40\%$ for equity tokens.
9. **Implement In-Memory Portfolio Cache with Redis:** Maintain user position state in Redis with sub-millisecond local LRU caching using `moka` crate to avoid unnecessary network hops.
10. **Build Pre-Order Margin Verification Pipeline:** Implement `CheckPreOrderMargin` RPC computing the incremental post-order margin $\Delta M$:
 - Check if Available Margin $\ge (\text{Current Margin} + \Delta M)$.
 - Return instant pass/fail verdict in $< 400\mu\text{s}$.
11. **Implement Intra-Day Margin Monitoring Daemon:** Build asynchronous Kafka consumer processing real-time trade fills and market ticks, recalculating margin utilization:
    $$\text{Utilization} = \frac{\text{VaR} + \text{ELM} + \text{MTM Loss}}{C_{\text{eff}}} \times 100\%$$
12. **Build Automated Margin Call & Liquidation Workflow:**
 - If $\text{Utilization} \ge 70\%$: Emit `MARGIN_CALL_WARNING` to Kafka topic `rms.margin_calls.v1`.
 - If $\text{Utilization} \ge 85\%$: Emit `MARGIN_CALL_CRITICAL` requiring immediate collateral infusion within $T+0$ cut-off.
 - If $\text{Utilization} \ge 100\%$: Emit `LIQUIDATION_TRIGGER` to `rms.liquidation_triggers.v1` and lock member order placement.
13. **Configure Prometheus Telemetry:** Export real-time metrics: `var_calculation_duration_nanos`, `active_portfolios_monitored`, `margin_utilization_ratio`, `margin_breaches_total`.
14. **Write SIMD Unit & Benchmarking Tests:** Write Rust benchmarks (`criterion`) verifying that 1,000-position portfolio VaR computes in $< 250\mu\text{s}$ with zero memory leaks.

## Interfaces / Contracts

### Protobuf Definition (`margin_engine.proto`)
```protobuf
syntax = "proto3";

package growww.rms.v1;

option go_package = "growww/rms/v1;rmsv1";

service MarginEngineService {
  rpc CheckPreOrderMargin (CheckPreOrderMarginRequest) returns (CheckPreOrderMarginResponse);
  rpc EvaluatePortfolioMargin (EvaluatePortfolioMarginRequest) returns (EvaluatePortfolioMarginResponse);
  rpc GetPortfolioRiskMetrics (GetPortfolioRiskMetricsRequest) returns (GetPortfolioRiskMetricsResponse);
  rpc UpdateAssetVolatility (UpdateAssetVolatilityRequest) returns (UpdateAssetVolatilityResponse);
}

enum MarginCheckStatus {
  MARGIN_CHECK_STATUS_UNSPECIFIED = 0;
  MARGIN_CHECK_STATUS_APPROVED = 1;
  MARGIN_CHECK_STATUS_INSUFFICIENT_MARGIN = 2;
  MARGIN_CHECK_STATUS_PORTFOLIO_LOCKED = 3;
  MARGIN_CHECK_STATUS_EXPOSURE_CAP_EXCEEDED = 4;
}

enum MarginAlertLevel {
  MARGIN_ALERT_NORMAL = 0;
  MARGIN_ALERT_WARNING_70 = 1;
  MARGIN_ALERT_CRITICAL_85 = 2;
  MARGIN_ALERT_BREACH_100 = 3;
}

message CheckPreOrderMarginRequest {
  string request_id = 1;
  string user_id = 2;
  string ledger_address = 3;
  string isin = 4;
  string side = 5; // BUY, SELL
  string order_type = 6; // LIMIT, MARKET
  string price = 7; // Decimal string INR
  string quantity = 8; // Fractional decimal string
}

message CheckPreOrderMarginResponse {
  string request_id = 1;
  string user_id = 2;
  MarginCheckStatus status = 3;
  string required_incremental_margin = 4;
  string available_margin = 5;
  string total_margin_after_order = 6;
  string projected_utilization_percent = 7;
  string rejection_reason = 8;
  int64 evaluated_at_unix_ns = 9;
}

message EvaluatePortfolioMarginRequest {
  string user_id = 2;
  bool include_historical_simulation = 3;
}

message EvaluatePortfolioMarginResponse {
  string user_id = 1;
  string total_portfolio_value = 2;
  string effective_collateral_value = 3;
  string parametric_var_99 = 4;
  string historical_var_99 = 5;
  string extreme_loss_margin = 6;
  string mark_to_market_loss = 7;
  string total_margin_required = 8;
  string margin_utilization_percent = 9;
  MarginAlertLevel alert_level = 10;
  int64 evaluated_at_unix_ns = 11;
}

message GetPortfolioRiskMetricsRequest {
  string user_id = 1;
}

message GetPortfolioRiskMetricsResponse {
  string user_id = 1;
  string gross_exposure = 2;
  string net_exposure = 3;
  string portfolio_beta = 4;
  string daily_portfolio_variance = 5;
  repeated PositionRiskItem positions = 6;
}

message PositionRiskItem {
  string isin = 1;
  string symbol = 2;
  string quantity = 3;
  string current_price = 4;
  string market_value = 5;
  string daily_volatility_sigma = 6;
  string elm_rate_percent = 7;
  string standalone_var = 8;
}

message UpdateAssetVolatilityRequest {
  string isin = 1;
  string ewma_volatility_sigma = 2;
  string elm_rate_percent = 3;
  string updated_by = 4;
}

message UpdateAssetVolatilityResponse {
  bool success = 1;
  int64 applied_at_unix_ns = 2;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE var_parameters (
    parameter_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    confidence_level NUMERIC(5, 4) NOT NULL DEFAULT 0.9900, -- 99%
    time_horizon_days INTEGER NOT NULL DEFAULT 1,
    ewma_lambda_intraday NUMERIC(5, 4) NOT NULL DEFAULT 0.9400,
    ewma_lambda_interday NUMERIC(5, 4) NOT NULL DEFAULT 0.9900,
    min_elm_rate_percent NUMERIC(5, 2) NOT NULL DEFAULT 3.50, -- 3.5% minimum ELM
    warning_threshold_percent NUMERIC(5, 2) NOT NULL DEFAULT 70.00,
    critical_threshold_percent NUMERIC(5, 2) NOT NULL DEFAULT 85.00,
    liquidation_threshold_percent NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE asset_volatilities (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    current_price NUMERIC(18, 4) NOT NULL,
    daily_volatility_sigma NUMERIC(10, 6) NOT NULL,
    annualized_volatility NUMERIC(10, 6) NOT NULL,
    elm_rate_percent NUMERIC(5, 2) NOT NULL DEFAULT 3.50,
    haircut_percent NUMERIC(5, 2) NOT NULL DEFAULT 20.00,
    last_price_update TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE margin_collateral_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    collateral_type VARCHAR(30) NOT NULL, -- CASH_INR, CBDC_E_RUPEE, GSEC_SOVEREIGN, TOKENIZED_EQUITY
    asset_identifier VARCHAR(50) NOT NULL, -- INR, ISIN, CBDC_TOKEN_ID
    raw_amount NUMERIC(24, 6) NOT NULL,
    market_price NUMERIC(18, 4) NOT NULL DEFAULT 1.0000,
    haircut_percent NUMERIC(5, 2) NOT NULL,
    effective_value_inr NUMERIC(18, 2) NOT NULL,
    is_locked_for_sgf BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE margin_breach_events (
    breach_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    utilization_percent NUMERIC(7, 2) NOT NULL,
    required_margin_inr NUMERIC(18, 2) NOT NULL,
    effective_collateral_inr NUMERIC(18, 2) NOT NULL,
    alert_level VARCHAR(30) NOT NULL, -- WARNING_70, CRITICAL_85, LIQUIDATION_100
    triggered_action VARCHAR(50) NOT NULL, -- NOTIFICATION_SENT, ORDER_ENTRY_BLOCKED, FORCED_LIQUIDATION_INITIATED
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collateral_user ON margin_collateral_allocations(user_id);
CREATE INDEX idx_breach_user_time ON margin_breach_events(user_id, created_at DESC);
```

## Security & Compliance Notes
- **SEBI Master Circular for Clearing Corporations Mandate:** Clearing corporations must collect Initial Margins (VaR + ELM) on an upfront basis with real-time intra-day peak margin monitoring across all client accounts.
- **Fail-Closed Pre-Trade Invariant:** If the Margin Engine fails to compute margin or encounters a timeout ($> 5\text{ms}$), downstream order routers must immediately reject the order (`MARGIN_CHECK_STATUS_PORTFOLIO_LOCKED`).
- **Precision Floating-Point Safeguard:** All financial value accounting and collateral deductions use high-precision fixed-point or `rust_decimal::Decimal` (28-digit precision) to prevent IEEE-754 floating-point rounding errors.
- **Zero In-Flight Margin Leakage:** Real-time margin holds are locked atomically in Redis before matching engine intake and released only upon trade cancellation or DvP settlement finality.

## Acceptance Criteria
- [ ] SIMD-accelerated 1-day 99% Parametric VaR computes across a 500-asset portfolio in $< 250\mu\text{s}$.
- [ ] Pre-order incremental margin check responds via gRPC in $< 500\mu\text{s}$ at $p99$.
- [ ] EWMA dynamic volatility updates correctly on every streaming market trade tick with $\lambda = 0.94$.
- [ ] Minimum 3.5% Extreme Loss Margin (ELM) enforced on all equity scrips, with higher ELM for high-volatility securities.
- [ ] Multi-asset collateral haircuts correctly applied (0% for Cash/e₹, 2% for G-Secs, $\ge 20\%$ for equity).
- [ ] Member margin utilization reaching 70%, 85%, and 100% emits respective Kafka alerts and updates account state within 10ms.
- [ ] Portfolio reaching 100% margin utilization locks order intake and triggers automated liquidation workflow.
- [ ] PostgreSQL audit tables record every margin breach and collateral state change immutably.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Standards), Prompt `203` (Wallet Account Service), Prompt `206` (Pre-Trade Risk Engine), Prompt `402` (Redis Patterns).
- **Parallel Tasks:** Prompt `205` (Matching Engine), Prompt `207` (Market Data Service).
- **Downstream Blockers:** Prompt `204` (Order Service margin integration), Prompt `230` (Settlement Guarantee Fund Service).
