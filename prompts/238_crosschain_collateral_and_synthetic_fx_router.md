# 238 - Cross-Chain Collateral & Synthetic FX Router Service

## Purpose
International and crypto-native institutional and retail investors accessing tokenized Indian equities, sovereign debt instruments, and GIFT City financial markets often maintain collateral denominated in major decentralized digital assets (Bitcoin [BTC], Ethereum [ETH], Solana [SOL], and USD Coin [USDC]). 

The **Cross-Chain Collateral & Synthetic FX Router Service** functions as the multi-chain collateralization, synthetic margin normalization, and FX hedging gateway. It ingests multi-chain deposits across heterogeneous distributed ledgers, normalizes them into standardized synthetic margin accounts denominated in synthetic USD (`eUSD`) or synthetic INR (`eINR`), applies real-time risk-adjusted collateral haircut matrices, tracks mark-to-market prices via high-frequency decentralized oracles, and automates slippage-protected hedging and portfolio rebalancing across centralized and decentralized liquidity venues.

By bridging decentralized digital assets with permissioned settlement ledgers and domestic trading margin accounts, this service enables capital-efficient global market participation while completely isolating the domestic equity exchange from cryptocurrency market volatility and liquidation shortfalls.

## What You Are Building
A specialized, resilient Python and Go microservice (`services/crosschain-fx-router`) comprising the following core architectural subsystems:
- **Multi-Chain Ingestion & Custodial Gateway:** Ingests and validates multi-chain transactions across Bitcoin (UTXO), Ethereum/EVM (ERC-20 and native), and Solana (SPL) rails connected to institutional Multi-Party Computation (MPC) custody vaults.
- **Collateral Haircut & Valuation Engine:** Enforces a real-time risk haircut matrix:
  - Bitcoin (BTC): 20% haircut (80% collateral loan-to-value [LTV]).
  - Ethereum (ETH): 25% haircut (75% collateral LTV).
  - Solana (SOL): 30% haircut (70% collateral LTV).
  - USD Coin (USDC): 0% haircut (100% collateral LTV).
- **Dual-Oracle Aggregation & Mark-Pricing Subsystem:** Ingests and aggregates sub-second price feeds from Chainlink Data Feeds and Pyth Network Hermes WebSocket oracles, enforcing confidence interval validation, staleness filters, and circuit breaker deviation controls.
- **Synthetic FX Normalization & Margin Allocator:** Converts risk-adjusted collateral values into normalized synthetic margin units (`eUSD` and `eINR`) using institutional spot FX feeds and Chainlink FX reference rates.
- **Automated Slippage-Protected Hedging Engine:** Executes delta-neutral spot and derivative hedges across liquid Centralized Exchanges (CEXs) and Decentralized Exchanges (DEXs) with strict maximum allowable slippage bounds (default 15 bps, hard ceiling 25 bps).
- **Automated Rebalancing & Margin Call Supervisor:** Continuously tracks collateral health ratios ($H_r$), issuing tiered margin calls at 115% threshold and triggering automated fractional liquidation or rebalancing at 105% threshold.

## Scope Boundaries
- **In Scope:**
  - Multi-chain deposit ingestion, confirmation tracking, and custodial indexing for BTC, ETH, SOL, and USDC.
  - Real-time mark-to-market collateral valuation utilizing combined Chainlink and Pyth oracle streams.
  - Calculation and enforcement of baseline and dynamic volatility-adjusted haircut matrices (BTC 20%, ETH 25%, SOL 30%, USDC 0%).
  - Synthetic margin routing, conversion, and allocation into normalized `eUSD` and `eINR` ledger accounts.
  - Automated delta-hedging execution across DEX/CEX venues with parameterized slippage protection.
  - Continuous margin health ratio monitoring, margin call generation, and automated tiered liquidation execution.
  - Publishing standardized event streams across Kafka topics for wallet, risk, and ledger updates.
- **Out of Scope / Handled Elsewhere:**
  - Direct matching engine order matching and execution (handled in Prompt 205).
  - Pre-trade equity Value-at-Risk (VaR) and Extreme Loss Margin (ELM) calculations on Indian equity portfolios (handled in Prompt 229).
  - Regulated domestic INR payment rails such as UPI, IMPS, NEFT, and RTGS (handled in Prompt 212).
  - Offshore fiat wire deposits in foreign currencies via SWIFT MT103 or ISO 20022 `pacs.008` (handled in Prompt 214).
  - Settlement Guarantee Fund (SGF) default waterfall handling (handled in Prompt 230).
  - Solidity smart contract logic for permissioned equity token issuance (handled in Prompt 303).

