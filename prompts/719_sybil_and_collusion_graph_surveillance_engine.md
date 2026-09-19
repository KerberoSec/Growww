# 719 - Sybil Ring & Cross-Account Collusion Graph Surveillance Engine (Python / Neo4j)

## Purpose
Establishes an institutional-grade graph intelligence, behavioral topology, and network analysis microservice pipeline (`services/graph-surveillance-engine`). The engine is engineered to detect, de-anonymize, and dismantle complex market abuse syndicates operating across ostensibly independent user accounts. Specifically, it targets coordinated wash trading rings, pump-and-dump cartels, circular transaction loops, marking the close/opening schemes, synchronized quote manipulation, and front-running cartels.

Operating under the statutory mandate of the **Securities and Exchange Board of India (Prohibition of Fraudulent and Unfair Trade Practices relating to Securities Market) Regulations, 2003 (PFUTP)**, the **SEBI Master Circular on Surveillance of Securities Markets**, the **Prevention of Money Laundering Act, 2002 (PMLA)**, and the **IFSCA (Capital Market Intermediaries) Regulations**, market infrastructure institutions and exchange operators must proactively identify non-genuine transactions where beneficial ownership does not change, as well as manipulative rings artificially inflating trading volumes or distorting price discovery.

Modern market manipulation syndicates rarely execute illicit strategies from a single trading account. Instead, sophisticated collusion rings orchestrate distributed operations across dozens or hundreds of Sybil accounts-utilizing mule accounts, family kinship networks, shell corporate LLPs, shared VPN/residential proxy infrastructures, common funding seed wallets, and millisecond-synchronized order submission latencies. Traditional rule-based, single-account alert engines fail to correlate these distributed signals. This engine constructs dynamic bipartite trade graphs, multi-layer identity graphs, and temporal transaction hypergraphs, applying cycle detection, community clustering, and Heterogeneous Graph Neural Networks (HGNNs) to expose coordinated market manipulation rings in real time.

## What You Are Building
A distributed, high-throughput graph analytics pipeline (`services/graph-surveillance-engine`) designed for continuous real-time stream processing and deep historical batch network forensics:
- **Dynamic Bipartite & Multi-Layer Graph Ingestion Engine:** Consumes real-time trade matching streams, order state changes, user identity graphs, and authentication sessions. It maps entities into a multi-relational property graph consisting of `Trader`, `Order`, `Trade`, `Device`, `IPSubnet`, `BankAccount`, and `BesuWallet` nodes connected via temporal, weighted edges.
- **Real-Time Circular Transaction & Directed Cycle Detector:** Implements streaming cycle traversal algorithms (Johnson's circuit finding algorithm, Tarjan's Strongly Connected Components, and bounded-depth DFS) over temporal trade directed graphs to detect circular wash trading ($A \to B \to C \to \dots \to A$) within sliding time horizons ($10\text{s}, 60\text{s}, 15\text{m}, 24\text{h}$) where asset volume transfers cyclically with negligible change in net beneficial ownership.
- **Sybil Community & Clique Detection Engine:** Executes distributed community detection (Louvain, Leiden, Infomap) and topological clustering across multi-attributed relationship graphs to identify tightly knit account clusters sharing common hardware fingerprints, IP subnets, bank payment origins, or rapid mutual fund transfers.
- **Synchronized Order Placement & Cartel Correlator:** Employs temporal point processes and cross-correlation time-series metrics (Dynamic Time Warping, Pearson cross-correlation, and mutual information) to flag statistically impossible synchronization across accounts submitting matching bids, asks, and order cancellations within microsecond windows.
- **Heterogeneous Graph Neural Network (HGNN) Inference Engine:** Deploys trained PyTorch Geometric Relational Graph Convolutional Networks (RGCN) and Graph Attention Networks (GAT) to compute node embeddings and edge-existence probabilities, scoring the likelihood that any pair or group of accounts belongs to a covert collusive syndicate.
- **Hyperledger Besu On-Chain Transfer Subgraph Forensics:** Continuously indexes private consortium ledger transactions on Hyperledger Besu, constructing wallet transfer subgraphs to correlate off-chain trading collusion with on-chain token movements, liquidity pool wash trades, and common gas funding distribution trees.
- **CollusionAlert & Automated Containment Dispatcher:** Emits standardized, high-priority `CollusionAlert` events to Kafka, triggers automated pre-trade containment blocks via the Real-Time Market Surveillance Engine (Prompt 228) and Risk Engine (Prompt 206), and publishes immutable audit evidence to the Audit Log Service (Prompt 218).
- **Interactive Forensic Graph Visualizer & Dossier Generator:** Exposes gRPC/GraphQL APIs for compliance officers to traverse $k$-hop collusion subgraphs, inspect chronological trade loops, and generate SEBI PFUTP-compliant evidentiary dossiers.

## Scope Boundaries
- **In Scope:**
  - Streaming ingestion and continuous graph indexing of matched trade executions, order cancellations, account logins, device fingerprints, and deposit/withdrawal events.
  - Real-time directed cycle detection on directed trade graphs for cycles of length $k \in [2, 8]$ within sliding time windows.
  - Computation of graph centrality metrics (PageRank, betweenness centrality, core number) and community partition coefficients.
  - Multi-layer identity graph resolution linking PAN hashes, bank accounts, device UUIDs, TLS client fingerprints, and IP prefixes.
  - Heterogeneous Graph Neural Network (HGNN) inference using PyTorch Geometric for automated collusion scoring.
  - Hyperledger Besu on-chain transaction subgraph extraction and cross-layer correlation with off-chain trading accounts.
  - Generation of structured `CollusionAlert` payloads containing subgraph topologies, participant account hashes, trade volume deltas, and statistical confidence scores.
  - PostgreSQL schema for case management, alert lifecycle tracking, and historical audit indices.
