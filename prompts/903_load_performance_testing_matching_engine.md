# 903 - Load & Performance Testing Plan (Matching Engine & Gateway Focus)

## Purpose
Establishes a rigorous, distributed load and stress testing framework capable of validating the Growww platform under extreme market conditions (e.g., market open volatility, high-frequency limit order bursts, and flash crash simulations). The primary objective is to verify that the Rust matching engine maintains sub-millisecond (p99 < 1ms) matching latency, the API Gateway sustains 50,000 requests/second with p99 < 50ms, and the on-chain Hyperledger Besu DvP settlement batching pipeline maintains resilience without transaction starvation.

## What You Are Building
A distributed, cloud-native performance testing suite (`tests/load/`):
- Distributed k6 load testing scenarios configured via k6 Operator on Kubernetes to simulate 50,000+ orders/sec sustained throughput and 100,000 concurrent WebSocket connections.
- Realistic market order stream generators replicating historical NSE/BSE tick-by-tick order flow distributions (limit orders, market orders, cancellations, modifications).
- Smart contract DvP batch settlement benchmarking harness to measure on-chain TPS, block gas utilization, and relayer queue latency on Hyperledger Besu.
- Automated Grafana load-test dashboards integrating Prometheus metrics, OpenTelemetry distributed traces, and JVM/Rust memory/CPU profiles.

## Scope Boundaries
- **In Scope:**
 - Order entry API ingestion throughput at the API Gateway and BFF layer (50,000 req/s).
 - In-memory order book matching engine latency and throughput (single-symbol and multi-symbol scaling).
 - Market data WebSocket streaming fan-out to 100,000 concurrent subscribers.
 - Asynchronous DvP settlement queue throughput and Besu blockchain mempool pressure testing.
 - Stress testing to identify system breaking points, CPU/memory leak thresholds, and backpressure behavior.
- **Out of Scope / Handled Elsewhere:**
 - Microservice unit/integration testing (Prompt 901).
 - Multi-tier UI user journey automation (Prompt 902).
 - Infrastructure fault injection and container kill testing (Prompt 904).

## Technology to Use
- **Primary Load Generator:** k6 (Go-based native binary executing modern JavaScript/TypeScript test scenarios) - chosen for its minimal memory footprint, asynchronous I/O, native WebSocket and gRPC protocol support, and cloud-native Kubernetes orchestration via the `k6-operator`.
- **Secondary Multi-Step User Simulator:** Locust (Python-based distributed worker framework) for complex multi-step stateful user journey simulations.
- **Metrics & Observability:** Prometheus for high-frequency time-series metrics collection, Grafana for real-time visualization, and Grafana Tempo / Pyroscope for continuous distributed tracing and CPU flamegraph analysis during load runs.

*Justification:* k6 can generate tens of thousands of requests per second per node with predictable resource overhead, making it ideal for 50,000 req/s financial benchmarks where client-side measurement jitter must be eliminated.

## Backend / Infra Touchpoints
- Dedicated Performance Testing Kubernetes Cluster (isolated node pools with c6i.4xlarge compute instances).
- API Gateway (Envoy / Traefik) and BFF instances under horizontal pod autoscaling (HPA).
- Rust Matching Engine service (`services/matching-engine`) connected via high-speed gRPC / Aeron.
- Redis 7+ Cluster (Cluster mode, 6 nodes) for fast order state caching and WebSocket pub/sub.
- Apache Kafka cluster (6 brokers, 16 partitions per order topic) for trade output ingestion.
- Hyperledger Besu QBFT 4-node validator cluster with dedicated transaction relayer nodes.

## Blockchain Interaction
- Benchmarks atomic DvP settlement on Hyperledger Besu by stress-testing the Go settlement service and `SettlementDvP.sol` contract.
- Evaluates relayer transaction batching strategies (e.g., submitting 100 DvP settlement proofs per aggregated on-chain transaction) to ensure blockchain gas limits and 2-second block intervals are not overwhelmed during 50,000 req/s market bursts.
- Analyzes mempool queuing, nonces synchronization across parallel relayer HSM keys, and gas price estimation stability under sustained transaction load.

