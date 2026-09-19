# 247 - Cross-Chain Reorg Invalidation and Clawback Saga Coordinator (Go / Rust)

## Purpose
In a cross-chain institutional financial exchange infrastructure, external blockchain networks (such as Bitcoin, Ethereum L1, Solana, EVM L2s, and Cosmos zones) operate under probabilistic or epoch-based finality models. Under adverse network conditions (e.g., deep chain reorganizations, validator consensus partitions, dishonest majority 51% attacks, or L2 sequencer equivocation), previously confirmed and credited cross-chain collateral deposits can be invalidated, orphaned, or replaced by conflicting transaction histories after credit has already been granted on the platform.

If an investor has deposited cross-chain collateral that is subsequently reorganized out of the canonical chain, the credited assets become phantom backing. By that time, the user may have already minted synthetic asset tokens, placed leveraged derivatives orders, established open margin positions, or triggered automatic external delta hedges. 

The **Cross-Chain Reorg Invalidation and Clawback Saga Coordinator** (`services/clawback-saga-coordinator`) is a mission-critical, distributed orchestration engine implemented in Go and Rust. It executes an atomic, deterministic, 5-phase compensating Saga workflow to instantly freeze compromised accounts, purge active orders, prioritize portfolio liquidations, unwind external CEX/DEX delta hedges, and burn orphaned on-chain synthetic tokens on the permissioned Hyperledger Besu network. This coordinator prevents exchange balance sheet insolvency, eliminates systemic shortfall, and satisfies CPMI-IOSCO Principle 8 (Settlement Finality) and Principle 17 (Operational Risk).

## What You Are Building
A fault-tolerant, high-speed distributed Saga orchestrator and compensation coordinator (`services/clawback-saga-coordinator`) combining a Go orchestration supervisor with a Rust high-concurrency event processing core (`crates/clawback-core`), backed by PostgreSQL 16, Redis 7.2 Cluster, and Apache Kafka. Concrete deliverables include:

- **Deep Reorg Invalidation Ingress & Detection Listener:** A sub-millisecond Kafka event listener subscribing to reorg detection streams from Bitcoin, EVM, and Solana bridge indexers (Prompts 234, 235, 236) to parse invalidated transaction hashes, orphaned block heights, and impacted user deposits.
- **Deterministic 5-Phase Saga Orchestration Engine:** A stateful, forward-recovery and backward-compensating workflow coordinator executing the non-negotiable clawback sequence:
  * **Phase 1 (Account Freeze):** Immediate transition of user account state to `MARGIN_RECOVERY_LOCKED` across Redis and PostgreSQL, revoking trading, withdrawal, and API session privileges.
  * **Phase 2 (Order Cancellation):** Synchronous mass cancellation of all resting limit, stop-loss, bracket, and algorithmic trigger orders in the Order Matching Engine (Prompt 205) and Advanced Order Engine (Prompt 226).
  * **Phase 3 (Priority Auto-Liquidation):** Escalated priority portfolio margin liquidation via the SPAN Risk Engine (Prompt 241) and Perpetuals Engine (Prompt 240) to close out open derivative positions and capture remaining equity.
  * **Phase 4 (Delta Hedge Unwinding on CEX/DEX):** Coordinated release and unwinding of off-platform delta hedges via the Cross-Chain Router (Prompt 238) and SOR Gateway (Prompt 242) to prevent toxic unhedged basis risk.
  * **Phase 5 (On-Chain Synthetic Token Burn):** Execution of authorized burn transactions on Hyperledger Besu (`SyntheticTokenRegistry.sol` / `CrossChainBridgeVault.sol`) via relayer HSM to destroy phantom synthetic representations and balance the Proof-of-Reserve ledger.
- **Distributed Lock & Idempotency Manager:** Redlock-backed single-flight execution coordinator ensuring that a single reorg event across multiple blocks or deposits triggers exactly one coordinated Saga without race conditions or duplicate liquidation calls.
- **SGF Default Escalation Circuit:** An automated fail-safe bridge that routes unrecoverable negative equity deficits directly to the Settlement Guarantee Fund Default Waterfall (Prompt 230) when liquidated collateral is insufficient to cover reorg losses.
- **Cryptographic Audit Trail & Attestation Logger:** An immutable PostgreSQL journal generating SHA-256 Merkle proofs for every completed Saga phase, published to the Audit Log Service (Prompt 218) and Regulatory Reporting Dispatcher (Prompt 233).

