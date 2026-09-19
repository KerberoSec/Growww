# 322 - Bitcoin SPV Header Relay & Discreet Log Contracts (DLC) Settlement Bridge (BitcoinSPVBridge.sol, DLCRegistry.sol)

## Purpose
Cross-border institutional trade finance, foreign portfolio investment in GIFT City IFSC, and Bitcoin-collateralized equity settlement require trustless, cryptographically verifiable interoperability between native Bitcoin (Layer 1) and the Hyperledger Besu permissioned ledger. Traditional cross-chain bridges rely on centralized custodian escrow or vulnerable multi-party multi-sig wrapping schemes, creating significant counterparty risk, custodial liability, and regulatory vulnerabilities under IFSCA and SEBI frameworks.

This prompt specifies the architecture, Solidity smart contracts (`BitcoinSPVBridge.sol`, `DLCRegistry.sol`), cryptographic libraries, and relayer infrastructure for validating Bitcoin Simplified Payment Verification (SPV) proofs and coordinating Discreet Log Contracts (DLC) directly on Hyperledger Besu. By maintaining an on-chain header relay with cumulative proof-of-work validation and evaluating Schnorr oracle attestations (BIP-340), the system enables atomic Delivery-versus-Payment (DvP) settlements between native Bitcoin UTXOs and tokenized fractional securities without centralized wrapping intermediaries or custodial risk.

## What You Are Building
A high-assurance, gas-optimized Bitcoin SPV verification engine and Discreet Log Contract settlement hub under `contracts/bitcoin-bridge/` and `services/spv-relayer/` comprising:
- `BitcoinSPVBridge.sol`: On-chain Bitcoin block header relay and SPV transaction inclusion proof validator. It verifies 80-byte block headers, computes double-SHA256 hashes, tracks cumulative chain work, enforces difficulty retargeting rules (2016-block epochs), and validates Merkle inclusion proofs for target transactions.
- `DLCRegistry.sol`: On-chain state machine and registry for Bitcoin Discreet Log Contracts. It registers DLC terms (CET commitments, collateral allocations, outcome payout vectors, and oracle public keys), verifies Schnorr oracle outcome attestations, coordinates funding and closing proofs via SPV, and unlocks corresponding tokenized settlement assets on Hyperledger Besu.
- `BTCUtils.sol` & `BitcoinHeaderParser.sol`: Core cryptographic libraries for endian reversal, CompactBits target calculation, block header parsing, transaction ID hashing, and Merkle tree path evaluation.
- `DLCSchnorrVerifier.sol`: EVM-optimized secp256k1 curve arithmetic library validating BIP-340 Schnorr signatures, tagged hashes, and oracle point commitments for discrete numerical/categorical settlement outcomes.
- `SPVRelayerDaemon` (`services/spv-relayer/` in Go): High-reliability off-chain daemon monitoring Bitcoin Core full nodes (via RPC/ZMQ), assembling and submitting serialized header batches to `BitcoinSPVBridge.sol`, and generating Merkle SPV inclusion proofs for registered DLC transactions.
- Comprehensive Foundry test suite (`test/bitcoin-bridge/`) verifying header reorgs, difficulty adjustments, false proof rejections, Schnorr oracle signature validations, and edge-case dispute resolutions.

## Scope Boundaries
- **In Scope:**
  - On-chain storage and validation of serialized 80-byte Bitcoin block headers.
  - Difficulty retargeting verification per 2016-block epoch (BIP-9/Consensus target limits).
  - Cumulative proof-of-work tracking and longest-chain fork resolution on Besu EVM.
  - Merkle tree path verification (SPV) proving Bitcoin transaction inclusion within a finalized block.
  - Parsing Bitcoin transaction outputs (OP_RETURN metadata, P2WSH, and P2TR/Taproot outpoints).
  - DLC agreement lifecycle: Registration, Funding Verification (via SPV), Oracle Attestation Verification, Settlement Execution, and Refund/Timeout processing.
  - BIP-340 Schnorr signature validation on secp256k1 over EVM precompiles and modular arithmetic.
  - Configurable transaction confirmation depth (default: 6 blocks) to guard against Bitcoin chain reorgs.
