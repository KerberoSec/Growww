# 262 - Automated Liquidity Provisioning & Market Maker Incentive Engine (Go)

## Purpose
Continuous 24/7 off-hours equity trading-introduced in Prompt 253 to decouple fractional tokenized Indian equity and synthetic asset trading from the traditional 09:15-15:30 IST market schedule-presents a critical liquidity challenge. Outside primary exchange operational windows, unmanaged order books risk liquidity drought, widening bid-ask spreads, shallow market depth, and high slippage for retail and institutional participants.

To establish institutional-grade execution quality, tight spreads, and continuous order book depth around the clock, regulated Designated Market Makers (DMMs) and Authorized Liquidity Providers (ALPs) are contracted under SEBI Market Making Regulations and IFSC GIFT City Capital Market frameworks. Under these regulatory covenants, DMMs contractually commit to maintaining continuous, executable two-sided (bid and ask) quotes within strict maximum spread limits and minimum notional depth thresholds for a designated percentage of every trading epoch (e.g., $\ge 90\%$ to $98\%$ uptime).

In return for absorbing inventory holding risk and providing continuous liquidity, DMMs are incentivized through dynamic Maker-Taker fee rebates, volume-tiered fee reimbursements, and epoch-based liquidity pool subsidies funded directly from the exchange's collected platform revenues (Prompt 210) and Treasury Reserve Pool (`NBSEFeeCollector.sol`, Prompt 329).

The **Automated Liquidity Provisioning & Market Maker Incentive Engine** (`services/mm-incentive-engine`) is the automated quantitative surveillance, scoring, and rebate disbursement authority for Growww and NBSE. It continuously ingests tick-by-tick Level-2 order book diffs, tracks DMM quote presence and persistence in real time via sliding-window data structures, evaluates compliance against contractual spread and depth obligations, detects predatory gaming (such as spoofing, quote flickering, and quote stuffing), computes tiered rebate settlements, and coordinates programmatic disbursements across off-chain double-entry ledgers and on-chain smart contracts.

---

## What You Are Building
A high-throughput, low-latency quantitative surveillance and rebate settlement microservice implemented in Go 1.22+ (`services/mm-incentive-engine`). Concrete architectural deliverables include:

- **Sliding-Window Quote Tracker & Uptime Evaluator:** A high-frequency quote ingestion engine tracking DMM two-sided quote presence using Redis Sorted Sets with millisecond-precision timestamp scores across sliding evaluation epochs (e.g., 1-second sampling intervals, 1-hour performance epochs, 24-hour settlement cycles).
- **Two-Sided Depth & Maximum Spread Compliance Engine:** A mathematical evaluation pipeline that verifies whether active DMM quotes fulfill contractual obligations:
  - *Maximum Allowable Spread:* $\text{Spread}_{\text{bps}} = \frac{P_{\text{ask}} - P_{\text{bid}}}{P_{\text{mid}}} \times 10,000 \le \text{MaxSpreadThreshold}$ (e.g., $\le 15$ bps for Large-Cap equities, $\le 30$ bps for Mid-Cap/Commodities).
  - *Minimum Notional Depth:* $\text{Depth}_{\text{notional}} = \min(Q_{\text{bid}} \cdot P_{\text{bid}}, Q_{\text{ask}} \cdot P_{\text{ask}}) \ge \text{MinDepthNotional}$ (e.g., ₹5,00,000 within the top 3 book levels).
- **Anti-Gaming, Spoofing & Quote-Stuffing Surveillance Guard:** A real-time market surveillance layer detecting manipulative behavior:
  - *Quote Flickering:* Sub-50ms cancellations intended to simulate artificial depth without execution intent.
  - *Phantom Liquidity:* Strategic quote withdrawal immediately prior to matching retail orders.
  - *Quote Stuffing:* Excessive Order-to-Trade Ratios (OTR $> 50:1$). Non-compliant DMM quotes are penalized or disqualified from rebate calculations.
