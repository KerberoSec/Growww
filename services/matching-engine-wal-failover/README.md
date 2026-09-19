# matching-engine-wal-failover

## Executive Overview
`matching-engine-wal-failover` is an institutional-grade microservice written in **Rust** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
NVMe append-only Write-Ahead Log (WAL) replicator and sub-500ms Raft failover manager.

## Interface & Communication Boundaries
- **Primary gRPC Service**: `WALFailoverService` (defined in `packages/proto/`)
- **Inbound Event Stream**: `wal.events.v1`
- **Outbound Event Stream**: `sync.events.v1`
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
- **Horizontal Autoscaling**: Active-shadow NVMe mmap WAL replication with sub-500ms failover.
- **Latency Target**: p99 < 5ms processing time per message.
- **Telemetry**: OpenTelemetry distributed tracing with Prometheus metrics exported on port `:9090`.
- **Zero-Trust**: Mutual TLS (mTLS) with SPIFFE/SPIRE dynamic X.509 certificate authentication.
