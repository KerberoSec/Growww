# 720 - Zero-Knowledge Proof of Liabilities (ZK-PoL) Sparse Merkle Tree Engine

## Purpose
In modern electronic exchanges and multi-asset financial platforms, verifying platform solvency without compromising customer confidentiality is a fundamental operational, fiduciary, and regulatory requirement. Traditional financial audits rely on periodic, centralized sampling by external accounting firms, requiring the disclosure of sensitive investor Personal Identifiable Information (PII), raw account balances, and portfolio compositions. Such practices violate data sovereignty laws, notably India's Digital Personal Data Protection Act (DPDP Act 2023) and global privacy regulations such as GDPR. Furthermore, conventional Merkle tree implementations leak sibling balances to users during inclusion checks, enabling adversarial observers to reconstruct the entire exchange liability distribution through repeated audits.

Crucially, without cryptographic non-negativity guarantees, a compromised or insolvent exchange can fabricate fictitious accounts with negative balances (underflow attacks) to artificially depress its aggregate liabilities, mathematically masking stolen or lost customer funds.

The purpose of this specification is to define the architecture, cryptographic circuits, state storage, and verification pipeline for the **Zero-Knowledge Proof of Liabilities (ZK-PoL) Sparse Merkle Tree Engine** (`services/zk-pol-engine`). The engine provides immutable, mathematically verifiable proof that the exchange's total liabilities (the exact sum of all customer balances) strictly match or are covered by total depository custodial reserves held at NSDL, CDSL, WDRA-accredited vaults, and banking partners. Concurrently, it empowers every individual investor to cryptographically verify their specific balance inclusion in the published liability root with zero knowledge of other users' balances, identities, or aggregate exchange distribution curves.

## What You Are Building
A high-throughput, fault-tolerant cryptographic engine (`services/zk-pol-engine`) and zero-knowledge pipeline comprising:
- **Sparse Merkle Sum Tree (SMST) Synthesizer:** A high-performance Rust core constructing deterministic, depth-64 Sparse Merkle Sum Trees over millions of user accounts using the algebraic Poseidon hash function over the BN254 scalar field.
- **ZK Non-Negativity & Solvency Circuit Suite (Circom 2.1+):** Arithmetic R1CS circuits enforcing that every user balance is strictly non-negative ($0 \le b_i < 2^{64}$), that parent nodes represent the exact sum of child balances ($B_{parent} = B_{left} + B_{right}$), and that the aggregate liability commitment matches the tree root.
- **Parallelized Proving Pipeline (snarkjs / RapidSNARK / C++ Prover):** Multi-threaded witness generation and Groth16 zk-SNARK prover generating succinct, constant-size solvency proofs within minutes across massive account sets.
- **Embedded LSM-Tree State Store (RocksDB):** Persistent storage layer for tree nodes, leaves, intermediate subtree hashes, and epoch snapshot history with copy-on-write versioning.
- **Cryptographic Blinding & Salt Vault:** Hardware-protected, salted key derivation module ensuring irreversible account pseudonymization and preventing dictionary balance-enumeration attacks.
- **User Inclusion Proof API & Verification SDK:** High-throughput gRPC and REST endpoints serving logarithmic $\mathcal{O}(\log N)$ inclusion paths, paired with a client-side WebAssembly (Wasm) verification library for web and mobile clients.
- **On-Chain Notarization Relayer:** Automated dispatcher publishing epoch root commitments, liability sums, and zk-SNARK proofs to `ZKProofOfSolvencyVerifier.sol` on Hyperledger Besu.

## Scope Boundaries
- **In Scope:**
  - Design and implementation of depth-64 Sparse Merkle Sum Trees (SMST) in Rust.
  - Implementation of Poseidon algebraic hashing over BN254 curve scalar field ($\mathbb{F}_r$).
  - Circom R1CS circuits for 64-bit balance range checking ($0 \le b_i < 2^{64}$) and sum aggregation.
  - Batch witness generation and Groth16 proving pipeline orchestration using snarkjs and C++ backends.
  - Integration with RocksDB for sparse subtree caching and epoch snapshot retrieval.
  - gRPC and REST APIs for user leaf inclusion proof querying.
  - Protobuf and JSON schemas for inclusion proofs, epoch manifests, and solvency proof payloads.
  - On-chain proof dispatch and transaction execution on `ZKProofOfSolvencyVerifier.sol` on Hyperledger Besu.
  - DPDP Act 2023 zero-PII compliance and zero-knowledge privacy enforcement.
