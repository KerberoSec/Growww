# 306 - Atomic Delivery-versus-Payment (DvP) Settlement Smart Contract (SettlementDvP.sol)

## Purpose
In modern securities clearing and settlement, Delivery-versus-Payment (DvP) Model 1 guarantees that the final transfer of digital securities from seller to buyer occurs if and only if the corresponding fiat payment (INR) from buyer to seller is simultaneously and unconditionally completed. Without atomic DvP, market participants face principal risk (counterparty default during the settlement window).

This prompt specifies the production-grade implementation of the **Atomic Delivery-versus-Payment Settlement Smart Contract (`SettlementDvP.sol`)** on Hyperledger Besu, expanding on the reference architecture. The contract coordinates multi-trade atomic batch settlements, verifies fiat banking settlement commitments signed by authorized banking/clearing relayers, transfers fractional equity units across buyer/seller balances, validates and distributes the canonical Fixed 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee on trade notional turnover via a 0.00% fee at launch (governed by FeeController.sol) revenue split, and records cryptographic proof hashes for FIFO capital gains computed strictly for user tax compliance (Section 111A/112A).

## What You Are Building
A high-throughput, secure settlement contract system under `contracts/settlement/` comprising:
- `SettlementDvP.sol`: Core settlement contract executing atomic single and batched trade settlements between verified buyers and sellers.
- Batch Settlement Engine: Optimized calldata packing allowing up to 100 bilateral trades settled atomically in a single Besu transaction.
- Fixed Transaction Fee & Multi-Vault Revenue Splitter: On-chain verification and allocation of the fixed 0.00% fee (No fee at all) on trade notional turnover routed according to the 0.00% fee at launch (governed by FeeController.sol) revenue split.
- Tax Compliance Proof Attestation: On-chain cryptographic leaf hash recording for off-chain FIFO capital gains tax calculations (Section 111A/112A).
- Trade Idempotency & Nonce Registry: Prevents replay attacks, double settlement, or out-of-order trade processing.
- Comprehensive Foundry test suite (`test/settlement/SettlementDvP.t.sol`) validating DvP atomicity, fee calculations, multi-vault splits, and reentrancy resistance.

## Scope Boundaries
- **In Scope:**
 - Atomic swap execution: token transfer from seller to buyer conditional upon valid fiat settlement proof.
 - Verification of EIP-712 trade authorization signatures from buyer, seller, and settlement orchestrator.
 - Multi-trade batching (processing multiple matched orders per transaction).
 - Fixed transaction fee deduction (0.00% (Zero Fee) on turnover) and 0.00% fee at launch (governed by FeeController.sol) revenue routing.
 - Attestation hash storage for tax compliance (Section 111A/112A).
 - Emitting detailed settlement events (`DvPSettlementExecuted`, `BatchSettled`).
- **Out of Scope / Handled Elsewhere:**
 - Off-chain Order Matching Engine (handled in Prompt 205).
 - Banking payment reservation and UPI/IMPS rail execution (handled in Prompt 212).
 - Off-chain Trade Settlement Orchestration Service (handled in Prompt 208).
 - Off-chain FIFO tax lot matching microservice (handled in Prompt 210).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Provides native custom errors, custom user-defined value types, and efficient cryptographic primitives (`ecrecover`, `keccak256`) with transient storage (`TSTORE`/`TLOAD`) to minimize gas overhead during batch loops.
- **Frameworks:** OpenZeppelin Contracts v5.0 (ReentrancyGuardUpgradeable, ECDSA, EIP712Upgradeable, AccessControlUpgradeable).
- **Tooling:** Foundry (`forge`, `cast`), Slither static analyzer.

## Backend / Infra Touchpoints
- **Trade Settlement Service:** Microservice (Prompt 208) that packages matched trades from the Matching Engine (Prompt 205) and calls `SettlementDvP.sol`.
- **Wallet & Account Service:** Fiat ledger service (Prompt 203) that confirms INR bank hold execution before submitting on-chain settlement transactions.
- **Fee Engine:** Service (Prompt 210) computing turnover fees (0.00% fee at launch (future fee parameters governed by FeeController.sol)) and tax compliance proof hashes verified by `SettlementDvP.sol`.
- **Blockchain Event Indexer:** Service (Prompt 309) ingesting `BatchSettled` events for immediate post-trade client balance updates.