- **Out of Scope / Handled Elsewhere:**
  - Single-account order book micro-manipulation such as single-trader spoofing, quote stuffing, and layering (handled in Prompt 711 & Prompt 228).
  - Corporate insider trading, Unpublished Price Sensitive Information (UPSI) tracking, and Structured Digital Database (SDD) maintenance (handled in Prompt 712).
  - Public blockchain cross-chain forensic tracing across Bitcoin, Ethereum mainnet, and Solana mixers (handled in Prompt 714).
  - Low-latency order matching, price determination, and continuous limit order book execution (handled in Prompt 205).
  - Pre-trade margin calculations, cash balance checks, and collateral reservations (handled in Prompt 206).
  - Statutory quarterly regulatory filing dispatch to regulatory portals (handled in Prompt 216 & Prompt 715).

## Technology to Use
- **Core Runtime & Machine Learning Framework:**
  - Python 3.11+ leveraging asynchronous event loops (`asyncio`, `uvloop`), optimized NumPy/Numba routines for numerical acceleration, and `pydantic` v2 for strict runtime data validation.
  - **PyTorch 2.3+ & PyTorch Geometric (PyG):** Heterogeneous Graph Neural Networks (RGCN, HeteroGAT) for semi-supervised Sybil node classification and edge-link prediction.
- **Graph Engines & In-Memory Analytics Libraries:**
  - **Neo4j Enterprise 5+ / Memgraph:** High-performance property graph database utilizing the Cypher query language, APOC (Awesome Procedures on Cypher), and Graph Data Science (GDS) library for persistent relationship mapping and exploratory investigator queries.
  - **Graph-Tool & NetworkX:** `graph-tool` (C++ Boost Graph Library Python wrapper) for ultra-fast in-memory directed cycle searches, strongly connected components (SCC), and network flow algorithms; `NetworkX` for rapid ad-hoc topological manipulation.
  - **Apache Spark 3.5+ with GraphX / GraphFrames:** Distributed graph analytics processing batch historical graphs over parquet/delta lakes for periodic macro-surveillance sweeps and community evolution analysis across millions of accounts.
- **Streaming & Data Storage Infrastructure:**
  - **Apache Kafka 3.7+:** Distributed streaming backbone (using `confluent-kafka` with librdkafka bindings) for consuming trade, order, and telemetry streams.
  - **PostgreSQL 16 with TimescaleDB & pgvector:** Stores persistent alert state, case workflows, trader feature vectors, and graph embedding representations with Row-Level Security (RLS).
  - **Redis 7.2 Cluster:** Low-latency sliding window cache storing active edge buffers, sub-millisecond account cross-indexes, and synchronized order counter keys.
- **Justification:** Python 3.11+ provides the ideal bridge between streaming graph algorithms and modern Deep Learning graph models (PyG). While Cypher on Neo4j enables complex multi-hop pattern queries and intuitive visualization for compliance officers, in-memory execution via `graph-tool` ensures sub-second cycle detection on high-velocity streaming trades without the I/O overhead of round-tripping to persistent disk.

## Backend / Infra Touchpoints
- **Upstream Microservices & Ingestion Feeds:**
  - `services/order-matching-engine` (Prompt 205): Real-time trade match events (`market.trades.matched.v1`) capturing matched buyer ID, seller ID, execution price, quantity, timestamp, aggressive/passive flags, and order IDs.
  - `services/order-service` (Prompt 204): Stream of order lifecycle events (`order.placed.v1`, `order.cancelled.v1`, `order.modified.v1`) to detect synchronized placement and cancellation patterns.
  - `services/kyc-aml-service` (Prompt 202): Salted identity metadata (PAN hash, bank account hashes, nominee structures, CKYC reference hashes).
  - `services/api-gateway` & `services/rate-limiting` (Prompts 219 & 220): HTTP/WebSocket session telemetry, device UUIDs, canvas/WebGL fingerprints, JA4 TLS fingerprints, and source IP CIDRs.
  - `services/wallet-account-service` (Prompt 203): Fiat and token deposit/withdrawal records, internal ledger transfers, and payment settlement references.
  - `services/hyperledger-besu-indexer`: Block receipts, ERC-20/1400 token transfers, and atomic settlement calls from the consortium ledger.
- **Downstream Microservices & Consumers:**
  - `services/surveillance-engine` (Prompt 228 - Real-Time Market Surveillance & Dynamic Volatility Engine): Receives real-time `CollusionAlert` events to dynamically adjust surveillance thresholds, flag correlated order books, and adjust symbol volatility parameters.
  - `services/risk-engine` (Prompt 206): Consumes containment signals to apply automated pre-trade trading blocks, liquidate collateral, or enforce maximum position limits on flagged Sybil rings.
  - `services/audit-log-service` (Prompt 218): Ingests immutable forensic records of all collusion detections, model weights, graph snapshots, and automated intervention actions with cryptographic chaining.
  - `services/regulatory-reporting-service` (Prompts 216 & 715): Aggregates collusion dossiers for automated filing of Suspicious Transaction Reports (STRs) to SEBI and the Financial Intelligence Unit - India (FIU-IND).
