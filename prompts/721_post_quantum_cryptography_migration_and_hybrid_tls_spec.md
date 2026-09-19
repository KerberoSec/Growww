# 721 - Post-Quantum Cryptography (PQC) Migration & Hybrid TLS Architecture

## Purpose
Electronic exchanges, clearing corporations, and depository participants operate within an adversarial threat landscape where the confidentiality, integrity, and non-repudiation of financial transactions must be preserved across multi-decade horizons. Modern financial infrastructure relies almost exclusively on classical public-key cryptography (RSA-2048/4096, ECDSA secp256k1/P-256, and ECDH X25519) for perimeter transport layer security, inter-service mutual TLS (mTLS), distributed ledger consensus, and custodial multi-signature authorizations. However, the anticipated realization of a Cryptanalytically Relevant Quantum Computer (CRQC) executing Shor's algorithm will render these classical discrete logarithm and integer factorization schemes completely insecure.

Of immediate operational concern is the Harvest-Now-Decrypt-Later (HNDL) attack vector. Hostile state actors and sophisticated cyber-syndicates are actively intercepting and archiving encrypted financial market data, depository transfers, investor Demat account records, and proprietary order book streams traversing public telecommunications networks. Once a CRQC becomes operational, this stored encrypted data can be retroactively decrypted, compromising long-term customer confidentiality, personal identifiable information (violating India's Digital Personal Data Protection Act - DPDP Act 2023), trade secrets, and sovereign financial records. Furthermore, quantum adversaries possessing Shor-enabled capabilities could forge digital signatures, enabling fraudulent block proposal on consortium ledgers, unauthorized treasury withdrawals, and undetected transaction tampering.

The purpose of this specification is to define the authoritative migration architecture and cryptographic agility framework for transitioning the Growww and National Blockchain Stock Exchange (NBSE) platform to NIST-standardized Post-Quantum Cryptography (PQC). This entails deploying Module-Lattice-Based Key-Encapsulation Mechanism (ML-KEM / FIPS 203, formerly CRYSTALS-Kyber) combined with classical Diffie-Hellman in hybrid transport layer security (X25519 + Kyber-768), implementing Module-Lattice-Based Digital Signature Algorithm (ML-DSA / FIPS 204, formerly CRYSTALS-Dilithium) for dual-signature blockchain consensus and institutional multi-signature custody, updating Hardware Security Module (HSM) firmware policies, and enforcing cryptographic agility across the API Gateway, FIX Gateway, and internal service mesh.

## What You Are Building
An institutional-grade, quantum-safe cryptographic architecture and migration framework across perimeter gateways, backend microservices, and distributed ledger components:
- **Hybrid Post-Quantum TLS 1.3 Edge Termination:** High-throughput perimeter configuration in Envoy Proxy and NGINX implementing hybrid key encapsulation (`X25519Kyber768Draft00` / `x25519_mlkem768`), defeating HNDL attacks while maintaining sub-millisecond connection establishment latency.
- **Cryptographic Agility Abstraction Layer (`crates/pqc-crypto-core`):** A zero-allocation Rust cryptographic library providing a unified interface for classical, hybrid, and post-quantum algorithms (ML-KEM-768/1024, ML-DSA-65/87, and SLH-DSA / SPHINCS+), featuring dynamic algorithm negotiation, key derivation, and signature verification.
- **PQC-Enabled API Gateway & Service Mesh (Envoy / Istio):** Ingress and inter-service mTLS proxy configuration compiled with OpenSSL 3.3+ and the Open Quantum Safe provider (`oqsprovider`), enforcing post-quantum cipher suites for zero-trust microservice communication.
- **Quantum-Resistant FIX 4.4 / 5.0 SP2 Gateway Wrapper:** TLS 1.3 hybrid transport tunnel terminating institutional broker-dealer connections to `services/fix-gateway` without requiring structural changes to legacy binary/tag-value financial trading pipelines.
- **Dual-Signature Hyperledger Besu QBFT Consensus Engine:** Extended validator consensus module producing and verifying dual cryptographic signatures (classical secp256k1 + quantum-safe ML-DSA-65) for block proposals, commit seals, and round changes on the permissioned consortium ledger.
- **Quantum-Safe Multi-Sig Treasury & Validator Registry Smart Contracts:** Solidity smart contracts (`PQCMultiSigTreasury.sol` and `PQCValidatorRegistry.sol`) on Hyperledger Besu with off-chain and precompile-assisted ML-DSA verification verifying dual-signature threshold spend authorizations.
- **HSM & CloudHSM Post-Quantum Key Lifecycle Policy:** Operational and firmware upgrade policies for AWS CloudHSM clusters and Thales Luna PCIe HSMs, governing quantum-safe key generation, partition allocation, PKCS#11 v3.0 extensions, and ceremony key destruction.
- **Dual-Certificate X.509 PKI & Certificate Issuance Engine:** Automated certificate management integration (cert-manager and HashiCorp Vault) issuing hybrid composite certificates and parallel classical/PQC certificate chains with OCSP stapling.
- **Telemetry & Downgrade Prevention Monitor:** Real-time OpenTelemetry and Prometheus instrumentation tracking TLS handshake durations, TCP packet fragmentation, cipher suite distribution, and alerting on unauthorized protocol downgrade attempts.

## Scope Boundaries
- **In Scope:**
  - Deployment of hybrid TLS 1.3 key exchange (`x25519_mlkem768` / `X25519Kyber768Draft00`) on external and internal Envoy proxies.
  - Rust cryptographic library (`crates/pqc-crypto-core`) integrating `pqcrypto-kyber`, `pqcrypto-dilithium`, and OpenSSL 3.3 `liboqs`.
  - Migration of `services/api-gateway` (Prompt 219) to terminate hybrid post-quantum TLS for mobile, web, and external REST/WebSocket clients.
  - Integration of hybrid TLS 1.3 wrappers on `services/fix-gateway` (Prompt 225) for institutional FIX protocol connectivity.
  - Firmware upgrade and PKCS#11 v3.0 integration guidelines for `services/cloudhsm-signer-daemon` (Prompt 717).
  - Hyperledger Besu QBFT validator block hybrid signature verification and block header schema extensions.
  - Dual-key treasury multi-sig smart contract architecture (`PQCMultiSigTreasury.sol`).
  - X.509 hybrid certificate profiles, certificate issuance schemas, and automated rotation pipelines.
  - TCP MTU optimization, MSS clamping, and packet fragmentation tuning for PQC handshakes.
  - Fallback negotiation and backward compatibility for legacy non-PQC clients.
- **Out of Scope / Handled Elsewhere:**
  - High-frequency order book matching and matching engine execution loops (Prompt 205: `services/order-matching-engine`).
  - Business logic of trade clearing, netting, and gross settlement (Prompt 208: `services/trade-settlement`).
  - User authentication, JWT issuance, and RBAC policy enforcement (Prompt 105 & Prompt 219).
  - Physical datacenter HSM rack installation, power redundancy, and physical tamper sensors (Prompt 717).
  - Sparse Merkle Sum Tree generation and zero-knowledge circuit proving for solvency (Prompt 720: `services/zk-pol-engine`).
  - Zero-knowledge Pedersen commitment blinding and transaction graph decorrelation (Prompt 718: `services/privacy-obfuscation-service`).
  - Mobile application UI component rendering and state management (Prompt 501 - 531).

## Technology to Use
- **Core Microservice & Cryptographic Library:** **Rust 1.78+** utilizing `tokio` asynchronous runtime, `pqcrypto` suite (`pqcrypto-kyber`, `pqcrypto-dilithium`, `pqcrypto-traits`), `oqs` (liboqs Rust bindings), `rustls` 0.23+ (with post-quantum key exchange extensions), and `subtle` for constant-time cryptographic comparisons.
  *Justification:* Rust guarantees zero memory safety bugs, eliminates garbage collection pauses in latency-critical network paths, and provides high-performance SIMD-accelerated lattice operations.
- **Post-Quantum Cryptographic Libraries:** **liboqs 0.10+** (Open Quantum Safe C library) and **OpenSSL 3.3+** with `oqsprovider`.
  *Justification:* NIST FIPS 203 (ML-KEM) and FIPS 204 (ML-DSA) reference implementations optimized with AVX2 and ARM Neon assembly instructions.
- **Reverse Proxy & Edge Ingress:** **Envoy Proxy 1.30+** compiled with BoringSSL / OpenSSL 3.3+ supporting PQC curves (`x25519_mlkem768`, `X25519Kyber768Draft00`, and `secp256r1_kyber768`).
  *Justification:* High-performance C++ edge proxy capable of handling 100,000+ concurrent TLS handshakes with native support for hybrid key encapsulation.
- **Go Cryptographic Toolkit (for Go-based microservices):** **Go 1.22+** utilizing `crypto/tls` and Cloudflare `circl` (Cloudflare Interoperable Reusable Cryptographic Library) for ML-KEM and ML-DSA support.
- **Hardware Security Modules (HSM):** **AWS CloudHSM** (LiquidSecurity PCIe cryptographic accelerators with firmware 5.x+) and **Thales Luna HSM 7** (firmware 7.8+ supporting PKCS#11 v3.0 and vendor post-quantum mechanisms).
- **Consensus & Blockchain Client:** **Hyperledger Besu 24.x** enterprise Ethereum client running on OpenJDK 21, modified with a native Rust/C JNI cryptographic extension for high-speed ML-DSA-65 signature verification during QBFT block ingestion.
- **Certificate Management & PKI:** **HashiCorp Vault 1.16+** with custom PQC CA plugin and **cert-manager** on Kubernetes managing hybrid X.509 certificate lifecycles.
- **Telemetry & Monitoring:** **Prometheus 2.51+**, **OpenTelemetry Collector**, and **Grafana 10.4+** for tracking TLS handshake duration distributions, ClientHello byte lengths, and cipher suite counters.

## Backend / Infra Touchpoints
- **Upstream Gateways & Ingress Points:**
  - `services/api-gateway` (Prompt 219): Terminates hybrid TLS 1.3 for incoming REST, GraphQL, and WebSocket traffic from retail Flutter mobile apps and web platforms.
  - `services/fix-gateway` (Prompt 225): Terminates hybrid TLS 1.3 sessions for institutional DMA (Direct Market Access) participants and broker-dealers over FIX 4.4 and FIX 5.0 SP2.
  - Kubernetes Ingress Controllers (Prompt 802): Edge routing controllers terminating perimeter TLS and forwarding traffic across internal mesh.
- **Downstream Services & Cryptographic Daemons:**
  - `services/cloudhsm-signer-daemon` (Prompt 717): Executes hardware-isolated post-quantum and hybrid digital signatures for block proposal, settlement batch authorizations, and CA signing.
  - `services/trade-settlement` (Prompt 208): Ingests dual-signed DvP settlement batches.
  - Hyperledger Besu Validator Nodes (Prompts 301, 310, 311): Peer-to-peer consortium nodes validating dual-signature QBFT blocks.
- **Message Broker & Event Topics (Apache Kafka):**
  - Consumes:
    - `security.hsm_firmware.upgrade_scheduled.v1`: Ingests operational commands to transition HSM partitions to PQC-active firmware states.
    - `pqc.migration.node_status.v1`: Receives heartbeat notifications regarding peer node cryptographic readiness.
  - Publishes:
    - `pqc.tls.handshake_telemetry.v1`: Streams granular handshake telemetry (negotiated cipher, key exchange duration, client hello size, client user-agent).
    - `pqc.certificate.rotated.v1`: Emits notifications when hybrid X.509 certificates are renewed.
    - `pqc.client_downgrade.flagged.v1`: Alerts SIEM on anomalous client connections attempting classical fallback when PQC capability was previously recorded.
- **Key & Session Cache:**
  - Redis 7.2 Enterprise Cluster: TLS 1.3 session ticket and resumption state cache configured with quantum-safe 256-bit Pre-Shared Keys (PSK).

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus Protocol & Validator Authentication:**
  - The platform operates a permissioned Hyperledger Besu consortium network governed by Istanbul/QBFT consensus with 2-second block times.
  - Classical QBFT consensus relies exclusively on secp256k1 ECDSA signatures contained in the `extraData` field of the block header. Under this specification, the validator consensus protocol is upgraded to a **Dual-Signature Scheme**:
    1. **Primary Classical Seal:** Standard 65-byte secp256k1 signature `(r, s, v)` preserving native Besu EVM toolchain compatibility.
    2. **Post-Quantum Seal:** 3293-byte ML-DSA-65 (Dilithium3) digital signature computed over the exact block header hash commitment concatenated with the round sequence number.
  - Blocks lacking a valid post-quantum signature are immediately rejected by consortium nodes once the hard fork activation block height (`PQC_FORK_BLOCK`) is crossed.
- **On-Chain Validator Registry (`PQCValidatorRegistry.sol`):**
  - Smart contract deployed at a deterministic pre-allocated address on Hyperledger Besu.
  - Maintains the mapping of validator Ethereum addresses to their respective 1952-byte ML-DSA-65 public keys.
  - Governed by an $M$-of-$N$ consortium admin multisig requiring dual classical and post-quantum administrative authorization for validator onboarding or rotation.
- **Quantum-Resistant Treasury Multi-Sig (`PQCMultiSigTreasury.sol`):**
  - Manages depository asset reserve allocations, settlement collateral pools, and emergency circuit-breaker contracts.
  - Spends require $M$-of-$N$ threshold authorizations. Each participating officer must submit both an ECDSA signature and an ML-DSA-65 signature verifying the EIP-712 structured data digest of the transaction.
  - Signature verification offloads ML-DSA modular polynomial arithmetic to a specialized native Besu precompile located at address `0x0000000000000000000000000000000000000880`, capping verification gas consumption to 45,000 gas per signature.
- **Zero-PII On-Chain Invariant:**
  - No investor identity, PAN, Demat account, or plaintext trade parameters are included in post-quantum block headers or signature payloads.
  - All public keys and signature attestations identify system validators, broker-dealer relayer nodes, and institutional escrow custodians exclusively.

## Cryptographic Architecture & Mathematical Mechanics

### 1. Hybrid Key Encapsulation Mechanism (KEM): X25519 + ML-KEM-768
To defend against Harvest-Now-Decrypt-Later attacks while adhering strictly to defence-in-depth principles, all TLS connections enforce hybrid key encapsulation. The platform couples the battle-tested classical Elliptic Curve Diffie-Hellman (ECDH) over Curve25519 with the lattice-based NIST FIPS 203 standard (ML-KEM-768 / CRYSTALS-Kyber-768).

```
Client                                                               Server
  |                                                                    |
  |  1. ClientHello                                                    |
  |     - KeyShare: X25519_pk (32 B) + ML-KEM-768_pk (1184 B)          |
  |     - SupportedGroups: x25519_mlkem768, x25519                     |
  |------------------------------------------------------------------->|
  |                                                                    |
  |  2. ServerHello                                                    |
  |     - KeyShare: X25519_sk_exchange (32 B)                          |
  |                 ML-KEM-768_ciphertext (1088 B)                     |
  |     - EncryptedExtensions                                          |
  |     - Certificate (Hybrid / Dual Chain)                            |
  |     - CertificateVerify                                            |
  |     - Finished                                                     |
  |<-------------------------------------------------------------------|
  |                                                                    |
  |  [Shared Secret Computation]:                                      |
  |  K_classical = ECDH(X25519_sk_c, X25519_pk_s)                      |
  |  K_pqc       = ML-KEM-Decaps(ML-KEM_sk_c, ML-KEM_ct)               |
  |  SS_hybrid   = HKDF-Extract(0, K_classical || K_pqc)               |
  |                                                                    |
  |  3. Finished                                                       |
  |------------------------------------------------------------------->|
  |                                                                    |
  |  [Secure Application Data - AES-256-GCM / ChaCha20-Poly1305]       |
  |<==================================================================>|
```

#### Mathematical Specification:
1. **Classical Component (X25519):**
   - Public key: $pk_{X} \in \mathbb{F}_{2^{255}-19}$ (32 bytes).
   - Private key: $sk_{X} \in \mathbb{Z}$ (32 bytes).
   - Shared secret: $K_{\text{classical}} = X25519(sk_{X}, pk_{X}) \in \mathbb{F}_{2^{255}-19}$ (32 bytes).
2. **Post-Quantum Component (ML-KEM-768, FIPS 203):**
   - Operates over the polynomial ring $R_q = \mathbb{Z}_q[X]/(X^{256} + 1)$ where prime $q = 3329$ and lattice dimension $k = 3$.
   - Matrix $A \in R_q^{k \times k}$ generated deterministically via SHAKE-128 from seed $\rho$.
   - Key generation: Sample secret vectors $s, e \leftarrow \chi_2^k$. Compute public key $t = A \cdot s + e \in R_q^k$. Public key $pk_{\text{kem}} = (t, \rho)$ (1184 bytes); private key $sk_{\text{kem}} = s$ (2400 bytes).
   - Encapsulation: Sample message $m \leftarrow \{0,1\}^{256}$, derive error vectors $r, e_1, e_2$. Compute ciphertext $c = (u, v)$ where $u = A^T \cdot r + e_1$ and $v = t^T \cdot r + e_2 + \text{Decompress}(m)$ (1088 bytes).
   - Decapsulation: Computes $m' = \text{Compress}(v - s^T \cdot u)$ and hashes to obtain $K_{\text{pqc}} \in \{0,1\}^{256}$ (32 bytes).
3. **Hybrid Shared Secret Combiner:**
   The hybrid secret is derived via HKDF-Extract and HKDF-Expand according to RFC 9180 / IETF TLS Hybrid Key Exchange Draft:
   $$SS_{\text{hybrid}} = \text{HKDF-Extract}(\text{salt} = 0^{32}, K_{\text{classical}} \parallel K_{\text{pqc}})$$
   $$K_{\text{session}} = \text{HKDF-Expand}(SS_{\text{hybrid}}, \text{"tls13 x25519_mlkem768 shared secret"}, 32)$$
   *Security Guarantee:* The connection remains completely secure if EITHER Curve25519 OR ML-KEM-768 remains unbroken. Confidentiality survives even if an adversary possesses a CRQC that completely breaks X25519.

### 2. Dual Digital Signature Scheme: secp256k1 + ML-DSA-65
Digital signatures across blockchain consensus and institutional custody operate in dual-signature mode to satisfy regulatory audit requirements while defending against quantum forgery.

#### Mathematical Specification:
1. **Classical Component (ECDSA secp256k1):**
   - Elliptic curve $y^2 = x^3 + 7 \pmod p$ over $\mathbb{F}_p$ ($p = 2^{256} - 2^{32} - 977$).
   - Signature $\sigma_{\text{classical}} = (r, s, v)$ (65 bytes).
2. **Post-Quantum Component (ML-DSA-65, FIPS 204):**
   - Based on Module Learning with Errors (M-LWE) and Module Short Integer Solution (M-SIS) over $R_q = \mathbb{Z}_q[X]/(X^{256} + 1)$ with $q = 8380417$, matrix dimension $k = 6, l = 5$.
   - Public key $pk_{\text{dsa}} \in \{0,1\}^{1952 \times 8}$ (1952 bytes).
   - Private key $sk_{\text{dsa}} \in \{0,1\}^{4032 \times 8}$ (4032 bytes).
   - Signature $\sigma_{\text{dsa}} = (z, h, c_1)$ (3293 bytes), where $z$ is the short lattice vector, $h$ is the hint vector, and $c_1$ is the challenge polynomial commitment.
3. **Composite Verification Invariant:**
   For a given transaction or block payload $M$, a signature bundle $\Sigma = (\sigma_{\text{classical}}, \sigma_{\text{dsa}})$ is valid if and only if:
   $$\text{ECDSA-Verify}(pk_{\text{classical}}, \text{Keccak256}(M), \sigma_{\text{classical}}) == 1 \quad \land \quad \text{ML-DSA-Verify}(pk_{\text{dsa}}, M, \sigma_{\text{dsa}}) == 1$$
   Failure of either verification check results in immediate transaction revert or block drop.

### 3. Network MTU Sizing, TCP MSS Clamping, and Packet Fragmentation
Classical TLS 1.3 ClientHello packets rarely exceed 400 bytes, fitting comfortably inside a single standard Ethernet Maximum Transmission Unit (MTU = 1500 bytes). However, the inclusion of the ML-KEM-768 public key (1184 bytes), classical X25519 key share (32 bytes), ALPN tokens, SNI headers, and TLS extensions expands the ClientHello payload to between 1,700 and 2,300 bytes.

```
+-------------------------------------------------------------------------------+
| Standard MTU: 1500 Bytes (Single Frame)                                        |
| +---------------------------------------------------------------------------+ |
| | IP Header (20 B) | TCP Header (20 B) | Classical ClientHello (~350 Bytes)  | |
| +---------------------------------------------------------------------------+ |
+-------------------------------------------------------------------------------+

+-------------------------------------------------------------------------------+
| PQC Hybrid TLS Handshake: 2 Fragmented Packets                                |
| Packet 1: IP (20 B) | TCP (20 B) | TLS Record Header | ClientHello Chunk 1    |
|           [Total: 1500 Bytes - Maximum MSS]                                   |
| Packet 2: IP (20 B) | TCP (20 B) | TLS Record Header | ClientHello Chunk 2    |
|           [Total: ~850 Bytes]                                                 |
+-------------------------------------------------------------------------------+
```

#### Performance & Fragmentation Mitigations:
1. **TCP MSS Clamping:** Border routers, AWS Application Load Balancers, and Envoy ingress pods configure TCP Maximum Segment Size clamping to 1460 bytes (IPv4) / 1440 bytes (IPv6) with Path MTU Discovery (PMTU) enabled.
2. **Handshake Buffer Tuning:** Socket receive and transmit buffers (`SO_RCVBUF`, `SO_SNDBUF`) on API Gateway and FIX Gateway listeners are set to a minimum of 64 KB to eliminate buffer drops during concurrent PQC handshake floods.
3. **Certificate Compression (RFC 8879):** Intermediate and leaf certificates in the PQC chain are compressed using Brotli (`brotli` algorithm identifier in TLS CertificateCompression extension), reducing leaf transmission overhead by up to 40%.
4. **Resumption via Session Tickets (RFC 8446):** Post-quantum session tickets carry cached 256-bit PSKs. Resumed TLS 1.3 connections bypass the 1184-byte KEM exchange entirely, completing in 1 RTT with standard packet sizes (< 500 bytes).

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Cryptographic Core Workspace (`crates/pqc-crypto-core`):**
   Initialize the Rust crate under `crates/pqc-crypto-core/` with `Cargo.toml` dependencies: `pqcrypto-kyber`, `pqcrypto-dilithium`, `pqcrypto-traits`, `subtle`, `ring`, `tokio`, and `tonic`. Implement memory-pinned zeroizing wrappers (`zeroize::ZeroizeOnDrop`) for all private key structures to ensure immediate erasure from RAM upon deallocation.

2. **Implement Hybrid Key Encapsulation Mechanism Engine:**
   In `crates/pqc-crypto-core/src/kem/hybrid.rs`, implement `HybridKemEngine`. Expose `generate_keypair()`, `encapsulate(peer_pk)`, and `decapsulate(sk, ciphertext)`. Combine Curve25519 and ML-KEM-768 using the HKDF combiner specified in RFC 9180. Validate serialization bounds and reject malformed ciphertexts in constant time.

3. **Implement Dual Signature Module (secp256k1 + ML-DSA-65):**
   In `crates/pqc-crypto-core/src/signatures/dual.rs`, develop the `DualSigner` and `DualVerifier` structs. Implement composite signing over message digests. Add deterministic serialization for dual signatures formatted as `[1 byte version || 65 bytes secp256k1 || 2 bytes dsa_len || 3293 bytes ML-DSA-65]`.

4. **Build Custom Envoy Proxy with OpenSSL 3.3 and OQS Provider:**
   Author the Dockerfile `deployments/docker/envoy-pqc.Dockerfile` compiling Envoy Proxy 1.30+ against OpenSSL 3.3+ and `liboqs` v0.10. Enable `oqsprovider` in `openssl.cnf`. Verify that `openssl list -kem-algorithms` reports `x25519_mlkem768`, `kyber768`, and `mlkem768`.

5. **Author Perimeter Envoy Configuration for Hybrid TLS Termination:**
   Develop `deployments/k8s/envoy/envoy-pqc-gateway.yaml`. Configure downstream TLS contexts with `ecdh_curves: ["x25519_mlkem768", "X25519Kyber768Draft00", "X25519", "P-256"]`. Set cipher suites to `TLS_AES_256_GCM_SHA384` and `TLS_CHACHA20_POLY1305_SHA256`. Configure TLS session caching in Redis 7.2 with 256-bit PSK rotation.

6. **Implement Cryptographic Agility gRPC Service:**
   Create `services/pqc-crypto-daemon/` implementing the Protobuf contract `proto/growww/crypto/pqc/v1/pqc_crypto.proto`. Provide RPC endpoints: `HybridEncapsulate`, `HybridDecapsulate`, `DualSign`, `DualVerify`, and `GetAlgorithmCapabilities`. Enforce mutual TLS (mTLS 1.3) with client certificate authorization.

7. **Develop FIX Gateway Hybrid TLS 1.3 Wrapper:**
   In `services/fix-gateway/src/transport/pqc_tls.rs`, implement a non-blocking asynchronous TLS handshake wrapper using `tokio-rustls` configured with `pqcrypto` KEM group providers. Wrap the raw TCP stream before handing over decrypted byte buffers to QuickFIX / internal tag-value parser engines (Prompt 225).

8. **Formulate CloudHSM & Thales Luna HSM Upgrade Policy Manifest:**
   Author `docs/hsm/pqc_firmware_lifecycle_runbook.md` and provisioning automation scripts `scripts/hsm/provision_pqc_partitions.sh`. Define partition allocation across AWS CloudHSM clusters, establish PKCS#11 v3.0 mechanism mappings (`CKM_ML_KEM_KEY_PAIR_GEN`, `CKM_ML_DSA_KEY_PAIR_GEN`), and implement firmware update validation procedures with physical key-custodian ceremonies.

9. **Extend CloudHSM Signing Daemon for Dual Signatures (Prompt 717):**
   Update `services/cloudhsm-signer-daemon` to support PKCS#11 post-quantum key handles. Implement an orchestration pipeline that fetches classical secp256k1 signatures from partition 0 and ML-DSA-65 signatures from partition 1, assembling the combined dual-signature payload in under 6 milliseconds.

10. **Implement Hyperledger Besu QBFT Dual-Signature Consensus Extension:**
    In the Hyperledger Besu codebase (`besu/consensus/qbft`), extend the block header validator. Implement `QBFTDualSignatureValidator` invoking the native ML-DSA JNI shared library. Validate both the ECDSA commit seal and the post-quantum extraData signature for every proposal, prepare, and commit round.

11. **Deploy Post-Quantum On-Chain Registries and Treasury Contracts:**
    Develop and test `contracts/pqc/PQCValidatorRegistry.sol` and `contracts/pqc/PQCMultiSigTreasury.sol`. Implement the off-chain verification fallback and integration with the Besu ML-DSA EVM precompile (`0x0880`). Deploy contracts on Besu testnet and seed active validator public keys.

12. **Configure Automated Hybrid PKI & Certificate Issuance Pipeline:**
    Configure HashiCorp Vault with the custom `vault-plugin-secrets-pqc-pki`. Create Kubernetes `cert-manager` ClusterIssuers for internal service mesh certificates. Configure automated 30-day rotation for leaf certificates with automated CRL publication and OCSP stapling.

13. **Build Fallback Engine & Downgrade Prevention Circuit Breaker:**
    In `services/api-gateway/src/middleware/pqc_downgrade_guard.rs`, implement client tracking using client certificate hashes and IP reputation buckets. If an institutional client or registered DMA broker connects using classical-only TLS after previously establishing hybrid PQC sessions, block the connection and emit `pqc.client_downgrade.flagged.v1` to Kafka.

14. **Implement OpenTelemetry Metrics & Telemetry Pipeline:**
    Instrument all gateway proxies and daemons with Prometheus metrics: `pqc_handshake_duration_seconds{cipher,kex_group}`, `pqc_handshake_bytes_received{fragmented="true|false"}`, `pqc_active_sessions{algorithm="hybrid|classical"}`, and `pqc_signature_verification_duration_seconds`.

15. **Execute Exhaustive Quantum-Chaos & Handshake Fuzzing Test Suite:**
    Develop test harnesses in `tests/pqc/` using `scapy` and automated test runners. Test packet drops on fragmented ClientHello packets, MTU drops at 1400 bytes, corrupt ML-KEM ciphertexts, forged ML-DSA signature vectors, and verify that classical legacy clients (Chrome/Firefox without PQC) gracefully fall back to X25519.

## Interfaces / Contracts

### 1. Protobuf Interface: Post-Quantum Cryptographic Service (`proto/growww/crypto/pqc/v1/pqc_crypto.proto`)
```protobuf
syntax = "proto3";

package growww.crypto.pqc.v1;

option go_package = "services/pqc-crypto-daemon/pb;pqcv1";
option java_package = "com.growww.crypto.pqc.v1";
option java_multiple_files = true;

// Service providing hardware-isolated post-quantum cryptographic primitives
service PqcCryptoService {
  // Generates a hybrid keypair (X25519 + ML-KEM-768)
  rpc GenerateHybridKemKeypair(GenerateHybridKemKeypairRequest) returns (GenerateHybridKemKeypairResponse);

  // Encapsulates a shared secret using a peer's hybrid public key
  rpc HybridEncapsulate(HybridEncapsulateRequest) returns (HybridEncapsulateResponse);

  // Decapsulates a shared secret using private key handle stored in HSM
  rpc HybridDecapsulate(HybridDecapsulateRequest) returns (HybridDecapsulateResponse);

  // Generates a composite digital signature (secp256k1 + ML-DSA-65)
  rpc DualSign(DualSignRequest) returns (DualSignResponse);

  // Verifies a composite digital signature
  rpc DualVerify(DualVerifyRequest) returns (DualVerifyResponse);

  // Queries cryptographic agility capabilities and active algorithm versions
  rpc GetAlgorithmCapabilities(GetAlgorithmCapabilitiesRequest) returns (GetAlgorithmCapabilitiesResponse);
}

enum KemAlgorithm {
  KEM_ALGORITHM_UNSPECIFIED = 0;
  KEM_ALGORITHM_X25519_MLKEM768 = 1;
  KEM_ALGORITHM_X25519_KYBER768 = 2;
  KEM_ALGORITHM_SECP256R1_MLKEM768 = 3;
  KEM_ALGORITHM_MLKEM1024 = 4;
}

enum DsaAlgorithm {
  DSA_ALGORITHM_UNSPECIFIED = 0;
  DSA_ALGORITHM_DUAL_SECP256K1_MLDSA65 = 1;
  DSA_ALGORITHM_DUAL_ED25519_MLDSA65 = 2;
  DSA_ALGORITHM_MLDSA87 = 3;
  DSA_ALGORITHM_SLHDSA_SPHINCS_PLUS = 4;
}

message GenerateHybridKemKeypairRequest {
  KemAlgorithm algorithm = 1;
  string key_alias = 2;
  bool export_public_only = 3;
}

message GenerateHybridKemKeypairResponse {
  string key_id = 1;
  bytes classical_public_key = 2; // 32 bytes (X25519)
  bytes pqc_public_key = 3;       // 1184 bytes (ML-KEM-768)
  int64 created_at_unix_ms = 4;
}

message HybridEncapsulateRequest {
  KemAlgorithm algorithm = 1;
  bytes classical_peer_public_key = 2; // 32 bytes
  bytes pqc_peer_public_key = 3;       // 1184 bytes
}

message HybridEncapsulateResponse {
  bytes classical_ciphertext = 1; // 32 bytes ephemeral public key
  bytes pqc_ciphertext = 2;       // 1088 bytes encapsulation
  bytes shared_secret = 3;        // 32 bytes derived HKDF key
}

message HybridDecapsulateRequest {
  string key_id = 1;
  bytes classical_ciphertext = 2;
  bytes pqc_ciphertext = 3;
}

message HybridDecapsulateResponse {
  bytes shared_secret = 1; // 32 bytes derived HKDF key
}

message DualSignRequest {
  DsaAlgorithm algorithm = 1;
  string key_alias = 2;
  bytes message_digest = 3; // 32-byte pre-hashed payload (Keccak256 or SHA-256)
}

message DualSignResponse {
  string key_id = 1;
  bytes classical_signature = 2; // 65 bytes ECDSA (r, s, v)
  bytes pqc_signature = 3;       // 3293 bytes ML-DSA-65
  bytes composite_signature = 4; // Serialized bundle with version prefix
  int64 signed_at_unix_ms = 5;
}

message DualVerifyRequest {
  DsaAlgorithm algorithm = 1;
  bytes message_digest = 2;
  bytes classical_public_key = 3; // 33 bytes compressed secp256k1
  bytes pqc_public_key = 4;       // 1952 bytes ML-DSA-65
  bytes composite_signature = 5;
}

message DualVerifyResponse {
  bool is_valid = 1;
  bool classical_valid = 2;
  bool pqc_valid = 3;
  string failure_reason = 4;
}

message GetAlgorithmCapabilitiesRequest {}

message GetAlgorithmCapabilitiesResponse {
  repeated KemAlgorithm supported_kem_algorithms = 1;
  repeated DsaAlgorithm supported_dsa_algorithms = 2;
  string openssl_version = 3;
  string oqs_provider_version = 4;
  bool hsm_hardware_acceleration_active = 5;
}
```

### 2. Envoy Proxy PQC Listener Configuration (`deployments/k8s/envoy/envoy-pqc-tls.yaml`)
```yaml
static_resources:
  listeners:
  - name: ingress_hybrid_tls
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 8443
    per_connection_buffer_limit_bytes: 65536
    filter_chains:
    - transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
          common_tls_context:
            tls_params:
              tls_minimum_protocol_version: TLSv1_3
              tls_maximum_protocol_version: TLSv1_3
              cipher_suites:
              - "TLS_AES_256_GCM_SHA384"
              - "TLS_CHACHA20_POLY1305_SHA256"
              ecdh_curves:
              - "x25519_mlkem768"
              - "X25519Kyber768Draft00"
              - "X25519"
              - "P-256"
            tls_certificates:
            - certificate_chain:
                filename: "/etc/envoy/certs/server-hybrid.crt"
              private_key:
                filename: "/etc/envoy/certs/server-hybrid.key"
            validation_context:
              trusted_ca:
                filename: "/etc/envoy/certs/nbse-root-ca.crt"
              require_client_certificate: false
          session_ticket_keys:
            keys:
            - filename: "/etc/envoy/secrets/session_ticket.key"
      filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: api_backend
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: api_gateway_service
                  timeout: 5s
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters:
  - name: api_gateway_service
    connect_timeout: 0.25s
    type: STRICT_DNS
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: api_gateway_service
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: api-gateway.internal.growww.in
                port_value: 8080
```

### 3. OpenSSL 3.3 Post-Quantum Provider Configuration (`deployments/config/openssl-pqc.cnf`)
```ini
openssl_conf = openssl_init

[openssl_init]
providers = provider_sect
alg_section = evp_properties

[provider_sect]
default = default_sect
oqsprovider = oqsprovider_sect

[default_sect]
activate = 1

[oqsprovider_sect]
activate = 1
module = /usr/local/lib64/ossl-modules/oqsprovider.so

[evp_properties]
default_properties = "oqsprovider.security_bits>=128"
```

### 4. Solidity Dual-Signature Validator Registry (`contracts/pqc/IPQCValidatorRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IPQCValidatorRegistry
 * @notice Authoritative registry managing dual classical and post-quantum validator identities
 *         for Hyperledger Besu QBFT consensus verification.
 */
interface IPQCValidatorRegistry {
    struct ValidatorIdentity {
        address classicalAddress;      // secp256k1 derived Ethereum address (20 bytes)
        bytes pqcPublicKey;           // ML-DSA-65 public key (1952 bytes)
        uint64 registeredBlock;        // Activation block height
        bool isActive;                 // Active status in current validator set
        bytes32 registrationHash;      // Keccak256 commitment of dual keys
    }

    event ValidatorRegistered(
        address indexed classicalAddress,
        bytes32 indexed pqcPublicKeyHash,
        uint64 activationBlock
    );

    event ValidatorDeactivated(
        address indexed classicalAddress,
        uint64 deactivationBlock
    );

    /**
     * @notice Registers a new validator with dual classical and ML-DSA-65 public keys.
     * @param classicalAddress The 20-byte secp256k1 address.
     * @param pqcPublicKey The 1952-byte ML-DSA-65 public key.
     * @param governanceDualSignature Composite signature of the governance multi-sig.
     */
    function registerValidator(
        address classicalAddress,
        bytes calldata pqcPublicKey,
        bytes calldata governanceDualSignature
    ) external;

    /**
     * @notice Deactivates an existing validator.
     * @param classicalAddress The validator address to revoke.
     * @param governanceDualSignature Composite signature authorizing revocation.
     */
    function deactivateValidator(
        address classicalAddress,
        bytes calldata governanceDualSignature
    ) external;

    /**
     * @notice Verifies whether a given block proposal seal contains valid dual signatures.
     * @param headerHash The 32-byte Keccak256 hash of the block header.
     * @param classicalSignature 65-byte secp256k1 signature (r, s, v).
     * @param pqcSignature 3293-byte ML-DSA-65 signature.
     * @param validatorAddress The declared proposing validator.
     */
    function verifyBlockSeal(
        bytes32 headerHash,
        bytes calldata classicalSignature,
        bytes calldata pqcSignature,
        address validatorAddress
    ) external view returns (bool isValid);

    /**
     * @notice Retrieves the ML-DSA-65 public key associated with a validator address.
     */
    function getValidatorPqcKey(address validatorAddress) external view returns (bytes memory);
}
```

### 5. Solidity Post-Quantum Multi-Sig Treasury (`contracts/pqc/IPQCMultiSigTreasury.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IPQCMultiSigTreasury
 * @notice Institutional threshold multi-sig contract requiring dual classical (ECDSA)
 *         and post-quantum (ML-DSA-65) authorizations for high-value fund movements.
 */
interface IPQCMultiSigTreasury {
    struct TransactionProposal {
        address destination;
        uint256 value;
        bytes data;
        bool executed;
        uint32 confirmationsCount;
        uint256 submissionTimestamp;
        bytes32 transactionDigest;
    }

    struct OfficerSignature {
        address officerAddress;
        bytes classicalSignature; // 65 bytes ECDSA
        bytes pqcSignature;       // 3293 bytes ML-DSA-65
    }

    event ProposalSubmitted(uint256 indexed proposalId, address indexed destination, uint256 value);
    event ProposalConfirmed(uint256 indexed proposalId, address indexed officer);
    event ProposalExecuted(uint256 indexed proposalId, bytes32 txHash);

    function submitTransaction(
        address destination,
        uint256 value,
        bytes calldata data
    ) external returns (uint256 proposalId);

    function confirmTransaction(
        uint256 proposalId,
        OfficerSignature calldata signature
    ) external;

    function executeTransaction(uint256 proposalId) external returns (bytes memory result);

    function isConfirmed(uint256 proposalId) external view returns (bool);

    function getConfirmationThreshold() external view returns (uint32 requiredConfirmations);
}
```

### 6. X.509 Hybrid Certificate Profile JSON Schema (`schemas/hybrid_certificate_profile.json`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "HybridCertificateProfile",
  "type": "object",
  "required": [
    "profileVersion",
    "subjectDistinguishedName",
    "classicalKeyAlgorithm",
    "pqcKeyAlgorithm",
    "validityDays",
    "keyUsages",
    "sanExtension"
  ],
  "properties": {
    "profileVersion": {
      "type": "string",
      "enum": ["1.0.0"]
    },
    "subjectDistinguishedName": {
      "type": "object",
      "required": ["commonName", "organization", "country"],
      "properties": {
        "commonName": { "type": "string" },
        "organization": { "type": "string" },
        "organizationalUnit": { "type": "string" },
        "country": { "type": "string", "pattern": "^[A-Z]{2}$" }
      }
    },
    "classicalKeyAlgorithm": {
      "type": "string",
      "enum": ["ECDSA-P256", "ECDSA-secp256k1", "RSA-4096"]
    },
    "pqcKeyAlgorithm": {
      "type": "string",
      "enum": ["ML-DSA-65", "ML-DSA-87", "SLH-DSA-SHA2-128s"]
    },
    "compositeBindingMode": {
      "type": "string",
      "enum": ["IETF-LAMPS-DualCertificate", "ITU-X509-AlternativePublicKeyExtension"],
      "default": "ITU-X509-AlternativePublicKeyExtension"
    },
    "validityDays": {
      "type": "integer",
      "minimum": 1,
      "maximum": 365
    },
    "keyUsages": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["digitalSignature", "keyEncipherment", "keyCertSign", "cRLSign"]
      }
    },
    "extendedKeyUsages": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["serverAuth", "clientAuth", "codeSigning"]
      }
    },
    "sanExtension": {
      "type": "object",
      "required": ["dnsNames"],
      "properties": {
        "dnsNames": {
          "type": "array",
          "items": { "type": "string" }
        },
        "ipAddresses": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    }
  },
  "additionalProperties": false
}
```

## Security & Compliance Notes

### 1. NIST FIPS 203 / 204 Compliance & Standardization Status
- **ML-KEM (FIPS 203):** The deployment utilizes Module-Lattice-Based Key Encapsulation Mechanism at Security Category 3 (ML-KEM-768), providing approximately 192 bits of classical security and full 128-bit quantum security against Grover's quantum search algorithm. ML-KEM-1024 (Category 5) is provisioned as an emergency option for core inter-datacenter backbone links.
- **ML-DSA (FIPS 204):** Digital signatures utilize Module-Lattice-Based Digital Signature Algorithm at Category 3 (ML-DSA-65), delivering robust unforgeability against quantum chosen-message attacks.
- **Constant-Time Verification Invariants:** All lattice operations in `crates/pqc-crypto-core` and the Envoy TLS termination layer enforce constant-time polynomial multiplication and rejection sampling to prevent microarchitectural cache-timing and branch-predictor side-channel attacks.

### 2. Backward Compatibility & Graceful Client Fallback
- **ClientHello Negotiation:** Ingress proxies enforce RFC 8446 standard TLS 1.3 negotiation. If a connecting client (such as an un-upgraded retail browser or legacy institutional terminal) does not advertise `x25519_mlkem768` in its `supported_groups` extension, the handshake gracefully falls back to classical `X25519` or `secp256r1`.
- **Zero Disruption Guarantee:** Retail order routing and market data feeds remain 100% operational for existing clients while newer clients immediately benefit from post-quantum protection.
- **Institutional Gateway Enforcement:** For institutional direct access participants connecting via FIX 4.4 / 5.0 SP2 (Prompt 225), hybrid PQC TLS is mandatory following a staged 90-day migration window, after which classical-only handshakes on port 9443 are rejected.

### 3. Downgrade Attack Prevention
- **Cryptographic Signaling in ServerHello:** In accordance with TLS 1.3 specifications, if the server negotiates a classical cipher suite when post-quantum capabilities are enabled, random sentinel bytes are injected into the ServerHello `Random` field to detect active man-in-the-middle downgrade attempts.
- **Historical Client Pinning:** `services/api-gateway` maintains a persistent cache of client certificates and institutional API key IDs that have successfully established PQC sessions. If an authenticated institutional client abruptly downgrades to classical TLS, the connection is quarantined, flagged in SIEM, and blocked pending security officer review.

### 4. Regulatory Alignment (SEBI CSCRF, RBI, and DPDP Act 2023)
- **DPDP Act 2023 Long-Term Confidentiality:** Personal identifiable data, investor bank details, and Demat holdings encrypted in transit are shielded against retroactive decryption (HNDL attacks), fulfilling the statutory duty of care for data fiduciaries under Section 8 of the DPDP Act.
- **SEBI CSCRF Cryptographic Modernization:** Directly addresses SEBI guidelines mandating resilience against emerging cyber threats and algorithmic obsolescence by establishing cryptographic agility and dual-signing frameworks across exchange systems.
- **Audit Trails:** Every TLS handshake negotiation failure, algorithm mismatch, and key rotation event is streamed to ClickHouse and Apache Kafka for regulatory reporting.

## Acceptance Criteria

- [ ] **Hybrid Handshake Performance:** The Envoy ingress proxy completes 95% of hybrid post-quantum TLS 1.3 handshakes (`x25519_mlkem768`) in under 2.5 milliseconds on local/metropolitan network links (32 vCPU, 64 GB RAM baseline).
- [ ] **Throughput Concurrency:** Ingress gateways sustain a minimum of 15,000 active concurrent hybrid TLS sessions without exceeding 60% CPU utilization or dropping packets.
- [ ] **Graceful Classical Fallback:** Automated test suites confirm that 100% of simulated legacy clients lacking PQC capabilities successfully negotiate classical X25519 TLS 1.3 without connection errors or timeout penalties.
- [ ] **Packet Fragmentation Resilience:** TCP handshakes with simulated Path MTU constrained to 1400 bytes successfully complete multi-packet ClientHello reassembly with zero handshake aborts.
- [ ] **CloudHSM Dual Signing Latency:** `services/cloudhsm-signer-daemon` generates dual composite signatures (secp256k1 + ML-DSA-65) in under 8 milliseconds (p99).
- [ ] **Hyperledger Besu Block Verification Overhead:** The dual-signature QBFT validator consensus engine validates blocks containing ML-DSA-65 seals within 15 milliseconds, maintaining the 2-second block production window without incurring consensus round delays.
- [ ] **Treasury Multi-Sig Gas Bound:** Executing `confirmTransaction` on `PQCMultiSigTreasury.sol` consuming the native ML-DSA precompile incurs fewer than 65,000 gas per officer signature.
- [ ] **Cryptographic Agility Test Coverage:** Unit and property-based fuzz tests in `crates/pqc-crypto-core` achieve >90% code coverage, with zero memory leaks detected under Valgrind/ASAN.
- [ ] **Downgrade Attack Detection:** Automated security regression tests verify that simulated man-in-the-middle strip attacks removing PQC extensions from previously pinned institutional clients trigger immediate connection drop and alert dispatch.

## Suggested Order / Dependencies

- **Upstream Dependencies (Must Be Built / Specified Prior to Deployment):**
  - **Prompt 101 (`101_system_architecture_overview.md`):** Establishes the foundational microservice network topology and perimeter boundaries.
  - **Prompt 105 (`105_authentication_and_authorization_architecture.md`):** Defines the identity management and client authentication standards.
  - **Prompt 108 (`108_environment_strategy_and_config_mgmt.md`):** Manages Kubernetes ConfigMaps and Helm deployment configurations.
  - **Prompt 109 (`109_secrets_management_architecture.md`):** Manages HashiCorp Vault key distribution and PKI root CA storage.
  - **Prompt 219 (`219_api_gateway_and_bff.md`):** The primary edge ingress gateway terminating external hybrid TLS traffic.
  - **Prompt 225 (`225_fix_gateway_institutional_integration.md`):** Institutional gateway requiring post-quantum TLS tunneling.
  - **Prompt 301 (`301_hyperledger_besu_network_and_genesis.md`):** Provides the consortium blockchain genesis block and QBFT configuration.
  - **Prompt 717 (`717_hsm_kms_cloudhsm_key_lifecycle_and_signing_daemon.md`):** Hardware security module signing daemon managing PKCS#11 sessions and key ceremonies.
- **Downstream Consumers (Build After / Concurrently):**
  - **Prompt 204 (`204_order_service.md`):** Adopts PQC mTLS for internal order ingress.
  - **Prompt 208 (`208_trade_settlement_service.md`):** Ingests dual-signed DvP settlement batch payloads.
  - **Prompt 329 (`329_settlement_dvp_and_token_exchange_contract.md`):** Validates post-quantum settlement commitments on-chain.
  - **Prompt 701 (`701_full_system_threat_model_stride.md`):** Ingests the PQC threat mitigation model into the enterprise STRIDE risk register.
  - **Prompt 806 (`806_monitoring_alerting_prometheus_grafana.md`):** Ingests PQC handshake latency and fragmentation dashboards.
  - **Prompt 901 (`901_integration_test_suite_master.md`):** Executes cross-service hybrid TLS validation in CI/CD pipelines.
  - **Prompt 914 (`914_full_platform_scenario_stress_and_chaos_testing.md`):** Conducts chaos tests injecting MTU truncation and quantum downgrade attacks.
