# 112 - Idempotency & Exactly-Once Processing Strategy Across Distributed Services

## Purpose
Establishes the enterprise-wide idempotency, duplicate prevention, and exactly-once processing (EOP) strategy across all synchronous APIs, asynchronous Kafka event streams, payment gateway webhooks, and permissioned blockchain transactions in the Growww platform. In a high-speed financial trading and fractional equity settlement system, duplicate requests, network retries, or distributed partial failures must never result in duplicate orders, double-debiting user funds, or duplicate asset minting.

This prompt provides backend engineers, platform architects, and blockchain developers with the canonical specifications for the IETF Idempotency-Key standard, Redis distributed locking and deduplication filters, PostgreSQL Transactional Outbox pattern, Kafka Exactly-Once Semantics (EOS), and on-chain nonce management on Hyperledger Besu.

## What You Are Building
A comprehensive idempotency and exactly-once processing architecture specification (`docs/architecture/idempotency_and_exactly_once.md`), reusable middleware interceptors, Transactional Outbox schema, and Kafka EOS configurations:
- IETF Idempotency-Key Protocol: Standardized HTTP header (`Idempotency-Key: <UUIDv7>`) and gRPC metadata interceptor with Redis distributed locking and response caching.
- Transactional Outbox Pattern: PostgreSQL outbox table definition (`outbox_events`), CDC de-queuing workers, and atomic business-state-plus-event commit mechanisms.
- Kafka Exactly-Once Semantics (EOS): Configuration guides for transactional Kafka producers (`transactional.id`, `enable.idempotence=true`) and read-committed consumers.
- Smart Contract Replay & Nonce Defense: Contract-level transaction deduplication, EIP-712 structured data domain separators, and transaction relayer nonce management.

## Scope Boundaries
- **In Scope:**
 - Idempotency key lifecycle, caching, TTLs, and concurrency conflict resolution (HTTP 409 Conflict).
 - Transactional Outbox database schema and event publisher workers.
 - Kafka transactional producer and consumer configurations.
 - Blockchain relayer nonce synchronization and smart contract replay prevention.
 - Reusable Go and Python idempotency middleware/interceptors.
- **Out of Scope / Handled Elsewhere:**
 - REST/gRPC general API design standards (handled in Prompt 103).
 - Kafka cluster infrastructure deployment (handled in Prompt 403 & 802).
 - Specific payment gateway webhook handlers (handled in Prompt 212).

## Technology to Use
- **Distributed Caching & Locking:** Redis Enterprise Cluster 7.2+ using Redis Redlock and atomic Lua scripts (`SET key val NX EX ttl`).
- **Relational Persistence:** PostgreSQL 16+ ACID transactions with `SELECT ... FOR UPDATE SKIP LOCKED` for outbox polling.
- **Messaging Framework:** Apache Kafka 3.7+ with Transactional Coordinator and idempotent producer protocols.
- **Middleware Libraries:**
 - *Go:* Custom gRPC server interceptor and `gin`/`echo` HTTP middleware.
 - *Python / FastAPI:* Custom Starlette middleware and SQLAlchemy transaction lifecycle hooks.
 - *Rust:* `tower` service middleware for matching engine RPCs.

## Backend / Infra Touchpoints
- **Envoy API Gateway:** Validates presence of `Idempotency-Key` header on state-changing REST endpoints (`POST`, `PUT`, `PATCH`, `DELETE`).
- **Microservice Layer:** Executes deduplication checks before executing domain logic and caches responses in Redis upon completion.
- **PostgreSQL Databases:** Stores operational records and outbox events in a single atomic SQL transaction.
- **Kafka Cluster:** Coordinates 2-phase commit transactions for producers emitting to multiple topics.

## Blockchain Interaction
Guarantees exactly-once execution for permissioned Hyperledger Besu smart contract transactions:
- **Relayer Monotonic Nonce Management:** The backend transaction relayer maintains an in-memory, Redis-backed sequence counter ensuring transactions to Besu nodes use strictly sequential nonces without gaps or duplicates.
- **Smart Contract Execution Nonces:** Smart contracts (`SettlementDvP.sol`, `DigitalSecurityToken.sol`) track executed settlement IDs in a contract-level mapping (`mapping(bytes32 => bool) public executedSettlements`). Attempting to re-execute an already processed settlement ID reverts immediately with `SETTLEMENT_ALREADY_PROCESSED`.
- **EIP-712 Typed Signing Nonces:** Investor authorization signatures incorporate an incremental user nonce and chain ID, preventing cross-chain and replay attacks.

## Step-by-Step Build Instructions
1. Author `docs/architecture/idempotency_and_exactly_once.md` detailing the multi-tier deduplication lifecycle.
2. Define the Idempotency Key protocol conforming to the IETF draft standard:
 - Header: `Idempotency-Key: <UUIDv7>`.
 - Scope: Per authenticated `user_id` + HTTP Method + Resource Path.
 - Lifetime: 24 hours TTL in Redis.
3. Design the 3-state Idempotency Key lifecycle in Redis:
 - State 1: `IN_FLIGHT` (Locks request processing for up to 30 seconds using Redis `SET key "IN_FLIGHT" NX EX 30`).
 - State 2: `COMPLETED` (Stores HTTP status code, headers, and serialized response body with 24h TTL).
 - State 3: `FAILED` (Deletes key on transient infrastructure failure, allowing immediate client retry).
