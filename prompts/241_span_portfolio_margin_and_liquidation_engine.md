# 241 - Real-Time SPAN Portfolio Margin & Liquidation Engine (Go / Rust / Redis)

## Purpose
In modern derivatives, multi-asset margin trading, and continuous 24/7 tokenized financial markets, traditional fixed-percentage or gross-notional margining models are capital inefficient and fail to capture complex nonlinear risk profiles. Under SEBI Master Circulars for Derivatives Risk Management, CPMI-IOSCO Principle 4 (Credit Risk), and the global Standard Portfolio Analysis of Risk (SPAN) framework, clearing corporations and prime brokers must evaluate portfolio risk holistically based on underlying price shifts, volatility shocks, inter-contract correlations, and tail-event non-linearities.

The **Real-Time SPAN Portfolio Margin & Liquidation Engine** calculates exact portfolio margin requirements at microsecond latency and executes deterministic automated liquidation cascades. By simulating 16 distinct price-volatility risk array scenarios, crediting inter-commodity and calendar spread correlations, enforcing short option minimum charges, continuously tracking intra-day maintenance margin deficits, and orchestrating multi-tiered partial liquidation auctions with zero negative balance guarantees, this engine ensures complete market solvency while maximizing capital efficiency for institutional and retail traders alike.

## What You Are Building
A mission-critical, ultra-low-latency distributed portfolio margining and automated liquidation system (`services/span-margin-engine`) built in Go and Rust, backed by Redis in-memory portfolio state and high-throughput Apache Kafka streaming. Concrete deliverables include:
- **16-Scenario SPAN Risk Array Simulation Core:** A SIMD-accelerated calculation core in Rust evaluating derivatives and underlying equity positions across 16 standardized price shift ($\pm 1/3, \pm 2/3, \pm 3/3$ price scan ranges) and volatility shock ($\pm$ volatility scan ranges, plus extreme moves scaled by loss fractions) scenarios to compute Scan Risk.
- **Inter-Commodity & Intra-Commodity Spread Credit Engine:** A delta-pairing optimizer that identifies offsetting delta risks across correlated underlyings (e.g., sector indices vs component equities, correlated derivative pairs) and contract maturities, applying calibrated spread credit rates (50% to 80%) to reduce excess margin burdens.
- **Short Option Minimum (SOM) & Delivery Risk Evaluator:** A rule engine enforcing statutory margin floors across deep out-of-the-money (OTM) short option contracts to prevent black-swan gamma explosions, layered with physical/tokenized delivery margin escalations during expiry settlement cycles.
- **Real-Time Account Health & Margin Deficit Monitor:** A streaming Go daemon evaluating continuous portfolio equity against Initial Margin (IM) and Maintenance Margin (MM) thresholds at sub-millisecond intervals upon every market tick or fill.
- **Automated Partial Liquidation Auction & Execution Bot:** A deterministic state machine executing multi-tiered liquidation cascades when account health falls below maintenance thresholds: cancelling working orders, executing calibrated partial position market auctions, slicing orders via TWAP/VWAP book engines, and handing over toxic positions to backstop liquidity pools to guarantee zero negative account balances.
- **PostgreSQL & Redis Dual-Tier Persistence Layer:** Ultra-fast Redis key-value storage for live active positions, SPAN parameter files, and health scores, coupled with PostgreSQL tables for audit trails, liquidation auctions, and regulatory compliance archives.

## Scope Boundaries
- **In Scope:**
  - Full implementation of the 16-scenario SPAN risk calculation matrix for equities, futures, and options.
  - Calculation of Scanning Risk, Intra-Commodity (calendar) Spread Charge, Inter-Commodity Spread Credit, and Short Option Minimum (SOM).
  - Continuous Mark-to-Market (MTM) margin valuation and real-time collateral haircut evaluation.
  - Account Health Ratio ($HR$) monitoring and multi-stage margin deficit alert dispatch (Warning, Margin Call, Liquidation Trigger).
  - Automated liquidation order routing: order cancellation, partial slice auctioning, TWAP/VWAP liquidation execution, and backstop takeover.
  - Zero negative account balance enforcement via Insurance Fund / Settlement Guarantee Fund backstop coordination.
