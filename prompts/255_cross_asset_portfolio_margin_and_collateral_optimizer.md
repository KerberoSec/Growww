# 255 - Unified Cross-Asset Portfolio Margining & Dynamic Collateral Optimization Service (Go / Rust)

## Purpose
Traditional financial market infrastructures, prime brokerages, and central counterparties (CCPs) operate in siloed margining regimes: spot equity holdings are held in depository custody under fixed statutory haircuts, derivatives margin is evaluated via isolated SPAN/VaR formulas, bullion warrants require separate warehouse margin deposits, and sovereign debt (G-Secs) remains non-fungible across trading segments. This fragmented paradigm generates severe capital inefficiencies, high cash drag, forced liquidation risk from localized margin calls despite aggregate portfolio solvency, and sub-optimal collateral allocations.

The **Unified Cross-Asset Portfolio Margining and Dynamic Collateral Optimization Service** (`services/collateral-optimizer`) solves this fragmentation by providing an ultra-low-latency, holistic cross-asset risk netting and automated collateral routing engine. Built in Rust (for high-speed SIMD matrix valuation) and Go (for distributed orchestration), the service computes cross-margining offsets across Spot Equities, Bullion Warrants, Options, Perpetual Synthetics, and G-Sec Bonds in under 100 microseconds. It incorporates inter-commodity spread offsets, covariance-based Extreme Loss Margin (ELM) reduction, Value at Risk (VaR) netting, and an automated linear programming (LP) collateral pledge optimizer that dynamically minimizes cash drag while enforcing regulatory 50:50 cash-to-collateral ratios and the platform invariant flat 0.00% transaction fee (No fee at all) (Platform Treasury, Core SGF, and Investor Protection Fund per FeeController governance).

## What You Are Building
A high-performance, real-time cross-asset portfolio risk and dynamic collateral routing microservice (`services/collateral-optimizer`) composed of:
- **Unified Cross-Asset Position & Greek Aggregator:** Aggregates real-time positions, notionals, and analytical Greeks (Delta, Gamma, Vega, Theta, Rho, DV01) across 5 asset classes: Spot Equities, Gold/Silver Bullion Warrants, Options Chains, Perpetual Synthetics, and Sovereign G-Sec Bonds.
- **SIMD-Accelerated Covariance & VaR Netting Core (Rust):** Computes multi-asset portfolio Value at Risk (VaR), Stressed Expected Shortfall (ES), and covariance-based Extreme Loss Margin (ELM) using vectorized AVX-512 / AVX2 matrix routines, achieving sub-100 microsecond evaluation times.
- **Inter-Commodity Spread & Basis Offset Engine:** Evaluates statistical correlations and delta offsets across economically linked instruments (e.g., Gold Bullion Warrants vs Gold Perpetuals, Sovereign Yield vs Interest Rate Futures, Equity Basket vs Index Options), applying calibrated margin credit reductions between 40% and 80%.
- **Dynamic Collateral Pledge Routing & LP Solver:** Implements a Mixed-Integer Linear Programming (MILP) / Simplex optimization algorithm that continuously rebalances pledged collateral assets (Cash INR, T-Bills, G-Secs, Bullion Warrants, Blue-Chip Equities) to minimize yield loss (cash drag) while satisfying exchange haircut requirements and the statutory 50:50 cash-equivalent rule.
- **Microsecond Pre-Trade Margin Delta Engine:** Provides synchronous gRPC interfaces to the Order Service (Prompt 204) and Risk Engine (Prompt 206) to evaluate marginal portfolio margin delta ($\Delta \text{Margin}$) in under 100 microseconds before order matching.
- **Collateral Rebalancing & Auto-Substitution Dispatcher:** Generates optimal collateral pledge/unpledge instructions, coordinating with the Custodian Depository Service (Prompt 213) and Wallet Service (Prompt 203).
- **Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically logs and distributes the mandatory 0.00% (Zero Fee) platform fee (0.00% fee at launch; future fee parameters governed by FeeController.sol) across all collateral rebalancing turnover and margin-funded transactions.

