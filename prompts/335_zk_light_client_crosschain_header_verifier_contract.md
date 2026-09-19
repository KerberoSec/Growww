# 335 - ZK Light Client Cross-Chain Header & Transaction Inclusion Verifier Contract (ZKCrossChainLightClient.sol)

## Purpose
Institutional cross-chain asset bridging and collateral verification traditionally rely on centralized multi-signature federations, trusted relayer networks, or external custodial bridges. These legacy mechanisms introduce systemic security vulnerabilities, single points of failure, censorship vectors, and regulatory fragility. For the Growww National Blockchain Stock Exchange (NBSE) on Hyperledger Besu, accepting foreign blockchain assets (such as Bitcoin, Ethereum, and Solana) as cross-chain margin collateral or tokenized backing requires a completely trust-minimized, cryptographic verification paradigm.

Zero-Knowledge Succinct Non-Interactive Arguments of Knowledge (ZK-SNARKs) allow the mathematical verification of entire foreign blockchain state transitions and transaction inclusion proofs inside smart contracts at constant gas costs without running full nodes on-chain.

This prompt specifies the **ZK Light Client Cross-Chain Header and Transaction Inclusion Verifier Smart Contract Suite (`ZKCrossChainLightClient.sol`, `IZKCrossChainLightClient.sol`, `IZKProofVerifier.sol`)**. The contracts implement trust-minimized ZK-SNARK light client verification on Hyperledger Besu for Bitcoin, Ethereum, and Solana. The suite verifies foreign block header chains (Proof-of-Work difficulty targets, beacon chain sync committees, and slot roots), validates transaction and state Merkle inclusion proofs for collateral deposits, enforces automated fraud-proof challenge windows, and integrates Growww's canonical Universal Zero-Fee Model (0.00% fee - No fee at all) with a 0.00% fee at launch (governed by FeeController.sol) revenue split (with FIFO capital gains computed strictly for user tax compliance under Section 111A/112A).

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/bridges/` comprising:
- `ZKCrossChainLightClient.sol`: Master cross-chain light client registry and state verifier. It maintains the canonical verified block header states for supported foreign chains (Bitcoin, Ethereum, Solana), verifies zero-knowledge Groth16 / SP1 proofs of header validity and consensus rules, verifies Merkle / Merkle-Patricia inclusion proofs for deposit transactions, manages fraud-proof challenge periods, mints verified cross-chain collateral receipts, and routes the Universal Zero-Fee Model (0.00% fee - No fee at all) to Treasury (60%), Core SGF (25%), and IPF (15%) reserve vaults.
- `IZKCrossChainLightClient.sol` and `IZKProofVerifier.sol`: Comprehensive Solidity interface specifications declaring multi-chain data structures, enums, events, custom errors, and view/mutator function signatures for proof submission, inclusion verification, and header synchronization.
- `Groth16BN254Verifier.sol`: Highly optimized pairing-based elliptic curve BN254 / alt_bn128 verifier contract for ZK-SNARK proof verification on Besu.
- Comprehensive Foundry test harness (`test/bridges/`): Unit, fuzz, invariant, and cryptographic proof verification test suites validating header continuity, reorg handling, Merkle inclusion validation, challenge window mechanics, and strict compliance with the fixed 0.00% transaction fee (No fee at all) invariant.

## Scope Boundaries
- **In Scope:**
  - Trust-minimized ZK-SNARK block header chain verification for Bitcoin, Ethereum, and Solana.
  - Bitcoin consensus verification: sha256d block header hashing, target difficulty adjustment (2016-block epochs), cumulative chainwork validation, and 6-block depth finality confirmation.
  - Ethereum consensus verification: Beacon chain sync committee attestations (EIP-4788 / beacon state roots), BLS12-381 signature aggregation verification via ZK-SNARKs, and execution block header state root tracking.
  - Solana consensus verification: Bank hash / epoch state transitions, ed25519 vote transaction aggregations via ZK-SNARK, and slot root commitment tracking.
  - Merkle inclusion proof verification: Bitcoin SPV Merkle trees, Ethereum MPT (Merkle-Patricia Trie) receipt/state proofs, and Solana account state proofs for collateral deposit transactions.
  - Collateral deposit registration and locked receipt issuance with unique cryptographic commitments.
  - Automated fraud-proof challenge window and optimistic fallback verification mechanics.
  - Growww fixed 0.00% transaction fee (No fee at all) calculation on gross collateral deposit value with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), with FIFO capital gains for tax compliance (Section 111A/112A).
  - Emergency circuit breakers, pause mechanisms, and timelocked multi-signature administrative governance.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain ZK-SNARK prover infrastructure (SP1, RISC Zero, Gnark, Halo2 prover agents running on GPU/FPGA clusters - handled in Prompt 244).
  - Off-chain relayer network for submitting proofs to Hyperledger Besu (handled in Prompt 213).
  - Margin trading risk engine and liquidation coordinator (handled in Prompt 206 and Prompt 312).
  - Physical commodity and demat equity tokenization contracts (handled in Prompt 303 and Prompt 330).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD`, and custom errors).
