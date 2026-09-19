# 206 - Pre-Trade Risk & Margin Engine (Go / Redis)

## Purpose
The Pre-Trade Risk & Margin Engine is the frontline protective barrier guarding Growww's financial and regulatory integrity. Before any order reaches the matching engine, this service conducts deterministic, sub-millisecond pre-trade risk validations. In compliance with SEBI Master Circulars on Market Surveillance, Risk Management Frameworks, and Algorithmic Trading Controls, the engine enforces dynamic price bands (circuit breakers), maximum single-order quantity and value thresholds (fat-finger protection), user-level daily gross turnover limits, and market-wide volatility circuit breakers.

By filtering out erroneous, fraudulent, or non-compliant orders before execution, the Risk Engine protects investors from catastrophic trading errors and guarantees exchange solvency.

## What You Are Building
A ultra-low-latency Go microservice (`services/risk-engine`) backed by Redis in-memory evaluation scripts (Lua) and local memory caches. Concrete deliverables include:
- High-speed gRPC server exposing `EvaluateOrderRisk` responding in $< 1.5\text{ms}$ at $p99$.
- Redis Lua script suite executing atomic sliding-window order velocity counters, cumulative turnover calculations, and exposure limits.
- Circuit breaker state manager tracking dynamic daily operating bands ($\pm 10\%$, $\pm 20\%$) and market-wide index halts.
- Fat-finger and outlier detection algorithms preventing orders with prices significantly deviating from the national best bid/offer (NBBO).
- Kafka consumer ingesting live market ticker feeds and trade fills to maintain real-time exposure states.
- PostgreSQL repository for risk parameter configurations, compliance overrides, and immutable risk breach audit logs.

## Scope Boundaries
- **In Scope:**
 - Pre-trade validation for Limit and Market orders across all asset classes.
 - User-level daily turnover limits, position size caps, and order velocity rate checks.
 - Price band validation against exchange reference and circuit limits.
 - Single order maximum value and volume checks.
 - Dynamic symbol-level and exchange-level trading halt states.
- **Out of Scope / Handled Elsewhere:**
 - User cash wallet reservation and ledger debits (Prompt 203).
 - In-memory order book matching (Prompt 205).
 - Post-trade AML transaction monitoring and suspicious pattern analysis (Prompt 704).
 - Physical share custody verification (Prompt 213).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+. Selected for its fast execution, lightweight concurrent goroutines, and minimal overhead when interacting with Redis via connection pooling.
- **In-Memory Cache & Scripts:** Redis 7.2+ utilizing Redis Enterprise / Cluster with custom Lua scripts for atomic single-roundtrip multi-check execution.
- **Database & Storage:** PostgreSQL 16+ for risk rule persistence and breach records; `pgx/v5` and `sqlc` for type-safe database queries.
- **Inter-Service Communication:** `google.golang.org/grpc` for sub-millisecond RPC evaluations; `segmentio/kafka-go` for market data feeds.

