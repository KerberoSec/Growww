# 239 - Real-Time Options Pricing, Greeks & Volatility Risk Engine (C++ / Rust / SIMD)

## Purpose
Operating a continuous, 24/7 on-chain equity and index options trading platform demands real-time computational infrastructure capable of handling non-linear payout profiles, extreme volatility shifts, and complex multi-leg portfolio exposures. Under SEBI Master Circulars on Derivatives Risk Management, the J.R. Varma Committee recommendations, and CPMI-IOSCO PFMI principles, an institutional derivatives exchange must enforce continuous pre-trade risk evaluation, accurate options valuation, and real-time Greek risk limits across all market participants.

The **Real-Time Options Pricing, Greeks & Volatility Risk Engine** serves as the ultra-low-latency quantitative core of Growww's derivatives infrastructure. It computes deterministic analytical option pricing for European and American equity/index options, calculates first-order and second-order Greeks (Delta, Gamma, Vega, Theta, Rho) utilizing vectorized SIMD instructions, dynamically calibrates arbitrage-free Implied Volatility (IV) smile and surface curves from streaming market order books, and conducts sub-millisecond pre-trade portfolio Greek risk and margin validations.

By evaluating portfolio risk at microsecond speeds, the engine protects the exchange and market makers from catastrophic gamma-squeeze events, prevents under-collateralized options writing, and maintains 24/7 market solvency.

## What You Are Building
A high-performance computational microservice (`services/options-risk-engine`) implemented in C++20 / Rust with AVX-512, AVX2, and ARM NEON hardware acceleration, backed by Redis in-memory caches, PostgreSQL persistence, and Apache Kafka event streams. Concrete deliverables include:
- **Vectorized Analytical Pricing Engine:** Analytical evaluation of Black-Scholes-Merton (BSM) for European options and Bjerksund-Stensland (2002) closed-form approximations for American equity options with continuous dividend yields, achieving $< 50\text{ns}$ per contract pricing.
- **SIMD Greek Calculation Pipeline:** Hardware-vectorized computation of primary Greeks ($\Delta, \Gamma, \mathcal{V}, \Theta, \rho$) and higher-order risk sensitivities (Charm, Vanna, Volga) over entire strike/expiry grids ($> 10,000$ contracts) in $< 100\mu\text{s}$.
- **Real-Time Implied Volatility Solver:** High-throughput root-finding engine combining vectorized Newton-Raphson with Brent-Dekker fallback routines to extract exact implied volatilities from streaming top-of-book quotes.
- **Dynamic Arbitrage-Free Volatility Surface Fitter:** Continuous calibrator implementing the Stochastic Volatility Inspired (SVI) and SABR parametric models across delta-tenor grids, preventing calendar and butterfly arbitrage violations.
- **Sub-Millisecond Pre-Trade Risk Gateway:** High-performance gRPC gateway evaluating incremental order Greek impacts ($\Delta$-cash exposure, $\Gamma$-risk, $\text{Vega}$ limits) and enforcing user-level and broker-level risk caps in $< 400\mu\text{s}$.
- **SPAN-Style Multi-Scenario Stress Grid Generator:** Real-time risk matrix evaluator executing a 16-scenario price-volatility shift array ($\pm 3\sigma$ spot shifts, $\pm 500\text{ bps}$ IV shifts) to compute worst-case portfolio loss for margin requirements.
- **PostgreSQL & Redis State Synchronizer:** High-speed synchronization layer maintaining options contract masters, calibrated surface parameters, real-time Greek caches, and immutable audit logs of risk rejections.

