# 817 - Dual-Environment Testnet (Demo) vs Mainnet (Real) Infrastructure & Matching Pipeline

## Purpose
In a regulated digital asset and tokenized securities exchange, retail traders, institutional market participants, and algorithmic trading desks require a realistic sandbox environment (Demo / Paper Trading Testnet) to test execution strategies, evaluate market dynamics, and integrate API clients without placing real capital at risk. Simultaneously, the production exchange (Mainnet) manages real fiat balances, central bank digital currency (e-Rupee / CBDC) reserves, sovereign custody allocations, and legally binding Delivery-versus-Payment (DvP) trade settlements.

Operating these two environments presents a fundamental architectural challenge: traders on the Testnet expect realistic execution against live, real-world market movements (such as BTC/USDT, ETH/USDT, and tokenized equity price feeds), yet under no circumstances can virtual sandbox orders, test tokens, or synthetic balances cross-contaminate live production order books, custodial ledgers, or banking rails. Furthermore, market data from external venues must flow unidirectionally into both matching pipelines without exposing production networks to sandbox attack surfaces or credential leakage.

This specification details the end-to-end architecture, DevOps automation, Kubernetes topology, and network policies for Growww's Dual-Environment Testnet vs Mainnet Matching Pipeline (`infra/k8s/environments/`). It establishes an impermeable barrier separating Testnet (Demo) from Mainnet (Real) matching workloads, orchestrates a secure unidirectional market data feed from global spot feeds, isolates database roles and Redis namespaces, and routes on-chain transactions to dedicated, cryptographically distinct Hyperledger Besu consortium networks.

## What You Are Building
A production-grade, highly isolated dual-environment infrastructure and deployment framework (`infra/k8s/environments/`) comprising:
- **Environment Separation Manifests & Helm Charts (`infra/k8s/environments/`):** Declarative Helm charts and environment-specific values (`values-testnet.yaml`, `values-mainnet.yaml`) provisioning identical microservice topologies across dedicated Kubernetes namespaces (`nbse-testnet` and `nbse-mainnet`).
- **Zero-Trust Network Isolation Policies:** Strict Istio `AuthorizationPolicy`, `PeerAuthentication` (mTLS STRICT), and Kubernetes / Cilium `NetworkPolicy` manifests enforcing total Layer 3, Layer 4, and Layer 7 isolation between `nbse-testnet` and `nbse-mainnet`, blocking all lateral pod-to-pod and database traffic.
- **Unidirectional Market Data Feeder Hub (Prompt 272):** A zero-backchannel streaming architecture where live external spot feeds (BTC/USDT, ETH/USDT) ingest real-time ticks and fan out identical pricing feeds concurrently to both the Mainnet Matching Engine and the Testnet Demo Engine via read-only Kafka topic subscriptions and UDP multicast reflectors.
- **Spot Order Service Environment Router (Prompt 275):** Routing logic in the Spot Order Service that inspects authenticated JWT session claims (`environment: "testnet" | "mainnet"`) and API gateway routing headers, dispatching orders deterministically to either the high-performance Rust Matching Engine (Mainnet) or the Demo Engine (Testnet).
- **Virtual Matching Demo Engine (Prompt 273):** An in-memory paper trading engine running in `nbse-testnet` that simulates order fills against live feeder market depth, updates virtual wallet balances, and logs paper executions without invoking banking or depository settlements.
- **Database & Persistence Tier Segregation:** Independent PostgreSQL databases (`nbse_testnet_db` vs `nbse_mainnet_db`) with mutually exclusive database credentials, isolated PgBouncer pools, and dedicated Redis Cluster instances ensuring complete physical separation between real investor balances and virtual funds.
- **Dedicated Blockchain RPC Gateway Topology:** Isolated RPC routing terminating at `https://testnet.besu.nbse.in` (Chain ID `13371`) for demo token transfers and mock DvP contracts, versus `https://mainnet.besu.nbse.in` (Chain ID `2026`) backed by AWS CloudHSM signing enclaves for real asset tokenization and institutional settlement.

