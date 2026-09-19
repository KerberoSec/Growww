# luld-circuit-breaker-engine

## Executive Overview
`luld-circuit-breaker-engine` is a mission-critical, microsecond-latency market surveillance microservice written in **Rust 1.78+** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
Real-time dynamic volatility dampening and SEBI Limit Up / Limit Down (LULD) price band enforcement:
- Ingests high-frequency tick feeds and evaluates 5-minute rolling Volume Weighted Average Price (VWAP).
- Enforces multi-tier dynamic price bands (Tier 1: ±5%, Tier 2: ±10%, Tier 3: ±20%) with 2x expansion during opening and closing windows.
- Transitions symbols from normal trading to 15-second Limit States and 5-minute Call Auction uncrossing when prices touch bands.
- Enforces the `LiquidationCascadeSuppressionGuard` during market halts to prevent wrongful margin default cascades.

## Interface & Communication Boundaries
- **Inbound Event Stream**: `growww.matching.trade.executed.v1`, `growww.market.ticker.updated.v1`
- **Outbound Event Stream**: `growww.market.luld.state_changed.v1`, `growww.market.trading_halted.v1`
- **Persistence Layer**: Local memory state with RocksDB write-ahead log for zero-latency restarts.

## Local Testing Setup
```bash
cargo test
cargo run --release
```

## High-Scale Production Topology
- **Performance**: Sub-10 microsecond state evaluation per trade tick.
- **Failover**: Paired active-shadow hot standby with Raft-sequenced state replication.
