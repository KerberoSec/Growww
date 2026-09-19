# 220 - Rate Limiting & Abuse Prevention Service (Redis / Token Bucket / Behavioral Scoring)

## Purpose
The Rate Limiting & Abuse Prevention Service shields Growww's financial trading infrastructure, public API gateways, and backend matching systems from distributed denial-of-service (DDoS) attacks, brute-force credential stuffing, order book spamming, high-frequency quote scraping, and market manipulation attempts. In an institutional securities exchange operating in real time, malicious traffic or runaway algorithmic bots can degrade order execution latency, destabilize settlement pipelines, or compromise investor accounts.

This service provides distributed, microsecond-latency rate limiting and behavioral risk scoring, utilizing multi-tiered sliding-window algorithms and Redis token buckets. It enforces differentiated throughput quotas based on user verification tier (Guest, KYC-Verified, Institutional Market Maker), IP reputation, and anomalous behavior patterns.

## What You Are Building
A high-throughput Go microservice and middleware library (`services/abuse-prevention-service`) delivering:
- **Distributed Multi-Tier Rate Limiting Engine:** Atomic Token Bucket and Sliding Window Counter implementations running on Redis Cluster.
- **Dynamic Tier-Based Quota Evaluator:** Dynamic limits based on client authentication level (Guest: 20 req/min; KYC-1: 120 req/min; Trading APIs: 300 req/min; Institutional FIX: 5000 req/min).
- **Behavioral Anomaly & Order-Spam Detector:** Real-time detection of rapid order placement/cancellation cycling (quote stuffing) and credential stuffing attempts.
- **Adaptive Challenge & IP Banning Manager:** Triggers CAPTCHA challenges, temporary IP quarantines, or permanent blacklists.
- **Artifacts Delivered:**
 - `services/abuse-prevention-service/cmd/server/main.go` - Go service entry point.
 - `services/abuse-prevention-service/internal/limiter/sliding_window.go` - Redis sliding window engine.
 - `services/abuse-prevention-service/internal/detector/order_spam.go` - Trading abuse detection heuristics.
 - `services/abuse-prevention-service/internal/lua/token_bucket.lua` - Atomic Redis Lua rate-limiting scripts.
 - `proto/growww/abuse/v1/abuse.proto` - Internal gRPC service definitions.

## Scope Boundaries
- **In Scope:**
 - Per-IP, per-user, per-device, and per-endpoint distributed rate limiting.
 - Low-latency Redis Lua script execution for atomic limit verification (<1ms).
 - Trading abuse pattern detection (e.g. order cancel ratio $> 95\%$, order submission spikes).
 - Integration with Cloudflare / AWS WAF for IP reputation and automated blocking.
- **Out of Scope / Handled Elsewhere:**
 - Layer 3/4 network DDoS mitigation (handled by Cloudflare Magic Transit / AWS Shield, Prompt 803).
 - Pre-trade financial risk and margin balance checks (handled by Prompt 206).
 - User password hashing and authentication verification (handled by Prompt 201).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `go-redis/v9` with optimized Lua scripts, and `oschwald/geoip2-golang` for GeoIP reputation parsing.
- **Justification:** Go delivers the ultra-low latency (<0.5ms evaluation time), high concurrency, and low memory footprint essential for a security middleware service that sits directly on the critical path of every incoming API request and trading order.
- **Dependencies & Libraries:**
 - Redis 7.2+ Cluster for distributed atomic token buckets and sliding window logs.
 - PostgreSQL 16+ for long-term abuse incident logs, IP blacklists, and policy configurations.
 - MaxMind GeoIP2 / IPQualityScore feeds for real-time proxy/VPN risk scoring.
 - Apache Kafka 3.7+ for emitting security incident events.

## Backend / Infra Touchpoints
- **Redis 7.2 Cluster:** Atomic sliding window keys, rate quota buckets, and ephemeral IP lockouts.
- **PostgreSQL 16:** Global blocklists, allowlists, abuse rule configurations, and audit records.
- **API Gateway (Prompt 219):** Executes rate-limiting middleware on all incoming HTTP/gRPC requests.
- **Kafka Topics:**
 - Publishes: `security.rate_limit.exceeded`, `security.abuse.detected`, `security.ip.blocked`.
 - Subscribes: `order.placed`, `order.cancelled`, `user.auth.failed`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network running QBFT consensus.
- **Relayer & Node RPC Protection:** The Rate Limiting Service protects the Hyperledger Besu JSON-RPC nodes and HSM Relayers by enforcing strict quota limits on blockchain verification and proof-of-reserve query endpoints.
- **Zero On-Chain Overhead:** Rate limiting decisions, heuristics, and IP reputation scores operate entirely off-chain to avoid consuming blockchain transaction bandwidth or polluting ledger state.
- **Zero PII:** IP addresses and device fingerprints are hashed with cryptographic salt before being emitted in security audit events.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/abuse-prevention-service` with standardized configuration and metrics scaffolding.
2. Define Protobuf definitions in `proto/growww/abuse/v1/abuse.proto` and compile Go gRPC stubs.
3. Configure PostgreSQL schema migrations for `abuse_policies`, `ip_reputation_records`, and `security_incidents`.
4. Implement atomic Redis Lua scripts for the Sliding Window Counter and Token Bucket rate-limiting algorithms.
5. Build the gRPC interceptor and HTTP middleware that queries the rate limiter with sub-millisecond timeout constraints.
6. Implement the dynamic quota resolution engine mapping JWT claims, API keys, and client tiers to rate policies.
7. Build the Order-Spam Detection module analyzing order-to-cancellation ratios and order velocity in real time.
8. Implement Credential-Stuffing detection heuristics tracking failed login spikes across IP subnets and user accounts.
9. Integrate GeoIP2 database and VPN/Proxy detection for automated high-risk IP scoring.
10. Integrate Kafka event publishers producing `security.abuse.detected` and consumers updating real-time threat scores.
11. Add Prometheus metrics (`rate_limit_checks_total`, `rate_limit_exceeded_total`, `abuse_check_latency_microseconds`).
12. Write rigorous benchmark and fuzz tests simulating 100,000 requests/second bursts, validating zero race conditions and sub-millisecond response latency.

## Interfaces / Contracts

### Protobuf Definition (`abuse.proto`)
```protobuf
syntax = "proto3";

