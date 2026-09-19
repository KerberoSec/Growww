# Shared Domain Types & Financial Models

## Purpose & Scope
`packages/domain_types` defines the canonical, polyglot financial domain types to prevent floating-point inaccuracies and semantic drift across the platform.

## Key Types
- `AmountE8`: 64-bit integer representing currency units with 8 fixed decimal places.
- `OrderStatus`: Deterministic state machine: `PENDING`, `OPEN`, `PARTIALLY_FILLED`, `FILLED`, `CANCELLED`, `REJECTED`.
- `TradingPair`: Immutable currency pair identifier (e.g. `BTC/USDT`, `ETH/USDT`, `BTC/eINR`).
- `ExecutionReport`: Comprehensive post-match clearing model with price, fill size, fee, and liquidity flag.
