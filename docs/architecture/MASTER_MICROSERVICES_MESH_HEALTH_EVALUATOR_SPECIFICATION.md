# Master Microservices Mesh Health & Latency Budget Evaluator

**Specification ID:** SPEC-MESH-300-ENG
**Document Version:** 1.0.0-PROD-SPEC
**Status:** Approved
**Owner:** Platform Engineering & Site Reliability Group
**Review Cadence:** Quarterly
**Last Review:** September 2026

---

## 1. Purpose & Scope

The Master Microservices Mesh Health & Latency Budget Evaluator is an autonomous control-plane agent responsible for continuous end-to-end health evaluation and latency budget enforcement across all 79 microservices in the Growww/NBSE exchange cluster. It consumes per-hop OpenTelemetry spans from the service mesh, computes rolling p50/p95/p99 latency percentiles per service edge, enforces a sub-5ms aggregate budget for internal operations, and trips circuit breakers when thresholds are breached.

**In Scope:**
- Core state machine governing evaluator lifecycle: INITIALIZING, HEALTHY, DEGRADED, HALTED, RECOVERING.
- Per-service latency budget accounting and rolling HDR histogram aggregation.
- gRPC service interface and Protobuf schema for programmatic health queries.
- Kafka event streaming integration for span ingestion and alert publication.
- Redis Enterprise caching for hot-path latency percentile reads.
- Hyperledger Besu integration for immutable health event anchoring.
- Failure modes, circuit breaker logic, and dead-letter queue recovery.

**Out of Scope:**
- End-user mobile theme styling (handled in Category 5).
- Physical server hardware procurement (handled in Category 8).

---

## 2. Domain Data Models

### 2.1 ServiceEdge

Represents a directed latency-bearing link between two microservices.

```
ServiceEdge {
    source_service_id  string       // Canonical service name (e.g. "matching-engine")
    target_service_id  string       // Downstream service (e.g. "settlement-relayer")
    edge_key           string       // sha256(source + ">" + target), hex-encoded
    latency_budget_us  uint32       // Allocated latency budget in microseconds
    protocol           Protocol     // GRPC | KAFKA | REDIS | HTTP2
    criticality        Criticality  // CRITICAL | HIGH | NORMAL | LOW
}
```

### 2.2 LatencyObservation

Immutable record produced from an OpenTelemetry span once it exits the mesh.

```
LatencyObservation {
    trace_id           string    // W3C TraceContext trace-id
    span_id            string    // W3C TraceContext span-id
    edge_key           string    // Matches ServiceEdge.edge_key
    observed_us        uint64    // Observed one-way latency in microseconds
    wall_clock_ns      uint64    // UTC wall-clock nanoseconds at span end
    status             SpanStatus // OK | ERROR | TIMEOUT
    service_version    string    // Semantic version of the emitting service
}
```

### 2.3 LatencyBudgetReport

Snapshot produced by the evaluator for a specific service edge over a rolling window.

```
LatencyBudgetReport {
    edge_key           string
    window_start_ns    uint64
    window_end_ns      uint64
    p50_us             uint64
    p95_us             uint64
    p99_us             uint64
    p999_us            uint64
    sample_count       uint64
    budget_us          uint32
    budget_utilization float64  // p99_us / budget_us; values > 1.0 are budget breaches
    health_state       HealthState
}
```

### 2.4 MeshHealthSnapshot

Aggregate view across all 79 services, published every 100ms to Redis and Kafka.

```
MeshHealthSnapshot {
    snapshot_id        string         // UUIDv7
    generated_at_ns    uint64
    cluster_health     ClusterHealth  // GREEN | YELLOW | RED | BLACK
    degraded_edges     []string       // edge_keys currently in DEGRADED state
    halted_edges       []string       // edge_keys currently in HALTED state
    total_edges        uint32         // 79 services => up to 79*78/2 directed edges tracked
    p99_budget_breach_count uint32    // edges with budget_utilization > 1.0 in current window
    besu_anchor_tx     string         // Besu transaction hash of last anchored snapshot
}
```

---

## 3. State Machine

The evaluator maintains per-edge health state. Transitions are driven by rolling window p99 measurements evaluated every 100ms.

```
                       p99 < 80% budget
  +--HEALTHY <---------------------------------+
  |     |                                     |
  |     | p99 >= 80% budget                   |
  |     v                                     |
  | DEGRADED ----> p99 < 80% budget ----> RECOVERING
  |     |                                     ^
  |     | p99 >= 100% budget                  |
  |     v                                     |
  | HALTED -----> manual resume or TTL -------+
  |                 expiry (300s)
  |
  | On INITIALIZING: collect 10s warm-up window before first evaluation.
```

State transition invariants:
- HEALTHY -> DEGRADED requires 3 consecutive windows at >= 80% budget utilization.
- DEGRADED -> HALTED requires 2 consecutive windows at >= 100% budget utilization.
- HALTED -> RECOVERING requires explicit operator resume via gRPC or automatic 300-second TTL.
- RECOVERING -> HEALTHY requires 5 consecutive clean windows (p99 < 80% budget).

---