## Scope Boundaries
- **In Scope:**
  - Kubernetes manifests, Helm templates, and configuration values in `infra/k8s/environments/` for `testnet` and `mainnet`.
  - Istio Service Mesh policies: `PeerAuthentication`, `AuthorizationPolicy`, `VirtualService`, and `Gateway` routing rules.
  - Cilium eBPF / Kubernetes `NetworkPolicy` configurations denying all ingress and egress between `nbse-testnet` and `nbse-mainnet`.
  - Unidirectional feed fanout from Feeder (Prompt 272) across both environments.
  - Integration contracts and routing mechanisms for Spot Order Service (Prompt 275) and Demo Engine (Prompt 273).
  - Storage tier boundary enforcement: dedicated PostgreSQL schemas/databases, distinct database user roles, and segregated Redis namespaces/instances.
  - Blockchain RPC routing and EIP-155 Chain ID enforcement (`13371` for Testnet vs `2026` for Mainnet).
  - Automated CI/CD boundary leak detection tests and credential validation suites.
- **Out of Scope / Handled Elsewhere:**
  - Internal algorithm implementation and memory-mapped WAL of the Rust Matching Engine (handled in Prompt 205 and Prompt 246).
  - Internal state transition algorithms and fill simulation logic inside Demo Engine (handled in Prompt 273).
  - Low-level external exchange WebSocket adapter connectors in Feeder (handled in Prompt 272).
  - Flutter UI client environment toggle switch and theme adjustments (handled in Prompt 527).
  - AWS CloudHSM physical key provisioning and PKCS#11 driver lifecycle (handled in Prompt 717).
  - Multi-region disaster recovery and cross-DC consensus failover (handled in Prompt 816).

## Technology to Use
- **Container Orchestration & Packaging:** **Kubernetes 1.30+** with **Helm 3.14+** utilizing parameterized values trees (`infra/k8s/environments/values-testnet.yaml` and `infra/k8s/environments/values-mainnet.yaml`).
- **Service Mesh & Zero-Trust Security:** **Istio 1.22+** enforcing `STRICT` mutual TLS (mTLS) with SPIFFE/SPIRE workload identities, and **Cilium eBPF** enforcing high-performance kernel-level Layer 3/4 packet filtering.
- **In-Memory Caching & State Storage:** **Redis 7.2+ Cluster** deployed with dedicated instances per environment (`redis.testnet.svc` vs `redis.mainnet.svc`), utilizing namespace prefixes (`testnet:*` vs `mainnet:*`) as secondary defense-in-depth.
- **Relational Persistence:** **PostgreSQL 16+** with strictly segregated databases (`nbse_testnet_db` and `nbse_mainnet_db`), dedicated service users, and isolated **PgBouncer** connection pooling tiers.
- **Event Streaming Bus:** **Apache Kafka 3.7+ (KRaft mode)** utilizing separate tenant topic namespaces (`testnet.*` vs `mainnet.*`) with read-only shared market data topic access (`marketdata.feeder.ticks.v1`).
- **Consortium Blockchain Nodes:** **Hyperledger Besu 24.1+** operating under QBFT consensus across two independent networks:
  - *Regulatory Testnet Sandbox:* Chain ID `13371`, RPC endpoint `https://testnet.besu.nbse.in`.
  - *Production Mainnet:* Chain ID `2026`, RPC endpoint `https://mainnet.besu.nbse.in`.
- **Secret & Key Management:** **HashiCorp Vault** with External Secrets Operator (ESO) syncing distinct Kubernetes secrets into `nbse-testnet` and `nbse-mainnet` without cross-environment read access.

## Backend / Infra Touchpoints
- **Feeder (Prompt 272):** Connects to external liquidity providers (Binance, Coinbase, Deribit, LMAX) over secure outbound WebSockets. Feeder ingests top-of-book and L2 order book deltas for BTC/USDT, ETH/USDT, and tokenized indices, publishing normalized tick events to the unified Kafka topic `marketdata.feeder.ticks.v1`. Both Mainnet Order Service and Testnet Demo Engine subscribe to this topic in read-only mode.
- **Spot Order Service (Prompt 275):** Entry point for all user order submissions (`POST /api/v1/orders`). Validates order schema, checks trading permissions, evaluates the user's JWT context, and forwards the order via gRPC:
  - If `environment == "mainnet"`: routes to Mainnet Pre-Trade Risk Service (Prompt 206) and Rust Matching Engine (Prompt 205).
  - If `environment == "testnet"`: routes to Demo Engine (Prompt 273) with virtual balance verification.