## Scope Boundaries
- **In Scope:**
  - Cross-asset position aggregation and real-time Greek normalization across all 5 supported asset tiers.
  - Multi-asset covariance matrix computation, parametric VaR, Stressed VaR, and ELM portfolio reduction.
  - Inter-commodity, calendar, and basis spread offset calculations with correlation threshold gating.
  - Linear programming optimization for automated collateral allocation minimizing opportunity cost and cash drag.
  - Enforcement of SEBI / CPMI-IOSCO collateral rules, including the 50:50 cash-equivalent collateral composition rule.
  - Sub-100 microsecond pre-trade incremental margin checks via gRPC.
  - Real-time collateral haircut evaluation across equity liquidity tiers, sovereign debt maturities, and bullion purities.
  - 0.00% (No fee at all) platform fee computation and 0.00% fee launch policy revenue ledger partitioning.
- **Out of Scope / Handled Elsewhere:**
  - Central Limit Order Book matching and execution (handled in Prompt 205 Order Matching Engine).
  - Default waterfall and auction liquidation execution (handled in Prompt 230 Settlement Guarantee Fund & Default Waterfall Service).
  - Physical warehouse inspection and assaying of bullion (handled in Prompt 243 MCX Commodity Adapter).
  - Primary sovereign debt issuance auctions (handled in Prompt 254 / Prompt 333).
  - Depository pledge confirmation with NSDL/CDSL APIs (handled in Prompt 213 Custodian Depository Integration Service).

## Technology to Use
- **Primary Languages & Runtimes:**
  - **Rust 1.78+ (`crates/portfolio-math`):** Computational math core utilizing SIMD intrinsics (`std::simd` / AVX-512 / AVX2) for matrix multiplication, covariance Cholesky decomposition, Greek aggregation, and ultra-fast linear solvers.
  - **Go 1.22+ (`services/collateral-optimizer`):** High-concurrency orchestration service, gRPC endpoints, Kafka event consumers, Redis state managers, and database persistence layer.
- **Linear Programming & Optimization:** `good_lp` / `minilp` (Rust) or `coin-or/clp` FFI bindings for high-speed Simplex collateral pledge optimization.
- **In-Memory Cache & Lock-Free Storage:** **Redis 7.2+ Cluster** with shared-memory cache for instant retrieval of asset price ticks, implied volatilities, and haircut matrices.
- **Event Streaming:** **Apache Kafka 3.7+** for ingesting real-time trades (`matching.trades.v1`), mark-to-market prices (`market.ticker.v1`), and collateral updates (`collateral.events.v1`).
- **Relational Persistence:** **PostgreSQL 16+** with `pgx/v5` and `sqlx` for storing historical portfolio margin snapshots, collateral allocation plans, haircut schedules, and fee distribution records.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with mTLS and Unix Domain Sockets for ultra-low latency internal RPC calls.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `portfolio_margin_accounts`, `cross_asset_positions`, `collateral_asset_inventory`, `collateral_pledge_allocations`, `haircut_matrix_schedules`, `inter_asset_correlation_matrices`, `collateral_optimization_logs`, `cross_margin_fee_ledgers`.
- **Redis Cache Keys:**
  - `margin:portfolio:{account_id}:positions`: Live hash of active cross-asset positions and Greeks.
  - `margin:collateral:{account_id}:inventory`: Available unencumbered and pledged collateral holdings.
  - `margin:account:{account_id}:summary`: Real-time required Initial Margin (IM), Maintenance Margin (MM), and effective collateral value.
  - `matrix:covariance:active`: Global active asset covariance and correlation matrix.
- **Kafka Topics:**
  - Consumes: `market.ticker.v1`, `matching.trades.v1`, `wallet.balance_updated.v1`, `custody.pledge_confirmed.v1`.
  - Publishes: `margin.recalculated.v1`, `collateral.rebalance_recommended.v1`, `collateral.pledge_action_dispatched.v1`, `margin.fee_assessed.v1`.
- **Pre-Trade Risk Engine (Prompt 206):** Ingests pre-order margin increments ($\Delta \text{Margin}$) synchronously.
- **Custodian Depository Service (Prompt 213):** Receives automated pledge/unpledge execution requests.
- **Fixed Fee & Revenue Distribution Engine (Prompt 244):** Consumes platform fee distribution logs for revenue accounting.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Collateral Registry & State Notarization:** Periodic cryptographic Merkle roots of cross-asset collateral allocations and margin valuations are notarized to `CollateralRegistry.sol` on Hyperledger Besu.
- **Atomic Collateral Tokenization:** Tokenized representation of unencumbered G-Sec bonds (`GSecToken.sol`), Bullion Warrants (`BullionWarrant.sol`), and Equity Receipts (`EquityReceipt.sol`) are escrowed on-chain as cross-asset margin backing.
- **Zero-PII Cryptographic Proofs:** All on-chain events and attestations reference only pseudonymous account hashes (`bytes32 accountHash`), token contract addresses, and 18-decimal fixed-point balances. Zero investor identity data or banking details are committed on-chain.
- **1:1 Custody Backing Guarantee:** Pledged digital tokens represent strictly 1:1 physical custody balances verified via Proof-of-Reserve oracles (Prompt 007).