- **Out of Scope / Handled Elsewhere:**
  - High-frequency matching engine central limit order book execution (handled in Prompt 205).
  - Cash wallet ledger double-entry debits and fiat banking integration (handled in Prompt 203).
  - Core Settlement Guarantee Fund (SGF) multi-tranche default waterfall orchestration (handled in Prompt 230).
  - Physical custodian depository settlement and share immobilization (handled in Prompt 213).
  - Smart contract on-chain DvP atomic token transfers (handled in Prompt 306).

## Technology to Use
- **Primary Languages & Runtime:**
  - **Rust 1.78+:** Computational core (`crates/span-math`) for SIMD-accelerated 16-scenario risk array calculations, delta matching algorithms, and matrix valuations.
  - **Go 1.22+:** Orchestration, network gateways, liquidation bots, gRPC API servers, and Kafka stream consumers (`services/span-engine`).
- **In-Memory State Store & Cache:** **Redis 7.2+ Cluster** with in-memory position state hashes, sub-millisecond pipeline queries, and pub/sub for real-time price updates.
- **Relational Database:** **PostgreSQL 16+** with `pgx/v5` and `sqlc` for persistent SPAN risk parameter files, liquidation audit logs, margin calls, and historical risk snapshots.
- **Message Broker & Event Streaming:** **Apache Kafka 3.7+** for tick-by-tick market data ingestion (`market.ticker.v1`), fill events (`matching.trades.v1`), and margin deficit triggers (`rms.span_liquidation.v1`).
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with bi-directional streaming for microsecond-level pre-order margin evaluations.

## Backend / Infra Touchpoints
- **Redis 7.2 Keys:**
  - `span:params:{underlying_id}`: Hash containing Price Scan Range (PSR), Volatility Scan Range (VSR), Intra-Spread charge rates, SOM rates, and extreme move loss fractions.
  - `span:portfolio:{user_id}:positions`: Hash of active positions by contract ID (quantity, cost basis, delta).
  - `span:collateral:{user_id}:valuation`: Effective collateral balance after haircuts.
  - `span:account:{user_id}:health`: Live Account Health Ratio ($HR$), required IM, MM, and state flag (NORMAL, WARNING, MARGIN_CALL, LIQUIDATING, LOCKED).
- **PostgreSQL 16 Tables:**
  - `span_risk_parameter_files`, `span_inter_commodity_spreads`, `span_account_margin_snapshots`, `span_margin_calls`, `span_liquidation_events`, `span_liquidation_auction_bids`.
- **Apache Kafka Topics:**
  - Consumes: `market.ticker.v1`, `matching.trades.v1`, `wallet.collateral_updated.v1`.
  - Publishes: `rms.margin_warning.v1`, `rms.margin_call.v1`, `rms.liquidation_initiated.v1`, `rms.liquidation_completed.v1`, `rms.insurance_fund_drawdown.v1`.
- **Order Service (Prompt 204) & Risk Service (Prompt 206):** Ingests pre-order margin check RPCs before order routing.
- **Advanced Order Types Engine (Prompt 226):** Coordinates automated TWAP/VWAP execution slices during liquidation cascades.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Collateral & Reserve Verification:** Interfaces with tokenized collateral and on-chain reserve registries (`SettlementGuaranteeFund.sol`, Prompt 315) on Hyperledger Besu to verify collateral haircuts and locked margin allocations.
- **On-Chain Deficit Settlement & Insurance Relaying:** If an account liquidation completely exhausts deposited collateral and incurs a deficit before auction completion, the engine logs the deficit event and invokes the on-chain Insurance Fund contract with a cryptographically attested deficit proof signed by the risk validator nodes, preventing any negative client account balance.
- **Zero On-Chain PII Guarantee:** All transactions and state proofs submitted to Hyperledger Besu reference only pseudonymous `user_id` UUID hashes, `ledger_address` (0x...), `contract_isin`, and numeric settlement quantities. Personal identifiers, bank details, and PANs remain strictly off-chain in encrypted compliance enclaves.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Monorepo Service Structure:** Initialize Go orchestration microservice `services/span-margin-engine` and embedded Rust FFI / native gRPC math library `crates/span-math` with comprehensive Makefile, benchmarks, and linting rules.
2. **Define Protobuf Specifications:** Author `proto/growww/span/v1/span_engine.proto` specifying `EvaluateSPANMargin`, `CheckPreOrderSPAN`, `SimulatePortfolioScenarios`, and `TriggerLiquidationCascade`. Generate Go and Rust stubs.
3. **Design Database Schema Migrations:** Write PostgreSQL migrations establishing tables for SPAN parameter definition files, inter-commodity spread tables, margin health snapshots, margin calls, and liquidation auction logs.
4. **Implement 16-Scenario SPAN Math Core in Rust:**
   - Model the standard 16 SPAN scenarios:
     * Scenarios 1-2: Price Unchanged, Volatility (+VSR / -VSR).
     * Scenarios 3-4: Price Up 1/3 PSR, Volatility (+VSR / -VSR).
     * Scenarios 5-6: Price Down 1/3 PSR, Volatility (+VSR / -VSR).
     * Scenarios 7-8: Price Up 2/3 PSR, Volatility (+VSR / -VSR).
     * Scenarios 9-10: Price Down 2/3 PSR, Volatility (+VSR / -VSR).
     * Scenarios 11-12: Price Up 3/3 PSR, Volatility (+VSR / -VSR).
     * Scenarios 13-14: Price Down 3/3 PSR, Volatility (+VSR / -VSR).
     * Scenario 15: Extreme Price Up (2x or 3x PSR), Volatility Unchanged (Loss multiplied by extreme move loss fraction, e.g., 35%).
     * Scenario 16: Extreme Price Down (2x or 3x PSR), Volatility Unchanged (Loss multiplied by extreme move loss fraction, e.g., 35%).
   - Calculate Scanning Risk ($SR$) as the maximum theoretical portfolio loss across all 16 scenarios.