- **Out of Scope / Handled Elsewhere:**
  - Real-time double-entry ledger mutation and cash balance tracking (Prompt 203: `services/wallet-account-service`).
  - Custodial depository reserve ingestion from NSDL, CDSL, and banks (Prompt 213 & Prompt 308).
  - General Proof of Reserve smart contract state registry (Prompt 308 & Prompt 327: `MultiChainProofOfReserve.sol`).
  - Smart contract EVM pairing precompile verifier implementation (Prompt 317: `ZKProofOfSolvencyVerifier.sol`).
  - Hardware Security Module (HSM) PKCS#11 key lifecycle operations (Prompt 717: `services/cloudhsm-signer-daemon`).
  - Mobile user interface rendering and presentation (Prompt 514: `flutter_settings_profile_ui`).

## Technology to Use
- **Core Microservice Language:** **Rust 1.78+** utilizing `tokio` asynchronous runtime, `tonic` for gRPC, `rayon` for data-parallel tree synthesis, `arkworks-rs` (`ark-bn254`, `ark-ff`), and `rocksdb` bindings.
  *Justification:* Guarantees memory safety, eliminates garbage collection pauses, and provides SIMD-accelerated modular arithmetic essential for hashing millions of leaves under tight latency windows.
- **Zero-Knowledge Circuit Framework:** **Circom 2.1+** (R1CS arithmetic constraint compiler) and **snarkjs** / **RapidSNARK** (C++ / CUDA multi-threaded witness and Groth16 proving backends).
  *Justification:* Circom generates mathematically audited R1CS constraint systems; Groth16 produces tiny proofs (128 bytes) verifiable on Besu EVM for under 250,000 gas.
- **Cryptographic Hash Function:** **Poseidon Hash (BN254 curve)** with parameters $t=3$ for 2-to-1 internal nodes and $t=4$ for 3-input leaf nodes, using S-box $x^5$ over the BN254 scalar field $\mathbb{F}_r$ ($p = 21888242871839275222246405745257275088548364400416034343698204186575808495617$).
  *Justification:* Requires only ~240 R1CS constraints per hash compared to over 25,000 constraints for SHA-256, reducing proving time by two orders of magnitude.
- **Embedded State Store:** **RocksDB 9.x** with custom column families, block-based table options, prefix bloom filters, and LZ4 compression.
  *Justification:* Provides high random-read IOPS and compact disk footprint for deep sparse tree nodes.
- **Distributed Cache & Task Broker:** **Redis 7.2 Enterprise** for queuing proving jobs and caching inclusion proofs for active user sessions.
- **Messaging Infrastructure:** **Apache Kafka (librdkafka)** for consuming daily balance snapshot events and publishing solvency attestation manifests.
- **Transport Security:** **gRPC over HTTP/2** with Mutual TLS (mTLS 1.3), enforcing X.509 certificate pinning and AES-256-GCM.

## Backend / Infra Touchpoints
- **Upstream Services & Ingestion Sources:**
  - `services/wallet-account-service` (Prompt 203): Emits end-of-day liability balance snapshot events on Kafka topic `ledger.snapshot.liabilities.v1`.
  - `services/por-smt-generator` (Prompt 409): Shares RocksDB storage layouts, node serialization formats, and epoch checkpoint markers.
  - `services/custodian-depository-service` (Prompt 213): Emits verified asset reserve balances on `custody.attestation.completed.v1`.
- **Downstream Consumers & Notarization Targets:**
  - `ZKProofOfSolvencyVerifier.sol` (Prompt 317) deployed on Hyperledger Besu: Receives the Groth16 proof, Merkle liability root, and public inputs.
  - `services/on-chain-por-publisher` (Prompt 308): Ingests the finalized proof manifest and coordinates dual notarization of assets and liabilities.
  - `services/cloudhsm-signer-daemon` (Prompt 717): Signs on-chain notarization transactions with the operator key.
- **Kafka Topics:**
  - Ingests:
    - `ledger.snapshot.liabilities.v1`: User account balance snapshot array containing `(account_id, balance_inr_paisa, balance_securities_map, epoch_id)`.
    - `por.custody_reserves_published.v1`: Custodial depository holding commitments for correlation.
  - Emits:
    - `zkpol.smst_constructed.v1`: Emits tree build metadata, root hash, liability sum, and leaf count.
    - `zkpol.snark_proof_generated.v1`: Emits completed Groth16 proof, public signals, and proving duration.
    - `zkpol.attestation_notarized.v1`: Emits Besu transaction hash and on-chain verification confirmation.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Execution Environment:** Hyperledger Besu permissioned consortium network running Istanbul/QBFT BFT consensus with 2-second deterministic block finality.
