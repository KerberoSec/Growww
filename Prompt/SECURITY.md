# Security Architecture, Cryptographic Standards & Statutory Compliance Framework
## Sovereign Financial Infrastructure Specification for the Growww / National Blockchain Stock Exchange (NBSE)

---

### Executive Summary & Institutional Mandate

This document establishes the definitive, binding security architecture, cryptographic invariant framework, defense-in-depth posture, and statutory compliance controls for the Growww / National Blockchain Stock Exchange (NBSE) platform. As an institutional-grade, sovereign financial infrastructure operating on permissioned Hyperledger Besu rails with Istanbul / QBFT consensus and 1:1 real-asset custodial backing (NSDL, CDSL, and RBI depositories), the NBSE platform is designed to withstand nation-state cyber threats, zero-day distributed vectors, and insider collusion while guaranteeing absolute data privacy and deterministic financial settlement.

The core tenets mandated across every prompt, microservice, smart contract, and operational procedure within this platform comprise:
- Absolute Zero Personally Identifiable Information (Zero-PII) on the blockchain ledger.
- Hardware-enforced cryptographic security utilizing FIPS 140-2 Level 3 Hardware Security Modules (HSM) and Multi-Party Computation Threshold Signature Schemes (MPC-TSS).
- Multi-layered defense-in-depth spanning Anycast edge filters down to immutable smart contract invariant guards.
- Exhaustive statutory alignment with the Securities and Exchange Board of India (SEBI) Cybersecurity and Cyber Resilience Framework (CSCRF), the Reserve Bank of India (RBI) CBDC and Digital Lending Directives, the International Financial Services Centres Authority (IFSCA) AML/CFT rules, and the Digital Personal Data Protection Act, 2023 (DPDP Act 2023).

---

## 1. Security Architecture & Threat Modeling Framework

### 1.1 Architectural Trust Boundaries & Isolation Zones

The NBSE infrastructure is segregated into seven distinct, non-overlapping security zones. Traversal across boundaries is strictly mediated through authenticated, cryptographically signed, rate-limited, and audited gateways.

| Zone Identifier | Zone Name | Description & Trust Level | Primary Workloads & Assets | Inbound Ingress Controls | Outbound Egress Controls |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Zone 0 | Untrusted Edge | Zero Trust; untrusted public internet networks. | Flutter multi-platform clients, web trading terminals, external institutional Direct Market Access (DMA) client systems. | TLS 1.3 termination, Anycast DDoS scrubbers, Web Application Firewall (WAF) filtering. | N/A (Client Initiated). |
| Zone 1 | Perimeter DMZ | Untrusted Ingress; heavily fortified DMZ. | Anycast BGP ingress points, Cloudflare Magic Transit, AWS Shield Advanced, Kong Enterprise API Gateway, Envoy proxy routers. | Strict Layer 7 inspection, OWASP Top 10 filter sets, geo-fencing, rate limiting per IP / API key. | mTLS restricted exclusively to Zone 2 service mesh ingress points. |
| Zone 2 | Application Mesh | High Trust; internal containerized microservice mesh. | Rust order matching engines, Go transactional relayers, Gemini AI Advisor inference engine, Apache Kafka event streaming clusters. | Bidirectional SPIFFE/SPIRE mTLS, strict JWT / OAuth2 bearer token cryptographic validation. | Restricted to explicit Zone 3 database sockets and Zone 4 JSON-RPC relays. Egress to public internet blocked. |
| Zone 3 | Financial State Tier | Restricted Trust; core transactional state and data persistence. | PostgreSQL 16 primary-replica clusters, Redis Sentinel in-memory state caches, Kafka transactional commit logs. | mTLS authenticated database connections, strict role-based PostgreSQL grants, Redis AUTH over TLS 1.3. | Default deny-all. No external network routes. Isolated private database subnets. |
| Zone 4 | Ledger Core | Sovereign High Trust; permissioned blockchain node network. | Hyperledger Besu validator nodes running QBFT consensus, non-validating RPC sync nodes, bootnodes. | JSON-RPC filtered over internal mTLS gateway; p2p discovery restricted via static enode allowlists. | P2P traffic strictly restricted to authorized consortium validator IP addresses. Zero internet egress. |
| Zone 5 | Cryptographic Enclave | Ultra-High Trust; isolated hardware root of trust. | AWS CloudHSM clusters, Thales Luna 7 Network HSMs, MPC-TSS distributed signing engine enclaves. | Dedicated point-to-point VPC peering via AWS PrivateLink, PKCS#11 API over mutual TLS, multi-factor operator authentication. | Absolute zero egress; no internet connectivity; hardware-enforced air gap. |
| Zone 6 | Regulated Clearing Rails | Sovereign External Trust; dedicated financial leased lines. | NSDL / CDSL depository clearing bridges, RBI Central Bank Digital Currency (CBDC) settlement gateways, Clearing Corporation of India connections. | Dedicated point-to-point leased lines, hardware IPSec VPN tunnels with BGP peering, mutual client certificates. | Restricted strictly to designated clearing corporation IPs and ports. |

### 1.2 STRIDE Threat Modeling Applied to Subsystems

Every subsystem specified across the NBSE platform is modeled against the STRIDE threat categorization:

#### 1.2.1 Order Ingestion & API Gateway Tier
- Spoofing: An adversary attempts to inject unauthorized trade orders by forging client identity tokens or replaying captured requests.
  - Architectural Mitigations: Ephemeral ECDSA client signatures over full order payloads, Ed25519-signed JSON Web Tokens (JWT) with 15-minute expiration, and strict nonce tracking in Redis to invalidate replays.
- Tampering: Modification of order limit prices, volumes, or order types in transit between the client and gateway.
  - Architectural Mitigations: Mandatory TLS 1.3 with forward-secret cipher suites, end-to-end cryptographic checksum verification, and Protobuf schema validation rejecting malformed or unexpected fields.
- Repudiation: A trader claims they did not place a loss-making market or limit order.
  - Architectural Mitigations: Immutable, non-repudiable audit logs containing client hardware signatures, IP attestation, client certificate thumbprints, and sequential order sequence numbers written to an append-only Kafka audit topic.
- Information Disclosure: Eavesdropping on resting orders to front-run institutional block orders.
  - Architectural Mitigations: In-transit TLS 1.3 encryption across all internal transport layers, zero logging of order book depths to public endpoints, and internal payload masking.
- Denial of Service: Flood of micro-orders designed to exhaust API gateway connection pools or buffer queues.
  - Architectural Mitigations: Token bucket rate limiting per client UID, global IP connection throttling, Layer 7 WAF inspection, and immediate blacklisting of anomalous socket spam.
- Elevation of Privilege: Exploitation of gateway routing bugs to invoke administrative settlement endpoints.
  - Architectural Mitigations: Strict separation of client trading gateways from administrative portals, zero administrative route registration on public endpoints, and mandatory mutual TLS with physical hardware token validation for admin paths.

#### 1.2.2 In-Memory Matching Engine & Market Data Distribution
- Spoofing: Injection of fake execution matches or synthetic order book updates.
  - Architectural Mitigations: The matching engine accepts inputs strictly from verified internal Go relayers over authenticated IPC / gRPC channels; all incoming messages require internal cryptographic sequence stamps.
- Tampering: State corruption within the order book tree data structure leading to mismatched trade fills.
  - Architectural Mitigations: Deterministic state-machine replication, memory isolation using non-allocating high-speed ring buffers, and periodic cryptographic snapshot hashing against secondary shadow engines.
- Repudiation: Trading counterparty disputing an in-memory fill confirmation.
  - Architectural Mitigations: Every matched trade produces an immutable deterministic trade event published directly to partitioned Kafka commit logs with monotonic sequence numbering before downstream distribution.
- Information Disclosure: Cache timing attacks or memory inspection to gain visibility into resting institutional orders.
  - Architectural Mitigations: Execution inside dedicated bare-metal or hardened Linux instances with CPU core pinning, disabled memory swapping (swapoff), and strict kernel memory protection (kernel.kptr_restrict = 2).
- Denial of Service: Flooding order cancellation messages to induce processing latency spikes.
  - Architectural Mitigations: Fixed-size circular memory buffers, O(1) order cancellation lookup maps, and mandatory asymmetric speed bumps (500us) applied equally across all participant order flows.
- Elevation of Privilege: Memory overflow exploitation to gain host shell access.
  - Architectural Mitigations: Core matching engine implemented exclusively in memory-safe systems languages (Rust), compiled with full stack protection (stack-protector-strong), Address Space Layout Randomization (ASLR), and Position Independent Executables (PIE).

#### 1.2.3 Blockchain Relayer & Smart Contract Execution (Hyperledger Besu)
- Spoofing: Submission of transactions to the Besu ledger purporting to originate from authorized clearing relayers.
  - Architectural Mitigations: Node-level authentication via static enode peering, JSON-RPC endpoint access restricted to mTLS-authenticated relayers, and transaction signing performed exclusively within FIPS 140-2 Level 3 HSM enclaves.
- Tampering: Modification of transaction calldata or execution parameters during relay.
  - Architectural Mitigations: Relayer transactions are constructed and signed inside the HSM or MPC enclave; any byte-level modification invalidates the cryptographic signature, causing immediate rejection by Besu QBFT validators.
- Repudiation: Relayer operator claims a batch settlement transaction was unauthorized.
  - Architectural Mitigations: Multi-party threshold signatures (MPC-TSS) requiring $t$-of-$n$ approvals across distinct operational units, producing an auditable threshold signature trace.
- Information Disclosure: Extraction of confidential trader identity or trade size from ledger state or block events.
  - Architectural Mitigations: Zero-PII ledger design; trade amounts protected via Zero-Knowledge Pedersen Commitments, and trader addresses abstracted through ERC-3643 ONCHAINID identity contracts.