## Technology to Use
- **Primary Service Frameworks:**
  - **Go 1.22+:** Core daemon for multi-chain RPC ingestion, WebSocket oracle connections (Pyth Hermes / Chainlink), real-time margin valuation loops, and high-concurrency gRPC interfaces.
  - **Python 3.12+ (FastAPI):** Quantitative hedging algorithms, smart routing optimization, DEX aggregator interfacing, and statistical rebalancing calculations.
- **Blockchain Ingestion & RPC Clients:**
  - `btcsuite/btcd` for Bitcoin wire protocol parsing and UTXO transaction verification.
  - `ethereum/go-ethereum` for EVM JSON-RPC interaction, ERC-20 log parsing, and receipt validation.
  - `gagliardetto/solana-go` for Solana RPC, SPL token program parsing, and transaction confirmation.
- **Oracle Integrations:**
  - Pyth Network Go SDK and Hermes WebSocket streaming client for high-frequency pricing and confidence interval validation.
  - Chainlink Data Feeds via EVM RPC clients for reference index cross-validation.
- **Transactional & In-Memory Storage:**
  - **PostgreSQL 16+:** Relational storage with `pgx/v5` for collateral deposit records, haircut parameters, historical valuations, synthetic accounts, and hedging trade journals.
  - **Redis 7.2+:** In-memory sub-millisecond price cache, distributed locking (Redlock), and real-time margin health scores.
- **Event Streaming:**
  - **Apache Kafka 3.7+:** Event distribution for collateral confirmations, synthetic margin updates, hedging execution telemetry, margin calls, and liquidations.
- **Observability & Metrics:**
  - OpenTelemetry and Prometheus for tracking oracle latency, haircut ratios, hedging execution slippage, and collateral health metrics.

## Backend / Infra Touchpoints
- **Upstream Services & Systems:**
  - Institutional MPC Custody (Fireblocks / Copper) webhook relays and deposit address managers.
  - Pyth Network Hermes Oracle WebSocket service and Chainlink Data Feed contracts.
  - CEX Liquidity Providers (Binance, OKX, Coinbase Prime) via FIX/REST and DEX aggregators via smart routing.
  - `services/foreign-investor-funding-service` (Prompt 214): International investor account verification and GIFT City compliance validation.
- **Downstream Services & Systems:**
  - `services/wallet-account-service` (Prompt 203): Crediting and debiting normalized `eUSD` and `eINR` trading balances.
  - `services/risk-and-margin-checks-service` (Prompt 206) & `services/real-time-var-margin-engine` (Prompt 229): Ingestion of live collateral margin values for pre-trade risk checks.
  - `services/trade-settlement-service` (Prompt 208): Settlement allocation for synthetic margin accounts.
  - `services/notification-service` (Prompt 211): Transmission of margin call warnings and liquidation notifications via push, SMS, and email.
  - `services/admin-back-office` (Prompt 217 & 605): Administrative risk console for collateral parameters and emergency overrides.
  - `services/audit-log-service` (Prompt 218): Append-only cryptographic audit logging of all deposit, valuation, hedging, and liquidation operations.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consortium Ledger:** Hyperledger Besu permissioned blockchain network operating under QBFT consensus.
- **Synthetic Margin Ledger Sync:** When multi-chain collateral is confirmed and locked within institutional MPC custody vaults, the service records the deposit and transmits an attestation to the Hyperledger Besu relayer. The relayer invokes `SyntheticMarginRegistry.sol` to record the corresponding synthetic margin units allocated to the investor's cryptographic address.
- **1:1 Custodial Backing & Proof of Reserve Integration:** Every synthetic `eUSD` or `eINR` unit created on the trading platform is backed 1:1 by segregated, verifiable crypto collateral held in MPC cold/warm storage. The service regularly computes aggregate reserve proofs and publishes them to `services/reconciliation-service` (Prompt 215) and `ProofOfReserveRegistry.sol` (Prompt 308).
- **Zero PII Mandate:** All transactions, events, and state records written to the Hyperledger Besu ledger use anonymized cryptographic public keys (`address` / `bytes32`), asset symbols, token amounts, haircut ratios, and normalized margin values. No investor names, tax identification numbers, passport numbers, or residential addresses are ever stored on-chain.