- **On-Chain Verifier Smart Contract (`ZKProofOfSolvencyVerifier.sol` - Prompt 317):**
  - The ZK-PoL engine calls `verifyAndRecordSolvency(epochId, a, b, c, publicInputs)` via the Besu JSON-RPC gateway.
  - Parameters correspond to the Groth16 proof points $A \in \mathbb{G}_1, B \in \mathbb{G}_2, C \in \mathbb{G}_1$ and 5 public signals:
    1. `merkleSumTreeRoot`: 32-byte Poseidon root hash of the liabilities SMST.
    2. `totalLiabilities`: 256-bit integer representing aggregate customer liabilities (in smallest currency unit / paisa).
    3. `totalDepositoryReserves`: 256-bit integer representing verified reserves from depository attestations.
    4. `custodianSignatureHash`: SHA-256 digest of multi-signatory custodial reserve manifests.
    5. `epochTimestamp`: Unix epoch timestamp of the audited snapshot.
  - The contract executes BN254 elliptic curve pairing checks (`ecPairing` precompile at address `0x08`). If valid and `totalDepositoryReserves >= totalLiabilities`, the epoch is marked as certified and insolvent states revert atomically.
- **Zero-PII On-Chain Guarantee:**
  - Zero plaintext customer identifiers, PAN numbers, bank accounts, or individual balances are submitted to the ledger.
  - Leaf commitments are computed as $K_i = \text{Poseidon}(\text{AccountId} \parallel \text{UserBlindingSalt}_i)$. Salts are generated from a 256-bit secret derivation master key, ensuring cryptographic irreversibility.
- **Public On-Chain Inclusion Verification:**
  - The verifier contract exposes a view function:
    `verifyUserInclusion(uint256 epochId, bytes32 leafHash, uint256 leafBalance, bytes32[] calldata pathElements, uint8[] calldata pathIndices)`
    allowing any investor or third-party auditor to independently recompute the root on Besu without trusted intermediaries.

## Cryptographic Architecture & Mathematical Mechanics

### 1. Sparse Merkle Sum Tree (SMST) Formulation
A Sparse Merkle Sum Tree of depth $D = 64$ spans a virtual key space of $2^{64}$ leaves. Each node $N$ in the tree is a tuple consisting of a cryptographic hash $H \in \mathbb{F}_r$ and a non-negative scalar balance $B \in \mathbb{Z}_{\ge 0}$.

#### Leaf Node Construction:
For user account $i$ with account identifier $UID_i \in \mathbb{F}_r$, balance $b_i \in [0, 2^{64}-1]$, and private user blinding salt $s_i \in \mathbb{F}_r$:
1. User Path Key:
   $$K_i = \text{Poseidon}_2(UID_i, s_i)$$
   The 64-bit prefix of $K_i$ determines the leaf's unique deterministic path index from root to level 0.
2. Leaf Hash:
   $$H_{0, i} = \text{Poseidon}_3(K_i, b_i, s_i)$$
3. Leaf Balance:
   $$B_{0, i} = b_i$$

#### Default Empty Nodes (Zero Subtrees):
Because the tree is sparse ($2^{64}$ capacity vs millions of actual users), unpopulated subtrees collapse into precomputed default zero nodes:
- At height 0 (empty leaf): $B_0^{(default)} = 0$, $H_0^{(default)} = 0$.
- At height $h \in [1, D]$:
  $$B_h^{(default)} = 2 \cdot B_{h-1}^{(default)} = 0$$
  $$H_h^{(default)} = \text{Poseidon}_4(H_{h-1}^{(default)}, 0, H_{h-1}^{(default)}, 0)$$
This allows $\mathcal{O}(N \log N)$ tree computation and storage, where empty branches require no persistent RocksDB entries.

#### Non-Leaf / Internal Node Aggregation:
For an internal node at height $h$ with left child $(H_L, B_L)$ and right child $(H_R, B_R)$:
1. Sum Aggregation Invariant:
   $$B_{parent} = B_L + B_R$$
2. Parent Hash Commitment:
   $$H_{parent} = \text{Poseidon}_4(H_L, B_L, H_R, B_R)$$
At height $D = 64$, the root tuple is $(H_{root}, B_{root})$, where $B_{root} = \sum_{i=1}^{N} b_i$ constitutes the platform's total liability.