- **Messaging Topics (Apache Kafka):**
  - Consumes:
    - `market.trades.matched.v1`
    - `market.orders.lifecycle.v1`
    - `user.auth.session_logged.v1`
    - `wallet.transfers.completed.v1`
    - `besu.blocks.mined.v1`
  - Publishes:
    - `surveillance.collusion.alert.v1`
    - `surveillance.sybil_ring.detected.v1`
    - `risk.account.restrict_trading.v1`

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consortium Ledger Architecture:** Operates against the Growww permissioned Hyperledger Besu network utilizing QBFT (Quorum Byzantine Fault Tolerant) consensus with dedicated validator nodes run by regulated entities.
- **On-Chain Wallet Transfer Subgraph Analysis:**
  - Ingests all smart contract events from `NBSESettlementDvP.sol` (Prompt 329), `CommodityToken1400.sol` (Prompt 330), and escrow contracts to construct directed asset transfer graphs on-chain.
  - Maps wallet addresses ($W_i$) to off-chain exchange trading accounts using cryptographically blinded account link keys.
  - Traverses on-chain transaction flows to detect circular token wash transfers ($W_1 \to W_2 \to W_3 \to W_1$) executed to artificially inflate on-chain token velocity or manipulate automated market maker (AMM) pricing.
  - Traces common funding sources ("feeder wallets") that distribute native gas tokens or settlement tokens to multiple Sybil wallets prior to coordinated market activities.
- **Cross-Layer Market Manipulation Correlation:**
  - Correlates on-chain wallet activity with off-chain order execution books. Identifies cross-layer collusion schemes where accounts trade synthetic or tokenized securities off-chain while executing offsetting or manipulative transfers on the Besu ledger.
- **Cryptographic Attestation & Evidence Anchoring:**
  - Computes a deterministic SHA-256 Merkle root hash of every detected collusion subgraph, its constituent trade IDs, and mathematical confidence metrics.
  - Commits the Merkle root and investigation metadata to `CollusionSurveillanceRegistry.sol` on Hyperledger Besu via an authorized transaction signed by the surveillance daemon's CloudHSM key (Prompt 717).
  - This on-chain commitment provides tamper-evident, non-repudiable legal proof of detection time, preserving chain of custody for SEBI PFUTP adjudication.
- **Zero-PII On-Chain Compliance:**
  - In strict compliance with the **Digital Personal Data Protection Act (DPDP Act 2023)**, zero personally identifiable information (PII) is stored or referenced on-chain. All node identifiers, wallet addresses, and account numbers are anonymized via keyed cryptographic hashes: $\text{node\_hash} = \text{HMAC-SHA256}(\text{account\_id}, \text{secret\_salt})$.

## Step-by-Step Build Instructions

### Step 1: Project Scaffolding & Environment Setup
- Initialize the microservice directory structure under `services/graph-surveillance-engine`:
  ```
  services/graph-surveillance-engine/
  ├── cmd/
  │   ├── engine/          # Main streaming daemon entrypoint
  │   └── worker/          # Batch Spark/analytics worker
  ├── config/              # Configuration files and YAML definitions
  ├── proto/               # Protobuf source definitions
  ├── src/
  │   ├── ingestion/       # Kafka consumers and schema validators
  │   ├── graph/           # In-memory graph-tool & Neo4j graph managers
  │   ├── algorithms/      # Cycle detection, community discovery, synchrony
  │   ├── gnn/             # PyTorch Geometric models, training, inference
  │   ├── besu/            # Hyperledger Besu on-chain subgraph indexer
  │   ├── alerts/          # Alert scoring, deduplication, Kafka producers
  │   ├── api/             # gRPC and GraphQL query endpoints
  │   └── storage/         # PostgreSQL/TimescaleDB & Redis adapters
  ├── tests/
  └── Dockerfile
  ```
- Configure Python 3.11 virtual environment with strict dependency locking via Poetry / Pipenv, installing `torch`, `torch-geometric`, `graph-tool`, `neo4j`, `networkx`, `confluent-kafka`, `web3.py`, `asyncpg`, `pydantic`, and `redis`.

### Step 2: Protobuf Definition & Contract Compilation
- Define the formal gRPC service and streaming event definitions in `proto/growww/surveillance/graph/v1/collusion_surveillance.proto`.
- Generate Python gRPC stubs, Pydantic data schemas, and event serialization schemas using `protoc` and `grpc_tools.protoc`.
- Verify bi-directional schema backwards-compatibility for event ingestion from the Order Matching Engine and Order Service.

### Step 3: Neo4j Graph Database Schema & Index Constraints
- Establish Neo4j 5+ database connectivity with connection pooling, retry policies, and APOC / GDS library initialization.
- Execute Cypher schema constraints to ensure indexing and uniqueness across all entity nodes:
  - Unique constraint on `(:Account {account_hash})`
  - Unique constraint on `(:Device {device_fingerprint})`
  - Unique constraint on `(:BankAccount {account_hash})`
  - Unique constraint on `(:IPSubnet {cidr})`
  - Unique constraint on `(:BesuWallet {address_hash})`
  - Range and text indexes on relationship attributes (`traded_at`, `volume`, `isin`, `confidence_score`).

### Step 4: High-Throughput Streaming Ingestion Pipeline
- Build asynchronous multi-topic Kafka consumers using `confluent-kafka` with consumer groups assigned to partitions of `market.trades.matched.v1`, `market.orders.lifecycle.v1`, and `user.auth.session_logged.v1`.
- Implement an in-memory ring buffer and Redis-backed sliding window accumulator ($T \in [5\text{s}, 86400\text{s}]$) to ingest trades at rates exceeding $50,000$ trades/second with sub-millisecond serialization overhead.
- Transform incoming events into directed edge tuples:
  $$\vec{e} = (u, v, t, q, p, \text{isin}, \text{order\_ids})$$
  where $u$ is the aggressive buyer/seller, $v$ is the passive counterparty, $t$ is the exchange match timestamp, $q$ is quantity, and $p$ is match price.