package growww.abuse.v1;

option go_package = "github.com/growww/services/abuse-prevention/gen/v1;abusev1";

service AbusePreventionService {
  rpc EvaluateRequest (EvaluateRequestPayload) returns (EvaluateResponse);
  rpc ReportAuthFailure (AuthFailureReport) returns (ReportResponse);
  rpc ManageBlocklist (BlocklistRequest) returns (BlocklistResponse);
  rpc GetAbuseMetrics (AbuseMetricsRequest) returns (AbuseMetricsResponse);
}

message EvaluateRequestPayload {
  string client_ip = 1;
  string user_id = 2; // Optional for unauthenticated requests
  string device_fingerprint = 3;
  string route_path = 4;
  string http_method = 5;
  string client_tier = 6; // GUEST / KYC_1 / TRADER / INSTITUTIONAL
  int32 cost_weight = 7; // Request complexity weight (e.g. 1 for get, 5 for order)
}

message EvaluateResponse {
  bool is_allowed = 1;
  string action = 2; // ALLOW / THROTTLE / CHALLENGE_CAPTCHA / BLOCK
  int64 remaining_quota = 3;
  int64 reset_duration_ms = 4;
  string rejection_reason = 5;
}

message AuthFailureReport {
  string client_ip = 1;
  string attempted_identifier = 2;
  string device_fingerprint = 3;
  int64 timestamp = 4;
}

message ReportResponse {
  bool action_taken = 1;
  string current_threat_level = 2;
}

message BlocklistRequest {
  string entity_type = 1; // IP / SUBNET / USER_ID / DEVICE_ID
  string entity_value = 2;
  string reason = 3;
  int64 duration_seconds = 4; // 0 for permanent
  string admin_id = 5;
}

message BlocklistResponse {
  bool success = 1;
  int64 expires_at = 2;
}

message AbuseMetricsRequest {}
message AbuseMetricsResponse {
  int64 total_requests_evaluated = 1;
  int64 total_throttled = 2;
  int64 total_blocked = 3;
  int64 active_blocked_ips_count = 4;
}
```

### Redis Lua Script (`sliding_window.lua`)
```lua
-- KEYS[1]: Rate limit key (e.g. "rate:{tier}:{identifier}:{route}")
-- ARGV[1]: Current timestamp in milliseconds
-- ARGV[2]: Window size in milliseconds
-- ARGV[3]: Max allowed requests in window
-- ARGV[4]: Cost weight of current request

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local clear_before = now - window

-- Remove elements older than window
redis.call('ZREMRANGEBYSCORE', key, 0, clear_before)

-- Count current requests
local current_requests = redis.call('ZCARD', key)

if current_requests + cost <= limit then
    for i = 1, cost do
        redis.call('ZADD', key, now, now .. '-' .. math.random(1000, 9999) .. '-' .. i)
    end
    redis.call('PEXPIRE', key, window)
    return {1, limit - (current_requests + cost), 0} -- Allowed, remaining, retry_after
else
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local retry_after = 0
    if #oldest > 0 then
        retry_after = (tonumber(oldest[2]) + window) - now
    end
    return {0, 0, math.max(0, retry_after)} -- Denied, remaining=0, retry_after_ms
end
```

## Security & Compliance Notes
- **False-Positive Prevention on Order Execution:** Strict priority queues guarantee that critical trade exit or margin-stop orders are never throttled during volatile market moves.
- **DPDP Act Anonymization:** Client IP addresses stored in PostgreSQL are hashed with a rotating cryptographic salt; raw IP addresses are only retained temporarily in Redis memory for active rate-window evaluation.
- **Fail-Open vs Fail-Secure Configuration:** In the event of a total Redis cluster outage, non-trading browse endpoints fail-open to preserve user experience, while sensitive state-changing trading endpoints fail-secure with strict fallback limits.

## Acceptance Criteria
- [ ] Go rate-limiting service builds cleanly and passes all static analysis checks.
- [ ] Atomic Redis Lua sliding-window script executes in under 0.8ms at 99th percentile under load.
- [ ] Tiered rate limits correctly distinguish Guest, KYC-Verified, and Institutional API keys.
- [ ] Order-spam detection automatically flags accounts with >95% cancellation ratios within 10-second sliding windows.
- [ ] Blocklisted IPs and devices are rejected immediately at the gateway layer.
- [ ] Automated benchmarks sustain 100,000 evaluations/sec with zero race conditions.
- [ ] Test coverage exceeds >=85%.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design), Prompt 201 (User Service), Prompt 219 (API Gateway).
- **Subsequent / Parallel Tasks:** Prompt 204 (Order Service), Prompt 205 (Matching Engine).
