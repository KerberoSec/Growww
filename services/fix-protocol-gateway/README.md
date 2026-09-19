# fix-protocol-gateway

## Executive Overview
`fix-protocol-gateway` is an ultra-low-latency, zero-allocation institutional gateway written in **Rust 2021 Edition** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
High-performance ingress and egress gateway processing:
- **FIX 5.0 SP2**: Tag-value engine for `NewOrderSingle` (MsgType=D), `OrderCancelRequest` (MsgType=F), `OrderCancelReplaceRequest` (MsgType=G), and `ExecutionReport` (MsgType=8).
- **OUCH 4.2+**: Fixed-length binary order entry protocol for sub-10 microsecond DMA order routing.
- **ITCH 5.0**: Binary MoldUDP64 multicast market data feed delivering Level-3 tick-by-tick order events.
- **Drop-Copy**: Dedicated regulatory and compliance risk monitoring feeds.
- **Cancel-on-Disconnect (COD)**: Automated heartbeat watchdog cancelling 100% of open resting orders within 50ms of socket termination.

## Interface & Communication Boundaries
- **Primary Transport**: TCP/IP zero-copy with `io_uring` and Linux `epoll` socket pooling.
- **Inbound Event Stream**: `growww.fix.inbound.v1`
- **Outbound Event Stream**: `growww.fix.outbound.v1`
- **Internal Microstructure Bus**: LMAX Disruptor ring buffer into Order Matching Engine.
- **Zero-Fee Institutional Tag**: Tag 136 (`FeeRate`) = `0.0000` at launch.

## Local Testing Setup
```bash
# Run unit and fuzz tests locally
cargo test

# Run gateway in local high-frequency benchmarking mode
cargo run --release
```

## High-Scale Production Topology
- **Deployment**: Core-pinned NUMA execution with dedicated CPU affinity (taskset).
- **Latency Target**: p99 < 10 microseconds end-to-end gateway parsing and order ingestion.
- **Telemetry**: OpenTelemetry microsecond tracing with Prometheus metrics exported on port `:9090`.
- **Zero-Trust**: Mutual TLS (mTLS) with SPIFFE/SPIRE dynamic X.509 certificate authentication.