### Step 5: In-Memory Dynamic Bipartite & Multi-Graph Formulation
- Implement dual-graph representations:
  - **Trade Directed Multi-Graph ($G_{\text{trade}}$):** Nodes represent trading accounts; directed edges represent trade executions pointing from seller to buyer (or vice-versa depending on asset flow direction), weighted by traded volume and execution value.
  - **Identity & Association Heterogeneous Hypergraph ($G_{\text{identity}}$):** Multi-modal graph linking accounts to shared physical or legal attributes:
    $$\text{Account} \xleftrightarrow{\text{AUTHENTICATED\_FROM}} \text{Device}$$
    $$\text{Account} \xleftrightarrow{\text{LOGGED\_IP}} \text{IPSubnet}$$
    $$\text{Account} \xleftrightarrow{\text{FUNDED\_VIA}} \text{BankAccount}$$
    $$\text{Account} \xleftrightarrow{\text{SETTLED\_TO}} \text{BesuWallet}$$
- Use C++ optimized `graph-tool` dynamic graph objects in memory, periodically pruning edges older than the maximum surveillance lookback window ($24\text{h}$) while syncing persistent subgraphs to Neo4j.

### Step 6: Real-Time Circular Wash Trading & Directed Cycle Detection
- Implement streaming cycle detection over $G_{\text{trade}}$ for every newly added trade edge $(u, v)$:
  - Identify Strongly Connected Components (SCC) using Tarjan’s algorithm within the relevant asset subgraph.
  - For SCCs with $\ge 2$ nodes, execute depth-bounded Johnson’s elementary cycle finding algorithm to detect all simple cycles $C = (v_1, v_2, \dots, v_k, v_1)$ of length $2 \le k \le 8$.
  - Calculate cycle volume conservation and price delta metrics:
    $$\Delta V_{\text{net}} = \sum_{i=1}^{k} \left| q_{v_i \to v_{i+1}} - \bar{q} \right|$$
    $$\text{Wash Ratio} = 1.0 - \frac{|\text{PositionDelta}(v_i)|}{\sum \text{Turnover}(v_i)}$$
  - If Wash Ratio $\ge 0.90$ and the cycle completes within window $\Delta t \le 1800\text{s}$, tag the cycle as an active circular wash trading ring.

### Step 7: Pump-and-Dump & Synchronized Trading Syndicate Detection
- Construct temporal accumulation/distribution volume curves for illiquid or mid-cap securities across all accounts.
- Calculate cross-account order synchrony scores:
  - For each pair of accounts $(A, B)$ trading the same ISIN, calculate the time delta between order submissions:
    $$\delta t_{ij} = |t_{A, i} - t_{B, j}|$$
  - Compute the Synchronized Order Ratio (SOR):
    $$\text{SOR}(A, B) = \frac{|\{ (i, j) : \delta t_{ij} < 50\text{ms} \land \text{side}_A \neq \text{side}_B \land |p_A - p_B| \le \epsilon \}|}{\min(N_A, N_B)}$$
  - High SOR values between accounts with zero formal legal connection indicate pre-arranged, synchronized trading rings designed to create artificial market depth and ignite price momentum.

### Step 8: Multi-Modal Identity Resolution & Sybil Clique Discovery
- Ingest identity telemetry events to materialize edges in $G_{\text{identity}}$.
- Execute community discovery algorithms:
  - Run the Louvain and Leiden community detection algorithms with modularity optimization on the projected account-account affinity matrix.
  - Calculate Sybil Affinity Score ($S_{\text{affinity}}$) combining shared device fingerprints (weight 0.40), identical /24 IP subnets with common user-agent hashes (weight 0.25), shared bank beneficiary IFSC/PAN hashes (weight 0.25), and synchronized trade executions (weight 0.10).
  - Extract dense subgraphs (cliques where edge density $\rho = \frac{2|E|}{|V|(|V|-1)} \ge 0.75$) as confirmed Sybil clusters.

### Step 9: Heterogeneous Graph Neural Network (HGNN) Inference Pipeline
- Design and load pre-trained PyTorch Geometric Relational Graph Convolutional Network (RGCN) with heterogeneous message passing:
  - Input features: Node degree, 30-day turnover, order cancellation ratio, mean holding duration, KYC risk score, device sharing degree.
  - Edge types: `TRADED_WITH`, `SHARED_DEVICE`, `SHARED_IP`, `SHARED_BANK`, `CHAIN_TRANSFER`.
  - Architecture: 3-layer RGCN with Graph Attention (GAT) layers and LeakyReLU activations, outputting a 64-dimensional latent embedding vector per account node.
- Compute cosine similarity and run an edge classification head to predict latent collusion edges between accounts that have intentionally avoided direct interaction in the trade book (e.g., $A$ trades with $B$, $C$ trades with $D$, but $A$ and $C$ share funding and hardware roots).
- Maintain an inference latency budget of $< 25\text{ms}$ per candidate subgraph using PyTorch TensorRT / ONNX Runtime.

