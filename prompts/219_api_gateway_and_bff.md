# 219 - API Gateway & Backend-for-Frontend (BFF) Layer (Go / Envoy / REST / GraphQL)

## Purpose
The API Gateway & Backend-for-Frontend (BFF) layer serves as the single secure entry point and reverse proxy for all client traffic across multi-platform client applications (Flutter mobile/desktop and Next.js web applications). It insulates internal gRPC microservices from the public internet, terminates TLS, authenticates user JWTs and session tokens, validates device security integrity, and aggregates backend microservice responses into unified, client-optimized data models.

By consolidating authentication, request filtering, response compression, client-specific data shaping, and WebSocket multiplexing at the edge, this layer guarantees low latency for trading interfaces while upholding zero-trust internal network boundaries.

## What You Are Building
A high-performance Go microservice and Envoy-based gateway architecture (`services/api-gateway` & `services/bff-service`) providing:
- **Edge API Gateway (Envoy / Go):** Reverse proxy, mTLS edge termination, OAuth2/JWT verification, and request routing.
- **Client-Specific BFF Layer (Go / GraphQL / REST):** Aggregates complex queries (e.g., Portfolio Dashboard combining Holdings, Live Ticker, Realized P&L, and Proof-of-Reserve) into single round-trips.
- **WebSocket Streaming Multiplexer:** Multiplexes live order book ticks (Prompt 207), notification events (Prompt 211), and order execution states over a single client connection.
- **Client Security & Device Binding Validator:** Enforces device fingerprints, app signature validation, and jailbreak/root detection flags.
- **Artifacts Delivered:**
 - `services/api-gateway/main.go` - Go gateway router and middleware engine.
 - `services/api-gateway/config/envoy.yaml` - Envoy proxy edge configuration.
 - `services/bff-service/cmd/server/main.go` - BFF aggregation server.
 - `services/bff-service/internal/aggregator/portfolio.go` - Multi-service dashboard aggregator.
 - `proto/growww/gateway/v1/gateway.proto` - Public gateway contract definitions.

## Scope Boundaries
- **In Scope:**
 - TLS 1.3 termination, CORS management, and security HTTP headers.
 - JWT / OIDC validation and token revocation checking with Redis.
 - Protocol translation from external REST/GraphQL/WebSocket to internal gRPC.
 - Client payload compression (Brotli/Gzip) and payload trimming.
 - WebSocket session multiplexing and heartbeat management.
- **Out of Scope / Handled Elsewhere:**
 - Fine-grained business logic execution (handled by specialized microservices, Prompts 201-218).
 - Algorithmic rate limiting and DDoS scoring (handled by Prompt 220).
 - User authentication credential verification (handled by Prompt 201).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `grpc-gateway/v2`, `gin-gonic/gin` (or `valyala/fasthttp`), and `gorilla/websocket` with Envoy Proxy as the edge L7 ingress.
- **Justification:** Go combined with Envoy provides extreme connection throughput, sub-millisecond proxy routing overhead, minimal CPU/memory footprint under 100,000+ concurrent WebSocket connections, and native gRPC protocol-buffer serialization support.
- **Dependencies & Libraries:**
 - Redis 7.2+ for JWT revocation checks, user session cache, and rate limit sync.
 - Envoy Proxy 1.30+ for edge routing and mTLS service mesh integration.
 - `grpc-gateway/v2` for generating OpenAPI 3.1 and REST proxies from Protobuf.
 - `open-telemetry/opentelemetry-go` for end-to-end trace propagation.