- Denial of Service: Spamming the ledger with zero-value transactions to stall block production.
  - Architectural Mitigations: Zero public RPC exposure; transaction admission regulated by private mempool filters that reject non-whitelisted contract invocations and throttle transaction submission rates per relayer account.
- Elevation of Privilege: Reentrancy or access-control bypass in smart contracts to mint unbacked token shares.
  - Architectural Mitigations: Formally verified smart contracts, open access controls replaced with OpenZeppelin AccessControlDefaultAdminRules, reentrancy guards on all external state-modifying functions, and hard checks enforcing 1:1 real-asset custodial backing reserves.

#### 1.2.4 Google Gemini Financial Advisor Service
- Spoofing: Adversary impersonating the Gemini Advisor service to emit fraudulent investment recommendations.
  - Architectural Mitigations: Service-to-service mTLS using SPIFFE/SPIRE x509 workload identities, client-side signature verification of advisor output payloads.
- Tampering: Prompt injection attacks manipulating LLM context to bypass SEBI non-advisory guardrails.
  - Architectural Mitigations: Strict multi-stage input sanitization pipelines, rigid system prompt isolation, secondary classifier verification passes, and deterministic AST verification of generated outputs.
- Repudiation: User disputing investment decisions influenced by AI advisory interactions.
  - Architectural Mitigations: Cryptographically signed, immutable interaction audit logs capturing exact prompt context, model temperature, output tokens, and timestamped user acknowledgments stored in compliance data lakes for 8 years.
- Information Disclosure: Extraction of proprietary trading data or cross-user portfolio context through LLM hallucination or prompt leakage.
  - Architectural Mitigations: Complete multi-tenant isolation, stateless API interactions with the model, zero user data retention on foundation model provider endpoints, and strict Retrieval-Augmented Generation (RAG) tenancy scoping.
- Denial of Service: Complex algorithmic prompt spam designed to exhaust GPU compute clusters or API token quotas.
  - Architectural Mitigations: Adaptive token throttling per user account, strict request timeout limits (5.0s), and semantic caching for common financial queries using vector embeddings.
- Elevation of Privilege: Exploitation of LLM function calling / tools to trigger unauthorized asset transfers.
  - Architectural Mitigations: The Gemini Advisor service is architecturally read-only; it possesses zero write credentials, zero transaction signing capabilities, and zero direct connectivity to transaction relayers or settlement contracts.

### 1.3 Attack Vectors Unique to Sovereign Real-World Asset (RWA) Blockchains

Operating a permissioned RWA exchange introduces attack classes distinct from both public decentralized finance (DeFi) and traditional centralized stock exchanges:

```
+----------------------------------------------------------------------------------------------------+
|                         Sovereign RWA Blockchain Attack Vector Landscape                           |
+----------------------------------------------------------------------------------------------------+
| 1. Front-Running & MEV         | 2. Toxic HFT Arbitrage         | 3. Oracle Manipulation           |
| - Private validator mempool    | - 500us asymmetric speed bumps | - Multi-source medianized feeds  |
| - Deterministic FIFO ordering  | - Minimum tick & lot sizes     | - Dynamic TWAP/VWAP collars      |
| - Zero public transaction pool | - Anti-sniping logic gates     | - Outlier rejection algorithms   |
+--------------------------------+--------------------------------+----------------------------------+
| 4. Custodial De-synchronization| 5. vAMM Invariant Drain        | 6. Validator Cartelization       |
| - Dual-ledger atomic locks     | - Constant-product bounds      | - QBFT Byzantine quorum (3f + 1) |
| - Automated 60s PoR sweeps     | - Flash-loan mitigation locks  | - Multi-institution governance   |
| - Hard-stop circuit breakers   | - Maximum slippage caps        | - Node geographic distribution   |
+--------------------------------+--------------------------------+----------------------------------+
| 7. Forensic De-Anonymization   | 8. Leased-Line Wiretap         | 9. Depository Partitioning       |
| - ERC-3643 claim hash masking  | - Hardware MACsec encryption   | - Dual-depository failover       |
| - ZK Pedersen balance blinding | - Redundant optical paths      | - Offline reconciliation queue   |
| - Ephemeral address mixing     | - BGP route attestation        | - Deterministic state recovery   |
+----------------------------------------------------------------------------------------------------+
```

1. Maximal Extractable Value (MEV) & Transaction Reordering:
   - In a permissioned ledger, validator collusion can reorder settlement batches to front-run off-market trades.
   - Mitigation: Hyperledger Besu mempools are configured with zero transaction prioritization fees (zero gas price auctions); transaction sequencing is enforced by deterministic FIFO timestamping recorded at the Go relayer ingress point, and blocks are finalized with single-block QBFT deterministic finality.

2. Toxic High-Frequency Trading (HFT) Arbitrage & Latency Sniping:
   - Institutional participants with microsecond network advantages attempting to pick off off-market vAMM quotes before depository updates propagate.
   - Mitigation: Hardware-level asymmetric speed bumps (500us randomized latency jitter applied to cancellation requests and aggressive market orders), coupled with minimum quote life rules.

3. Cross-Market Oracle Manipulation:
   - Artificial distortion of underlying equity spot prices on external secondary venues to trigger automated liquidations or unfavorable fractional redemptions on the NBSE ledger.
   - Mitigation: Multi-source medianized oracles ingesting tick data from NSE, BSE, and global depositary receipt (GDR) venues simultaneously; dynamic price bands (5% circuit collars per rolling 60-second window); rejection of any oracle tick that deviates by more than 3 standard deviations from the 5-minute Time-Weighted Average Price (TWAP).

4. Custodial Reconciliation De-synchronization (Split-Brain Ledger):
   - A condition where the on-chain digital token balance diverges from physical depository share holdings in NSDL / CDSL due to a communication partition or unhandled corporate action (stock split, bonus issue, demerger).
   - Mitigation: Real-time automated reconciliation daemons running on 60-second cycles; cryptographic assertion of Proof-of-Reserve (PoR) prior to opening every continuous trading window; instant platform-wide automated circuit breaker engagement if divergence exceeds 0.00000001 shares.

5. vAMM Liquidity Invariant Drain:
   - Algorithmic attempts to drain off-market hybrid vAMM liquidity pools using non-atomic cross-pool arbitrage loops.
   - Mitigation: Strict bounding of maximum trade size to 2% of virtual pool depth per block; automated volatility fees that dynamically scale transaction costs in proportion to pool imbalance; absolute prohibition of uncollateralized flash loans within the token smart contracts.

6. Validator Collusion & Consensus Liveness Stalls:
   - A minority validator cartel attempting to halt consensus to prevent liquidation executions or regulatory freeze orders.
   - Mitigation: QBFT consensus tolerance guarantees safety and liveness provided less than one-third of validators are Byzantine ($f < (n - 1) / 3$). Validators are distributed across legally independent institutional entities: National Depository, Clearing Corporation, Regulated Broker Consortium, and State Sovereign Entity.

### 1.4 Comprehensive Threat Modeling & DREAD Risk Matrix

The DREAD framework evaluates each vector across five parameters scored from 1 (lowest) to 10 (highest): Damage potential, Reproducibility, Exploitability, Affected users, and Discoverability. The cumulative score determines architectural priority.

| Threat ID | Threat Description | Architectural Component | STRIDE Class | D | R | E | A | D | Total Score | Risk Level | Primary Mitigating Architectural Control |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| THR-01 | Relayer Hot Wallet Key Exfiltration | Go Transaction Relayer | Spoofing / Elevation | 10 | 3 | 3 | 10 | 3 | 29 | Critical | FIPS 140-2 Level 3 HSM delegation; MPC-TSS 3-of-5 signing; transaction whitelist constraints. |
| THR-02 | Smart Contract Reentrancy Minting | ERC-3643 Settlement Contract | Tampering / Elevation | 10 | 8 | 4 | 10 | 5 | 37 | Critical | Formal verification via Certora; OpenZeppelin non-reentrant mutex locks; single-state invariant assertions. |
| THR-03 | In-Memory Engine State Corruption | Rust Matching Engine | Tampering | 9 | 4 | 3 | 9 | 3 | 28 | Critical | Memory safety via Rust; deterministic state-machine replication; ring-buffer isolation. |
| THR-04 | In-Transit Order Manipulation | API Gateway Ingress | Tampering | 8 | 3 | 4 | 8 | 4 | 27 | High | Mandatory TLS 1.3; client ECDSA signature verification on order payloads; Protobuf typing. |
| THR-05 | Depository Feed Replay / Desync | NSDL/CDSL Leased-Line Bridge | Tampering / Repudiation | 9 | 4 | 3 | 9 | 4 | 29 | Critical | Monotonic sequence counters; cryptographic transmission tokens; 60s automated PoR reconciliation. |
| THR-06 | PII Leakage on Ledger Layer | Hyperledger Besu Nodes | Information Disclosure | 10 | 8 | 6 | 10 | 7 | 41 | Critical | Zero-PII ledger design; ERC-3643 claim hashes; ZK Pedersen commitments; off-chain PII vaulting. |
| THR-07 | Gemini AI Prompt Injection Exploitation | Gemini Financial Advisor | Tampering / Elevation | 6 | 7 | 6 | 6 | 8 | 33 | High | Strict input token sanitization; read-only service role; architectural isolation from trade execution rails. |
| THR-08 | Byzantine Validator Collusion | Besu QBFT Consensus | Tampering / DoS | 10 | 2 | 2 | 10 | 2 | 26 | High | Multi-institutional validator governance; statutory legal oversight; 3f+1 consensus quorum. |
| THR-09 | Distributed Denial of Service (DDoS) | Public Edge Gateways | Denial of Service | 7 | 8 | 8 | 8 | 9 | 40 | High | Cloudflare Magic Transit Anycast scrubbers; AWS Shield Advanced; token bucket rate limiters. |
| THR-10 | Insider Data Exfiltration from Database | Core PostgreSQL Tier | Information Disclosure | 9 | 4 | 4 | 9 | 4 | 30 | Critical | Transparent Database Encryption (TDE); column-level AES-256-GCM; dual-custody access controls. |
| THR-11 | Unauthorized Administrative State Change | Master Configuration Contracts | Elevation of Privilege | 10 | 3 | 3 | 10 | 3 | 29 | Critical | OpenZeppelin AccessControlDefaultAdminRules; 48-hour Timelock; multi-sig governance execution. |
| THR-12 | Off-Market Price Oracle Poisoning | Price Oracle Aggregate Relay | Tampering | 9 | 4 | 4 | 9 | 5 | 31 | Critical | Multi-source medianized feeds; dynamic 5% rolling volatility bands; TWAP sanity checks. |