## Scope Boundaries
- **In Scope:**
  - Ingesting reorg and invalidation events from bridge ingress adapters (Prompts 234, 235, 236).
  - Enforcing the strict 5-phase sequential Saga workflow (`MARGIN_RECOVERY_LOCKED` -> Order Purge -> Auto-Liquidation -> Hedge Unwind -> Besu Token Burn).
  - Orchestrating gRPC and Kafka RPC commands to Risk, Order, Matching, Ledger, Router, and Smart Contract services.
  - Managing distributed Saga state persistence, transition timeouts, retry policies, and compensation rollbacks in PostgreSQL.
  - Triggering SGF Default Waterfall (Prompt 230) allocation if negative equity remains after liquidation.
  - Maintaining zero-PII cryptographic attestation receipts for all clawback actions.
- **Out of Scope / Handled Elsewhere:**
  - Raw blockchain block header ingestion and SPV light client proofs (handled in Prompts 234, 235, 236, 322, 324).
  - Low-level matching engine order book execution (handled in Prompt 205).
  - Real-time SPAN margin array calculation algorithms (handled in Prompt 241).
  - External exchange API order execution connectors (handled in Prompt 238).
  - KYC/AML sanctions screening and regulatory identity verification (handled in Prompt 202).

## Technology to Use
- **Primary Languages & Runtimes:**
  - **Go 1.22+:** Saga supervisor, gRPC client/server, Kafka consumer group management, and database transaction coordination (`services/clawback-saga-coordinator`).
  - **Rust 1.78+:** High-throughput event parsing core, cryptographic Merkle proof generator, and concurrent state machine validator (`crates/clawback-core`).
- **Database & State Storage:**
  - **PostgreSQL 16+:** ACID-compliant relational store for persistent Saga instance records, phase execution logs, and compensation journal entries.
  - **Redis 7.2+ Cluster:** In-memory distributed lock manager (Redlock algorithm), real-time account freeze bitmap flags, and sub-millisecond execution state caching.
- **Message Broker & Event Streaming:**
  - **Apache Kafka 3.7+:** Partitioned event streaming with exactly-once semantic (EOS) transactional producers for phase state transitions and invalidation notifications.
- **Inter-Service Communication:**
  - **gRPC / Protocol Buffers v3:** Low-latency mTLS RPCs for direct command dispatch to Order Service, Matching Engine, and Risk Service.
- **Distributed Coordination & Workflow:**
  - **Temporal.io / Embedded Go Saga Pattern:** Deterministic workflow execution engine with configurable backoff retries, heartbeat monitoring, and compensation branches.
- **Cryptography & Proofs:**
  - `ring` / `sha2` (Rust) for sub-microsecond SHA-256 Merkle tree leaf generation and attestation hashing.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:**
  - `saga_instances`, `saga_phase_executions`, `saga_compensations`, `reorg_invalidation_events`, `clawback_ledger_journals`, `sgf_deficit_transfers`.
- **Redis 7.2 Keys:**
  - `lock:saga:reorg:{chain_id}:{reorg_id}`: Distributed mutex ensuring single coordinator execution per reorg incident.
  - `account:status:override:{user_id}`: High-priority bitflag evaluated by API Gateway and Risk Engine before routing any trade.
  - `saga:state:{saga_id}`: Ephemeral cached state of active Saga instance for real-time observability.
- **Apache Kafka Topics:**
  - Consumes: `bridge.reorg.detected.v1`, `bridge.deposit.invalidated.v1`, `rms.liquidation.completed.v1`, `hedging.unwind.completed.v1`, `blockchain.besu.burn_confirmed.v1`.
  - Publishes: `saga.clawback.initiated.v1`, `saga.clawback.phase_transition.v1`, `saga.clawback.completed.v1`, `saga.clawback.failed.v1`, `account.frozen.v1`.
