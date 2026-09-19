# order-matching-engine

## Executive Overview
`order-matching-engine` is an institutional-grade microservice written in **Rust** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
Ultra-low-latency in-memory Level 3 orderbook and matching engine with strict Price-Time Priority.

## Interface & Communication Boundaries
- **Primary gRPC Service**: `MatchingEngineService` (defined in `packages/proto/`)
- **Inbound Event Stream**: `market.orders.v1`
- **Outbound Event Stream**: `market.trades.v1`
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
- **Horizontal Autoscaling**: Single-writer active-shadow topology with deterministic failover and core-pinned thread affinity.
- **Latency Target**: p99 < 5ms processing time per message.
- **Telemetry**: OpenTelemetry distributed tracing with Prometheus metrics exported on port `:9090`.
- **Zero-Trust**: Mutual TLS (mTLS) with SPIFFE/SPIRE dynamic X.509 certificate authentication.