## Step-by-Step Build Instructions
1. Initialize the `tests/load/` repository structure with subdirectories for `scenarios/`, `data_generators/`, `k8s_manifests/`, and `dashboards/`.
2. Develop synthetic market data generators that produce mathematically realistic order book distributions matching Indian equity trading characteristics (price clustering around tick sizes, bid-ask spread distributions).
3. Implement the baseline order placement k6 test scenario (`tests/load/scenarios/order_placement_burst.js`) targeting the REST/gRPC API Gateway.
4. Implement the high-concurrency WebSocket subscription k6 scenario (`tests/load/scenarios/market_data_fanout.js`) maintaining 100,000 active persistent connections streaming order book deltas.
5. Deploy the `k6-operator` into the dedicated Kubernetes test cluster to orchestrate 20+ distributed load-generator pods.
6. Configure Prometheus scrape jobs for the Rust matching engine custom metrics (`engine_order_latency_microseconds`, `engine_matches_per_second`, `engine_orderbook_depth`).
7. Execute ramp-up baseline tests (from 1,000 to 10,000 req/s) to establish baseline system resource utilization curves.
8. Execute peak load tests (ramping up to 50,000 req/s for 30 minutes sustained) to measure p50, p90, p99, and p99.9 latencies.
9. Execute soak tests (15,000 req/s continuous load over 12 hours) to identify memory leaks, file descriptor leaks, or Redis connection pool degradation.
10. Execute spike tests (instant jump from 2,000 req/s to 60,000 req/s within 5 seconds) to validate rate limiter throttling, circuit breakers, and HPA auto-scaling responsiveness.
11. Run smart contract DvP batch settlement load tests measuring maximum on-chain settlement throughput and verifying transaction finality on Hyperledger Besu.
12. Generate automated performance benchmark reports and integrate automated regression performance gates into the release pipeline.

## Interfaces / Contracts
```javascript
// k6 Load Test Scenario Spec: 50,000 req/s Order Placement Burst
// File: tests/load/scenarios/order_placement_burst.js

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    order_burst: {
      executor: 'ramping-arrival-rate',
      startRate: 5000,
      timeUnit: '1s',
      preAllocatedVUs: 2000,
      maxVUs: 10000,
      stages: [
        { duration: '2m', target: 10000 },  // Warm-up to 10k req/s
        { duration: '5m', target: 50000 },  // Ramp to 50k req/s
        { duration: '15m', target: 50000 }, // Sustain 50k req/s peak load
        { duration: '3m', target: 0 },      // Cool-down
      ],
    },
  },
  thresholds: {
    'http_req_duration{type:order_place}': ['p(95)<30', 'p(99)<50'], // API Gateway p99 < 50ms
    'http_req_failed': ['rate<0.001'],                               // Error rate < 0.1%
    'matching_engine_p99_latency': ['value<1000'],                   // In-engine p99 < 1ms
  },
};

export default function () {
  const payload = JSON.stringify({
    isin: 'INE002A01018',
    order_type: 'LIMIT',
    side: Math.random() > 0.5 ? 'BUY' : 'SELL',
    quantity: (Math.random() * 5 + 0.1).toFixed(4),
    price: (2800.0 + (Math.random() * 20 - 10)).toFixed(2),
    idempotency_key: `load-${__VU}-${__ITER}-${Date.now()}`
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.TEST_AUTH_TOKEN}`,
    },
    tags: { type: 'order_place' },
  };

  const res = http.post(`${__ENV.API_BASE_URL}/api/v1/orders`, payload, params);
  check(res, {
    'order accepted': (r) => r.status === 201 || r.status === 200,
  });
}
```

## Security & Compliance Notes
- All load tests must execute in isolated staging/performance environments strictly firewalled from production networks.
- Synthetic load traffic must have mock bank/custodian headers so external partner networks (NPCI, NSDL) are never accidentally contacted.
- Rate limiting and WAF rules must be evaluated during load tests to ensure malicious traffic patterns are dropped while legitimate high-volume customer orders are not throttled.

## Acceptance Criteria
- [ ] API Gateway reliably handles sustained 50,000 req/s order ingestion with p99 latency < 50ms and error rate < 0.05%.
- [ ] Rust matching engine executes order matching with p99 in-engine latency < 1ms under sustained peak volume.
- [ ] Market data WebSocket service successfully broadcasts price updates to 100,000 concurrent client connections with message broadcast latency < 100ms.
- [ ] On-chain settlement batching relayer processes DvP settlements without dropping transactions or exhausting Besu gas limits.
- [ ] Soak test runs for 12 continuous hours with zero memory leaks or unhandled service restarts.

## Suggested Order / Dependencies
- **Prerequisites:** 010 (NFRs), 204 (Order Service), 205 (Matching Engine), 207 (Market Data), 219 (API Gateway), 806 (Observability).
- **Parallel Tasks:** 904 (Chaos Engineering), 905 (Security Testing).