- **Order Matching Engine (Prompt 205):** Accepts immediate synchronous batch order cancel requests for specific account IDs.
- **SPAN Portfolio Margin & Liquidation Engine (Prompt 241):** Accepts priority liquidation triggers bypassing standard margin call grace periods.
- **Cross-Chain Collateral & Synthetic FX Router (Prompt 238):** Ingests delta hedge unwinding orders to liquidate off-platform inventory on CEXs and DEXs.
- **Settlement Guarantee Fund Service (Prompt 230):** Receives deficit claims when liquidation proceeds fail to cover invalidated collateral.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Phase 5 Synthetic Token Burn:** The coordinator interfaces with Hyperledger Besu via the `TokenIssuanceRedemption` (Prompt 303/304) and `CrossChainBridgeVault` (Prompt 319/323) smart contracts.
- **HSM-Authorized Burn Transaction:** Upon successful completion of Phases 1 through 4, the coordinator constructs an EIP-1559 transaction signed by a dedicated FIPS 140-2 Level 3 HSM Relayer key invoking `burnSyntheticTokens(address userContractAddress, uint256 assetId, uint256 amount, bytes32 reorgProofHash)`.
- **Proof-of-Reserve Merkle Root Update:** The transaction emits `SyntheticTokensBurnedOnReorg(bytes32 indexed sagaId, uint256 assetId, uint256 burnedAmount, bytes32 merkleRoot)` which is picked up by the Proof-of-Reserve Registry (Prompt 327) to maintain strict 1:1 custody backing parity.
- **Zero-PII On-Chain Policy:** All contract parameters contain strictly pseudonymous account hashes, internal asset identifiers, monetary integer amounts, and cryptographic proof digests. No user PAN, legal name, email, or IP address is ever transmitted to the blockchain.

## Saga State Machine & Distributed Transaction Architecture

```
                  +-----------------------------------+
                  |   REORG_DETECTED_EVENT_INGRESS   |
                  +-----------------+-----------------+
                                    |
                                    v
                     +-----------------------------+
                     | [Phase 1] ACCOUNT_FREEZE    |
                     | Lock: MARGIN_RECOVERY_LOCKED|
                     +--------------+--------------+
                                    |
                                    v
                     +-----------------------------+
                     | [Phase 2] ORDER_CANCELLATION|
                     | Purge active CLOB/SL orders |
                     +--------------+--------------+
                                    |
                                    v
                     +-----------------------------+
                     | [Phase 3] PRIORITY_LIQUIDATE|
                     | Forced SPAN position close  |
                     +--------------+--------------+
                                    |
                                    v
                     +-----------------------------+
                     | [Phase 4] DELTA_HEDGE_UNWIND|
                     | Unwind CEX/DEX hedge book   |
                     +--------------+--------------+
                                    |
                                    v
                     +-----------------------------+
                     | [Phase 5] BESU_TOKEN_BURN   |
                     | On-chain burn via HSM       |
                     +--------------+--------------+
                                    |
                                    v
                     +-----------------------------+
                     | SAGA_COMPLETED / SGF_REVERT |
                     | Deficit routed to SGF       |
                     +-----------------------------+
```

### Phase Transition Table

