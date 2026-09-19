# 301 - Permissioned Blockchain Platform Evaluation & Selection (Hyperledger Besu vs Fabric vs Polygon Supernets)

## Purpose
The core value proposition of Growww is providing an immutable, cryptographically verifiable, and auditable settlement and proof-of-reserve layer for fractionalized Indian equities backed 1:1 by real securities held in SEBI-regulated depositories (NSDL/CDSL). Choosing the appropriate permissioned distributed ledger platform is a foundational architectural decision. The ledger must deliver high deterministic throughput (>=2,000 TPS), instant finality with zero fork probability, strict regulatory compliance (zero public PII, deterministic smart contract execution, node-level permissioning), low operational latency, and compatibility with enterprise Hardware Security Modules (HSMs).

This prompt directs the architecture team to produce a comprehensive benchmark, technical trade-off evaluation, and formal Architectural Decision Record (ADR) analyzing three leading permissioned distributed ledger technologies: **Hyperledger Besu (Enterprise Ethereum with QBFT consensus)**, **Hyperledger Fabric (v2.5+ with Raft consensus)**, and **Polygon Supernets / CDK (Enterprise EVM App-Chain)**.

## What You Are Building
A rigorous benchmark evaluation suite, empirical performance test report, and a formal Architectural Decision Record (`docs/architecture/adr_001_blockchain_platform_selection.md`) that documents:
- Comparative matrix analyzing throughput, finality latency, EVM tooling compatibility, HSM integration, privacy controls, and operational overhead.
- Automated benchmarking harness (using Hyperledger Caliper / custom k6 load generators) testing ERC-3643 transfer, batch DvP settlement, and proof-of-reserve Merkle root updates under sustained 2,000+ TPS.
- Formal selection justification confirming **Hyperledger Besu with QBFT consensus** as the primary consortium ledger, outlining network topology, node permissioning smart contracts, and production trade-offs.

## Scope Boundaries
- **In Scope:**
 - Quantitative and qualitative evaluation of Hyperledger Besu, Hyperledger Fabric, and Polygon Supernets/CDK.
 - Evaluation of consensus algorithms: QBFT (Quorum Byzantine Fault Tolerance) vs IBFT 2.0 vs Raft vs Polygon PoA/PoS.
 - Evaluation of smart contract ecosystems: Solidity EVM (ERC-3643, SettlementDvP.sol) vs Fabric Chaincode (Go/Rust).
 - Production benchmarking harness producing reproducible TPS, p95/p99 latency, and CPU/memory utilization figures.
 - Architectural Decision Record (ADR) detailing selection rationale, regulatory compliance alignment, and risk mitigations.
- **Out of Scope / Handled Elsewhere:**
 - Production validator node cluster deployment (handled in Prompt 302).
 - Smart contract implementations (handled in Prompts 303-307).
 - Validator HSM setup and key signing proxy (handled in Prompt 311).

## Technology to Use
- **Primary Selected Platform:** **Hyperledger Besu (v24.x+) running QBFT Consensus**.
  *Justification:* Hyperledger Besu is an enterprise-grade, Apache 2.0-licensed Ethereum client designed specifically for consortium environments. Its QBFT consensus mechanism provides immediate 1-block deterministic finality with zero risk of chain reorganizations (vital for legal securities settlement), full 100% EVM bytecode compatibility (allowing the team to leverage battle-tested Solidity standards like ERC-3643 and OpenZeppelin contracts), native on-chain node and account permissioning rules, native Web3Signer integration with cloud/on-premise HSMs (FIPS 140-2 Level 3), and clean auditability for regulatory observers (SEBI/RBI/IFSCA).
- **Comparative Baseline Platforms:** Hyperledger Fabric v2.5 (Channel/Orderer architecture) and Polygon CDK / Edge (App-chain EVM).
- **Benchmarking Tools:** Hyperledger Caliper, Prometheus, Grafana, Docker Compose, Python / Go benchmark client scripts.
- **EVM Tooling:** Foundry (forge/cast) for Solidity smart contract compilation, test suites, and gas profiling.