- **Demo Engine (Prompt 273):** Deployed strictly within the `nbse-testnet` namespace. Maintains in-memory virtual order books, tracks virtual balance accounts, matches incoming demo orders against live market ticks published by Feeder (Prompt 272), and records synthetic fills in `nbse_testnet_db`.
- **Database Architecture:**
  - Testnet: Hosts tables `testnet_accounts`, `testnet_orders`, `testnet_trades`, `testnet_virtual_balances`. Accessed via role `testnet_app_user`.
  - Mainnet: Hosts tables `accounts`, `orders`, `trades`, `fiat_balances`, `dvp_settlements`. Accessed via role `mainnet_app_user`. No shared credentials or cross-database grants exist.
- **Kafka Topic Structure:**
  - Shared Topic (Read-Only to Testnet): `marketdata.feeder.ticks.v1`
  - Testnet Pipeline Topics: `testnet.orders.inbound.v1`, `testnet.orders.matched.v1`, `testnet.wallet.virtual_updated.v1`
  - Mainnet Pipeline Topics: `mainnet.orders.inbound.v1`, `mainnet.orders.matched.v1`, `mainnet.settlement.dvp_pending.v1`

## Blockchain Interaction
Growww operates dual, completely isolated Hyperledger Besu consortium networks to eliminate any risk of testnet asset minting affecting real financial ledgers:
- **Testnet Sandbox Network:**
  - RPC Endpoint: `https://testnet.besu.nbse.in`
  - Chain ID: `13371` (EIP-155 replay protected)
  - Consensus: 4-node QBFT with 1-second block periods.
  - Relayer Signer: Automated software private keys managed via HashiCorp Vault transit engine.
  - Smart Contracts: Mock token contracts (`MockERC20.sol`, `MockSettlementDvP.sol`, `MockDigitalSecurityToken.sol`).
  - Capabilities: Developers and retail demo users can trigger automated faucets, re-genesis states, and test mock atomic settlements.
- **Production Mainnet Network:**
  - RPC Endpoint: `https://mainnet.besu.nbse.in`
  - Chain ID: `2026` (EIP-155 replay protected)
  - Consensus: 7-node geographically distributed QBFT cluster spanning Mumbai (DC), GIFT City (DR), and Witness nodes.
  - Relayer Signer: Dedicated FIPS 140-2 Level 3 AWS CloudHSM partitions with multi-signature authorization.
  - Smart Contracts: Audited production contracts (`SettlementDvP.sol`, `DigitalSecurityToken.sol`, `ComplianceRegistry.sol`).
  - Restrictions: Testnet nodes, containers, and deployment service accounts have zero network egress or cryptographic access to Mainnet RPC endpoints or validator nodes.
- **EIP-155 Transaction Replay Defense:** All transactions are signed with strict EIP-155 chain identification. An EVM signature generated for Chain ID `13371` is mathematically rejected by Mainnet Besu nodes (Chain ID `2026`), preventing cross-network transaction replay even in the event of an accidental broadcast.