| Phase | Phase Name | Primary Action | Target Service | Success Condition | Compensation on Fatal Failure |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Phase 1** | `ACCOUNT_FREEZE` | Set account status to `MARGIN_RECOVERY_LOCKED` | Redis / PostgreSQL / Risk Engine | Bitflag set in Redis and DB row locked | Alert SecOps, lock platform globally for account |
| **Phase 2** | `ORDER_CANCELLATION` | Cancel all active resting and algorithmic orders | Matching Engine / Advanced Order Engine | Open order count = 0 in matching memory | Re-attempt cancel via emergency drain API |
| **Phase 3** | `PRIORITY_AUTO_LIQUIDATION`| Liquidate all open futures/options positions | SPAN Risk Engine / Perpetuals Engine | Net open contracts = 0, PnL realized | Route residual deficit to SGF Waterfall (Prompt 230) |
| **Phase 4** | `DELTA_HEDGE_UNWINDING` | Execute market unwinds for off-platform hedges | Cross-Chain Collateral Router (CEX/DEX) | Off-platform delta neutral (= 0) | Settle at prevailing market, claim slippage from SGF |
| **Phase 5** | `BESU_TOKEN_BURN` | Submit burn transaction to Besu contract | Hyperledger Besu via Relayer HSM | Transaction mined in QBFT block with receipt | Escalate to Validator Governance Multisig (Prompt 307) |

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Workspace:** Create `services/clawback-saga-coordinator` in Go and embedded computational library `crates/clawback-core` in Rust with strict linting and module structure.
2. **Define Protocol Buffer Specifications:** Author `proto/growww/clawback/v1/clawback_saga.proto` declaring gRPC methods for triggering manual invalidations, querying active Saga progress, and simulating clawback impacts.
3. **Design PostgreSQL Database Schema:** Write migration scripts creating `saga_instances`, `saga_phase_executions`, `saga_compensations`, `reorg_invalidation_events`, and `clawback_ledger_journals` with foreign keys and index constraints.
4. **Implement Reorg Event Ingress Consumer:** Build high-throughput Kafka consumer in Go subscribing to `bridge.reorg.detected.v1` and `bridge.deposit.invalidated.v1` with dead-letter queue (DLQ) support and schema validation.
5. **Implement Distributed Redlock & Concurrency Guard:** Integrate Redis 7.2 Redlock manager to enforce a mutex across `lock:saga:reorg:{chain_id}:{reorg_id}`, ensuring only one coordinator executes the Saga for any given reorg event.
6. **Implement Phase 1 (Account Freeze):** Build atomic executor writing `MARGIN_RECOVERY_LOCKED` state to Redis cache (`account:status:override:{user_id}`) and PostgreSQL `accounts` table, publishing `account.frozen.v1` to Kafka.
7. **Implement Phase 2 (Mass Order Cancellation):** Build synchronous gRPC client dispatching batch order cancel commands to Order Matching Engine (Prompt 205) and Advanced Order Engine (Prompt 226), verifying matching memory eviction.
8. **Implement Phase 3 (Priority Auto-Liquidation):** Construct liquidation trigger client invoking SPAN Risk Engine (Prompt 241) and Perpetuals Engine (Prompt 240) in emergency liquidation mode, capturing realized shortfall.
9. **Implement Phase 4 (Delta Hedge Unwinding):** Build hedge release dispatcher instructing Cross-Chain Router (Prompt 238) to execute off-platform market hedge closeouts on external CEXs (Binance, OKX, Deribit) and DEX liquidity pools.
10. **Implement Phase 5 (Besu Token Burn Dispatcher):** Build blockchain relayer client interfacing with Hyperledger Besu JSON-RPC, signing burn transactions via FIPS 140-2 Level 3 HSM, and validating transaction receipts.
11. **Implement SGF Deficit Escalation Handler:** Develop automated bridge routing net unrecoverable losses to Settlement Guarantee Fund Default Waterfall (Prompt 230) whenever liquidated user equity is insufficient.
12. **Implement Double-Entry Ledger Journal Generator:** Formulate balanced journal postings for Wallet & Ledger Service (Prompt 203) crediting/debiting synthetic liability, bridge custody, and SGF default asset accounts.
13. **Implement Cryptographic Proof & Merkle Attestation Engine:** Implement Rust Merkle proof generator in `crates/clawback-core` generating SHA-256 state transition root hashes for audit verification.
14. **Configure Prometheus Observability & Alerting:** Instrument Prometheus metrics (`clawback_sagas_initiated_total`, `clawback_phase_duration_ms`, `clawback_deficit_inr_total`, `clawback_tokens_burned_total`) and PagerDuty alert rules.
15. **Implement Comprehensive End-to-End Simulation Test Suite:** Construct end-to-end integration tests with simulated deep reorg vectors across Bitcoin, Ethereum, and Solana deposits, verifying zero invariant violations.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/clawback/v1/clawback_saga.proto`)
```protobuf
syntax = "proto3";