- **Elliptic Curve Cryptography:** Native EVM precompiles `ecAdd` (0x06), `ecMul` (0x07), and `ecPairing` (0x08) on BN254 / alt_bn128 curve.
- **Zero-Knowledge Proof Systems:** **SP1 / Groth16** on BN254 curve with optimized verification gas footprint (<220,000 gas per proof).
- **Security & Modularity:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation and invariant fuzzing, `cast` for Besu RPC interactions).
- **Static Analysis & Auditing:** Slither, Mythril, and Solhint automated CI pipelines.

## Backend / Infra Touchpoints
- **ZK Prover Cluster & Relayer Service (Prompt 213 & Prompt 244):** Monitors foreign chains (Bitcoin bitcoind, Ethereum Geth/Prysm, Solana validator RPC), generates SP1 / Groth16 ZK-SNARK proofs of headers and transaction inclusions, and submits them to `ZKCrossChainLightClient.sol`.
- **Pre-Trade Risk & Margin Engine (Prompt 206):** Ingests `CollateralDepositVerified` events to grant real-time cross-chain trading margins on the NBSE exchange.
- **Liquidation & Settlement Service (Prompt 208 & Prompt 312):** Monitors collateral health based on cryptographically verified on-chain deposit states.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes Universal Zero-Fee Model (0.00% fee - No fee at all) deductions (0.00% fee at launch (governed by FeeController.sol) split) and computes Section 111A/112A capital gains tax compliance records.
- **Blockchain Event Indexer (Prompt 309):** Indexes `HeaderUpdated`, `ProofVerified`, `CollateralDepositVerified`, and `FraudChallengeOpened` events for downstream accounting and audit dashboards.
- **Proof-of-Reserve Registry (Prompt 308):** Correlates foreign lockup UTXOs / smart contract balances with verified on-chain light client commitments.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Executes on Hyperledger Besu private consortium network with QBFT consensus, 2-second deterministic block times, and immediate finality without chain reorganizations.
- **Zero Centralized Bridge Custody Invariant:** Collateral verification requires mathematical ZK-proofs of foreign consensus and Merkle inclusion proofs. No trusted multi-sig relayer can mint or verify collateral unbacked by verified foreign chain state.
- **Deterministic Cryptographic Verification:** All foreign block hashes, difficulty targets, and state roots are committed immutably. Once verified via ZK-SNARK, header roots are finalized following standard confirmation depths (e.g. 6 blocks for Bitcoin, 32 epochs for Ethereum, 31 confirmations for Solana).
- **Zero On-Chain PII:** Smart contract state stores only `bytes32 chainId`, `bytes32 blockHash`, `bytes32 txHash`, `bytes32 commitmentHash`, `address investorWallet`, `uint256 collateralValueSubPaise`, and proof calldata arrays. No legal names, PAN numbers, or IP addresses exist on-chain.
- **Multi-Party Governance:** Administrative functions (updating ZK verifier addresses, configuring challenge window durations, whitelisting supported foreign chain IDs, and updating fee treasury vaults) require 3-of-5 threshold signatures from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout:** Initialize `contracts/bridges/ZKCrossChainLightClient.sol`, `contracts/bridges/verifiers/Groth16BN254Verifier.sol`, interfaces under `contracts/interfaces/bridges/`, and test suites under `test/bridges/`.
2. **Define Data Models and Interfaces:** Write `IZKCrossChainLightClient.sol` and `IZKProofVerifier.sol` declaring all enums (`ChainIdentifier`, `ProofType`, `DepositState`, `ChallengeStatus`), structs, custom errors, and events.
3. **Implement Groth16 BN254 Verifier Contract (`Groth16BN254Verifier.sol`):**
   - Implement optimized pairing check on the BN254 / alt_bn128 curve utilizing EVM precompiles at addresses `0x06`, `0x07`, and `0x08`.
   - Implement `verifyProof(uint256[2] a, uint256[2][2] b, uint256[2] c, uint256[] input)` returning a boolean verification outcome.