### Step 10: Hyperledger Besu On-Chain Transfer Subgraph Indexer
- Connect to the permissioned Besu node via WebSockets (`web3.py`) with TLS client mutual authentication.
- Listen for `Transfer(address,address,uint256)` and `TradeSettled(bytes32,address,address,uint256,uint256)` events emitted by exchange settlement and token contracts.
- Ingest on-chain wallet transactions into the Besu subgraph:
  - Detect on-chain circular routing ($W_1 \to W_2 \to \dots \to W_1$) used to simulate token velocity or game Proof-of-Reserve validation.
  - Perform BFS traversal from funding contract addresses to detect Sybil wallet dispersal ("fan-out" from a single seed wallet) and subsequent consolidation ("fan-in" to a cash-out wallet).
  - Map on-chain wallet clusters to off-chain account IDs to uncover cross-layer manipulation schemes.

### Step 11: Multi-Dimensional Risk Scoring & CollusionAlert Event Engine
- Formulate the Composite Collusion Index ($\text{CCI} \in [0.0, 100.0]$):
  $$\text{CCI} = w_1 \cdot S_{\text{cycle}} + w_2 \cdot S_{\text{sybil\_affinity}} + w_3 \cdot S_{\text{synchrony}} + w_4 \cdot S_{\text{hgnn\_embedding}} + w_5 \cdot S_{\text{besu\_onchain}}$$
  where weights satisfy $\sum w_i = 1.0$ (calibrated default: $w_1=0.30, w_2=0.25, w_3=0.20, w_4=0.15, w_5=0.10$).
- When $\text{CCI} \ge 75.0$ (or statutory wash criteria are met deterministically), generate a cryptographically signed `CollusionAlert` event.
- Publish `CollusionAlert` to Kafka topic `surveillance.collusion.alert.v1` and persist the record, subgraph topology, and trace metrics to PostgreSQL/TimescaleDB.

### Step 12: Automated Containment & Inter-Service Mitigation Dispatch
- Integrate with `services/surveillance-engine` (Prompt 228) and `services/risk-engine` (Prompt 206):
  - When a severe collusion alert ($\text{CCI} \ge 85.0$ or active 2-party wash cycle) is generated, immediately issue a pre-trade restriction command to `risk.account.restrict_trading.v1`.
  - The Risk Engine instantaneously invalidates active session tokens, cancels outstanding resting limit orders, and transitions participant accounts to `RESTRICTED_LIQUIDATION_ONLY` status within $< 50\text{ms}$.
  - Dispatch audit record to `services/audit-log-service` (Prompt 218) with immutable payload checksum and HMAC signature.

### Step 13: Interactive Forensic Graph Visualizer & SEBI PFUTP Export Engine
- Expose gRPC and GraphQL query endpoints for internal compliance investigation tools:
  - Path traversal queries returning nodes, edges, weights, and trade timelines up to 5 hops.
  - Real-time Cytoscape.js / D3.js compatible JSON serialization of detected collusion rings.
- Implement an automated report generator assembling SEBI PFUTP Regulatory Investigation Dossiers:
  - Formatted PDF/JSON dossiers containing trade chronologies, price impact charts, counterparty volume tables, KYC entity summaries, and visual relationship graph diagrams.
  - Cryptographic attestation anchoring receipt from Hyperledger Besu included in the final dossier package.

### Step 14: Comprehensive Verification, Tuning & Load Testing
- Implement automated unit and integration tests covering:
  - Cycle detection accuracy over synthetic cyclical trade graphs (testing cycles of length 2, 3, 4, 5, and 8).
  - Graph partition and community detection robustness under edge perturbations.
  - HGNN inference performance and memory stability under continuous batch inference.
  - End-to-end latency benchmarks verifying alert generation within $< 100\text{ms}$ of the closing trade execution.
  - Resilience against Kafka partition rebalances and Neo4j transient network disconnections.

## Interfaces / Contracts