## Step-by-Step Build Instructions
1. **Scaffold Directory Layout:** Create directory structure `infra/k8s/environments/` containing subdirectories: `base/`, `testnet/`, `mainnet/`, `policies/`, `helm/`, and `tests/`.
2. **Author Base Helm Chart:** Define standard microservice deployment and service templates in `infra/k8s/environments/helm/matching-pipeline/` with configurable environment parameters, resource limits, and service mesh annotations.
3. **Configure Testnet Helm Values:** Author `infra/k8s/environments/testnet/values.yaml` configuring target namespace `nbse-testnet`, Demo Engine enablement (`demoEngine.enabled=true`), Testnet Besu RPC `https://testnet.besu.nbse.in`, Chain ID `13371`, and PostgreSQL database `nbse_testnet_db`.
4. **Configure Mainnet Helm Values:** Author `infra/k8s/environments/mainnet/values.yaml` configuring target namespace `nbse-mainnet`, Demo Engine disablement (`demoEngine.enabled=false`), Mainnet Besu RPC `https://mainnet.besu.nbse.in`, Chain ID `2026`, and PostgreSQL database `nbse_mainnet_db`.
5. **Implement Istio PeerAuthentication Manifests:** Create `infra/k8s/environments/policies/peer-authentication.yaml` enforcing `STRICT` mTLS across both `nbse-testnet` and `nbse-mainnet` namespaces, requiring valid cryptographically signed SPIFFE workload certificates for all pod communications.
6. **Implement Istio AuthorizationPolicies:** Author `infra/k8s/environments/policies/istio-isolation-policy.yaml` explicitly rejecting any request originating from `cluster.local/ns/nbse-testnet/*` destined for any workload in `cluster.local/ns/nbse-mainnet/*`.
7. **Implement Cilium / Kubernetes NetworkPolicies:** Author `infra/k8s/environments/policies/cilium-network-policies.yaml` defining Layer 3/4 egress and ingress rules. Block all CIDR-level routing from `nbse-testnet` pods to Mainnet database, Redis, and HSM subnets.
8. **Configure Feeder Unidirectional Fanout:** Deploy Feeder (Prompt 272) in a dedicated shared ingress namespace (`nbse-marketdata`). Configure Kafka producer ACLs allowing Feeder to publish ticks to `marketdata.feeder.ticks.v1`, and grant read-only consumer ACLs to both `nbse-mainnet` (Order Matching Engine) and `nbse-testnet` (Demo Engine).
9. **Implement Spot Order Service Environment Dispatcher:** In Spot Order Service (Prompt 275), configure dynamic gRPC client routing. Extract environment claim from incoming request context: route `testnet` requests to `demo-engine.nbse-testnet.svc.cluster.local:50051` and `mainnet` requests to `risk-service.nbse-mainnet.svc.cluster.local:50051`.
10. **Deploy Dedicated Redis Clusters:** Deploy Redis Cluster in `nbse-testnet` with service name `redis-cluster.nbse-testnet.svc` and in `nbse-mainnet` with service name `redis-cluster.nbse-mainnet.svc`. Enforce strict password authentication and namespace isolation.
11. **Deploy Isolated PostgreSQL Databases & Roles:** Configure PostgreSQL instances with dedicated users `testnet_app` (with permissions limited exclusively to `nbse_testnet_db`) and `mainnet_app` (with permissions limited exclusively to `nbse_mainnet_db`). Validate that `testnet_app` cannot list or read `nbse_mainnet_db`.
12. **Configure External Secrets Operator (ESO):** Deploy ESO syncing secrets from HashiCorp Vault. Ensure Vault path `secret/nbse/testnet/*` is synced only to Kubernetes namespace `nbse-testnet`, and `secret/nbse/mainnet/*` is synced only to `nbse-mainnet`.
13. **Configure Hyperledger Besu RPC Gateway Endpoints:** Deploy Envoy-based internal egress gateways in each namespace. Configure Testnet egress gateway to route JSON-RPC traffic to `https://testnet.besu.nbse.in` and Mainnet egress gateway to `https://mainnet.besu.nbse.in`.
14. **Build Automated Isolation Verification Harness:** Author integration verification script `infra/k8s/environments/tests/verify_isolation.sh`. The test suite attempts cross-namespace port connections, executes simulated SQL injection against Mainnet databases from Testnet pods, asserts Feeder tick delivery in both environments, and validates EIP-155 chain ID rejections.
15. **Integrate Prometheus & Grafana Monitoring:** Deploy Prometheus ServiceMonitors tracking dual-pipeline metrics: tick ingestion latency, Demo Engine execution latency, Mainnet Matching Engine WAL write latency, and zero-tolerance alerts for any cross-namespace network policy violation.

## Interfaces / Contracts

### 1. Helm Values Specification: Testnet vs Mainnet

```yaml
# infra/k8s/environments/testnet/values.yaml
global:
  environment: testnet
  clusterDomain: cluster.local
  region: ap-south-1

network:
  namespace: nbse-testnet
  istio:
    mtlsMode: STRICT
    allowedIngressNamespaces:
      - nbse-ingress
      - nbse-marketdata

services:
  spotOrderService:
    replicaCount: 3
    env:
      NBSE_ENVIRONMENT: testnet
      DEMO_ENGINE_GRPC_ENDPOINT: "demo-engine.nbse-testnet.svc.cluster.local:50051"
      MAINNET_ROUTING_ALLOWED: "false"

  demoEngine:
    enabled: true
    replicaCount: 2
    env:
      NBSE_ENVIRONMENT: testnet
      MARKET_DATA_FEEDER_TOPIC: "marketdata.feeder.ticks.v1"
      KAFKA_BOOTSTRAP_SERVERS: "kafka.nbse-marketdata.svc.cluster.local:9092"
      DATABASE_URL_SECRET: "testnet-db-credentials"
      REDIS_ENDPOINT: "redis-cluster.nbse-testnet.svc.cluster.local:6379"

  blockchain:
    rpcUrl: "https://testnet.besu.nbse.in"
    chainId: 13371
    vaultKeyPath: "secret/nbse/testnet/relayer-key"
```