## Backend / Infra Touchpoints
- **Redis 7.2:** Keys `risk:symbol:{isin}:bands`, `risk:user:{user_id}:turnover:daily`, `risk:user:{user_id}:velocity:1m`, `risk:exchange:status`.
- **PostgreSQL 16:** Tables `risk_parameters`, `risk_profiles`, `circuit_breakers`, `risk_breach_audit_log`.
- **Apache Kafka:** Consumes from `matching.trades.v1`, `market.ticker.v1`; publishes alerts to `risk.alerts.v1`.
- **Order Service (Prompt 204):** Calls `EvaluateOrderRisk` prior to order dispatch.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Compliance Registry Rule Enforcement:** Synchronizes with on-chain holding limits codified in `ComplianceRegistry.sol` on Hyperledger Besu (e.g., SEBI Regulatory Sandbox holding limits restricting retail investors to a maximum of ₹1,00,000 in total tokenized assets).
- **On-Chain Address Freezes:** Subscribes to on-chain `ComplianceRegistry.AddressFrozen` events emitted from Hyperledger Besu under QBFT consensus, immediately updating Redis blacklists to block pre-trade orders for sanctioned or suspended addresses within 100ms.
- **Zero On-Chain PII:** The Risk Engine operates purely on `user_id`, `ledger_address`, and `isin`, guaranteeing zero PII disclosure.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/risk-engine` with strict formatting, Makefile, and standard layout.
2. **Define Protobuf Schema:** Create `proto/growww/risk/v1/risk_service.proto` defining `EvaluateOrderRisk`, `UpdateCircuitBreakers`, and `SetRiskProfile`.
3. **Compile Protobuf & gRPC Stubs:** Generate Go code using `protoc`.
4. **Design PostgreSQL Schema:** Write database migrations for `risk_parameters`, `symbol_risk_limits`, and `risk_breach_audit_log`.
5. **Configure Redis Connection Pool & Lua Scripts:** Write optimized Lua scripts for atomic pre-trade evaluation:
 - Check if exchange or symbol is halted.
 - Verify order price within lower/upper circuit band ($[P_{\text{lower}}, P_{\text{upper}}]$).
 - Verify order value $\le \text{MaxSingleOrderValue}$.
 - Atomically increment and verify 1-minute order velocity $\le \text{MaxOrdersPerMinute}$.
 - Verify cumulative daily turnover $\le \text{DailyTurnoverLimit}$.
6. **Implement In-Memory Cache for Static Limits:** Build concurrent in-memory cache using `sync.RWMutex` with 30-second background refresh from PostgreSQL to avoid Redis roundtrips for static symbol limits.
7. **Build Price Band Calculation Module:** Implement calculation engine deriving daily static bands ($\pm 10\%$ from previous day closing price) and dynamic circuit filters based on rolling 15-minute VWAP.
8. **Build Fat-Finger Protection Module:** Implement validation rejecting orders deviating by $> 5\%$ from the current best bid/offer for market orders.
9. **Implement On-Chain Freeze Listener:** Consume Kafka events originating from blockchain indexer (Prompt 309) for on-chain address freeze transactions, updating Redis blacklists instantly.
10. **Implement Kafka Market Data Consumer:** Ingest ticker updates from `market.ticker.v1` to refresh dynamic reference prices in Redis.
11. **Implement gRPC Server Handlers:** Implement `RiskServiceServer` with strict timeout controls (context timeout 5ms).
12. **Implement Risk Breach Auditing:** Asynchronously write all rejected order attempts and risk breaches to `risk_breach_audit_log` and publish to Kafka topic `risk.alerts.v1`.
13. **Configure Prometheus Metrics:** Instrument histograms for `risk_evaluation_duration_seconds` and counters for `risk_rejections_total` categorized by reason code.
14. **Write Comprehensive Benchmark & Integration Tests:** Write Go benchmarks ensuring evaluation latency $\le 1.0\text{ms}$ and integration tests verifying circuit breaker enforcement under multi-threaded concurrency.

## Interfaces / Contracts

### Protobuf Definition (`risk_service.proto`)
```protobuf
syntax = "proto3";

package growww.risk.v1;

option go_package = "growww/risk/v1;riskv1";

service RiskService {
  rpc EvaluateOrderRisk (EvaluateOrderRiskRequest) returns (EvaluateOrderRiskResponse);
  rpc UpdateCircuitBreakers (UpdateCircuitBreakersRequest) returns (UpdateCircuitBreakersResponse);
  rpc GetSymbolRiskStatus (GetSymbolRiskStatusRequest) returns (GetSymbolRiskStatusResponse);
}

enum RiskDecision {
  RISK_DECISION_UNSPECIFIED = 0;
  RISK_DECISION_APPROVED = 1;
  RISK_DECISION_REJECTED = 2;
}

enum RiskRejectReason {
  REJECT_REASON_NONE = 0;
  REJECT_REASON_EXCHANGE_HALTED = 1;
  REJECT_REASON_SYMBOL_HALTED = 2;
  REJECT_REASON_CIRCUIT_BREAKER_VIOLATION = 3;
  REJECT_REASON_MAX_ORDER_VALUE_EXCEEDED = 4;
  REJECT_REASON_MAX_ORDER_QUANTITY_EXCEEDED = 5;
  REJECT_REASON_DAILY_TURNOVER_LIMIT_EXCEEDED = 6;
  REJECT_REASON_ORDER_VELOCITY_RATE_EXCEEDED = 7;
  REJECT_REASON_FAT_FINGER_PRICE_DEVIATION = 8;
  REJECT_REASON_COMPLIANCE_ADDRESS_FROZEN = 9;
}