## Backend / Infra Touchpoints
- **Kubernetes / Cloud Infrastructure:** AWS EKS / GCP GKE clusters hosting benchmarking test harnesses.
- **Key Management:** HashiCorp Vault Transit Engine / AWS CloudHSM mock endpoints for transaction signing benchmarks.
- **Observability:** Prometheus pulling Besu `/metrics` endpoint, Grafana displaying throughput and block production latency dashboards.
- **Relational Stores:** PostgreSQL database ingestion sink to verify downstream event indexing throughput.

## Blockchain Interaction
This component establishes the foundation for all blockchain interactions across the Growww platform:
- **Consensus & Block Production:** Hyperledger Besu network configured with QBFT consensus, 2-second block intervals, and 0 gas fee (or fixed gas price) consortium token economics.
- **Custody Backing & Zero PII:** On-chain state stores strictly anonymous cryptographic investor identifiers (e.g. `bytes32 identityHash`, ERC-734/ERC-735 claim IDs) and fractional token balances backed 1:1 by NSDL/CDSL depository receipts. No customer names, PAN, Aadhaar, or bank details ever touch the ledger.
- **Smart Contract Target Ecosystem:** Standard Solidity EVM contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`, `MultiSigGovernance.sol`).
- **Institutional Node Participation:** Consensus topology supporting multi-institution validator participation (Growww Domestic Entity, Regulated Custodian, Clearing Partner, and read-only SEBI/IFSCA regulatory audit nodes).

## Step-by-Step Build Instructions
1. Scaffold the architecture evaluation repository structure under `benchmarks/blockchain/` with subdirectories for `besu-qbft/`, `fabric-raft/`, and `polygon-cdk/`.
2. Configure a 4-node Hyperledger Besu QBFT local cluster using Docker Compose with 2-second block times, zero gas price, and on-chain permissioning enabled.
3. Configure a 4-peer, 3-orderer Hyperledger Fabric v2.5 network with Raft consensus and a single channel representing the equity settlement domain.
4. Configure a 4-node Polygon CDK / Edge local testnet with PoA consensus and EVM compatibility.
5. Implement a standardized synthetic benchmark test suite:
 - Scenario A: High-concurrency ERC-20 / ERC-3643 fractional token transfers (1,000 to 5,000 concurrent accounts).
 - Scenario B: Atomic Delivery-versus-Payment (`SettlementDvP.sol`) batch settlement executions (100 trades per batch).
 - Scenario C: Proof-of-Reserve 32-level Merkle root publishing with state verification.
6. Instrument all node instances with Prometheus metric exporters capturing CPU, memory, disk I/O (RocksDB/LevelDB), network bandwidth, block propagation delay, and TPS.
7. Execute automated Caliper load testing ramping from 500 TPS to 5,000 TPS across 10-minute steady-state intervals for all three platforms.
8. Measure and document transaction finality: verify 100% immediate finality in QBFT vs probabilistic/block confirmation delays in PoA vs commit endorsement delays in Fabric.
9. Evaluate HSM signer integration complexity by benchmarking Web3Signer (Besu) vs Fabric CA / HSM client vs Polygon RPC signers.
10. Evaluate developer velocity and ecosystem tooling: contract testing in Foundry vs Fabric chaincode testing in Go/Docker.
11. Compile quantitative benchmark results into structured Markdown tables and Grafana dashboard snapshots.
12. Draft the formal Architectural Decision Record (`adr_001_blockchain_platform_selection.md`) presenting criteria scores, trade-offs, security properties, regulatory defensibility, and concluding selection of Hyperledger Besu (QBFT).
13. Submit ADR for review and institutional sign-off by Security, DevOps, and Compliance teams.

## Interfaces / Contracts

### Architectural Decision Record (ADR) Metadata Schema
```markdown
# ADR 001: Permissioned Consortium Ledger Selection

## Status
Accepted

## Context
Growww requires a deterministic, compliant, high-throughput permissioned distributed ledger to record fractional equity ownership, execute atomic Delivery-versus-Payment (DvP) settlements, and publish cryptographic proof-of-reserve attestations.

## Evaluated Options
1. Hyperledger Besu (v24.x, QBFT Consensus)
2. Hyperledger Fabric (v2.5, Raft Consensus)
3. Polygon Supernets / CDK (EVM App-Chain)