- **Out of Scope / Handled Elsewhere:**
  - Direct execution of Bitcoin script or mining validation beyond SPV headers (handled by external Bitcoin nodes).
  - Off-chain Bitcoin P2P Lightning channel routing (handled via separate payment layer).
  - Domestic INR fiat settlement and banking reservation (handled in Prompt 208 and Prompt 212).
  - GIFT City foreign exchange conversion (handled in Prompt 214).
  - EVM-to-EVM cross-chain bridging via CCIP/LayerZero (handled in Prompt 319).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Cancun/Prague on Hyperledger Besu).
  *Justification:* Provides native custom errors, transient storage (`TSTORE`/`TLOAD`) for gas-efficient Merkle verification loops, and strict overflow checking.
- **Blockchain Ledger:** **Hyperledger Besu v24.x+** (QBFT consensus, 2-second block period, Chain ID `13370`).
- **Cryptographic Specifications & BIP Standards:**
  - **BIP-340:** Schnorr Signatures for secp256k1.
  - **BIP-341 / BIP-342:** Taproot and Tapscript structure validation.
  - **BIP-141:** Segregated Witness (SegWit) transaction parsing.
  - **DLC Specification (v0.1):** Discreet Log Contracts for conditional digital settlement.
- **Libraries & Standards:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `UUPSUpgradeable`).
- **Relayer Service Language:** **Go (v1.22+)** with `btcsuite/btcd`, `ethereum/go-ethereum`, and PostgreSQL for state caching.
- **Testing & Verification Tooling:** Foundry (`forge`, `cast`), Slither static analyzer, and Halmos formal verification tools.

## Backend / Infra Touchpoints
- **Trade Settlement & DvP Orchestration Service (Prompt 208):** Coordinates simultaneous release of fractional equity tokens upon confirmed Bitcoin DLC settlement.
- **Foreign Investor Funding & FX Service (Prompt 214):** Utilizes Bitcoin SPV verification to confirm cross-border institutional capital inflows into GIFT City escrow vaults.
- **Settlement DvP Smart Contract (Prompt 306):** Executes atomic equity delivery against verified on-chain Bitcoin DLC payouts.
- **MultiSig Governance Smart Contract (Prompt 307):** Controls bridge parameters, trusted oracle public key registries, epoch checkpoints, and emergency pause controls.
- **Event Indexing Service (Prompt 309):** Indexes `BitcoinHeaderInserted`, `TransactionVerifiedSPV`, and `DLCSettled` events for sub-second UI notifications and risk monitoring.
- **Validator Key Management HSM (Prompt 311):** Secures relayer and governance keys in FIPS 140-2 Level 3 hardware.

## Blockchain Interaction
- **Bitcoin Header Ingestion:** Relayer calls `BitcoinSPVBridge.relayHeaders(bytes headers)` with a contiguous byte array of 80-byte headers. The bridge validates proof-of-work, calculates double-SHA256 hashes, checks difficulty targets, and extends the active tip.
- **SPV Inclusion Verification:** Client or settlement service calls `BitcoinSPVBridge.verifyTransactionInclusion(bytes32 txid, uint256 blockHeight, bytes intermediateNodes, uint256 index)`. The contract confirms the Merkle root matches the stored block header and satisfies confirmation depth $D \ge 6$.
- **DLC Lifecycle Management:**
  1. *Registration:* Parties submit DLC terms via `DLCRegistry.createDLC(DLCTerms terms)` including oracle public keys, collateral commitments, and outcome settlement vectors.
  2. *Funding Attestation:* Relayer provides SPV proof that the 2-of-2 multi-sig / Taproot funding UTXO is confirmed on Bitcoin L1.
  3. *Oracle Attestation:* At maturity, the designated oracle signs the outcome value. The signature is submitted to `DLCRegistry.settleDLC(bytes32 dlcId, uint256 outcomeValue, bytes oracleSignature)`.
  4. *Atomic Payout:* The contract validates the Schnorr signature against the oracle public key, computes the payout split, and triggers atomic equity/token transfers on Besu.
  5. *Dispute / Refund:* If the oracle fails to attest before the timeout, either party can trigger `DLCRegistry.refundDLC(bytes32 dlcId)` after the timelock expires.