- **Multi-Tiered Rebate & Liquidity Scoring Calculator:** A deterministic financial scoring engine computing composite Liquidity Provisioning Scores ($LPS$) and assigning DMM performance tiers:
  - *Tier 1 (Diamond):* $\ge 98\%$ Uptime, $\le 10$ bps Spread, ₹15L Depth $\rightarrow$ Negative maker fee (0.002% net rebate on executed maker volume) + Tier 1 Treasury Pool subsidy split.
  - *Tier 2 (Gold):* $\ge 95\%$ Uptime, $\le 20$ bps Spread, ₹10L Depth $\rightarrow 100\%$ Maker fee rebate (0.000% net fee) + Tier 2 Treasury Pool subsidy split.
  - *Tier 3 (Silver):* $\ge 90\%$ Uptime, $\le 30$ bps Spread, ₹5L Depth $\rightarrow 70\%$ Maker fee rebate.
  - *Sub-Threshold / Disqualified:* $< 90\%$ Uptime or spread breach $\rightarrow 0\%$ Rebate (standard 0.00% fee (No fee at all) applies) and automated alert to exchange compliance.
- **Dual-Dispatch Rebate Settlement Relayer:** Automated disbursement coordinator that executes:
  - Off-chain ledger credits via the Wallet & Account Service (Prompt 203) using idempotent double-entry journal postings.
  - On-chain Treasury Pool distributions via `NBSEFeeCollector.sol` (Prompt 329) deployed on Hyperledger Besu under QBFT consensus.
- **High-Performance gRPC Service (`MMIncentiveService`):** Internal API exposing DMM performance scorecards, real-time compliance telemetry, historical rebate statements, and administrative threshold configurations.
- **Admin & Back-Office Integration (Prompt 217):** Maker-Checker dual-control workflow bindings for updating DMM tier parameters, minimum depth thresholds, and authorizing exception payouts or disqualifications.

---

## Scope Boundaries

### In Scope
- Real-time ingestion and validation of DMM quotes from Market Data Level-2 streams (Prompt 207) and execution drop copies from the Matching Engine (Prompt 205 / Prompt 253).
- Sub-second sliding-window state tracking in Redis Sorted Sets (`ZADD`, `ZREMRANGEBYSCORE`, `ZRANGEBYSCORE`).
- Exact mathematical calculation of Uptime Percentage, Time-Weighted Average Spread (TWAS), and Depth-Weighted Liquidity Provisioning Scores ($LPS$).
- Surveillance heuristics to detect and penalize quote flickering ($<50\text{ms}$ cancellations), high Order-to-Trade Ratios ($>50:1$), and spoofing.
- Calculation of volume-weighted maker rebates and statutory Treasury Pool subsidy distributions.
- Dual-dispatch of rebate instructions: off-chain double-entry ledgers (Prompt 203) and on-chain smart contracts (Prompt 329).
- Definition and compilation of Protobuf contracts (`MMPerformanceReport`, `DMMTierConfig`, `EpochEvaluationResult`).
- Audit trail instrumentation and Prometheus metrics for exchange surveillance and SEBI compliance.