4. **Implement Bitcoin Light Client Consensus Verifier in Master Contract:**
   - Implement `updateBitcoinHeaderChain` accepting block header batches and accompanying Groth16 / SP1 ZK-proofs.
   - Verify sha256d header hash chains, target difficulty retargeting across 2016-block epochs, and cumulative chainwork.
   - Store verified block hashes in `mapping(uint256 => bytes32) btcBlockHeaders` and update highest verified tip.
5. **Implement Ethereum Beacon Sync Committee Verifier in Master Contract:**
   - Implement `updateEthereumBeaconRoot` accepting beacon header sync committee ZK-proofs (verifying 512 BLS12-381 signatures aggregated via circuit).
   - Validate execution block header hash against EIP-4788 beacon state root commitments.
   - Store verified execution state roots in `mapping(uint256 => bytes32) ethExecutionStateRoots`.
6. **Implement Solana Epoch and Slot Verifier in Master Contract:**
   - Implement `updateSolanaSlotRoot` accepting ZK-proofs of validator epoch stake consensus and bank hash commitments.
   - Store verified slot roots in `mapping(uint64 => bytes32) solanaSlotRoots`.
7. **Implement Bitcoin SPV Merkle Transaction Inclusion Verifier:**
   - Implement `verifyBitcoinDepositInclusion` validating that a raw deposit transaction hashes into the Merkle root of a verified Bitcoin block at least 6 blocks deep.
   - Extract deposit output value, recipient script, and locked collateral commitment.
8. **Implement Ethereum MPT (Merkle-Patricia Trie) Inclusion Verifier:**
   - Implement `verifyEthereumDepositInclusion` validating RLP-encoded receipt / state trie proofs against verified execution block roots.
   - Verify deposit event logs emitted by verified custodial lockup contracts on Ethereum.
9. **Implement Solana Account State Inclusion Verifier:**
   - Implement `verifySolanaDepositInclusion` validating account state data proofs against verified slot bank hashes.
10. **Implement Collateral Receipt Minting & Lifecycle State Machine:**
    - Record verified collateral deposit in `mapping(bytes32 => CollateralDeposit) deposits`.
    - Enforce unique deposit commitment hashing: `keccak256(abi.encode(chainId, txHash, outputIndex, investorWallet, amount))`.
    - Mark deposit status as `VERIFIED_PENDING_CHALLENGE` or `FINALIZED`.
11. **Implement Automated Fraud-Proof Challenge Mechanism:**
    - Implement `openFraudChallenge` allowing whitelisted challengers or public bond posters to dispute invalid headers or double-spent UTXOs within the `challengeWindowSeconds` (e.g. 1800 seconds).
    - Implement `resolveFraudChallenge` executing slashed bond distributions and rolling back invalid header tips upon confirmed mathematical invalidity.
