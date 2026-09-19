# 254 - Bond Yield Curve & Dirty Price Calculator Service (Rust / gRPC / SIMD)

## Purpose
Fixed-income secondary market trading requires deterministic, continuous, and sub-millisecond debt instrument valuation across a broad spectrum of sovereign paper, including Central Government Securities (G-Secs), Treasury Bills (T-Bills), State Development Loans (SDLs), and STRIPS. Unlike equities, which trade strictly on spot equity price quotations, bond trading relies on dual-pricing mechanics: Clean Price (trading quote per INR 100 face value) and Dirty Price (actual invoice cash settlement amount incorporating accrued interest). Furthermore, institutional market makers and algorithmic trading desks require real-time Yield to Maturity (YTM), Macaulay Duration, Modified Duration, and Convexity to manage interest rate duration risk and yield curve arbitrage.

The **Bond Yield Curve & Dirty Price Calculator Service** (`services/bond-pricing-engine`) is an ultra-low latency, mission-critical valuation engine built in Rust. It continuously ingests trade prints and order book updates from RBI NDS-OM, CCIL clearing feeds, and domestic exchange debt segments. The service fits real-time sovereign benchmark yield curves using the 6-parameter Nelson-Siegel-Svensson (NSS) model, computes analytical bond yield and risk metrics via vectorized SIMD routines, publishes dirty price order book ladders via high-throughput gRPC streaming, and assesses the platform invariant fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) (0.00% fee at launch; future fee parameters governed by FeeController.sol) across all bond transactions.

## What You Are Building
A high-throughput, sub-millisecond sovereign debt pricing microservice (`services/bond-pricing-engine`) written in Rust. Concrete modules include:
- **Nelson-Siegel-Svensson (NSS) Curve Calibration Engine:** Calibrates the 6-parameter continuous sovereign spot yield curve ($\beta_0, \beta_1, \beta_2, \beta_3, \tau_1, \tau_2$) in real-time from active G-Sec and T-Bill benchmark quotes using nonlinear least squares optimization with Levenberg-Marquardt and L-BFGS algorithms.
- **SIMD-Accelerated YTM & Cash Flow Solver:** Resolves bond Yield to Maturity from clean/dirty market prices using vectorized Newton-Raphson and Brent root-finding algorithms, achieving sub-microsecond convergence per instrument.
- **Accrued Interest & Day-Count Valuation Engine:** Evaluates precise accrued interest and dirty cash prices adhering to FIMMDA and RBI standards across Actual/Actual ICMA, Actual/365, and 30/360 Indian conventions.
- **Duration & Convexity Risk Calculator:** Computes Macaulay Duration, Modified Duration ($\text{DV01} / \text{PV01}$), and Convexity in real-time for dynamic portfolio interest rate risk immunization.
- **T-Bill Money Market Discount Engine:** Translates between Money Market Discount Yield ($d$), Money Market Yield (MMY), and Bond Equivalent Yield (BEY) for short-term government paper.
- **gRPC Dirty Price Streaming Bridge:** Streams live dirty price ladders and risk metrics to the Order Matching Engine (Prompt 205) and FIX Protocol Gateway (Prompt 225).
- **Universal Flat Fee Accounting:** Computes the mandatory 0.00% transaction fee (No fee at all) on total dirty turnover, logging the canonical 0.00% fee launch policy distribution.

## Scope Boundaries
- **In Scope:**
  - Continuous fitting and recalibration of the Indian sovereign zero-coupon yield curve using the Nelson-Siegel-Svensson model.
  - Multi-convention day-count accrued interest derivation (Actual/Actual ICMA, Actual/365, 30/360).
  - High-speed numerical solving for YTM from clean/dirty prices and vice versa.
  - Risk metric derivation: Macaulay Duration, Modified Duration, PV01, and Convexity.
  - Money market discount rate, MMY, and BEY calculations for 91D, 182D, and 364D T-Bills.
  - Low-latency gRPC and WebSocket streaming of bond valuations and dirty quote matrices.
  - 0.00% (No fee at all) platform fee assessment and 0.00% fee at launch (governed by FeeController.sol) accounting.