## 4. Latency Budget Allocation

The aggregate p99 latency budget for any synchronous critical path traversal is 5,000 microseconds (5ms). Budget is partitioned statically across service hops according to criticality tier:

| Tier | Budget per Hop (us) | Services |
| :--- | :--- | :--- |
| CRITICAL | 200 | Matching Engine, Settlement Relayer, Risk Engine |
| HIGH | 500 | Order Gateway, KYC Gate, Identity Registry |
| NORMAL | 1,000 | Market Data Aggregator, Audit Logger, Event Indexer |
| LOW | 2,000 | Reporting Service, Notification Service |

Total critical path budget example for order-to-ack:

```
Order Gateway (HIGH) + Risk Engine (CRITICAL) + Matching Engine (CRITICAL) + Settlement Relayer (CRITICAL)
= 500 + 200 + 200 + 200 = 1,100 us p99 internal budget
Remaining slack: 5,000 - 1,100 = 3,900 us reserved for serialization, queueing, and network jitter.
```

HDR histogram buckets use 1-microsecond resolution with a max trackable value of 10,000,000 microseconds (10 seconds), stored per edge in Redis with a 60-second rolling window sliding by 1-second steps.

---

## 5. gRPC Service Interface

```protobuf
syntax = "proto3";

package growww.master.v1;

option go_package = "growww/packages/proto/growww/master/v1;masterv1";

message MastermicroservicesmeshhealthevaluatorRequest {
  string request_id = 1;
  string entity_id = 2;
  uint64 amount_e8 = 3;
  uint64 timestamp_ms = 4;
  map<string, string> metadata = 5;
}

message MastermicroservicesmeshhealthevaluatorResponse {
  string request_id = 1;
  bool success = 2;
  string transaction_hash = 3;
  uint64 block_number = 4;
  string error_message = 5;
}

service MastermicroservicesmeshhealthevaluatorService {
  rpc Execute(MastermicroservicesmeshhealthevaluatorRequest) returns (MastermicroservicesmeshhealthevaluatorResponse);
}

// Extended interfaces for direct health query.

message GetEdgeHealthRequest {
  string edge_key = 1;
}

message GetEdgeHealthResponse {
  string edge_key = 1;
  string health_state = 2;
  uint64 p99_us = 3;
  uint32 budget_us = 4;
  double budget_utilization = 5;
}

message GetMeshSnapshotRequest {}

message GetMeshSnapshotResponse {
  string snapshot_id = 1;
  string cluster_health = 2;
  repeated string degraded_edges = 3;
  repeated string halted_edges = 4;
  uint32 p99_budget_breach_count = 5;
  string besu_anchor_tx = 6;
}

service MeshHealthService {
  rpc GetEdgeHealth(GetEdgeHealthRequest) returns (GetEdgeHealthResponse);
  rpc GetMeshSnapshot(GetMeshSnapshotRequest) returns (GetMeshSnapshotResponse);
  rpc ResumeEdge(GetEdgeHealthRequest) returns (GetEdgeHealthResponse);
}
```

---

## 6. Kafka Integration

### 6.1 Span Ingestion Topic

| Property | Value |
| :--- | :--- |
| Topic | `mesh.otel.spans.raw` |
| Partitions | 79 (one per logical service source) |
| Retention | 1 hour |
| Compression | zstd |
| Key | `edge_key` (ensures per-edge ordering) |
| Value Schema | `LatencyObservation` (Protobuf, schema registry ID enforced) |

The evaluator runs a Kafka consumer group `mesh-health-evaluator-cg` with `auto.offset.reset=latest` and processes spans in batches of up to 1,000 records per poll with a max poll interval of 100ms.

### 6.2 Alert Publication Topic

When a health state transition occurs, the evaluator publishes to `mesh.health.alerts`:

| Field | Value |
| :--- | :--- |
| Topic | `mesh.health.alerts` |
| Partitions | 8 |
| Key | `edge_key` |
| Value | JSON: `{ "edge_key", "previous_state", "new_state", "p99_us", "budget_us", "timestamp_ns" }` |

Consumers of `mesh.health.alerts` include: Alertmanager bridge, circuit-breaker controller, WORM audit logger.

### 6.3 Dead-Letter Queue

Spans that cannot be parsed or attributed to a known edge are routed to `mesh.otel.spans.dlq` with the original bytes and a `dlq_reason` header. A separate reconciliation worker re-processes the DLQ every 30 seconds.

---

## 7. Redis Enterprise Caching

Hot-path percentile reads are served from Redis Enterprise with the following key schema:

```
mesh:edge:{edge_key}:p99          -> uint64 (microseconds, TTL 10s)
mesh:edge:{edge_key}:histogram    -> HdrHistogram binary blob (TTL 65s)
mesh:snapshot:latest              -> MeshHealthSnapshot JSON (TTL 200ms)
mesh:edge:{edge_key}:state        -> HealthState string (TTL none, updated on transition)
```

All writes use Redis pipelines with `MULTI/EXEC` to guarantee atomic histogram + p99 updates. Reads for the gRPC `GetEdgeHealth` path must complete within 500 microseconds (Redis call included), enforced by a 500us deadline on the Go context.

