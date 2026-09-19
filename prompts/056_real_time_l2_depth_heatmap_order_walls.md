# 056 - Real-Time Level 2 / Level 3 Depth Heatmap & Order Wall Visualization

## Purpose
Empower crypto traders with instant visual insight into market depth, liquidity clusters, and institutional order walls through an animated Level 2 orderbook depth heatmap and cumulative volume visualizer.

This document establishes the authoritative technical blueprint, mathematical models, state machines, API interfaces, failure recovery mechanisms, and regulatory compliance mapping for Real-Time Level 2 / Level 3 Depth Heatmap & Order Wall Visualization across the Growww / NBSE platform.

## What You Are Building
A high-performance orderbook depth component that renders real-time bid and ask ladders, horizontal volume bars with opacity gradients, cumulative depth charts, and visual highlight flags for whale order walls (>5 BTC).

Key capabilities include:
- Scalable, low-latency microservice architecture designed for high-concurrency 24/7 financial operations.
- Deterministic state machine governing all resource lifecycles with absolute transactional consistency.
- Comprehensive gRPC service interfaces and schema definitions ensuring clean polyglot service boundaries.
- Resilient failover, automated circuit breakers, and zero-loss disaster recovery mechanisms.

## Scope Boundaries
- **In Scope:**
- Level 2 orderbook display: top 20 bids and asks updating at 20 FPS with zero layout shift or flickering.
- Visual depth heatmap bars: horizontal bars behind price rows reflecting relative volume at each price level.
- Whale wall detection: distinct visual badge highlighting price levels with unusually large resting liquidity.
- Cumulative market depth visualization (green vs red visual mountain chart) with interactive tooltip inspection.
- **Out of Scope / Handled Elsewhere:**
- Database query optimization.
- Legal dispute arbitration.

## Technology to Use
- Core Technologies: Flutter CustomPainter / Canvas, React Canvas / SVG rendering, WebSocket conflated book stream, WebGL.
- Performance Standards: Sub-millisecond internal processing, microsecond-precision OpenTelemetry tracing.
- Data Integrity: ACID relational persistence, distributed Redis caching, and immutable ledger anchoring.

## Backend / Infra Touchpoints
- Orderbook Screen Widget, Market Data Service, WebSocket Streaming Gateway, Mobile & Web Frontends.

## Blockchain Interaction
Displays on-chain settlement liquidity status for resting limit orders registered on Hyperledger Besu.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review system architectural boundaries and regulatory invariants for Real-Time Level 2 / Level 3 Depth Heatmap & Order Wall Visualization.
2. Define the core domain entities, data models, and state lifecycle transitions for Real-Time Level 2 / Level 3 Depth Heatmap & Order Wall Visualization.
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

package growww.ui.depth.v1;

option go_package = "growww/packages/proto/growww/ui/depth/v1;depthv1";

message OrderbookHeatmapRequest {
  string request_id = 1;
  string user_id = 2;
  string symbol = 3;
  uint64 amount_e8 = 4;
  uint64 price_e8 = 5;
  uint64 timestamp_ms = 6;
  map<string, string> metadata = 7;
}

message OrderbookHeatmapResponse {
  string request_id = 1;
  bool success = 2;
  string execution_id = 3;
  string transaction_hash = 4;
  uint64 block_number = 5;
  string error_message = 6;
  int64 processed_at_ms = 7;
}

message OrderbookHeatmapEvent {
  string event_id = 1;
  string aggregate_id = 2;
  string event_type = 3;
  bytes payload = 4;
  uint64 sequence_number = 5;
  uint64 timestamp_ms = 6;
}

service DepthVisualizerService {
  rpc GetOrderbookHeatmap(OrderbookHeatmapRequest) returns (OrderbookHeatmapResponse);
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
- [ ] Core specification and domain data models for Real-Time Level 2 / Level 3 Depth Heatmap & Order Wall Visualization fully documented.
- [ ] Protobuf service contracts and schema interfaces fully specified with zero syntax ambiguity.
- [ ] Hyperledger Besu smart contract and cryptographic verification integration points defined.
- [ ] End-to-end latency budget (p99 < 5ms for internal operations, p99 < 50ms for gateway) satisfied.
- [ ] 8 comprehensive failure modes and recovery actions documented with deterministic recovery paths.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Full compliance mapping to SEBI, RBI, IFSCA, FIU-IND, or DPDP Act 2023 validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** IOSCO Market Transparency Standards, SEBI guidelines on fair display of orderbook depth.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