5. **Implement Intra-Commodity (Calendar) Spread Charge:**
   - Group positions into maturity tiers (e.g., Tier 1: Current Month, Tier 2: Next Month, Tier 3: Far Month).
   - Match opposing long and short delta exposures within the same underlying asset across different tiers.
   - Apply configured calendar spread charge per matched delta pair to cover basis and term structure risk.
6. **Implement Inter-Commodity Spread Credit Engine:**
   - Identify correlated product pairs (e.g., NIFTY 50 vs BANK NIFTY, Index vs Sector ETF).
   - Compute delta equivalents across contract groups based on price ratios and correlation factors.
   - Pair qualifying deltas and apply spread credit rates (e.g., deducting 60-80% of the combined scanning risk on paired legs).
7. **Implement Short Option Minimum (SOM) & Delivery Charge Calculator:**
   - For every short call and short put position in the portfolio, compute $SOM = \sum |Q_{\text{short}}| \times SOM\_Rate$.
   - Calculate statutory Initial Margin: $\text{IM} = \max(SR + \text{IntraCharge} - \text{InterCredit}, SOM) + \text{DeliveryMargin} + \text{ExtremeLossMargin}$.
   - Compute Maintenance Margin: $\text{MM} = \text{IM} \times \text{MaintenanceFactor}$ (typically 80% of IM).
8. **Build High-Performance Redis State Store Adapter:** Implement atomic Redis Lua scripts and pipelined readers to maintain active portfolio holdings, latest mark prices, implied volatilities, and collateral valuations in memory.
9. **Build Pre-Trade Incremental SPAN Check Gateway:** Implement gRPC endpoint `CheckPreOrderSPAN` executing hypothetical pre-trade portfolio margin delta $\Delta \text{IM}$ in $< 350\mu\text{s}$ to ensure account equity remains $\ge \text{IM}_{\text{new}}$.
10. **Build Streaming Market Tick Ingestion & Real-Time Margin Monitor:** Ingest Kafka ticks from `market.ticker.v1` and fills from `matching.trades.v1`. Compute real-time Account Health Ratio:
    $$HR = \frac{\text{Effective Collateral} + \text{Unrealized MTM PnL}}{\text{Maintenance Margin}}$$
11. **Implement Multi-Stage Margin Deficit Alerting System:**
    - If $1.15 < HR \le 1.25$: Emit `MARGIN_WARNING` event (send push/email notification).
    - If $1.00 < HR \le 1.15$: Emit `MARGIN_CALL` event requiring collateral deposit within statutory intraday window.
    - If $HR \le 1.00$: Immediately transition account state to `LIQUIDATING` and dispatch execution trigger to Liquidation Bot.