### Protobuf Service & Event Definitions (`collusion_surveillance.proto`)
```protobuf
syntax = "proto3";

package growww.surveillance.graph.v1;

option go_package = "github.com/growww/services/graph-surveillance-engine/gen/v1;graphsurveillancev1";
option py_generic_services = false;

service CollusionGraphSurveillanceService {
  // Queries subgraph neighborhoods for a suspicious account
  rpc GetAccountCollusionSubgraph (AccountSubgraphRequest) returns (AccountSubgraphResponse);
  
  // Real-time analysis of a candidate trade cycle or cluster
  rpc EvaluateTradeCycle (EvaluateTradeCycleRequest) returns (EvaluateTradeCycleResponse);
  
  // Triggers distributed community detection across a specific asset or market
  rpc RunCommunityDetection (CommunityDetectionRequest) returns (CommunityDetectionResponse);
  
  // Generates SEBI PFUTP regulatory inspection dossier
  rpc GenerateCollusionDossier (GenerateCollusionDossierRequest) returns (GenerateCollusionDossierResponse);
}

enum CollusionCategory {
  COLLUSION_CATEGORY_UNSPECIFIED = 0;
  CIRCULAR_WASH_TRADING = 1;
  SYNCHRONIZED_PUMP_AND_DUMP = 2;
  MARKING_THE_CLOSE = 3;
  FRONT_RUNNING_CARTEL = 4;
  SYBIL_MULE_RING = 5;
  ONCHAIN_CROSS_LAYER_MANIPULATION = 6;
}

enum AlertSeverity {
  ALERT_SEVERITY_UNSPECIFIED = 0;
  SEVERITY_INFO = 1;
  SEVERITY_LOW = 2;
  SEVERITY_MEDIUM = 3;
  SEVERITY_HIGH = 4;
  SEVERITY_CRITICAL = 5;
}

message GraphNode {
  string node_id = 1;                  // HMAC-SHA256 blinded entity hash
  string node_type = 2;                // ACCOUNT, DEVICE, IP_SUBNET, BANK_ACCOUNT, BESU_WALLET
  map<string, string> properties = 3;  // Non-PII metadata (e.g., account_age_days, risk_tier)
}

message GraphEdge {
  string edge_id = 1;
  string source_node_id = 2;
  string target_node_id = 3;
  string relationship_type = 4;        // TRADED_WITH, SHARED_DEVICE, TRANSFERRED_FUNDS, SAME_SUBNET
  double weight = 5;
  int64 timestamp_utc_ms = 6;
  map<string, string> metadata = 7;    // isin, trade_id, volume, price
}

message SubgraphData {
  repeated GraphNode nodes = 1;
  repeated GraphEdge edges = 2;
  double graph_density = 3;
  int32 cycle_length = 4;
}

message CollusionAlert {
  string alert_id = 1;
  CollusionCategory category = 2;
  AlertSeverity severity = 3;
  string symbol_isin = 4;
  repeated string participant_account_hashes = 5;
  repeated string involved_trade_ids = 6;
  
  double composite_collusion_score = 7; // 0.0 to 100.0
  double wash_trading_ratio = 8;        // 0.0 to 1.0 (proportion of volume with no beneficial ownership change)
  double total_manipulated_volume = 9;
  double total_turnover_inr = 10;
  
  int64 window_start_utc_ms = 11;
  int64 window_end_utc_ms = 12;
  int64 detected_at_utc_ms = 13;
  
  SubgraphData evidence_subgraph = 14;
  string besu_attestation_tx_hash = 15;
  string digital_signature = 16;
}

message AccountSubgraphRequest {
  string account_hash = 1;
  int32 max_hops = 2;                 // Recommended 1 to 4
  int64 lookback_seconds = 3;
}

message AccountSubgraphResponse {
  string root_account_hash = 1;
  SubgraphData subgraph = 2;
  double sybil_risk_score = 3;
}

message EvaluateTradeCycleRequest {
  repeated string trade_ids = 1;
  string symbol_isin = 2;
}

message EvaluateTradeCycleResponse {
  bool is_circular_cycle = 1;
  int32 cycle_hops = 2;
  double volume_retention_ratio = 3;
  double price_impact_bps = 4;
  CollusionAlert generated_alert = 5;
}

message CommunityDetectionRequest {
  string symbol_isin = 1;
  int64 start_time_ms = 2;
  int64 end_time_ms = 3;
  double min_affinity_threshold = 4;
}

message CommunityDetectionResponse {
  int32 detected_communities_count = 1;
  repeated SubgraphData suspected_cliques = 2;
}

message GenerateCollusionDossierRequest {
  string alert_id = 1;
  string compliance_officer_id = 2;
  string regulatory_jurisdiction = 3; // "SEBI_INDIA" or "IFSCA_GIFT_CITY"
}

message GenerateCollusionDossierResponse {
  string dossier_id = 1;
  string pdf_download_url = 2;
  string sha256_checksum = 3;
  string onchain_receipt_hash = 4;
  int64 generated_at_utc_ms = 5;
}
```

### Neo4j Cypher Graph Schema & Core Detection Queries
```cypher
// Schema Constraints & Uniqueness Indexes
CREATE CONSTRAINT unique_account_hash IF NOT EXISTS
FOR (a:Account) REQUIRE a.account_hash IS UNIQUE;

CREATE CONSTRAINT unique_device_fingerprint IF NOT EXISTS
FOR (d:Device) REQUIRE d.fingerprint IS UNIQUE;

CREATE CONSTRAINT unique_bank_hash IF NOT EXISTS
FOR (b:BankAccount) REQUIRE b.bank_hash IS UNIQUE;

CREATE CONSTRAINT unique_besu_wallet IF NOT EXISTS
FOR (w:BesuWallet) REQUIRE w.address_hash IS UNIQUE;

CREATE INDEX idx_trade_isin_time IF NOT EXISTS
FOR ()-[r:TRADED_WITH]-() ON (r.isin, r.timestamp_ms);

// Query 1: Detect Directed Circular Trading Cycles (Length 2 to 6) in Sliding Window
MATCH path = (a1:Account)-[r1:TRADED_WITH]->(a2:Account)-[r2:TRADED_WITH]->(a3:Account)-[r3:TRADED_WITH]->(a1:Account)
WHERE r1.isin = $target_isin
  AND r2.isin = $target_isin
  AND r3.isin = $target_isin
  AND r1.timestamp_ms >= $window_start_ms
  AND r3.timestamp_ms <= $window_end_ms
  AND a1 <> a2 AND a2 <> a3 AND a1 <> a3
WITH path, a1, a2, a3, [r1, r2, r3] AS trades,
     reduce(v = 0.0, r IN [r1, r2, r3] | v + r.quantity) AS total_cycle_volume,
     reduce(p = 0.0, r IN [r1, r2, r3] | p + (r.price * r.quantity)) AS total_turnover
RETURN a1.account_hash AS node_a,
       a2.account_hash AS node_b,
       a3.account_hash AS node_c,
       total_cycle_volume,
       total_turnover,
       [r IN trades | r.trade_id] AS trade_ids,
       length(path) AS cycle_length;

// Query 2: Multi-Hop Sybil Ring Detection via Shared Hardware, IP & Banking Roots
MATCH (target:Account {account_hash: $seed_account_hash})
MATCH path = (target)-[r1:SHARED_DEVICE|SHARED_IP|SHARED_BANK*1..3]-(peer:Account)
WHERE target <> peer
WITH peer, path, relationships(path) AS rels
MATCH (target)-[t:TRADED_WITH]-(peer)
WHERE t.timestamp_ms >= $lookback_start_ms
RETURN peer.account_hash AS colluding_peer,
       [rel IN rels | type(rel)] AS shared_modalities,
       count(t) AS direct_trades_count,
       sum(t.quantity) AS total_matched_shares,
       sum(t.price * t.quantity) AS total_matched_value_inr;
```