## Blockchain Interaction
- **Direct Asset Movement:** `SettlementDvP.sol` calls `IERC20(tokenAddress).transferFrom(seller, buyer, tokenAmount)`.
- **Token Compliance Checks:** Transfers automatically trigger underlying ERC-3643 compliance hooks (Prompt 305) verifying both buyer and seller are active, KYC-verified investors.
- **Atomic All-or-Nothing Batches:** If any individual trade in an atomic batch fails compliance or verification, the entire batch transaction reverts cleanly, preserving ledger state consistency.
- **Zero PII:** Trade payloads contain only `bytes32 tradeId`, `address buyer`, `address seller`, `address tokenAddress`, `uint256 tokenAmount`, `uint256 inrGrossAmount`, `uint256 feeAmount`, `bytes32 taxProofHash`, and signatures.

## Step-by-Step Build Instructions
1. Scaffold settlement directory: `contracts/settlement/SettlementDvP.sol` and `contracts/interfaces/ISettlementDvP.sol`.
2. Define EIP-712 domain separator: `EIP712("GrowwwSettlementDvP", "1")`.
3. Define trade data structs: `struct TradeOrder { bytes32 tradeId; address buyer; address seller; address token; uint256 tokenAmount; uint256 inrGrossAmount; uint256 feeAmount; bytes32 taxProofHash; uint64 matchedTimestamp; uint64 expiryTimestamp; uint256 nonce; }`.
4. Define settlement batch struct: `struct SettlementBatch { bytes32 batchId; TradeOrder[] trades; bytes[] orchestratorSignatures; }`.
5. Implement storage layout with trade execution status: `mapping(bytes32 => bool) public isTradeSettled` and `mapping(address => uint256) public userNonces`.
6. Implement `initialize(address adminMultisig, address settlementRelayer, address treasuryVault, address coreSgfVault, address ipfVault)` with upgradeable access control.
7. Implement `settleTrade(TradeOrder calldata trade, bytes calldata orchestratorSig)`:
 - Check caller has `SETTLEMENT_OPERATOR_ROLE` or verify orchestrator signature.
 - Verify `trade.expiryTimestamp >= block.timestamp`.
 - Verify `!isTradeSettled[trade.tradeId]`.
 - Mark `isTradeSettled[trade.tradeId] = true`.
 - Validate fee computation: require `trade.feeAmount == (trade.inrGrossAmount * feeBps) / 10000 (where feeBps == 0 at launch)` (exact 0.00% (Zero Fee) / 0 bps (0.00% fee at launch) of gross turnover).
 - Execute token delivery: `IERC20(trade.token).transferFrom(trade.seller, trade.buyer, trade.tokenAmount)`.
 - Emit `DvPSettlementExecuted(trade.tradeId, trade.token, trade.buyer, trade.seller, trade.tokenAmount, trade.inrGrossAmount, trade.feeAmount, trade.taxProofHash)`.
8. Implement `settleBatch(SettlementBatch calldata batch)` iterating over `batch.trades` with local memory caching to maximize throughput.
9. Implement emergency settlement pause/unpause restricted to multi-sig governance (Prompt 307).
10. Implement trade cancellation: `cancelTrade(bytes32 tradeId, string calldata reason)` restricted to authorized settlement operators.
11. Write Foundry unit tests testing valid single trade settlement, fee deductions, and event emissions.
12. Write negative tests verifying execution failure on expired timestamp, replay attack with duplicate `tradeId`, invalid signature, and unverified buyer/seller address.
13. Write batch load tests in Foundry simulating batches of 50, 100, and 200 trades per block, measuring gas consumption and throughput on Besu.
14. Perform Slither static analysis and verify zero reentrancy or state inconsistencies.
15. Deploy to Hyperledger Besu devnet, bind with `DigitalSecurityToken.sol`, and execute automated settlement test runs.

## Interfaces / Contracts