## Step-by-Step Build Instructions (10-15 steps)
1. Initialize the `services/crosschain-fx-router` repository layout with modular Go packages (`cmd/`, `internal/ingestor/`, `internal/oracle/`, `internal/margin/`, `internal/hedger/`) and Python quantitative worker packages (`hedging_optimizer/`, `dex_router/`).
2. Define Protobuf contracts in `proto/growww/crosschain/v1/crosschain_router.proto` and generate Go and Python gRPC client/server stubs.
3. Construct PostgreSQL migration scripts defining schemas for collateral deposits, haircut matrices, oracle price logs, synthetic margin accounts, hedging executions, and liquidation events.
4. Build the **Multi-Chain Deposit Ingestion Engine**:
   - Implement blockchain RPC listeners and webhook parsers for Bitcoin (UTXO), Ethereum/Arbitrum (EVM/ERC-20), and Solana (SPL).
   - Enforce confirmation depth requirements: Bitcoin (3 confirmations), Ethereum/Arbitrum (finalized / 32 epochs), Solana (finalized / 32 slots).
   - Validate cryptographic deposit signatures and verify custody settlement with MPC APIs.
5. Implement the **Dual-Oracle Aggregator (Pyth & Chainlink)**:
   - Establish continuous WebSocket streaming with Pyth Network Hermes endpoint for sub-second mark pricing and confidence interval bands ($\sigma$).
   - Configure Chainlink Data Feeds via JSON-RPC polling as the secondary benchmark.
   - Enforce staleness checks ($\Delta t \le 30\text{ seconds}$), confidence interval limits ($\sigma \le 0.50\%$), and oracle divergence limits ($|\text{Price}_{\text{Pyth}} - \text{Price}_{\text{Chainlink}}| / \text{Price}_{\text{Chainlink}} \le 1.00\%$).
6. Build the **Dynamic Haircut Calculation Module**:
   - Implement baseline haircut rules: BTC (20% haircut, 80% LTV), ETH (25% haircut, 75% LTV), SOL (30% haircut, 70% LTV), and USDC (0% haircut, 100% LTV).
   - Add dynamic volatility surcharges based on 30-day realized volatility metrics ($HV_{30}$) when volatility exceeds the 80th historical percentile.
7. Implement the **Synthetic FX Normalization Engine**:
   - Compute aggregate effective collateral value in USD:
     $$V_{\text{USD}} = \sum_{i \in \{\text{BTC}, \text{ETH}, \text{SOL}, \text{USDC}\}} Q_i \times P_i \times (1 - H_i)$$
   - Apply real-time USD/INR exchange rate ($R_{\text{FX}}$) to mint and allocate normalized synthetic margin units (`eUSD` or `eINR`) to the investor margin account.
8. Implement the **Automated Slippage-Protected Hedging Engine**:
   - Calculate delta exposure on non-stable collateral (BTC, ETH, SOL) relative to platform risk limits.
   - Route spot or perpetual hedging orders across connected CEX and DEX venues.
   - Enforce strict slippage tolerance checks: reject or split orders if estimated execution slippage exceeds 15 bps (hard ceiling 25 bps).
9. Implement the **Margin Health & Liquidation Monitor**:
   - Continuously compute the margin health ratio:
     $$H_r = \frac{\text{Effective Collateral Value}}{\text{Maintenance Margin Requirement}} \times 100\%$$
   - Emit warning alert events when $H_r \le 115\%$.
   - Trigger automated fractional liquidation and collateral rebalancing when $H_r \le 105\%$.
10. Integrate the **Hyperledger Besu On-Chain State Relayer**:
    - Submit signed cryptographic state attestations to `SyntheticMarginRegistry.sol` for confirmed collateral deposits and margin adjustments.
11. Build the **Kafka Messaging Pipeline**:
    - Publish structured events for deposit confirmations, synthetic margin allocations, hedging executions, margin call warnings, and liquidations.
12. Implement **Audit Logging and Cryptographic Journaling**:
    - Stream structured audit events to `services/audit-log-service` for every state transition, valuation update, and hedging trade.