12. **Implement Automated Partial Liquidation Auction Cascade:**
    - **Step 1 (Order Purge):** Call Order Service (Prompt 204) to immediately cancel all open unexecuted working limit orders for the account, freeing locked margin reserves. Recalculate $HR$; if $HR \ge 1.15$, halt liquidation.
    - **Step 2 (Partial Position Auction):** Identify the contract position generating the largest margin load. Execute partial market liquidation in calibrated tranches (e.g., 20% to 33% of position size) via microsecond batch auction or aggressive limit orders.
    - **Step 3 (TWAP/VWAP Slicing):** Route liquidation slices to the Advanced Order Types Engine (Prompt 226) to minimize market impact and order book slippage.
    - **Step 4 (Backstop Takeover):** If $HR \le 0.50$ (imminent bankruptcy or fast market dislocation), immediately transfer the residual position to the CCP Liquidation Backstop Pool / Market Maker liquidity network at the bankruptcy price ($P_{\text{bank}} = \text{EntryPrice} \pm \frac{\text{Remaining Collateral}}{\text{Position Size}}$).
13. **Implement Zero Negative Balance Invariant & Insurance Fund Drawdown:**
    - If total realized liquidation proceeds are insufficient to cover open losses (collateral balance drops to zero or below), execute an automated drawdown from the Platform Insurance Fund / Settlement Guarantee Fund (Prompt 230).
    - Credit the user ledger to restore account equity exactly to zero, guaranteeing that the trader never incurs negative equity liability or uncollectible debt.
14. **Configure Observability & Prometheus Metrics:** Instrument metrics for `span_calculation_duration_microseconds`, `span_evaluations_total`, `account_health_ratio_distribution`, `liquidation_events_total`, `liquidation_slippage_bps`, and `insurance_fund_drawdowns_inr`.
15. **Implement Comprehensive Test Suite & Benchmarks:** Write property-based simulation tests verifying SPAN mathematical accuracy against official exchange test vectors, along with stress benchmarks validating $\ge 50,000$ portfolio health evaluations per second under volatile market tick storms.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/span/v1/span_engine.proto`)
```protobuf
syntax = "proto3";

package growww.span.v1;

option go_package = "growww/span/v1;spanv1";

service SpanMarginEngineService {
  rpc EvaluateSPANMargin (EvaluateSPANMarginRequest) returns (EvaluateSPANMarginResponse);
  rpc CheckPreOrderSPAN (CheckPreOrderSPANRequest) returns (CheckPreOrderSPANResponse);
  rpc GetAccountHealth (GetAccountHealthRequest) returns (GetAccountHealthResponse);
  rpc UpdateRiskParameters (UpdateRiskParametersRequest) returns (UpdateRiskParametersResponse);
  rpc TriggerManualLiquidation (TriggerManualLiquidationRequest) returns (TriggerManualLiquidationResponse);
}

enum AccountRiskState {
  ACCOUNT_RISK_STATE_UNSPECIFIED = 0;
  ACCOUNT_RISK_STATE_HEALTHY = 1;
  ACCOUNT_RISK_STATE_WARNING = 2;
  ACCOUNT_RISK_STATE_MARGIN_CALL = 3;
  ACCOUNT_RISK_STATE_LIQUIDATING = 4;
  ACCOUNT_RISK_STATE_LOCKED = 5;
}

enum LiquidationStage {
  LIQUIDATION_STAGE_UNSPECIFIED = 0;
  LIQUIDATION_STAGE_CANCEL_OPEN_ORDERS = 1;
  LIQUIDATION_STAGE_PARTIAL_SLICE_AUCTION = 2;
  LIQUIDATION_STAGE_AGGRESSIVE_MARKET_TWAP = 3;
  LIQUIDATION_STAGE_BACKSTOP_TAKEOVER = 4;
  LIQUIDATION_STAGE_COMPLETED = 5;
}

message ContractPosition {
  string contract_id = 1;
  string underlying_id = 2;
  string isin = 3;
  string contract_type = 4; // EQUITY, FUTURE, CALL_OPTION, PUT_OPTION
  string strike_price = 5; // Decimal string (INR)
  int64 expiry_timestamp = 6;
  string side = 7; // LONG, SHORT
  string quantity = 8; // Fractional decimal string
  string average_entry_price = 9;
  string current_mark_price = 10;
  string delta = 11;
  string gamma = 12;
  string implied_volatility = 13;
}