## Scope Boundaries
- **In Scope:**
  - Analytical pricing for European options (Black-Scholes-Merton formula).
  - Analytical approximation pricing for American options (Bjerksund-Stensland 2002 model).
  - SIMD-vectorized calculation of Delta, Gamma, Vega, Theta, and Rho.
  - Vectorized Implied Volatility (IV) solving via Newton-Raphson and robust fallback.
  - Volatility smile and 3D surface calibration using SVI/SABR parameterizations with no-arbitrage constraints.
  - Pre-trade and intra-day user portfolio Greek exposure tracking and threshold enforcement.
  - Multi-scenario SPAN-style portfolio risk grid generation for margin evaluation.
  - High-speed gRPC API for risk evaluation and real-time Kafka Greek broadcast streaming.
- **Out of Scope / Handled Elsewhere:**
  - Physical share custody and depository integration (handled in Prompt 213).
  - Core order book matching and order queue processing (handled in Prompt 205).
  - Cash wallet ledger debit/credit and fiat banking integration (handled in Prompt 203).
  - Smart contract token minting, burning, and on-chain options settlement contracts (handled in Prompt 303 / Prompt 306).
  - Central Counterparty Settlement Guarantee Fund default waterfall execution (handled in Prompt 230).

## Technology to Use
- **Primary Computational Language:** **C++20** (GCC 13+ / Clang 17+) or **Rust 1.78+** utilizing cache-aligned memory buffers, zero-heap allocations in hot paths, and lock-free concurrency.
- **SIMD Hardware Vectorization:** Hardware intrinsics via `core::arch::x86_64` (AVX-512, AVX2, FMA) and `core::arch::aarch64` (ARM NEON) for 8-wide (double precision) and 16-wide (single precision) vectorized mathematical evaluations.
- **Inter-Service Communication:** `tonic` (Rust) or `grpc++` (C++) for sub-millisecond gRPC risk validations; Protobuf v3 schemas.
- **In-Memory Cache & State Store:** **Redis 7.2+ Cluster** with pipeline batching and in-process cache (`moka` in Rust or high-performance concurrent hash tables in C++) for instant Greek lookups.
- **Event Streaming:** **Apache Kafka** via `rdkafka` / `librdkafka` consuming streaming trade fills and L2 order book quotes, and publishing computed Greeks to `options.greeks.v1`.
- **Database & Storage:** **PostgreSQL 16+** with async connection pooling for persistent contract master records, historical surface snapshots, user risk caps, and risk breach logs.

