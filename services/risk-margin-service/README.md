# risk-margin-service

## Executive Overview
`risk-margin-service` is the central pre-trade and real-time portfolio risk evaluation service written in **Go 1.22+** (with high-performance Redis Lua scripts and Rust SIMD math kernels) within the Growww / NBSE trading platform architecture.

## Primary Responsibility
Pre-trade and post-trade risk management across cash equities, crypto, and derivatives:
- Validates pre-trade margin sufficiency and capital adequacy in sub-500 microseconds.
- Enforces the SEBI 50:50 cash-to-collateral ratio and statutory collateral haircut schedules.
- Monitors real-time Account Health Ratios ($HR = E / \text{TMM}$) and triggers automated margin cure timers and Auto-Deleveraging (ADL) waterfalls.
- Calculates portfolio-level SPAN margins and Options Value-at-Risk (VaR).

## Interface & Communication Boundaries
- **Primary gRPC Service**: `PreTradeRiskService`, `PortfolioMarginService`
- **Inbound Event Stream**: `growww.engine.matches.v1`, `growww.market.price_ticks.v1`
- **Outbound Event Stream**: `growww.risk.margin_call.v1`, `growww.risk.liquidation_triggered.v1`
- **Persistence Layer**: TigerBeetle Ledger (holds) + Redis Cluster (real-time balances) + PostgreSQL (audit trail).

## Local Testing Setup
```bash
go test ./...
go run main.go
```

## High-Scale Production Topology
- **Latency Budget**: p99 < 500 microseconds for pre-trade balance checks.
- **Autoscaling**: Scaled across trading partition clusters with Redis Enterprise clustering.