message ScenarioLoss {
  int32 scenario_number = 1; // 1 through 16
  string price_shift_fraction = 2; // e.g. "+0.3333", "-1.0000", "+2.0000"
  string vol_shift_fraction = 3; // e.g. "+1.0000", "-1.0000", "0.0000"
  string theoretical_loss_inr = 4;
}

message EvaluateSPANMarginRequest {
  string user_id = 1;
  string ledger_address = 2;
  repeated ContractPosition positions = 3;
  string total_effective_collateral_inr = 4;
}

message EvaluateSPANMarginResponse {
  string user_id = 1;
  string scanning_risk_inr = 2;
  string intra_commodity_charge_inr = 3;
  string inter_commodity_credit_inr = 4;
  string short_option_minimum_inr = 5;
  string extreme_loss_margin_inr = 6;
  string delivery_margin_inr = 7;
  string total_initial_margin_inr = 8;
  string maintenance_margin_inr = 9;
  string unrealized_mtm_pnl_inr = 10;
  string current_equity_inr = 11;
  string health_ratio = 12;
  AccountRiskState risk_state = 13;
  repeated ScenarioLoss scenario_breakdown = 14;
  int64 calculated_at_unix_ns = 15;
}

message CheckPreOrderSPANRequest {
  string user_id = 1;
  string ledger_address = 2;
  string contract_id = 3;
  string side = 4; // BUY, SELL
  string order_type = 5; // LIMIT, MARKET
  string price = 6;
  string quantity = 7;
}

message CheckPreOrderSPANResponse {
  bool approved = 1;
  string current_initial_margin_inr = 2;
  string required_initial_margin_inr = 3;
  string post_trade_equity_inr = 4;
  string post_trade_health_ratio = 5;
  string reject_reason = 6;
}

message GetAccountHealthRequest {
  string user_id = 1;
}

message GetAccountHealthResponse {
  string user_id = 1;
  AccountRiskState risk_state = 2;
  string health_ratio = 3;
  string effective_collateral_inr = 4;
  string total_margin_required_inr = 5;
  string free_margin_inr = 6;
  string unrealized_mtm_inr = 7;
  bool is_liquidation_active = 8;
  int64 last_evaluation_time_unix_ms = 9;
}

message UpdateRiskParametersRequest {
  string underlying_id = 1;
  string price_scan_range_inr = 2;
  string vol_scan_range_percent = 3;
  string intra_spread_charge_rate = 4;
  string short_option_minimum_rate = 5;
  string extreme_move_loss_fraction = 6;
  string maintenance_margin_factor = 7;
}

message UpdateRiskParametersResponse {
  bool success = 1;
  string message = 2;
}

message TriggerManualLiquidationRequest {
  string user_id = 1;
  string reason = 2;
  string operator_id = 3;
}