## Backend / Infra Touchpoints
- **Redis 7.2 Keys:** `opt:surface:{underlying_isin}:svi`, `opt:greeks:{contract_isin}`, `opt:portfolio:{user_id}:greeks`, `opt:config:risk_caps`, `opt:spot:{underlying_isin}`.
- **PostgreSQL 16 Tables:** `options_contracts`, `volatility_surface_snapshots`, `user_greek_limits`, `options_portfolio_risk_snapshots`, `options_risk_breaches`.
- **Apache Kafka Topics:** Consumes `market.ticker.v1`, `matching.trades.v1`, `market.depth.l2.v1`; publishes `options.greeks.v1`, `options.surface_updates.v1`, `options.risk_alerts.v1`.
- **Order Matching Engine (Prompt 205):** Ingests fair-value and theoretical price bands to enforce dynamic options volatility circuit breakers.
- **Pre-Trade Risk Engine (Prompt 206) & Order Service (Prompt 204):** Ingests pre-order Greek evaluation results to approve or reject incoming options orders.
- **Real-Time VaR Engine (Prompt 229):** Consumes delta-equivalent equity exposures to aggregate options risk into cross-asset margin requirements.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Collateral & Custody Verification:** Validates tokenized underlying equity collateral locked in `TokenCustodyVault.sol` on Hyperledger Besu to verify 1:1 asset backing for covered call writes and cash-secured put reservations.
- **On-Chain Address Freezes & Sanctions Enforcement:** Subscribes to `ComplianceRegistry.AddressFrozen` events emitted on Hyperledger Besu under QBFT consensus, immediately updating local memory blacklists to block pre-trade options calculations for sanctioned wallets within 50ms.
- **Zero On-Chain PII Guarantee:** The engine processes only `user_id` UUIDs, contract ISINs, and Ethereum addresses (`0x...`). All personal identity records, PANs, and bank details remain isolated off-chain in compliance with DPDP Act 2023.
- **Expiry Settlement Manifest Attestation:** Upon contract expiration, the engine generates an immutable batch exercise/assignment settlement manifest with Merkle proofs, committed to `SettlementDvP.sol` on Hyperledger Besu for on-chain delivery-versus-payment finality.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project & Build Pipeline:** Scaffold `services/options-risk-engine` with optimized compiler configurations (`-O3 -march=native -flto` or Cargo `lto = "fat"`, `codegen-units = 1`, `opt-level = 3`).
2. **Define Protobuf Schema:** Create `proto/growww/options_risk/v1/options_risk_service.proto` specifying `EvaluateOptionOrderRisk`, `ComputeOptionGreeksBatch`, `GetVolatilitySurface`, `CalculatePortfolioGreekRisk`, and `UpdateSurfaceParameters`.
3. **Generate gRPC Stubs:** Compile Protobuf definitions into high-performance C++ / Rust server traits and data models.
4. **Design PostgreSQL Schema:** Write database migration scripts establishing tables for `options_contracts`, `volatility_surface_snapshots`, `user_greek_limits`, `options_portfolio_risk_snapshots`, and `options_risk_breaches`.
5. **Implement SIMD Cumulative Normal Distribution (CDF) & PDF Functions:** Implement fast vectorized approximations for the standard Gaussian cumulative distribution $\Phi(x)$ and probability density $\phi(x)$ using Rational Chebyshev or Abramowitz-Stegun approximations with AVX-512 / NEON FMA instructions.
6. **Implement Vectorized Black-Scholes-Merton (BSM) Analytical Pricing:** Write vectorized mathematical kernels computing European Call and Put theoretical prices:
   $$d_1 = \frac{\ln(S / K) + (r - q + \frac{1}{2}\sigma^2)T}{\sigma \sqrt{T}}, \quad d_2 = d_1 - \sigma \sqrt{T}$$
   $$C = S e^{-q T} \Phi(d_1) - K e^{-r T} \Phi(d_2), \quad P = K e^{-r T} \Phi(-d_2) - S e^{-q T} \Phi(-d_1)$$
7. **Implement Bjerksund-Stensland (2002) American Options Pricing:** Code the closed-form analytical approximation for American options on equities with early exercise boundaries, calculating trigger prices ($I_1, I_2$) and exercise probabilities without expensive lattice/tree traversals.
8. **Build Hardware-Vectorized Greeks Calculation Engine:** Implement analytical first-order and second-order Greek routines computing across 8 or 16 option contracts concurrently per CPU vector register:
   - Delta ($\Delta$): $\frac{\partial V}{\partial S}$
   - Gamma ($\Gamma$): $\frac{\partial^2 V}{\partial S^2} = \frac{e^{-q T} \phi(d_1)}{S \sigma \sqrt{T}}$
   - Vega ($\mathcal{V}$): $\frac{\partial V}{\partial \sigma} = S e^{-q T} \phi(d_1) \sqrt{T}$
   - Theta ($\Theta$): $\frac{\partial V}{\partial T}$
   - Rho ($\rho$): $\frac{\partial V}{\partial r}$
   - Higher-order cross Greeks: Vanna ($\frac{\partial \Delta}{\partial \sigma}$) and Volga ($\frac{\partial \mathcal{V}}{\partial \sigma}$).
9. **Implement Real-Time Vectorized Implied Volatility (IV) Root Finder:** Build root-finding engine calculating IV from streaming market prices:
   - Primary: Vectorized Newton-Raphson iteration using analytical Vega as derivative: $\sigma_{n+1} = \sigma_n - \frac{V(\sigma_n) - V_{\text{mkt}}}{\mathcal{V}(\sigma_n)}$.
   - Secondary / Fallback: Vectorized Brent-Dekker bracketed solver for deep out-of-the-money or low-vega boundary conditions where Newton-Raphson diverges.
