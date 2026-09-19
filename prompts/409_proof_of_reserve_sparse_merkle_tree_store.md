# 409 - Proof-of-Reserve Sparse Merkle Sum Tree State Store & ZK Generator (services/por-smt-generator)

## Purpose
In a fractional investment and digital asset infrastructure trading real asset-backed Indian equities (NSDL/CDSL demat holdings), physical precious commodities (WDRA-accredited gold and silver vault reserves), and multichain collateral assets (Bitcoin, Ethereum, Solana, USDC), cryptographic Proof of Reserve (PoR) is the cornerstone of institutional trust and solvency verification. Without verifiable, non-custodial mathematical solvency attestations, platforms remain vulnerable to fractional-reserve balance manipulation, undisclosed liabilities, and internal accounting discrepancies.

This prompt defines the architecture, storage layouts, cryptographic algorithms, Zero-Knowledge (ZK) circuit specifications, and API contracts for the **Proof-of-Reserve Sparse Merkle Sum Tree State Store & ZK Generator** (`services/por-smt-generator`). The service builds daily deterministic Sparse Merkle Sum Trees (SMST) over millions of investor accounts, reconciles aggregated liabilities against physical and on-chain custodial reserves, produces succinct Zero-Knowledge proofs (ZK-SNARKs) certifying platform solvency without leaking user balances or identities, and submits immutable epoch commitments to `MultiChainProofOfReserve.sol` on Hyperledger Besu.

## What You Are Building
A production-grade, high-performance cryptographic backend and storage microservice located in `services/por-smt-generator/` comprising:
- **Sparse Merkle Sum Tree (SMST) Engine:** High-performance, multi-threaded tree computation pipeline in Rust building deterministic depth-256 Sparse Merkle Sum Trees over user liabilities with algebraic hashing (Poseidon) and EVM hashing (Keccak-256/SHA-256).
- **LSM-Tree State Store (RocksDB):** Embedded, persistent key-value store optimized for massive sparse Merkle trees, managing copy-on-write snapshot versioning across daily epochs and sub-millisecond node retrieval.
- **ZK-SNARK Solvency Prover:** Proving backend utilizing Groth16 / BN254 circuits to generate non-interactive zero-knowledge proofs certifying that all account balances are non-negative ($\ge 0$) and sum up exactly to the total platform liability root.
- **Multichain & Depository Custody Reconciler:** Automated ingestion worker fetching and validating verified holdings from NSDL/CDSL depository files, WDRA electronic Negotiable Warehouse Receipts (eNWRs from NERL/CCRL), Bitcoin UTXOs, EVM lockbox contracts, and Solana escrow accounts.
- **On-Chain Notarization Relayer:** Secure transaction dispatcher integrating with AWS CloudHSM to submit signed daily epoch root commitments, asset breakdown summaries, and M-of-N custodian threshold signatures to `MultiChainProofOfReserve.sol`.
- **User Inclusion Proof API:** High-throughput gRPC and REST server allowing users, auditors, and mobile/web clients to retrieve $\mathcal{O}(\log N)$ cryptographic inclusion proofs to independently verify that their balances are included in the published on-chain root.

## Scope Boundaries
- **In Scope:**
  - Sparse Merkle Sum Tree (SMST) mathematical specification, parent sum aggregation, and node hashing.
  - RocksDB key-value schema, column families, epoch snapshot management, and prune/compaction policies.
  - Zero-Knowledge circuit specifications for balance non-negativity and total liability sum verification.
  - Depository reconciliation ingestion (NSDL/CDSL demat equities, WDRA eNWR commodities, BTC, ETH, SOL, USDC).
  - PostgreSQL `por_store` schema for epoch metadata, custodian signature registries, and dispute logs.
  - gRPC and REST interface definitions for proof serving and regulatory audit queries.
  - Automated mathematical solvency assertion and notarization payload assembly for `MultiChainProofOfReserve.sol`.
- **Out of Scope / Handled Elsewhere:**
  - On-chain registry contract implementation and EVM execution (handled in Prompt 327: `MultiChainProofOfReserve.sol`).
  - IoT calibrated scale weighment and assay inspector portal (handled in Prompt 716).
  - Off-chain in-memory matching engine execution (handled in Prompt 205).
  - Double-entry relational fiat ledger management (handled in Prompt 203 and Prompt 401).
  - Primary fiat payment gateway processing (handled in Prompt 212).