```yaml
# infra/k8s/environments/mainnet/values.yaml
global:
  environment: mainnet
  clusterDomain: cluster.local
  region: ap-south-1

network:
  namespace: nbse-mainnet
  istio:
    mtlsMode: STRICT
    allowedIngressNamespaces:
      - nbse-ingress
      - nbse-marketdata

services:
  spotOrderService:
    replicaCount: 5
    env:
      NBSE_ENVIRONMENT: mainnet
      RISK_SERVICE_GRPC_ENDPOINT: "risk-engine.nbse-mainnet.svc.cluster.local:50051"
      MATCHING_ENGINE_GRPC_ENDPOINT: "matching-engine.nbse-mainnet.svc.cluster.local:50052"
      DEMO_ENGINE_ALLOWED: "false"

  demoEngine:
    enabled: false

  matchingEngine:
    enabled: true
    replicaCount: 2
    env:
      NBSE_ENVIRONMENT: mainnet
      MARKET_DATA_FEEDER_TOPIC: "marketdata.feeder.ticks.v1"
      KAFKA_BOOTSTRAP_SERVERS: "kafka.nbse-marketdata.svc.cluster.local:9092"
      DATABASE_URL_SECRET: "mainnet-db-credentials"
      REDIS_ENDPOINT: "redis-cluster.nbse-mainnet.svc.cluster.local:6379"

  blockchain:
    rpcUrl: "https://mainnet.besu.nbse.in"
    chainId: 2026
    hsmEndpoint: "cloudhsm.ap-south-1.amazonaws.com"
```

### 2. Istio Network Isolation & Authorization Policy

```yaml
# infra/k8s/environments/policies/istio-isolation-policy.yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default-strict-mtls
  namespace: nbse-mainnet
spec:
  mtls:
    mode: STRICT
---
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: block-testnet-cross-traffic
  namespace: nbse-mainnet
spec:
  action: DENY
  rules:
    - from:
        - source:
            principals: ["cluster.local/ns/nbse-testnet/*"]
---
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-ingress-and-marketdata-only
  namespace: nbse-mainnet
spec:
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
              - "cluster.local/ns/nbse-mainnet/*"
              - "cluster.local/ns/nbse-ingress/*"
              - "cluster.local/ns/nbse-marketdata/*"
```

### 3. Cilium Layer 3/4 Egress & Ingress Boundary Policy

```yaml
# infra/k8s/environments/policies/cilium-network-policies.yaml
apiVersion: "cilium.io/v2"
kind: CiliumNetworkPolicy
metadata:
  name: testnet-isolation-policy
  namespace: nbse-testnet
spec:
  endpointSelector:
    matchLabels: {}
  ingress:
    - fromEndpoints:
        - matchLabels:
            io.kubernetes.pod.namespace: nbse-testnet
        - matchLabels:
            io.kubernetes.pod.namespace: nbse-ingress
  egress:
    - toEndpoints:
        - matchLabels:
            io.kubernetes.pod.namespace: nbse-testnet
        - matchLabels:
            io.kubernetes.pod.namespace: nbse-marketdata
    - toCIDR:
        - 10.100.0.0/16 # Testnet Database & Redis Subnet
    - toEntities:
        - cluster
        - world # Permitted outbound for external Testnet Besu RPC
    - toEndpoints:
        - matchLabels:
            io.kubernetes.pod.namespace: nbse-mainnet
      # Explicitly denied: No rule matching nbse-mainnet endpoints
```