10. **Implement Arbitrage-Free SVI / SABR Volatility Surface Fitter:** Build continuous surface calibration worker fitting the Quasi-Explicit SVI (Stochastic Volatility Inspired) model:
    $$w(k) = a + b \left( \rho (k - m) + \sqrt{(k - m)^2 + \sigma^2} \right)$$
    enforcing Lee's moment formula constraints and checking Roger Lee no-arbitrage conditions ($b(1 + |\rho|) \le 4/T$) to guarantee absence of butterfly and calendar spread arbitrage.
11. **Implement SPAN-Style Multi-Scenario Risk Grid Generator:** Build parallel 16-scenario risk matrix evaluator assessing portfolio loss across spot shifts ($\Delta S \in \{-3\sigma, -2\sigma, -1\sigma, 0, +1\sigma, +2\sigma, +3\sigma\}$) and volatility shifts ($\Delta \sigma \in \{-500\text{ bps}, 0, +500\text{ bps}\}$), identifying worst-case portfolio drawdown.
12. **Build Pre-Trade Greek Exposure & Margin Validation Gateway:** Implement `EvaluateOptionOrderRisk` RPC:
    - Ingest pending order, compute post-order portfolio Greeks ($\Delta_{\text{net}}, \Gamma_{\text{gross}}, \mathcal{V}_{\text{net}}$).
    - Compare against user-level and broker-level risk thresholds.
    - Return instantaneous APPROVE/REJECT decision in $< 400\mu\text{s}$.
13. **Implement Kafka Streaming Producers & Consumers:** Consume L2 market depth and trade ticks; stream computed Greeks every 10ms to Kafka topic `options.greeks.v1` for downstream consumption by Matching Engine and UI WebSockets.
14. **Configure Observability & Prometheus Metrics:** Instrument metrics for `options_pricing_latency_nanos`, `greeks_computation_batch_size`, `iv_solver_iterations_count`, `svi_calibration_error_rmse`, and `pre_trade_risk_rejections_total`.
15. **Write Comprehensive Mathematical & Performance Benchmarks:** Execute automated test suites verifying mathematical accuracy against standard QuantLib reference benchmarks, confirming SIMD vectorization throughput of $> 2,000,000$ option pricing evaluations per second per core.

## Interfaces / Contracts

