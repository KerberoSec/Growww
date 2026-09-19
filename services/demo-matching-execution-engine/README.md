# demo-matching-execution-engine

## Executive Overview
`demo-matching-execution-engine` is an institutional-grade microservice written in **Rust** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
Isolated in-memory matching engine executing virtual paper trades against mirrored live prices.

## Interface & Communication Boundaries
- **Primary gRPC Service**: `DemoMatchingExecutionService` (defined in `packages/proto/`)
- **Inbound Event Stream**: `demo.orders.v1`
- **Outbound Event Stream**: `demo.executions.v1`
- **Persistence Layer**: Distributed Redis Cluster for in-memory caching and PostgreSQL for ACID durability.
- **Blockchain Touchpoint**: Integrates with the Hyperledger Besu enterprise settlement layer where applicable.

## Local Testing Setup
```bash
# Run unit tests locally
cargo test

# Run service in local development mode
cargo run
```

## High-Scale Production Topology
- **Horizontal Autoscaling**: Kubernetes HPA dynamically scales from 3 to 30 replicas based on CPU/memory utilization and Kafka consumer lag.
- **Latency Target**: p99 < 5ms processing time per message.
- **Telemetry**: OpenTelemetry distributed tracing with Prometheus metrics exported on port `:9090`.
- **Zero-Trust**: Mutual TLS (mTLS) with SPIFFE/SPIRE dynamic X.509 certificate authentication.
