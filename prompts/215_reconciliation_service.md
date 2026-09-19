# 215 - On-Chain vs Off-Chain Ledger Reconciliation Engine (Go / Rust)

## Purpose
The Reconciliation Engine is the central financial and asset integrity guardian of the Growww architecture. In an institutional investment platform bridging physical depository custody with a permissioned blockchain ledger and off-chain relational databases, any variance between these three records of state represents severe regulatory non-compliance, financial loss, or security compromise.

This service continuously executes deterministic, multi-way reconciliations across: (1) Depository physical Demat shares held at NSDL/CDSL, (2) Off-chain PostgreSQL transactional databases, fiat balances, and portfolio holdings, and (3) On-chain Hyperledger Besu token balances, `SettlementDvP.sol` escrow states, and `ProofOfReserveRegistry.sol` Merkle roots. If any discrepancy exceeds zero tolerance, the engine automatically triggers platform circuit-breakers, freezes affected asset lines, and alerts regulatory compliance officers.

## What You Are Building
A high-assurance, deterministic Go microservice (`services/reconciliation-engine`) delivering:
- **Continuous 3-Way Reconciliation Worker:** Real-time event-driven verification of balance transitions across Depository, SQL, and Blockchain.
- **End-of-Day (EOD) Batch Reconciliation Engine:** High-performance vectorized balance verification using DuckDB and Apache Arrow.
- **Automated Anomaly & Circuit-Breaker Trigger:** Instantaneous trading suspension on specific ISINs or token contracts upon detection of unbacked tokens or fiat imbalances.
- **Proof-of-Reserve Cryptographic Attestation Verifier:** Validates off-chain physical share counts against on-chain Merkle root commitments.
- **Artifacts Delivered:**
 - `services/reconciliation-engine/cmd/server/main.go` - Go service entry point.
 - `services/reconciliation-engine/internal/recon/three_way.go` - Core 3-way reconciliation logic.
 - `services/reconciliation-engine/internal/chain/besu_client.go` - Web3 JSON-RPC client reading on-chain contract states.
 - `services/reconciliation-engine/internal/breaker/circuit_breaker.go` - Circuit breaker and multi-sig alert manager.
 - `proto/growww/reconciliation/v1/reconciliation.proto` - Internal gRPC service definitions.

## Scope Boundaries
- **In Scope:**
 - Continuous streaming reconciliation of trade settlements, token mints, burns, and transfers.
 - T+0 and EOD batch reconciliation of user holdings, fiat ledgers, pool accounts, and token supplies.
 - Cryptographic verification of Merkle proofs for Proof-of-Reserve.
 - Emitting discrepancy alerts, audit reports, and circuit breaker trip signals.
- **Out of Scope / Handled Elsewhere:**
 - Executing depository SFTP uploads/downloads (handled by Prompt 213).
 - Executing smart contract minting/burning (handled by Prompt 303/304).
 - Generating SEBI statutory PDF reports (handled by Prompt 216).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `go-ethereum/ethclient` (for Hyperledger Besu RPC), `pgx/v5`, and `apache/arrow/go/v16` for vectorized batch processing.
- **Justification:** Go provides the memory safety, determinism, high-speed concurrency, and native blockchain library integration needed to compare millions of balance records across SQL databases and Ethereum-compatible Besu nodes within tight sub-second settlement windows.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for transaction ledgers and reconciliation audit runs.
 - Hyperledger Besu JSON-RPC endpoint.
 - Redis 7.2+ for mismatch caching and real-time circuit-breaker flags.
 - Apache Kafka 3.7+ for high-throughput event ingestion.
 - DuckDB Go driver for fast OLAP in-memory EOD reconciliation aggregation.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Core ledger tables (`portfolio_holdings`, `fiat_ledger`, `depository_transactions`).
- **Hyperledger Besu (QBFT):** Direct RPC queries to `DigitalSecurityToken.sol`, `SettlementDvP.sol`, and `ProofOfReserveRegistry.sol`.
- **Redis 7.2:** Real-time variance counters and platform circuit-breaker state keys.
- **Kafka Topics:**
 - Subscribes: `custody.holdings.snapshot`, `trade.settled`, `wallet.balance.updated`, `chain.events.indexed`.
 - Publishes: `reconciliation.verified`, `reconciliation.discrepancy.flagged`, `reconciliation.circuit_breaker.tripped`.
- **Admin Alerting:** PagerDuty / Opsgenie and Slack webhook alerts for immediate engineering escalation.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network with QBFT consensus (2-second block finality).
- **Contracts Read:**
 - `DigitalSecurityToken.sol`: Queries `totalSupply(isin)` and per-account balances.
 - `ProofOfReserveRegistry.sol`: Queries current Merkle root and block attestation timestamps.
 - `SettlementDvP.sol`: Checks pending and locked escrow balances.
- **Verification Formula:**
  $$\sum \text{Demat Physical Shares (NSDL/CDSL)} \equiv \sum \text{Off-Chain Database Holdings} \equiv \text{On-Chain Total Supply (Besu)}$$