- **Zero PII Policy:** All on-chain records consist exclusively of cryptographic transaction IDs, public keys, Merkle roots, outpoint hashes, satoshi balances, and execution nonces.

## Step-by-Step Build Instructions
1. Scaffold project directory structure:
   - `contracts/bitcoin-bridge/BitcoinSPVBridge.sol`
   - `contracts/bitcoin-bridge/DLCRegistry.sol`
   - `contracts/bitcoin-bridge/libraries/BTCUtils.sol`
   - `contracts/bitcoin-bridge/libraries/BitcoinHeaderParser.sol`
   - `contracts/bitcoin-bridge/libraries/DLCSchnorrVerifier.sol`
   - `contracts/interfaces/IBitcoinSPVBridge.sol`
   - `contracts/interfaces/IDLCRegistry.sol`
   - `services/spv-relayer/` (Go daemon)
   - `test/bitcoin-bridge/` (Foundry test suite)
2. Implement `BTCUtils.sol`:
   - Implement `reverseEndianness(bytes32 input) internal pure returns (bytes32)` for little-endian to big-endian hash conversion.
   - Implement `dsha256(bytes memory data) internal pure returns (bytes32)` for Bitcoin standard double-SHA256 hashing.
   - Implement `extractTarget(uint32 bits) internal pure returns (uint256)` decoding Bitcoin compact target representation (`nBits`).
   - Implement `calculateWork(uint256 target) internal pure returns (uint256)` calculating cumulative chain work: $W = \lfloor 2^{256} / (T + 1) \rfloor$.
3. Implement `BitcoinHeaderParser.sol`:
   - Implement extraction helpers for 80-byte serialized header fields:
     - `version` (bytes 0-3, uint32 little-endian)
     - `prevBlockHash` (bytes 4-35, bytes32 little-endian)
     - `merkleRoot` (bytes 36-67, bytes32 little-endian)
     - `timestamp` (bytes 68-71, uint32 little-endian)
     - `bits` (bytes 72-75, uint32 little-endian)
     - `nonce` (bytes 76-79, uint32 little-endian)
4. Implement `DLCSchnorrVerifier.sol`:
   - Implement BIP-340 tagged hashing: `taggedHash(string tag, bytes data)`.
   - Implement Schnorr signature validation over secp256k1 curve ($y^2 = x^3 + 7 \pmod p$):
     - Extract $R_x$ (first 32 bytes) and $s$ (last 32 bytes).
     - Verify $s < n$ and $R_x < p$.
     - Compute challenge $e = \text{BIP340Hash}(\text{"BIP0340/challenge"}, R_x \parallel P \parallel m) \pmod n$.
     - Verify curve point equality: $s \cdot G = R + e \cdot P$.
5. Implement `BitcoinSPVBridge.sol`:
   - Storage layout:
     - `mapping(bytes32 => BlockHeader) public headers;`
     - `mapping(uint256 => bytes32) public canonicalChain;` (height to block hash)
     - `bytes32 public highestBlockHash;`
     - `uint256 public highestBlockHeight;`
     - `uint256 public minConfirmations;` (default: 6)
   - Implement `initialize(bytes memory genesisHeader, uint256 genesisHeight, uint256 confirmations)` with UUPS proxy support.
   - Implement `relayHeaders(bytes calldata headers)`:
     - Verify calldata length is a positive multiple of 80 bytes.
     - Loop through headers, validating that each header's `prevBlockHash` matches an existing validated header.
     - Verify header double-SHA256 meets the target specified by its `bits` field.
     - Validate difficulty retargeting every 2016 blocks against epoch start timestamps.
     - Compute cumulative chain work and update `highestBlockHash` if the newly submitted branch exceeds the current tip work.
     - Emit `BitcoinHeaderInserted(bytes32 indexed blockHash, uint256 indexed height, bytes32 merkleRoot)`.
   - Implement `verifyTransactionInclusion(bytes32 txid, uint256 blockHeight, bytes calldata proof, uint256 index)`:
     - Retrieve canonical block hash at `blockHeight`.
     - Confirm `highestBlockHeight >= blockHeight + minConfirmations - 1`.
     - Compute Merkle root by hashing `txid` with `proof` sibling hashes using `index` bit-shifts.
     - Revert with `InvalidMerkleProof` if computed root does not match stored `merkleRoot`.
     - Return `true` upon successful proof validation.
