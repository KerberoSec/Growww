# automated-liquidity-provisioning-engine

## Executive Overview
`automated-liquidity-provisioning-engine` is an institutional-grade microservice written in **Rust** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
Algorithmic market-making daemon providing narrow spreads and continuous liquidity.

## Interface & Communication Boundaries
- **Primary gRPC Service**: `AutoLiquidityEngineService` (defined in `packages/proto/`)
- **Inbound Event Stream**: `market.depth.v1`
- **Outbound Event Stream**: `quote.orders.v1`
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