### 2. Negative Balance Prevention & Non-Negativity Circuit Mechanics
In a naive Merkle Sum Tree without zero-knowledge range constraints, a malicious operator can hide an asset shortfall of amount $X$ by inserting a fake account with negative balance $-X$. Since the sum property $B_{parent} = B_L + B_R$ holds algebraically in modular arithmetic, the root balance becomes:
$$B_{root} = (\sum b_{real}) - X$$
In the finite field $\mathbb{F}_r$ where $p \approx 2^{254}$, $-X$ wraps to $p - X$. Without range constraints, an auditor summing field elements cannot distinguish a legitimate sum from an underflowed modular sum.

To definitively eliminate this vulnerability, the ZK-PoL engine compiles a Circom circuit enforcing a strict 64-bit binary decomposition constraint on every leaf balance:
$$b_i = \sum_{j=0}^{63} 2^j \cdot \text{bit}_j \quad \text{where} \quad \text{bit}_j \cdot (1 - \text{bit}_j) = 0$$
Because $2^{64} - 1 \ll p$, modular wrap-around is mathematically impossible. A negative balance cannot satisfy the binary decomposition within 64 bits, causing the R1CS constraint check to fail during witness generation.

### 3. Circom Circuit Architecture
The verification circuit decomposes into modular templates:
- `RangeCheck64`: Enforces that an input scalar belongs strictly to $[0, 2^{64}-1]$ using 64 boolean constraints.
- `SMSTInternalNode`: Verifies that parent balance equals $B_L + B_R$ and parent hash matches $\text{Poseidon}_4(H_L, B_L, H_R, B_R)$.
- `SMSTSubtree(subdepth)`: Proves validity across a batch of $2^{\text{subdepth}}$ leaves in parallel.
- `SolvencyProof`: Aggregates the root liabilities sum and proves that:
  $$\text{TotalLiabilities} \le \text{TotalCustodialReserves}$$
  while verifying the top-level tree root $H_{root}$.

### 4. User Inclusion Verification Path
When a user requests proof of inclusion, the engine provides an inclusion path consisting of 64 sibling tuples:
$$\mathcal{P}_i = [ (H_0^{(sib)}, B_0^{(sib)}, \text{dir}_0), (H_1^{(sib)}, B_1^{(sib)}, \text{dir}_1), \dots, (H_{D-1}^{(sib)}, B_{D-1}^{(sib)}, \text{dir}_{D-1}) ]$$
where $\text{dir}_h \in \{0, 1\}$ indicates whether the sibling is on the left or right.

The client recalculates iteratively from $h = 0$ to $D-1$:
- If $\text{dir}_h = 0$ (sibling is left, current is right):
  $$B_{h+1} = B_h^{(sib)} + B_h, \quad H_{h+1} = \text{Poseidon}_4(H_h^{(sib)}, B_h^{(sib)}, H_h, B_h)$$
- If $\text{dir}_h = 1$ (sibling is right, current is left):
  $$B_{h+1} = B_h + B_h^{(sib)}, \quad H_{h+1} = \text{Poseidon}_4(H_h, B_h, H_h^{(sib)}, B_h^{(sib)})$$
The user asserts that $H_D = H_{root}$ and $B_D = B_{root}$.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Engine Repository and Module Layout:**
   Establish the project workspace under `services/zk-pol-engine/`, `circuits/pol/`, and `crates/smst-core/`. Configure Rust `Cargo.toml` with `arkworks`, `rocksdb`, `tokio`, `tonic`, `rayon`, and `circom-witnesscalc`. Configure Circom build pipelines with `snarkjs` and `rapidsnark`.
2. **Implement Poseidon Hash Primitives in Rust:**
   Develop `crates/smst-core/src/poseidon.rs` implementing round constants, MDS matrices, and S-box operations for $t=3$ and $t=4$ over the BN254 scalar field. Ensure exact bitwise and arithmetic parity with the Circom `Poseidon` circuit implementation.
3. **Build RocksDB Sparse Merkle Sum Tree Storage Engine:**
   Implement `crates/smst-core/src/storage.rs` wrapping RocksDB with column families `cf_nodes`, `cf_leaves`, `cf_epochs`, and `cf_zero_nodes`. Implement precomputed zero-subtree hash lookup tables for levels 0 through 64.
4. **Develop In-Memory Tree Synthesis Pipeline:**
   Implement `crates/smst-core/src/tree.rs` using `rayon` for parallel multi-threaded bottom-up construction. Ingest account records, sort by path keys $K_i$, insert active leaves, compute sibling sums, and bubble up Poseidon hashes to derive $H_{root}$ and $B_{root}$.
5. **Develop Circom Non-Negativity Range Check Circuit:**
   Create `circuits/pol/range_check.circom` with template `RangeCheck64()`. Implement 64-bit boolean decomposition and verify constraint completeness using the Ecne constraint analyzer to ensure zero underflow vectors.
