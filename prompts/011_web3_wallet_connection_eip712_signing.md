# 011 - Web3 Wallet Connection & EIP-712 Structured Signing Specification

## Purpose
Non-custodial and hybrid Web3 wallet connectivity for retail and institutional traders, enabling seamless authentication, session management, and cryptographic trade authorization via EIP-712 structured data signing without exposing private keys.

This document establishes the authoritative technical blueprint, mathematical models, state machines, API interfaces, failure recovery mechanisms, and regulatory compliance mapping for Web3 Wallet Connection & EIP-712 Structured Signing Specification across the Growww / NBSE platform.

## What You Are Building
A comprehensive Web3 wallet integration and cryptographic signing specification that standardizes multi-chain wallet adapters (MetaMask, Phantom, WalletConnect v2, Coinbase Wallet), EIP-712 structured typed signature verification, replay attack prevention using monotonic nonces, and session-key delegation.

Key capabilities include:
- Scalable, low-latency microservice architecture designed for high-concurrency 24/7 financial operations.
- Deterministic state machine governing all resource lifecycles with absolute transactional consistency.
- Comprehensive gRPC service interfaces and schema definitions ensuring clean polyglot service boundaries.
- Resilient failover, automated circuit breakers, and zero-loss disaster recovery mechanisms.

## Scope Boundaries
- **In Scope:**
- EIP-712 domain separator and structured typed data schemas for Spot Order Placement, Cancellation, and Transfer.
- Replay attack mitigation via account nonce tracking in Redis and Hyperledger Besu state.
- Session key delegation allowing timed high-frequency order placement without repeated wallet popups.
- Wallet connection lifecycle state machine across Flutter mobile and Next.js web terminals.
- **Out of Scope / Handled Elsewhere:**
- MPC-TSS key shard computation.
- Hardware Security Module root key storage.
- Exchange matching engine execution.

## Technology to Use
- Core Technologies: EIP-712 typed signing standard, Web3Modal / Wagmi v2 client libraries, ethers.js / viem, Go go-ethereum crypto package, Redis cluster for nonce caching.
- Performance Standards: Sub-millisecond internal processing, microsecond-precision OpenTelemetry tracing.
- Data Integrity: ACID relational persistence, distributed Redis caching, and immutable ledger anchoring.

## Backend / Infra Touchpoints
- API Gateway, Auth Service, Order Service, Flutter Mobile App, Next.js Web Terminal, Hyperledger Besu JSON-RPC nodes.

## Blockchain Interaction
Hyperledger Besu permissioned EVM ledger with EIP-712 signature verification in smart contracts (DvPAtomicSettlement.sol and OrderRegistry.sol). Validates that trader public addresses match recovered ECDSA secp256k1 signatures.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review system architectural boundaries and regulatory invariants for Web3 Wallet Connection & EIP-712 Structured Signing Specification.
2. Define the core domain entities, data models, and state lifecycle transitions for Web3 Wallet Connection & EIP-712 Structured Signing Specification.
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

package growww.web3.v1;

option go_package = "growww/packages/proto/growww/web3/v1;web3v1";

message OrderSignatureRequest {
  string request_id = 1;
  string user_id = 2;
  string symbol = 3;
  uint64 amount_e8 = 4;
  uint64 price_e8 = 5;
  uint64 timestamp_ms = 6;
  map<string, string> metadata = 7;
}

message OrderSignatureResponse {
  string request_id = 1;
  bool success = 2;
  string execution_id = 3;
  string transaction_hash = 4;
  uint64 block_number = 5;
  string error_message = 6;
  int64 processed_at_ms = 7;
}

message OrderSignatureEvent {
  string event_id = 1;
  string aggregate_id = 2;
  string event_type = 3;
  bytes payload = 4;
  uint64 sequence_number = 5;
  uint64 timestamp_ms = 6;
}

service Web3SignatureService {
  rpc VerifyOrderSignature(OrderSignatureRequest) returns (OrderSignatureResponse);
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
- [ ] Core specification and domain data models for Web3 Wallet Connection & EIP-712 Structured Signing Specification fully documented.
- [ ] Protobuf service contracts and schema interfaces fully specified with zero syntax ambiguity.
- [ ] Hyperledger Besu smart contract and cryptographic verification integration points defined.
- [ ] End-to-end latency budget (p99 < 5ms for internal operations, p99 < 50ms for gateway) satisfied.
- [ ] 8 comprehensive failure modes and recovery actions documented with deterministic recovery paths.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Full compliance mapping to SEBI, RBI, IFSCA, FIU-IND, or DPDP Act 2023 validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** SEBI Cyber Security Framework (CSCRF), FIU-IND Non-Custodial Wallet Guidance, DPDP Act 2023 zero-PII wallet address pseudonymous identity rules.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