6. Implement `DLCRegistry.sol`:
   - Storage layout:
     - `mapping(bytes32 => DLCState) public dlcAgreements;`
     - `mapping(address => bool) public authorizedOracles;`
     - `mapping(bytes32 => bool) public usedFundingUTXOs;`
   - Implement `initialize(address admin, address spvBridgeAddress)` with UUPS proxy support.
   - Implement `createDLC(DLCTerms calldata terms)`:
     - Validate collateral amounts, maturity timestamps, refund timelocks, and oracle public keys.
     - Lock collateral tokens (or fractional equity units) from the initiating party into contract escrow.
     - Transition DLC status to `Offered` or `Active`.
     - Emit `DLCOffered(bytes32 indexed dlcId, address indexed initiator, address counterparty)`.
   - Implement `acceptDLC(bytes32 dlcId, bytes32 fundingTxid, uint256 fundingVout)`:
     - Lock counterparty collateral into contract escrow.
     - Bind Bitcoin funding UTXO outpoint (`fundingTxid:fundingVout`).
     - Transition DLC status to `FundedPendingSPV`.
     - Emit `DLCAccepted(bytes32 indexed dlcId, bytes32 fundingTxid, uint256 fundingVout)`.
   - Implement `verifyFundingSPV(bytes32 dlcId, uint256 blockHeight, bytes calldata proof, uint256 index)`:
     - Call `IBitcoinSPVBridge(spvBridge).verifyTransactionInclusion(fundingTxid, blockHeight, proof, index)`.
     - Require `!usedFundingUTXOs[fundingUTXO]`; mark UTXO as used.
     - Transition DLC status to `ConfirmedActive`.
     - Emit `DLCFundedConfirmed(bytes32 indexed dlcId, bytes32 blockHash)`.
   - Implement `settleDLC(bytes32 dlcId, uint256 outcomeValue, bytes calldata oracleSignature)`:
     - Require DLC status is `ConfirmedActive` and `block.timestamp >= terms.maturityTimestamp`.
     - Verify `DLCSchnorrVerifier.verifySignature(terms.oraclePubKey, outcomeValue, oracleSignature)`.
     - Calculate payout shares for Party A and Party B based on the pre-committed payout function vector.
     - Distribute locked collateral and tokenized securities atomically to respective parties.
     - Transition DLC status to `Settled`.
     - Emit `DLCSettled(bytes32 indexed dlcId, uint256 outcomeValue, uint256 payoutPartyA, uint256 payoutPartyB)`.
   - Implement `refundDLC(bytes32 dlcId)`:
     - Require DLC status is `ConfirmedActive` and `block.timestamp >= terms.refundTimestamp`.
     - Return locked collateral 100% to respective depositors.
     - Transition DLC status to `Refunded`.
     - Emit `DLCRefunded(bytes32 indexed dlcId)`.
7. Build `SPVRelayerDaemon` in Go (`services/spv-relayer/`):
   - Connect to Bitcoin Core RPC endpoint and Hyperledger Besu JSON-RPC.
   - Poll Bitcoin block tip; on new blocks, serialize 80-byte headers into batches (e.g. up to 10 headers per Besu tx).
   - Generate Merkle transaction inclusion proofs using Bitcoin Core `gettxoutproof` or internal block traversal.
   - Automatically submit `relayHeaders` and `verifyFundingSPV` transactions to Besu with mTLS and HSM signing.
