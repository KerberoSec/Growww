# MPC Key Generation Ceremonies & Vault Disaster Recovery Specification

**Specification ID:** SPEC-ARCH-034-MPC  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Cryptographic Key Management & Institutional Disaster Recovery  
**Owner:** Cryptographic Key Management & Custody Operations Group  

---

## 1. Executive Summary & Cryptographic Guarantees
This specification establishes the ceremonies, key-share distribution, and emergency disaster recovery procedures for the Multi-Party Computation (MPC-TSS, CGGMP21 protocol) institutional custody cluster:
- **Distributed Key Generation (DKG)**: Generates institutional secp256k1 and ed25519 threshold keys across 5 independent hardware security boundaries without ever assembling the private key in plaintext at any point in time.
- **Threshold Policy**: Mandatory **3-of-5 institutional quorum** required for transaction signing:
  - Share 1: Primary Exchange Enclave (AWS Nitro Enclaves - Mumbai).
  - Share 2: Secondary Cloud Enclave (GCP Confidential Computing - Frankfurt).
  - Share 3: Third Cloud Enclave (Azure Confidential Computing - Singapore).
  - Share 4: Independent Custodian Bank HSM (FIPS 140-3 Level 3).
  - Share 5: Offline Disaster Recovery Vault HSM (Air-Gapped Cold Storage in GIFT City).
- **Proactive Secret Sharing (PSS)**: Key shares refresh every 24 hours via polynomial re-sharing; compromised shares become completely useless after the refresh epoch without altering the public on-chain address.
- **Sub-80ms Online Signing**: Powered by an institutional pre-signing nonce pool maintained in hardware-isolated memory.

---

## 2. Distributed Key Generation (DKG) Ceremony Protocol

```
+-------------------------------------------------------------------------------------------------------------------------+
| 5-PARTY DISTRIBUTED KEY GENERATION (DKG) CEREMONY (CGGMP21)                                                             |
|                                                                                                                         |
| [ AWS Nitro Mumbai ]  [ GCP Frankfurt ]  [ Azure Singapore ]  [ Custodian Bank HSM ]  [ GIFT City Air-Gapped HSM ]      |
|         |                   |                   |                     |                         |                       |
|         +-------------------+-------------------+---------------------+-------------------------+                       |
|                                                         |                                                               |
|                                                         v                                                               |
|                                  [ Round 1: Feldman Verifiable Secret Sharing ]                                         |
|                                  - Each party i selects random polynomial f_i(x) of degree t-1 = 2                      |
|                                  - Broadcasts commitment to coefficients: C_{i,j} = g^{a_{i,j}}                         |
|                                                         |                                                               |
|                                                         v                                                               |
|                                  [ Round 2: Pairwise Share Exchange over mTLS ]                                         |
|                                  - Transmit secret evaluation s_{i,j} = f_i(j) encrypted via ECIES                      |
|                                                         |                                                               |
|                                                         v                                                               |
|                                  [ Round 3: Consistency Check & Public Key Derivation ]                                 |
|                                  - Verify commitments: g^{s_{i,j}} == Prod_{k=0}^{t-1} (C_{i,k})^{j^k}                 |
|                                  - Master Public Key Derived: Y = Prod_{i=1}^5 C_{i,0}                                  |
|                                                         |                                                               |
|                                                         v                                                               |
|                                  [ Ceremony Successful: Zero Plaintext Key Exposure ]                                   |
+-------------------------------------------------------------------------------------------------------------------------+
```

---

## 3. Offline Pre-Signing Nonce Pool Lifecycle (Sub-80ms Latency)

### 3.1 Nonce Pre-Computation Protocol
To eliminate multi-round interactive latency from blocking automated settlement batches and high-throughput withdrawals:
1. **Background Batch Generation**: The online institutional cluster (AWS Mumbai, GCP Frankfurt, Azure Singapore) executes Rounds 1-3 of the CGGMP21 pre-signing protocol during platform idle periods.
2. **Pool Capacity & Watermarks**:
   - Target Pool Capacity: **5,000 pre-signed nonces** ($R = k^{-1} \cdot G$) and partial shares cached in encrypted memory.
   - Low-Watermark Trigger: When available nonces drop below **1,000 nonces (20%)**, the orchestrator triggers an asynchronous batch pre-computation of 4,000 additional nonces.