## Comparison Matrix
| Dimension | Hyperledger Besu (QBFT) | Hyperledger Fabric (Raft) | Polygon Supernets/CDK |
|---|---|---|---|
| Consensus & Finality | QBFT (Instant, 1-block, zero-fork) | Raft (Instant, leader-based) | PoA / IBFT (Near-instant) |
| EVM Compatibility | 100% Native EVM (Solidity 0.8.x) | None (WASM / Go Chaincode) | 100% Native EVM |
| Throughput (Settlement DvP) | ~2,400 TPS (2s blocks, batching) | ~1,800 TPS (Endorsement bottleneck)| ~2,100 TPS |
| Standard Token Support | ERC-3643 (T-REX), ERC-20, ERC-721 | Custom Chaincode models | ERC-3643, ERC-20 |
| HSM / Key Custody Integration | Native Web3Signer (CloudHSM/Vault) | PKCS#11 fabric-ca client | Custom RPC Relayers |
| Node Permissioning | Smart contract-based (On-Chain) | Fabric MSP / CA certificates | Allowlist contract |
| Regulatory Audit Readability | Standard Ethereum JSON-RPC / Explorers | Complex channel client required | Standard Ethereum JSON-RPC |

## Decision
Adopt **Hyperledger Besu with QBFT Consensus** as the enterprise permissioned ledger for Growww.
```

### Benchmark Metric JSON Output Schema
```json
{
  "test_id": "BENCH-BESU-QBFT-001",
  "platform": "Hyperledger Besu v24.1.0",
  "consensus": "QBFT",
  "block_period_seconds": 2,
  "nodes_count": 4,
  "workload": "BatchDvPSettlement_50TradesPerTx",
  "duration_seconds": 600,
  "metrics": {
    "total_transactions_submitted": 1200000,
    "successful_transactions": 1200000,
    "failed_transactions": 0,
    "average_tps": 2415.6,
    "peak_tps": 2850.2,
    "latency_p50_ms": 412,
    "latency_p95_ms": 1180,
    "latency_p99_ms": 1850,
    "forks_detected": 0,
    "reorgs_detected": 0
  },
  "resource_utilization": {
    "validator_avg_cpu_cores": 2.8,
    "validator_avg_memory_gb": 4.6,
    "rocksdb_disk_growth_mb_per_hour": 1420
  }
}
```

## Security & Compliance Notes
- **Consensus Liveness & Fault Tolerance:** QBFT requires $3F + 1$ validators to tolerate $F$ malicious or offline nodes. For a 4-node consortium network, 1 node failure is tolerated ($N=4, F=1$ with quorum $2F+1=3$).
- **Zero PII on Ledger:** All contract state must be rigorously designed to accept only hashed identity proofs or abstract compliance claim IDs. Under no circumstances should Aadhaar, PAN, phone numbers, or investor email addresses be stored on-chain.
- **Deterministic Smart Contract Execution:** Smart contracts must avoid non-deterministic EVM opcodes (e.g. relying on block hashes for entropy) and enforce strict gas limits even in zero-gas consortium networks to prevent infinite loop denial-of-service.
- **Regulatory Transparency:** Node configuration must support permissioned, read-only "Observer Nodes" that allow SEBI and RBI auditors to sync the full block ledger in real-time without holding block-signing or consensus voting keys.

## Acceptance Criteria
- [ ] Benchmarking repository with automated Docker Compose harnesses for Hyperledger Besu (QBFT), Hyperledger Fabric, and Polygon Supernets is functional and documented.
- [ ] Reproducible benchmark test runs execute at >=2,000 TPS under simulated high-load DvP settlement conditions.
- [ ] Comprehensive Architectural Decision Record (`adr_001_blockchain_platform_selection.md`) is completed, presenting empirical data, consensus trade-offs, and clear rationale for selecting Hyperledger Besu.
- [ ] Slither / Mythril security profiles for EVM contracts pass without high/critical issues on Besu target environment.
- [ ] Formal sign-off on the ADR achieved from Architecture, Security, and Compliance stakeholders.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `000` (Project North Star), Prompt `003` (Regulatory Pathway Overview), Prompt `101` (System Architecture Overview).
- **Parallel Tasks:** Prompt `302` (Network Topology & Validator Setup), Prompt `109` (Secrets Management Architecture).
- **Subsequent Prompts Enabled:** Prompts `303` through `314` (All smart contracts, services, and operations in Category 3).