6. **Develop Circom Sparse Merkle Sum Tree Subtree Circuit:**
   Create `circuits/pol/smst_subtree.circom` validating internal node summation $B_P = B_L + B_R$ and Poseidon hash transitions. Support parameterizable subtree depths (depth 4 to depth 16) for recursive or batched proving.
7. **Develop Top-Level Solvency Circuit:**
   Create `circuits/pol/solvency_proof.circom`. Accept public inputs `merkleSumTreeRoot`, `totalLiabilities`, `totalDepositoryReserves`, and `epochTimestamp`. Accept private inputs for subtree roots and custodial breakdown. Enforce invariant `totalLiabilities <= totalDepositoryReserves`.
8. **Execute Trusted Setup Ceremony Pipeline:**
   Download and verify Phase 1 Powers of Tau ceremony transcript (BN254, $2^{21}$ constraints). Perform Phase 2 circuit-specific setup using `snarkjs groth16 setup`, apply entropy contributions, and export `pol_final.zkey` and `verification_key.json`.
9. **Build High-Performance Witness & Prover Daemon:**
   Implement `services/zk-pol-engine/src/prover.rs` integrating RapidSNARK (C++/CUDA) and Rust witness calculation. Orchestrate CPU/GPU memory allocation to generate Groth16 proofs for 1,000,000 accounts in under 180 seconds.
10. **Implement User Blinding Salt Key Vault:**
    Develop `services/zk-pol-engine/src/vault.rs` interfacing with AWS CloudHSM / HashiCorp Vault. Implement HMAC-SHA256 master key derivation to deterministically compute per-epoch user blinding salts $s_i = \text{HMAC-SHA256}(K_{master}, UID_i \parallel epochId)$.
11. **Implement User Inclusion Proof gRPC and REST Service:**
    Develop `services/zk-pol-engine/src/grpc/` implementing `ZkPolService`. Provide RPC method `GetInclusionProof(GetInclusionProofRequest)` returning 64 sibling tuples and leaf commitments. Include in-memory LRU and Redis caching.
12. **Implement Hyperledger Besu Notarization Relayer:**
    Develop `services/zk-pol-engine/src/relayer.rs` consuming completed proof manifests, packing calldata matching `IZKProofOfSolvencyVerifier.sol`, and dispatching signed transactions to Besu via `services/cloudhsm-signer-daemon` (Prompt 717).
13. **Build Client-Side WebAssembly (Wasm) Verification SDK:**
    Develop `packages/zk-pol-wasm/` compiling Rust/Poseidon to Wasm. Expose JavaScript/TypeScript and Dart APIs allowing investor web and mobile clients to verify inclusion proofs locally in under 10 milliseconds.
14. **Implement Re-Audit and Sanity Invariant Worker:**
    Build an independent background reconciliation daemon in `services/zk-pol-engine/src/auditor.rs` that verifies every generated proof against the raw PostgreSQL balance snapshot before Besu submission, halting immediately on any sum mismatch.
15. **Construct Comprehensive Negative-Balance and Fuzzing Test Harness:**
    Implement exhaustive test suites in `tests/` verifying rejection of negative balances, field overflow injections ($p - 1$), altered sibling hashes, tampered liability sums, and replayed epoch proofs.

## Interfaces / Contracts