### Protobuf Definition (`options_risk_service.proto`)
```protobuf
syntax = "proto3";

package growww.options_risk.v1;

option go_package = "growww/options_risk/v1;optionsriskv1";

service OptionsRiskService {
  rpc EvaluateOptionOrderRisk (EvaluateOptionOrderRiskRequest) returns (EvaluateOptionOrderRiskResponse);
  rpc ComputeOptionGreeksBatch (ComputeOptionGreeksBatchRequest) returns (ComputeOptionGreeksBatchResponse);
  rpc GetVolatilitySurface (GetVolatilitySurfaceRequest) returns (GetVolatilitySurfaceResponse);
  rpc CalculatePortfolioGreekRisk (CalculatePortfolioGreekRiskRequest) returns (CalculatePortfolioGreekRiskResponse);
  rpc UpdateSurfaceParameters (UpdateSurfaceParametersRequest) returns (UpdateSurfaceParametersResponse);
}

enum OptionExerciseStyle {
  OPTION_EXERCISE_STYLE_UNSPECIFIED = 0;
  OPTION_EXERCISE_STYLE_EUROPEAN = 1;
  OPTION_EXERCISE_STYLE_AMERICAN = 2;
}

enum OptionType {
  OPTION_TYPE_UNSPECIFIED = 0;
  OPTION_TYPE_CALL = 1;
  OPTION_TYPE_PUT = 2;
}

enum RiskEvaluationDecision {
  RISK_EVALUATION_DECISION_UNSPECIFIED = 0;
  RISK_EVALUATION_DECISION_APPROVED = 1;
  RISK_EVALUATION_DECISION_REJECTED = 2;
}

enum OptionRiskRejectReason {
  OPTION_REJECT_REASON_NONE = 0;
  OPTION_REJECT_REASON_DELTA_LIMIT_EXCEEDED = 1;
  OPTION_REJECT_REASON_GAMMA_LIMIT_EXCEEDED = 2;
  OPTION_REJECT_REASON_VEGA_LIMIT_EXCEEDED = 3;
  OPTION_REJECT_REASON_SPAN_MARGIN_SHORTFALL = 4;
  OPTION_REJECT_REASON_MAX_SHORT_CONTRACTS_EXCEEDED = 5;
  OPTION_REJECT_REASON_UNCOVERED_WRITING_DISALLOWED = 6;
  OPTION_REJECT_REASON_VOLATILITY_SURFACE_DISCONTINUITY = 7;
  OPTION_REJECT_REASON_UNDERLYING_HALTED = 8;
}

message OptionContractInput {
  string contract_isin = 1;
  string underlying_isin = 2;
  OptionType option_type = 3;
  OptionExerciseStyle exercise_style = 4;
  string strike_price = 5; // Decimal string INR
  string spot_price = 6; // Decimal string INR
  string risk_free_rate = 7; // Decimal string e.g. "0.0650" for 6.50%
  string dividend_yield = 8; // Decimal string e.g. "0.0120" for 1.20%
  double time_to_expiry_years = 9; // e.g. 0.08219 (30 days / 365)
  string market_price = 10; // Decimal string INR for IV solving
  string implied_volatility = 11; // Decimal string e.g. "0.1850" (if pre-known)
}

message OptionGreeksResult {
  string contract_isin = 1;
  string theoretical_price = 2;
  string implied_volatility = 3;
  string delta = 4;
  string gamma = 5;
  string vega = 6;
  string theta = 7;
  string rho = 8;
  string vanna = 9;
  string volga = 10;
  bool is_iv_converged = 11;
  int64 evaluated_at_unix_ns = 12;
}

message ComputeOptionGreeksBatchRequest {
  repeated OptionContractInput contracts = 1;
}

message ComputeOptionGreeksBatchResponse {
  repeated OptionGreeksResult results = 1;
  int64 computation_duration_nanos = 2;
}

message EvaluateOptionOrderRiskRequest {
  string request_id = 1;
  string user_id = 2;
  string ledger_address = 3;
  string contract_isin = 4;
  string underlying_isin = 5;
  string side = 6; // BUY, SELL
  string order_type = 7; // LIMIT, MARKET
  string price = 8; // Decimal string INR
  string quantity = 9; // Decimal string contracts
  OptionType option_type = 10;
  OptionExerciseStyle exercise_style = 11;
}

message EvaluateOptionOrderRiskResponse {
  string request_id = 1;
  string user_id = 2;
  RiskEvaluationDecision decision = 3;
  OptionRiskRejectReason reject_reason = 4;
  string message = 5;
  string incremental_delta_cash = 6;
  string projected_net_delta_cash = 7;
  string projected_net_gamma = 8;
  string projected_net_vega = 9;
  string required_span_margin = 10;
  string available_margin = 11;
  int64 evaluated_at_unix_ns = 12;
}

message GetVolatilitySurfaceRequest {
  string underlying_isin = 1;
}

message SviSliceParameters {
  double time_to_expiry_years = 1;
  string param_a = 2;
  string param_b = 3;
  string param_rho = 4;
  string param_m = 5;
  string param_sigma = 6;
  string rmse_fit_error = 7;
}

message GetVolatilitySurfaceResponse {
  string underlying_isin = 1;
  string underlying_spot_price = 2;
  repeated SviSliceParameters slices = 3;
  int64 calibrated_at_unix_ns = 4;
}

message CalculatePortfolioGreekRiskRequest {
  string user_id = 1;
  bool include_span_scenario_grid = 2;
}

message SpanScenarioResult {
  int32 scenario_index = 1;
  string spot_shift_percent = 2;
  string vol_shift_bps = 3;
  string projected_portfolio_pnl = 4;
}

message CalculatePortfolioGreekRiskResponse {
  string user_id = 1;
  string total_portfolio_delta_cash = 2;
  string total_portfolio_gamma = 3;
  string total_portfolio_vega = 4;
  string total_portfolio_theta = 5;
  string worst_case_scenario_loss = 6;
  string recommended_span_margin = 7;
  repeated SpanScenarioResult scenario_grid = 8;
  int64 evaluated_at_unix_ns = 9;
}

message UpdateSurfaceParametersRequest {
  string underlying_isin = 1;
  repeated SviSliceParameters slices = 2;
  string updated_by = 3;
}

message UpdateSurfaceParametersResponse {
  bool success = 1;
  int64 applied_at_unix_ns = 2;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE options_contracts (
    contract_isin VARCHAR(12) PRIMARY KEY,
    underlying_isin VARCHAR(12) NOT NULL,
    symbol VARCHAR(30) NOT NULL,
    option_type VARCHAR(4) NOT NULL CHECK (option_type IN ('CALL', 'PUT')),
    exercise_style VARCHAR(10) NOT NULL CHECK (exercise_style IN ('EUROPEAN', 'AMERICAN')),
    strike_price NUMERIC(18, 4) NOT NULL,
    lot_size INTEGER NOT NULL DEFAULT 1,
    expiration_date TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE volatility_surface_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    underlying_isin VARCHAR(12) NOT NULL,
    spot_price NUMERIC(18, 4) NOT NULL,
    time_to_expiry_years NUMERIC(8, 6) NOT NULL,
    param_a NUMERIC(12, 8) NOT NULL,
    param_b NUMERIC(12, 8) NOT NULL,
    param_rho NUMERIC(12, 8) NOT NULL,
    param_m NUMERIC(12, 8) NOT NULL,
    param_sigma NUMERIC(12, 8) NOT NULL,
    rmse_error NUMERIC(10, 8) NOT NULL,
    calibrated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_greek_limits (
    limit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    ledger_address VARCHAR(42) NOT NULL,
    tier_name VARCHAR(50) NOT NULL DEFAULT 'RETAIL_DERIVATIVES',
    max_net_delta_cash NUMERIC(18, 2) NOT NULL DEFAULT 1000000.00, -- ₹10,00,000 max Delta exposure
    max_gross_gamma NUMERIC(18, 4) NOT NULL DEFAULT 50000.00,
    max_net_vega NUMERIC(18, 2) NOT NULL DEFAULT 250000.00,
    max_short_options_contracts INTEGER NOT NULL DEFAULT 50,
    allow_uncovered_writing BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE options_portfolio_risk_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    net_delta_cash NUMERIC(18, 2) NOT NULL,
    gross_gamma NUMERIC(18, 4) NOT NULL,
    net_vega NUMERIC(18, 2) NOT NULL,
    daily_theta_decay NUMERIC(18, 2) NOT NULL,
    worst_case_span_loss NUMERIC(18, 2) NOT NULL,
    required_span_margin NUMERIC(18, 2) NOT NULL,
    effective_collateral NUMERIC(18, 2) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE options_risk_breaches (
    breach_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(64) NOT NULL,
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    contract_isin VARCHAR(12) NOT NULL,
    underlying_isin VARCHAR(12) NOT NULL,
    breach_type VARCHAR(50) NOT NULL, -- DELTA_LIMIT, GAMMA_LIMIT, VEGA_LIMIT, SPAN_MARGIN
    attempted_quantity NUMERIC(18, 4) NOT NULL,
    attempted_price NUMERIC(18, 4) NOT NULL,
    projected_metric_value NUMERIC(18, 4) NOT NULL,
    allowed_limit_value NUMERIC(18, 4) NOT NULL,
    rejection_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_options_contracts_underlying ON options_contracts(underlying_isin, expiration_date);
CREATE INDEX idx_surface_snapshots_isin ON volatility_surface_snapshots(underlying_isin, calibrated_at DESC);
CREATE INDEX idx_options_breaches_user ON options_risk_breaches(user_id, occurred_at DESC);
```

