# 402 - Redis Usage Patterns (Session, Cache, Real-Time Order Book State)

## Purpose
In a high-throughput financial trading and fractional asset platform, sub-millisecond data access is essential for real-time market data streaming, fast pre-trade checks, session validation, and distributed resource synchronization. Redis 7.2 serves as the distributed in-memory data tier for Growww, offloading read pressure from PostgreSQL and providing ultra-low-latency data structures.

This prompt defines standard Redis caching patterns, key topologies, serialization protocols, TTL policies, atomic Lua scripts, and distributed locking mechanisms across all backend services. It ensures consistent cache-aside, write-through, and read-through behavior, preventing stale market state, cache stampedes, and race conditions during high-volume trading hours.

## What You Are Building
A production-ready Redis data architecture and reusable client SDK package containing:
- **Redis Architecture & Key Namespace Standard:** Documentation and specification (`docs/data/redis_patterns.md`) establishing hierarchical key conventions (`growww:<env>:<domain>:<entity>:<id>`) and TTL matrices.
- **Standardized Client Libraries:** Unified, resilient Redis client wrappers in Go (`pkg/redis/`) and Python (`core/redis/`) featuring connection pooling, circuit breaking, exponential backoff, and distributed tracing.
- **Real-Time Order Book & Market Depth Cache:** Optimized Redis Sorted Set (`ZSET`) and Hash structures mirroring L2 order book bids and asks for sub-millisecond price quote queries.
- **Sliding-Window Rate Limiter Engine:** Atomic Lua scripts implementing rolling sliding-window rate limiting per user, IP, and API key.
- **Distributed Lock (Redlock) Implementation:** Safe distributed mutex with fencing tokens, preventing duplicate order placement and double-spend race conditions.
- **Session & Compliance Whitelist Cache:** High-speed in-memory store for active JWT session revocation lists and on-chain KYC whitelist lookups.

## Scope Boundaries
- **In Scope:** Redis key schemas, eviction policies, Lua scripts, distributed lock protocols, client SDK wrappers, cache invalidation strategies, and cluster topology.
- **Out of Scope / Handled Elsewhere:**
 - In-memory order matching engine state inside the core matching engine process (handled in Prompt 205).
 - PostgreSQL persistent transactional storage (handled in Prompt 401).
 - Durable event log and message brokering (handled in Prompt 403).
 - WebSocket market data broadcast to Flutter clients (handled in Prompt 207).

## Technology to Use
Redis 7.2 Enterprise / Cluster mode is selected as the distributed in-memory caching engine. Redis provides deterministic sub-millisecond response times, single-threaded atomic command execution, expressive native data structures (Hashes, Sorted Sets, Streams, Bitmaps), and robust cluster sharding. Compared to Memcached, Redis offers complex collection types, persistence options (AOF/RDB), and server-side Lua scripting required for atomic financial operations.

- **Engine:** Redis 7.2 Cluster with Multi-AZ replication (3 Shards, 1 Primary + 1 Replica per shard).
- **Go Client:** `go-redis/v9` with OpenTelemetry hook integration.
- **Python Client:** `redis-py` (asyncio) 5.0+.
- **Serialization:** Google Protocol Buffers / MsgPack for high-density payloads; JSON for metadata inspection.
- **Eviction Policy:** `volatile-lru` (evicts least recently used keys with an explicit TTL set).

## Backend / Infra Touchpoints
- **Redis Cluster:** Multi-AZ Kubernetes deployment via Redis Operator with automated failover and slot rebalancing.
- **Microservices:** API Gateway (219), Rate Limiter (220), Order Service (204), Market Data Service (207), User Service (201).
- **PostgreSQL 16:** Backing database for cache-aside reads and write-through cache invalidation.
- **Prometheus & Grafana:** Redis Exporter tracking memory fragmentation, cache hit/miss ratio, connected clients, and slow log.

## Blockchain Interaction
Redis accelerates off-chain performance by caching read-heavy blockchain state from the Hyperledger Besu consortium network (QBFT consensus, 2-second block finality).

### Detailed On-Chain Integration Mechanics:
- **KYC Whitelist Mirror:** Caches investor whitelist verification status from `ComplianceRegistry.sol` (Key: `growww:prod:compliance:whitelist:<0x_address>`, TTL: 5 minutes) to perform instant pre-trade validation without calling Besu JSON-RPC on every order.
- **Validator Node Health & Sync State:** Caches validator node connectivity, block height, peer counts, and QBFT consensus round state for rapid health reporting (TTL: 2 seconds).
- **Recent Block & Tx Receipts:** Caches recent transaction hashes (`tx_hash`) and confirmation status emitted by `SettlementDvP.sol` and `DigitalSecurityToken.sol` for fast UI status polling before DB indexing completes.
- **Proof-of-Reserve Root Cache:** Caches the latest published Merkle root hash from `ProofOfReserveRegistry.sol` for real-time verification against depository balances.