### 4. Market Data Feeder Tick Contract (JSON Schema)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "FeederMarketTickEvent",
  "description": "Normalized market data tick published by Feeder (Prompt 272) consumed by both Testnet Demo Engine and Mainnet Matching Engine",
  "type": "object",
  "required": [
    "specversion",
    "type",
    "source",
    "id",
    "time",
    "datacontenttype",
    "data"
  ],
  "properties": {
    "specversion": { "type": "string", "const": "1.0" },
    "type": { "type": "string", "const": "in.nbse.marketdata.tick.v1" },
    "source": { "type": "string", "const": "/services/feeder" },
    "id": { "type": "string", "format": "uuid" },
    "time": { "type": "string", "format": "date-time" },
    "datacontenttype": { "type": "string", "const": "application/json" },
    "data": {
      "type": "object",
      "required": [
        "symbol",
        "sequence_id",
        "bid_price",
        "bid_quantity",
        "ask_price",
        "ask_quantity",
        "last_price",
        "timestamp_ns"
      ],
      "properties": {
        "symbol": { "type": "string", "enum": ["BTC-USDT", "ETH-USDT", "NBSE-RELIANCE-INR", "NBSE-TATA-INR"] },
        "sequence_id": { "type": "integer", "minimum": 1 },
        "bid_price": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
        "bid_quantity": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
        "ask_price": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
        "ask_quantity": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
        "last_price": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
        "timestamp_ns": { "type": "integer", "minimum": 0 }
      }
    }
  }
}
```

### 5. Order Routing Context Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "OrderRoutingContext",
  "description": "Internal routing envelope evaluated by Spot Order Service (Prompt 275)",
  "type": "object",
  "required": [
    "order_id",
    "user_id",
    "environment",
    "symbol",
    "side",
    "order_type",
    "quantity",
    "price",
    "created_at"
  ],
  "properties": {
    "order_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "environment": { "type": "string", "enum": ["testnet", "mainnet"] },
    "symbol": { "type": "string" },
    "side": { "type": "string", "enum": ["BUY", "SELL"] },
    "order_type": { "type": "string", "enum": ["LIMIT", "MARKET"] },
    "quantity": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
    "price": { "type": "string", "pattern": "^[0-9]+(\\.[0-9]{1,8})?$" },
    "created_at": { "type": "string", "format": "date-time" }
  }
}
```

## Security & Compliance Notes
- **Zero Cross-Contamination Invariant:** Under no condition may a testnet transaction mutate mainnet state, trigger real bank debits, or modify production order books. This invariant is guaranteed at three concentric layers:
  1. *Network Layer:* Istio mTLS and Cilium NetworkPolicies drop all packets between `nbse-testnet` and `nbse-mainnet`.
  2. *Application Layer:* Spot Order Service asserts that requests with `environment: "testnet"` cannot access mainnet gRPC stubs or database connections.
  3. *Persistence Layer:* Completely distinct database instances with non-overlapping database user credentials and isolated Redis clusters.
- **Unidirectional Market Data Diode Pattern:** The Feeder (Prompt 272) operates as a software data diode. It writes tick data to the shared Kafka topic `marketdata.feeder.ticks.v1`. Testnet consumers have read-only access (`Describe`, `Read`) enforced via Kafka SASL/SCRAM ACLs. Testnet workloads have no produce permissions on market data topics, preventing any simulated or spoofed ticks from reaching Mainnet.
- **Strict Database Role Separation:**
  - `testnet_app_user` has `ALL PRIVILEGES` on database `nbse_testnet_db`, but no `CONNECT` privilege on `nbse_mainnet_db`.
  - `mainnet_app_user` has `ALL PRIVILEGES` on database `nbse_mainnet_db`, but no `CONNECT` privilege on `nbse_testnet_db`.
  - Database administrator accounts are held in separate HashiCorp Vault enclaves with distinct access policies and audit logging.
- **Cryptographic Key Isolation & HSM Boundaries:**
  - Testnet relayer and operator keys are stored in software Vault instances and are explicitly tagged as ephemeral/sandbox keys.
  - Mainnet signing keys are permanently held inside FIPS 140-2 Level 3 Hardware Security Modules (AWS CloudHSM) requiring multi-party quorum activation, physically inaccessible to testnet pipelines or developer environments.