### PostgreSQL Case Management & Alert Audit Schema (`storage/schema.sql`)
```sql
-- PostgreSQL 16 Schema for Collusion Alert Auditing and Case Workflows
CREATE TABLE IF NOT EXISTS collusion_alerts (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(64) NOT NULL,
    severity VARCHAR(32) NOT NULL,
    symbol_isin VARCHAR(12) NOT NULL,
    composite_collusion_score NUMERIC(5, 2) NOT NULL,
    wash_trading_ratio NUMERIC(5, 4) NOT NULL,
    total_manipulated_volume NUMERIC(18, 4) NOT NULL,
    total_turnover_inr NUMERIC(18, 2) NOT NULL,
    participant_accounts JSONB NOT NULL,
    involved_trade_ids JSONB NOT NULL,
    subgraph_topology JSONB NOT NULL,
    window_start_utc TIMESTAMPTZ NOT NULL,
    window_end_utc TIMESTAMPTZ NOT NULL,
    detected_at_utc TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    besu_attestation_tx_hash VARCHAR(66),
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN', -- OPEN, INVESTIGATING, CONFIRMED_FRAUD, FALSE_POSITIVE, REFERRED_TO_SEBI
    assigned_officer_id VARCHAR(64),
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collusion_alerts_isin ON collusion_alerts (symbol_isin, detected_at_utc DESC);
CREATE INDEX idx_collusion_alerts_score ON collusion_alerts (composite_collusion_score DESC);
CREATE INDEX idx_collusion_alerts_status ON collusion_alerts (status);
CREATE INDEX idx_collusion_alerts_accounts ON collusion_alerts USING GIN (participant_accounts);

CREATE TABLE IF NOT EXISTS sybil_cluster_registry (
    cluster_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_label VARCHAR(128) NOT NULL,
    lead_account_hash VARCHAR(64) NOT NULL,
    member_accounts JSONB NOT NULL,
    shared_devices JSONB,
    shared_subnets JSONB,
    shared_banks JSONB,
    affinity_density NUMERIC(5, 4) NOT NULL,
    first_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active_block BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_sybil_cluster_lead ON sybil_cluster_registry (lead_account_hash);
CREATE INDEX idx_sybil_cluster_members ON sybil_cluster_registry USING GIN (member_accounts);
```