---

## 2. The Zero-PII Ledger Invariant

### 2.1 The Fundamental Design Axiom

Under no circumstances shall Personally Identifiable Information (PII), Sensitive Personal Data or Information (SPDI), or direct financial identifiers ever be written to the Hyperledger Besu blockchain state, event logs, transaction calldata, block headers, or node commit histories.

Prohibited data elements include, but are not limited to:
- Real names, aliases, patronymics, or signatures.
- Permanent Account Number (PAN), Aadhaar Number, Passport Number, Voter ID, or Driving License details.
- Bank Account Numbers, International Bank Account Numbers (IBAN), Indian Financial System Codes (IFSC), or UPI Virtual Payment Addresses (VPA).
- Physical residential or commercial addresses, postal codes, and GPS location coordinates.
- Mobile telephone numbers, email addresses, and instant messaging handles.
- IP addresses, MAC addresses, device fingerprints, and browser user-agent strings.
- Biometric templates, facial recognition vectors, and photographic imagery.

This invariant is immutable and enforced via automated static analysis in the CI/CD pipeline, pre-commit ledger hooks, and deterministic schema sanitizers.

### 2.2 ERC-3643 (T-REX Compliant) Identity Registry & Cryptographic Claim Hashes

To achieve full regulatory compliance with SEBI and IFSCA KYC/AML mandates without leaking PII on-chain, the platform deploys the ERC-3643 (Token for Regulated EXchanges) standard, utilizing ONCHAINID contract abstractions.

```
+----------------------------------------------------------------------------------------------------+
|                             Zero-PII ERC-3643 Identity Architecture                                |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|  [ Physical Real Person ] ---> [ Off-Chain PII Vault ] ---> [ Trusted Claim Issuer (KYC Provider) ]  |
|                                (AES-256-GCM / HSM)                   |                             |
|                                                                      v                             |
|                                                         Generates Cryptographic Claim              |
|                                                         Topic: 1 (KYC Approved)                    |
|                                                         Data: keccak256(RawClaim + Salt)           |
|                                                                      |                             |
|                                                                      v                             |
|  [ Investor Wallet Address ] <---> [ ONCHAINID Identity Contract ] <---+                           |
|              |                                    |                                                |
|              v                                    v                                                |
|  [ Settlement Transaction ] -------> [ ERC-3643 Compliance Module ]                                |
|                                       - Verifies Claim Signature from Trusted Issuer               |
|                                       - Verifies Claim Expiration Date                             |
|                                       - Zero PII Inspected or Disclosed                            |
+----------------------------------------------------------------------------------------------------+
```

1. Architectural Separation:
   - A trader's on-chain presence is represented exclusively by an Ethereum address mapped 1:1 to an ONCHAINID identity smart contract.
   - The identity contract contains zero user identity data; it contains only an array of cryptographic claims signed by authorized Claim Issuers (e.g., licensed KYC registration agencies, custodians, and compliance verification nodes).

2. Cryptographic Claim Construction:
   - Claim Topic: A 32-byte standardized numerical category (e.g., Topic 1: SEBI KYC Verified, Topic 2: FPI Registration Valid, Topic 3: Accredited Investor, Topic 4: FATF Non-Sanctioned Jurisdiction).
   - Claim Data Hash: A cryptographic digest computed off-chain by the trusted issuer:
     $$\text{ClaimHash} = \text{keccak256}(\text{IdentityContractAddress} \parallel \text{ClaimTopic} \parallel \text{ClaimPayload} \parallel \text{SecretSalt})$$
   - Issuer Signature: The Claim Issuer signs the computed ClaimHash using their HSM-backed private key:
     $$\text{Signature} = \text{ECDSA\_Sign}(\text{IssuerPrivateKey}, \text{keccak256}(\text{"\x19Ethereum Signed Message:\n32"} \parallel \text{ClaimHash}))$$
   - The resulting signature, topic, and issuer address are recorded on the identity contract. The underlying personal data ($\text{ClaimPayload}$) and the $\text{SecretSalt}$ remain strictly within the isolated off-chain PII vault.

3. Dynamic Compliance Verification:
   - When a token transfer or order settlement is executed, the ERC-3643 compliance module inspects the identity registry of both sender and receiver.
   - It mathematically verifies that both identity contracts hold valid, non-expired signatures from approved Claim Issuers for all mandatory claim topics.
   - If any claim is missing, revoked, or expired, the smart contract transaction reverts with a deterministic error code, preventing settlement without ever reading an identity attribute.

### 2.3 Zero-Knowledge Pedersen Blinding & Commitments

To preserve confidentiality of trade amounts, balance states, and institutional positions on the permissioned ledger while maintaining mathematical certainty of solvency and balance conservation, the platform uses Zero-Knowledge Pedersen Commitments over the alt_bn128 (BN254) elliptic curve.

1. Mathematical Formulation:
   - For an asset balance or trade value $v \in \mathbb{Z}_q$, where $q$ is the prime group order, the system selects two elliptic curve generator points $G$ and $H$ such that the discrete logarithm $\log_G(H)$ is verifiably unknown (computed via standard Nothing-Up-My-Sleeve point derivation).
   - A cryptographically secure random blinding factor $r \in \mathbb{Z}_q$ is generated inside the client or relayer security enclave.
   - The Pedersen Commitment $C$ is computed as:
     $$C = r \cdot G + v \cdot H$$

2. Homomorphic Balance Invariance:
   - Due to the additively homomorphic property of Pedersen commitments, the settlement contract verifies that the sum of output commitments equals the sum of input commitments plus transaction fees without decrypting the underlying values:
     $$\sum C_{\text{in}} - \sum C_{\text{out}} - C_{\text{fee}} = \left(\sum r_{\text{in}} - \sum r_{\text{out}} - r_{\text{fee}}\right) \cdot G + \left(\sum v_{\text{in}} - \sum v_{\text{out}} - v_{\text{fee}}\right) \cdot H = 0 \cdot H = \mathcal{O}$$
   - When balances are balanced ($\sum v_{\text{in}} = \sum v_{\text{out}} + v_{\text{fee}}$), the equation reduces strictly to a commitment to zero with respect to generator $H$:
     $$\sum C_{\text{in}} - \sum C_{\text{out}} - C_{\text{fee}} = \Delta r \cdot G$$
   - The transacting party submits a zero-knowledge Schnorr proof demonstrating knowledge of $\Delta r$, proving balance conservation without revealing trade sizes.

3. Zero-Knowledge Range Proofs (Bulletproofs):
   - To prevent negative value exploits (integer underflow attacks where $v < 0$), every state update transaction must accompany a zero-knowledge range proof demonstrating that:
     $$v \in [0, 2^{64} - 1]$$
   - The platform utilizes non-interactive Bulletproofs verified via precompiled contracts on the Besu virtual machine, incurring zero trusted setup requirements and logarithmic proof size.

4. Authorized Regulatory View Keys:
   - For statutory auditing under SEBI and FIU-IND inspection mandates, transacting parties encrypt the blinding factor $r$ and balance $v$ using an asymmetric ElGamal encryption scheme under the sovereign regulatory public key:
     $$E_{\text{reg}} = (k \cdot G, \; M + k \cdot Y_{\text{reg}})$$
   - Where $Y_{\text{reg}}$ is the public key of the regulatory authority, and $k$ is an ephemeral session scalar. Authorized statutory inspectors equipped with the corresponding private view key can deterministically decrypt on-chain trade values without disrupting platform confidentiality.

### 2.4 Off-Chain PII Vaulting Architecture

All customer personal identity data, banking credentials, tax documentation, and regulatory verification artifacts are quarantined within a physically and logically isolated Off-Chain PII Vault.

```
+----------------------------------------------------------------------------------------------------+
|                                 Off-Chain PII Vault Architecture                                   |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|   +--------------------------------------------------------------------------------------------+   |
|   | AWS CloudHSM / Thales Luna 7 HSM (FIPS 140-2 Level 3)                                      |   |
|   | - Holds Master Key Encryption Key (KEK)                                                    |   |
|   +--------------------------------------------------------------------------------------------+   |
|                                                |                                                   |
|                         Unwraps DEK via PKCS#11 | Generates Ephemeral DEKs                          |
|                                                v                                                   |
|   +--------------------------------------------------------------------------------------------+   |
|   | Secure PII Vault Enclave (Isolated Private VPC / Zero Internet Route)                      |   |
|   |                                                                                            |   |
|   |   Raw PII Payload ---> Encrypted with User-Unique DEK (AES-256-GCM)                        |   |
|   |                        - 96-bit Unique Initialization Vector (IV)                          |   |
|   |                        - 128-bit Authentication Tag (MAC)                                  |   |
|   |                        - Aadhaar / PAN / Bank Details Encrypted at Rest                    |   |
|   |                                                                                            |   |
|   |   Synthetic Identifier Generator:                                                          |   |
|   |   - Outputs: Pseudonymous Tenant UUIDv4                                                    |   |
|   |   - Links to: ERC-3643 ONCHAINID Address                                                   |   |
|   |   - Zero PII leaves the enclave perimeter                                                  |   |
|   +--------------------------------------------------------------------------------------------+   |
|                                                |                                                   |
|                                                v Encrypted Ciphertext + Wrapped DEK                |
|   +--------------------------------------------------------------------------------------------+   |
|   | PostgreSQL 16 Isolated Storage Tier (Encrypted Tablespaces + Transparent Data Encryption)  |   |
|   +--------------------------------------------------------------------------------------------+   |
+----------------------------------------------------------------------------------------------------+
```