13. Configure **Prometheus Metrics and Distributed Tracing**:
    - Instrument OpenTelemetry spans and export Prometheus metrics for oracle latency, haircut utilization, hedging slippage, and health ratio distributions.
14. Build **Comprehensive Unit, Integration, and Simulation Suites**:
    - Implement automated integration tests validating multi-chain deposit ingestion, dual-oracle failover, haircut calculation accuracy, slippage bounds enforcement, and tiered liquidation execution.

## Interfaces / Contracts

### Protobuf Definition (`crosschain_router.proto`)
```protobuf
syntax = "proto3";

package growww.crosschain.v1;

option go_package = "github.com/growww/services/crosschain-fx-router/gen/v1;crosschainv1";

service CrossChainCollateralRouterService {
  rpc RegisterCollateralDeposit (RegisterDepositRequest) returns (RegisterDepositResponse);
  rpc GetCollateralValuation (CollateralValuationRequest) returns (CollateralValuationResponse);
  rpc CalculateSyntheticMargin (SyntheticMarginRequest) returns (SyntheticMarginResponse);
  rpc ExecuteHedgingOrder (HedgingOrderRequest) returns (HedgingOrderResponse);
  rpc GetMarginHealth (MarginHealthRequest) returns (MarginHealthResponse);
  rpc TriggerCollateralRebalance (RebalanceRequest) returns (RebalanceResponse);
}

enum CollateralAsset {
  COLLATERAL_ASSET_UNSPECIFIED = 0;
  COLLATERAL_ASSET_BTC = 1;
  COLLATERAL_ASSET_ETH = 2;
  COLLATERAL_ASSET_SOL = 3;
  COLLATERAL_ASSET_USDC = 4;
}

enum TargetMarginCurrency {
  MARGIN_CURRENCY_UNSPECIFIED = 0;
  MARGIN_CURRENCY_EUSD = 1;
  MARGIN_CURRENCY_EINR = 2;
}

enum BlockchainNetwork {
  NETWORK_UNSPECIFIED = 0;
  NETWORK_BITCOIN_MAINNET = 1;
  NETWORK_ETHEREUM_MAINNET = 2;
  NETWORK_SOLANA_MAINNET = 3;
  NETWORK_ARBITRUM_ONE = 4;
}

enum DepositStatus {
  DEPOSIT_STATUS_UNSPECIFIED = 0;
  DEPOSIT_STATUS_DETECTED = 1;
  DEPOSIT_STATUS_CONFIRMING = 2;
  DEPOSIT_STATUS_CONFIRMED = 3;
  DEPOSIT_STATUS_REJECTED = 4;
}

enum HedgingVenueType {
  VENUE_TYPE_UNSPECIFIED = 0;
  VENUE_TYPE_CEX_SPOT = 1;
  VENUE_TYPE_CEX_PERPETUAL = 2;
  VENUE_TYPE_DEX_SWAP = 3;
}

message RegisterDepositRequest {
  string investor_account_id = 1;
  BlockchainNetwork network = 2;
  CollateralAsset asset = 3;
  string tx_hash = 4;
  string deposit_address = 5;
  string amount_raw = 6;
  uint32 confirmation_count = 7;
}

message RegisterDepositResponse {
  string deposit_id = 1;
  DepositStatus status = 2;
  string haircut_percentage = 3; // e.g., "20.00"
  string effective_collateral_usd = 4;
  int64 registered_at_utc = 5;
}

message CollateralValuationRequest {
  string investor_account_id = 1;
}

message CollateralPosition {
  CollateralAsset asset = 1;
  string quantity = 2;
  string mark_price_usd = 3;
  string oracle_source = 4;
  string haircut_percentage = 5;
  string gross_value_usd = 6;
  string net_collateral_value_usd = 7;
}

message CollateralValuationResponse {
  string investor_account_id = 1;
  repeated CollateralPosition positions = 2;
  string total_gross_value_usd = 3;
  string total_net_collateral_usd = 4;
  string synthetic_margin_e_usd = 5;
  string synthetic_margin_e_inr = 6;
  string fx_rate_usd_inr = 7;
  int64 valuation_timestamp_utc = 8;
}

message SyntheticMarginRequest {
  string investor_account_id = 1;
  TargetMarginCurrency target_currency = 2;
}

message SyntheticMarginResponse {
  string investor_account_id = 1;
  TargetMarginCurrency target_currency = 2;
  string available_margin = 3;
  string used_margin = 4;
  string free_margin = 5;
  string health_ratio_percentage = 6;
}

message HedgingOrderRequest {
  string investor_account_id = 1;
  CollateralAsset asset = 2;
  string hedge_quantity = 3;
  HedgingVenueType preferred_venue = 4;
  uint32 max_slippage_bps = 5; // e.g., 15 bps = 0.15%
}

message HedgingOrderResponse {
  string hedge_order_id = 1;
  string venue_name = 2;
  string executed_price_usd = 3;
  string executed_quantity = 4;
  uint32 actual_slippage_bps = 5;
  string status = 6; // FILLED / PARTIALLY_FILLED / REJECTED
  int64 executed_at_utc = 7;
}

message MarginHealthRequest {
  string investor_account_id = 1;
}

message MarginHealthResponse {
  string investor_account_id = 1;
  string health_ratio_percentage = 2;
  string maintenance_margin_required_usd = 3;
  string effective_collateral_value_usd = 4;
  bool is_margin_call_active = 5;
  bool is_liquidation_eligible = 6;
}

message RebalanceRequest {
  string investor_account_id = 1;
  string reason = 2;
}

message RebalanceResponse {
  string rebalance_id = 1;
  string status = 2;
  string liquidated_collateral_usd = 3;
  string restored_health_ratio = 4;
  int64 processed_at_utc = 5;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Cross-Chain Collateral Deposits Table
CREATE TABLE crosschain_collateral_deposits (
    deposit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_account_id UUID NOT NULL,
    network VARCHAR(32) NOT NULL CHECK (network IN ('BITCOIN_MAINNET', 'ETHEREUM_MAINNET', 'SOLANA_MAINNET', 'ARBITRUM_ONE')),
    asset_symbol VARCHAR(16) NOT NULL CHECK (asset_symbol IN ('BTC', 'ETH', 'SOL', 'USDC')),
    tx_hash VARCHAR(128) NOT NULL UNIQUE,
    deposit_address VARCHAR(128) NOT NULL,
    amount_raw NUMERIC(36, 18) NOT NULL,
    confirmation_count INT NOT NULL DEFAULT 0,
    required_confirmations INT NOT NULL,
    deposit_status VARCHAR(32) NOT NULL DEFAULT 'DETECTED' CHECK (deposit_status IN ('DETECTED', 'CONFIRMING', 'CONFIRMED', 'REJECTED')),
    haircut_percentage NUMERIC(5, 2) NOT NULL,
    mark_price_usd NUMERIC(18, 6) NOT NULL,
    gross_value_usd NUMERIC(18, 2) NOT NULL,
    effective_collateral_usd NUMERIC(18, 2) NOT NULL,
    custody_vault_id VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);

-- Real-Time Collateral Haircut Matrix Configuration
CREATE TABLE collateral_haircut_matrix (
    asset_symbol VARCHAR(16) PRIMARY KEY CHECK (asset_symbol IN ('BTC', 'ETH', 'SOL', 'USDC')),
    base_haircut_percentage NUMERIC(5, 2) NOT NULL,
    max_ltv_percentage NUMERIC(5, 2) NOT NULL,
    volatility_add_on_percentage NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    liquidation_penalty_percentage NUMERIC(5, 2) NOT NULL DEFAULT 5.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Collateral Haircuts
INSERT INTO collateral_haircut_matrix (asset_symbol, base_haircut_percentage, max_ltv_percentage, volatility_add_on_percentage, liquidation_penalty_percentage)
VALUES 
    ('BTC', 20.00, 80.00, 0.00, 5.00),
    ('ETH', 25.00, 75.00, 0.00, 5.00),
    ('SOL', 30.00, 70.00, 0.00, 6.00),
    ('USDC', 0.00, 100.00, 0.00, 2.00);

-- Oracle Mark Price History Table
CREATE TABLE oracle_price_ticks (
    tick_id BIGSERIAL PRIMARY KEY,
    asset_symbol VARCHAR(16) NOT NULL,
    pyth_price_usd NUMERIC(18, 6) NOT NULL,
    pyth_confidence_usd NUMERIC(18, 6) NOT NULL,
    pyth_timestamp TIMESTAMPTZ NOT NULL,
    chainlink_price_usd NUMERIC(18, 6) NOT NULL,
    chainlink_timestamp TIMESTAMPTZ NOT NULL,
    effective_mark_price_usd NUMERIC(18, 6) NOT NULL,
    deviation_bps INT NOT NULL,
    usd_inr_fx_rate NUMERIC(12, 6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Synthetic Margin Accounts Table
CREATE TABLE synthetic_margin_accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_account_id UUID NOT NULL UNIQUE,
    total_effective_collateral_usd NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    allocated_e_usd_margin NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    allocated_e_inr_margin NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    locked_margin_usd NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    health_ratio_percentage NUMERIC(8, 2) NOT NULL DEFAULT 100.00,
    is_margin_called BOOLEAN NOT NULL DEFAULT FALSE,
    last_valuation_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Hedging Execution Journal Table
CREATE TABLE hedging_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_account_id UUID NOT NULL,
    asset_symbol VARCHAR(16) NOT NULL,
    venue_type VARCHAR(32) NOT NULL CHECK (venue_type IN ('CEX_SPOT', 'CEX_PERPETUAL', 'DEX_SWAP')),
    venue_name VARCHAR(64) NOT NULL,
    order_side VARCHAR(8) NOT NULL CHECK (order_side IN ('BUY', 'SELL')),
    order_quantity NUMERIC(36, 18) NOT NULL,
    target_price_usd NUMERIC(18, 6) NOT NULL,
    executed_price_usd NUMERIC(18, 6) NOT NULL,
    max_slippage_bps INT NOT NULL,
    realized_slippage_bps INT NOT NULL,
    execution_status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (execution_status IN ('PENDING', 'FILLED', 'PARTIAL', 'REJECTED')),
    tx_or_order_ref VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ
);

-- Collateral Rebalancing & Liquidation Events Table
CREATE TABLE collateral_rebalancing_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_account_id UUID NOT NULL,
    trigger_type VARCHAR(32) NOT NULL CHECK (trigger_type IN ('DYNAMIC_REBALANCE', 'MARGIN_CALL', 'TIERED_LIQUIDATION')),
    pre_event_health_ratio NUMERIC(8, 2) NOT NULL,
    post_event_health_ratio NUMERIC(8, 2) NOT NULL,
    liquidated_asset_symbol VARCHAR(16) NOT NULL,
    liquidated_amount_raw NUMERIC(36, 18) NOT NULL,
    recovery_value_usd NUMERIC(18, 2) NOT NULL,
    liquidation_fee_usd NUMERIC(18, 2) NOT NULL,
    event_status VARCHAR(32) NOT NULL DEFAULT 'COMPLETED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for Query Performance
CREATE INDEX idx_deposits_investor ON crosschain_collateral_deposits (investor_account_id, deposit_status);
CREATE INDEX idx_oracle_ticks_asset ON oracle_price_ticks (asset_symbol, created_at DESC);
CREATE INDEX idx_margin_health ON synthetic_margin_accounts (health_ratio_percentage);
CREATE INDEX idx_hedging_investor ON hedging_executions (investor_account_id, created_at DESC);
```