## Cross-Asset Margining & Dynamic Collateral Optimization Mechanics

### 1. Cross-Asset Risk & Greek Aggregation
The engine aggregates positions across $M$ financial instruments grouped into 5 primary asset classes:
1. Spot Equities ($\text{EQ}$)
2. Bullion Warrants ($\text{BL}$)
3. Equity & Index Options ($\text{OPT}$)
4. Perpetual Synthetics ($\text{PERP}$)
5. Sovereign Debt & G-Secs ($\text{GS}$)

For an account holding position vector $\mathbf{q} = [q_1, q_2, \dots, q_M]^T$ with spot/mark prices $\mathbf{S} = [S_1, S_2, \dots, S_M]^T$:
- **Portfolio Notional:**
  $$V_{\text{port}} = \sum_{i=1}^{M} |q_i| \cdot S_i$$
- **Delta-Equivalent Exposure Vector:**
  $$\mathbf{\Delta} = [\Delta_1, \Delta_2, \dots, \Delta_M]^T, \quad \text{where } \Delta_i = q_i \cdot \frac{\partial V_i}{\partial S_i}$$
- **Gamma Vector:**
  $$\mathbf{\Gamma} = [\Gamma_1, \Gamma_2, \dots, \Gamma_M]^T, \quad \text{where } \Gamma_i = q_i \cdot \frac{\partial^2 V_i}{\partial S_i^2}$$
- **Vega Vector:**
  $$\mathbf{\mathcal{V}} = [\mathcal{V}_1, \mathcal{V}_2, \dots, \mathcal{V}_M]^T, \quad \text{where } \mathcal{V}_i = q_i \cdot \frac{\partial V_i}{\partial \sigma_i}$$

### 2. Covariance-Based Parametric VaR & Extreme Loss Margin (ELM)
Let $\mathbf{\Sigma} \in \mathbb{R}^{M \times M}$ be the exponentially weighted covariance matrix of asset price returns:
$$\sigma_{\text{port}} = \sqrt{\mathbf{\Delta}^T \mathbf{\Sigma} \mathbf{\Delta}}$$

The 99% 1-Day Value at Risk ($\text{VaR}_{0.99}$) with Cornish-Fisher expansion incorporating Gamma non-linearity:
$$\text{VaR}_{\text{port}} = Z_{0.99} \cdot \sigma_{\text{port}} - \frac{1}{2} \mathbf{\Delta}^T \text{diag}(\mathbf{\Gamma}) \mathbf{\Delta} \cdot (\Delta S_{\text{shock}})^2$$
where $Z_{0.99} = 2.3263$.

**Extreme Loss Margin (ELM) Reduction:**
Instead of gross summation across asset classes, cross-asset diversification reduces total ELM:
$$\text{ELM}_{\text{gross}} = \sum_{c \in \{\text{EQ}, \text{BL}, \text{OPT}, \text{PERP}, \text{GS}\}} \text{ELM}_c$$
$$\text{Diversification Factor } \Phi = \sqrt{\frac{\mathbf{\Delta}^T \mathbf{\Sigma} \mathbf{\Delta}}{\sum_i \Delta_i^2 \sigma_i^2}}$$
$$\text{ELM}_{\text{net}} = \max\left(\text{ELM}_{\text{gross}} \times \Phi, \quad 0.40 \times \text{ELM}_{\text{gross}}\right)$$

### 3. Inter-Commodity Spread Offset Formulation
For correlated pairs $(A, B)$ with historical correlation $\rho_{A,B} \ge 0.70$ (e.g., Gold Bullion Spot vs Gold Perpetual Future, or Nifty Index Future vs 50-Stock Replicating Equity Basket):
1. **Delta Hedge Ratio:**
   $$\beta_{A,B} = \rho_{A,B} \cdot \frac{\sigma_A}{\sigma_B}$$