1. Cryptographic Envelope Encryption:
   - Raw personal data fields are encrypted using authenticated Advanced Encryption Standard in Galois/Counter Mode (AES-256-GCM).
   - Every individual user record is encrypted under a unique, cryptographically random Data Encryption Key (DEK).
   - The DEK is encrypted (wrapped) using an institutional Key Encryption Key (KEK) stored inside a FIPS 140-2 Level 3 Hardware Security Module:
     $$\text{Ciphertext} = \text{AES-256-GCM}_{\text{DEK}}(\text{RawPII}, \text{IV}, \text{AAD})$$
     $$\text{WrappedDEK} = \text{AES-KeyWrap}_{\text{KEK}}(\text{DEK})$$
   - The Additional Authenticated Data (AAD) binds the ciphertext to the user's synthetic UUIDv4, preventing ciphertext transplant attacks across database records.

2. In-Memory Lifecycle & Volatile Scrubbing:
   - Decrypted PII attributes exist exclusively within volatile, protected memory (RAM) allocated via operating system primitives that prevent paging to disk (`mlock`).
   - Core dumps are disabled globally (`prctl(PR_SET_DUMPABLE, 0)`).
   - Upon completion of a verification operation, memory buffers containing plaintext identity attributes are explicitly overwritten with cryptographically random noise before release (`explicit_bzero` / volatile zeroization).

3. Synthetic Pseudonymization:
   - Downstream trading microservices, matching engines, and relayer queues reference users exclusively through a 128-bit cryptographically random synthetic UUIDv4.
   - The mapping between the synthetic UUIDv4 and the encrypted PII record is confined to the isolated PII Vault database; no other service possesses database read grants to this lookup table.

### 2.5 Data Sovereignty & DPDP Act 2023 Enforcement

1. Sovereign Indian Data Residency:
   - In accordance with the Digital Personal Data Protection Act, 2023 (DPDP Act 2023) and Reserve Bank of India data localization circulars, all databases, message queues, caches, and cryptographic vaults holding personal or financial records reside exclusively within physical data centers located within the territorial borders of the Republic of India (Primary: Mumbai Region; Disaster Recovery: Chennai / Hyderabad Region).
   - Cross-border transmission of raw personal data is prohibited. International investors participating via GIFT City IFSCA routes have their PII vaulted within sovereign IFSCA-approved Indian infrastructure, with zero replication to foreign cloud regions.

2. Technical Enforcement of the Right to Erasure (Crypto-Shredding):
   - A fundamental challenge in blockchain engineering is reconciling the immutable nature of distributed ledgers with statutory data erasure mandates (Section 12 of the DPDP Act 2023 and GDPR Article 17).
   - The NBSE platform resolves this through deterministic cryptographic shredding:
     - All on-chain records reference only claim hashes and synthetic ONCHAINID contract addresses.
     - When a customer exercises their statutory right to erasure (and subject to mandatory statutory financial record retention windows under the Prevention of Money Laundering Act), the platform permanently purges the user's unique Data Encryption Key (DEK) from the PII Vault.
     - The corresponding wrapped key is zeroized inside the HSM.
     - Without the DEK, the vaulted encrypted personal data is rendered mathematically indistinguishable from random noise, permanently and irreversibly severing all ties between the physical person and the immutable on-chain state hashes.

---

## 3. Cryptographic Key Management & HSM Architecture

### 3.1 Hardware Security Module (HSM) Deployment Architecture

All operational cryptographic keys utilized across the NBSE platform are generated, stored, and executed inside tamper-reactive Hardware Security Modules validated to Federal Information Processing Standards (FIPS) 140-2 Level 3.

```
+----------------------------------------------------------------------------------------------------+
|                                Cryptographic Key & HSM Hierarchy                                   |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|  [ Tier 0: Root Key Encryption Key (Master Seed) ]                                                 |
|  - Air-gapped, FIPS 140-2 Level 3 Physical HSM                                                     |
|  - Never leaves hardware boundary; multi-custodian M-of-N physical smart cards                     |
|                                |                                                                   |
|                                v Derives via HKDF-SHA256                                           |
|  [ Tier 1: Service Key Encryption Keys (KEKs) ]                                                    |
|  - Dedicated partition per microservice domain                                                     |
|  - PII Vault KEK, Settlement Bridge KEK, Treasury Governance KEK                                   |
|                                |                                                                   |
|                                v Wraps / Generates                                                 |
|  [ Tier 2: Operational Data Encryption Keys (DEKs) ]                                               |
|  - Per-user unique keys; rotates on data modification; envelope encryption                        |
|                                                                                                    |
|  ------------------------------------------------------------------------------------------------  |
|                                                                                                    |
|  [ Tier 3: Ledger Validator & Settlement Signing Infrastructure ]                                  |
|                                                                                                    |
|    +-----------------------------+                 +-----------------------------------------+     |
|    | QBFT Validator Private Keys |                 | Relayer MPC-TSS Threshold Signing Pool  |     |
|    | - Dedicated CloudHSM Enclave|                 | - FROST (Schnorr) / GG20 (ECDSA)         |     |
|    | - Signs 2.0s block proposals|                 | - 3-of-5 Institutional Threshold        |     |
|    | - Auto-zeroizes on physical |                 |   * Custodian Node Share                |     |
|    |   tamper detection          |                 |   * Clearing House Node Share           |     |
|    +-----------------------------+                 |   * Exchange Risk Node Share            |     |
|                                                    |   * Regulatory Compliance Node Share    |     |
|                                                    |   * Disaster Recovery Offline Share     |     |
|                                                    +-----------------------------------------+     |
+----------------------------------------------------------------------------------------------------+
```

1. Hardware Specifications:
   - Primary Cloud Infrastructure: AWS CloudHSM clusters provisioned across three distinct Availability Zones in the Mumbai region. Each HSM instance is a dedicated, single-tenant cryptographic appliance running in customer-controlled VPC subnets.
   - On-Premises Secondary Infrastructure: Thales Luna PCIe / Network HSM 7000 appliances deployed within secure cages at the primary physical data center.
   - Physical Security: Hardware zeroization triggers on physical enclosure penetration, voltage fluctuation, temperature anomalies, and unauthorized chassis access.

### 3.2 Enterprise Cryptographic Key Hierarchy

The key management lifecycle is organized into five functional tiers:

1. Tier 0: Root Key Encryption Key (Root KEK / Master Seed):
   - The root of all cryptographic trust across the exchange.
   - Generated during an audited key ceremony; never exists in unencrypted form outside the HSM hardware boundary.
   - Backed up using $M$-of-$N$ physical smart cards distributed across designated executive trust officers.

2. Tier 1: Domain Service Key Encryption Keys (Service KEKs):
   - Cryptographically derived from the Root KEK using HMAC-based Extract-and-Expand Key Derivation Function (HKDF-SHA256).
   - Dedicated service KEKs are segregated by operational domain: PII Vault KEK, Database Storage KEK, Kafka Transport KEK, and Internal Token Signing KEK.

3. Tier 2: Operational Data Encryption Keys (DEKs):
   - Ephemeral, single-purpose AES-256 keys generated to encrypt specific tables, files, or user identity records.
   - Wrapped by the corresponding Tier 1 Service KEK and stored alongside the ciphertext.

4. Tier 3: Blockchain Consensus & Transaction Signing Keys:
   - QBFT Validator Keys: secp256k1 private keys held inside CloudHSM instances utilized by Hyperledger Besu validator nodes to sign block proposals and consensus rounds.
   - Relayer Signing Keys: Hot keys utilized to inject aggregated batch settlement transactions into the Besu ledger.

5. Tier 4: Cold Governance & Emergency Treasury Keys:
   - Highly restricted keys utilized to execute smart contract upgrades, parameter modifications, and emergency contract pausing.
   - Stored in deep cold storage requiring physical multi-party presence to activate.

### 3.3 Multi-Party Computation Threshold Signature Scheme (MPC-TSS)

To eliminate single points of compromise across high-value operations (minting fractional RWA tokens, updating system contracts, releasing custodial reserves), the platform mandates Multi-Party Computation Threshold Signature Schemes.

1. Algorithmic Foundation:
   - Elliptic Curve Digital Signature Algorithm (ECDSA) for secp256k1 rails uses the Gennaro-Goldfeder 2020 protocol (GG20) with proactive security.
   - Schnorr / Ed25519 rails utilize the Flexible Round-Optimized Schnorr Threshold (FROST) protocol.

2. Threshold Topologies:
   - Normal Batch Relayer Settlement: 3-of-5 threshold.
     - Share 1: Primary Exchange Matching Engine Relayer (Automated Enclave).
     - Share 2: Independent Custodian Bridge Node (Automated Enclave).
     - Share 3: Clearing Corporation Verification Node (Automated Enclave).
     - Share 4: Sovereign Compliance Oversight Monitor (Automated Enclave).
     - Share 5: Cold Disaster Recovery Key Share (Offline Physical Air-Gap).
   - Emergency Treasury & Upgrade Governance: 4-of-6 threshold involving designated human key custodians (Executive CISO, Head of Clearing, Independent Depository Trustee, Lead Legal Counsel, Technical Architecture Lead, External Auditor).

3. Distributed Key Generation (DKG):
   - Private keys are generated collaboratively across participating nodes without ever being assembled at any single physical location, server memory, or network node.
   - Utilizes Verifiable Secret Sharing (VSS) with Feldman commitments over elliptic curves.
   - Each participant receives a private key share $s_i$ and a public verification vector; the complete private key $x$ exists only as a mathematical abstraction:
     $$x = \sum_{i \in S} \lambda_i \cdot s_i \pmod q$$
     Where $\lambda_i$ represents the Lagrange basis polynomial coefficient for participant set $S$.