### Kafka Event Specifications

#### Event: `collateral.deposit.confirmed.v1`
- **Topic:** `collateral.deposit.confirmed.v1`
- **Key:** `investor_account_id`
- **Payload Schema:**
```json
{
  "event_id": "8f7e2a1b-3c4d-4e5f-9a0b-1c2d3e4f5a6b",
  "event_type": "COLLATERAL_DEPOSIT_CONFIRMED",
  "investor_account_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
  "network": "BITCOIN_MAINNET",
  "asset_symbol": "BTC",
  "tx_hash": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
  "amount_raw": "1.50000000",
  "mark_price_usd": "68500.00",
  "haircut_percentage": 20.00,
  "gross_value_usd": "102750.00",
  "effective_collateral_usd": "82200.00",
  "timestamp_utc": 1774000000000
}
```

#### Event: `margin.synthetic.allocated.v1`
- **Topic:** `margin.synthetic.allocated.v1`
- **Key:** `investor_account_id`
- **Payload Schema:**
```json
{
  "event_id": "7d6c5b4a-2e1f-4a3b-8c9d-0e1f2a3b4c5d",
  "event_type": "SYNTHETIC_MARGIN_ALLOCATED",
  "investor_account_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
  "total_effective_collateral_usd": "82200.00",
  "allocated_e_usd": "82200.00",
  "allocated_e_inr": "6822600.00",
  "usd_inr_fx_rate": "83.00",
  "health_ratio_percentage": 180.50,
  "timestamp_utc": 1774000000500
}
```