### 1. Protocol Buffers API (`proto/zk_pol.proto`)
```protobuf
syntax = "proto3";

package growww.zkpol.v1;

option go_package = "services/zk-pol-engine/pb;zkpolv1";

service ZkPolService {
  // Returns inclusion proof path and leaf commitment for a specific account and epoch
  rpc GetInclusionProof(GetInclusionProofRequest) returns (GetInclusionProofResponse);
  
  // Triggers batch tree construction and proof generation for a closed epoch
  rpc TriggerEpochProving(TriggerEpochProvingRequest) returns (TriggerEpochProvingResponse);
  
  // Queries notarized solvency attestation details for an epoch
  rpc GetEpochSolvencyStatus(GetEpochSolvencyStatusRequest) returns (GetEpochSolvencyStatusResponse);
}

message GetInclusionProofRequest {
  string account_id = 1;
  uint64 epoch_id = 2;
  string user_auth_token = 3;
}

message MerkleSiblingNode {
  uint32 height = 1;
  bytes sibling_hash = 2;       // 32 bytes Poseidon hash
  uint64 sibling_balance = 3;    // Subtree balance sum in paisa
  bool is_right_sibling = 4;     // true if sibling is right child, false if left
}

message GetInclusionProofResponse {
  uint64 epoch_id = 1;
  uint64 epoch_timestamp = 2;
  bytes root_hash = 3;           // 32 bytes Merkle Sum Tree root
  uint64 total_liabilities = 4;  // Total exchange liabilities
  bytes user_leaf_hash = 5;      // 32 bytes Poseidon leaf hash
  uint64 user_balance = 6;       // User balance in paisa
  bytes user_salt = 7;           // 32 bytes private blinding salt
  repeated MerkleSiblingNode path_elements = 8; // 64 sibling path elements
  bool is_verified = 9;
}

message TriggerEpochProvingRequest {
  uint64 epoch_id = 1;
  uint64 total_custodial_reserves = 2; // Verified depository reserves in paisa
  bytes custodian_manifest_hash = 3;   // SHA-256 hash of depository proof
}

enum TriggerEpochProvingStatus {
  TRIGGER_STATUS_QUEUED = 0;
  TRIGGER_STATUS_PROCESSING = 1;
  TRIGGER_STATUS_FAILED = 2;
}

message TriggerEpochProvingResponse {
  string job_id = 1;
  TriggerEpochProvingStatus status = 2;
  string message = 3;
}

message GetEpochSolvencyStatusRequest {
  uint64 epoch_id = 1;
}

message GetEpochSolvencyStatusResponse {
  uint64 epoch_id = 1;
  bytes root_hash = 2;
  uint64 total_liabilities = 3;
  uint64 total_custodial_reserves = 4;
  bool is_solvent = 5;
  string besu_tx_hash = 6;
  uint64 block_number = 7;
  bytes groth16_proof_json = 8;
}
```

### 2. Solvency Proof JSON Schema (`schemas/solvency_proof.json`)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "SolvencyProofManifest",
  "type": "object",
  "required": [
    "epochId",
    "epochTimestamp",
    "rootHash",
    "totalLiabilities",
    "totalCustodialReserves",
    "proof",
    "publicSignals"
  ],
  "properties": {
    "epochId": { "type": "integer", "minimum": 1 },
    "epochTimestamp": { "type": "integer", "minimum": 1700000000 },
    "rootHash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "totalLiabilities": { "type": "string", "pattern": "^[0-9]+$" },
    "totalCustodialReserves": { "type": "string", "pattern": "^[0-9]+$" },
    "proof": {
      "type": "object",
      "required": ["pi_a", "pi_b", "pi_c", "protocol"],
      "properties": {
        "pi_a": {
          "type": "array",
          "items": { "type": "string", "pattern": "^[0-9]+$" },
          "minItems": 3,
          "maxItems": 3
        },
        "pi_b": {
          "type": "array",
          "items": {
            "type": "array",
            "items": { "type": "string", "pattern": "^[0-9]+$" },
            "minItems": 2,
            "maxItems": 2
          },
          "minItems": 3,
          "maxItems": 3
        },
        "pi_c": {
          "type": "array",
          "items": { "type": "string", "pattern": "^[0-9]+$" },
          "minItems": 3,
          "maxItems": 3
        },
        "protocol": { "type": "string", "enum": ["groth16"] }
      }
    },
    "publicSignals": {
      "type": "array",
      "items": { "type": "string", "pattern": "^[0-9]+$" },
      "minItems": 5,
      "maxItems": 5
    }
  }
}
```

### 3. Circom Circuit Templates (`circuits/pol/`)

#### 64-Bit Range Check Template (`circuits/pol/range_check.circom`):
```circom
pragma circom 2.1.6;

template RangeCheck64() {
    signal input in;
    signal bits[64];
    var accumulated = 0;

    for (var i = 0; i < 64; i++) {
        bits[i] <-- (in >> i) & 1;
        // Enforce boolean constraint: bit * (1 - bit) == 0
        bits[i] * (1 - bits[i]) === 0;
        accumulated += bits[i] * (1 << i);
    }

    // Enforce exact binary reconstruction
    in === accumulated;
}
```

#### SMST Node Verification Template (`circuits/pol/smst_node.circom`):
```circom
pragma circom 2.1.6;

include "../node_modules/circomlib/circuits/poseidon.circom";
include "./range_check.circom";