- **EIP-155 Replay Protection:** Hyperledger Besu enforces strict EIP-155 Chain IDs (`13371` vs `2026`). Any raw transaction signed for Testnet will fail consensus verification if submitted to Mainnet nodes due to invalid elliptic curve recovery identifiers ($v$).
- **Regulatory Disclosure & Paper Trading Disclaimers:** In accordance with SEBI guidelines on simulated trading platforms (SEBI Master Circular for Stock Brokers), all Testnet API responses and UI headers must include statutory headers: `X-NBSE-Paper-Trading: TRUE` and `X-NBSE-Simulated-Environment: Sandbox-ChainID-13371`, explicitly disclaiming real execution or financial value.

## Acceptance Criteria
- [ ] Helm charts in `infra/k8s/environments/` successfully deploy isolated matching pipelines into `nbse-testnet` and `nbse-mainnet` namespaces without configuration drift.
- [ ] Istio `PeerAuthentication` enforces `STRICT` mTLS across all pods in both `nbse-testnet` and `nbse-mainnet`.
- [ ] Istio `AuthorizationPolicy` explicitly blocks all lateral traffic between `nbse-testnet` and `nbse-mainnet`, returning HTTP 403 / RBAC access denied.
- [ ] Cilium NetworkPolicies drop all direct IP packets routed from `nbse-testnet` pods to Mainnet PostgreSQL, Redis, and Besu node IP addresses.
- [ ] Feeder (Prompt 272) successfully broadcasts live BTC/USDT and ETH/USDT tick events to `marketdata.feeder.ticks.v1`, which are consumed concurrently by Demo Engine (Prompt 273) and Mainnet Matching Engine (Prompt 205).
- [ ] Testnet consumers attempting to publish messages to `marketdata.feeder.ticks.v1` are rejected by Kafka ACLs with `TopicAuthorizationException`.
- [ ] Spot Order Service (Prompt 275) routes orders with `environment: "testnet"` exclusively to Demo Engine and orders with `environment: "mainnet"` exclusively to Mainnet Matching Engine.
- [ ] Demo Engine executes paper trades against live market ticks, updating virtual balances in `nbse_testnet_db` without triggering real custodial ledger entries.
- [ ] PostgreSQL role `testnet_app_user` is provably denied connection access to `nbse_mainnet_db`.
- [ ] Redis Cluster keys are strictly isolated, verifying that virtual user balances in `redis-cluster.nbse-testnet.svc` cannot be queried or updated from Mainnet services.
- [ ] Hyperledger Besu RPC endpoint for Testnet (`https://testnet.besu.nbse.in`) responds with Chain ID `13371`, and Mainnet (`https://mainnet.besu.nbse.in`) responds with Chain ID `2026`.
- [ ] Automated isolation test script (`infra/k8s/environments/tests/verify_isolation.sh`) executes during CI/CD and passes with zero network leakage or cross-environment read/write violations.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 101 (System Architecture Overview & C4 Model)
  - Prompt 104 (Event Schema & Kafka Topic Standards)
  - Prompt 105 (Authentication & Authorization Architecture)
  - Prompt 108 (Environment Strategy, Configuration Management & 12-Factor App Architecture)
  - Prompt 205 (High-Performance Order Matching Engine)
  - Prompt 272 (External Market Data Feeder & Tick Ingestion Service)
  - Prompt 273 (Virtual Matching Demo Engine & Paper Trading Simulator)
  - Prompt 275 (Spot Order Service & Pre-Trade Routing Engine)
  - Prompt 301 (Permissioned Blockchain Architecture & QBFT Configuration)
  - Prompt 802 (Kubernetes Cluster Architecture & Core Infrastructure)
  - Prompt 812 (Dual-Environment Testnet Sandbox, Mainnet Isolation & Orchestration Suite)
  - Prompt 815 (Zero-Trust Service Mesh SPIFFE/SPIRE Workload Identity)
- **Downstream & Parallel Dependencies:**
  - Prompt 527 (Flutter Environment Switcher & Sandbox Mode)
  - Prompt 609 (Developer Portal & Interactive Sandbox Faucet)
  - Prompt 804 (GitOps Progressive Canary Rollout Pipeline with ArgoCD)
  - Prompt 811 (Dual-Environment GitOps CI/CD Pipeline: Testnet vs Mainnet)
  - Prompt 908 (Production Launch Runbook & Operational Go-Live Checklist)