12. **Implement Fixed Transaction Fee & Multi-Vault Allocation Engine:**
    - Compute gross collateral value: `turnoverSubPaise = depositValue`.
    - Calculate fixed platform fee: `feeSubPaise = (turnoverSubPaise * feeBps) / 10000 (where feeBps == 0 at launch)` (exact 0.00% (Zero Fee) / 0 bps (0.00% fee at launch)).
    - Distribute fee: routed dynamically per FeeController governance (0.00% at launch).
    - Record tax compliance leaf hash for off-chain FIFO capital gains reporting (Section 111A/112A).
13. **Implement Circuit Breakers and Governance Controls:**
    - Inherit OpenZeppelin `PausableUpgradeable` and `AccessControlUpgradeable`.
    - Restrict emergency pause, unpause, verifier contract upgrades, and chain configuration to `DEFAULT_ADMIN_ROLE` (`MultiSigGovernance.sol`).
14. **Author Comprehensive Foundry Unit and Invariant Tests:**
    - Test Bitcoin header chain verification and reject invalid difficulty transitions.
    - Test Ethereum sync committee ZK-proof verification and Merkle-Patricia inclusion proofs.
    - Test Solana bank hash verification and account inclusion validation.
    - Invariant test: duplicate transaction hashes or reused deposit nonces must revert.
    - Invariant test: platform fee deduction must equal exactly 0 bps (0.00% fee at launch) with 0.00% fee launch policy distribution.
15. **Perform Slither Static Analysis and Gas Benchmarking:**
    - Run `slither contracts/bridges/` and verify zero high/medium security vulnerabilities.
    - Benchmark EVM pairing precompile gas usage to verify proof verification executes under 220,000 gas.

## Interfaces / Contracts

### 1. ZK Proof Verifier Interface (`IZKProofVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IZKProofVerifier
 * @notice Generic interface for Groth16 / SP1 SNARK proof verification on BN254 / alt_bn128 curve.
 */
interface IZKProofVerifier {
    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct Groth16Proof {
        uint256[2] a;
        uint256[2][2] b;
        uint256[2] c;
    }

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    /**
     * @notice Verifies a Groth16 ZK-SNARK proof against public inputs.
     * @param proof The Groth16 elliptic curve proof points (a, b, c).
     * @param publicInputs The array of public inputs associated with the circuit execution.
     * @return isValid True if the pairing check succeeds, false otherwise.
     */
    function verifyProof(
        Groth16Proof calldata proof,
        uint256[] calldata publicInputs
    ) external view returns (bool isValid);
}
```

### 2. ZK Cross-Chain Light Client Interface (`IZKCrossChainLightClient.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IZKProofVerifier} from "./IZKProofVerifier.sol";

/**
 * @title IZKCrossChainLightClient
 * @notice Master interface for the Trust-Minimized ZK Light Client Cross-Chain Verifier on Hyperledger Besu.
 * @dev Supports Bitcoin, Ethereum, and Solana header verification, Merkle inclusion proofs, and 0.00% fee (No fee at all) split.
 */
interface IZKCrossChainLightClient {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum ChainIdentifier {
        BITCOIN_MAINNET,
        ETHEREUM_MAINNET,
        SOLANA_MAINNET
    }

    enum ProofType {
        HEADER_CHAIN_UPDATE,
        TRANSACTION_INCLUSION,
        RECEIPT_INCLUSION,
        ACCOUNT_STATE_INCLUSION
    }

    enum DepositStatus {
        NONE,
        VERIFIED_PENDING_CHALLENGE,
        FINALIZED,
        DISPUTED_SLASHED
    }

    enum ChallengeStatus {
        NONE,
        OPEN,
        UPHELD_FRAUD_PROVEN,
        DISMISSED_VALID
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct BlockHeaderCommitment {
        bytes32 blockHash;           // Foreign canonical block hash
        bytes32 stateOrMerkleRoot;   // Merkle root (BTC) or State root (ETH/SOL)
        uint64 blockNumberOrSlot;    // Block height (BTC/ETH) or Slot number (SOL)
        uint64 timestamp;            // Foreign block timestamp
        uint256 cumulativeWorkOrEpoch;// Cumulative difficulty chainwork or epoch index
        uint64 verifiedAtBesuBlock;  // Besu block number when header was verified
    }