template SMSTInternalNode() {
    signal input leftHash;
    signal input leftBalance;
    signal input rightHash;
    signal input rightBalance;

    signal output parentHash;
    signal output parentBalance;

    // 1. Verify non-negativity of both children
    component checkLeft = RangeCheck64();
    checkLeft.in <== leftBalance;

    component checkRight = RangeCheck64();
    checkRight.in <== rightBalance;

    // 2. Sum child balances
    parentBalance <== leftBalance + rightBalance;

    // 3. Verify parent non-negativity
    component checkParent = RangeCheck64();
    checkParent.in <== parentBalance;

    // 4. Compute Poseidon parent commitment: Poseidon(leftHash, leftBalance, rightHash, rightBalance)
    component hasher = Poseidon(4);
    hasher.inputs[0] <== leftHash;
    hasher.inputs[1] <== leftBalance;
    hasher.inputs[2] <== rightHash;
    hasher.inputs[3] <== rightBalance;

    parentHash <== hasher.out;
}
```

#### Top-Level Solvency Attestation Template (`circuits/pol/solvency_proof.circom`):
```circom
pragma circom 2.1.6;

include "./range_check.circom";
include "../node_modules/circomlib/circuits/comparators.circom";

template SolvencyProof() {
    // Public Inputs
    signal input merkleSumTreeRoot;
    signal input totalLiabilities;
    signal input totalCustodialReserves;
    signal input custodianManifestHash;
    signal input epochTimestamp;

    // Output assertion flag
    signal output isSolvent;

    // Enforce non-negativity of public totals
    component checkLiabilities = RangeCheck64();
    checkLiabilities.in <== totalLiabilities;

    component checkReserves = RangeCheck64();
    checkReserves.in <== totalCustodialReserves;

    // Solvency Invariant: totalLiabilities <= totalCustodialReserves
    component comp = LessEqThan(64);
    comp.in[0] <== totalLiabilities;
    comp.in[1] <== totalCustodialReserves;

    // Invariant must strictly evaluate to 1 (true)
    comp.out === 1;
    isSolvent <== comp.out;
}