4. Proactive Secret Sharing & Share Refresh:
   - Key shares are dynamically refreshed every 30 days without changing the underlying public key or on-chain address.
   - If an adversary compromises a key share during epoch $T$, that share is rendered mathematically useless in epoch $T+1$, thwarting slow mobile adversary attacks.

### 3.4 Formal Key Ceremonies

All Root KEK generations, validator provisioning ceremonies, and MPC initialization events must adhere to strict operational ceremony protocols:

1. Physical Environment:
   - Conducted inside a Faraday-shielded, TEMPEST-certified Sensitive Compartmented Information Facility (SCIF) with zero electronic emitting devices, mobile phones, or external network connections.
   - Continuous audio-visual recording from four independent angles; video feeds archived in tamper-evident physical safes.

2. Custodian Roles:
   - Minimum five ceremony participants: Ceremony Administrator, Cryptographic Security Officer, Independent External Auditor (CERT-In empanelled), Legal Compliance Observer, and Platform Engineering Lead.
   - Dual-control access enforced via physical dual-key locks on the HSM safe enclosures.

3. Execution Steps:
   - Step 1: Physical verification of HSM tamper seals and hardware serial numbers against manufacturer delivery manifests.
   - Step 2: Booting hardware appliances using air-gapped, verifiable live Linux media with verified SHA-512 checksums.
   - Step 3: Initialization of the HSM randomness generator using hardware quantum noise sources combined with environmental entropy pools.
   - Step 4: Execution of the DKG protocol or Master Seed generation inside the HSM hardware boundary.
   - Step 5: Splitting the master recovery secret into $N$ smart cards via Shamir Secret Sharing with a threshold of $M$.
   - Step 6: Individual custodians verify their PINs and seal their cards in tamper-evident forensic evidence bags with unique serial barcodes.
   - Step 7: Cryptographic signing of the final ceremony attestation certificate by all participants.

### 3.5 Key Rotation Schedules & Emergency Zeroization

| Key Classification | Cryptographic Algorithm | Storage Medium | Rotation Frequency | Rotation Procedure | Emergency Zeroization Trigger |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Root KEK | AES-256 | FIPS 140-2 Level 3 Physical HSM | 2 Years | Formal Key Ceremony with M-of-N Custodians | Physical enclosure tamper detection; legal mandate. |
| Service KEKs | AES-256 | CloudHSM Partition | 1 Year | Automated HKDF re-derivation and re-encryption | Unauthorized partition login attempts; node compromise. |
| User DEKs | AES-256-GCM | Database (Wrapped) | 90 Days / On Update | Envelope re-encryption via HSM batch job | DPDP Act erasure request; credential leak. |
| QBFT Validator Keys | ECDSA (secp256k1) | CloudHSM Appliance | 6 Months | On-chain validator rotation proposal & QBFT vote | Host system root compromise; consensus equivocation. |
| Relayer MPC Shares | GG20 / FROST | Isolated Enclaves | 30 Days | Proactive Secret Sharing refresh (zero address change) | Share exfiltration detection; relayer anomaly. |
| API Gateway TLS Keys | ECDSA (P-384) | Vault / Cloud KMS | 60 Days | Automated ACME protocol renewal via internal CA | Revocation of intermediate CA; private key leakage. |
| Internal JWT Signing | Ed25519 | HashiCorp Vault | 15 Days | Overlapping key sets with published JWKS keys | Vault unseal key compromise; token forgery alert. |

---

## 4. Defense-in-Depth Across Architectural Layers

```
+----------------------------------------------------------------------------------------------------+
|                                 Defense-in-Depth Layered Architecture                              |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|  [ LAYER 1: Edge & Perimeter Security ]                                                            |
|  - Anycast BGP Routing, Cloudflare Magic Transit, AWS Shield Advanced (DDoS Scrubbing)             |
|  - Enterprise Web Application Firewall (WAF), OWASP Top 10 Rules, Geo-Fencing, Rate Limiting       |
|  - Strict TLS 1.3 Termination (Zero TLS 1.0/1.1/1.2; Strict Cipher Suites)                         |
|                                |                                                                   |
|                                v                                                                   |
|  [ LAYER 2: API Gateway & Ingress Control ]                                                        |
|  - Kong Enterprise / Envoy API Gateway Mesh Ingress                                                |
|  - Mutual TLS (mTLS) with Institutional Certificate Pinning for DMA Clients                        |
|  - OAuth2 / OIDC Token Verification (Ed25519 JWT, 15-minute TTL, jti Nonce Revocation Cache)        |
|  - Strict Protobuf Schema Validation & Deep Payload Sanitization                                  |
|                                |                                                                   |
|                                v                                                                   |
|  [ LAYER 3: Microservice Service Mesh & Container Runtime ]                                       |
|  - Istio / Cilium eBPF Service Mesh with SPIFFE/SPIRE Cryptographic Workload Identities            |
|  - Per-hop Bidirectional mTLS across all inter-service communication                               |
|  - Hardened Kubernetes (Pod Security Standards: Restricted, Read-Only Root, Non-Root UID 10001)    |
|  - Calico / Cilium Default-Deny Network Policies (Zero Egress to Public Internet)                  |
|                                |                                                                   |
|                                v                                                                   |
|  [ LAYER 4: Storage & Data Persistence Security ]                                                  |
|  - PostgreSQL 16 Transparent Data Encryption (TDE) with AES-256-XTS                                |
|  - Column-Level Field Encryption via pgcrypto / AES-256-GCM for sensitive financial identifiers    |
|  - Redis Enterprise TLS 1.3 Encryption In-Transit with Auth Isolation                              |
|  - Apache Kafka SASL/SCRAM Authentication & Topic-Level Access Control Lists (ACLs)                |
|                                |                                                                   |
|                                v                                                                   |
|  [ LAYER 5: Blockchain Smart Contract Layer ]                                                      |
|  - OpenZeppelin AccessControlDefaultAdminRules with 48-Hour Governance Timelock                    |
|  - Formally Verified State Invariants (Certora Prover & Halmos Symbolic Analysis)                  |
|  - Checks-Effects-Interactions & Non-Reentrant Mutex Locks on External Functions                   |
|  - UUPS Proxy Pattern with 50-Slot Storage Gaps & Multi-Sig Authorization                          |
|  - Autonomous Emergency Circuit Breakers (PausableEmergencyStop)                                  |
+----------------------------------------------------------------------------------------------------+
```

### 4.1 Layer 1: Edge & Perimeter Defense
- Anycast BGP Routing & Volumetric DDoS Scrubbing: All external traffic routes through Anycast points of presence equipped with Cloudflare Magic Transit and AWS Shield Advanced, capable of absorbing > 15 Tbps multi-vector volumetric floods (SYN floods, UDP amplification, DNS reflection).
- Enterprise Web Application Firewall (WAF): Managed rule sets mitigating OWASP Top 10 vulnerabilities, automated credential stuffing protection, bot mitigation algorithms inspecting TLS client fingerprints (JA3/JA4), and dynamic IP reputation scoring.
- Cryptographic Transport Security: Strict TLS 1.3 termination. Deprecated protocols (SSLv3, TLS 1.0, TLS 1.1, TLS 1.2) are permanently disabled. Permitted cipher suites are restricted to:
  - TLS_AES_256_GCM_SHA384
  - TLS_CHACHA20_POLY1305_SHA256
- Strict HTTP Strict Transport Security (HSTS) enforced with a minimum duration of 2 years (`max-age=63072000; includeSubDomains; preload`).
- Jurisdictional Geo-Fencing: Immediate dropping of network packets originating from jurisdictions classified as high-risk or uncooperative by the Financial Action Task Force (FATF) and the Ministry of Home Affairs (MHA), enforced at the border BGP routing layer.

### 4.2 Layer 2: Ingress & API Gateway Security
- Gateway Routing Architecture: Managed Kong Enterprise and Envoy gateway instances act as the exclusive ingress point into the internal network.
- Institutional DMA Mutual TLS (mTLS): Direct Market Access (DMA) and broker-dealer gateways mandate client-authenticated TLS (mTLS) with hardware-backed certificate pinning (X.509 certificates rooted in the NBSE Private Institutional CA).
- Authentication & Authorization:
  - Stateless JSON Web Tokens (JWT) signed using Ed25519 asymmetric keys.
  - Token validity capped at 15 minutes (900 seconds).
  - Replay prevention: Every token embeds a unique `jti` (JWT ID) UUIDv4 nonce; used nonces are recorded in high-speed Redis caches with an automated 15-minute expiration. Duplicate nonces trigger instant connection termination and security alerting.
- Deep Payload Sanitization: Strict Protobuf schema parsing; incoming payloads exceeding maximum size limits (64 KB for orders, 1 MB for document verification) or containing unrecognized fields are dropped at the gateway boundary prior to queue dispatch.

### 4.3 Layer 3: Service Mesh & Container Runtime Security
- Workload Attestation & SPIFFE/SPIRE: Microservices deployed across Kubernetes clusters obtain ephemeral x509 SVID (SPIFFE Verifiable Identity Document) certificates issued by an internal SPIRE server backed by TPM hardware root of trust.
- In-Mesh Communication: 100% of inter-service network packets are encrypted via per-hop bidirectional mTLS mediated by Istio / Cilium service meshes with strict cipher enforcement. Plaintext TCP transport inside the cluster is mathematically and administratively impossible.
- Hardened Container Security Profiles:
  - Kubernetes Pod Security Standards enforced at `Restricted` level across all production namespaces.
  - Root execution strictly blocked (`runAsNonRoot: true`, `runAsUser: 10001`).
  - Read-only root filesystems (`readOnlyRootFilesystem: true`); temporary storage restricted to ephemeral, non-executable memory-backed tmpfs mounts (`noexec, nosuid, nodev`).
  - Total capability dropping: All Linux kernel capabilities dropped (`capabilities: drop: ["ALL"]`).
  - Kernel security modules: Mandatory deployment of custom Seccomp profiles blocking prohibited system calls (`ptrace`, `clone3`, `bpf`) and enforcing strict AppArmor profiles.
