# 251 - Cross-Currency Collateral Dynamic FX Haircut & Auto-Hedging Engine (Rust / Redis / PostgreSQL)

## Purpose
International investors participating in Indian capital markets on the National Blockchain Stock Exchange (NBSE) deposit multi-currency collateral (such as USD, EUR, GBP, AED, USDT, and USDC) while trading INR-denominated equities, derivatives, and commodities. Because exchange rates fluctuate continuously during trading hours and over weekend market closures, static collateral valuation creates severe under-collateralization risks during adverse foreign exchange movements.

The **Cross-Currency Collateral Dynamic FX Haircut & Auto-Hedging Engine** provides real-time multi-currency risk management. The service ingests live composite foreign exchange feeds from RBI reference rates, interbank FX feeds, and IFSC GIFT City liquidity pools. It dynamically calculates exponentially weighted moving average (EWMA) volatility haircuts, establishes multi-tier margin buffer collars, executes automated FX spot conversions upon margin threshold triggers, and enforces the platform invariant: a flat 0.00% transaction fee (No fee at all) assessed across all multi-currency conversions and trading activities.

## What You Are Building
A ultra-low latency, deterministic multi-currency collateral valuation and auto-hedging microservice (`services/fx-haircut-engine`) written in Rust. Concrete components include:
- **Dynamic FX Volatility Haircut Calculator:** Computes rolling 30-day EWMA exchange rate volatility ($\sigma_{FX}$) and applies dynamic haircuts:
  $$H_{FX} = \max(H_{\text{base}}, k \cdot \sigma_{FX} \cdot \sqrt{\Delta t})$$
  ensuring adequate margin coverage under extreme currency depreciation scenarios.
- **Cross-Currency Margin Valuation Engine:** Revalues global user collateral portfolios in sub-milliseconds, converting foreign asset values into normalized INR margin credits.
- **Automated FX Spot Auto-Hedger:** Interfaces with IFSC GIFT City banking APIs and liquidity providers to execute automated micro-hedges when cross-currency exposures exceed predetermined risk limits.
- **Weekend / Overnight FX Buffer Scaler:** Automatically widens collateral haircuts prior to weekend trading halts to absorb global currency shocks before Monday morning market open.
- **Universal Fee Integration:** Assesses the mandatory 0.00% transaction fee (No fee at all) (0.00% fee at launch; future fee parameters governed by FeeController.sol) on all currency conversions, collateral swaps, and hedging fills.

## Scope Boundaries
- **In Scope:**
  - Real-time foreign exchange market data ingestion and triangulation.
  - Multi-currency collateral haircut computation and margin credit derivation.
  - Liquidity monitoring across IFSC GIFT City banking rails.
  - Automated micro-hedging order dispatch for cross-currency exposures.
  - Valuation integration with SPAN Portfolio Margin Engine (Prompt 241).
- **Out of Scope / Handled Elsewhere:**
  - Core SPAN margin scanning and risk checks (handled in Prompt 241).
  - On-chain token minting/burning of synthetic stablecoins (handled in Prompt 238 / Prompt 326).
  - Cross-chain bridge relayer monitoring (handled in Prompt 247).
  - Fiat bank deposits and withdrawals (handled in Prompt 203 / Prompt 214).

