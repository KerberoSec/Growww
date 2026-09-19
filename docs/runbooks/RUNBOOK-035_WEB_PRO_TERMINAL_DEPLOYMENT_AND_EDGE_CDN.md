# SRE Operational Runbook 035: Web Pro Terminal Deployment & Edge CDN Architecture

**Runbook ID:** RUNBOOK-035-WEB-PRO-DEPLOY  
**Severity Classification:** Tier 1 (Mission-Critical Frontend Delivery)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** Web Pro Trading Terminal (`apps/growww_web`)  
**Target Environments:** Cloudflare Edge / Fastly CDN / Vercel Enterprise / Kubernetes Web Pods  
**Last Review:** September 2026  

---

## 1. Executive Summary & Production Objectives

The Web Pro Trading Terminal (`apps/growww_web`) is the primary browser-based point of access for institutional traders, retail day traders, and algorithmic scalpers. Any frontend deployment downtime, cache poisoning, or Content Security Policy (CSP) misconfiguration can lock traders out of active risk management during fast-moving markets.

### Service Level Objectives (SLOs):
- **Edge Availability:** 99.999% global edge uptime.
- **Time to First Byte (TTFB):** < 50ms globally across 250+ Edge Points of Presence (PoPs).
- **First Contentful Paint (FCP):** < 400ms on 4G / broadband networks.
- **Cumulative Layout Shift (CLS):** Exactly `0.00` (strict tabular typography).
- **Cache Invalidation Latency:** < 500ms global purge on atomic releases.

---

## 2. Global Edge CDN Topology

```
+-------------------------------------------------------------------------------------------------------+
|                                  WEB PRO TERMINAL GLOBAL CDN TOPOLOGY                                 |
|                                                                                                       |
|  [ End User Browser ]                                                                                 |
|          |                                                                                            |
|          v (Anycast DNS / HTTP/3 QUIC)                                                                |
|  [ Global Edge CDN Layer ] (Cloudflare / Fastly)                                                      |
|  - TLS 1.3 Termination (Zero-RTT 0-RTT Session Resumption)                                             |
|  - Brotli Level 11 Static Compression                                                                 |
|  - Strict Content Security Policy (CSP) & HSTS Injection                                              |
|  - DDoS & Web Application Firewall (WAF) Rate Limiting                                                |
|          |                                                                                            |
|          +----------------------------+----------------------------+                                  |
|          | (Static Cache Hit)         | (Dynamic RSC Request)      | (WSS Market Stream)              |
|          v                            v                            v                                  |
|  [ Edge Static Cache ]       [ Next.js 14 App Router ]    [ WebSocket Gateway ]                       |
|  - `.next/static/*`          - Dynamic Server Components  - `wss://ws.growww.in`                      |
|  - Immutable (1 Year Cache)  - Non-Cacheable Dynamic Data - 50ms Conflation Stream                    |
|  - Zero-Copy WASM Modules    - Multi-Region K8s Cluster   - Dedicated Web Worker                      |
+-------------------------------------------------------------------------------------------------------+
```

---

## 3. Production Build & Container Pipeline

### 3.1 Next.js 14 Standalone Production Container
The application compiles using Next.js 14 `output: 'standalone'` mode, generating a minimal Docker image (<90MB) based on Google Distroless Node.js runtime:

```dockerfile
# apps/growww_web/Dockerfile (Production Specification)
FROM node:20-alpine AS builder
WORKDIR /app
RUN apk add --no-cache libc6-compat
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN pnpm build

FROM gcr.io/distroless/nodejs20-debian12:nonroot AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

USER nonroot
EXPOSE 3000
CMD ["server.js"]
```

---

## 4. Content Security Policy (CSP) & Web3 Headers

To protect non-custodial Web3 credentials, prevent Cross-Site Scripting (XSS), and allow WebAssembly and Web Workers:

```text
Content-Security-Policy: 
  default-src 'self';
  script-src 'self' 'wasm-unsafe-eval' 'strict-dynamic';
  worker-src 'self' blob:;
  connect-src 'self' wss://ws.growww.in wss://*.walletconnect.com https://*.walletconnect.com https://besu.growww.in;
  img-src 'self' data: blob: https://assets.growww.in;
  font-src 'self' data:;
  style-src 'self' 'unsafe-inline';
  frame-ancestors 'none';
  base-uri 'self';
  form-action 'self';
```

---

## 5. Deployment Verification & Smoke Test Protocol

Upon completion of a production canary rollout:
1. Verify HTTP/3 connection: `curl --http3 -I https://trade.growww.in`
2. Validate zero-fee invariant rendered on order ticket:
   - Verify `Trading Fee: ₹0.00`
   - Verify `0.00% Zero-Fee Badge` is visible.
3. Validate Web Worker streaming: Inspect browser console, ensure worker connects to `wss://ws.growww.in` and SharedArrayBuffer initializes cleanly.
4. Verify PWA Service Worker registration: Confirm Workbox registers and caches app shell in IndexedDB cache.