## Technology to Use
- **Core Service Language:** **Rust 1.78+** (with `tokio` asynchronous runtime, `tonic` gRPC framework, `rocksdb` bindings, and `alloy` / `ethers-rs` for Ethereum RPC).
  *Justification:* Provides deterministic memory safety, zero garbage collection pauses, SIMD-accelerated cryptographic hashing, and native interoperability with ZK-SNARK proving libraries.
- **Storage Engine:** **Embedded RocksDB 9.x** with custom Block-Based Table options, Prefix Bloom Filters, LZ4/ZSTD compression, and multi-threaded LSM-tree compaction.
- **Cryptographic Hash Functions:**
  - **Poseidon Hash (BN254):** Algebraic hash function optimized for arithmetic circuits, minimizing R1CS constraints in ZK solvency proofs.
  - **Keccak-256 & SHA-256:** Standard cryptographic hashes used for EVM smart contract interoperability and depository data manifests.
- **Zero-Knowledge Prover Framework:** **Groth16 on BN254 Curve** (implemented via `arkworks` / `gnark` / `circom-witnesscalc`) for fast proof generation and cheap on-chain gas verification ($< 250,000$ gas).
- **Relational Metadata Database:** **PostgreSQL 16** (`por_store` schema) for attestation histories, custodian digital signatures, and audit logs.
- **Inter-Service Protocol:** **gRPC / Protocol Buffers v3** for internal service queries; **REST / JSON** for public and auditor verification endpoints.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Database:** Stores epoch history, custodian signature manifests, asset category configurations, and dispute records.
- **Dedicated NVMe SSD Storage:** Dedicated high-IOPS persistent storage volume for RocksDB holding SMST node states across consecutive epochs.
- **Apache Kafka Cluster:** Ingests daily user account balance snapshots from topic `ledger.user_balances_snapshot` and custodian telemetry from `custody.attestations`.
- **AWS CloudHSM / HashiCorp Vault Transit:** Hardware-protected signing key for the PoR relayer account submitting transactions to Hyperledger Besu.
- **Hyperledger Besu Consortium Ledger:** Network hosting `MultiChainProofOfReserve.sol` under QBFT consensus with 2-second block intervals.
- **Custodial Node Interfaces:**
  - NSDL & CDSL SFTP/API connectors for Demat holding files.
  - NERL & CCRL repository gateways for WDRA commodity eNWRs.
  - Bitcoin Core full node (RPC/ZMQ) for Bitcoin cold wallet UTXO verification.
  - EVM RPC nodes for Ethereum/Polygon/Arbitrum USDC and lockbox contracts.
  - Solana RPC cluster for SPL token account balances.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Block Finality:** Permissioned Hyperledger Besu network running QBFT consensus with 2-second block finality, guaranteeing immediate non-revertible execution of reserve notarizations.
- **Daily Epoch Commitment:** Submits the root commitment tuple `(epochId, smstMerkleRoot, totalLiabilitySum, assetReserves, custodianSignatures)` to `MultiChainProofOfReserve.sol`.
- **On-Chain Solvency Verification:** The smart contract programmatically validates that for every registered asset category, total custodial reserves are greater than or equal to on-chain liabilities:
  $$\text{CustodialReserves}_{asset} \ge \text{OnChainLiabilities}_{asset}$$
- **Zero On-Chain PII Guarantee:** No user identities, PANs, account numbers, or individual balances are written to the blockchain. All leaf commitments are pseudonymous SHA-256/Poseidon hashes salted with cryptographically secure user blinds.
- **Self-Sovereign Public Verification:** Any investor can take their balance, blinding salt, and the inclusion proof path returned by the gRPC API, and call the view function `IMultiChainProofOfReserve.verifyUserInclusionProof()` to cryptographically verify their inclusion on-chain.

## Storage Architecture & Cryptographic Mechanics