8. Write comprehensive Foundry tests (`test/bitcoin-bridge/BitcoinSPVBridge.t.sol`):
   - Test sequential header ingestion with real mainnet Bitcoin header slices.
   - Test fork choice rule: simulate a 3-block reorg and assert canonical chain realignment to higher cumulative work.
   - Test difficulty retargeting validation at epoch boundaries.
   - Test invalid proof-of-work rejection and malformed header length rejection.
   - Test Merkle inclusion verification with valid transaction paths and corrupted sibling nodes.
9. Write Foundry tests (`test/bitcoin-bridge/DLCRegistry.t.sol`):
   - Test full DLC lifecycle: Create -> Accept -> SPV Fund -> Settle with valid Schnorr signature.
   - Test settlement rejection on forged oracle signature or mismatched outcome value.
   - Test refund execution after `refundTimestamp` expiration and rejection prior to expiry.
   - Test reentrancy attack resistance on collateral payout transfers.
10. Execute Slither static analysis and gas optimization benchmarks on Merkle and Schnorr verification paths.
11. Deploy contracts to Hyperledger Besu testnet and configure relayer daemon with Bitcoin testnet4/signet.

## Interfaces / Contracts

### Bitcoin SPV Bridge Interface (`IBitcoinSPVBridge.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IBitcoinSPVBridge {
    struct BlockHeader {
        bytes32 blockHash;
        bytes32 prevBlockHash;
        bytes32 merkleRoot;
        uint32 timestamp;
        uint32 bits;
        uint32 nonce;
        uint32 version;
        uint256 height;
        uint256 chainWork;
    }

    event BitcoinHeaderInserted(
        bytes32 indexed blockHash,
        uint256 indexed height,
        bytes32 merkleRoot,
        uint256 chainWork
    );

    event CanonicalChainReorganized(
        bytes32 indexed oldTipHash,
        bytes32 indexed newTipHash,
        uint256 newHeight,
        uint256 reorgDepth
    );

    event MinimumConfirmationsUpdated(uint256 oldConfirmations, uint256 newConfirmations);

    // Errors
    error InvalidHeaderLength(uint256 providedLength);
    error PreviousHeaderNotFound(bytes32 prevBlockHash);
    error InsufficientProofOfWork(bytes32 blockHash, uint256 target);
    error InvalidDifficultyBits(uint32 bits, uint32 expectedBits);
    error BlockTimestampTooOld(uint32 timestamp, uint32 medianTimestamp);
    error InsufficientConfirmations(uint256 currentHeight, uint256 requiredHeight);
    error InvalidMerkleProof(bytes32 txid, bytes32 computedRoot, bytes32 expectedRoot);
    error InvalidTransactionIndex(uint256 index, uint256 maxIndex);

    function relayHeaders(bytes calldata headers) external returns (bytes32 newTip);

    function verifyTransactionInclusion(
        bytes32 txid,
        uint256 blockHeight,
        bytes calldata proof,
        uint256 index
    ) external view returns (bool isValid);

    function getHeaderByHeight(uint256 height) external view returns (BlockHeader memory);

    function getHeaderByHash(bytes32 blockHash) external view returns (BlockHeader memory);

    function getHighestBlockHeight() external view returns (uint256);

    function getHighestBlockHash() external view returns (bytes32);

    function isCanonical(bytes32 blockHash, uint256 height) external view returns (bool);
}
```

### DLC Registry Interface (`IDLCRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IDLCRegistry {
    enum DLCStatus {
        None,
        Offered,
        Accepted,
        FundedPendingSPV,
        ConfirmedActive,
        Settled,
        Refunded,
        Disputed
    }

    struct PayoutCurvePoint {
        uint256 outcomeValue;      // Discrete outcome / price index value
        uint256 payoutPartyA;      // Payout to Party A in token units / wei
        uint256 payoutPartyB;      // Payout to Party B in token units / wei
    }

    struct DLCTerms {
        bytes32 dlcId;
        address partyA;
        address partyB;
        address settlementToken;   // ERC-20 token or fractional equity token
        uint256 collateralPartyA;
        uint256 collateralPartyB;
        bytes32 oraclePubKey;      // BIP-340 32-byte x-only public key
        bytes32 oracleRPoint;      // Oracle pre-committed nonce point
        uint64 maturityTimestamp;
        uint64 refundTimestamp;
        bytes32 fundingTxid;
        uint32 fundingVout;
        uint256 totalOutcomes;
    }

    struct DLCState {
        DLCTerms terms;
        DLCStatus status;
        uint256 confirmedBlockHeight;
        uint256 settledOutcome;
        uint256 settledTimestamp;
    }

    event DLCOffered(bytes32 indexed dlcId, address indexed partyA, address indexed partyB, uint256 collateralA);
    event DLCAccepted(bytes32 indexed dlcId, address indexed partyB, uint256 collateralB);
    event DLCFundingSubmitted(bytes32 indexed dlcId, bytes32 indexed fundingTxid, uint32 fundingVout);
    event DLCFundedConfirmed(bytes32 indexed dlcId, bytes32 indexed blockHash, uint256 height);
    event DLCSettled(bytes32 indexed dlcId, uint256 outcomeValue, uint256 payoutPartyA, uint256 payoutPartyB);
    event DLCRefunded(bytes32 indexed dlcId, uint256 refundPartyA, uint256 refundPartyB);
    event OracleRegistered(bytes32 indexed oraclePubKey, string oracleName);
    event OracleRevoked(bytes32 indexed oraclePubKey);

    // Errors
    error DLCAlreadyExists(bytes32 dlcId);
    error DLCNotFound(bytes32 dlcId);
    error InvalidDLCStatus(bytes32 dlcId, DLCStatus currentStatus, DLCStatus expectedStatus);
    error UnauthorizedCaller(address caller);
    error InvalidCollateralAmount(uint256 provided, uint256 required);
    error InvalidTimestamps(uint64 maturity, uint64 refund);
    error FundingUTXOAlreadyUsed(bytes32 txid, uint32 vout);
    error MaturityTimestampNotReached(uint64 maturity, uint256 currentTimestamp);
    error RefundTimestampNotReached(uint64 refund, uint256 currentTimestamp);
    error InvalidOracleSignature(bytes32 dlcId, bytes32 oraclePubKey);
    error UnauthorizedOracle(bytes32 oraclePubKey);
    error PayoutPointMismatch(uint256 outcome, uint256 totalCollateral);

    function createDLC(DLCTerms calldata terms, PayoutCurvePoint[] calldata payoutCurve) external returns (bytes32 dlcId);

    function acceptDLC(bytes32 dlcId) external;

    function submitFundingProof(
        bytes32 dlcId,
        bytes32 fundingTxid,
        uint32 fundingVout,
        uint256 blockHeight,
        bytes calldata spvProof,
        uint256 txIndex
    ) external;

    function settleDLC(
        bytes32 dlcId,
        uint256 outcomeValue,
        bytes calldata schnorrSignature
    ) external;

    function refundDLC(bytes32 dlcId) external;

    function getDLCState(bytes32 dlcId) external view returns (DLCState memory);
}
```

### Relayer Batch Ingestion Schema (`spv-header-batch.json`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "BitcoinSPVHeaderBatch",
  "type": "object",
  "required": [
    "batchId",
    "network",
    "startHeight",
    "endHeight",
    "headerCount",
    "serializedHeadersHex",
    "relayerAddress",
    "timestamp"
  ],
  "properties": {
    "batchId": {
      "type": "string",
      "example": "BTC-BATCH-MAINNET-840000-840010"
    },
    "network": {
      "type": "string",
      "enum": ["mainnet", "testnet4", "signet"]
    },
    "startHeight": {
      "type": "integer",
      "minimum": 0,
      "example": 840000
    },
    "endHeight": {
      "type": "integer",
      "minimum": 0,
      "example": 840010
    },
    "headerCount": {
      "type": "integer",
      "minimum": 1,
      "maximum": 50,
      "example": 10
    },
    "serializedHeadersHex": {
      "type": "string",
      "pattern": "^0x[0-9a-fA-F]{160,}$",
      "description": "Concatenated 80-byte raw Bitcoin block headers in hex format"
    },
    "relayerAddress": {
      "type": "string",
      "pattern": "^0x[0-9a-fA-F]{40}$"
    },
    "timestamp": {
      "type": "integer",
      "example": 1713567890
    }
  }
}
```

