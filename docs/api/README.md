# API Specifications & Developer Reference (`docs/api`)

Welcome to the centralized API documentation directory for the Growww / NBSE platform. This directory contains complete interface specifications across all supported communication protocols:

---

## Protocol Specifications Catalog

| Document | Format / Standard | Target Audience | Primary Use Case |
| :--- | :--- | :--- | :--- |
| **[REST API Specification](./REST_API_SPECIFICATION.md)** | OpenAPI 3.1 / JSON | Web & Mobile Developers, Third-Party Integrators | Account lifecycle, balance queries, spot orders, demo faucets, crypto transfers |
| **[WebSocket Streaming Protocol](./WEBSOCKET_STREAMING_PROTOCOL.md)** | WSS / JSON & Protobuf | TradingView widgets, Flutter clients, market makers | Sub-millisecond Level-2 depth, live ticks, candlestick series, execution reports |
| **[gRPC Service Definitions](./GRPC_SERVICE_DEFINITIONS.md)** | Protocol Buffers v3 / HTTP/2 | Internal Microservice Engineers | Inter-service RPCs, matching engine sequencer, double-entry wallet balance reservations |
| **[FIX 5.0 SP2 Protocol Specification](./FIX_50SP2_SPECIFICATION.md)** | Binary FIX / Tag-Value TCP | Institutional HFTs, Colocated Prop Desks | High-speed binary order entry, Drop Copy risk feeds, automated market making |

---

## Core Operational Invariants
- **Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)):** All trading protocols report and calculate the exact 0.00% (Zero Fee) (0 bps (0.00% fee at launch) / 0 bps at launch) platform fee.
- **Strict Idempotency:** Every mutating state transition requires a unique client-generated UUIDv4 `Idempotency-Key`.
- **Environment Isolation:** Completely separated endpoints and database states for `TESTNET` (Demo / Paper Trading) and `MAINNET` (Real-Money Production).