### 1. Sparse Merkle Sum Tree (SMST) Specification
A Sparse Merkle Sum Tree is a deterministic full binary tree of fixed depth $D = 256$ representing a 256-bit key space ($0 \text{ to } 2^{256}-1$). Unlike a standard Sparse Merkle Tree where nodes store only cryptographic hashes, each SMST node stores both a **cryptographic hash commitment** and an **aggregated balance sum vector** across all supported asset categories.

#### Leaf Node Definition:
For a user account with index $K_i = \text{Poseidon}(\text{AccountId} \parallel \text{Salt}_i)$, the leaf node at level $0$ stores:
- $\text{LeafHash}_i = \text{Poseidon}(K_i \parallel B_{i,1} \parallel B_{i,2} \parallel \dots \parallel B_{i,m} \parallel \text{Salt}_i)$
- $\text{LeafSums}_i = [B_{i,1}, B_{i,2}, \dots, B_{i,m}]$, where $B_{i,k} \ge 0$ is the user balance for asset category $k$.

#### Non-Leaf / Internal Node Definition:
For an internal node at depth $d$ with left child $L = (H_L, \vec{S}_L)$ and right child $R = (H_R, \vec{S}_R)$:
- $\vec{S}_{parent} = \vec{S}_L + \vec{S}_R = [S_{L,1} + S_{R,1}, S_{L,2} + S_{R,2}, \dots, S_{L,m} + S_{R,m}]$
- $H_{parent} = \text{Poseidon}(H_L \parallel \vec{S}_L \parallel H_R \parallel \vec{S}_R)$

#### Zero Subtree Optimization:
Since the tree has $2^{256}$ potential leaves and most subtrees are completely empty (all balances zero), the store pre-computes default zero hashes and zero sum vectors for all levels $d \in [0, 256]$:
- $Z_0 = (\text{Poseidon}(0 \parallel \vec{0} \parallel 0), \vec{0})$
- $Z_{d+1} = (\text{Poseidon}(Z_d.H \parallel \vec{0} \parallel Z_d.H \parallel \vec{0}), \vec{0})$
Empty intermediate nodes are not written to RocksDB, reducing physical disk storage to $\mathcal{O}(N \log D)$ for $N$ active accounts.

```
                  [Root Node: (H_root, TotalSum)]
                            /           \
                           /             \
                  [Node: (H_L, Sum_L)]   [Node: (H_R, Sum_R)]
                        /       \              /       \
                     ...         ...        ...         ...
                     /             \        /             \
             [Leaf_1]        [Leaf_2]   [Leaf_3]      [Default Zero Z_0]
         (H_1, Sum_1)    (H_2, Sum_2)  (H_3, Sum_3)    (0x00, [0...0])
```

### 2. Zero-Knowledge Solvency Circuit (Groth16 / BN254)
The ZK-SNARK circuit proves that:
1. Every included leaf has non-negative balances: $\forall k \in [1, m], B_{i,k} \ge 0$ (preventing negative balance injection to hide deficits).
2. The Merkle sum propagation is mathematically valid at every branch: $\vec{S}_{parent} = \vec{S}_L + \vec{S}_R$.
3. The computed root hash and root total sum match the public signals $(H_{root}, \vec{S}_{total})$ committed to the blockchain.
4. User identities and balance distributions remain completely private.