    struct BitcoinInclusionProof {
        uint64 blockHeight;          // Bitcoin block height containing transaction
        bytes32 txHash;              // Bitcoin transaction identifier (sha256d)
        uint32 outputIndex;          // Transaction output index (vout)
        uint256 valueSats;           // Deposited satoshi amount
        bytes rawTransaction;        // Serialized raw transaction bytes
        bytes32[] merkleBranch;      // SPV Merkle inclusion path
        uint32 txIndexInBlock;       // Transaction index in block Merkle tree
    }

    struct EthereumInclusionProof {
        uint64 blockNumber;          // Ethereum execution block height
        bytes32 txHash;              // Ethereum transaction hash
        bytes rlpReceipt;            // RLP-encoded transaction receipt
        bytes rlpPostStateOrStatus;  // Post-state root or status code
        bytes[] mptReceiptNodes;     // Merkle-Patricia Trie proof nodes
        bytes[] mptAccountNodes;     // State trie proof nodes (optional)
    }

    struct SolanaInclusionProof {
        uint64 slot;                 // Solana slot number
        bytes32 txSignature;         // Solana transaction signature
        bytes32 accountPubkey;       // Vault account address hash
        bytes accountStateBytes;     // Serialized account state data
        bytes32[] stateMerkleBranch; // Account state Merkle proof branch
    }

    struct CollateralDeposit {
        bytes32 depositCommitment;   // Unique deposit commitment hash
        ChainIdentifier chainId;     // Source foreign blockchain identifier
        bytes32 foreignTxHash;       // Source transaction hash
        address investorWallet;      // On-chain recipient ledger address
        uint256 collateralUnits;     // Normalized collateral units (18 decimals)
        uint256 grossConsiderationPaise; // Valuation in paise for fee assessment
        uint256 platformFeePaise;    // Deducted 0.00% (Zero Fee) platform fee
        DepositStatus status;        // Current deposit verification status
        uint64 submittedTimestamp;   // Submission timestamp
        uint64 challengeExpiry;      // Timestamp after which deposit becomes finalized
    }

    struct FraudChallenge {
        bytes32 challengeId;         // Unique challenge identifier
        bytes32 depositCommitment;   // Associated deposit commitment hash
        address challenger;          // Address of challenging party
        uint256 challengerBond;      // Escrowed bond amount
        ChallengeStatus status;      // Challenge resolution status
        bytes evidenceData;          // Cryptographic fraud evidence payload
        uint64 openedTimestamp;      // Timestamp when challenge was opened
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event HeaderChainUpdated(
        ChainIdentifier indexed chainId,
        bytes32 indexed blockHash,
        uint64 indexed blockNumberOrSlot,
        bytes32 stateOrMerkleRoot,
        uint256 cumulativeWorkOrEpoch
    );

    event CollateralDepositSubmitted(
        bytes32 indexed depositCommitment,
        ChainIdentifier indexed chainId,
        bytes32 indexed foreignTxHash,
        address investorWallet,
        uint256 collateralUnits,
        uint256 grossConsiderationPaise,
        uint64 challengeExpiry
    );

    event CollateralDepositFinalized(
        bytes32 indexed depositCommitment,
        address indexed investorWallet,
        uint256 netCollateralUnits,
        uint256 platformFeePaise,
        bytes32 taxProofHash
    );

    event FraudChallengeOpened(
        bytes32 indexed challengeId,
        bytes32 indexed depositCommitment,
        address indexed challenger,
        uint256 challengerBond
    );

    event FraudChallengeResolved(
        bytes32 indexed challengeId,
        bytes32 indexed depositCommitment,
        ChallengeStatus status,
        address indexed recipientOfBond
    );