3. **Sub-80ms Online Phase**: When an atomic settlement batch is ready, the coordinator pops an unused pre-signed nonce and broadcasts the scalar challenge. The 3 online nodes return single scalar multiplications $s_i$, completing the threshold ECDSA signature in a single network round (<80ms).

### 3.2 Nonce Reuse Elimination & Atomic Consumption
To mathematically prevent ECDSA nonce reuse ($k = (m_1 - m_2) / (s_1 - s_2) \pmod q$):
- Each pre-signed nonce is stored in Redis Enterprise with a hardware monotonic counter identifier.
- Nonces are consumed atomically using Redis Lua scripts executing `GETDEL`. Once a nonce index is popped, it is permanently marked as consumed in hardware-anchored RocksDB logs and cannot be re-read.

---

## 4. Disaster Recovery Ceremonies

### 4.1 Regional Datacenter Loss & Hot Failover
- If one cloud region (e.g. AWS Mumbai) experiences complete network partition or physical destruction:
  - The cluster automatically reconfigures to use the remaining online participants (GCP Frankfurt, Azure Singapore, Custodian Bank HSM).
  - Since $3 \ge 3$ threshold quorum is satisfied, transaction signing continues uninterrupted.
  - Recovery Time Objective (RTO) < 30s; Recovery Point Objective (RPO) = 0.

### 4.2 Cold Disaster Recovery Enclave Activation Ceremony
If two primary cloud regions fail simultaneously ($M < 3$), the air-gapped GIFT City Cold Vault HSM must be activated:
1. **Quorum Assembly**: Minimum 3-of-5 designated Crypto Officers must assemble in person at the GIFT City secure enclave facility.
2. **Physical Authentication**: Each Crypto Officer presents their physical FIDO2 cryptographic key fob and completes biometric fingerprint / iris verification.
3. **Air-Gapped Optical Data Diode**: Transaction batches are transmitted into the air-gapped vault via one-way optical QR code data diodes.
4. **Offline Signing**: The air-gapped HSM validates transaction batch signatures and produces Share 5's partial signature, transmitted out of the vault via QR diode to restore 3-of-5 quorum.

---

## 5. Proactive Secret Sharing (PSS) Emergency Resharing

- **Scheduled Resharing**: Every 24 hours at 00:00 UTC, all active nodes execute an automated polynomial resharing ceremony:
  $$\sum_{i=1}^5 f_i(z) = 0$$
- **Emergency Resharing**: If an intrusion attempt or key share compromise is detected on any node:
  - The security orchestrator revokes the affected node's TLS certificate.
  - The remaining 4 nodes execute an immediate out-of-band PSS resharing ceremony.
  - The compromised share becomes cryptographically invalid immediately, preserving the master on-chain public address and all locked custody assets.

---

## 6. Multi-Region Failover Matrix

| Scenario | Active Quorum | Latency Impact | RTO | RPO |
| :--- | :--- | :--- | :--- | :--- |
| **Normal Operations** | AWS Mumbai + GCP Frankfurt + Azure Singapore | Sub-80ms online signing | 0s | 0s |
| **Single Region Outage** | 2 Cloud Enclaves + Custodian Bank HSM | 95ms online signing | <15s | 0s |
| **Dual Region Catastrophe** | 1 Cloud Enclave + Custodian Bank HSM + GIFT City Cold HSM | Manual Air-Gapped Flow | <4 Hours | 0s |
| **Total Cloud Compromise** | Custodian Bank HSM + GIFT City Cold HSM + Offline Reserve | Emergency Key Reconstruction | <24 Hours | 0s |
