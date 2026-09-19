# ADR 001: Permissioned Consortium Ledger Selection

## Status
Accepted

## Date
September 2026

## Deciders
- Lead Blockchain Architect
- Chief Technology Officer
- Head of Information Security
- Chief Compliance Officer

---

## 1. Context & Business Drivers

The Growww institutional exchange platform requires an immutable, cryptographically verifiable, and auditable settlement and proof-of-reserve layer for fractionalized Indian securities backed 1:1 by real dematerialized equities held in SEBI-regulated depositories (NSDL and CDSL).

Selecting the appropriate distributed ledger platform is a foundational architectural decision. The ledger must fulfill the following operational and regulatory requirements:
1. **Deterministic High Throughput:** Process sustained loads of 2,000 or more transactions per second (TPS) during peak settlement windows.
2. **Instant Finality with Zero Fork Risk:** Equities clearing and Delivery-versus-Payment (DvP) transactions require absolute finality. Probabilistic finality or chain reorganizations are unacceptable for legal settlement.
3. **Strict Regulatory Compliance & Zero PII:** No customer names, Permanent Account Numbers (PAN), Aadhaar numbers, or bank account details may ever be committed to the blockchain. All on-chain records must use pseudonymous cryptographic hashes (e.g. `bytes32 identityHash`, ERC-734/ERC-735 claim identifiers).
4. **Enterprise Key Custody & HSM Integration:** Node validators and settlement relayers must integrate seamlessly with FIPS 140-2 Level 3 Hardware Security Modules (CloudHSM, HashiCorp Vault Transit) without exposing signing keys in memory.
5. **EVM Tooling & Standard Security Tokens:** Native support for standard Solidity smart contracts, ERC-3643 (T-REX) compliance standards, and established security tooling (Foundry, Slither).
6. **Consortium Multi-Institution Participation:** Support for distributed validator topology across Growww, Regulated Custodians, Clearing Corporations, and read-only regulatory audit nodes (SEBI, IFSCA, RBI).

---

## 2. Evaluated Options

Three enterprise distributed ledger candidates were evaluated under rigorous empirical benchmarks:
1. **Hyperledger Besu (v24.1.0) with QBFT Consensus:** Enterprise Ethereum client running Quorum Byzantine Fault Tolerance.
2. **Hyperledger Fabric (v2.5.4) with Raft Consensus:** Permissioned enterprise ledger using an execute-order-validate pipeline.
3. **Polygon CDK / Edge (v0.7.1) with IBFT Consensus:** Modular EVM app-chain framework with Proof-of-Authority/IBFT.

---

## 3. Comparison Matrix

| Evaluation Dimension | Hyperledger Besu (QBFT) | Hyperledger Fabric (Raft) | Polygon CDK / Edge |
|---|---|---|---|
| **Consensus Mechanism** | QBFT (Quorum BFT) | etcdraft (Crash Fault Tolerant) | IBFT (Istanbul BFT) |
| **Finality Guarantee** | Immediate 1-block deterministic | Immediate upon Raft block commit | Immediate 1-block deterministic |
| **Fault Tolerance Model** | Byzantine (Tolerates $F < N/3$ malicious nodes) | Crash-fault only (Cannot tolerate malicious nodes) | Byzantine (Tolerates $F < N/3$ malicious nodes) |
| **EVM Compatibility** | 100% Native EVM (Solidity 0.8.24) | None (Go, Java, Node chaincode) | 100% Native EVM (Solidity 0.8.24) |
| **DvP Settlement Throughput** | 2,415.6 average TPS (Peak 2,850.2 TPS) | 1,807.0 average TPS (Peak 2,040.5 TPS) | 2,104.3 average TPS (Peak 2,420.0 TPS) |
| **p95 Settlement Latency** | 1,180 ms | 1,820 ms | 1,450 ms |
| **Standard Token Support** | ERC-3643, ERC-20, ERC-1400 native | Proprietary asset state structures | ERC-3643, ERC-20 native |
| **Node Permissioning** | On-chain smart contracts (EIP-712/QBFT) | Fabric CA and MSP X.509 certs | Genesis validator allowlist |
| **HSM / Key Management** | Native Web3Signer (CloudHSM, Vault) | PKCS#11 fabric-ca client | Custom RPC signing proxy |
| **Regulatory Audit Access** | Standard JSON-RPC Observer Nodes | Complex channel client setup | Standard JSON-RPC |
| **Developer Velocity** | Foundry, Hardhat, OpenZeppelin | Custom Fabric test mocks | Foundry, Hardhat |

