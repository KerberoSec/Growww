# 050 - Consolidated Tax Contract Note (Form 16A / Form 26AS) Generator

## Purpose
Deliver institutional-grade financial transparency by automatically generating legally binding, digitally signed Contract Notes for every spot crypto trade, complete with brokerage, exchange fees, GST, and zero on-chain TDS breakdowns.

This document establishes the authoritative technical blueprint, mathematical models, state machines, API interfaces, failure recovery mechanisms, and regulatory compliance mapping for Consolidated Tax Contract Note (Form 16A / Form 26AS) Generator across the Growww / NBSE platform.

## What You Are Building
A high-throughput document generation service that aggregates daily trade executions, computes statutory levies (18% GST on brokerage, zero on-chain TDS, stamp duty), and produces tamper-proof, digitally signed PDF contract notes.

Key capabilities include:
- Scalable, low-latency microservice architecture designed for high-concurrency 24/7 financial operations.
- Deterministic state machine governing all resource lifecycles with absolute transactional consistency.
- Comprehensive gRPC service interfaces and schema definitions ensuring clean polyglot service boundaries.
- Resilient failover, automated circuit breakers, and zero-loss disaster recovery mechanisms.

## Scope Boundaries
- **In Scope:**
- Automated generation of daily consolidated Contract Notes formatted in accordance with SEBI broker standards.
- Detailed financial breakdown: Gross Trade Value, Brokerage, Exchange Turnover Fee, 18% GST, zero on-chain TDS, and Net Consideration.
- PKI digital signature applied to every PDF using exchange Class 3 digital signature certificate (DSC).
- Direct user download via mobile/web apps and automated encrypted email dispatch within 24 hours of trade execution.
- **Out of Scope / Handled Elsewhere:**
- Real-time WebSocket market feeds.
- High-frequency limit order matching.

## Technology to Use
- Core Technologies: Go document engine, UniPDF / wkhtmltopdf, PKCS#7 digital signature module, AWS S3 / MinIO WORM storage.
- Performance Standards: Sub-millisecond internal processing, microsecond-precision OpenTelemetry tracing.
- Data Integrity: ACID relational persistence, distributed Redis caching, and immutable ledger anchoring.

## Backend / Infra Touchpoints
- Contract Note Service, Trade Ledger, Tax Service, Email Dispatcher, Mobile Client, Web Portal.

## Blockchain Interaction
Embeds the Hyperledger Besu settlement transaction hash and block number on every generated contract note for absolute provenance.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review system architectural boundaries and regulatory invariants for Consolidated Tax Contract Note (Form 16A / Form 26AS) Generator.
2. Define the core domain entities, data models, and state lifecycle transitions for Consolidated Tax Contract Note (Form 16A / Form 26AS) Generator.
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

package growww.tax.contractnote.v1;

option go_package = "growww/packages/proto/growww/tax/contractnote/v1;contractnotev1";

message GenerateContractNoteRequest {
  string request_id = 1;
  string user_id = 2;
  string symbol = 3;
  uint64 amount_e8 = 4;
  uint64 price_e8 = 5;
  uint64 timestamp_ms = 6;
  map<string, string> metadata = 7;
}

message GenerateContractNoteResponse {
  string request_id = 1;
  bool success = 2;
  string execution_id = 3;
  string transaction_hash = 4;
  uint64 block_number = 5;
  string error_message = 6;
  int64 processed_at_ms = 7;
}

message GenerateContractNoteEvent {
  string event_id = 1;
  string aggregate_id = 2;
  string event_type = 3;
  bytes payload = 4;
  uint64 sequence_number = 5;
  uint64 timestamp_ms = 6;
}

service ContractNoteService {
  rpc GenerateContractNote(GenerateContractNoteRequest) returns (GenerateContractNoteResponse);
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
- [ ] Core specification and domain data models for Consolidated Tax Contract Note (Form 16A / Form 26AS) Generator fully documented.
- [ ] Protobuf service contracts and schema interfaces fully specified with zero syntax ambiguity.
- [ ] Hyperledger Besu smart contract and cryptographic verification integration points defined.
- [ ] End-to-end latency budget (p99 < 5ms for internal operations, p99 < 50ms for gateway) satisfied.
- [ ] 8 comprehensive failure modes and recovery actions documented with deterministic recovery paths.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Full compliance mapping to SEBI, RBI, IFSCA, FIU-IND, or DPDP Act 2023 validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** SEBI circular on Electronic Contract Notes (ECN), Information Technology Act 2000 digital signature legal validity.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