#### Event: `hedging.trade.executed.v1`
- **Topic:** `hedging.trade.executed.v1`
- **Key:** `investor_account_id`
- **Payload Schema:**
```json
{
  "event_id": "6c5b4a3f-1e2d-4c5b-7a8b-9e0f1a2b3c4d",
  "event_type": "HEDGING_TRADE_EXECUTED",
  "execution_id": "5b4a3f2e-0d1c-4b5a-6f7e-8d9c0b1a2e3f",
  "investor_account_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
  "asset_symbol": "ETH",
  "venue_name": "OKX_PERPETUAL",
  "order_side": "SELL",
  "executed_quantity": "10.000000",
  "executed_price_usd": "3500.25",
  "realized_slippage_bps": 8,
  "timestamp_utc": 1774000001000
}
```

## Security & Compliance Notes
- **Zero PII Principle:** No personally identifiable information (such as Aadhaar, PAN, passport, physical address, or legal names) is ingested, processed, or persisted by this service. The service operates exclusively on anonymized account UUIDs and cryptographic public keys.
- **MPC Custody & Segregated Vaults:** All multi-chain collateral is deposited directly into institutional Multi-Party Computation (MPC) cold/warm vaults (e.g., Fireblocks or Copper) governed by multi-signature approval quorums and programmatic address whitelisting.
- **Oracle Manipulation & Flash Loan Defense:** Mark prices are calculated by cross-verifying Pyth Hermes sub-second feeds against Chainlink Data Feeds. If oracle divergence exceeds 100 bps or confidence interval exceeds 50 bps, trading margin expansion is halted automatically.
- **Slippage Bounds & MEV Protection:** All automated hedging trades routed to DEX venues are submitted via private mempool endpoints (e.g., Flashbots Protect for Ethereum, Jito bundles for Solana) with a strict maximum slippage threshold of 15 bps (hard ceiling 25 bps) to eliminate front-running and sandwich attacks.
- **IFSCA Regulatory Segregation:** Collateral assets held for foreign/international investors participating through GIFT City are maintained in legally ring-fenced offshore custody structures, completely segregated from domestic Indian retail equity margin pools.