### Kafka Event Schema (`surveillance.collusion.alert.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CollusionAlertEvent",
  "type": "object",
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "alertId": { "type": "string", "format": "uuid" },
    "category": { "type": "string", "enum": ["CIRCULAR_WASH_TRADING", "SYNCHRONIZED_PUMP_AND_DUMP", "MARKING_THE_CLOSE", "FRONT_RUNNING_CARTEL", "SYBIL_MULE_RING", "ONCHAIN_CROSS_LAYER_MANIPULATION"] },
    "severity": { "type": "string", "enum": ["INFO", "LOW", "MEDIUM", "HIGH", "CRITICAL"] },
    "symbolIsin": { "type": "string", "pattern": "^[A-Z]{2}[A-Z0-9]{9}[0-9]$" },
    "compositeCollusionScore": { "type": "number", "minimum": 0.0, "maximum": 100.0 },
    "washTradingRatio": { "type": "number", "minimum": 0.0, "maximum": 1.0 },
    "participantAccountHashes": {
      "type": "array",
      "items": { "type": "string", "pattern": "^[a-f0-9]{64}$" }
    },
    "involvedTradeIds": {
      "type": "array",
      "items": { "type": "string" }
    },
    "subgraphMetrics": {
      "type": "object",
      "properties": {
        "nodeCount": { "type": "integer" },
        "edgeCount": { "type": "integer" },
        "cycleLength": { "type": "integer" },
        "graphDensity": { "type": "number" }
      },
      "required": ["nodeCount", "edgeCount", "cycleLength", "graphDensity"]
    },
    "evidenceDigestSha256": { "type": "string", "pattern": "^[a-f0-9]{64}$" },
    "besuAttestationTxHash": { "type": "string", "pattern": "^0x[a-f0-9]{64}$" },
    "timestampUtcMs": { "type": "integer" }
  },
  "required": [
    "eventId",
    "alertId",
    "category",
    "severity",
    "symbolIsin",
    "compositeCollusionScore",
    "participantAccountHashes",
    "involvedTradeIds",
    "evidenceDigestSha256",
    "timestampUtcMs"
  ]
}
```

## Security & Compliance Notes
- **SEBI PFUTP Regulations Compliance:**
  - Strict adherence to Regulations 3, 4, and 5 of the SEBI (Prohibition of Fraudulent and Unfair Trade Practices relating to Securities Market) Regulations, 2003. Specifically targets acts deemed fraudulent under Regulation 4(2)(a) (creating a false or misleading appearance of active trading), Regulation 4(2)(b) (dealing in securities not intended to effect transfer of beneficial ownership), and Regulation 4(2)(g) (entering into transactions with a view to artificially depressing or raising prices).
- **PMLA 2002 & Suspicious Transaction Reporting (STR):**
  - Graph surveillance alerts feed directly into automated Suspicious Transaction Report (STR) queues. In compliance with the Financial Intelligence Unit - India (FIU-IND) Red Flag Indicators (RFIs), rings exhibiting circular fund routing or mule account clustering trigger automatic STR drafting within 24 hours of confirmation.
- **DPDP Act 2023 & Pseudonymization Framework:**
  - In strict compliance with the **Digital Personal Data Protection Act, 2023**, the graph surveillance engine operates solely on cryptographically salted hashes (`account_hash`, `pan_hash`, `device_fingerprint`). Raw PII (names, phone numbers, Aadhaar numbers, unmasked PANs) is never ingested, stored, or processed within the graph database or GNN embedding layers.
  - De-anonymization is restricted to authorized Compliance Officers and operates under strict dual-control ("four-eyes" principle) RBAC with mandatory audit log recording.
- **Indian Evidence Act Section 65B & Chain of Custody:**
  - Graph topologies, transaction logs, and analytical scores used for regulatory prosecution must maintain an unbroken chain of custody. Every alert payload is digitally signed with an HSM-backed private key and anchored to the Hyperledger Besu consortium ledger. This satisfies admissibility requirements under Section 65B of the Indian Evidence Act, 1872 (and corresponding provisions of the Bharatiya Sakshya Adhiniyam, 2023).
- **Anti-Adversarial Model Protections:**
  - Collusion rings may attempt to poison or evade GNN detection by introducing synthetic noise trades with uninvolved retail accounts. The engine mitigates adversarial graph attacks through volume-weighted message passing, edge-dropout regularization, and temporal decay functions that discount low-volume non-recurrent edges.

## Acceptance Criteria
- [ ] **Real-Time Cycle Traversal Latency:** Directed cycle detection algorithm identifies $100\%$ of circular wash trading loops of length $k \in [2, 6]$ in an active $10,000$-edge sliding trade window in $< 35\text{ms}$.
- [ ] **Streaming Throughput & Backpressure:** Sustains streaming consumption and dynamic graph indexing at $\ge 50,000$ trades/second with zero Kafka message loss and p99 pipeline latency $< 100\text{ms}$.
- [ ] **Sybil Community Detection Accuracy:** Correctly partitions synthetic multi-account Sybil test rings (incorporating shared device, IP, and banking links) with a clustering purity $\ge 98.5\%$ and modularity $Q \ge 0.70$.
- [ ] **Wash Trading Volume Detection:** Successfully flags $100\%$ of simulated zero-net-delta wash trades where $\text{Wash Ratio} \ge 0.90$ across synthetic exchange backtest datasets.
- [ ] **GNN Inference Latency & Accuracy:** PyTorch Geometric HGNN inference model achieves an AUC-ROC $\ge 0.94$ on historical collusion test suites with per-batch inference latency $< 25\text{ms}$.
- [ ] **Hyperledger Besu On-Chain Correlation:** Accurately correlates on-chain token transfer cycles with off-chain trading accounts, registering attestation Merkle roots to `CollusionSurveillanceRegistry.sol` within 2 QBFT block cycles ($\le 4\text{s}$).
- [ ] **Pre-Trade Mitigation Trigger:** End-to-end latency from the final closing trade execution of a high-confidence collusion ring ($\text{CCI} \ge 85.0$) to the publication of the `risk.account.restrict_trading.v1` command is $< 100\text{ms}$.
- [ ] **Tamper-Evident Audit & WORM Logging:** Every emitted alert is immutably logged to PostgreSQL with an unbroken SHA-256 hash chain and dispatched to Prompt 218 (Audit Service) with zero failure rate.
- [ ] **SEBI PFUTP Export Generation:** Automatically generates fully rendered, cryptographically attested PDF/JSON investigation dossiers with embedded visual network graphs within $< 2.5\text{s}$ of compliance request.
- [ ] **Test Coverage:** Unit, integration, and property-based graph permutation tests maintain strict code coverage of $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites (Upstream Dependencies):**
  - Prompt `004` (Domestic KYC/AML Policy) & Prompt `202` (KYC/AML Service): Identity models and PAN/CKYC hashing schemes.
  - Prompt `204` (Order Lifecycle Service) & Prompt `205` (Order Matching Engine): Stream of order events and executed trade matches.
  - Prompt `208` (Trade Settlement Service): Settlement records and gross-to-net clearing outputs.
  - Prompt `218` (Audit Log Service): WORM audit trail persistence.
  - Prompt `228` (Real-Time Market Surveillance & Dynamic Volatility Engine): Surveillance baseline and real-time alert coordination.
  - Prompt `329` (NBSE Settlement DvP Smart Contract): Consortium ledger trade settlement events on Hyperledger Besu.
- **Downstream & Parallel Work:**
  - Prompt `711` (Market Surveillance & Anti-Manipulation Engine): Complementary micro-structure surveillance (spoofing, layering).
  - Prompt `712` (Insider Trading & UPSI Graph Analytics Engine): Shared Neo4j graph infrastructure and kinship relationship resolution.
  - Prompt `714` (Cross-Chain AML & Blockchain Forensics Screener): External public blockchain transaction screener.
  - Prompt `715` (NBSE Market Surveillance & SEBI Reporting Engine): Exchange-level regulatory reporting and statutory STR dispatch.
  - Prompt `206` (Risk & Margin Checks Service): Ingests automated trading block containment signals.