### Out of Scope / Handled Elsewhere
- Central Limit Order Book matching and execution generation (handled in Prompt 205 and Prompt 253).
- Direct binary connectivity to primary domestic exchanges (NSE/BSE Bhavcopy ingestion, handled in Prompt 242).
- Pre-trade margin verification and real-time SPAN liquidation triggers (handled in Prompt 206 and Prompt 241).
- General institutional onboarding, KYC, and AML screening (handled in Prompt 202 and Prompt 214).
- User-facing Flutter and Web mobile charting interfaces (handled in Category 5 and Category 6).
- Physical demat asset custody and depository lockups (handled in Prompt 213).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Go provides ultra-efficient concurrency via lightweight goroutines, low-latency garbage collection tuned for financial streams, and native support for gRPC and Redis.
- **In-Memory Sliding Window State Store:** Redis 7.2+ Cluster utilizing Redis Sorted Sets (`ZSET`) for millisecond-precision quote timestamp indexing, Redis Hashes for active DMM configuration caches, and Redis Pub/Sub for sub-second surveillance breach broadcasting.
- **Relational Persistence & Audit Ledger:** PostgreSQL 16+ using `pgx/v5` connection pooling and `sqlc` for compile-time verified, zero-allocation database queries.
- **Message Broker:** Apache Kafka 3.7+ (`segmentio/kafka-go`) consuming order book updates and trade execution events, and publishing compliance evaluations and rebate settlements.
- **Financial Mathematics Library:** `github.com/shopspring/decimal` for fixed-point arbitrary precision arithmetic (utilizing `ROUND_HALF_EVEN` / Banker's Rounding to 4 decimal places for INR paise and 8 decimal places for tokenized equity units).
- **Blockchain RPC & Cryptography:** `github.com/ethereum/go-ethereum/ethclient` for interacting with Hyperledger Besu JSON-RPC endpoints and CloudHSM-backed relayer keys for signing EIP-712 disbursement attestations.

---

## Backend / Infra Touchpoints
- **Market Data Service L2 Order Books (Prompt 207):** Ingests streaming Level-2 order book diffs via Kafka topic `matching.depth.v1` to monitor top 5-20 price levels, quote persistence, and active quotes tagged with DMM participant IDs.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes base transaction fee debits (standard 0.00% (Zero Fee) platform fee) and reconciles offset credits from computed maker rebates.
- **Admin & Back-Office Service (Prompt 217):** Provides Maker-Checker dual-control operational workflows for DMM onboarding, parameter adjustments, and emergency disqualifications.
- **Order Matching Engine & After-Hours Gateway (Prompt 205 / Prompt 253):** Ingests raw execution drop copies (`matching.trades.v1`) to correlate resting passive maker orders placed by DMMs versus aggressive taker executions.
- **Wallet & Account Service (Prompt 203):** Dispatches double-entry journal entries crediting DMM operating accounts (`REBATE_INCENTIVE_EXPENSE` debit, `DMM_OPERATING_ACCOUNT` credit).
- **Real-Time Market Surveillance & Audit Log Service (Prompt 218 / Prompt 228):** Ingests surveillance alerts (excessive OTR, flickering, spoofing) for SEBI regulatory reporting and institutional audit logging.

---

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Treasury Pool Distribution (`NBSEFeeCollector.sol`, Prompt 329):** The engine interfaces with the exchange Treasury Reserve pool established under Prompt 329. A statutory allocation (e.g., 15% of the 60% Exchange Treasury fee pool) is earmarked for DMM liquidity maintenance.
- **Cryptographic Rebate Attestation:** At the close of each performance epoch (e.g., hourly or daily at 09:00 IST), the engine constructs a Merkle tree of all DMM performance metrics and publishes an on-chain cryptographic attestation:
  $$\text{RebateEpochRoot} = \text{MerkleRoot}\Big(\big\{\text{SHA256}(\text{dmm\_id} \,\|\, \text{epoch\_id} \,\|\, \text{uptime\_pct} \,\|\, \text{twas} \,\|\, \text{rebate\_amount})\big\}\Big)$$
- **On-Chain Disbursement Invocation:** An authorized settlement relayer invokes `claimMarketMakerRebate(uint256 epochId, address dmmAddress, uint256 rebateAmount, bytes32[] calldata merkleProof)` on `NBSEFeeCollector.sol`. The contract verifies the proof and transfers Digital Rupee (eINR / CBDC) from the Treasury pool directly to the DMM's whitelisted custodian wallet.
- **Zero On-Chain PII Invariant:** The ledger records only pseudonymous institutional IDs (`DMM_IN_XXXX`), whitelisted contract addresses, cryptographic Merkle roots, and numerical rebate amounts. No trader identities, PANs, or banking credentials exist on-chain.

---

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/mm-incentive-engine` with modular directory structure:
   - `cmd/server/main.go`: Application entry point and lifecycle manager.
   - `internal/evaluator/`: Spread, depth, and uptime compliance mathematical engine.
   - `internal/slidingwindow/`: Redis-backed sliding-window quote tracker.
   - `internal/rebate/`: Maker-taker rebate and Treasury Pool subsidy calculator.
   - `internal/surveillance/`: Anti-spoofing, quote flickering, and OTR enforcement guards.
   - `internal/store/`: PostgreSQL schemas and SQL queries via `sqlc`.
   - `internal/relayer/`: Blockchain disbursement and Wallet Service dispatcher.
2. **Define Protobuf Contracts:** Author `proto/growww/mm/v1/mm_incentive_service.proto` specifying `MMPerformanceReport`, `DMMTierConfig`, `EpochEvaluationResult`, `QueryDMMPerformanceRequest`, and `RebateDisbursementInstruction`.
3. **Generate Go Stubs:** Compile protobuf schemas into Go gRPC server interfaces and serialization types using `protoc-gen-go` and `protoc-gen-go-grpc`.
4. **Design PostgreSQL Schema:** Author SQL migrations establishing tables: `dmm_registrations`, `dmm_asset_obligations`, `dmm_evaluation_epochs`, `dmm_rebate_ledgers`, and `dmm_surveillance_penalties`.
5. **Implement Redis Sliding-Window Quote Buffer:** Build Redis client managing sorted sets (`zset:mm:quotes:{isin}:{dmm_id}`) where elements store serialized quote snapshots and scores represent Unix millisecond timestamps. Implement automated trimming (`ZREMRANGEBYSCORE`) to prune items older than the evaluation window.
6. **Implement Market Data L2 Stream Ingestor:** Build Kafka consumer listening to `matching.depth.v1` and `marketdata.ticks.v1`. Filter order book updates to extract top-of-book and multi-level quotes tagged with DMM participant IDs.
7. **Build Spread & Depth Compliance Validator:** Implement real-time mathematical validation checking whether active DMM quotes satisfy:
   - $\text{Spread}_{\text{bps}} \le \text{MaxSpreadThreshold}$
   - $\text{NotionalDepth} \ge \text{MinDepthNotional}$
8. **Build Anti-Spoofing & Quote-Stuffing Guard:** Implement sliding-window surveillance tracking:
   - Quote lifetimes: flag cancellations occurring $< 50\text{ms}$ from placement without execution.
   - Order-to-Trade Ratio (OTR): calculate actions per execution; flag DMMs exceeding regulatory limits ($> 50:1$).
   - Apply automatic uptime deduction penalties or epoch disqualification upon confirmed violations.
9. **Implement Uptime & Time-Weighted Average Spread (TWAS) Evaluator:** Execute deterministic sampling every 1,000ms. Measure continuous compliant seconds against total epoch seconds to compute Uptime % and integrate spread over time for TWAS.
10. **Implement Multi-Tiered Liquidity Scoring Engine:** Calculate composite score $LPS$ using weighted coefficients across Uptime, TWAS compliance, and Depth multiples. Map scores to Diamond, Gold, Silver, or Disqualified tiers.
11. **Implement Maker-Taker Fee Rebate Calculator:** Ingest execution events from `matching.trades.v1`. Identify passive fills where the DMM provided liquidity, calculate the maker rebate based on the DMM's tier, and allocate the corresponding share of the Treasury Pool subsidy.
12. **Implement Off-Chain Wallet Disbursement Relayer:** Construct multi-legged journal entries in PostgreSQL and invoke `WalletService.CreditAccount` (Prompt 203) using deterministic idempotency keys.
13. **Implement On-Chain Smart Contract Relayer:** Generate EIP-712 typed data payloads and Merkle proofs for `claimMarketMakerRebate`, sign payloads with relayer keys stored in CloudHSM (Prompt 717), and broadcast transactions to `NBSEFeeCollector.sol` on Hyperledger Besu (Prompt 329).
14. **Build Admin & Back-Office Control Hook (Prompt 217):** Implement gRPC handlers enabling compliance officers to configure tier thresholds, inspect DMM performance scorecards, and enforce dual-control overrides.
15. **Configure Prometheus Telemetry & Invariant Tests:** Expose Prometheus gauges for DMM uptime, TWAS histograms, and rebate disbursement counters. Implement comprehensive unit, fuzz, and stress tests verifying Banker's Rounding precision, zero-leakage conservation, and sub-5ms evaluation latency.

---

## Interfaces / Contracts

### Protobuf Contract (`proto/growww/mm/v1/mm_incentive_service.proto`)
```protobuf
syntax = "proto3";

package growww.mm.v1;

option go_package = "growww/mm/v1;mmv1";

service MMIncentiveService {
  rpc GetDMMPerformanceReport (GetDMMPerformanceReportRequest) returns (MMPerformanceReport);
  rpc QueryEpochEvaluation (QueryEpochEvaluationRequest) returns (EpochEvaluationResult);
  rpc UpdateDMMTierConfig (UpdateDMMTierConfigRequest) returns (UpdateDMMTierConfigResponse);
  rpc TriggerManualRebateSettlement (TriggerManualRebateSettlementRequest) returns (RebateDisbursementInstruction);
}

enum DMMComplianceTier {
  DMM_COMPLIANCE_TIER_UNSPECIFIED = 0;
  DMM_COMPLIANCE_TIER_DIAMOND = 1;    // >= 98% Uptime, <= 10 bps Spread, >= 15L Depth
  DMM_COMPLIANCE_TIER_GOLD = 2;       // >= 95% Uptime, <= 20 bps Spread, >= 10L Depth
  DMM_COMPLIANCE_TIER_SILVER = 3;     // >= 90% Uptime, <= 30 bps Spread, >= 5L Depth
  DMM_COMPLIANCE_TIER_DISQUALIFIED = 4; // < 90% Uptime or Surveillance Breach
}

message DMMTierConfig {
  string config_id = 1;
  string asset_class = 2; // EQUITY_LARGE_CAP, EQUITY_MID_CAP, COMMODITY, DERIVATIVE
  uint32 min_uptime_basis_points = 3; // e.g. 9500 for 95.00%
  uint32 max_spread_basis_points = 4; // e.g. 15 for 0.15%
  string min_notional_depth_inr = 5;  // e.g. "1000000.0000"
  int64 maker_rebate_basis_points = 6; // e.g. -2 for -0.02% (negative fee) or 100 for 100% reimbursement
  string treasury_subsidy_weight = 7; // Fractional share of epoch pool (e.g. "0.40")
  uint32 max_order_to_trade_ratio = 8; // e.g. 50 (max 50 order actions per trade)
}

message MMPerformanceReport {
  string report_id = 1;
  string dmm_id = 2;
  string isin = 3;
  string epoch_id = 4;
  int64 epoch_start_unix_ms = 5;
  int64 epoch_end_unix_ms = 6;
  
  // Quantitative Compliance Metrics
  double uptime_percentage = 7;         // Actual compliant uptime (e.g. 96.45%)
  double time_weighted_avg_spread_bps = 8; // TWAS across epoch
  string avg_two_sided_depth_inr = 9;   // Average resting two-sided depth
  uint32 total_samples = 10;
  uint32 compliant_samples = 11;
  
  // Surveillance & Anti-Gaming Metrics
  uint32 order_to_trade_ratio = 12;
  uint32 sub_50ms_cancellation_count = 13;
  bool spoofing_flagged = 14;
  
  // Tier & Rebate Accounting
  DMMComplianceTier achieved_tier = 15;
  string gross_maker_volume_inr = 16;
  string maker_fee_rebate_inr = 17;
  string treasury_subsidy_inr = 18;
  string total_rebate_payable_inr = 19;
  string computation_proof_hash = 20; // SHA-256 Merkle leaf
}

message GetDMMPerformanceReportRequest {
  string dmm_id = 1;
  string isin = 2;
  string epoch_id = 3;
}

message QueryEpochEvaluationRequest {
  string epoch_id = 1;
}

message EpochEvaluationResult {
  string epoch_id = 1;
  int64 evaluated_at_unix = 2;
  repeated MMPerformanceReport reports = 3;
  string total_treasury_pool_distributed_inr = 4;
  string on_chain_merkle_root = 5;
}

message UpdateDMMTierConfigRequest {
  string admin_user_id = 1;
  string maker_approval_token = 2;
  DMMTierConfig config = 3;
}

message UpdateDMMTierConfigResponse {
  bool success = 1;
  string config_id = 2;
  int64 updated_at_unix = 3;
}

message TriggerManualRebateSettlementRequest {
  string epoch_id = 1;
  string admin_user_id = 2;
  string checker_approval_token = 3;
}

message RebateDisbursementInstruction {
  string instruction_id = 1;
  string epoch_id = 2;
  string dmm_id = 3;
  string dmm_wallet_address = 4;
  string amount_inr = 5;
  string off_chain_journal_id = 6;
  string on_chain_tx_hash = 7;
  string status = 8; // PENDING, DISPATCHED, CONFIRMED, FAILED
}
```

---

### PostgreSQL Database Schema DDL
```sql
-- DMM Institutional Registration
CREATE TABLE dmm_registrations (
    dmm_id VARCHAR(32) PRIMARY KEY, -- e.g. 'DMM_IN_ALPHA_CAPITAL'
    entity_name VARCHAR(128) NOT NULL,
    sebi_registration_no VARCHAR(64) NOT NULL UNIQUE,
    settlement_wallet_address VARCHAR(42) NOT NULL, -- On-chain 0x Besu address
    operational_status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (operational_status IN ('ACTIVE', 'SUSPENDED', 'TERMINATED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contractual Obligations per Asset Class / Symbol
CREATE TABLE dmm_asset_obligations (
    obligation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dmm_id VARCHAR(32) NOT NULL REFERENCES dmm_registrations(dmm_id) ON DELETE CASCADE,
    isin VARCHAR(12) NOT NULL,
    market_session VARCHAR(16) NOT NULL CHECK (market_session IN ('PRIMARY_HOURS', 'AFTER_HOURS_24_7', 'WEEKEND')),
    min_uptime_pct NUMERIC(5, 2) NOT NULL DEFAULT 95.00,
    max_spread_bps NUMERIC(6, 2) NOT NULL DEFAULT 15.00,
    min_notional_depth NUMERIC(18, 4) NOT NULL DEFAULT 1000000.0000,
    max_otr_threshold INTEGER NOT NULL DEFAULT 50,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (dmm_id, isin, market_session)
);

-- Evaluation Epochs (Hourly / Daily Settlement Windows)
CREATE TABLE dmm_evaluation_epochs (
    epoch_id VARCHAR(64) PRIMARY KEY, -- e.g. 'EPOCH_2026_09_18_H23'
    epoch_type VARCHAR(16) NOT NULL CHECK (epoch_type IN ('HOURLY', 'DAILY_SETTLEMENT')),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    treasury_subsidy_allocated NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    merkle_root VARCHAR(66),
    is_evaluated BOOLEAN NOT NULL DEFAULT FALSE,
    is_settled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Historical DMM Performance & Rebate Accounting
CREATE TABLE dmm_rebate_ledgers (
    ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epoch_id VARCHAR(64) NOT NULL REFERENCES dmm_evaluation_epochs(epoch_id) ON DELETE CASCADE,
    dmm_id VARCHAR(32) NOT NULL REFERENCES dmm_registrations(dmm_id),
    isin VARCHAR(12) NOT NULL,
    uptime_pct NUMERIC(5, 2) NOT NULL,
    twas_bps NUMERIC(8, 4) NOT NULL,
    avg_depth_inr NUMERIC(18, 4) NOT NULL,
    order_to_trade_ratio INTEGER NOT NULL,
    flicker_cancel_count INTEGER NOT NULL DEFAULT 0,
    compliance_tier VARCHAR(16) NOT NULL CHECK (compliance_tier IN ('DIAMOND', 'GOLD', 'SILVER', 'DISQUALIFIED')),
    maker_volume_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    maker_rebate_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    treasury_subsidy_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_rebate_inr NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    proof_hash VARCHAR(64) NOT NULL,
    off_chain_journal_id UUID,
    on_chain_tx_hash VARCHAR(66),
    settlement_status VARCHAR(16) NOT NULL DEFAULT 'PENDING' CHECK (settlement_status IN ('PENDING', 'DISPATCHED', 'SETTLED', 'FAILED', 'PENALIZED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (epoch_id, dmm_id, isin)
);

-- Surveillance Infractions & Penalties
CREATE TABLE dmm_surveillance_penalties (
    penalty_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dmm_id VARCHAR(32) NOT NULL REFERENCES dmm_registrations(dmm_id),
    epoch_id VARCHAR(64) NOT NULL REFERENCES dmm_evaluation_epochs(epoch_id),
    isin VARCHAR(12) NOT NULL,
    infraction_type VARCHAR(32) NOT NULL CHECK (infraction_type IN ('QUOTE_FLICKERING', 'EXCESSIVE_OTR', 'SPOOFING_ATTEMPT', 'PHANTOM_WITHDRAWAL')),
    details JSONB NOT NULL,
    deduction_pct NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    is_disqualified BOOLEAN NOT NULL DEFAULT FALSE,
    flagged_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for Low-Latency Querying
CREATE INDEX idx_dmm_rebate_lookup ON dmm_rebate_ledgers(dmm_id, epoch_id);
CREATE INDEX idx_dmm_obligation_lookup ON dmm_asset_obligations(dmm_id, isin, is_active);
CREATE INDEX idx_dmm_epoch_status ON dmm_evaluation_epochs(is_evaluated, is_settled);
CREATE INDEX idx_dmm_penalties_epoch ON dmm_surveillance_penalties(epoch_id, dmm_id);
```

---

### Mathematical Rebate & Compliance Formulas

1. **Spread in Basis Points ($\text{Spread}_{\text{bps}}$):**
   $$P_{\text{mid}} = \frac{P_{\text{bid}} + P_{\text{ask}}}{2}$$
   $$\text{Spread}_{\text{bps}} = \left(\frac{P_{\text{ask}} - P_{\text{bid}}}{P_{\text{mid}}}\right) \times 10,000$$

2. **Two-Sided Notional Depth ($\text{Depth}_{\text{notional}}$):**
   $$\text{Depth}_{\text{bid}} = \sum_{i=1}^{K} Q_{\text{bid}, i} \cdot P_{\text{bid}, i}, \quad \text{Depth}_{\text{ask}} = \sum_{j=1}^{K} Q_{\text{ask}, j} \cdot P_{\text{ask}, j}$$
   $$\text{Depth}_{\text{notional}} = \min(\text{Depth}_{\text{bid}}, \text{Depth}_{\text{ask}}) \ge D_{\text{min}}$$

3. **Time-Weighted Average Spread (TWAS):**
   $$\text{TWAS} = \frac{\sum_{t=1}^{N} \text{Spread}_{\text{bps}}(t) \cdot \Delta t_t}{\sum_{t=1}^{N} \Delta t_t}$$

4. **Compliant Uptime Percentage ($\text{Uptime}_{\%}$):**
   $$\text{Sample Compliant } C_t = \begin{cases} 1 & \text{if } \text{Spread}_{\text{bps}}(t) \le S_{\text{max}} \land \text{Depth}_{\text{notional}}(t) \ge D_{\text{min}} \land \neg \text{Flicker}(t) \\ 0 & \text{otherwise} \end{cases}$$
   $$\text{Uptime}_{\%} = \left(\frac{\sum_{t=1}^{N} C_t}{N}\right) \times 100$$

5. **Order-to-Trade Ratio (OTR):**
   $$\text{OTR} = \frac{\text{Total Order Actions (Submissions + Modifications + Cancellations)}}{\max(1, \text{Executed Trades})}$$
   $$\text{If } \text{OTR} > \text{MaxOTR}_{\text{regulatory}} \implies \text{Apply Penalty Deduction } P_{\text{OTR}}$$

6. **Composite Liquidity Provisioning Score ($LPS$):**
   $$LPS = w_1 \cdot \text{Uptime}_{\%} + w_2 \cdot \min\left(1.0, \frac{S_{\text{target}}}{\text{TWAS}}\right) \times 100 + w_3 \cdot \min\left(1.5, \frac{\bar{D}}{D_{\text{min}}}\right) \times 100 - \text{Penalty}_{\text{surveillance}}$$
   *(Nominal weights: $w_1 = 0.50$, $w_2 = 0.30$, $w_3 = 0.20$)*

7. **Net Rebate Calculation:**
   $$\text{MakerRebate}_{\text{inr}} = \text{MakerVolume}_{\text{inr}} \times R_{\text{tier}}$$
   $$\text{TreasurySubsidy}_{\text{inr}} = \text{TreasuryPool}_{\text{epoch}} \times \left(\frac{LPS_{\text{dmm}}}{\sum_{k} LPS_k}\right)$$
   $$\text{TotalRebatePayable} = \text{MakerRebate}_{\text{inr}} + \text{TreasurySubsidy}_{\text{inr}}$$

---

## Security & Compliance Notes
- **SEBI Market Making Guidelines Alignment:** System architecture strictly complies with SEBI circulars governing market making on nationwide stock exchanges and IFSC GIFT City stock exchanges. Designated Market Makers must maintain mandatory two-sided quotes for at least 90% (regular) or 95% (after-hours) of the trading session. Quotes must not exceed pre-defined spread bands (e.g., 15-30 bps) and must support minimum executable lots.
- **Anti-Spoofing & Quote Flickering Prevention:** The engine inspects quote cancellation velocities. Quotes posted and withdrawn within $< 50\text{ms}$ are flagged as non-bona-fide "flicker" quotes. They are disqualified from depth and uptime accumulation and penalized under automated market abuse rules. Repeated violations trigger automatic suspension and dispatch alert tickets to the Surveillance Desk (Prompt 228).
- **Order-to-Trade Ratio (OTR) Regulatory Enforcement:** In accordance with SEBI algorithmic trading guidelines, DMMs exceeding an OTR of 50:1 incur progressive economic penalties (10% to 50% rebate forfeiture). Sustained OTR $> 100:1$ results in immediate epoch disqualification.
- **Self-Trade Prevention (STP):** Orders submitted by a DMM that would match against resting orders from the same DMM participant ID are intercepted and blocked at the Matching Engine (Prompt 205). The engine scrubs any attempted self-trades from qualifying maker turnover volumes.
- **Treasury Pool Solvency Invariant:** Total maker rebates and liquidity subsidies distributed across all DMMs in an epoch are mathematically capped by the statutory Market Maker Subsidy Reserve allocated in `NBSEFeeCollector.sol`. Subsidies are pro-rata normalized if aggregate claims exceed available treasury funds, guaranteeing zero deficit risk.
- **Cryptographic Auditability & Idempotency:** Each epoch settlement produces a deterministic SHA-256 Merkle root. Every disbursement instruction requires an idempotent token derived from `SHA256(dmm_id || epoch_id || total_rebate_inr)`. Duplicate debit or credit execution is physically impossible across both off-chain PostgreSQL ledgers and on-chain Besu smart contracts.

---

## Acceptance Criteria
- [ ] Engine ingests Level-2 order book updates from Kafka (`matching.depth.v1`) and validates DMM quote presence and spread within $< 5\text{ms}$ latency.
- [ ] Redis sliding-window sorted sets accurately index millisecond quote timestamps and prune expired records without memory exhaustion.
- [ ] Compliant Uptime Percentage and TWAS calculations match analytical test matrices with $10^{-6}$ numerical precision.
- [ ] Quotes exceeding maximum spread thresholds or failing minimum notional depth are excluded from compliant uptime accumulation.
- [ ] Surveillance module detects quote flickering ($<50\text{ms}$ cancellations) and high OTR ($>50:1$), properly logging infractions and applying score penalties.
- [ ] Composite Liquidity Provisioning Score ($LPS$) accurately assigns DMMs to Diamond, Gold, Silver, or Disqualified tiers.
- [ ] Maker fee rebates and Treasury Pool subsidies are computed without micro-paisa leakage using Banker's Rounding (`ROUND_HALF_EVEN`).
- [ ] Idempotent double-entry journal postings are dispatched to Wallet Service (Prompt 203) with zero balance mismatches.
- [ ] On-chain Merkle roots and EIP-712 transaction payloads for `NBSEFeeCollector.sol` (Prompt 329) are correctly formatted and verifiable via Besu RPC.
- [ ] Admin Back-Office (Prompt 217) Maker-Checker workflows successfully authorize parameter modifications and manual overrides.
- [ ] Prometheus metrics accurately report real-time DMM uptime, TWAS distributions, and cumulative rebate amounts.

---

## Suggested Order / Dependencies
- **Prerequisites:**
  - `006_fee_model_specification.md` (0.00% fee (No fee at all) model and 0.00% fee launch policy revenue split)
  - `205_order_matching_engine.md` (Execution drop copies and trade matching)
  - `207_market_data_service.md` (L2 order book streaming diffs)
  - `210_fee_and_realized_pnl_engine.md` (Platform fee calculation and journal postings)
  - `253_continuous_24_7_synthetic_market_and_after_hours_gateway.md` (24/7 after-hours trading session states)
  - `329_nbse_settlement_dvp_and_fee_collector.md` (`NBSEFeeCollector.sol` Treasury Reserve disbursement)
- **Parallel Tasks:**
  - `217_admin_back_office_service.md` (Maker-Checker DMM configuration UI)
  - `218_audit_log_service.md` (Immutable regulatory event storage)
  - `228_real_time_market_surveillance.md` (Cross-market spoofing and manipulation detection)
  - `244_nbse_fixed_fee_and_revenue_distribution_engine.md` (Automated statutory treasury distribution)
- **Downstream Blockers:**
  - `512_flutter_trade_and_performance_history.md` (DMM institutional analytics dashboard)
  - `604_admin_web_operations_and_surveillance_dashboard.md` (Exchange operator surveillance console)