---

## 8. Hyperledger Besu Blockchain Integration

Every 10 seconds, the evaluator submits a Keccak-256 hash of the current `MeshHealthSnapshot` to the Besu permissioned ledger via a pre-deployed `MeshHealthAnchor` smart contract. This provides an immutable, non-repudiable audit trail for health state history.

**Contract Interface:**

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IMeshHealthAnchor {
    event HealthSnapshotAnchored(
        bytes32 indexed snapshotHash,
        uint64 generatedAtNs,
        uint8 clusterHealth,
        uint32 breachCount
    );

    function anchorSnapshot(
        bytes32 snapshotHash,
        uint64 generatedAtNs,
        uint8 clusterHealth,
        uint32 breachCount
    ) external;
}
```

**Security Invariant:** The Besu node submitting anchor transactions uses an HSM-backed key (YubiHSM 2 or CloudHSM). No private keys are present on disk or in environment variables. Anchoring is non-blocking; the evaluator does not wait for Besu confirmation before continuing its evaluation loop.

**Consensus & Block Finality:** Hyperledger Besu QBFT with 2-second block intervals and deterministic finality. The anchor transaction hash is stored in `MeshHealthSnapshot.besu_anchor_tx` for cross-referencing.

---

## 9. OpenTelemetry Observability

Each evaluation cycle produces the following spans and metrics:

**Spans (microsecond precision):**
- `mesh_evaluator.ingest_batch` - duration of Kafka batch processing.
- `mesh_evaluator.compute_percentiles` - duration of HDR histogram percentile computation.
- `mesh_evaluator.redis_write` - duration of atomic Redis pipeline write.
- `mesh_evaluator.besu_anchor` - duration of Besu transaction submission (async, does not block evaluation loop).

**Prometheus Metrics:**
| Metric | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `mesh_evaluator_edge_p99_us` | Gauge | `edge_key`, `source`, `target` | Current p99 latency in microseconds |
| `mesh_evaluator_budget_utilization` | Gauge | `edge_key` | p99 / budget ratio |
| `mesh_evaluator_state_transitions_total` | Counter | `edge_key`, `from_state`, `to_state` | State transition count |
| `mesh_evaluator_breached_edges` | Gauge | - | Count of edges with utilization > 1.0 |
| `mesh_evaluator_kafka_lag` | Gauge | `partition` | Consumer group lag per partition |
| `mesh_evaluator_evaluation_cycle_us` | Histogram | - | Duration of each 100ms evaluation cycle |

---

## 10. Failure Modes & Recovery

| Failure Scenario | Trigger Condition | System Behavior & Mitigation |
| :--- | :--- | :--- |
| Network Partition | Distributed nodes lose connectivity | Circuit breaker trips; requests queue with backoff; fallback to secondary nodes within 500ms. |
| Invalid Payload | Malformed or spoofed cryptographic payload | Rejection with INVALID_ARGUMENT; client flagged in rate limiter; event logged to WORM audit trail. |
| Replay Attack | Duplicate transaction re-submitted | Monotonic nonce check rejects duplicate; Redis lock prevents race conditions; dropped silently. |
| Upstream Timeout | Upstream service exceeds latency budget | Falls back to cached state if safe, or returns DEADLINE_EXCEEDED; client retries with idempotent token. |
| State Mismatch | In-memory cache diverges from database | Periodic reconciliation detects hash mismatch; locks resource; rolls back state from immutable WAL. |
| Resource Spike | Pod memory/CPU reaches 90% threshold | Autoscaler scales replicas; non-critical telemetry sampled; rate limiter sheds low-priority requests. |
| Blockchain Stall | Consensus node halt or delay | Status stays PENDING; gas escalator increases replacement fee; alert triggered if block interval > 6s. |
| Edge Discrepancy | Arithmetic overflow or corner-case occurs | Safe transaction abort; all balance locks released immediately; core dump captured for offline replay. |

---

## 11. Regulatory & Compliance Mapping

| Requirement | Framework | Implementation |
| :--- | :--- | :--- |
| 7-year WORM audit trail | SEBI LODR 2015 Reg. 25, RBI IT Framework | Every `MeshHealthSnapshot` hash is anchored to Besu. Raw Kafka events are archived to immutable S3-compatible object storage with WORM policy. |
| Cryptographic non-repudiation | FIU-IND PMLA 2002 | Besu QBFT consensus provides deterministic block finality; HSM-signed transactions provide non-repudiation of anchor submissions. |
| Zero on-chain PII | DPDP Act 2023, RBI Data Localisation | Only Keccak-256 hashes of health snapshots are anchored; no personal data, IP addresses, or service credentials appear on-chain. |
| Automated compliance reporting | SEBI Circular SEBI/HO/MRD/DRMNP | The `GetMeshSnapshot` gRPC endpoint provides a real-time compliance-grade view consumable by the automated reporting pipeline. |
| Operational resilience | IFSCA BFSR Circular 2023 | Evaluator is deployed with N+2 redundancy across availability zones; circuit breaker and DLQ ensure no span data loss during partial failures. |