## Acceptance Criteria
- [ ] Correctly ingests and validates deposits across Bitcoin (BTC), Ethereum (ETH), Solana (SOL), and USD Coin (USDC) with strict confirmation depth validation.
- [ ] Enforces collateral haircut matrix: BTC 20%, ETH 25%, SOL 30%, and USDC 0% with mathematical accuracy.
- [ ] Aggregates real-time price feeds from Pyth Hermes and Chainlink Data Feeds, correctly enforcing staleness ($\le 30\text{s}$) and deviation ($\le 1.0\%$) checks.
- [ ] Accurately normalizes multi-asset collateral into synthetic USD (`eUSD`) and synthetic INR (`eINR`) trading margin balances.
- [ ] Executes delta hedging orders with slippage protection, rejecting or splitting any order exceeding the configured slippage threshold (15-25 bps).
- [ ] Monitors margin health ratios continuously, issuing margin call alerts at $H_r \le 115\%$ and initiating automated rebalancing/liquidation at $H_r \le 105\%$.
- [ ] Publishes cryptographic state proofs to `SyntheticMarginRegistry.sol` on Hyperledger Besu with zero PII.
- [ ] Emits all required Kafka events (`collateral.deposit.confirmed.v1`, `margin.synthetic.allocated.v1`, `hedging.trade.executed.v1`) with valid JSON payloads.
- [ ] Unit and integration test coverage across all packages exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 002 (Two-Entity Legal/Technical Structure), Prompt 103 (API Standards), Prompt 104 (Kafka Standards), Prompt 203 (Wallet Account Service), Prompt 214 (Foreign-Investor Funding & FX Service).
- **Subsequent / Parallel Tasks:** Prompt 206 (Risk & Margin Checks Service), Prompt 229 (Real-Time VaR Margin Engine), Prompt 308 (On-Chain Proof of Reserve Publishing), Prompt 319 (Institutional Custody Bridge).