## Technology to Use
- **Primary Language & Framework:** **Rust 1.78+** utilizing `tokio` asynchronous runtime, `glidesort` for high-speed rate sorting, and `tonic` gRPC. Rust guarantees zero garbage collection latency and high mathematical accuracy for risk critical valuations.
- **In-Memory Cache & Pipeline:** **Redis Cluster** for caching real-time FX spot rates, historical EWMA variance matrices, and account collateral balances.
- **Message Broker:** **Apache Kafka** for ingesting tick streams (`fx.market_data.ticks.v1`) and emitting margin valuation events (`margin.fx_revalued.v1`).
- **Database & Persistence:** **PostgreSQL 16+** with `sqlx` for storing FX haircut policies, currency master records, and auto-hedge execution history.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `currency_master`, `fx_haircut_policies`, `account_multicurrency_collateral`, `fx_hedging_orders`, `fx_conversion_fee_ledgers`.
- **Kafka Topics:** Consumes `fx.market_data.ticks.v1`, `account.deposit.multicurrency.v1`; publishes `margin.fx_revalued.v1`, `hedging.order_submitted.v1`, `risk.fx_margin_call.v1`.
- **SPAN Margin Engine (Prompt 241):** Queries FX Haircut Engine via low-latency gRPC to obtain current effective INR collateral values for international accounts.
- **Risk & Margin Service (Prompt 206):** Ingests real-time margin alerts when foreign currency collateral values breach maintenance thresholds.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Proof-of-Collateral Multi-Currency Notarization:** Periodic cryptographic commitments of aggregate foreign currency reserves and applicable haircuts are posted to `ProofOfReserve.sol`.
- **On-Chain DvP Multi-Currency Settlement:** Cross-currency settlement transactions on `SettlementDvP.sol` consume notarized exchange rate oracles with signed EIP-712 price attestations from authorized FX nodes.
- **Zero PII Exposure:** On-chain records contain only pseudonymous account hashes, currency identifiers (e.g., `USD`, `EUR`), haircut percentages, and tokenized quantities.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Create Rust Cargo project `services/fx-haircut-engine` with high-performance math optimizations and strict clippy rules.
2. **Define Protobuf Schema:** Create `proto/growww/fx/v1/fx_haircut_service.proto` defining `GetEffectiveCollateralValue`, `CalculateFxHaircut`, `StreamFxRates`, and `ExecuteFxHedge`.
3. **Generate gRPC Stubs:** Build Rust gRPC client and server interfaces using `tonic-build`.
4. **Design PostgreSQL Schema:** Write migrations creating `currency_master`, `fx_haircut_policies`, and `fx_hedging_orders` tables.
5. **Implement FX Rate Ingestion Engine:** Build WebSocket and FIX adapters to ingest live exchange rates for USD/INR, EUR/INR, GBP/INR, AED/INR, and USDT/USD.
6. **Implement EWMA Volatility Engine:** Write online streaming algorithm to compute 30-day exponentially weighted moving average variance and dynamic haircut bounds ($H_{FX}$).
7. **Implement Multi-Currency Valuation Service:** Construct sub-millisecond calculation pipeline converting multi-currency wallet balances to effective INR margin credits.
8. **Implement Weekend / Overnight Haircut Multiplier:** Build automated cron logic to increase collateral haircuts by $+5\%$ at 17:00 IST on Fridays and revert at 08:30 IST on Mondays.
9. **Implement Auto-Hedging Decision Engine:** Evaluate net platform foreign exchange delta against risk thresholds and generate hedge orders for IFSC GIFT City liquidity pools.
10. **Implement Universal 0.00% fee (No fee at all) Assessment:** Compute fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) on all currency conversion notionals and distribute 0.00% fee at launch (governed by FeeController.sol).
11. **Implement Redis Pipeline Caching:** Cache effective collateral valuations in Redis with 100ms TTL to support ultra-fast matching engine pre-trade risk checks.
12. **Implement Kafka Event Publisher:** Stream real-time margin valuation adjustments to `margin.fx_revalued.v1`.
13. **Write Unit & Benchmark Tests:** Benchmark gRPC response times (< 100 microseconds) and verify mathematical accuracy of EWMA calculations against standard financial test vectors.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/fx/v1/fx_haircut_service.proto`)
```protobuf
syntax = "proto3";

package growww.fx.v1;

option go_package = "github.com/growww/proto/gen/go/fx/v1;fxv1";

message CurrencyBalance {
  string currency_code = 1; // e.g. "USD", "EUR", "USDT"
  uint64 raw_amount_e6 = 2;
}

message GetEffectiveCollateralRequest {
  string account_id = 1;
  repeated CurrencyBalance balances = 2;
  bool apply_weekend_buffer = 3;
}