package growww.clawback.v1;

option go_package = "growww/clawback/v1;clawbackv1";

service ClawbackSagaService {
  rpc TriggerReorgClawback (TriggerClawbackRequest) returns (TriggerClawbackResponse);
  rpc GetSagaStatus (GetSagaStatusRequest) returns (GetSagaStatusResponse);
  rpc ListActiveSagas (ListActiveSagasRequest) returns (ListActiveSagasResponse);
  rpc RetryFailedPhase (RetryFailedPhaseRequest) returns (RetryFailedPhaseResponse);
  rpc EscalateToSGF (EscalateToSGFRequest) returns (EscalateToSGFResponse);
}

enum ChainIdentifier {
  CHAIN_IDENTIFIER_UNSPECIFIED = 0;
  CHAIN_IDENTIFIER_BITCOIN = 1;
  CHAIN_IDENTIFIER_ETHEREUM = 2;
  CHAIN_IDENTIFIER_SOLANA = 3;
  CHAIN_IDENTIFIER_ARBITRUM = 4;
  CHAIN_IDENTIFIER_POLYGON = 5;
}

enum SagaStatus {
  SAGA_STATUS_UNSPECIFIED = 0;
  SAGA_STATUS_INITIATED = 1;
  SAGA_STATUS_IN_PROGRESS = 2;
  SAGA_STATUS_COMPLETED_SOLVENT = 3;
  SAGA_STATUS_COMPLETED_DEFICIT_ESCALATED = 4;
  SAGA_STATUS_COMPENSATING = 5;
  SAGA_STATUS_FAILED_MANUAL_INTERVENTION = 6;
}

enum SagaPhase {
  SAGA_PHASE_UNSPECIFIED = 0;
  SAGA_PHASE_1_ACCOUNT_FREEZE = 1;
  SAGA_PHASE_2_ORDER_CANCELLATION = 2;
  SAGA_PHASE_3_PRIORITY_LIQUIDATION = 3;
  SAGA_PHASE_4_DELTA_HEDGE_UNWIND = 4;
  SAGA_PHASE_5_BESU_TOKEN_BURN = 5;
}

enum PhaseExecutionStatus {
  PHASE_STATUS_UNSPECIFIED = 0;
  PHASE_STATUS_PENDING = 1;
  PHASE_STATUS_EXECUTING = 2;
  PHASE_STATUS_SUCCEEDED = 3;
  PHASE_STATUS_FAILED_RETRYABLE = 4;
  PHASE_STATUS_FAILED_FATAL = 5;
  PHASE_STATUS_COMPENSATED = 6;
}

message TriggerClawbackRequest {
  string reorg_event_id = 1;
  ChainIdentifier source_chain = 2;
  int64 reorg_depth_blocks = 3;
  string orphaned_block_hash = 4;
  string canonical_block_hash = 5;
  string invalidated_tx_hash = 6;
  string affected_user_id = 7;
  string asset_symbol = 8;
  string invalidated_amount = 9; // Fixed-point numeric string (18 decimals)
  string reason_code = 10;
}

message TriggerClawbackResponse {
  string saga_id = 1;
  SagaStatus status = 2;
  int64 initiated_at_unix_ns = 3;
  string idempotency_key = 4;
}

message PhaseDetail {
  SagaPhase phase = 1;
  PhaseExecutionStatus status = 2;
  int64 started_at_unix_ns = 3;
  int64 completed_at_unix_ns = 4;
  int32 retry_count = 5;
  string error_message = 6;
  string result_payload_json = 7;
}

message GetSagaStatusRequest {
  string saga_id = 1;
}

message GetSagaStatusResponse {
  string saga_id = 1;
  string reorg_event_id = 2;
  string user_id = 3;
  SagaStatus status = 4;
  SagaPhase current_phase = 5;
  repeated PhaseDetail phase_history = 6;
  string total_invalidated_amount = 7;
  string liquidated_equity_recovered_inr = 8;
  string remaining_deficit_inr = 9;
  string besu_burn_tx_hash = 10;
  string merkle_proof_root = 11;
  int64 created_at_unix_ns = 12;
  int64 updated_at_unix_ns = 13;
}

message ListActiveSagasRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListActiveSagasResponse {
  repeated GetSagaStatusResponse active_sagas = 1;
  string next_page_token = 2;
}

message RetryFailedPhaseRequest {
  string saga_id = 1;
  SagaPhase phase = 2;
  bool force_override = 3;
  string override_reason = 4;
}

message RetryFailedPhaseResponse {
  string saga_id = 1;
  SagaPhase phase = 2;
  PhaseExecutionStatus new_status = 3;
}

message EscalateToSGFRequest {
  string saga_id = 1;
  string deficit_amount_inr = 2;
  string rationale = 3;
}

message EscalateToSGFResponse {
  string saga_id = 1;
  string sgf_claim_id = 2;
  bool claim_accepted = 3;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Reorg Invalidation Ingress Events Table
CREATE TABLE reorg_invalidation_events (
    reorg_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_chain VARCHAR(32) NOT NULL, -- BITCOIN, ETHEREUM, SOLANA, ARBITRUM, POLYGON
    reorg_depth_blocks INTEGER NOT NULL,
    orphaned_block_hash VARCHAR(66) NOT NULL,
    canonical_block_hash VARCHAR(66) NOT NULL,
    invalidated_tx_hash VARCHAR(66) NOT NULL,
    user_id UUID NOT NULL,
    asset_symbol VARCHAR(16) NOT NULL,
    invalidated_amount NUMERIC(36, 18) NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_reorg_chain_tx UNIQUE (source_chain, invalidated_tx_hash)
);

-- Master Saga Instances Table
CREATE TABLE saga_instances (
    saga_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reorg_event_id UUID NOT NULL REFERENCES reorg_invalidation_events(reorg_event_id),
    user_id UUID NOT NULL,
    current_phase VARCHAR(40) NOT NULL DEFAULT 'PHASE_1_ACCOUNT_FREEZE',
    saga_status VARCHAR(40) NOT NULL DEFAULT 'INITIATED',
    total_invalidated_amount NUMERIC(36, 18) NOT NULL,
    liquidated_equity_recovered_inr NUMERIC(28, 4) NOT NULL DEFAULT 0.0000,
    remaining_deficit_inr NUMERIC(28, 4) NOT NULL DEFAULT 0.0000,
    is_sgf_escalated BOOLEAN NOT NULL DEFAULT FALSE,
    sgf_claim_id UUID,
    besu_burn_tx_hash VARCHAR(66),
    merkle_proof_root VARCHAR(64),
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Granular Phase Execution Audit Logs
CREATE TABLE saga_phase_executions (
    phase_execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id UUID NOT NULL REFERENCES saga_instances(saga_id) ON DELETE CASCADE,
    phase_name VARCHAR(40) NOT NULL,
    execution_status VARCHAR(40) NOT NULL, -- PENDING, EXECUTING, SUCCEEDED, FAILED_RETRYABLE, FAILED_FATAL
    attempt_number INTEGER NOT NULL DEFAULT 1,
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT uq_saga_phase_attempt UNIQUE (saga_id, phase_name, attempt_number)
);

-- Saga Compensating Actions Log
CREATE TABLE saga_compensations (
    compensation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id UUID NOT NULL REFERENCES saga_instances(saga_id),
    failed_phase VARCHAR(40) NOT NULL,
    compensation_action VARCHAR(64) NOT NULL,
    execution_status VARCHAR(40) NOT NULL,
    compensation_payload JSONB,
    error_message TEXT,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Clawback Double-Entry Accounting Journals
CREATE TABLE clawback_ledger_journals (
    journal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id UUID NOT NULL REFERENCES saga_instances(saga_id),
    user_id UUID NOT NULL,
    debit_account_id UUID NOT NULL,
    credit_account_id UUID NOT NULL,
    amount_inr NUMERIC(28, 4) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'INR',
    journal_type VARCHAR(40) NOT NULL, -- MARGIN_REVERSAL, SYNTHETIC_BURN, SGF_DEFICIT_COVER
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SGF Deficit Escalation Transfers
CREATE TABLE sgf_deficit_transfers (
    transfer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id UUID NOT NULL REFERENCES saga_instances(saga_id),
    user_id UUID NOT NULL,
    claimed_deficit_inr NUMERIC(28, 4) NOT NULL,
    sgf_tranche_applied VARCHAR(40) NOT NULL, -- TRANCHE_1_CORE_RESERVES, TRANCHE_2_MEMBER_ASSESSMENT
    is_approved BOOLEAN NOT NULL DEFAULT TRUE,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance Indices
CREATE INDEX idx_saga_user_status ON saga_instances(user_id, saga_status);
CREATE INDEX idx_saga_created_at ON saga_instances(created_at DESC);
CREATE INDEX idx_phase_exec_saga ON saga_phase_executions(saga_id, phase_name);
CREATE INDEX idx_reorg_user ON reorg_invalidation_events(user_id);
CREATE INDEX idx_clawback_ledger_saga ON clawback_ledger_journals(saga_id);
```

### Kafka Event Schemas

#### Topic: `bridge.reorg.detected.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "BridgeReorgDetectedEvent",
  "type": "object",
  "required": [
    "reorg_event_id",
    "source_chain",
    "reorg_depth_blocks",
    "orphaned_block_hash",
    "canonical_block_hash",
    "invalidated_deposits",
    "detected_at_unix_ns"
  ],
  "properties": {
    "reorg_event_id": { "type": "string", "format": "uuid" },
    "source_chain": { "type": "string", "enum": ["BITCOIN", "ETHEREUM", "SOLANA", "ARBITRUM", "POLYGON"] },
    "reorg_depth_blocks": { "type": "integer", "minimum": 1 },
    "orphaned_block_hash": { "type": "string" },
    "canonical_block_hash": { "type": "string" },
    "invalidated_deposits": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["tx_hash", "user_id", "asset_symbol", "amount"],
        "properties": {
          "tx_hash": { "type": "string" },
          "user_id": { "type": "string", "format": "uuid" },
          "asset_symbol": { "type": "string" },
          "amount": { "type": "string" }
        }
      }
    },
    "detected_at_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `saga.clawback.phase_transition.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SagaClawbackPhaseTransitionEvent",
  "type": "object",
  "required": [
    "saga_id",
    "reorg_event_id",
    "user_id",
    "previous_phase",
    "current_phase",
    "phase_status",
    "attempt_number",
    "timestamp_unix_ns"
  ],
  "properties": {
    "saga_id": { "type": "string", "format": "uuid" },
    "reorg_event_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "previous_phase": { "type": "string" },
    "current_phase": { "type": "string", "enum": ["PHASE_1_ACCOUNT_FREEZE", "PHASE_2_ORDER_CANCELLATION", "PHASE_3_PRIORITY_LIQUIDATION", "PHASE_4_DELTA_HEDGE_UNWIND", "PHASE_5_BESU_TOKEN_BURN", "SAGA_COMPLETED"] },
    "phase_status": { "type": "string", "enum": ["PENDING", "EXECUTING", "SUCCEEDED", "FAILED"] },
    "attempt_number": { "type": "integer" },
    "recovered_equity_inr": { "type": "string" },
    "residual_deficit_inr": { "type": "string" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `saga.clawback.completed.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SagaClawbackCompletedEvent",
  "type": "object",
  "required": [
    "saga_id",
    "reorg_event_id",
    "user_id",
    "final_status",
    "total_invalidated_amount",
    "recovered_equity_inr",
    "remaining_deficit_inr",
    "is_sgf_covered",
    "besu_burn_tx_hash",
    "merkle_proof_root",
    "completed_at_unix_ns"
  ],
  "properties": {
    "saga_id": { "type": "string", "format": "uuid" },
    "reorg_event_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "final_status": { "type": "string", "enum": ["COMPLETED_SOLVENT", "COMPLETED_DEFICIT_ESCALATED"] },
    "total_invalidated_amount": { "type": "string" },
    "recovered_equity_inr": { "type": "string" },
    "remaining_deficit_inr": { "type": "string" },
    "is_sgf_covered": { "type": "boolean" },
    "sgf_claim_id": { "type": "string", "format": "uuid" },
    "besu_burn_tx_hash": { "type": "string" },
    "merkle_proof_root": { "type": "string", "pattern": "^[0-9a-fA-F]{64}$" },
    "completed_at_unix_ns": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **Atomic Single-Flight Isolation:** Distributed Redlock mutex locks on `lock:saga:reorg:{chain_id}:{reorg_id}` prevent race conditions, split-brain executions, or concurrent duplicate clawbacks across horizontally scaled coordinator instances.
- **Fail-Closed Account Quarantine:** Phase 1 applies `MARGIN_RECOVERY_LOCKED` synchronously across Redis cluster bitflags and the primary database. The API Gateway (Prompt 219) and Pre-Trade Risk Engine (Prompt 206) reject all outbound requests with HTTP 423 Locked / gRPC Status `PERMISSION_DENIED` immediately upon flag detection.
- **Strict Execution Order Invariant:** Under no circumstance may Phase 5 (Besu Token Burn) or Phase 4 (Hedge Unwind) execute prior to Phase 2 (Order Cancellation) and Phase 3 (Priority Liquidation). Executing out of order risks unhedged open orders executing against a purged collateral balance.
- **Systemic Deficit Absorption (SEBI & CPMI-IOSCO Standards):** If total liquidation proceeds and remaining user wallet balances are insufficient to offset the invalidated collateral deficit, the coordinator guarantees zero balance sheet deficit by atomically invoking the Settlement Guarantee Fund Default Waterfall (Prompt 230).
- **Zero-PII On-Chain Compliance:** All cryptographic attestation receipts, Merkle leaves, and Hyperledger Besu transaction logs store exclusively SHA-256 hashes, transaction IDs, monetary integers, and UUIDs, maintaining strict compliance with DPDP Act 2023 and GDPR privacy laws.

## Acceptance Criteria
- [ ] Go/Rust coordinator subscribes to `bridge.reorg.detected.v1` and initializes a validated Saga instance in PostgreSQL within 5ms of event receipt.
- [ ] Distributed Redlock prevents duplicate Saga execution across multiple coordinator replicas for identical reorg event IDs.
- [ ] Phase 1 successfully applies `MARGIN_RECOVERY_LOCKED` status in Redis cache and database, blocking subsequent order placement and withdrawal attempts.
- [ ] Phase 2 synchronously invokes Order Matching Engine and Advanced Order Engine to mass-cancel 100% of open resting orders for the target user.
- [ ] Phase 3 triggers priority SPAN liquidation, closing out all open derivative and perpetual positions and posting realized PnL to the audit log.
- [ ] Phase 4 coordinates delta hedge unwinding via Cross-Chain Router, eliminating off-platform inventory and neutralizing market delta.
- [ ] Phase 5 submits an HSM-signed burn transaction to Hyperledger Besu smart contract and validates receipt confirmation in a finalized QBFT block.
- [ ] In the event of residual negative equity, the coordinator programmatically escalates deficit claims to the SGF Default Waterfall (Prompt 230).
- [ ] Balanced double-entry accounting journal entries are posted to the Wallet & Ledger Service for every phase.
- [ ] Cryptographic SHA-256 Merkle proofs are generated for all state transitions and persisted to the immutable audit log.
- [ ] Specification adheres strictly to the 12 mandatory sections with zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `104` (Event Schema & Kafka Standards), Prompt `112` (Idempotency & Exactly-Once Processing), Prompt `203` (Wallet & Ledger Service), Prompt `205` (Order Matching Engine), Prompt `206` (Risk & Margin Checks Service).
- **Parallel Tasks:** Prompt `230` (Settlement Guarantee Fund & Default Waterfall), Prompt `238` (Cross-Chain Collateral & Synthetic FX Router), Prompt `241` (SPAN Portfolio Margin & Liquidation Engine).
- **Downstream Blockers:** Prompt `304` (Token Redemption Smart Contract), Prompt `319` (Institutional Custody Bridge), Prompt `327` (Multi-Chain Proof-of-Reserve Registry Contract).