message EvaluateOrderRiskRequest {
  string order_id = 1;
  string user_id = 2;
  string ledger_address = 3;
  string isin = 4;
  string side = 5; // BUY, SELL
  string order_type = 6; // LIMIT, MARKET
  string price = 7; // Decimal string in INR
  string quantity = 8; // Fractional decimal string
}

message EvaluateOrderRiskResponse {
  string order_id = 1;
  RiskDecision decision = 2;
  RiskRejectReason reject_reason = 3;
  string message = 4;
  int64 evaluated_at_unix_ns = 5;
}

message UpdateCircuitBreakersRequest {
  string isin = 1;
  string lower_circuit_price = 2;
  string upper_circuit_price = 3;
  bool is_halted = 4;
  string reason = 5;
}

message UpdateCircuitBreakersResponse {
  bool success = 1;
}

message GetSymbolRiskStatusRequest {
  string isin = 1;
}

message GetSymbolRiskStatusResponse {
  string isin = 1;
  string lower_circuit_price = 2;
  string upper_circuit_price = 3;
  bool is_halted = 4;
  string max_order_quantity = 5;
  string max_order_value = 6;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE risk_parameters (
    parameter_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_name VARCHAR(50) NOT NULL UNIQUE, -- RETAIL_STANDARD, ACCREDITED_INVESTOR, INSTITUTIONAL
    max_single_order_value NUMERIC(18, 2) NOT NULL DEFAULT 500000.00, -- ₹5,00,000
    max_daily_turnover NUMERIC(18, 2) NOT NULL DEFAULT 2000000.00, -- ₹20,00,000
    max_orders_per_minute INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE symbol_risk_limits (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    lower_circuit_price NUMERIC(18, 4) NOT NULL,
    upper_circuit_price NUMERIC(18, 4) NOT NULL,
    max_order_quantity NUMERIC(18, 6) NOT NULL DEFAULT 10000.000000,
    is_trading_halted BOOLEAN NOT NULL DEFAULT FALSE,
    halt_reason TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE risk_breach_audit_log (
    breach_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    reject_reason VARCHAR(100) NOT NULL,
    attempted_price NUMERIC(18, 4),
    attempted_quantity NUMERIC(18, 6),
    attempted_value NUMERIC(18, 4),
    breach_details JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_breach_user ON risk_breach_audit_log(user_id, occurred_at);
CREATE INDEX idx_risk_breach_isin ON risk_breach_audit_log(isin, occurred_at);
```

## Security & Compliance Notes
- **SEBI Market Surveillance Mandate:** All orders breaching price bands or velocity limits must be rejected with immutable audit records retained for 5 years.
- **Fail-Closed Security Invariant:** If the Risk Engine fails or times out ($> 5\text{ms}$), downstream order routers must default to **REJECT** (fail-closed) to protect system solvency.
- **Zero In-Memory Drift:** User turnover and velocity counters are maintained atomically in Redis with automatic midnight IST key expiration.

## Acceptance Criteria
- [ ] Pre-trade risk engine evaluates all incoming order parameters against configured price bands and turnover limits.
- [ ] Orders with prices outside the lower/upper circuit band are rejected with `REJECT_REASON_CIRCUIT_BREAKER_VIOLATION`.
- [ ] Single orders exceeding maximum order value caps are rejected with `REJECT_REASON_MAX_ORDER_VALUE_EXCEEDED`.
- [ ] Redis Lua evaluation executes in $< 1.0\text{ms}$ per order under 10,000 concurrent evaluations/sec.
- [ ] Trading halt on a symbol instantly blocks all incoming orders for that symbol across all cluster instances.
- [ ] Frozen ledger addresses are immediately blocked from placing orders.
- [ ] Every rejected order creates an immutable entry in `risk_breach_audit_log` and emits a Kafka alert.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Standards), 204 (Order Service), 402 (Redis Patterns), 407 (Master Data Management).
- **Parallel Tasks:** 203 (Wallet Service), 205 (Matching Engine).
- **Downstream Blockers:** 204 (Order Service integration testing).