- Microsegmentation & Network Policies: Default-deny inbound and outbound network policies enforced at Layer 3 and Layer 4 via Cilium eBPF filters. A microservice can communicate only with explicitly whitelisted peer services and database endpoints. Egress from application containers directly to the public internet is disabled at the hypervisor level.

### 4.4 Layer 4: Storage & Data Persistence Security
- Database Layer (PostgreSQL 16):
  - Transparent Data Encryption (TDE) implemented at the tablespace level utilizing AES-256-XTS.
  - Column-level field encryption via `pgcrypto` for high-sensitivity financial attributes (demat account numbers, settlement tracking codes) using AES-256-GCM.
  - Network transport between microservices and databases mandates TLS 1.3 with client certificate validation (`sslmode=verify-full`).
  - Dedicated PostgreSQL role-based grants: Microservices are barred from executing DDL statements (`CREATE`, `ALTER`, `DROP`); write operations are restricted to specific tables, and auditing is enforced via `pgaudit` extensions streaming to immutable S3 buckets with Object Lock.
- In-Memory Cache (Redis Enterprise Cluster):
  - Transport encryption: TLS 1.3 enforced across all client-to-cluster and node-to-node replication paths.
  - Authentication: Strong, randomized authentication tokens (minimum 64 characters) rotated on a 30-day schedule.
  - Command isolation: Dangerous administrative commands (`FLUSHALL`, `CONFIG`, `KEYS`, `EVAL`, `DEBUG`) are permanently disabled or renamed via configuration.
- Event Streaming (Apache Kafka):
  - In-transit encryption using TLS 1.3 with mutual certificate authentication across all brokers, producers, and consumers.
  - Authentication via SASL/SCRAM-SHA-512.
  - Fine-grained Topic-Level Access Control Lists (ACLs): The matching engine possesses write-only access to trade execution topics; settlement relayers possess read-only access; unauthorized topic discovery or subscription is rejected.

### 4.5 Layer 5: Blockchain Smart Contract Layer Security
- Access Control Architecture: Smart contracts enforce OpenZeppelin `AccessControlDefaultAdminRules` with a mandatory 48-hour time-lock delay for all administrative role changes, preventing instant privilege elevation.
- Defensive Programming Invariants:
  - Strict adherence to the Checks-Effects-Interactions design pattern across all contract methods.
  - Deployment of non-reentrant mutex locks (`ReentrancyGuardUpgradeable`) on every external function that alters token balances or contract state.
  - Integer overflow protection enforced at compiler level (Solidity 0.8.24+).
  - Storage gap reservations (minimum 50 slots) maintained across all base upgradeable contracts to prevent storage collision vulnerabilities during proxy upgrades.
- Upgradeability Governance:
  - Contracts utilize the Universal Upgradeable Proxy Standard (UUPS) with ERC-1967 storage slots.
  - The implementation upgrade function is gated behind a 4-of-6 institutional multi-signature timelock contract; upgrade execution emits an on-chain event and cannot execute until the 48-hour challenge window expires.
- Formal Verification & Static Analysis:
  - 100% of production smart contracts undergo formal verification using the Certora Prover to mathematically prove safety invariants (e.g., total token supply can never exceed verified custodial vault balances).
  - Symbolic execution and property testing executed via Halmos and Echidna.
  - Automated continuous static analysis integration in CI pipelines utilizing Slither and Mythril.
- Autonomous Circuit Breakers: Smart contracts inherit custom `PausableEmergencyStop` modules capable of freezing specific trading pairs, redemptions, or minting operations within a single block upon activation by authorized automated risk monitors.

---

## 5. Regulatory & Statutory Compliance Requirements

### 5.1 SEBI Cybersecurity & Cyber Resilience Framework (CSCRF)

As a sovereign financial infrastructure providing market rails for Indian securities, the NBSE platform adheres strictly to the comprehensive SEBI Cybersecurity and Cyber Resilience Framework (CSCRF) for Market Infrastructure Institutions (MIIs) and Qualified Regulated Entities:

1. Governance & Oversight Structure:
   - Standing Technology Committee and Information Security Steering Committee providing monthly oversight.
   - Dedicated Chief Information Security Officer (CISO) possessing minimum 15 years domain experience, operating with direct reporting authority to the Board of Directors and functional independence from operational IT engineering.
   - 24/7/365 Cyber Security Operations Centre (C-SOC) staffed by certified security analysts, maintaining automated Security Information and Event Management (SIEM) and Security Orchestration, Automation, and Response (SOAR) platforms.

2. Certification & Audit Compliance:
   - Mandatory annual compliance certification against ISO/IEC 27001:2022 (Information Security Management), ISO 22301 (Business Continuity Management), and SOC 2 Type II controls.
   - Comprehensive semi-annual system and cybersecurity audits conducted by independent CERT-In empanelled auditing agencies, with audit findings submitted directly to SEBI within 30 days of completion.

3. Business Continuity & Disaster Recovery Timelines:
   - Recovery Time Objective (RTO): Less than 4 hours for full resumption of trading and settlement operations following a catastrophic failure.
   - Recovery Point Objective (RPO): Less than 15 minutes (target: zero data loss for finalized blocks).
   - Geographic Separation:
     - Primary Data Center (PDC): Mumbai, Maharashtra.
     - Near Site (NS): Navi Mumbai (synchronous replication, RPO = 0).
     - Disaster Recovery Site (DRS): Chennai / Hyderabad (minimum 500 km geographical separation, situated in a different seismic zone, operating asynchronous continuous replication).
   - Quarterly live disaster recovery switchover drills conducted during market hours, demonstrating live traffic failover without data loss.

### 5.2 RBI Guidelines for Digital Lending, Payments & CBDC

The integration of Indian Rupee payment rails and Reserve Bank of India Central Bank Digital Currency (eINR / CBDC-R / CBDC-W) enforces strict central banking security standards:

1. Fund Segregation & Escrow Invariants:
   - Complete architectural segregation between investor settlement funds, custodial reserves, and exchange corporate treasury balances.
   - Client funds are held in dedicated scheduled commercial bank escrow accounts or direct RBI settlement nodes; co-mingling of platform operational fees with investor capital is mathematically and administratively impossible.

2. Delivery versus Payment (DvP) Model 1:
   - Secondary market asset settlements operate strictly under DvP Model 1 standards: The transfer of digital security tokens and the transfer of eINR CBDC funds occur simultaneously, atomically, and gross within the same blockchain transaction.
   - If the cash leg fails, the asset leg reverts completely, eliminating principal settlement risk.

3. Direct Payment Routing:
   - In accordance with RBI Digital Lending directives, all disbursement and repayment flows must execute directly between the verified bank account / CBDC wallet of the borrower/investor and the regulated lending entity, without passing through any intermediate pooling accounts of third-party technology service providers.

### 5.3 IFSCA AML/CFT Regulations (GIFT City IFSC)

For cross-border investment flows mediated through the Gujarat International Finance Tec-City (GIFT City) International Financial Services Centre (IFSC), the platform enforces IFSCA Anti-Money Laundering, Counter-Terrorist Financing, and Know Your Customer rules:

1. FATF Recommendation 16 (The Travel Rule):
   - For all cross-border transfers of digital tokens, the platform cryptographically packages and transmits verified originator and beneficiary information using the InterVASP Messaging Standard (IVMS 101).
   - Information payloads are encrypted using the counterparty institution's public key and transmitted out-of-band; zero Travel Rule data is exposed on the public or permissioned ledger.

2. Automated Sanctions Screening & PEP Identification:
   - Real-time pre-transaction screening of all transacting entities against:
     - United Nations Security Council (UNSC) Sanctions Lists.
     - Ministry of Home Affairs (MHA) Unlawful Activities (Prevention) Act (UAPA) Terrorist Lists.
     - Office of Foreign Assets Control (OFAC) Specially Designated Nationals (SDN) Lists.
     - European Union Consolidated Financial Sanctions Lists.
     - Politically Exposed Persons (PEP) global databases.
   - Any transaction involving a sanctioned entity or jurisdiction is automatically blocked at the API gateway layer before order book entry.

3. Suspicious Transaction Reporting (STR) Pipeline:
   - Automated AML transaction monitoring systems continuously calculate transaction velocity, structuring patterns (smurfing), round-trip trading anomalies, and sudden volume deviations.
   - Anomalies exceeding high-confidence risk thresholds trigger automated alert generation for the Principal Compliance Officer.
   - Statutory Suspicious Transaction Reports (STRs) are transmitted to the Financial Intelligence Unit - India (FIU-IND) and the Financial Intelligence Unit - IFSCA within 48 hours of detection.

### 5.4 Digital Personal Data Protection Act, 2023 (DPDP Act 2023)

The NBSE platform enforces the statutory data protection principles mandated by India's DPDP Act 2023:

1. Notice & Consent Architecture (Section 6):
   - Clear, itemized, standalone consent notices presented in plain language, accessible across all 22 languages specified in the Eighth Schedule to the Constitution of India.
   - Consent records are cryptographically signed by the user and stored in an immutable Consent Ledger capturing the exact scope, purpose, and timestamp of authorization.

2. Purpose Limitation & Data Minimization:
   - Personal data collected during onboarding is strictly restricted to attributes legally mandated by SEBI, RBI, and PMLA for financial customer due diligence.
   - Processing for unstated secondary purposes (marketing, cross-selling, algorithmic profiling outside exchange risk management) is blocked at the data access layer.