### 3. RocksDB Key-Value Namespace Layout
RocksDB uses dedicated column families to isolate high-throughput node mutations from historical epoch metadata:
- `cf_default`: System configuration, current active epoch pointers.
- `cf_nodes`: SMST internal nodes indexed as `epoch:{epoch_id}:node:{depth}:{node_index_hex}` $\rightarrow$ `(Hash: [32]byte, Sums: [m]uint256)`.
- `cf_leaves`: Leaf nodes indexed as `epoch:{epoch_id}:leaf:{account_hash_hex}` $\rightarrow$ `(Salt: [32]byte, Sums: [m]uint256, LeafHash: [32]byte)`.
- `cf_history`: Compacted epoch root index `epoch:{epoch_id}:summary` $\rightarrow$ Protobuf serialized epoch descriptor.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure:** Initialize `services/por-smt-generator/` containing `src/`, `proto/`, `circuits/`, `migrations/`, `config/`, and `tests/`.
2. **Configure Embedded RocksDB Instance:** Implement storage initialization in `src/storage/db.rs` with custom column families (`cf_nodes`, `cf_leaves`, `cf_history`), LZ4 compression, block cache size of 8GB, and write buffer size of 512MB.
3. **Implement Precomputed Zero-Subtree Cache:** Build static tables in `src/tree/zero_cache.rs` computing default zero hashes $Z_0 \dots Z_{256}$ using the Poseidon BN254 cryptographic primitive.
4. **Implement Multi-Threaded SMST Builder:** Build the core tree construction pipeline in `src/tree/builder.rs` utilizing Rayon/Tokio thread pools for parallel bottom-up leaf insertion, sorting, and internal node hashing.
5. **Implement RocksDB Batch Persistence:** Implement atomic write batching in `src/storage/writer.rs` persisting newly calculated non-zero nodes and leaf records to RocksDB within an isolated epoch transaction.
6. **Implement Depository Holdings Ingestion Worker:** Author `src/ingest/demat.rs` to ingest daily NSDL/CDSL holding statements, parse ISO 6166 ISIN quantities, and compute depository asset reserve totals.
7. **Implement WDRA Commodity Vault Ingestion Worker:** Author `src/ingest/commodity.rs` to fetch electronic Negotiable Warehouse Receipts (eNWRs) from NERL/CCRL repositories, validate physical bar serial lists, and aggregate precious metal weights.
8. **Implement Multichain Custody Listeners:** Author `src/ingest/multichain.rs` to query Bitcoin Core RPC (UTXO balances), EVM JSON-RPC (lockbox smart contract balances), and Solana RPC (SPL token vaults).
9. **Define Zero-Knowledge Solvency Circuit:** Create Circom / Gnark circuit specifications in `circuits/solvency_sum_tree.circom` asserting leaf non-negativity and parent sum conservation.
10. **Implement ZK-SNARK Prover Worker:** Author `src/zk/prover.rs` using `arkworks`/`gnark` to compile witness inputs from RocksDB leaf states and generate Groth16 proofs.
11. **Implement gRPC Inclusion Proof Service:** Define `proto/por_service.proto` and build `src/server/grpc.rs` delivering user leaf inclusion proofs ($\mathcal{O}(\log N)$ sibling nodes) with sub-millisecond query latency.
12. **Implement REST Public Verification Server:** Author `src/server/rest.rs` exposing public endpoints for regulatory inspectors and clients to query epoch summaries, root hashes, and asset reserve allocations.
13. **Implement Hyperledger Besu Notarization Relayer:** Author `src/relayer/besu.rs` wrapping `alloy`/`ethers-rs` to format notarization transactions, sign payloads using AWS CloudHSM, and call `MultiChainProofOfReserve.notarizeEpochRoot()`.
14. **Implement Automated Solvency Alerting & Circuit Breaker:** Build `src/reconciliation/monitor.rs` verifying reserves against liabilities; trigger Prometheus alerts and notify `MultiChainProofOfReserve.triggerReserveDispute()` if any deficit is found.
15. **Write Comprehensive Test Suites:** Author unit tests, property-based fuzz tests for tree math, and end-to-end integration tests verifying 1,000,000 accounts processed in under 30 seconds.

## Interfaces / Contracts