    event PlatformFeeDistributed(
        bytes32 indexed depositCommitment,
        uint256 totalFeePaise,
        uint256 treasurySharePaise,
        uint256 coreSgfSharePaise,
        uint256 ipfSharePaise
    );

    event ZKVerifierUpdated(ChainIdentifier indexed chainId, address indexed newVerifier);

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error InvalidZKProof(ChainIdentifier chainId, ProofType proofType);
    error HeaderNotFound(ChainIdentifier chainId, uint64 blockNumberOrSlot);
    error InvalidBlockContinuity(bytes32 expectedParent, bytes32 providedParent);
    error InsufficientChainWork(uint256 providedWork, uint256 currentTipWork);
    error InsufficientConfirmationDepth(uint64 currentTip, uint64 requestedBlock, uint64 requiredDepth);
    error InvalidMerkleInclusionProof(bytes32 computedRoot, bytes32 expectedRoot);
    error DepositAlreadyExists(bytes32 depositCommitment);
    error DepositNotFound(bytes32 depositCommitment);
    error DepositNotPending(bytes32 depositCommitment, DepositStatus status);
    error ChallengePeriodStillActive(uint64 currentTimestamp, uint64 challengeExpiry);
    error ChallengePeriodExpired(uint64 currentTimestamp, uint64 challengeExpiry);
    error InsufficientChallengerBond(uint256 provided, uint256 required);
    error UnauthorizedCaller(address caller);
    error UnsupportedChainIdentifier(ChainIdentifier chainId);

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function updateBitcoinHeaderChain(
        bytes calldata blockHeadersBatch,
        IZKProofVerifier.Groth16Proof calldata zkProof,
        uint256[] calldata publicInputs
    ) external;

    function updateEthereumBeaconRoot(
        bytes32 beaconRoot,
        bytes32 executionStateRoot,
        uint64 executionBlockNumber,
        IZKProofVerifier.Groth16Proof calldata zkProof,
        uint256[] calldata publicInputs
    ) external;

    function updateSolanaSlotRoot(
        uint64 slotNumber,
        bytes32 bankHash,
        IZKProofVerifier.Groth16Proof calldata zkProof,
        uint256[] calldata publicInputs
    ) external;

    function submitBitcoinCollateralDeposit(
        BitcoinInclusionProof calldata proof,
        address recipientInvestor,
        uint256 valuationPaise
    ) external returns (bytes32 depositCommitment);

    function submitEthereumCollateralDeposit(
        EthereumInclusionProof calldata proof,
        address recipientInvestor,
        uint256 valuationPaise
    ) external returns (bytes32 depositCommitment);

    function submitSolanaCollateralDeposit(
        SolanaInclusionProof calldata proof,
        address recipientInvestor,
        uint256 valuationPaise
    ) external returns (bytes32 depositCommitment);

    function finalizeCollateralDeposit(
        bytes32 depositCommitment
    ) external returns (uint256 netCollateralUnits);

    function openFraudChallenge(
        bytes32 depositCommitment,
        bytes calldata evidenceData
    ) external payable returns (bytes32 challengeId);