## Security & Compliance Notes
- **51% Attack and Reorg Protection:** The bridge enforces a configurable confirmation threshold ($D \ge 6$ on Bitcoin L1) before any transaction is treated as finalized for DLC funding or settlement. In the event of a chain reorganization, the bridge re-evaluates the canonical chain strictly by cumulative proof-of-work, not by block height alone.
- **Difficulty Time-Warp Mitigation:** Difficulty adjustments every 2016 blocks must adhere to consensus rules bounding adjustments to a factor of 4 in either direction ($T_{\text{new}} \in [T_{\text{old}} / 4, T_{\text{old}} \cdot 4]$). Block timestamps must be strictly greater than the median of the previous 11 blocks (Median Time Past / MTP).
- **Schnorr Oracle Nonce Replay Prevention:** DLC settlement requires BIP-340 Schnorr signature validation with unique pre-committed oracle $R$-points. Oracles that sign contradictory outcomes with the same nonce point compromise their private keys via standard discrete logarithm extraction, creating a strong economic deterrent against oracle equivocation.
- **UTXO Double-Spend & Replay Protection:** `DLCRegistry.sol` maintains a global registry of used Bitcoin outpoints (`bytes32 txid + uint32 vout`). Once an outpoint is verified via SPV for funding or settlement, it is permanently marked as spent, preventing replay across multiple DLC agreements.
- **IFSCA GIFT City Cross-Border Compliance:** All DLC settlements between Bitcoin collateral and tokenized Indian securities operate in adherence to IFSCA digital asset guidelines. Counterparties must be KYC-verified entities on the Growww Identity Registry (Prompt 305), and all settlement records provide full cryptographic auditability without exposing investor PII.