component main {public [merkleSumTreeRoot, totalLiabilities, totalCustodialReserves, custodianManifestHash, epochTimestamp]} = SolvencyProof();
```

### 4. Solidity Verifier Contract Interface (`contracts/zk/IZKProofOfSolvencyVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IZKProofOfSolvencyVerifier {
    struct SolvencyRecord {
        uint256 epochId;
        uint256 timestamp;
        bytes32 merkleSumTreeRoot;
        uint256 totalLiabilities;
        uint256 totalCustodialReserves;
        bytes32 custodianManifestHash;
        address notarizedBy;
        uint256 blockNumber;
    }

    event SolvencyAttestationRecorded(
        uint256 indexed epochId,
        bytes32 indexed merkleSumTreeRoot,
        uint256 totalLiabilities,
        uint256 totalCustodialReserves,
        uint256 timestamp
    );

    event UserInclusionVerified(
        uint256 indexed epochId,
        bytes32 indexed leafHash,
        uint256 userBalance,
        bool isValid
    );

    function verifyAndRecordSolvency(
        uint256 epochId,
        uint256[2] calldata a,
        uint256[2][2] calldata b,
        uint256[2] calldata c,
        uint256[5] calldata publicInputs
    ) external returns (bool);

    function verifyUserInclusion(
        uint256 epochId,
        bytes32 leafHash,
        uint256 userBalance,
        bytes32[] calldata pathElements,
        uint64[] calldata siblingBalances,
        uint8[] calldata pathDirections
    ) external view returns (bool);

    function getEpochRecord(uint256 epochId) external view returns (SolvencyRecord memory);
}
```

## Security & Compliance Notes
- **DPDP Act 2023 Zero-PII Guarantee:**
  Under the Digital Personal Data Protection Act (DPDP Act 2023), customer balances and transactional holdings represent protected personal data. The ZK-PoL engine ensures that no customer identifiers (names, Permanent Account Numbers [PAN], Aadhaar hashes, bank account numbers, or plain account UUIDs) are ever exposed to the proving system, published on the blockchain, or visible in public APIs. Every account leaf is shielded by a 256-bit cryptographically secure blinding salt derived via HMAC-SHA256 from an HSM master key.
- **Mathematical Impossibility of Negative Balance Injections:**
  In classical financial reporting and naive Merkle sum schemes, insolvency can be masked by introducing synthetic credit balances (negative user accounts). Because the Circom arithmetic circuits strictly enforce $b_i \in [0, 2^{64}-1]$ via bit decomposition, any balance $b_i < 0$ or $b_i \ge 2^{64}$ cannot generate a valid witness. An exchange cannot artificially reduce its published liability root without triggering immediate proof generation failure.
- **Resistance to Balance Dictionary & Enumeration Attacks:**
  If account leaves were computed simply as $\text{Hash}(UID, b_i)$, an attacker knowing a user's $UID$ could perform offline rainbow table attacks to discover the user's exact balance $b_i$. By computing leaf hashes as $\text{Poseidon}(K_i, b_i, s_i)$ with high-entropy 256-bit secret salt $s_i$, dictionary and brute-force attacks are computationally infeasible ($> 128$ bits of security).
- **Trusted Setup Security & Toxic Waste Destruction:**
  The Groth16 zk-SNARK proving system relies on a multi-party computation (MPC) trusted setup ceremony. The ZK-PoL engine utilizes the publicly verifiable Hermez/Ethereum Foundation Perpetual Powers of Tau Phase 1 ceremony combined with an internally audited, air-gapped Phase 2 ceremony involving multiple independent institutional witnesses. Toxic waste parameters are provably destroyed upon completion.
- **Constant-Time Cryptographic Execution & Side-Channel Immunity:**
  All Rust field operations and C++ witness generators use constant-time Montgomery ladder multiplications to prevent cache-timing, branch-prediction, or power-analysis side-channel attacks during proof generation.
- **Regulatory Audit Trail & Tamper-Evidence:**
  All generated proofs, public inputs, and verification transactions are preserved immutably across both ClickHouse long-term audit storage and the Hyperledger Besu consortium ledger, satisfying SEBI circulars on continuous system auditability and financial records retention.

## Acceptance Criteria
- [ ] **SMST Synthesis Throughput:** The Rust tree engine constructs a depth-64 Sparse Merkle Sum Tree containing 1,000,000 active user account leaves in under 60 seconds on standard cloud infrastructure (32 vCPU, 64 GB RAM).
- [ ] **ZK Proving Performance:** The Groth16 prover produces a valid proof across all account liabilities within 180 seconds using multi-threaded CPU/GPU acceleration.
- [ ] **On-Chain Gas Consumption:** The `verifyAndRecordSolvency` transaction on `ZKProofOfSolvencyVerifier.sol` (Hyperledger Besu) consumes less than 250,000 gas, fitting comfortably within standard block gas limits.
- [ ] **Zero Underflow / Negative Balance Rejection:** Automated negative-testing test harnesses verify that injecting a balance of $-1$, $2^{64}$, or $p-1$ causes 100% immediate failure in witness generation.
- [ ] **Inclusion Proof Generation Latency:** The gRPC endpoint `GetInclusionProof` returns a complete 64-element sibling path with intermediate sums in under 5 milliseconds (p99).
- [ ] **Client-Side Verification Latency:** The WebAssembly verification SDK validates an inclusion proof against a published root in under 10 milliseconds in Chrome, Firefox, and Flutter mobile WebViews.
- [ ] **Zero PII Exposure Verification:** Static analysis and data flow scanners certify that no PAN, name, or plain UUID exists in generated proofs, public inputs, or on-chain payloads.
- [ ] **100% Automated Invariant Reconciliation:** The background audit daemon detects and blocks any discrepancy greater than zero paisa between wallet database ledger totals and the SMST root balance.

## Suggested Order / Dependencies
- **Upstream Dependencies (Must Be Built / Specified Prior to Deployment):**
  - **Prompt 203 (`203_wallet_account_service.md`):** Supplies daily double-entry user account balance snapshots (`ledger.snapshot.liabilities.v1`).
  - **Prompt 308 (`308_on_chain_proof_of_reserve_publishing.md`):** Coordinates on-chain reserve notarization and triggers liability proof generation.
  - **Prompt 317 (`317_zk_proof_of_solvency_verifier.md`):** Deploys the EVM Groth16 verifier contract (`ZKProofOfSolvencyVerifier.sol`) on Hyperledger Besu.
  - **Prompt 409 (`409_proof_of_reserve_sparse_merkle_tree_store.md`):** Establishes the underlying RocksDB schema, column families, and copy-on-write snapshotting standards.
  - **Prompt 717 (`717_hsm_kms_cloudhsm_key_lifecycle_and_signing_daemon.md`):** Provides CloudHSM master key derivation for user blinding salts and transaction signing.
- **Downstream Consumers (Build After / Concurrently):**
  - **Prompt 514 (`514_flutter_settings_profile_ui.md`):** Integrates the mobile client-side Wasm self-audit verification UI for end investors.
  - **Prompt 606 (`606_admin_proof_of_reserve_reconciliation.md`):** Displays operator dashboards for solvency verification, reserve matching, and audit tracking.
  - **Prompt 715 (`715_nbse_market_surveillance_and_sebi_reporting_engine.md`):** Consumes notarized solvency proofs for automated regulatory filings with SEBI and FIU-IND.
  - **Prompt 913 (`913_cross_market_reconciliation_and_settlement_fuzzing.md`):** Executes automated fuzz testing, stress tests, and invariant validation on the ZK-PoL engine.