message CurrencyValuationDetail {
  string currency_code = 1;
  uint64 raw_amount_e6 = 2;
  uint64 spot_rate_paise_e6 = 3; // Rate to INR in paise
  uint32 base_haircut_bps = 4;
  uint32 volatility_haircut_bps = 5;
  uint32 total_haircut_bps = 6;
  uint64 gross_inr_value_paise = 7;
  uint64 net_collateral_inr_value_paise = 8;
}

message GetEffectiveCollateralResponse {
  string account_id = 1;
  repeated CurrencyValuationDetail currency_details = 2;
  uint64 total_effective_inr_collateral_paise = 3;
  uint64 total_fee_deduction_paise = 4; // 0.00% (Zero Fee) platform fee
  int64 evaluated_at_ns = 5;
}

service FxHaircutService {
  rpc GetEffectiveCollateralValue (GetEffectiveCollateralRequest) returns (GetEffectiveCollateralResponse);
}
```

### PostgreSQL Database Schema (`services/fx-haircut-engine/migrations/001_initial_schema.sql`)
```sql
CREATE TABLE currency_master (
    currency_code VARCHAR(8) PRIMARY KEY,
    currency_name VARCHAR(64) NOT NULL,
    is_fiat BOOLEAN NOT NULL DEFAULT TRUE,
    decimal_places INT NOT NULL DEFAULT 2,
    base_haircut_percentage NUMERIC(5, 2) NOT NULL DEFAULT 5.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fx_haircut_policies (
    policy_id VARCHAR(64) PRIMARY KEY,
    currency_code VARCHAR(8) NOT NULL REFERENCES currency_master(currency_code),
    rolling_30d_volatility NUMERIC(8, 4) NOT NULL DEFAULT 0.0000,
    dynamic_haircut_percentage NUMERIC(5, 2) NOT NULL DEFAULT 5.00,
    weekend_multiplier NUMERIC(4, 2) NOT NULL DEFAULT 1.50,
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fx_hedging_orders (
    hedge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    currency_pair VARCHAR(16) NOT NULL,
    order_side VARCHAR(4) NOT NULL,
    notional_amount NUMERIC(24, 6) NOT NULL,
    executed_rate NUMERIC(18, 6) NOT NULL,
    platform_fee_inr NUMERIC(18, 4) NOT NULL, -- 0.00% (Zero Fee) flat fee
    treasury_split_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    ipf_split_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    venue VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fx_policies_currency ON fx_haircut_policies(currency_code);
CREATE INDEX idx_fx_hedging_pair ON fx_hedging_orders(currency_pair, created_at);
```

## Security & Compliance Notes
- **Sub-Paise Precision:** All currency arithmetic is performed using 128-bit fixed-point representations to eliminate floating-point precision leakage.
- **Oracle Anti-Manipulation:** FX spot rates are derived from median filtered multi-source feeds with maximum 0.50% single-tick deviation limits.
- **Strict Fee Ring-Fencing:** Every currency conversion and collateral rebalancing event logs fee distributions (0.00% fee at launch (governed by FeeController.sol)) to the immutable audit database.
- **Zero PII Exposure:** Cryptographic account hashes are used throughout gRPC payloads and database tables.

## Acceptance Criteria
- [ ] Sub-millisecond gRPC response time for `GetEffectiveCollateralValue` (< 100 microseconds p99).
- [ ] Dynamic EWMA haircut expands automatically during high-volatility currency market conditions.
- [ ] Weekend buffer increases collateral haircuts deterministically at 17:00 IST every Friday.
- [ ] 0.00% (No fee at all) platform fee is assessed accurately on all currency conversion notionals with 0.00% fee at launch (future fee parameters governed by FeeController.sol).
- [ ] Unit tests verify zero floating-point precision loss across multi-currency portfolio valuations.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 102 (Bounded Contexts), Prompt 203 (Wallets).
- **Parallel Work:** Prompt 241 (SPAN Portfolio Margin), Prompt 250 (FPI Headroom Engine), Prompt 238 (Cross-Chain Router).
- **Enables:** Global foreign investor onboarding and multi-currency collateral risk management.