3. Data Principal Rights Fulfillment:
   - Right to Access & Correction: Self-service portal enabling investors to view active processing purposes and request correction of inaccurate data.
   - Right to Grievance Redressal: Automated ticketing pipeline ensuring data grievances are acknowledged within 24 hours and fully resolved within 7 business days by the designated Data Protection Officer (DPO).
   - Right of Erasure vs. Statutory Retention: Customer erasure requests are processed by crypto-shredding off-chain data encryption keys immediately upon the expiration of statutory financial retention windows mandated by Section 12 of the Prevention of Money Laundering Act (currently 5 years post-account closure).

---

## 6. Vulnerability Disclosure, Penetration Testing & Bug Bounty Protocol

### 6.1 Coordinated Vulnerability Disclosure (CVD) & Safe Harbor Policy

The NBSE project maintains an open, collaborative security posture with ethical security researchers. Research conducted in good faith within the bounds of this framework is granted legal safe harbor protection.

1. Safe Harbor Commitments:
   - The exchange will not pursue legal action, civil claims, or law enforcement referrals against security researchers who discover and report vulnerabilities in accordance with this policy.
   - The exchange waives potential claims under the Information Technology Act, 2000 (Section 43, Section 66) and applicable copyright or anti-circumvention provisions for research activities conducted strictly within policy boundaries.

2. Scope Boundaries:
   - In-Scope Assets:
     - Public API Gateways and Ingress Routers (`api.nbse.exchange`, `dma.nbse.exchange`).
     - Core Smart Contract Codebases (ERC-3643, Identity Registries, vAMM engines, DvP Settlement modules).
     - Official Flutter Mobile and Desktop Applications (Android, iOS, Windows, Linux, macOS).
     - Web Trading Terminals and Admin Compliance Portals.
   - Out-of-Scope Assets & Prohibited Vectors:
     - Physical security of data centers, corporate offices, or hardware HSM appliances.
     - Social engineering, phishing, spear-phishing, or vishing targeting exchange employees, contractors, or custodial partners.
     - Distributed Denial of Service (DDoS) attacks against network infrastructure.
     - Third-party clearing rails (NSDL, CDSL, RBI payment gateways) outside our direct technological control.
     - Exploitation of vulnerabilities causing actual economic loss, theft of user assets, or permanent destruction of ledger data.

3. Submission Protocol:
   - Vulnerability submissions must be transmitted via PGP-encrypted email to `security@nbse.exchange` using the official exchange security public key, or submitted through our verified vulnerability disclosure portal.
   - Submissions must contain a detailed description, reproduction steps, proof-of-concept scripts or transaction traces, and an assessment of potential impact.
   - Initial human acknowledgment SLA: Within 12 hours of receipt.
   - Triage and severity classification SLA: Within 24 hours of receipt.

### 6.2 Penetration Testing Cadence & Rigor

1. Continuous Automated Vulnerability Management:
   - Daily dynamic application security testing (DAST) executed against staging and pre-production environments.
   - Automated Software Composition Analysis (SCA) tracking open-source dependencies against known Common Vulnerabilities and Exposures (CVE) databases, generating Software Bill of Materials (SBOM) in CycloneDX format on every merge request.
   - Static Application Security Testing (SAST) integrated into CI/CD pipelines; builds fail automatically on any reported High or Critical vulnerability.

2. External Third-Party Penetration Testing:
   - Comprehensive external penetration testing executed quarterly by independent, certified offensive security firms empanelled by CERT-In.
   - Tests cover black-box, grey-box, and white-box methodologies across network perimeter, Kubernetes container clusters, API gateways, and web/mobile endpoints.

3. Red Team Adversarial Simulations:
   - Annual multi-week Red Team adversarial engagement simulating sophisticated Advanced Persistent Threat (APT) actors.
   - Scope includes assumed-breach scenarios, lateral movement defense verification, container escape testing, and internal C-SOC detection and response measurement.

4. Independent Smart Contract Audits:
   - Zero smart contracts are deployed to the `nbse-mainnet` production network without completing at least two independent security audits from tier-1 specialized blockchain security auditing firms.
   - All audit reports, identified issues, remediation commits, and verification sign-offs are published in full transparency to the public documentation portal before contract activation.

### 6.3 Bug Bounty Program Structure & Reward Matrix

The platform operates both a private invite-only bug bounty program for vetted institutional researchers and a public program hosted on leading security platforms (e.g., Immunefi and HackerOne).

1. Severity Classification:
   - Vulnerabilities are rated using the Common Vulnerability Scoring System (CVSS v3.1 / v4.0), supplemented by blockchain-specific financial impact assessments.

2. Reward Tiers:

| Severity Level | CVSS v3.1 Score | Operational / Financial Impact Definition | Maximum Reward (USD) | Maximum Reward (INR) | Triage SLA | Remediation SLA |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Critical | 9.0 - 10.0 | Direct theft or freezing of user custodial assets; arbitrary token minting without reserves; validator private key exfiltration; consensus halt; bypass of Zero-PII invariant. | Up to $500,000 | Up to ₹4,00,00,000 | 2 Hours | 24 Hours (Hotfix) |
| High | 7.0 - 8.9 | Temporary freezing of trading order book without fund loss; oracle manipulation resulting in partial slippage; PII disclosure from off-chain caches; bypass of KYC claim verification. | Up to $100,000 | Up to ₹80,00,000 | 6 Hours | 72 Hours |
| Medium | 4.0 - 6.9 | State bloat attacks causing increased validator compute load; minor order book griefing; rate limit evasion without system disruption; non-exploitable logic deviations. | Up to $25,000 | Up to ₹20,00,000 | 12 Hours | 7 Days |
| Low | 0.1 - 3.9 | Informational security deviations; minor TLS configuration anomalies; non-sensitive error stack trace leaks; UI spoofing without financial credential capture. | Up to $5,000 | Up to ₹4,00,000 | 24 Hours | 14 Days |

---

## 7. Incident Response, Circuit Breakers & Emergency Operational Runbooks

### 7.1 Incident Classification & Escalation Hierarchy

All operational, security, and integrity anomalies are classified into five standardized severity tiers:

```
+----------------------------------------------------------------------------------------------------+
|                                Incident Severity Classification Matrix                             |
+----------------------------------------------------------------------------------------------------+
| P0 - CATASTROPHIC  | Critical exploit in progress, fund theft, consensus halt, validator key leak. |
| P1 - MAJOR         | Matching engine outage, oracle desync > 2%, depository connection severed.    |
| P2 - MODERATE      | Single microservice degradation, failover triggered, elevated error rates.    |
| P3 - MINOR         | Transient alert, non-exploitable anomaly, user-reported cosmetic discrepancy.  |
| P4 - INFORMATIONAL | Routine audit log deviation, scheduled security alert review.                 |
+----------------------------------------------------------------------------------------------------+
```

1. Computer Security Incident Response Team (CSIRT) Command Hierarchy:
   - Gold Command (Strategic / Executive):
     - Composition: Chief Executive Officer (CEO), Chief Information Security Officer (CISO), Chief Technology Officer (CTO), Chief Legal Officer (CLO), Head of Depository Clearing.
     - Responsibility: Regulatory notifications (SEBI, CERT-In, RBI), public communication, board escalation, authorization of emergency hard forks or platform-wide shutdowns.
   - Silver Command (Tactical / Operational):
     - Composition: Security Operations Centre (SOC) Lead, Lead Blockchain Architect, Infrastructure / SRE Lead, Financial Risk Lead.
     - Responsibility: Incident containment strategy, coordination across operational engineering teams, management of recovery runbooks.
   - Bronze Command (Technical / Incident Responders):
     - Composition: On-call Site Reliability Engineers, Smart Contract Engineers, Cryptographic Engineers, Database Administrators.
     - Responsibility: Hands-on diagnostic isolation, patch development, network isolation, log preservation, post-incident forensic collection.

### 7.2 Architectural Circuit Breakers & Autonomous Halts

To prevent systemic market contagion and protect physical asset backing, the NBSE architecture incorporates autonomous hardware and software circuit breakers:

```
+----------------------------------------------------------------------------------------------------+
|                                Autonomous Circuit Breaker Architecture                             |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|  [ Real-Time State Monitor ] ---> Analyzes block executions, oracle ticks, and depository states    |
|                                                                                                    |
|    |---> Deviation > 5% in 60s --------> [ Level 1 Breaker ]: 15-Minute Trading Pair Cool-Off      |
|    |                                                                                               |
|    |---> Market Index Swing > 10% -----> [ Level 2 Breaker ]: Exchange-Wide Trading Halt & Cancel   |
|    |                                                                                               |
|    |---> vAMM Depth Drain > 15% -------> [ Level 3 Breaker ]: Smart Contract Liquidity Freeze      |
|    |                                                                                               |
|    |---> Proof-of-Reserve Mismatch ----> [ Level 4 Breaker ]: Hard-Lock on Minting & Redemption     |
+----------------------------------------------------------------------------------------------------+
```

1. Level 1 Breaker (Pair-Specific Volatility Collar):
   - Trigger: Single-asset trading price deviates by more than 5.0% from the 5-minute rolling TWAP within a rolling 60-second window.
   - Action: Automated 15-minute trading halt on the specific trading pair. Resting market orders are purged; limit orders remain frozen. The matching engine resumes with a 3-minute pre-open call auction.

2. Level 2 Breaker (Market-Wide Circuit Breaker):
   - Trigger: Benchmark exchange index deviates by 10.0%, 15.0%, or 20.0% (aligned with SEBI equity market circuit breakers).
   - Action: Platform-wide trading halt across all continuous matching engines and vAMM pools. Coordinated order book purge and automated notification to clearing corporations.

3. Level 3 Breaker (vAMM Invariant Drain Guard):
   - Trigger: Virtual liquidity pool reserve depth decreases by more than 15.0% within a single block.
   - Action: Autonomous, contract-level execution of `pauseTrading()` on the liquidity pool smart contract. Settlement transactions referencing the pool revert immediately.