2. **Matched Netting Delta:**
   $$\Delta_{\text{matched}} = \min\left(|\Delta_A|, |\beta_{A,B} \cdot \Delta_B|\right) \cdot \mathbb{I}(\text{sgn}(\Delta_A) \ne \text{sgn}(\Delta_B))$$
3. **Spread Margin Credit:**
   $$\text{Credit}_{A,B} = \Delta_{\text{matched}} \cdot \left(\text{MarginRate}_A + \text{MarginRate}_B\right) \times \Psi_{\text{spread}}(\rho_{A,B})$$
   where $\Psi_{\text{spread}}(\rho) = \min(0.80, \rho^2)$.

### 4. Dynamic Collateral Pledge Optimizer (Linear Programming)
Traders hold diverse collateral assets $k \in \mathcal{K}$ (Cash INR, T-Bills, G-Secs, Bullion Warrants, Blue-Chip Equities). Each asset $k$ has:
- Quantity available: $U_k$
- Current unit price: $P_k$
- Haircut percentage: $h_k \in [0, 1)$
- Annualized yield earned while pledged: $r_k^{\text{yield}}$
- Opportunity cost / cash drag rate: $c_k^{\text{opp}}$

The optimization problem seeks to allocate pledged amounts $x_k$ (in units) to satisfy the Total Margin Requirement ($\text{TMR} = \text{VaR}_{\text{port}} + \text{ELM}_{\text{net}} - \sum \text{Credit} + \text{SOM}$) while minimizing total opportunity cost:

$$\min_{\mathbf{x}} \sum_{k \in \mathcal{K}} \left( c_k^{\text{opp}} - r_k^{\text{yield}} \right) \cdot (x_k \cdot P_k)$$

Subject to:
1. **Total Collateral Coverage Constraint:**
   $$\sum_{k \in \mathcal{K}} (1 - h_k) \cdot (x_k \cdot P_k) \ge \text{TMR}$$
2. **50:50 Cash-Equivalent Rule Constraint (SEBI / Regulatory Mandate):**
   $$\sum_{k \in \text{CashEquivalents}} (1 - h_k) \cdot (x_k \cdot P_k) \ge 0.50 \times \text{TMR}$$
   where $\text{CashEquivalents} = \{\text{Cash INR}, \text{T-Bills}, \text{Sovereign G-Secs} \le 1\text{Y maturity}\}$.
3. **Asset Inventory Bound:**
   $$0 \le x_k \le U_k, \quad \forall k \in \mathcal{K}$$
4. **Single-Issuer Concentration Limit:**
   $$(1 - h_k) \cdot (x_k \cdot P_k) \le 0.15 \times \text{TMR}, \quad \forall k \in \text{Corporate Equities}$$

### 5. 0.00% (No fee at all) Platform Fee Model (0.00% fee at launch (future fee parameters governed by FeeController.sol))
All portfolio margin borrowing, collateral conversion turnovers, and cross-asset trades incur the standard 0.00% (Zero Fee) (0.00% fee) fee:
$$\text{Gross Turnover} = \text{Units} \times \text{Execution Price}$$
$$\text{Total Fee} = \text{Gross Turnover} \times 0.0000 = 0$$
$$\text{Platform Treasury Share} = \text{Total Fee} \times \text{treasuryRate}$$
$$\text{Core SGF Share} = \text{Total Fee} \times \text{sgfRate}$$
*(At launch, Total Fee = 0)*
$$\text{Investor Protection Fund Share (15\%)} = \text{Total Fee} \times 0.15$$

