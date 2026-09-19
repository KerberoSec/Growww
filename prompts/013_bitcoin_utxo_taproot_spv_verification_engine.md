# 013 - Bitcoin UTXO Ingress, Taproot Scripts & SPV Verification Engine

## Purpose
Provide institutional-grade Bitcoin (BTC) deposit, custody, and settlement rails by implementing native Taproot (BIP-341/342) address derivation, real-time UTXO tracking, and trust-minimized Simplified Payment Verification (SPV) block header validation.

This document establishes the authoritative technical blueprint, mathematical models, state machines, API interfaces, failure recovery mechanisms, and regulatory compliance mapping for Bitcoin UTXO Ingress, Taproot Scripts & SPV Verification Engine across the Growww / NBSE platform.

## What You Are Building
A native Bitcoin UTXO processing engine that tracks on-chain BTC deposits, verifies block headers via cryptographic SPV proofs, enforces 6-block finality for high-value spot balances, and derives deterministic Taproot deposit addresses for every registered trader.

Key capabilities include:
- Scalable, low-latency microservice architecture designed for high-concurrency 24/7 financial operations.
- Deterministic state machine governing all resource lifecycles with absolute transactional consistency.
- Comprehensive gRPC service interfaces and schema definitions ensuring clean polyglot service boundaries.
- Resilient failover, automated circuit breakers, and zero-loss disaster recovery mechanisms.

## Scope Boundaries
- **In Scope:**
- Taproot (BIP-341/342) script-path and key-path address derivation using master public keys (xpub).
- UTXO state indexing, mempool transaction monitoring, and RBF (Replace-By-Fee) detection.
- SPV light-client header chain verification with difficulty adjustment validation.
- Automated credit to off-chain BTC spot balance upon meeting confirmation thresholds.
- **Out of Scope / Handled Elsewhere:**
- EVM token smart contracts.
- Fiat currency banking rails.

## Technology to Use
- Core Technologies: Bitcoin Core JSON-RPC, btcutil / btcsuite (Go), Taproot Schnorr signatures, RocksDB for UTXO indexing, Kafka deposit event topics.
- Performance Standards: Sub-millisecond internal processing, microsecond-precision OpenTelemetry tracing.
- Data Integrity: ACID relational persistence, distributed Redis caching, and immutable ledger anchoring.

## Backend / Infra Touchpoints
- BTC UTXO Ingress Service, Wallet Service, Ledger Service, MPC-TSS Custody Daemon, Core Matching Engine.

## Blockchain Interaction
Bitcoin Mainnet and Testnet4 integration via P2TR addresses, linked to Hyperledger Besu via peg-in proof anchoring smart contracts.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review system architectural boundaries and regulatory invariants for Bitcoin UTXO Ingress, Taproot Scripts & SPV Verification Engine.
2. Define the core domain entities, data models, and state lifecycle transitions for Bitcoin UTXO Ingress, Taproot Scripts & SPV Verification Engine.
3. Establish the primary microservice boundaries, gRPC service definitions, and Protobuf contracts.
4. Design high-performance state transitions, in-memory caches, and database schemas ensuring sub-millisecond execution.
5. Configure the Hyperledger Besu blockchain interface, smart contract interactions, and deterministic event listeners.
6. Implement robust pre-execution validation checks, balance reservations, and anti-tamper security controls.
7. Connect distributed telemetry, OpenTelemetry microsecond tracing, and Prometheus metrics for real-time observability.
8. Establish fallback mechanisms, dead-letter queues, and graceful degradation paths for upstream/downstream service failures.
9. Implement automated reconciliation loops verifying off-chain state consistency against immutable on-chain records.
10. Construct comprehensive unit, integration, and fuzz testing suites achieving >90% code branch coverage.
11. Perform end-to-end chaos engineering and stress test simulations under 10,000 requests per second burst load.
12. Review security posture with the Chief Information Security Officer (CISO) and obtain regulatory compliance sign-off.

## Interfaces / Contracts
```protobuf
syntax = "proto3";

package growww.bitcoin.v1;

option go_package = "growww/packages/proto/growww/bitcoin/v1;bitcoinv1";

message DepositAddressRequest {
  string request_id = 1;
  string user_id = 2;
  string symbol = 3;
  uint64 amount_e8 = 4;
  uint64 price_e8 = 5;
  uint64 timestamp_ms = 6;
  map<string, string> metadata = 7;
}

message DepositAddressResponse {
  string request_id = 1;
  bool success = 2;
  string execution_id = 3;
  string transaction_hash = 4;
  uint64 block_number = 5;
  string error_message = 6;
  int64 processed_at_ms = 7;
}

message DepositAddressEvent {
  string event_id = 1;
  string aggregate_id = 2;
  string event_type = 3;
  bytes payload = 4;
  uint64 sequence_number = 5;
  uint64 timestamp_ms = 6;
}

service BitcoinIngressService {
  rpc GetDepositAddress(DepositAddressRequest) returns (DepositAddressResponse);
}
```

## Failure Modes & Edge Cases
| Failure Scenario | Trigger Condition | System Behavior & Mitigation |
| :--- | :--- | :--- |
| Network Partition / Disconnection | Distributed nodes or external feeds lose connectivity | Automated circuit breaker trips; requests queue with exponential backoff; fallback to secondary failover nodes within 500ms. |
| Invalid Signature / Payload Tampering | Malformed or spoofed cryptographic payload submitted | Immediate rejection with gRPC code INVALID_ARGUMENT; client IP flagged in rate limiter; security event logged to WORM audit trail. |
| Replay Attack Attempt | Attacker re-submits previously executed transaction or signature | Monotonic nonce check rejects duplicate; Redis nonce lock prevents race conditions; transaction dropped silently after alert dispatch. |
| High Latency / Upstream Timeout | Upstream service or blockchain node exceeds latency budget | Execution falls back to cached state if safe, or returns DEADLINE_EXCEEDED; client prompted to retry with idempotent token. |
| State Desynchronization | In-memory cache diverges from database or on-chain truth | Periodic reconciliation job detects hash mismatch; locks affected resource; performs deterministic state rollback from immutable WAL. |
| Hardware / Memory Exhaustion | Pod memory or CPU reaches 90% threshold under burst traffic | Horizontal Pod Autoscaler scales replicas; non-critical telemetry sampled; rate-limiting shedder rejects low-priority requests. |
| Blockchain Re-org or Stall | Consensus node halt or unexpected transaction delay | Transaction status stays PENDING; automated gas escalator increases replacement fee; alert triggered if block generation exceeds 6 seconds. |
| Unhandled Edge Discrepancy | Corner-case market condition or arithmetic overflow occurs | Safe transaction abort; all balance locks released immediately; full core dump captured for deterministic offline replay. |

## Acceptance Criteria
- [ ] Core specification and domain data models for Bitcoin UTXO Ingress, Taproot Scripts & SPV Verification Engine fully documented.
- [ ] Protobuf service contracts and schema interfaces fully specified with zero syntax ambiguity.
- [ ] Hyperledger Besu smart contract and cryptographic verification integration points defined.
- [ ] End-to-end latency budget (p99 < 5ms for internal operations, p99 < 50ms for gateway) satisfied.
- [ ] 8 comprehensive failure modes and recovery actions documented with deterministic recovery paths.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Full compliance mapping to SEBI, RBI, IFSCA, FIU-IND, or DPDP Act 2023 validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** FIU-IND Reporting for Virtual Digital Asset (VDA) deposits, FATF Travel Rule for cross-border UTXO origin screening.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
