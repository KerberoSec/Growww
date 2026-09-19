# Services Architecture & Microservice Directory

## Executive Overview
The `services/` directory contains the 79 sovereign microservices comprising the backend infrastructure for the Growww / NBSE trading exchange. The system is architected as an ultra-low-latency, event-driven polyglot service mesh (Go, Rust, and Python) designed for both local testability and enterprise horizontal scalability (1 Crore / 10 Million concurrent users).

## Local Development & Testing Architecture
When running in local development mode:
- All services communicate via local gRPC on loopback interfaces (`127.0.0.1`) and local Docker Compose networks.
- In-memory mock brokers and local single-node Kafka clusters provide event streaming.
- Local SQLite / PostgreSQL instances provide ACID persistence without cloud dependencies.
- Local Hyperledger Besu single-node QBFT testnet provides deterministic on-chain settlement.

## Enterprise Scale Production Topology
In production environments:
- Services are deployed as distributed Kubernetes workloads across AWS EKS Multi-AZ (`ap-south-1` Mumbai and `ap-south-2` Hyderabad).
- Zero-Trust mTLS service mesh with SPIFFE/SPIRE dynamic workload authentication.
- Raft consensus sequencers and NVMe append-only Write-Ahead Logs (WAL) for sub-millisecond order processing.
- Direct integration with institutional FIX 5.0 SP2 gateways, NPCI UPI 2.0 banking rails, and NSDL/CDSL depositories.

## Service Domains
1. **Core Matching & Execution**: Sub-millisecond in-memory L3 orderbook and matching engine.
2. **Clearing & DvP Settlement**: Atomic Delivery-versus-Payment blockchain settlement relayer.
3. **Custody & Blockchain Ingress**: Taproot Bitcoin UTXO listener, multichain USDT, and MPC-TSS signers.
4. **Market Data & Feeds**: WebSocket ticker broadcasting, Level 2 depth diffs, and FIX gateway.
5. **Demo / Paper Trading**: Isolated simulation sandbox with virtual testnet faucets and live price mirroring.
6. **Risk & Margin Engines**: Pre-trade risk evaluation, fat-finger collars, and cross-margin calculations.
7. **Compliance & Taxation**: Automated zero on-chain TDS (Section 194S), 30% VDA reporting (Section 115BBH), and FIU-IND AML.