## Step-by-Step Build Instructions (10-15 steps)
1. Initialize the monorepo workspace containing Go orchestration service `services/collateral-optimizer` and Rust SIMD computational library `crates/portfolio-math`.
2. Define Protobuf contracts in `proto/growww/margin/v1/collateral_optimizer.proto` specifying valuation requests, Greek structures, collateral optimization requests, and fee allocations.
3. Generate Go and Rust RPC bindings using `buf generate` and `tonic-build`.
4. Create PostgreSQL migration scripts in `services/collateral-optimizer/migrations/001_collateral_optimizer_schema.sql` establishing tables for accounts, positions, haircuts, and optimization history.
5. Implement the Rust SIMD vector math module (`crates/portfolio-math/src/simd_var.rs`) for vectorized covariance matrix dot products and portfolio VaR computation.
6. Implement the Greek normalization engine mapping Spot, Options, Perpetuals, and G-Secs into standard continuous Delta, Gamma, and DV01 equivalents.
7. Implement the Inter-Commodity and Basis Spread Offset engine evaluating paired correlations and calculating spread margin credits.
8. Implement the Extreme Loss Margin (ELM) reduction calculator using the multi-asset diversification factor $\Phi$.
9. Implement the Mixed-Integer Linear Programming (MILP) collateral optimizer using `minilp` in Rust, encoding the objective function, haircut constraints, concentration caps, and 50:50 cash-equivalent rules.
10. Implement the Go service layer managing Redis caching for active account portfolios, collateral balances, and global covariance matrices.
11. Implement synchronous gRPC endpoints `EvaluateCrossMargin` and `OptimizeCollateralAllocation` achieving sub-100 microsecond response times.
12. Implement Kafka consumers for real-time market data ticks (`market.ticker.v1`) and trade executions (`matching.trades.v1`), triggering asynchronous margin re-evaluations.
13. Implement the automated Collateral Rebalancing Dispatcher that emits `collateral.pledge_action_dispatched.v1` events to Custody and Wallet services.
14. Implement the fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Calculator with strict double-entry ledger allocation (0.00% fee at launch; future fee parameters governed by FeeController.sol).
15. Build a comprehensive unit and stress-testing suite validating math accuracy against QuantLib benchmarks and verifying zero numerical drift across high-frequency tick bursts.

## Interfaces / Contracts