## Step-by-Step Build Instructions
1. Define the global Redis key namespace hierarchy and document TTL guidelines in `docs/data/redis_patterns.md`.
2. Configure Redis Cluster deployment manifests with `maxmemory`, `maxmemory-policy volatile-lru`, and `appendonly yes` (AOF everysec).
3. Author Lua script for atomic sliding-window rate limiting using Redis Sorted Sets (`ZSET`).
4. Author Lua script for atomic order balance reservation locks with automatic timeout release.
5. Author Lua script for L2 order book price level batch updates (`ZADD` / `ZREMRANGEBYSCORE`).
6. Implement Go Redis client wrapper (`pkg/redis/client.go`) with connection pool tuning, automatic retries, and distributed tracing.
7. Implement Python async Redis client wrapper (`core/redis/client.py`) with connection recycling and error handling.
8. Implement distributed lock manager with fencing tokens and lease renewal extensions.
9. Implement cache-aside helper utilities with Probabilistic Early Expiration (XFetch / Jitter) to eliminate cache stampedes.
10. Implement session revocation blacklist using Redis Bitmaps or Hashes with automated expiration.
11. Implement Redis Sentinel / Cluster failover test harness verifying zero data loss for locked keys during leader election.
12. Establish Prometheus alerts for memory usage >80%, eviction rates >100/sec, and slow log commands taking >10ms.

## Interfaces / Contracts

```lua
-- Lua Script: Atomic Sliding-Window Rate Limiter
-- KEYS[1]: Rate limit key (e.g., growww:prod:ratelimit:user:123:orders)
-- ARGV[1]: Current timestamp in milliseconds
-- ARGV[2]: Window size in milliseconds (e.g., 60000 for 1 minute)
-- ARGV[3]: Max requests allowed in window (e.g., 100)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local clearBefore = now - window

-- Remove timestamps outside the sliding window
redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)

-- Count current requests in window
local currentRequests = redis.call('ZCARD', key)

if currentRequests < limit then
    -- Add current request timestamp
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return {1, limit - currentRequests - 1} -- Allowed (1), remaining capacity
else
    return {0, 0} -- Blocked (0), 0 capacity
end
```

```yaml
# Redis Key Hierarchy & TTL Matrix
Key Schemes:
  User Session:
    Pattern: "growww:<env>:session:<user_id>:<device_id>"
    Type: "HASH"
    TTL: "86400s (24h with rolling extension)"
    Fields: ["jwt_jti", "role", "entity_type", "last_active_at"]

  Order Book L2 Depth:
    Pattern: "growww:<env>:orderbook:<symbol>:bids"
    Type: "ZSET (score = price, member = price_level_json)"
    TTL: "No TTL (Continuous sync from Matching Engine)"

  Distributed Mutex (Redlock):
    Pattern: "growww:<env>:lock:<resource_type>:<resource_id>"
    Type: "STRING (value = fencing_token_uuid)"
    TTL: "5000ms (Explicit release or timeout)"

  Compliance Whitelist:
    Pattern: "growww:<env>:compliance:whitelist:<blockchain_address>"
    Type: "STRING (value = 'ACTIVE' | 'FROZEN')"
    TTL: "300s (5 minutes)"

  Idempotency Key:
    Pattern: "growww:<env>:idempotency:<idempotency_key>"
    Type: "STRING (value = response_payload_hash)"
    TTL: "120s (2 minutes)"
```

## Security & Compliance Notes
- **Network Isolation & Authentication:** Redis Cluster runs inside a private VPC subnet with no public internet ingress. Mandatory Redis AUTH with 64-character rotated passwords and TLS 1.3 encryption on all client-server and node-to-node communication.
- **Zero PII in Keys & Values:** Redis keys must NEVER contain user phone numbers, PAN cards, email addresses, or investor real names. Only internal UUIDs and cryptographic wallet addresses are permitted.
- **Access Control Lists (ACLs):** Individual microservice service accounts are provisioned with restricted ACLs (e.g. Market Data service has read-only access to order books; Rate Limiter has access only to `ratelimit:*` keys).
- **Non-Persistence of Sensitive Crypto Keys:** Validator private keys, HSM PINs, and KMS tokens must NEVER be stored or cached in Redis.

## Acceptance Criteria
- [ ] Redis 7.2 Cluster deployed with 3 shards across multi-AZ with automatic failover verified in <3 seconds.
- [ ] Standard key namespace hierarchy and TTL policies validated across all Category 2 services.
- [ ] Atomic sliding-window rate limiting Lua script benchmarked at >20,000 evaluations/sec with zero race conditions.
- [ ] Redlock distributed lock implementation verified with fencing tokens to prevent double-spending or duplicate order submission.
- [ ] Order book L2 depth data structures (`ZSET`) tested with sub-millisecond price insertion and top-of-book slicing (`ZREVRANGE`).
- [ ] Cache-aside wrappers with jitter successfully prevent cache stampede under simulated burst traffic (10,000 concurrent requests).
- [ ] Compliance whitelist caching reduces Hyperledger Besu RPC node queries by >95% for order validation flows.
- [ ] Prometheus metrics and alerts operational for memory fragmentation, slow queries, and cluster node health.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 103 (API Standards), Prompt 204 (Order Service).
- **Parallel Tasks:** Prompt 401 (PostgreSQL Schema), Prompt 403 (Kafka Cluster), Prompt 220 (Rate Limiter).