4. Level 4 Breaker (Proof-of-Reserve Mismatch Lockdown):
   - Trigger: The real-time reconciliation daemon detects a divergence between total issued on-chain ERC-3643 token supply and physical custodial depository shares held at NSDL/CDSL exceeding 0.00000001 shares.
   - Action: Immediate, irreversible hard lock on token minting and redemption contracts. Platform transitions to Read-Only mode. Gold Command convened within 5 minutes.

### 7.3 Operational Runbooks

#### Runbook 1: Validator Compromise or Consensus Fork Remediation
- Severity: P0 (Catastrophic)
- Objective: Isolate a compromised or equivocating Hyperledger Besu validator node, re-establish consensus stability, and restore network integrity.
- Execution Protocol:
  - Step 1: Detect consensus stall, round-trip timeout alerts, or equivocation evidence via Prometheus alerts (`besu_consensus_qbFT_round_timeout_total` > 3).
  - Step 2: Convene Silver Command instantly on an out-of-band, encrypted voice bridge (Signal / Wire).
  - Step 3: Identify the misbehaving validator node address from consensus round logs and block proposal headers.
  - Step 4: The remaining institutional validator nodes execute an administrative QBFT consensus proposal to vote out the compromised validator address from the validator set:
    - Each validator node invokes the QBFT validator removal method via their local HSM-authenticated management interface.
    - Upon reaching consensus ($2f + 1$ votes), the Besu network mathematically strips the compromised validator of block proposing and signing privileges.
  - Step 5: Network firewalls drop all peer-to-peer traffic and enode connections associated with the compromised validator IP address.
  - Step 6: Quarantine the affected physical or virtual node instance; trigger a volatile memory snapshot (RAM dump) for post-mortem forensic analysis before reboot.
  - Step 7: Provision a clean validator instance from immutable Golden AMI images; generate a new validator key pair inside a newly initialized CloudHSM partition.
  - Step 8: Submit an on-chain proposal to admit the new validator address; achieve $2f + 1$ quorum to re-establish full institutional consortium participation.

#### Runbook 2: Relayer Private Key Exposure or Anomaly Response
- Severity: P0 (Catastrophic)
- Objective: Invalidate an exposed or anomalous transaction relayer account, neutralize pending transactions, and rotate relayer infrastructure without halting exchange trading.
- Execution Protocol:
  - Step 1: Automated intrusion detection or anomalous volume monitor flags an unexpected transaction signed by an active relayer hot key.
  - Step 2: Bronze Command immediately invokes the emergency nonce-exhaustion protocol:
    - Dispatch a batch of zero-value dummy transactions directly to the Besu node with the highest possible sequence nonces, overwriting and invalidating any attacker transactions resting in node mempools.
  - Step 3: Invoke the emergency access control contract using the 3-of-5 Timelock administration multi-sig:
    - Revoke the `RELAYER_ROLE` from the compromised relayer address in the `OnChainIdentityRegistry` and `DvPSettlementEngine` contracts.
    - Once stripped of this role, all transactions submitted by the compromised key revert at the EVM execution boundary.
  - Step 4: Revoke CloudHSM / MPC credentials associated with the compromised relayer instance; trigger automated zeroization of local key handles.
  - Step 5: Activate the secondary, standby relayer node running in an isolated AWS availability zone, backed by an independent MPC key share partition.
  - Step 6: Grant `RELAYER_ROLE` to the new standby relayer address via multi-signature consensus.
  - Step 7: Verify transaction flow resumption and audit all transactions mined within the past 100 blocks to identify and remediate any unauthorized state modifications.

#### Runbook 3: Smart Contract Exploit, Flash Freeze & Emergency UUPS Migration
- Severity: P0 (Catastrophic)
- Objective: Neutralize an active smart contract vulnerability, protect custodial backing reserves, and deploy a formally verified patched implementation via UUPS proxy migration.
- Execution Protocol:
  - Step 1: Invariant monitor detects an anomalous state change, balance divergence, or reentrancy signature.
  - Step 2: Automated Guardian Bot or on-call Security Officer executes the `PausableEmergencyStop` function on the affected proxy contract:
    - All token transfers, order fills, minting, and burning operations are instantly frozen on-chain.
  - Step 3: Gold Command convenes; official notification transmitted to SEBI, NSDL/CDSL, and market participants within 15 minutes of pause execution.
  - Step 4: Lead Smart Contract Engineers isolate the vulnerability and develop a patched implementation contract in an isolated sandbox environment.
  - Step 5: The patch undergoes automated regression testing, mutation testing, and rapid formal verification via the Certora Prover to mathematically prove the vulnerability is neutralized without introducing state regressions.
  - Step 6: Deploy the verified patch implementation contract to the `nbse-mainnet` network.
  - Step 7: Gold Command executes the UUPS `upgradeToAndCall()` function via the 4-of-6 Timelock Governance Multi-Signature Contract:
    - The ERC-1967 implementation slot on the active proxy is updated to point to the new patch contract address.
    - Proxy storage layouts and 50-slot storage gaps are verified against pre-upgrade snapshots.
  - Step 8: Execute post-migration state validation scripts asserting that total token balances match physical custodial depository assets down to the exact fractional decimal.
  - Step 9: Lift the emergency pause using the multi-sig governance key; restore continuous trading under a 1-hour phased rate-limiting mode.

#### Runbook 4: Off-Chain PII Vault Intrusion & Key Shredding Protocol
- Severity: P0 (Catastrophic)
- Objective: Respond to an unauthorized intrusion into the PII Vault enclave, prevent bulk data exfiltration, and execute cryptographic key-shredding if physical compromise is suspected.
- Execution Protocol:
  - Step 1: Intrusion Detection System (Cilium eBPF / Wazuh / GuardDuty) detects an anomalous process injection, unauthorized root escalation, or bulk database exfiltration attempt within the PII Vault VPC.
  - Step 2: Perimeter network isolation triggers automatically:
    - Security groups drop all inbound and outbound network connectivity to the PII Vault VPC.
    - Database read connections are terminated instantly.
  - Step 3: SOC Lead assesses the extent of the breach within 15 minutes:
    - If evidence indicates an attacker has obtained root access to database files and an attempt to extract the CloudHSM KEK is underway, the CISO authorizes the Key Shredding Protocol.
  - Step 4: Execution of Cryptographic Key Shredding:
    - The Security Officer transmits an authenticated zeroization command to the CloudHSM cluster via the physical management interface.
    - The Master KEK and all user-specific DEKs are zeroized inside the HSM hardware memory.
    - Without the keys, the exfiltrated ciphertext is rendered mathematically unrecoverable (AES-256-GCM ciphertext becomes indistinguishable from pure entropy).
  - Step 5: Preserve forensic snapshots of virtual disks and volatile memory dumps across all vault instances for law enforcement and CERT-In investigation.
  - Step 6: Prepare and transmit statutory breach notifications to the Data Protection Board of India (DPBI) and CERT-In within the mandatory 6-hour statutory window mandated by CERT-In Cyber Security Directions.
  - Step 7: Restore vault infrastructure from clean, air-gapped immutable backup snapshots located in the secondary Disaster Recovery region; re-derive service keys via formal key restoration ceremony.

#### Runbook 5: Depository Disconnect & Proof-of-Reserve Mismatch Reconciliation
- Severity: P1 (Major)
- Objective: Re-establish synchronization between the permissioned blockchain ledger and external depository clearing rails (NSDL / CDSL) following a network partition or record divergence.
- Execution Protocol:
  - Step 1: Automated Reconciliation Daemon flags an un-reconciled settlement balance at the end of an hourly or daily settlement cycle.
  - Step 2: The bridge ingestion engine immediately engages the Level 4 Circuit Breaker, freezing tokenization and redemption pipelines while permitting secondary market trading to continue within existing on-chain balances.
  - Step 3: SRE Lead verifies the physical and logical status of dedicated leased lines and IPSec VPN tunnels connecting to NSDL/CDSL data centers:
    - If a physical fiber cut is identified, initiate BGP failover to the secondary redundant telecommunication carrier path.
  - Step 4: Retrieve the authoritative depository settlement clearing file (e.g., NSDL DPM / CDSL DGS daily holding statement) via out-of-band, cryptographically signed SFTP transfer.
  - Step 5: Execute the automated Ledger Differential Engine:
    - Ingest the depository statement and compare against the on-chain ERC-3643 token total supply and holder distribution.
    - Generate an itemized break report detailing every un-reconciled share transaction.
  - Step 6: Clearing Operations and Custodial Trustees review the break report:
    - If breaks represent delayed settlement batches, the bridge engine re-submits the missing batch proofs to the on-chain relayer with high priority.
    - If breaks represent failed depository debits, execute compensatory on-chain adjustment transactions signed jointly by the Depository Trustee and the Exchange Clearing House via 3-of-5 MPC threshold signatures.
  - Step 7: Proof-of-Reserve Daemon verifies 100% mathematical parity across all asset classes ($1.00000000 \text{ Digital Token} = 1 \text{ Physical Share}$).
  - Step 8: Deactivate the Level 4 Circuit Breaker; resume normal tokenization and redemption workflows.
  - Step 9: Publish the cryptographic Proof-of-Reserve attestation hash to the public transparency portal.

---

### Document Control & Authority

- Document Classification: Strictly Confidential - Internal Financial Infrastructure Architecture Standard.
- Governing Authority: Board of Directors & Technology Committee, Growww / National Blockchain Stock Exchange (NBSE).
- Effective Date: September 2026.
- Mandatory Review Cycle: Semi-Annual (March / September) or immediately following a P0 incident or significant regulatory circular from SEBI, RBI, or IFSCA.
- Enforceability: All engineers, architects, autonomous agents, and system modules must comply with the specifications herein without exception. Non-compliance is flagged as a blocking architectural failure.