### 1. Protobuf Service Contract (`proto/growww/margin/v1/collateral_optimizer.proto`)
```protobuf
syntax = "proto3";

package growww.margin.v1;

option go_package = "github.com/growww/proto/gen/go/margin/v1;marginv1";

enum AssetClass {
  ASSET_CLASS_UNSPECIFIED = 0;
  ASSET_CLASS_SPOT_EQUITY = 1;
  ASSET_CLASS_BULLION_WARRANT = 2;
  ASSET_CLASS_EQUITY_OPTION = 3;
  ASSET_CLASS_PERPETUAL_SYNTHETIC = 4;
  ASSET_CLASS_GSEC_SOVEREIGN_DEBT = 5;
  ASSET_CLASS_CASH_FIAT = 6;
}

enum CollateralTier {
  COLLATERAL_TIER_UNSPECIFIED = 0;
  COLLATERAL_TIER_CASH_AND_EQUIVALENT = 1; // Cash INR, T-Bills, G-Secs < 1Y
  COLLATERAL_TIER_SOVEREIGN_DEBT = 2;      // G-Secs > 1Y, SDLs
  COLLATERAL_TIER_BULLION = 3;             // Gold/Silver 999 Vault Warrants
  COLLATERAL_TIER_BLUE_CHIP_EQUITY = 4;    // Group 1 Equities (Haircut <= 20%)
  COLLATERAL_TIER_MIDCAP_EQUITY = 5;       // Group 2 Equities (Haircut <= 40%)
}

message PositionItem {
  string isin = 1;
  AssetClass asset_class = 2;
  int64 quantity = 3;                     // Positive for long, negative for short
  uint64 mark_price_paise = 4;            // 10000 = INR 100.00
  int64 delta_e4 = 5;                     // Delta scaled by 10^4
  int64 gamma_e6 = 6;                     // Gamma scaled by 10^6
  int64 vega_paise = 7;                   // Vega in paise
  int64 dv01_paise = 8;                   // Interest rate DV01 in paise
}

message CollateralInventoryItem {
  string asset_id = 1;
  string isin = 2;
  AssetClass asset_class = 3;
  CollateralTier tier = 4;
  uint64 available_quantity = 5;
  uint64 already_pledged_quantity = 6;
  uint64 unit_price_paise = 7;
  uint32 haircut_bps = 8;                 // e.g. 1500 = 15.00%
  uint32 annualized_yield_bps = 9;        // e.g. 715 = 7.15% (for G-Secs/T-Bills)
  uint32 opportunity_cost_bps = 10;       // Cash drag / funding rate
}

message CrossMarginEvaluationRequest {
  string account_id = 1;
  repeated PositionItem positions = 2;
  repeated CollateralInventoryItem collateral_inventory = 3;
  bool include_optimization_plan = 4;
}

message SpreadOffsetCredit {
  string leg_a_isin = 1;
  string leg_b_isin = 2;
  uint64 matched_delta_units = 3;
  uint64 credit_amount_paise = 4;
  uint32 correlation_bps = 5;
}

message CollateralPledgeInstruction {
  string isin = 1;
  AssetClass asset_class = 2;
  uint64 pledge_quantity = 3;
  uint64 gross_value_paise = 4;
  uint64 haircut_value_paise = 5;
  uint64 effective_collateral_paise = 6;
}

message CrossMarginFeeBreakdown {
  uint64 gross_turnover_paise = 1;
  uint64 platform_fee_paise = 2;          // 0.00% (No fee at all)
  uint64 treasury_share_paise = 3;        // Governed by FeeController (0.00% at launch)
  uint64 core_sgf_share_paise = 4;        // Governed by FeeController (0.00% at launch)
  uint64 ipf_share_paise = 5;             // Governed by FeeController (0.00% at launch)
}

message CrossMarginEvaluationResponse {
  string account_id = 1;
  uint64 portfolio_notional_paise = 2;
  uint64 parametric_var_paise = 3;
  uint64 gross_elm_paise = 4;
  uint64 net_elm_paise = 5;
  uint64 total_spread_credits_paise = 6;
  uint64 initial_margin_required_paise = 7;
  uint64 maintenance_margin_required_paise = 8;
  uint64 effective_collateral_available_paise = 9;
  uint64 cash_equivalent_collateral_paise = 10;
  bool is_50_50_cash_rule_satisfied = 11;
  int64 margin_surplus_deficit_paise = 12;
  repeated SpreadOffsetCredit spread_credits = 13;
  repeated CollateralPledgeInstruction recommended_pledges = 14;
  CrossMarginFeeBreakdown fee_breakdown = 15;
  uint32 evaluation_duration_micros = 16;
  int64 evaluated_at_ns = 17;
}

message PreTradeMarginDeltaRequest {
  string account_id = 1;
  PositionItem incremental_order = 2;
}

message PreTradeMarginDeltaResponse {
  string account_id = 1;
  bool is_margin_sufficient = 2;
  int64 initial_margin_delta_paise = 3;
  uint64 new_total_margin_required_paise = 4;
  int64 remaining_margin_surplus_paise = 5;
  uint32 evaluation_duration_micros = 6;
}

service CollateralOptimizerService {
  rpc EvaluateCrossMargin (CrossMarginEvaluationRequest) returns (CrossMarginEvaluationResponse);
  rpc CheckPreTradeMarginDelta (PreTradeMarginDeltaRequest) returns (PreTradeMarginDeltaResponse);
  rpc OptimizeCollateralAllocation (CrossMarginEvaluationRequest) returns (CrossMarginEvaluationResponse);
}
```

### 2. PostgreSQL Database Schema (`services/collateral-optimizer/migrations/001_collateral_optimizer_schema.sql`)
```sql
CREATE TABLE portfolio_margin_accounts (
    account_id VARCHAR(64) PRIMARY KEY,
    user_uuid UUID NOT NULL,
    entity_id VARCHAR(32) NOT NULL DEFAULT 'DOMESTIC_NBSE',
    is_cross_margin_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    current_health_ratio NUMERIC(8, 4) NOT NULL DEFAULT 1.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE collateral_asset_inventory (
    inventory_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL REFERENCES portfolio_margin_accounts(account_id),
    isin VARCHAR(12) NOT NULL,
    asset_class VARCHAR(32) NOT NULL,
    collateral_tier VARCHAR(32) NOT NULL,
    total_unencumbered_units BIGINT NOT NULL DEFAULT 0,
    total_pledged_units BIGINT NOT NULL DEFAULT 0,
    unit_price_paise BIGINT NOT NULL,
    haircut_bps INT NOT NULL,
    annualized_yield_bps INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_account_isin UNIQUE (account_id, isin)
);

CREATE TABLE haircut_matrix_schedules (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    asset_class VARCHAR(32) NOT NULL,
    collateral_tier VARCHAR(32) NOT NULL,
    haircut_bps INT NOT NULL,
    concentration_limit_bps INT NOT NULL DEFAULT 1500, -- 15% max
    is_eligible_for_margin BOOLEAN NOT NULL DEFAULT TRUE,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE collateral_optimization_logs (
    optimization_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL REFERENCES portfolio_margin_accounts(account_id),
    portfolio_notional_paise BIGINT NOT NULL,
    initial_margin_required_paise BIGINT NOT NULL,
    effective_collateral_pledged_paise BIGINT NOT NULL,
    cash_drag_saved_bps INT NOT NULL,
    solver_iterations INT NOT NULL,
    solver_duration_micros INT NOT NULL,
    optimized_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cross_margin_fee_ledgers (
    ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL REFERENCES portfolio_margin_accounts(account_id),
    transaction_ref VARCHAR(64) NOT NULL UNIQUE,
    gross_turnover_paise BIGINT NOT NULL,
    platform_fee_paise BIGINT NOT NULL,   -- 0.00% (No fee at all)
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    besu_proof_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collateral_inventory_acc ON collateral_asset_inventory(account_id);
CREATE INDEX idx_collateral_opt_logs_acc ON collateral_optimization_logs(account_id, optimized_at DESC);
CREATE INDEX idx_cross_margin_fee_acc ON cross_margin_fee_ledgers(account_id, created_at DESC);
```

