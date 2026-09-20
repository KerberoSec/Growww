# Envoy Proxy Rate Limiting & DDoS Mitigation Architecture (Prompt 814)

## Architecture Overview
Growww's perimeter ingress gateway leverages Envoy Proxy v1.30+ to deliver institutional-grade L4/L7 threat mitigation, DDoS absorption, and tiered rate limiting.

### Defense-in-Depth Layer Stack
1. **L4 TCP Connection Flooding Mitigation**: `envoy.filters.network.connection_limit` bounds concurrent open sockets (up to 50,000 active connections) with delayed connection acceptance.
2. **Slowloris & Buffer Exhaustion Protection**: Clamped connection buffer limits (`per_connection_buffer_limit_bytes: 32768`) and granular idle stream timeouts.
3. **L7 Local Token-Bucket Burst Defense**: `envoy.filters.http.local_ratelimit` absorbs volumetric micro-bursts locally in-process without overloading the Redis ratelimit tier.
4. **Automated Anomaly & Scanner Filtering**: `envoy.filters.http.lua` script drops malicious scanning user-agents (sqlmap, nikto, masscan) and validates timestamp skew on high-frequency order placement.
5. **Distributed Tiered Rate Limiting**: `envoy.filters.http.ratelimit` connects via gRPC to Envoy's global RateLimit service backed by Redis 7.2.
   - **Tier 0 (Public/Unauthenticated)**: 10 req/s
   - **Tier 1 (Retail Clients)**: 100 req/s
   - **Tier 2 (Institutional API)**: 2,000 req/s
   - **Tier 3 (Market Makers / Co-Lo)**: 10,000 req/s
   - **Brute-Force Auth Protection**: 5 attempts/minute per IP
6. **Adaptive Concurrency Control**: Dynamic gradient controller protecting the Order Matching Engine and upstream gateways from cascading latencies and brownout conditions.
7. **Circuit Breakers & Outlier Detection**: Automatic outlier pod ejection on consecutive 5xx failures with configurable recovery backoff.