## Security & Compliance Notes
- **SEBI Derivatives Risk Management Mandate:** Enforces mandatory upfront collection of initial and exposure margins, portfolio-level SPAN risk assessments, and real-time market-wide position limit (MWPL) surveillance.
- **Fail-Closed Security Invariant:** In the event of engine computation timeout ($> 1\text{ms}$), memory exhaustion, or volatility surface discontinuity, all downstream order gateways must default to **REJECT** (`RISK_EVALUATION_DECISION_REJECTED`) to protect platform liquidity.
- **No-Arbitrage Surface Constraints:** Calibrated SVI slices are mathematically validated against Roger Lee tail bounds and butterfly arbitrage density conditions ($g(k) \ge 0$) to eliminate unphysical pricing anomalies.
- **Numerical Stability & Overflow Safeguards:** All mathematical operations are protected against divide-by-zero (e.g. $T \rightarrow 0$ or $\sigma \rightarrow 0$), negative variance roots, and extreme deep out-of-the-money float underflow using strict numerical bounding and IEEE-754 exception traps.
- **Zero In-Flight Exposure Leakage:** Position Greek mutations and margin reservations are locked atomically in memory during pre-trade authorization and released only upon trade cancellation or settlement execution.

## Acceptance Criteria
- [ ] Black-Scholes-Merton and Bjerksund-Stensland (2002) pricing models execute with mathematical accuracy matching QuantLib reference benchmarks within $< 10^{-6}$ precision.
- [ ] SIMD-vectorized Greeks pipeline computes full Greeks ($\Delta, \Gamma, \mathcal{V}, \Theta, \rho, \text{Vanna}, \text{Volga}$) for 10,000 contracts in $< 100\mu\text{s}$ on modern AVX-512 / NEON hardware.
- [ ] Vectorized Implied Volatility solver achieves $> 99.99\%$ convergence across all active strike/maturity quotes using Newton-Raphson with Brent-Dekker fallback.
- [ ] SVI volatility surface calibrates continuous, arbitrage-free curves with RMSE fitting error $\le 0.0050$ (50 bps).
- [ ] Pre-trade risk validation gateway (`EvaluateOptionOrderRisk`) responds via gRPC in $< 400\mu\text{s}$ at $p99$.
- [ ] SPAN-style 16-scenario risk grid accurately determines maximum portfolio drawdown across $\pm 3\sigma$ spot and $\pm 500\text{ bps}$ volatility shocks.
- [ ] Frozen or sanctioned on-chain addresses are blocked from placing options orders within 50ms of event emission.
- [ ] Every risk breach is immutably persisted to `options_risk_breaches` and published to Kafka topic `options.risk_alerts.v1`.
- [ ] Zero single-point-of-failure: engine handles market feed bursts of $> 500,000$ updates/second with zero dropped ticks.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `104` (Event Schema & Kafka Topic Standards), Prompt `205` (Order Matching Engine), Prompt `206` (Pre-Trade Risk Engine), Prompt `229` (Real-Time VaR Margin Engine).
- **Parallel Tasks:** Prompt `207` (Market Data Service), Prompt `226` (Advanced Order Types Engine), Prompt `228` (Real-Time Market Surveillance Engine).
- **Downstream Blockers:** Prompt `204` (Order Service options contract intake), Prompt `208` (Trade Settlement Service options DvP), Prompt `509` (Flutter Options Trading Flow).