---

## 4. Empirical Benchmark Findings

Automated benchmarks were executed using 4-node dedicated clusters (AWS c6i.2xlarge instances with NVMe SSD storage) running over 10-minute steady-state intervals under three synthetic scenarios:
- **Scenario A (ERC-3643 Transfers):** High-concurrency fractional security token transfers between 5,000 verified accounts. Besu sustained 2,620 TPS with average CPU utilization of 2.6 cores.
- **Scenario B (Batch DvP Settlement):** Atomic Delivery-versus-Payment execution (`SettlementDvP.sol`) bundling 50 bilateral trades per transaction. Besu achieved 2,415.6 average TPS with p95 latency of 1,180 ms and zero transaction failures. Fabric reached 1,807 TPS but encountered endorsement contention conflicts on high-velocity asset keys (115,800 failed transactions). Polygon CDK reached 2,104.3 TPS with higher p99 latency (2,310 ms).
- **Scenario C (Proof-of-Reserve Updates):** 32-level sparse Merkle root publishing with state verification. Besu processed state transitions in 412 ms median latency.

The detailed benchmark metrics are recorded in `benchmarks/blockchain/results/benchmark_results.json`.

---

## 5. Architectural Decision

Growww adopts **Hyperledger Besu with QBFT Consensus** as the foundational permissioned distributed ledger for securities settlement, tokenization, and proof-of-reserve attestation.

### Key Architectural Specifications:
- **Consensus & Block Production:** QBFT consensus algorithm configured with 2.0-second block period, 4.0-second request timeout, and 30,000 block epoch length.
- **Gas Economics:** Zero gas price (`zeroBaseFee: true`, `fixedBaseFee: 0`) within the permissioned consortium network to eliminate gas market volatility while preserving EVM gas metering to prevent contract denial-of-service loops.
- **Storage Engine:** Bonsai Tries key-value format for RocksDB, providing rolling state pruning and bounded disk footprint.
- **Consortium Network Topology:**
  - Node 1: Growww Primary Settlement Node (Mumbai AWS Region).
  - Node 2: Regulated Depository/Custodian Node (GIFT City Equinix DC).
  - Node 3: Clearing Corporation Node (Mumbai Primary Data Center).
  - Node 4: Regulatory Trust Node (Delhi Trust Facility).
  - Observer Nodes: Non-validating, read-only JSON-RPC nodes provisioned for SEBI and RBI audit teams.
- **Permissioning Architecture:**
  - On-chain node permissioning rules governed by smart contract at `0x0000000000000000000000000000000000008888`.
  - On-chain account permissioning rules governed by smart contract at `0x0000000000000000000000000000000000009999`.
- **Key Custody:**
  - Validator keys managed via Besu Web3Signer backed by AWS CloudHSM (FIPS 140-2 Level 3).

---

## 6. Security, Compliance, & Risk Controls

1. **Byzantine Fault Tolerance ($3F + 1$):** With $N=4$ validators, the consortium tolerates $F=1$ offline or compromised node, requiring a quorum of $2F + 1 = 3$ signatures to produce and finalize every block.
2. **Zero PII Ledger Invariant:** All customer identity verification is anchored through the `IIdentityRegistry` contract storing only `keccak256(abi.encode(panHash, kycTier, salt))` and numeric ISO country codes.
3. **Smart Contract Security Posture:** Contracts adhere to Solidity 0.8.24 with custom errors, OpenZeppelin upgradeable contracts, ReentrancyGuard, and automated Slither static analysis.
4. **Disaster Recovery:** Automated RocksDB hourly snapshot replication to secondary cold storage in Hyderabad (RPO < 1 hour, RTO < 15 minutes).

---

## 7. Stakeholder Sign-Off

- **Blockchain Architecture:** Approved
- **Enterprise Security:** Approved
- **Regulatory Compliance:** Approved
- **Infrastructure Operations:** Approved
