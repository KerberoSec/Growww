# Canonical Protocol Buffers & gRPC Schemas

## Purpose & Scope
`packages/proto` contains the canonical Protocol Buffers v3 (`.proto`) service definitions and message schemas governing all inter-service gRPC communication, Kafka event serialization, and client WebSocket streaming.

## Package Domains
- `growww/market/v1/`: Market data feeds, Level 2/3 depth diffs, OHLCV bars, and ticker updates.
- `growww/order/v1/`: Order placement, cancellation, conditional triggers, and execution reports.
- `growww/settlement/v1/`: On-chain DvP batch settlement payloads and Besu relayer events.
- `growww/demo/v1/`: Simulated order execution, virtual faucet claims, and sandbox balance schemas.
- `growww/wallet/v1/`: Multichain crypto deposits, fiat INR payment webhooks, and withdrawal requests.
- `growww/common/v1/`: Shared types (UUID, Timestamp, Money, BigInt, SecurityIdentifier).