## Acceptance Criteria
- [ ] `BitcoinSPVBridge.sol` successfully parses and validates 80-byte serialized Bitcoin block headers on Hyperledger Besu.
- [ ] Difficulty retargeting validation correctly verifies 2016-block epoch boundaries and rejects non-compliant difficulty adjustments.
- [ ] Fork resolution correctly selects the chain branch with the highest cumulative proof-of-work ($W_{\text{cumulative}}$).
- [ ] `verifyTransactionInclusion` verifies valid Merkle SPV inclusion proofs against canonical block headers and rejects forged or corrupted paths.
- [ ] `DLCSchnorrVerifier.sol` passes 100% of BIP-340 official test vectors for secp256k1 Schnorr signature verification.
- [ ] `DLCRegistry.sol` executes the full contract lifecycle: Create -> Accept -> SPV Funding Verification -> Schnorr Settlement -> Payout.
- [ ] Attempting to settle a DLC with an invalid oracle signature or unauthorized outcome reverts with `InvalidOracleSignature`.
- [ ] Timelocked refund function executes successfully after `refundTimestamp` expiration and reverts if triggered early.
- [ ] Duplicate SPV funding proofs on the same Bitcoin UTXO are rejected with `FundingUTXOAlreadyUsed`.
- [ ] Go-based `SPVRelayerDaemon` automatically ingests Bitcoin headers, constructs SPV proofs, and submits transactions to Besu via mTLS.
- [ ] Static analysis with Slither and Halmos passes with zero high/critical vulnerabilities.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `306` (Settlement DvP Smart Contract), Prompt `307` (MultiSig Governance Smart Contract).
- **Parallel Tasks:** Prompt `311` (Validator Key Management HSM), Prompt `313` (Cross-Entity Ledger Bridge), Prompt `319` (Cross-Chain Institutional Custody Bridge).
- **Subsequent Prompts Enabled:** Prompt `208` (Trade Settlement & DvP Orchestration Service), Prompt `214` (Foreign Investor Funding & FX Service), Prompt `216` (Regulatory Reporting Service Integration).