- **Zero PII:** Operates purely on anonymized wallet addresses, ISIN codes, transaction hashes, and quantitative balance totals.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/reconciliation-engine` with clean architectural layers (`domain`, `repository`, `adapter`, `usecase`).
2. Define Protobuf definitions in `proto/growww/reconciliation/v1/reconciliation.proto` and generate Go stubs.
3. Configure PostgreSQL schema migrations for `reconciliation_runs`, `discrepancy_records`, and `asset_balance_snapshots`.
4. Implement the Hyperledger Besu Web3 client with contract ABI bindings generated via `abigen`.
5. Build the real-time event streaming reconciler consuming Kafka events (`trade.settled`, `chain.events.indexed`).
6. Implement the 3-Way Asset Reconciliation algorithm comparing depository snapshot records with SQL balances and on-chain contract states.
7. Implement the Fiat-to-Escrow Reconciliation algorithm comparing bank account cash balances against user wallet ledger sums.
8. Build the high-performance EOD batch reconciliation engine using DuckDB / Apache Arrow for large-scale dataset joins.
9. Implement the automated Circuit Breaker mechanism that triggers a Kafka alert and calls admin freeze APIs upon detecting any variance $> 0$.
10. Implement cryptographic verification of Proof-of-Reserve Merkle trees against live depository holding files.
11. Add Prometheus metrics (`reconciliation_variance_gauge`, `recon_duration_seconds`, `discrepancy_count_total`) and health endpoints.
12. Write comprehensive test cases including automated chaos tests simulating depository mismatch, dropped chain events, and out-of-order settlements.

## Interfaces / Contracts

### Protobuf Definition (`reconciliation.proto`)
```protobuf
syntax = "proto3";

package growww.reconciliation.v1;

option go_package = "github.com/growww/services/reconciliation-engine/gen/v1;reconciliationv1";

service ReconciliationService {
  rpc TriggerManualReconciliation (ReconciliationRequest) returns (ReconciliationResponse);
  rpc GetReconciliationStatus (ReconciliationStatusRequest) returns (ReconciliationStatusResponse);
  rpc GetDiscrepancyReport (DiscrepancyReportRequest) returns (DiscrepancyReportResponse);
  rpc ResolveDiscrepancy (ResolveDiscrepancyRequest) returns (ResolveDiscrepancyResponse);
}

message ReconciliationRequest {
  string recon_type = 1; // ASSET_3WAY / FIAT_ESCROW / PROOF_OF_RESERVE
  string target_date = 2; // YYYY-MM-DD
  string isin_filter = 3; // Optional ISIN
}

message ReconciliationResponse {
  string run_id = 1;
  string status = 2; // STARTED / QUEUED
  int64 triggered_at = 3;
}

message ReconciliationStatusRequest {
  string run_id = 1;
}

message ReconciliationStatusResponse {
  string run_id = 1;
  string status = 2; // IN_PROGRESS / COMPLETED / DISCREPANCY_DETECTED / FAILED
  int64 total_records_checked = 3;
  int64 total_discrepancies = 4;
  string execution_time_ms = 5;
  int64 completed_at = 6;
}

message DiscrepancyReportRequest {
  string run_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message DiscrepancyReportResponse {
  string run_id = 1;
  repeated DiscrepancyItem items = 2;
  string next_page_token = 3;
}

message DiscrepancyItem {
  string discrepancy_id = 1;
  string isin = 2;
  string account_identifier = 3;
  string custody_balance = 4;
  string database_balance = 5;
  string on_chain_balance = 6;
  string variance_amount = 7;
  string severity = 8; // CRITICAL / WARNING
  string status = 9; // OPEN / RESOLVED / ESCALATED
  int64 detected_at = 10;
}

message ResolveDiscrepancyRequest {
  string discrepancy_id = 1;
  string resolution_notes = 2;
  string resolved_by_officer_id = 3;
}

message ResolveDiscrepancyResponse {
  string discrepancy_id = 1;
  bool success = 2;
  int64 resolved_at = 3;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE reconciliation_runs (
    run_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recon_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'RUNNING',
    snapshot_date DATE NOT NULL,
    total_records_analyzed BIGINT NOT NULL DEFAULT 0,
    discrepancies_found INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE discrepancy_records (
    discrepancy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES reconciliation_runs(run_id),
    isin VARCHAR(12) NOT NULL,
    identifier VARCHAR(64) NOT NULL,
    custody_qty NUMERIC(28, 8) NOT NULL,
    db_qty NUMERIC(28, 8) NOT NULL,
    chain_qty NUMERIC(28, 8) NOT NULL,
    variance NUMERIC(28, 8) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    circuit_breaker_tripped BOOLEAN NOT NULL DEFAULT FALSE,
    resolution_status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE proof_of_reserve_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date DATE NOT NULL,
    merkle_root VARCHAR(64) NOT NULL,
    on_chain_tx_hash VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Security & Compliance Notes
- **Zero-Tolerance Asset Integrity:** Any detected variance between physical custody and token supply immediately flags a critical security incident and freezes automated minting for the affected ISIN.
- **SEBI Audit Trail Compliance:** Detailed immutable reconciliation runs are retained for a minimum of 8 years in compliance with SEBI record retention mandates.
- **mTLS & Role-Based RPC Access:** Communication with Hyperledger Besu nodes and database replicas is strictly authenticated via mutual TLS with read-only database roles.
- **Maker-Checker Discrepancy Resolution:** Manual resolution of discrepancies requires dual approval from authorized compliance officers.

## Acceptance Criteria
- [ ] Go reconciliation service builds cleanly with zero static analysis warnings.
- [ ] Real-time reconciler processes simulated trade events with sub-100ms verification latency.
- [ ] EOD 3-way batch reconciler verifies 1,000,000 holdings records across SQL, Depository, and Besu in under 30 seconds using DuckDB/Arrow.
- [ ] Circuit breaker immediately trips and publishes alerting events when simulated custody variance is injected.
- [ ] Proof-of-Reserve Merkle root validator correctly detects altered balance leaves.
- [ ] Unit and integration test suite achieves >=85% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 203 (Wallet Service), Prompt 208 (Settlement Service), Prompt 209 (Holdings Service), Prompt 213 (Custody Adapter), Prompt 303 (Token Issuance).
- **Subsequent / Parallel Tasks:** Prompt 216 (Regulatory Reporting Service), Prompt 308 (On-Chain Proof-of-Reserve).