message TriggerManualLiquidationResponse {
  string liquidation_id = 1;
  bool initiated = 2;
  string message = 3;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE span_risk_parameter_files (
    parameter_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    underlying_id VARCHAR(30) NOT NULL UNIQUE, -- e.g. NIFTY50, BANKNIFTY, RELIANCE
    price_scan_range_inr NUMERIC(18, 4) NOT NULL,
    vol_scan_range_percent NUMERIC(8, 4) NOT NULL, -- e.g. 0.0500 (5.00%)
    intra_spread_charge_rate NUMERIC(18, 4) NOT NULL,
    short_option_minimum_rate NUMERIC(18, 4) NOT NULL DEFAULT 50.0000, -- ₹50 per short contract
    extreme_move_loss_fraction NUMERIC(5, 4) NOT NULL DEFAULT 0.3500, -- 35%
    extreme_loss_margin_rate NUMERIC(5, 4) NOT NULL DEFAULT 0.0350, -- 3.5%
    maintenance_margin_factor NUMERIC(5, 4) NOT NULL DEFAULT 0.8000, -- 80% of IM
    version_sequence BIGINT NOT NULL DEFAULT 1,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE span_inter_commodity_spreads (
    spread_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    primary_underlying_id VARCHAR(30) NOT NULL REFERENCES span_risk_parameter_files(underlying_id),
    secondary_underlying_id VARCHAR(30) NOT NULL REFERENCES span_risk_parameter_files(underlying_id),
    primary_delta_ratio NUMERIC(10, 4) NOT NULL, -- e.g. 1.0000
    secondary_delta_ratio NUMERIC(10, 4) NOT NULL, -- e.g. 2.5000
    credit_rate_percent NUMERIC(5, 2) NOT NULL, -- e.g. 65.00%
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_spread_pair UNIQUE(primary_underlying_id, secondary_underlying_id)
);

CREATE TABLE span_account_margin_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    effective_collateral NUMERIC(18, 2) NOT NULL,
    unrealized_mtm NUMERIC(18, 2) NOT NULL,
    current_equity NUMERIC(18, 2) NOT NULL,
    scanning_risk NUMERIC(18, 2) NOT NULL,
    intra_spread_charge NUMERIC(18, 2) NOT NULL,
    inter_spread_credit NUMERIC(18, 2) NOT NULL,
    short_option_minimum NUMERIC(18, 2) NOT NULL,
    total_initial_margin NUMERIC(18, 2) NOT NULL,
    maintenance_margin NUMERIC(18, 2) NOT NULL,
    health_ratio NUMERIC(8, 4) NOT NULL,
    risk_state VARCHAR(30) NOT NULL, -- HEALTHY, WARNING, MARGIN_CALL, LIQUIDATING, LOCKED
    scenario_details JSONB NOT NULL DEFAULT '{}'::jsonb,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE span_margin_calls (
    margin_call_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    deficit_amount_inr NUMERIC(18, 2) NOT NULL,
    initial_margin_required NUMERIC(18, 2) NOT NULL,
    maintenance_margin_required NUMERIC(18, 2) NOT NULL,
    current_equity NUMERIC(18, 2) NOT NULL,
    health_ratio NUMERIC(8, 4) NOT NULL,
    deadline_timestamp TIMESTAMPTZ NOT NULL,
    status VARCHAR(30) NOT NULL, -- PENDING, RESOLVED_COLLATERAL_DEPOSIT, RESOLVED_POSITION_REDUCTION, EXPIRED_LIQUIDATED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE span_liquidation_events (
    liquidation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    trigger_health_ratio NUMERIC(8, 4) NOT NULL,
    trigger_equity NUMERIC(18, 2) NOT NULL,
    maintenance_margin_at_trigger NUMERIC(18, 2) NOT NULL,
    current_stage VARCHAR(40) NOT NULL, -- CANCEL_OPEN_ORDERS, PARTIAL_SLICE_AUCTION, TWAP_MARKET, BACKSTOP_TAKEOVER, COMPLETED
    total_positions_count INTEGER NOT NULL,
    liquidated_positions_count INTEGER NOT NULL DEFAULT 0,
    gross_realized_loss NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    insurance_fund_drawdown NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    is_zero_balance_secured BOOLEAN NOT NULL DEFAULT TRUE,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE span_liquidation_auction_bids (
    bid_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    liquidation_id UUID NOT NULL REFERENCES span_liquidation_events(liquidation_id),
    contract_id VARCHAR(50) NOT NULL,
    contract_isin VARCHAR(12) NOT NULL,
    auction_tranche_index INTEGER NOT NULL,
    offered_quantity NUMERIC(18, 6) NOT NULL,
    reserve_price NUMERIC(18, 4) NOT NULL,
    clearing_price NUMERIC(18, 4),
    buyer_participant_id VARCHAR(50),
    status VARCHAR(30) NOT NULL, -- PENDING, MATCHED, EXPIRED, CANCELLED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    matched_at TIMESTAMPTZ
);

CREATE INDEX idx_span_snapshot_user_time ON span_account_margin_snapshots(user_id, recorded_at DESC);
CREATE INDEX idx_span_margin_call_status ON span_margin_calls(status, deadline_timestamp);
CREATE INDEX idx_span_liq_events_user ON span_liquidation_events(user_id, initiated_at DESC);
CREATE INDEX idx_span_liq_events_stage ON span_liquidation_events(current_stage);
```

### Kafka Event Schemas

#### Topic: `rms.span_liquidation.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SpanLiquidationInitiatedEvent",
  "type": "object",
  "required": [
    "liquidation_id",
    "user_id",
    "ledger_address",
    "trigger_health_ratio",
    "current_equity_inr",
    "maintenance_margin_inr",
    "initiated_at_unix_ns",
    "positions"
  ],
  "properties": {
    "liquidation_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "ledger_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "trigger_health_ratio": { "type": "string" },
    "current_equity_inr": { "type": "string" },
    "maintenance_margin_inr": { "type": "string" },
    "initiated_at_unix_ns": { "type": "integer" },
    "positions": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["contract_id", "isin", "side", "quantity", "mark_price"],
        "properties": {
          "contract_id": { "type": "string" },
          "isin": { "type": "string" },
          "side": { "type": "string", "enum": ["LONG", "SHORT"] },
          "quantity": { "type": "string" },
          "mark_price": { "type": "string" },
          "margin_contribution_inr": { "type": "string" }
        }
      }
    }
  }
}
```

#### Topic: `rms.insurance_fund_drawdown.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "InsuranceFundDrawdownEvent",
  "type": "object",
  "required": [
    "drawdown_id",
    "liquidation_id",
    "user_id",
    "ledger_address",
    "drawdown_amount_inr",
    "account_equity_pre_drawdown_inr",
    "account_equity_post_drawdown_inr",
    "timestamp_unix_ns"
  ],
  "properties": {
    "drawdown_id": { "type": "string", "format": "uuid" },
    "liquidation_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "ledger_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "drawdown_amount_inr": { "type": "string" },
    "account_equity_pre_drawdown_inr": { "type": "string" },
    "account_equity_post_drawdown_inr": { "type": "string", "enum": ["0.00", "0"] },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **SEBI Derivatives Margining Compliance:** Initial Margins (SPAN Scanning Risk + Spread Charges + SOM + Extreme Loss Margin) must be collected upfront before order entry. Real-time intraday monitoring must prevent any under-margined trading activity.
- **Fail-Closed Verification Invariant:** If the SPAN Margin Engine fails to compute margin or encounters an RPC timeout ($> 3\text{ms}$), all pre-order check routers must fail-closed and reject the order with status `ORDER_REJECTED_MARGIN_SERVICE_UNAVAILABLE`.
- **Zero Negative Balance Guarantee:** Liquidation bots must prioritize partial position shedding to restore margin compliance. If a market dislocation causes uncollateralized losses, the CCP Insurance Fund backstops the deficit immediately, ensuring retail and institutional clients never face negative account balances.
- **Arbitrage-Resistant Liquidation Auctions:** Liquidation tranche mini-auctions enforce minimum tick steps and reserve pricing based on the real-time NBBO (National Best Bid/Offer) to prevent predatory front-running or malicious liquidation manipulation by market makers.
- **Arbitrary Precision Arithmetic:** All margin, equity, and scenario loss calculations are executed using 128-bit fixed-point arithmetic or Rust `rust_decimal::Decimal` (28 decimal places) to completely eliminate IEEE-754 floating-point rounding inaccuracies.

## Acceptance Criteria
- [ ] Rust SPAN calculation core executes all 16 risk array scenario simulations across 200 derivative positions in $< 200\mu\text{s}$.
- [ ] Inter-commodity spread credits correctly pair offsetting delta exposures between correlated underlyings and apply configured credit discounts (e.g., 65%).
- [ ] Short option minimum (SOM) charge is enforced on all deep out-of-the-money short call/put positions.
- [ ] Pre-order incremental margin check responds via gRPC in $< 400\mu\text{s}$ at $p99$.
- [ ] Real-time margin monitoring loop evaluates account health upon every incoming price tick and emits margin warnings at $HR \le 1.25$ and margin calls at $HR \le 1.15$.
- [ ] Account health falling below $HR \le 1.00$ triggers immediate open order cancellation and initiates stepwise partial liquidation auctions.
- [ ] Partial liquidation restores account health to $HR \ge 1.15$ without liquidating more contracts than necessary.
- [ ] Fast market gap moves causing deficit balances trigger automatic Insurance Fund backstop, ensuring account equity is restored to zero with zero negative balance.
- [ ] PostgreSQL audit tables record all SPAN snapshots, margin calls, liquidation events, and auction bids immutably.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `203` (Wallet & Account Ledger Service), Prompt `206` (Pre-Trade Risk Engine), Prompt `401` (PostgreSQL Schema), Prompt `402` (Redis Patterns & Caching).
- **Parallel Tasks:** Prompt `205` (Order Matching Engine), Prompt `207` (Market Data Service), Prompt `226` (Advanced Order Types Engine).
- **Downstream Blockers:** Prompt `204` (Order Lifecycle Service integration testing), Prompt `230` (Settlement Guarantee Fund & Default Waterfall Service).