- **Out of Scope / Handled Elsewhere:**
  - Smart contract tokenization and coupon disbursement (handled in Prompt 333 Tokenized G-Sec Smart Contracts).
  - Primary NDS-OM and CCIL binary connectivity (handled in Prompt 213 Custodian Depository Integration Service).
  - Core order matching and trade execution (handled in Prompt 205 Order Matching Engine).
  - Pre-trade SPAN margin calculations and collateral haircutting (handled in Prompt 241 / Prompt 251).
  - Web and mobile UI bond trading screens (handled in Prompt 531 / Prompt 609).

## Technology to Use
- **Primary Language & Runtime:** **Rust 1.78+** utilizing `tokio` multi-threaded asynchronous runtime and `tonic` for low-latency gRPC services.
  *Justification:* Provides zero garbage collection overhead, deterministic memory layout, and compiler-level SIMD auto-vectorization for intensive numerical root-finding.
- **Numerical & Optimization Libraries:** `nalgebra` / `faer` for dense linear algebra, `argmin` for nonlinear Levenberg-Marquardt curve fitting, and `statrs` for statistical distributions.
- **Vectorized Hardware Acceleration:** SIMD instructions (`std::simd` / AVX2 / AVX-512) for parallelized multi-maturity bond cash flow discounting.
- **In-Memory Cache & Broadcast:** **Redis Cluster** with Pub/Sub for sub-millisecond distribution of calibrated NSS parameters and benchmark yields.
- **Event Streaming:** **Apache Kafka** for ingesting market trade ticks (`debt.market_data.ticks.v1`) and emitting calibrated yield curve snapshots (`debt.yield_curve.nss.v1`).
- **Relational Persistence:** **PostgreSQL 16+** with `sqlx` for historical yield curve parameter archives, bond master definitions, and fee audit trails.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `bond_master_specs`, `yield_curve_snapshots`, `nss_model_parameters`, `bond_realtime_valuations`, `debt_fee_distribution_ledgers`.
- **Kafka Topics:** Consumes `debt.market_data.ticks.v1`, `gsec.issuance_registered.v1`; publishes `debt.yield_curve.nss.v1`, `debt.valuation_tick.v1`, `debt.fee_assessed.v1`.
- **Tokenized G-Sec Contracts (Prompt 333):** Ingests on-chain bond specifications and coupon dates for model calibration.
- **Order Matching Engine (Prompt 205):** Consumes live dirty price ladders via gRPC to match clean-quoted orders with exact dirty cash settlement amounts.
- **Pre-Trade Risk Engine (Prompt 206):** Ingests PV01 and duration metrics to enforce interest rate risk limits on institutional trading desks.
- **Market Data Service (Prompt 207):** Broadcasts real-time YTM and dirty price feeds to retail mobile and web clients.
- **Fixed Fee & Revenue Distribution Engine (Prompt 244):** Consumes debt fee distribution events for statutory ledger posting.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Yield Curve Notarization:** Periodic cryptographic commitments of the calibrated sovereign NSS yield curve parameters are hashed and submitted to `OracleAggregator.sol` on Hyperledger Besu for on-chain valuation verification.
- **DvP Settlement Verification:** On-chain atomic settlements on `SettlementDvP.sol` consume signed EIP-712 dirty price attestations generated by this service to validate that cash debits match on-chain bond token transfers.
- **Zero-PII Architecture:** All pricing inputs, curve parameters, and gRPC valuation payloads reference purely financial identifiers (ISINs, yield basis points, paise amounts, timestamps) with zero investor Personally Identifiable Information.

## Nelson-Siegel-Svensson Curve & Bond Valuation Mechanics

### 1. Nelson-Siegel-Svensson (NSS) Sovereign Spot Yield Model
The NSS parametric model specifies the zero-coupon instantaneous spot yield $y(t)$ for maturity $t$ (in years) using 6 parameters: $\beta_0$ (long-term asymptotic yield), $\beta_1$ (short-term component), $\beta_2$ (medium-term curvature 1), $\beta_3$ (medium-term curvature 2), and $\tau_1, \tau_2$ (decay scale parameters):
$$y(t) = \beta_0 + \beta_1 \left(\frac{1 - e^{-t/\tau_1}}{t/\tau_1}\right) + \beta_2 \left(\frac{1 - e^{-t/\tau_1}}{t/\tau_1} - e^{-t/\tau_1}\right) + \beta_3 \left(\frac{1 - e^{-t/\tau_2}}{t/\tau_2} - e^{-t/\tau_2}\right)$$

