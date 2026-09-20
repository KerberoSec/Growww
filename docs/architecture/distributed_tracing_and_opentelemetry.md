# Growww Platform Distributed Tracing & OpenTelemetry Standards

## 1. Scope & Latency Budget
The Growww / NBSE platform implements sub-millisecond end-to-end distributed tracing using the CNCF OpenTelemetry standard across Flutter apps, Envoy gateways, Go microservices, Rust matching engines, Kafka streams, and Hyperledger Besu private blockchain consensus.

### Latency Budget Hierarchy:
| Component / Boundary | Target p50 | Target p99 | Target p99.9 |
| :--- | :--- | :--- | :--- |
| **API Gateway Ingress (Envoy)** | < 1.0 ms | < 5.0 ms | < 15.0 ms |
| **Pre-Trade Risk Engine (Go)** | < 500 µs | < 2.0 ms | < 5.0 ms |
| **Matching Engine Hot-Path (Rust)** | < 25 µs | < 100 µs | < 250 µs |
| **Kafka Event Streaming Ingestion** | < 2.0 ms | < 10.0 ms | < 25.0 ms |
| **DvP Settlement Relayer (Go)** | < 5.0 ms | < 25.0 ms | < 50.0 ms |
| **Hyperledger Besu QBFT Commit** | < 2.0 s (1 block) | < 4.0 s (2 blocks) | < 6.0 s |

---

## 2. W3C TraceContext Wire Protocol
All network boundaries strictly propagate the W3C `traceparent` header format:
```
traceparent: {version}-{trace_id}-{parent_id}-{trace_flags}
```
- `version`: `00` (Current W3C specification).
- `trace_id`: 32-hex-character (16-byte) globally unique identifier.
- `parent_id`: 16-hex-character (8-byte) span identifier.
- `trace_flags`: `01` (Sampled) or `00` (Not sampled).

---

## 3. Propagation Across Transport Protocols
1. **HTTP/1.1 & HTTP/2 REST:** `traceparent` passed as standard HTTP request header.
2. **gRPC over HTTP/2:** `traceparent` passed inside gRPC request metadata (`metadata.MD`).
3. **Apache Kafka:** `traceparent` injected into Kafka `RecordHeaders` as UTF-8 byte string.
4. **Hyperledger Besu JSON-RPC:** Correlated via `eth.tx_hash` and EIP-712 structured payload digests.

---

## 4. Zero-Allocation Rust Matching Engine Tracing
- Lock-free SPSC ring buffer for high-frequency matching hot-path.
- Avoids heap allocations during order execution.
- Hardware cycle counter (RDTSC) converts to nanoseconds for micro-benchmarks.