    function resolveFraudChallenge(
        bytes32 challengeId,
        bool upholdFraud
    ) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function getLatestVerifiedHeader(ChainIdentifier chainId) external view returns (BlockHeaderCommitment memory);
    function getHeaderByNumber(ChainIdentifier chainId, uint64 blockNumberOrSlot) external view returns (BlockHeaderCommitment memory);
    function getDeposit(bytes32 depositCommitment) external view returns (CollateralDeposit memory);
    function getChallenge(bytes32 challengeId) external view returns (FraudChallenge memory);
    function zkVerifiers(ChainIdentifier chainId) external view returns (address);
    function challengeWindowSeconds() external view returns (uint64);
    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
    function computeTransactionFee(
        uint256 grossValuationPaise
    ) external pure returns (uint256 totalFee, uint256 treasuryShare, uint256 coreSgfShare, uint256 ipfShare);
}
```

## Security & Compliance Notes
- **Zero Centralized Federation Reliance:** The light client relies entirely on mathematical ZK-SNARK proofs on the BN254 / alt_bn128 curve and cryptographic Merkle tree inclusion paths. No multi-sig relayer keys or centralized custody federations possess the capability to forge header state or mint unbacked collateral receipts.
- **Fixed Platform Fee Integrity:** In strict accordance with the Growww Core Fee Model (Prompt 006), the platform fee is assessed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover with an automated 0.00% fee at launch (governed by FeeController.sol) revenue split. Capital gains are computed strictly for off-chain user tax compliance (Section 111A/112A). Zero holding or custody fees are charged.
- **Reorganization & Confirmation Depth Protections:** Foreign block headers require rigorous minimum finality confirmation depths (e.g. 6 Bitcoin blocks, 32 Ethereum epochs) before collateral inclusion proofs can be submitted, rendering deep chain reorganization attacks economically infeasible.
- **Automated Optimistic Fraud Challenge Period:** In addition to ZK validity proofs, high-value collateral deposits enter a timelocked challenge window (`challengeWindowSeconds`, default 1800 seconds). Bonded challengers can dispute fraudulent double-spent UTXOs or manipulated state roots, triggering automated slashing of offending relayer deposits.
- **Replay Protection & Unique Commitments:** Every deposit creates a unique cryptographic commitment `depositCommitment = keccak256(abi.encode(chainId, txHash, outputIndex, investorWallet, amount))`. Re-submitting identical foreign deposit transactions reverts immediately.
- **Zero On-Chain PII:** The smart contract stores only cryptographic hashes (`depositCommitment`, `foreignTxHash`, `taxProofHash`) and public addresses. All investor identities and KYC compliance data remain strictly off-chain in encrypted compliance registries.
- **Multi-Signature Governance:** Privileged actions (upgrading ZK verifier contracts, tuning challenge periods, and configuring supported foreign chain parameters) require 3-of-5 multi-signature authorization from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Acceptance Criteria
- [ ] `IZKProofVerifier.sol` and `IZKCrossChainLightClient.sol` compiled under Solidity 0.8.24 with zero warnings.
- [ ] `Groth16BN254Verifier.sol` integrates EVM precompiles (`0x06`, `0x07`, `0x08`) and verifies valid Groth16 proofs under 220,000 gas.
- [ ] Bitcoin header verifier validates sha256d hash chains, 2016-block difficulty target transitions, and cumulative chainwork.
- [ ] Ethereum beacon sync committee verifier tracks execution block state roots via EIP-4788 commitments.
- [ ] Solana verifier validates bank hashes and slot root continuity.
- [ ] Bitcoin SPV Merkle proof verification correctly verifies transaction inclusion in verified block headers at >= 6 confirmation depth.
- [ ] Ethereum Merkle-Patricia Trie verification correctly validates receipt / state inclusion proofs.
- [ ] Replay protection prevents duplicate deposit commitment creation.
- [ ] Fraud challenge window allows bonded dispute submission and enforces timelocked collateral finalization.
- [ ] Platform fee computation mathematically enforces exact 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) deduction on collateral valuation with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol).
- [ ] Invariant fuzz tests in Foundry execute 10,000 runs confirming zero state corruption or unauthorized collateral minting.
- [ ] Slither and Mythril static analysis suites pass with zero high or medium severity vulnerabilities.
- [ ] Zero em dashes and zero en dashes present across entire specification documentation.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `303` (Token Issuance Smart Contract - ERC-3643), Prompt `305` (Transfer Compliance Hooks), Prompt `306` (Settlement DvP Smart Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `206` (Risk & Margin Engine), Prompt `208` (Trade Settlement Service), Prompt `210` (Fee & Realized PnL Engine), Prompt `213` (Custodian Depository Integration Service), Prompt `244` (ZK Prover Infrastructure).
- **Subsequent Prompts Enabled:** Prompt `308` (Proof of Reserve Registry), Prompt `309` (Event Indexing Service), Prompt `312` (Liquidation Smart Contract), Prompt `509` (Flutter Order Placement Flow).