Constraints enforced during optimization:
$$\beta_0 > 0, \quad \beta_0 + \beta_1 > 0, \quad \tau_1 > 0, \quad \tau_2 > 0$$

### 2. Clean Price to Yield to Maturity (YTM) Analytical Formulation
For a coupon bond paying semi-annual coupon $C = F \times \frac{c}{2}$ across $N$ remaining periods with fractional period $\alpha$ elapsed since last coupon:
$$\text{Dirty Price}(y) = \sum_{k=1}^{N} \frac{C}{\left(1 + \frac{y}{2}\right)^{k - 1 + \alpha}} + \frac{F}{\left(1 + \frac{y}{2}\right)^{N - 1 + \alpha}}$$
$$\text{Clean Price}(y) = \text{Dirty Price}(y) - \text{Accrued Interest}$$

The solver resolves for $y^*$ such that:
$$f(y) = \text{Clean Price}(y) - P_{\text{market}}^{\text{clean}} = 0$$

The Newton-Raphson iteration updates:
$$y_{n+1} = y_n - \frac{f(y_n)}{f'(y_n)}$$
where $f'(y)$ is the analytical derivative with respect to yield (Modified Duration scaled).

### 3. Duration, PV01, and Convexity Metrics
1. **Macaulay Duration ($D_{\text{mac}}$):** Weighted average time until cash flows are received:
   $$D_{\text{mac}} = \frac{1}{\text{Dirty Price}} \left[ \sum_{k=1}^{N} \frac{t_k \cdot C}{\left(1 + \frac{y}{m}\right)^{m \cdot t_k}} + \frac{T \cdot F}{\left(1 + \frac{y}{m}\right)^{m \cdot T}} \right]$$
2. **Modified Duration ($D_{\text{mod}}$):** Percentage price sensitivity to yield shift:
   $$D_{\text{mod}} = \frac{D_{\text{mac}}}{1 + \frac{y}{m}}$$
3. **Price Value of a Basis Point (PV01 / DV01):**
   $$\text{PV01} = \text{Dirty Price} \times D_{\text{mod}} \times 0.0001$$
4. **Convexity ($C$):** Second derivative of price with respect to yield:
   $$C = \frac{1}{\text{Dirty Price} \cdot \left(1 + \frac{y}{m}\right)^2} \left[ \sum_{k=1}^{N} \frac{t_k(t_k + 1/m) \cdot C}{\left(1 + \frac{y}{m}\right)^{m \cdot t_k}} + \frac{T(T + 1/m) \cdot F}{\left(1 + \frac{y}{m}\right)^{m \cdot T}} \right]$$
5. **Taylor Series Price Expansion:**
   $$\frac{\Delta P}{P} \approx -D_{\text{mod}} \cdot \Delta y + \frac{1}{2} C \cdot (\Delta y)^2$$

### 4. Treasury Bill (T-Bill) Money Market Yield Formulations
For a T-Bill with face value $F = 100$, price $P$, and days to maturity $t$:
1. **Discount Rate ($d$):**
   $$d = \frac{F - P}{F} \times \frac{365}{t}$$
2. **Money Market Yield (MMY):**
   $$\text{MMY} = \frac{F - P}{P} \times \frac{360}{t}$$
3. **Bond Equivalent Yield (BEY / Investment Yield):**
   $$\text{BEY} = \frac{F - P}{P} \times \frac{365}{t}$$

### 5. 0.00% (No fee at all) Platform Fee Model
The service computes the statutory 0.00% fee (No fee at all) on bond turnover:
$$\text{Gross Turnover} = \text{Order Quantity} \times \text{Dirty Price}$$
$$\text{Total Fee} = \text{Gross Turnover} \times 0.0000 = 0$$
$$\text{Platform Treasury (60\%)} = \text{Total Fee} \times 0.60$$
$$\text{Core SGF (25\%)} = \text{Total Fee} \times 0.25$$
$$\text{Investor Protection Fund - IPF (15\%)} = \text{Total Fee} \times 0.15$$

## Step-by-Step Build Instructions (10-15 steps)
1. Initialize the Rust workspace project `services/bond-pricing-engine` with high-performance compiler flags (`codegen-units = 1`, `lto = "fat"`, `target-cpu = "native"`).
2. Define Protobuf definitions in `proto/growww/debt/v1/bond_pricing_service.proto` specifying valuation requests, NSS parameters, dirty ladders, and fee distributions.
3. Generate Rust client and server stubs using `tonic-build` in `build.rs`.
4. Create PostgreSQL migration scripts in `services/bond-pricing-engine/migrations/001_bond_pricing_schema.sql` defining bond masters, yield curves, and valuation logs.
5. Implement the Day-Count calculation module supporting Actual/Actual ICMA, Actual/365, and 30/360 Indian conventions with FIMMDA leap-year handling.
6. Implement the vectorized Bond Cash Flow generator using SIMD arrays to construct coupon and principal timing vectors.
7. Implement the high-speed Newton-Raphson and Brent hybrid YTM solver with analytical derivative stepping and fallback bracket search.
8. Implement the Nelson-Siegel-Svensson (NSS) non-linear curve fitting engine using Levenberg-Marquardt optimization with parameter box constraints.
9. Implement the Risk Metrics module computing Macaulay Duration, Modified Duration, DV01, and Convexity.
10. Implement the T-Bill Money Market calculation engine handling Discount Rate, MMY, and BEY translations.
11. Implement the gRPC server implementing `CalculateBondValuation`, `CalibrateNSSYieldCurve`, `GetYieldCurveSnapshot`, and `StreamDirtyPriceLadder`.
12. Implement the Kafka consumer for live tick data (`debt.market_data.ticks.v1`) and publisher for calibrated NSS models (`debt.yield_curve.nss.v1`).
13. Implement the fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Calculator with strict 0.00% fee launch policy revenue partition ledgers.
14. Implement Redis caching with sub-millisecond TTLs for fast matching engine pre-trade lookups.
15. Write comprehensive benchmark and verification test suites validating NSS curve fitting against RBI/CCIL benchmark yield curves and FIMMDA reference pricing tables.

## Interfaces / Contracts

### 1. Protobuf Service Contract (`proto/growww/debt/v1/bond_pricing_service.proto`)
```protobuf
syntax = "proto3";

package growww.debt.v1;

option go_package = "github.com/growww/proto/gen/go/debt/v1;debtv1";

enum SovereignSecurityType {
  SOVEREIGN_SECURITY_TYPE_UNSPECIFIED = 0;
  SOVEREIGN_SECURITY_TYPE_CENTRAL_GSEC = 1;
  SOVEREIGN_SECURITY_TYPE_TREASURY_BILL = 2;
  SOVEREIGN_SECURITY_TYPE_STATE_DEVELOPMENT_LOAN = 3;
  SOVEREIGN_SECURITY_TYPE_SOVEREIGN_GREEN_BOND = 4;
  SOVEREIGN_SECURITY_TYPE_STRIPS_PRINCIPAL = 5;
  SOVEREIGN_SECURITY_TYPE_STRIPS_COUPON = 6;
}

enum DayCountConvention {
  DAY_COUNT_CONVENTION_UNSPECIFIED = 0;
  DAY_COUNT_CONVENTION_ACTUAL_ACTUAL_ICMA = 1;
  DAY_COUNT_CONVENTION_ACTUAL_365 = 2;
  DAY_COUNT_CONVENTION_THIRTY_360_ISMA = 3;
}

message BondValuationRequest {
  string isin = 1;
  uint64 clean_price_paise = 2;       // Clean price per unit (Paise, 10000 = INR 100.00)
  int64 settlement_timestamp_ns = 3;  // Target settlement timestamp
  uint64 quantity_units = 4;          // Number of bond units
}

message BondRiskMetrics {
  uint32 yield_to_maturity_bps = 1;   // YTM in basis points (e.g. 715 = 7.15%)
  uint64 macaulay_duration_days = 2;  // Macaulay duration in days
  uint32 modified_duration_bps = 3;   // Modified duration (e.g. 680 = 6.80 years)
  uint64 pv01_paise_per_unit = 4;     // Price value of 0.00% fee per unit in paise
  uint64 convexity_e4 = 5;            // Convexity scaled by 10^4
  uint32 money_market_discount_bps = 6; // For T-Bills (discount rate)
  uint32 bond_equivalent_yield_bps = 7; // For T-Bills (BEY)
}

message DebtFeeBreakdown {
  uint64 turnover_paise = 1;          // Gross trade value in paise
  uint64 total_fee_paise = 2;         // 0.00% (No fee at all) platform fee
  uint64 treasury_share_paise = 3;    // Governed by FeeController (0.00% at launch)
  uint64 core_sgf_share_paise = 4;    // Governed by FeeController (0.00% at launch)
  uint64 ipf_share_paise = 5;         // Governed by FeeController (0.00% at launch)
}

message BondValuationResponse {
  string isin = 1;
  uint64 clean_price_paise = 2;
  uint64 accrued_interest_paise = 3;
  uint64 dirty_price_paise = 4;
  uint64 gross_settlement_amount_paise = 5;
  uint32 accrued_days = 6;
  uint32 days_in_period = 7;
  BondRiskMetrics risk_metrics = 8;
  DebtFeeBreakdown fee_breakdown = 9;
  int64 calculated_at_ns = 10;
}

message NSSCalibrationParameters {
  int64 calibrated_at_ns = 1;
  double beta_0 = 2;                  // Long-term asymptotic level
  double beta_1 = 3;                  // Short-term component
  double beta_2 = 4;                  // Medium-term curvature 1
  double beta_3 = 5;                  // Medium-term curvature 2
  double tau_1 = 6;                   // Decay scale factor 1
  double tau_2 = 7;                   // Decay scale factor 2
  double root_mean_square_error_bps = 8;
  uint32 active_benchmark_count = 9;
}

message GetYieldCurveRequest {
  int64 as_of_timestamp_ns = 1;
}

message YieldCurvePoint {
  double maturity_years = 1;
  uint32 zero_coupon_yield_bps = 2;
  uint32 par_yield_bps = 3;
  uint32 forward_rate_bps = 4;
}

message GetYieldCurveResponse {
  NSSCalibrationParameters nss_parameters = 1;
  repeated YieldCurvePoint curve_points = 2;
  int64 snapshot_timestamp_ns = 3;
}

message DirtyPriceLadderRequest {
  string isin = 1;
  int64 settlement_timestamp_ns = 2;
}

message PriceLadderLevel {
  uint64 clean_price_paise = 1;
  uint64 dirty_price_paise = 2;
  uint32 ytm_bps = 3;
  uint64 quantity_units = 4;
}

message DirtyPriceLadderResponse {
  string isin = 1;
  uint64 accrued_interest_paise = 2;
  repeated PriceLadderLevel bids = 3;
  repeated PriceLadderLevel asks = 4;
  int64 generated_at_ns = 5;
}

service BondPricingService {
  rpc CalculateBondValuation (BondValuationRequest) returns (BondValuationResponse);
  rpc GetYieldCurve (GetYieldCurveRequest) returns (GetYieldCurveResponse);
  rpc StreamDirtyPriceLadder (DirtyPriceLadderRequest) returns (stream DirtyPriceLadderResponse);
}
```

### 2. PostgreSQL Database Schema (`services/bond-pricing-engine/migrations/001_bond_pricing_schema.sql`)
```sql
CREATE TABLE bond_master_specs (
    isin VARCHAR(12) PRIMARY KEY,
    security_name VARCHAR(128) NOT NULL,
    security_type VARCHAR(32) NOT NULL, -- CENTRAL_GSEC, TREASURY_BILL, STATE_DEVELOPMENT_LOAN, etc.
    face_value_paise BIGINT NOT NULL DEFAULT 10000, -- INR 100.00
    coupon_rate_bps INT NOT NULL, -- e.g. 718 for 7.18%
    coupon_frequency VARCHAR(20) NOT NULL, -- SEMI_ANNUAL, ANNUAL, ZERO_COUPON
    day_count_convention VARCHAR(32) NOT NULL DEFAULT 'ACTUAL_ACTUAL_ICMA',
    issue_date DATE NOT NULL,
    maturity_date DATE NOT NULL,
    first_coupon_date DATE,
    coupon_months INT[] NOT NULL DEFAULT '{6, 12}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE yield_curve_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    curve_type VARCHAR(32) NOT NULL DEFAULT 'INDIA_SOVEREIGN_NSS',
    beta_0 DOUBLE PRECISION NOT NULL,
    beta_1 DOUBLE PRECISION NOT NULL,
    beta_2 DOUBLE PRECISION NOT NULL,
    beta_3 DOUBLE PRECISION NOT NULL,
    tau_1 DOUBLE PRECISION NOT NULL,
    tau_2 DOUBLE PRECISION NOT NULL,
    rmse_bps DOUBLE PRECISION NOT NULL,
    benchmark_count INT NOT NULL,
    calibrated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bond_realtime_valuations (
    valuation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL REFERENCES bond_master_specs(isin),
    clean_price_paise BIGINT NOT NULL,
    accrued_interest_paise BIGINT NOT NULL,
    dirty_price_paise BIGINT NOT NULL,
    ytm_bps INT NOT NULL,
    modified_duration_bps INT NOT NULL,
    macaulay_duration_days INT NOT NULL,
    pv01_paise BIGINT NOT NULL,
    convexity_e4 BIGINT NOT NULL,
    settlement_date DATE NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE debt_fee_distribution_ledgers (
    ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_id VARCHAR(64) NOT NULL UNIQUE,
    isin VARCHAR(12) NOT NULL REFERENCES bond_master_specs(isin),
    quantity_units BIGINT NOT NULL,
    clean_price_paise BIGINT NOT NULL,
    dirty_price_paise BIGINT NOT NULL,
    gross_turnover_paise BIGINT NOT NULL,
    platform_fee_paise BIGINT NOT NULL,   -- 0.00% (No fee at all)
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    besu_tx_hash VARCHAR(66),
    settled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bond_valuations_isin_time ON bond_realtime_valuations(isin, calculated_at DESC);
CREATE INDEX idx_yield_curve_snapshots_time ON yield_curve_snapshots(calibrated_at DESC);
CREATE INDEX idx_debt_fee_isin ON debt_fee_distribution_ledgers(isin, settled_at DESC);
```

## Security & Compliance Notes
- **FIMMDA & RBI Valuation Standards:** Accrued interest and yield metrics strictly adhere to the Fixed Income Money Market and Derivatives Association of India (FIMMDA) handbook and RBI Master Directions.
- **Numerical Stability Safeguards:** The Nelson-Siegel-Svensson solver implements strict boundary clamping ($\beta_0 > 0, \tau_1 > 0, \tau_2 > 0$) and Singular Value Decomposition (SVD) fallbacks to prevent numerical divergence or singular matrix inversions during market stress.
- **Zero-PII Compliance:** The valuation engine operates entirely on financial parameters and anonymous market quotes without storing or processing investor Personally Identifiable Information.
- **Fixed Fee Invariant:** The 0.00% transaction fee (No fee at all) is verified via double-entry arithmetic checks: $\text{Turnover} \times \text{feeRate} = \text{Treasury} + \text{SGF} + \text{IPF}$ to within exact integer paise precision.
- **Deterministic SIMD Operations:** All SIMD vector discount routines maintain deterministic IEEE 754 floating-point operations across CPU architectures to ensure cross-validator reproducibility.

## Acceptance Criteria
- [ ] Protobuf schemas compile cleanly with `tonic-build` producing typed Rust stubs with zero compiler warnings.
- [ ] Nelson-Siegel-Svensson calibration completes in under 2 milliseconds across 50 active benchmark sovereign bonds with RMSE < 2.0 basis points.
- [ ] YTM solver resolves yields from clean prices in under 5 microseconds per bond with precision within 0.0.00% fees.
- [ ] Actual/Actual ICMA day-count logic precisely reproduces published FIMMDA accrued interest tables across regular and irregular coupon periods.
- [ ] Duration and Convexity calculations match reference quantitative analytics (QuantLib) within 0.00% (Zero Fee) relative tolerance.
- [ ] T-Bill money market conversion correctly maps discount rate, MMY, and BEY across 91D, 182D, and 364D tenors.
- [ ] gRPC dirty price streaming service delivers real-time order book ladders with sub-millisecond p99 latency.
- [ ] 0.00% (No fee at all) platform fee is assessed on all bond notionals and partitioned strictly into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
- [ ] High unit and integration test coverage across all numerical routines and optimization algorithms.
- [ ] Specification adheres strictly to the 12 mandatory sections with zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `102` (Service Boundary Map), Prompt `103` (API Design Standards), Prompt `205` (Order Matching Engine).
- **Parallel Work:** Prompt `333` (Tokenized G-Sec Bonds & Coupon Accrual Smart Contracts), Prompt `242` (NSE/BSE Market Data Adapter), Prompt `244` (NBSE Fixed Fee Engine).
- **Subsequent Prompts Enabled:** Prompt `531` (Flutter Sovereign Debt & G-Sec Investment Portal), Prompt `609` (G-Sec Primary Auction & Secondary Trading Web Dashboard).