### 1. Protocol Buffers Definition (`por_service.proto`)
```protobuf
syntax = "proto3";

package growww.por.v1;

enum AssetCategory {
  ASSET_CATEGORY_UNSPECIFIED = 0;
  ASSET_CATEGORY_DEMAT_EQUITY = 1;
  ASSET_CATEGORY_BITCOIN_COLD_WALLET = 2;
  ASSET_CATEGORY_EVM_LOCKBOX = 3;
  ASSET_CATEGORY_SOLANA_LOCKBOX = 4;
  ASSET_CATEGORY_WDRA_VAULT_COMMODITY = 5;
  ASSET_CATEGORY_BANK_FIAT_ESCROW = 6;
}

message AssetBalance {
  string asset_id = 1;
  AssetCategory category = 2;
  string balance_units = 3; // String representation of high-precision uint256
  int32 decimals = 4;
}

message UserInclusionProofRequest {
  uint64 epoch_id = 1;
  string account_id = 2;
  string user_blinding_salt = 3; // Hex-encoded 32-byte salt
}

message InclusionProofSiblingNode {
  string sibling_hash = 1; // Hex-encoded 32-byte hash
  repeated AssetBalance sibling_sums = 2;
  bool is_right_sibling = 3;
  uint32 depth = 4;
}

message UserInclusionProofResponse {
  uint64 epoch_id = 1;
  string account_hash = 2;
  string leaf_hash = 3;
  repeated AssetBalance declared_balances = 4;
  string root_hash = 5;
  repeated AssetBalance total_liabilities = 6;
  repeated InclusionProofSiblingNode proof_path = 7;
  int64 epoch_timestamp = 8;
  string on_chain_tx_hash = 9;
}

message VerifyInclusionProofRequest {
  uint64 epoch_id = 1;
  string leaf_hash = 2;
  repeated AssetBalance declared_balances = 3;
  repeated InclusionProofSiblingNode proof_path = 4;
  string expected_root_hash = 5;
}

message VerifyInclusionProofResponse {
  bool is_valid = 1;
  string computed_root_hash = 2;
  string verification_message = 3;
}

message EpochSummaryRequest {
  uint64 epoch_id = 1;
}

message AssetReserveSummary {
  string asset_id = 1;
  AssetCategory category = 2;
  string custodial_reserve_units = 3;
  string on_chain_liability_units = 4;
  bool is_solvent = 5;
  string custody_statement_hash = 6;
  int64 last_verified_timestamp = 7;
}

message EpochSummaryResponse {
  uint64 epoch_id = 1;
  string smst_merkle_root = 2;
  repeated AssetBalance total_liabilities = 3;
  repeated AssetReserveSummary asset_reserves = 4;
  bool overall_solvency_status = 5;
  string zk_proof_groth16_hex = 6;
  string besu_notarization_tx_hash = 7;
  int64 notarized_at_timestamp = 8;
}

service ProofOfReserveService {
  rpc GetUserInclusionProof (UserInclusionProofRequest) returns (UserInclusionProofResponse);
  rpc VerifyInclusionProof (VerifyInclusionProofRequest) returns (VerifyInclusionProofResponse);
  rpc GetEpochSummary (EpochSummaryRequest) returns (EpochSummaryResponse);
}
```

### 2. PostgreSQL DDL Schema (`por_store`)
```sql
CREATE SCHEMA IF NOT EXISTS por_store;

CREATE TABLE por_store.epochs (
    epoch_id BIGINT PRIMARY KEY,
    smst_merkle_root CHAR(66) NOT NULL, -- 0x + 64 hex characters
    total_accounts_indexed INT NOT NULL,
    zk_proof_data JSONB,
    zk_public_inputs JSONB,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'CALCULATING', 'PROVEN', 'NOTARIZED', 'DISPUTED')),
    besu_tx_hash CHAR(66),
    besu_block_number BIGINT,
    notarized_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE por_store.asset_reserves (
    reserve_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epoch_id BIGINT NOT NULL REFERENCES por_store.epochs(epoch_id) ON DELETE CASCADE,
    asset_id VARCHAR(64) NOT NULL,
    asset_category VARCHAR(32) NOT NULL CHECK (asset_category IN ('DEMAT_EQUITY', 'BITCOIN_COLD_WALLET', 'EVM_LOCKBOX', 'SOLANA_LOCKBOX', 'WDRA_VAULT_COMMODITY', 'BANK_FIAT_ESCROW')),
    custodial_reserve_units NUMERIC(38, 8) NOT NULL,
    liability_units NUMERIC(38, 8) NOT NULL,
    is_solvent BOOLEAN GENERATED ALWAYS AS (custodial_reserve_units >= liability_units) STORED,
    custody_statement_hash CHAR(66) NOT NULL,
    custody_source_uri TEXT NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_epoch_asset UNIQUE (epoch_id, asset_id)
);

CREATE TABLE por_store.custodian_attestations (
    attestation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epoch_id BIGINT NOT NULL REFERENCES por_store.epochs(epoch_id) ON DELETE CASCADE,
    custodian_name VARCHAR(128) NOT NULL,
    signer_address CHAR(42) NOT NULL,
    asset_category VARCHAR(32) NOT NULL,
    raw_statement_payload JSONB NOT NULL,
    statement_hash CHAR(66) NOT NULL,
    signature_bytes BYTEA NOT NULL,
    signed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE por_store.dispute_logs (
    dispute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epoch_id BIGINT NOT NULL REFERENCES por_store.epochs(epoch_id),
    asset_id VARCHAR(64) NOT NULL,
    custodial_units NUMERIC(38, 8) NOT NULL,
    liability_units NUMERIC(38, 8) NOT NULL,
    deficit_amount NUMERIC(38, 8) NOT NULL,
    alert_severity VARCHAR(16) NOT NULL DEFAULT 'CRITICAL',
    resolved BOOLEAN DEFAULT FALSE,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_asset_reserves_epoch ON por_store.asset_reserves(epoch_id);
CREATE INDEX idx_custodian_attestations_epoch ON por_store.custodian_attestations(epoch_id);
CREATE INDEX idx_epochs_status ON por_store.epochs(status);
```