### Settlement DvP Interface (`ISettlementDvP.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ISettlementDvP {
    struct TradeOrder {
        bytes32 tradeId;
        address buyer;
        address seller;
        address token;
        uint256 tokenAmount;       // 18 decimals fractional equity units
        uint256 inrGrossAmount;    // 2 decimals (paise precision)
        uint256 feeAmount;         // 2 decimals (0.00% (Zero Fee) of inrGrossAmount)
        bytes32 taxProofHash;      // SHA-256 hash for Section 111A/112A tax compliance
        uint64 matchedTimestamp;
        uint64 expiryTimestamp;
        uint256 nonce;
    }

    struct SettlementBatch {
        bytes32 batchId;
        TradeOrder[] trades;
        bytes[] signatures;
    }

    event DvPSettlementExecuted(
        bytes32 indexed tradeId,
        address indexed token,
        address indexed buyer,
        address seller,
        uint256 tokenAmount,
        uint256 inrGrossAmount,
        uint256 feeAmount,
        bytes32 taxProofHash
    );

    event BatchSettled(
        bytes32 indexed batchId,
        uint256 totalTrades,
        uint256 totalVolumeINR,
        uint256 totalFeeINR
    );

    event TradeCancelled(bytes32 indexed tradeId, string reason);
    event RevenueVaultsUpdated(address indexed treasury, address indexed coreSgf, address indexed ipf);

    // Errors
    error TradeAlreadySettled(bytes32 tradeId);
    error TradeExpired(bytes32 tradeId, uint64 expiryTimestamp);
    error InvalidFeeCalculation(uint256 expectedFee, uint256 providedFee);
    error InvalidTradeSignature(bytes32 tradeId);
    error UnauthorizedOperator(address caller);
    error BatchSizeZero();

    // Settlement Functions
    function settleTrade(TradeOrder calldata trade, bytes calldata signature) external;
    function settleBatch(SettlementBatch calldata batch) external;
    function cancelTrade(bytes32 tradeId, string calldata reason) external;

    // View Functions
    function isTradeSettled(bytes32 tradeId) external view returns (bool);
    function getRevenueVaults() external view returns (address treasury, address coreSgf, address ipf);
}
```

## Security & Compliance Notes
- **Atomicity Invariant:** Either token delivery and fee accounting execute simultaneously, or the entire state rolls back. Under no condition can tokens leave the seller's wallet without the settlement transaction confirming.
- **Fixed Platform Fee & Multi-Vault Allocation Enforcement:** In accordance with Growww's canonical fee policy, platform fees are assessed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover and routed per FeeController governance (0.00% at launch). Capital gains calculations are computed strictly for off-chain user tax compliance (Section 111A/112A) and anchored via `taxProofHash`. The contract verifies on-chain that `feeAmount == (inrGrossAmount * feeBps) / 10000 (where feeBps == 0 at launch)`.
- **Replay Protection:** Every `tradeId` is permanently recorded in storage. Any secondary attempt to submit an identical `tradeId` reverts immediately with `TradeAlreadySettled`.
- **mTLS & Relayer Authorization:** Only authenticated relayers possessing the `SETTLEMENT_OPERATOR_ROLE` backed by HSM keys (Prompt 311) can submit settlement batches to the node.

## Acceptance Criteria
- [ ] `SettlementDvP.sol` contract deployed, verified, and integrated on Hyperledger Besu.
- [ ] Single trade DvP settlement successfully executes token transfer from seller to buyer upon signature verification.
- [ ] Batch settlement executes up to 100 trades in a single transaction in < 1.5 seconds.
- [ ] Fixed 0.00% (Zero Fee) turnover fee logic tested: fee is correctly verified at 0 bps (0.00% fee at launch) of gross trade consideration.
- [ ] 0.00% fee at launch (governed by FeeController.sol) revenue split and tax compliance proof hashes are recorded.
- [ ] Duplicate trade submission reverts with `TradeAlreadySettled`.
- [ ] Slither / Mythril security scans pass with zero high/medium warnings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance), Prompt `305` (Transfer Compliance Hooks), Prompt `006` (Fee Model Specification), Prompt `208` (Trade Settlement Service).
- **Parallel Tasks:** Prompt `210` (Fee Engine), Prompt `309` (Event Indexing Service).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (Proof of Reserve).