## Backend / Infra Touchpoints
- **Redis 7.2 Cluster:** Token revocation blacklists, active user sessions, device verification nonces.
- **Internal gRPC Microservices Mesh:** User Service, Order Service, Market Data Service, Portfolio Service, Fee Engine, etc.
- **Kafka Topics:** Subscribes to real-time client notifications and system maintenance alerts.
- **Hyperledger Besu (QBFT):** Routes read-only public Proof-of-Reserve queries to Besu JSON-RPC read replicas; zero transaction signing or private keys reside on the gateway.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium ledger running QBFT consensus.
- **Read-Only Ledger Gateway:** Exposes public read endpoints `/api/v1/blockchain/proof-of-reserve` and `/api/v1/blockchain/settlement-status/{tx_hash}` that forward requests to off-chain event indexers (Prompt 309) and Besu read nodes.
- **Zero Key Custody:** The API Gateway holds no blockchain private keys and does not submit signed transactions. All write actions are submitted as gRPC commands to internal services which interface with HSM relayers.
- **Zero PII Leakage:** Public blockchain query responses contain strictly anonymized Merkle proofs, smart contract addresses, and transaction hashes.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/api-gateway` and `services/bff-service` with standardized configuration loaders.
2. Configure Envoy Proxy edge configuration with TLS 1.3, strict cipher suites, and mTLS internal upstream clusters.
3. Implement JWT / OIDC validation middleware verifying RSA256 / Ed25519 token signatures against Redis revocation lists.
4. Implement Device Integrity & Header Verification middleware validating App-Version, Device-ID, and anti-tamper signatures.
5. Set up `grpc-gateway` to automatically generate RESTful endpoints from internal Protobuf service definitions.
6. Build the BFF GraphQL / REST aggregator combining user profile, wallet balance, active orders, and portfolio holdings into unified payloads.
7. Implement the WebSocket Gateway multiplexer managing persistent client connections, heartbeats, and topic subscriptions.
8. Connect WebSocket stream to internal Redis Pub/Sub and Kafka consumers for live order execution and market tick delivery.
9. Implement response caching and compression middleware (Brotli / Gzip) with ETag validation.
10. Integrate OpenTelemetry trace context injection propagating `traceparent` headers to all downstream gRPC calls.
11. Add Prometheus metrics (`gateway_requests_total`, `gateway_latency_histogram`, `ws_active_connections_count`).
12. Write load and security tests using `k6` and OWASP ZAP verifying performance under 50,000 concurrent connections and resilience to API attacks.

## Interfaces / Contracts

### Protobuf Definition (`gateway.proto`)
```protobuf
syntax = "proto3";

package growww.gateway.v1;

import "google/api/annotations.proto";

option go_package = "github.com/growww/services/api-gateway/gen/v1;gatewayv1";

service ClientBFFService {
  rpc GetDashboardOverview (DashboardRequest) returns (DashboardResponse) {
    option (google.api.http) = {
      get: "/api/v1/dashboard/overview"
    };
  }

  rpc GetProofOfReserveSummary (PoRSummaryRequest) returns (PoRSummaryResponse) {
    option (google.api.http) = {
      get: "/api/v1/public/proof-of-reserve"
    };
  }
}

message DashboardRequest {
  string user_id = 1;
}

message DashboardResponse {
  string user_id = 1;
  string inr_wallet_balance = 2;
  string total_portfolio_value_inr = 3;
  string total_unrealized_pnl_inr = 4;
  string total_realized_pnl_inr = 5;
  repeated HoldingItem top_holdings = 6;
  repeated ActiveOrderItem pending_orders = 7;
  int64 server_time = 8;
}

message HoldingItem {
  string isin = 1;
  string symbol = 2;
  string company_name = 3;
  string fractional_units = 4;
  string average_cost_price = 5;
  string current_market_price = 6;
  string current_value = 7;
}

message ActiveOrderItem {
  string order_id = 1;
  string symbol = 2;
  string order_type = 3;
  string side = 4;
  string quantity = 5;
  string price = 6;
  string status = 7;
}

message PoRSummaryRequest {}

message PoRSummaryResponse {
  string latest_merkle_root = 1;
  uint64 block_number = 2;
  string on_chain_tx_hash = 3;
  string verified_custody_timestamp = 4;
  int64 total_backed_securities_count = 5;
}
```

## Security & Compliance Notes
- **OWASP API Security Top 10 Defenses:** Strict protection against Broken Object Level Authorization (BOLA), Broken Authentication, Mass Assignment, and Unrestricted Resource Consumption.
- **Zero-Trust Internal Mesh:** All egress traffic from API Gateway to backend microservices requires mutual TLS (mTLS) with cryptographically validated service identities (SPIFFE/SPIRE).
- **Client Security Headers:** Enforces strict CSP, HSTS (`max-age=63072000; includeSubDomains; preload`), `X-Frame-Options: DENY`, and `X-Content-Type-Options: nosniff`.
- **Sensitive Data Masking:** Gateway automatically sanitizes and masks sensitive PII (PAN, bank account numbers, MPINs) in transit logs.

## Acceptance Criteria
- [ ] API Gateway and BFF services build and execute cleanly inside Docker containers.
- [ ] JWT authentication interceptor validates valid tokens and rejects revoked/expired tokens within <1ms.
- [ ] BFF dashboard aggregator concurrently fetches data from User, Wallet, Order, and Holdings services, reducing client round-trips to 1.
- [ ] WebSocket multiplexer sustains 50,000+ concurrent connections with <5ms message broadcast latency.
- [ ] Public Proof-of-Reserve endpoint successfully streams on-chain attestation data without exposing internal infrastructure.
- [ ] End-to-end integration and load tests pass with >=85% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design Standards), Prompt 105 (Auth Architecture), Prompt 201 (User Service), Prompt 207 (Market Data).
- **Subsequent / Parallel Tasks:** Prompt 220 (Rate Limiting), Prompt 512 (Flutter App Shell), Prompt 601 (Web App).