## Security & Compliance Notes
- **50:50 Cash-Equivalent SEBI Mandate:** The LP optimizer strictly enforces that at least 50% of the effective required margin is backed by Cash INR, Treasury Bills, or Sovereign G-Secs with maturity under 1 year, ensuring absolute compliance with Indian clearing corporation risk norms.
- **Microsecond Latency Isolation:** The SIMD math core in Rust executes entirely in deterministic, zero-allocation memory buffers to avoid garbage collection pauses, guaranteeing sub-100 microsecond pre-trade risk checks.
- **Concentration Risk Caps:** Single-issuer equity collateral is capped at a strict 15% concentration limit of the total required margin to prevent systemic exposure to single-stock volatility.
- **Zero-PII Data Boundaries:** All portfolio margin calculations, inventory records, and Besu ledger hashes operate strictly on pseudonymous account IDs and financial identifiers with zero investor PII.
- **Double-Entry Fee Invariant:** The 0.00% transaction fee (No fee at all) is verified using integer arithmetic: $\text{Turnover} \times \text{feeRate} = \text{Treasury} + \text{Core SGF} + \text{IPF}$ with zero truncation drift.

## Acceptance Criteria
- [ ] Protobuf schemas compile cleanly with `buf build` and `tonic-build` with zero warnings.
- [ ] Cross-asset position aggregation accurately unifies Spot, Options, Perpetuals, Bullion, and G-Sec positions into standard Greek equivalents.
- [ ] SIMD-accelerated covariance VaR and ELM reduction calculations execute in under 100 microseconds for portfolios with up to 200 distinct instruments.
- [ ] Inter-commodity spread offset engine correctly identifies correlated pairs and applies calibrated margin credits up to 80%.
- [ ] Mixed-Integer Linear Programming (MILP) solver successfully minimizes cash drag while strictly satisfying the 50:50 cash-to-collateral rule and single-issuer concentration caps.
- [ ] Pre-trade margin delta check endpoint returns accurate margin headroom in under 100 microseconds p99 latency.
- [ ] Collateral rebalancing dispatcher coordinates seamlessly with custodian depository pledge adapters.
- [ ] 0.00% (No fee at all) platform fee is assessed on all margin rebalancing turnover and divided into Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
- [ ] PostgreSQL schema and Redis caching structures pass high-concurrency integration tests without race conditions.
- [ ] Adheres strictly to the 12 mandatory sections with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `203` (Wallet & Double-Entry Account Ledger Service), Prompt `206` (Pre-Trade Risk & Margin Engine), Prompt `241` (Real-Time SPAN Portfolio Margin & Liquidation Engine).
- **Parallel Work:** Prompt `251` (Cross-Currency Collateral Dynamic FX Haircut & Auto-Hedging Engine), Prompt `254` (Bond Yield Curve & Dirty Price Calculator Service), Prompt `315` (Settlement Guarantee Fund Smart Contract).
- **Subsequent Prompts Enabled:** Prompt `256` (Limit-Up / Limit-Down LULD Volatility Dampener Service), Prompt `532` (Institutional Portfolio Margin & Collateral Management Dashboard).