### 3. ZK-SNARK Solvency Circuit Specification (Circom Interface Schema)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SolvencySumTreeCircuitSignals",
  "type": "object",
  "required": [
    "public_signals",
    "private_signals"
  ],
  "properties": {
    "public_signals": {
      "type": "object",
      "required": [
        "expectedRootHash",
        "totalLiabilitySum",
        "epochId"
      ],
      "properties": {
        "expectedRootHash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
        "totalLiabilitySum": { "type": "array", "items": { "type": "string" } },
        "epochId": { "type": "integer", "minimum": 1 }
      }
    },
    "private_signals": {
      "type": "object",
      "required": [
        "accountIds",
        "salts",
        "balances",
        "merklePathElements"
      ],
      "properties": {
        "accountIds": { "type": "array", "items": { "type": "string" } },
        "salts": { "type": "array", "items": { "type": "string" } },
        "balances": { "type": "array", "items": { "type": "array", "items": { "type": "string" } } },
        "merklePathElements": { "type": "array", "items": { "type": "array", "items": { "type": "string" } } }
      }
    }
  }
}
```

### 4. Solidity Interface Binding Reference (`IMultiChainProofOfReserve.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IMultiChainProofOfReserve {
    enum AssetCategory {
        DEMAT_EQUITY,
        BITCOIN_COLD_WALLET,
        EVM_LOCKBOX,
        SOLANA_LOCKBOX,
        WDRA_VAULT_COMMODITY,
        BANK_FIAT_ESCROW
    }

    struct AssetReserveEntry {
        string assetId;
        AssetCategory category;
        uint256 custodialReserveUnits;
        uint256 onChainLiabilityUnits;
        bytes32 custodyStatementHash;
        uint256 lastVerifiedTimestamp;
    }

    struct InclusionProofNode {
        bytes32 siblingHash;
        uint256 siblingSum;
        bool isRightSibling;
    }

    struct CustodianSignature {
        address signer;
        bytes signature;
    }

    function notarizeEpochRoot(
        uint256 epochId,
        bytes32 smstMerkleRoot,
        uint256 totalLiabilitySum,
        AssetReserveEntry[] calldata assetReserves,
        CustodianSignature[] calldata custodianSignatures,
        bytes calldata zkProof
    ) external;

    function verifyUserInclusionProof(
        uint256 epochId,
        bytes32 leafHash,
        uint256 leafSum,
        InclusionProofNode[] calldata proofPath
    ) external view returns (bool isValid);

    function triggerReserveDispute(
        uint256 epochId,
        string calldata assetId,
        string calldata reason
    ) external;
}
```

## Security & Compliance Notes
- **Mathematical Non-Negativity Invariant:** The core solvency model strictly enforces $B_{i,k} \ge 0$ for every user balance. Inserting negative balance entries into the SMST is mathematically rejected by both tree validation checks and the ZK-SNARK range constraint circuit ($2^{64}$ range bit checks), preventing malicious actors from artificially deflating aggregate liabilities.
- **Zero-Knowledge Privacy Compliance:** In strict adherence to India's Digital Personal Data Protection (DPDP) Act 2023 and global financial secrecy standards, no personally identifiable information (PII) or plaintext asset holdings are revealed. User leaves are pseudonymized using cryptographically secure blinding salts:
  $$K_i = \text{Poseidon}(\text{AccountId} \parallel \text{Salt}_i)$$
- **Cryptographic Blinding Factor Protection:** User blinding salts are derived through HMAC-SHA256 using a cluster master secret stored within AWS CloudHSM. Users receive their individual salt over an authenticated, TLS-encrypted session, allowing self-audit while preventing external attackers from enumerating account numbers.
- **Multi-Custodian M-of-N Threshold Signatures:** Depository and vault custody statements require digital signatures from authorized M-of-N custodian signers (NSDL, CDSL, NERL, Brink's, Sequel, BitGo/Copper). Unsigned or invalidly signed statements are rejected prior to epoch tree calculation.
- **SEBI 8-Year Audit Trail & Immutability:** All calculated SMST roots, attestation records, and RocksDB epoch snapshots are archived to WORM (Write Once, Read Many) S3 Object Lock storage with an 8-year compliance retention lock matching SEBI and PMLA statutory guidelines.
- **HSM-Protected Relayer Security:** Transactions dispatched to Hyperledger Besu are signed exclusively within an AWS CloudHSM cluster or HashiCorp Vault Transit backend, preventing relayer key exposure or unauthorized manual root overrides.

## Acceptance Criteria
- [ ] `services/por-smt-generator` successfully compiles under Rust 1.78+ with zero compiler warnings or lint errors.
- [ ] Zero application implementation code included in interface specifications (only data schemas, protobufs, DDLs, circuit definitions, and method signatures).
- [ ] Sparse Merkle Sum Tree engine constructs a 1,000,000 account tree with depth 256 in $< 30\text{ seconds}$ on an 8-core instance.
- [ ] Pre-computed zero-subtree cache properly initializes default empty hashes $Z_0 \dots Z_{256}$ for Poseidon BN254.
- [ ] RocksDB storage layer achieves $< 2\text{ms}$ retrieval latency for individual user inclusion proof paths across 256 depth levels.
- [ ] Groth16 / BN254 ZK-SNARK circuit proves balance non-negativity and total liability sum equality with proof generation time $< 5\text{ seconds}$.
- [ ] Depository reconciliation worker correctly parses NSDL/CDSL holding statements and matches ISIN units with 6 decimal precision.
- [ ] WDRA vault reconciliation worker ingests NERL/CCRL eNWR receipts and physical bar serials for `gGOLD` and `gSILVER`.
- [ ] Multichain custody worker accurately verifies UTXOs on Bitcoin Core, smart contract lockbox balances on EVM chains, and SPL balances on Solana.
- [ ] gRPC service implements `GetUserInclusionProof` and `VerifyInclusionProof` with p99 response time $< 10\text{ms}$ under 5,000 concurrent requests.
- [ ] Relayer successfully packages notarization payloads and commits verified epoch records to `MultiChainProofOfReserve.sol` on Hyperledger Besu.
- [ ] Automated dispute detection triggers immediate critical alerts and contract dispute calls if $\text{CustodialReserves} < \text{Liabilities}$.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (Microservice & API Design Standards), Prompt 104 (Kafka Event-Driven Architecture), Prompt 301 (Permissioned Blockchain Platform Evaluation), Prompt 302 (Network Topology & Validator Setup), Prompt 327 (`MultiChainProofOfReserve.sol`), Prompt 401 (PostgreSQL Schema Design).
- **Parallel Tasks:** Prompt 215 (Automated Reconciliation Engine), Prompt 218 (Immutable Audit Log Service), Prompt 716 (MCX & WDRA Physical Vault Audit Portal).
- **Subsequent Prompts Enabled:** Prompt 309 (Blockchain Event Indexer & State Sync), Prompt 525 (Flutter API Client Layer & Offline Sync), Prompt 530 (Flutter Commodity Physical Delivery & Vault Redemption Flow).