4. Implement concurrent collision handling: If a concurrent request arrives with the same key while status is `IN_FLIGHT`, return `HTTP 409 Conflict` with `Retry-After: 2`.
5. Implement the Go gRPC idempotency interceptor (`pkg/middleware/idempotency.go`).
6. Implement the Python FastAPI idempotency middleware (`src/middleware/idempotency.py`).
7. Design the Transactional Outbox SQL schema:
 - Table `outbox_events`: `id` (UUIDv7), `aggregate_type`, `aggregate_id`, `event_type`, `payload` (JSONB), `headers` (JSONB), `created_at`, `published_at`, `status` (`PENDING`, `PUBLISHED`, `FAILED`).
8. Implement the Outbox Publisher Worker:
 - Polls `outbox_events` using `SELECT * FROM outbox_events WHERE status = 'PENDING' ORDER BY created_at ASC LIMIT 100 FOR UPDATE SKIP LOCKED`.
 - Publishes events to Kafka within an idempotent producer session.
 - Marks outbox records as `PUBLISHED` with `published_at = NOW()`.
9. Configure Kafka Producers for Exactly-Once Semantics:
 - `enable.idempotence = true`
 - `acks = all`
 - `max.in.flight.requests.per.connection = 1` (or `<= 5` with idempotence)
 - `transactional.id = "<service-name>-<pod-id>"`
10. Configure Kafka Consumers for Exactly-Once Processing:
 - `isolation.level = read_committed`
 - Implement consumer-side message deduplication using database unique constraints on `event_id`.
11. Implement blockchain relayer nonce management and smart contract deduplication guards.
12. Formulate integration tests simulating network timeouts, client retry storms, pod crashes, and broker failovers.
13. Publish `docs/architecture/idempotency_and_exactly_once.md` to repository.

## Interfaces / Contracts

### Transactional Outbox PostgreSQL Schema
```sql
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    topic VARCHAR(255) NOT NULL,
    partition_key VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_pending ON outbox_events (created_at ASC) 
WHERE status = 'PENDING';
```

### Redis Idempotency Record JSON Structure
```json
{
  "key": "idemp:usr_01HZX89AB:POST:/api/v1/orders:d8e6a1b2-c3d4-4e5f-6a7b-8c9d0e1f2a3b",
  "status": "COMPLETED",
  "created_at": 1718811900,
  "http_status_code": 201,
  "response_headers": {
    "Content-Type": "application/json",
    "X-Request-ID": "req_01HZX89AB72K9M12P5QRSTUVWX"
  },
  "response_body": {
    "order_id": "ord_01HZX89AB72K9M12P5QRSTUVWX",
    "status": "OPEN",
    "quantity": "10.500000",
    "limit_price": { "currency_code": "INR", "units": 2450, "nanos": 500000000 }
  }
}
```

### Smart Contract Deduplication Pattern (Solidity Snippet)
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract SettlementDvP {
    mapping(bytes32 => bool) public executedSettlements;

    event DvPExecuted(bytes32 indexed settlementId, address indexed buyer, address indexed seller, uint256 amount);

    function executeDvP(
        bytes32 settlementId,
        address buyer,
        address seller,
        address tokenContract,
        uint256 tokenQuantity,
        uint256 fiatAmountPaise
    ) external {
        // Enforce exact-once settlement execution
        require(!executedSettlements[settlementId], "SETTLEMENT_ALREADY_PROCESSED");
        executedSettlements[settlementId] = true;

        // Perform atomic token and payment verification logic...
        emit DvPExecuted(settlementId, buyer, seller, tokenQuantity);
    }
}
```

## Security & Compliance Notes
- **Prevention of Double-Spending:** Strict idempotency at the API Gateway, Service Layer, and Blockchain Contract layers guarantees that under no circumstances (retry bursts, duplicate webhook deliveries, packet duplication) can financial balances be debited twice.
- **Race Condition Immunity:** Redis atomic `SET NX` locks ensure that concurrent identical requests from a compromised client or automated script cannot bypass deduplication guards.
- **Audit Logging of Duplicate Invocations:** Every detected duplicate invocation is logged with an `IDEMPOTENT_REPLAY` security tag to detect potential replay attacks or malfunctioning client software.

## Acceptance Criteria
- [ ] Complete Idempotency & Exactly-Once Processing document (`docs/architecture/idempotency_and_exactly_once.md`) published.
- [ ] IETF Idempotency-Key protocol and Redis 3-state caching lifecycle specified with concrete code examples.
- [ ] Transactional Outbox PostgreSQL schema and de-queuing worker algorithm defined and verified.
- [ ] Kafka Exactly-Once Semantics (EOS) producer and consumer configurations documented.
- [ ] Smart contract deduplication pattern and blockchain relayer nonce management finalized.
- [ ] Automated tests verify that 100 concurrent duplicate requests result in exactly 1 execution and 99 cached or conflict responses.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 103 (API Standards), Prompt 104 (Event Standards), Prompt 111 (Domain Model).
- **Parallel Work:** Prompt 105 (Auth Architecture), Prompt 109 (Secrets Management).
- **Blocks:** Prompt 203 (Wallet Service), Prompt 204 (Order Service), Prompt 208 (Settlement Service), Prompt 212 (Payment Gateway), Prompt 306 (DvP Contract).
