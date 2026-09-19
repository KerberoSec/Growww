# 346 - P2P Escrow & 2-of-3 Multi-Sig Arbitration Smart Contract (Solidity)

## Purpose
Direct peer-to-peer (P2P) fiat-to-crypto and fiat-to-tokenized-asset settlement on the Growww National Blockchain Stock Exchange (NBSE) requires trustless collateral locking to prevent seller default and automated dispute resolution workflows to protect buyers against fraudulent counterparty claims. In standard OTC or P2P trading, transacting parties face principal risk: a buyer may send off-chain fiat (via UPI, IMPS, RTGS, or Open Banking) and the seller may refuse to release the crypto, or a fraudulent buyer may fabricate payment confirmations without transferring actual bank funds. 

Under ADR-0039 (P2P Automated Escrow Lockbox & Dispute Arbitration) and RUNBOOK-27 (P2P Fiat Dispute Maker-Checker Arbitration), Growww eliminates counterparty credit risk and manual operational bottlenecks by deploying a decentralized 2-of-3 threshold multi-signature escrow smart contract. Escrow deposits are programmatically secured on the permissioned Hyperledger Besu consortium ledger. Custody is non-custodial and conditional: funds can only be released or refunded with the valid cryptographic co-signatures of any 2 out of the 3 designated entities:
1. The Buyer (confirming fiat payment dispatched and receipt requested);
2. The Seller (confirming fiat receipt into their KYC-verified bank account);
3. The Exchange Compliance Officer / Arbitrator (adjudicating disputed transactions backed by FIPS 140-2 Level 3 CloudHSM signing keys and automated bank UTR clearance proofs).

This prompt specifies the enterprise-grade implementation of the **P2P Escrow & 2-of-3 Multi-Sig Arbitration Smart Contract (`P2PEscrow.sol`, `IP2PEscrow.sol`)**. The contract orchestrates the complete P2P trade lifecycle: locking collateral tokens upon order matching, enforcing dynamic timelocks for payment windows, recording cryptographic pre-image receipts (bank UTR payment leaf hashes), managing dispute freezes, and executing deterministic fund releases or refunds under 2-of-3 EIP-712 threshold authorizations, while strictly collecting Growww's canonical 0.00% (Zero Fee) platform fee (0.00% fee at launch; future fee parameters governed by FeeController.sol).

## What You Are Building
A production-ready, upgradeable Solidity smart contract suite located under `contracts/src/settlement/` and `contracts/interfaces/settlement/` comprising:
- `P2PEscrow.sol`: Enterprise-grade Solidity ^0.8.24 escrow engine deployed on Hyperledger Besu. It executes time-locked deposits, buyer payment marking with hashed bank reference receipts (UTR hashes), dispute freezing, seller confirmations, mutual cancellations, 2-of-3 multi-signature arbitrated dispute settlements, and automated fee splits. It natively handles ERC-20 stablecoins (USDT, USDC), sovereign digital rupee (`eINR`), and ERC-3643 permissioned security tokens.
- `IP2PEscrow.sol`: Complete Solidity interface specifying state enums, order structs, 2-of-3 multi-signature resolution payloads, fee breakdown records, custom errors, events, and all mutator and view function signatures.
- Cryptographic 2-of-3 Multi-Signature Verifier: Native EVM ECDSA signature verification compliant with EIP-712 structured data hashing. Verifies that any 2 distinct authorized signers (Buyer + Seller, Buyer + Compliance, or Seller + Compliance) authorize fund disbursement, rejecting single-signature tampering, unauthorized third-party keys, and signature replay attacks.
- Time-Locked Deposit & Griefing Prevention Engine: Dual-phase timelock tracking:
  1. $T_{\text{payment}}$ (Default 15 minutes): Window for buyer to transfer fiat and submit on-chain payment proof. If expired without buyer action, seller can unilaterally cancel and reclaim escrowed collateral.
  2. $T_{\text{confirm}}$ (Default 30 minutes): Window for seller to confirm bank receipt after buyer marks payment. If seller fails to act, buyer can escalate to dispute freeze.
  3. $T_{\text{dispute\_max}}$ (Default 7 days): Safety ceiling preventing permanent capital lockup during protracted disputes, allowing multi-sig governance override if compliance fails to arbitrate.
- Cryptographic Pre-Image Receipt Registry: Stores and verifies SHA-256 / Keccak-256 hashes of the off-chain banking Unique Transaction Reference (UTR), transaction timestamp, and payment receipt payload, creating an immutable cryptographic bridge between Indian Open Banking rails and on-chain settlement.
- Canonical Universal Zero-Fee Model (0.00% fee - No fee at all) & Multi-Vault Revenue Splitter: Automated deduction of Growww's 0.00% fee (No fee at all) on gross trade turnover upon successful order release, with dynamic routing governed by FeeController.sol (0.00% at launch) vaults, while waiving 100% of fees on cancelled or refunded trades.
- Comprehensive Foundry Test Suite (`test/settlement/P2PEscrow.t.sol`): Property-based fuzz tests, state machine transition tests, cryptographic signature recovery validations, timelock boundary checks, and reentrancy attack simulations.

## Scope Boundaries
- **In Scope:**
  - Solidity ^0.8.24 implementation of `P2PEscrow.sol` and `IP2PEscrow.sol` on Hyperledger Besu consortium network.
  - 2-of-3 threshold signature verification for dispute resolution and order settlement using OpenZeppelin `ECDSA` and `MessageHashUtils` with EIP-712 domain separation.
  - Multi-token support: Standard ERC-20 tokens, wrapped `eINR` (RBI CBDC bridge), and ERC-3643 permissioned real-world asset tokens.
  - State machine lifecycle management: `CREATED`, `PAYMENT_MARKED`, `COMPLETED`, `DISPUTED`, `REFUNDED`, and `CANCELLED`.
  - Timelock management: Payment window enforcement, auto-cancellation eligibility, dispute freeze protection, and anti-griefing limits.
  - Cryptographic UTR proof hashing binding off-chain fiat settlement to on-chain escrow records without exposing plain-text PII.
  - Automated Growww 0.00% (Zero Fee) platform fee calculation and deterministic 0.00% fee at launch (governed by FeeController.sol) distribution.
  - OpenZeppelin v5.0 UUPS upgradeability (`UUPSUpgradeable`), reentrancy protection (`ReentrancyGuardUpgradeable`), circuit breaker pause (`PausableUpgradeable`), and role-based access control (`AccessControlUpgradeable`).
- **Out of Scope / Handled Elsewhere:**
  - Off-chain P2P order matching, maker order-books, and taker quote matching (handled in Prompt 269 / P2P Escrow Service).
  - Off-chain Indian Open Banking API integration, Account Aggregator (AA) queries, and UTR clearance verification (handled in Prompt 269 and Prompt 212).
  - Compliance officer administrative review UI and Maker-Checker arbitration workflow management (handled in Prompt 217 / Admin Back-Office and Prompt 605).
  - Cryptographic CloudHSM arbitration key generation, PKCS#11 hardware isolation, and signing daemon (handled in Prompt 717).
  - Real-time investor KYC verification and AML sanctions screening (handled in Prompt 202 and Prompt 703).
  - User mobile and web P2P trading interface screens (handled in Prompt 521 / Flutter P2P Escrow Flow).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Optimized for EVM Cancun/Shanghai with checked arithmetic, custom errors for minimal bytecode footprint, and transient storage `TSTORE`/`TLOAD` support where available).
- **Cryptographic Primitives:**
  - OpenZeppelin `ECDSA` library for elliptic curve signature recovery (`secp256k1`).
  - OpenZeppelin `MessageHashUtils` for EIP-712 typed structured data digest generation (`toTypedDataHash`).
  - Keccak-256 and SHA-256 native EVM hash functions for order hashing and banking UTR receipt verification.
- **Frameworks & Core Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0:
    - `Initializable`: Safe proxy initialization without constructor state.
    - `UUPSUpgradeable`: Universal Upgradeable Proxy Standard with authorized upgrade gates.
    - `AccessControlUpgradeable`: Role-based access control for admin, operator, and compliance arbitrator roles.
    - `ReentrancyGuardUpgradeable`: Non-reentrant execution guards on all state-modifying token transfer operations.
    - `PausableUpgradeable`: Emergency circuit breaker halting new escrow creations during critical operational events.
    - `SafeERC20`: Non-standard ERC-20 token transfer handling (reverting on failure and handling boolean/void returns).
- **Testing & Verification Toolchain:**
  - **Foundry (`forge`, `cast`):** Unit testing, cryptographic property testing, and differential fuzzing.
  - **Static Security Analysis:** Slither and Mythril linters integrated into CI/CD build gates.

## Backend / Infra Touchpoints
- **P2P Escrow Service (Prompt 269):**
  - High-throughput Go/gRPC backend coordinating off-chain order creation, fiat timer tracking, and payment notification dispatches.
  - Submits on-chain `createEscrow()` transactions as relayer or orchestrates direct seller deposits.
  - Submits `markPaymentSent()` on behalf of buyers with the hashed Open Banking UTR proof.
  - Ingests on-chain events (`EscrowCreated`, `PaymentMarked`, `DisputeOpened`, `EscrowReleased`, `EscrowRefunded`) via Kafka to update order states in PostgreSQL (`p2p_orders`, `p2p_disputes`).
- **Admin & Back-Office Service (Prompt 217):**
  - Provides the Maker-Checker operational arbitration console used by exchange dispute officers (RUNBOOK-27).
  - Ingests buyer and seller evidence, verifies bank statement UTR entries, and coordinates Maker approval and Checker authorization.
  - Calls CloudHSM Signer (Prompt 717) to produce the compliance officer's cryptographic EIP-712 signature upon 2-of-3 dispute resolution.
- **CloudHSM Key Lifecycle & Signing Daemon (Prompt 717):**
  - FIPS 140-2 Level 3 hardware security module holding the private keys for `ARBITRATOR_ROLE`.
  - Signs the EIP-712 `ArbitrationResolution` digest only after verifying Maker-Checker dual-approval authorizations.
- **Settlement Guarantee Fund & Treasury Vaults (Prompt 230 / Prompt 315):**
  - Receives allocations per FeeController governance from the 0.00% (Zero Fee) platform fee upon successful trade completion.
- **Investor Protection Fund (IPF) Vault (Prompt 315):**
  - Receives the Investor Protection Fund fee allocation to capitalize investor dispute and counterparty default guarantees.
- **Blockchain Event Indexer (Prompt 309):**
  - Subscribes to Hyperledger Besu logs, indexing escrow state changes to Redis cache keys (`p2p:escrow:<orderId>`) for sub-second mobile UI status synchronization.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Immediate Finality:** Deployed on the Growww NBSE Hyperledger Besu consortium network operating QBFT consensus with 2-second deterministic block times. Zero chain reorganization risk guarantees that once an escrow is locked, completed, or arbitrated, the settlement is final and irreversible.
- **Trustless Non-Custodial Collateral Locking:** When a seller initiates or accepts a P2P sell order, the escrow contract pulls the required token amount into its contract address (`contracts/src/settlement/P2PEscrow.sol`). Neither party nor the exchange operator can unilaterally withdraw funds while the escrow is active.
- **2-of-3 Threshold Cryptographic Signatures:**
  Dispute resolution and emergency disbursements require any 2 distinct signatures out of the 3 designated keys:
  $$\text{Signers} = \{\text{Buyer}, \text{Seller}, \text{Compliance Arbitrator}\}$$
  The contract checks:
  1. Both signatures recover to distinct addresses in $\text{Signers}$;
  2. The signatures authorize an identical `orderId`, `beneficiary` (Buyer or Seller), `payoutAmountPaise`, and `nonce`;
  3. The signature digest adheres strictly to the EIP-712 domain separator:
     $$\text{Domain} = \text{keccak256}(\text{"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"})$$
     $$\text{TypeHash} = \text{keccak256}(\text{"ArbitrationResolution(bytes32 orderId,address beneficiary,uint256 amount,uint256 feePaise,uint256 nonce,uint64 expiry)"})$$
- **Zero On-Chain PII Invariant:** In accordance with DPDP Act 2023, SEBI cybersecurity guidelines, and ADR-0039, no personally identifiable information (investor legal names, PAN, Aadhaar, bank account numbers, IFSC codes, mobile numbers) is ever stored or logged on-chain. Bank settlements are referenced purely via `bytes32 utrHash` (Keccak-256 hash of the bank UTR combined with order salt).
- **ERC-3643 Permissioned Asset Compatibility:** When locking tokenized real-world assets or securities, the escrow contract checks the asset identity registry to ensure buyer and seller hold active, non-sanctioned ONCHAINID investor claims.
- **Platform Fee Distribution Mechanics:**
  On successful completion, the contract calculates:
  $$\text{totalFee} = (\text{escrow.amount} \times \text{feeBps}) / 10000 \quad (0.00\% \text{ / 0 bps at launch})$$
  $$\text{treasuryShare} = (\text{totalFee} \times 60) / 100$$
  $$\text{coreSgfShare} = (\text{totalFee} \times 25) / 100$$
  $$\text{ipfShare} = \text{totalFee} - \text{treasuryShare} - \text{coreSgfShare} \quad (15\% \text{ exact without rounding loss})$$
  $$\text{netPayout} = \text{escrow.amount} - \text{totalFee}$$

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout & Tooling:**
   - Initialize directories `contracts/src/settlement/`, `contracts/interfaces/settlement/`, and `test/settlement/`.
   - Verify Foundry configuration (`foundry.toml`) specifies EVM version `cancun`, optimizer runs `200`, and remappings for OpenZeppelin Contracts Upgradeable v5.0.
2. **Define State Enums, Structs, and Events (`IP2PEscrow.sol`):**
   - Define enum `EscrowStatus`: `UNINITIALIZED`, `LOCKED`, `PAYMENT_MARKED`, `DISPUTED`, `COMPLETED`, `REFUNDED`, `CANCELLED`.
   - Define enum `ArbitrationOutcome`: `NONE`, `RELEASE_TO_BUYER`, `REFUND_TO_SELLER`, `SPLIT_SETTLEMENT`.
   - Define structs: `EscrowOrder`, `OrderTimelocks`, `FeeAllocation`, `ArbitrationPayload`.
   - Declare custom errors: `OrderAlreadyExists`, `OrderDoesNotExist`, `InvalidStatus`, `PaymentWindowExpired`, `PaymentWindowActive`, `UnauthorizedCaller`, `InvalidSignerCount`, `DuplicateSignatures`, `InvalidThresholdSigners`, `SignatureExpired`, `ZeroAddressNotAllowed`, `ZeroAmountNotAllowed`, `FeeCalculationOverflow`.
   - Declare events: `EscrowCreated`, `PaymentMarked`, `DisputeOpened`, `EscrowCompleted`, `EscrowRefunded`, `EscrowCancelled`, `DisputeResolved`, `FeeDistributed`, `TimelockConfigUpdated`.
3. **Draft Base Smart Contract & Storage Architecture (`P2PEscrow.sol`):**
   - Inherit `Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`.
   - Define roles: `DEFAULT_ADMIN_ROLE`, `OPERATOR_ROLE`, `ARBITRATOR_ROLE`, `EMERGENCY_ROLE`.
   - Configure EIP-712 domain separator: name `"GrowwwP2PEscrow"`, version `"1"`.
   - Declare persistent state mappings:
     - `mapping(bytes32 => EscrowOrder) private _orders;`
     - `mapping(bytes32 => uint256) private _orderNonces;`
     - `mapping(address => bool) private _authorizedTokens;`
   - Declare vault addresses: `treasuryVault`, `coreSgfVault`, `ipfVault`.
   - Declare configurable operational parameters: `minPaymentWindow` (900 seconds / 15 min), `maxPaymentWindow` (7200 seconds / 2 hrs), `maxDisputeDuration` (604800 seconds / 7 days).
4. **Implement Contract Initializer:**
   - Implement `initialize(address admin, address operator, address arbitrator, address treasury, address coreSgf, address ipf)` with `initializer` modifier.
   - Enforce non-zero address validation across all constructor arguments.
   - Assign initial roles and configure default payment/dispute timelock windows.
5. **Implement Escrow Deposit & Lock Creation (`createEscrow`):**
   - Accept parameters: `orderId`, `buyer`, `token`, `amount`, `fiatPaise`, `paymentWindowSeconds`.
   - Validate `_orders[orderId].status == EscrowStatus.UNINITIALIZED`.
   - Validate `msg.sender != buyer`, `buyer != address(0)`, `amount > 0`.
   - Validate `paymentWindowSeconds` between `minPaymentWindow` and `maxPaymentWindow`.
   - Compute payment deadline: $T_{\text{payment}} = \text{block.timestamp} + \text{paymentWindowSeconds}$.
   - Pull tokens into escrow from `msg.sender` (Seller) via `SafeERC20.safeTransferFrom(msg.sender, address(this), amount)`.
   - Record `EscrowOrder` with status `LOCKED` and emit `EscrowCreated`.
6. **Implement Buyer Payment Marking (`markPaymentSent`):**
   - Accept parameters: `orderId`, `bytes32 utrHash`.
   - Validate caller is `order.buyer` or authorized `OPERATOR_ROLE` relayer.
   - Validate order status is `LOCKED`.
   - Enforce payment window invariant: require $\text{block.timestamp} \le order.timelocks.paymentDeadline$, reverting with `PaymentWindowExpired` if elapsed.
   - Store `order.utrHash = utrHash` and transition status to `PAYMENT_MARKED`.
   - Calculate confirmation deadline: $T_{\text{confirm}} = \text{block.timestamp} + 1800\text{ seconds}$ (30 minutes).
   - Emit `PaymentMarked(orderId, msg.sender, utrHash, order.timelocks.confirmDeadline)`.
7. **Implement Normal Settlement Path (`confirmPaymentAndRelease`):**
   - Accept parameter: `orderId`.
   - Validate caller is `order.seller`.
   - Validate order status is `PAYMENT_MARKED`.
   - Execute Checks-Effects-Interactions: transition status to `COMPLETED`.
   - Calculate 0.00% (Zero Fee) platform fee and multi-vault split.
   - Transfer net tokens to `order.buyer`.
   - Transfer fee portions to Treasury, Core SGF, and IPF vaults.
   - Emit `EscrowCompleted` and `FeeDistributed`.
8. **Implement Dispute Opening & Collateral Freezing (`openDispute`):**
   - Accept parameters: `orderId`, `bytes32 disputeReasonHash`.
   - Validate caller is either `order.buyer` or `order.seller`.
   - If caller is Buyer: order must be in `PAYMENT_MARKED` status and seller has failed to release.
   - If caller is Seller: order must be in `PAYMENT_MARKED` status (e.g. buyer provided invalid/uncredited UTR) or in `LOCKED` status if buyer falsely claimed action.
   - Transition order status to `DISPUTED`.
   - Calculate dispute expiration deadline: $T_{\text{dispute}} = \text{block.timestamp} + maxDisputeDuration$.
   - Emit `DisputeOpened(orderId, msg.sender, disputeReasonHash, order.timelocks.disputeDeadline)`.
9. **Implement 2-of-3 Multi-Signature Dispute Arbitration (`resolveDisputeWith2of3`):**
   - Accept `orderId`, `ArbitrationPayload` (outcome, beneficiary, payoutAmount, feePaise, nonce, expiry), and `bytes[] signatures` (two 65-byte ECDSA signatures).
   - Verify order status is `DISPUTED`.
   - Verify payload expiry: require $\text{block.timestamp} \le payload.expiry$.
   - Verify payload nonce matches and increment `_orderNonces[orderId]`.
   - Construct EIP-712 typed data digest.
   - Recover distinct signer addresses using `ECDSA.recover`.
   - Verify that exactly 2 valid, distinct signatures belong to the set $\{\text{order.buyer}, \text{order.seller}, \text{arbitratorAddress}\}$, where `arbitratorAddress` possesses `ARBITRATOR_ROLE`.
   - Execute state transition to `COMPLETED` (if beneficiary is Buyer) or `REFUNDED` (if beneficiary is Seller).
   - If split settlement: calculate split allocations, deduct 0.00% fee (No fee at all) on buyer portion, and disburse tokens accordingly.
   - Disburse platform fee to Treasury, SGF, and IPF vaults.
   - Emit `DisputeResolved` and corresponding settlement events.
10. **Implement Unilateral Timelock Expiry Refund (`refundExpiredPaymentWindow`):**
    - Accept parameter: `orderId`.
    - Validate caller is `order.seller`.
    - Validate order status is `LOCKED` (buyer never marked payment).
    - Enforce timelock condition: require $\text{block.timestamp} > order.timelocks.paymentDeadline$.
    - Transition status to `REFUNDED`.
    - Return 100% of escrowed tokens to `order.seller` with zero platform fee deduction.
    - Emit `EscrowRefunded(orderId, order.seller, order.amount, "PaymentWindowExpired")`.
11. **Implement Mutual Order Cancellation (`mutualCancel`):**
    - Accept parameter: `orderId` and signatures from both Buyer and Seller authorizing cancellation.
    - Status can be `LOCKED`, `PAYMENT_MARKED`, or `DISPUTED`.
    - Verify both Buyer and Seller co-signed the cancellation authorization digest.
    - Transition status to `CANCELLED`.
    - Return 100% of escrowed collateral to Seller with zero fee deduction.
    - Emit `EscrowCancelled(orderId, order.seller)`.
12. **Implement View, Query, and Inspection Methods:**
    - Implement `getEscrowOrder(bytes32 orderId)` returning complete `EscrowOrder` struct.
    - Implement `getOrderStatus(bytes32 orderId)` returning status enum.
    - Implement `getTimelockStatus(bytes32 orderId)` returning remaining seconds for payment, confirmation, or dispute windows.
    - Implement `computeFeeBreakdown(uint256 amount)` returning Treasury, SGF, and IPF fee shares.
    - Implement `hashArbitrationPayload(ArbitrationPayload memory payload)` returning the EIP-712 digest.
13. **Implement Administrative & Governance Functions:**
    - Implement `setAuthorizedToken(address token, bool authorized)` restricted to `DEFAULT_ADMIN_ROLE`.
    - Implement `setFeeVaults(address treasury, address coreSgf, address ipf)` restricted to `DEFAULT_ADMIN_ROLE`.
    - Implement `updateTimelockParameters(uint64 minPay, uint64 maxPay, uint64 maxDispute)` with parameter sanity bounds.
    - Implement `pause()` and `unpause()` circuit breakers restricted to `EMERGENCY_ROLE`.
    - Implement `_authorizeUpgrade(address newImplementation)` restricted to multi-sig governance (`DEFAULT_ADMIN_ROLE`).
14. **Construct Comprehensive Foundry Unit & Invariant Test Suite:**
    - Unit test normal trade lifecycle: create escrow -> mark payment -> seller confirms -> net release & 0.00% fee (No fee at all) split.
    - Unit test seller expiry refund: create escrow -> allow 15-minute timer to elapse -> seller reclaims tokens cleanly.
    - Unit test dispute resolution combinations:
      - Buyer + Compliance Officer signs -> release to Buyer.
      - Seller + Compliance Officer signs -> refund to Seller.
      - Buyer + Seller signs -> mutually resolves dispute without compliance officer.
    - Negative test: Single signature submitted -> reverts with `InvalidSignerCount`.
    - Negative test: Two signatures from identical party -> reverts with `DuplicateSignatures`.
    - Negative test: Unauthorized third-party signature -> reverts with `InvalidThresholdSigners`.
    - Fuzz test fee split: Assert `treasury + coreSgf + ipf == totalFee` for all amounts $1 \le \text{amount} \le 10^{30}$.
    - Invariant test: Contract token balance must always equal sum of active balances in `LOCKED`, `PAYMENT_MARKED`, and `DISPUTED` orders.
15. **Execute Security Linting, Gas Profiling & Formal Verification:**
    - Execute Slither static analysis and verify zero high, medium, or reentrancy warnings.
    - Profile gas costs ensuring `createEscrow`, `confirmPaymentAndRelease`, and `resolveDisputeWith2of3` execute well within Besu gas thresholds.

## Interfaces / Contracts

### 1. P2P Escrow Interface (`contracts/interfaces/settlement/IP2PEscrow.sol`)

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IP2PEscrow
 * @notice Interface for the Peer-to-Peer (P2P) Automated Escrow and 2-of-3 Multi-Signature Arbitration Engine on Hyperledger Besu.
 * @dev Manages time-locked deposits, cryptographic bank UTR receipts, dispute freezing, and 2-of-3 threshold signature arbitration (ADR-0039, RUNBOOK-27).
 */
interface IP2PEscrow {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum EscrowStatus {
        UNINITIALIZED,
        LOCKED,
        PAYMENT_MARKED,
        DISPUTED,
        COMPLETED,
        REFUNDED,
        CANCELLED
    }

    enum ArbitrationOutcome {
        NONE,
        RELEASE_TO_BUYER,
        REFUND_TO_SELLER,
        SPLIT_SETTLEMENT
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct OrderTimelocks {
        uint64 createdAt;          // Block timestamp when escrow deposit was locked
        uint64 paymentDeadline;    // T_payment: Deadline for buyer to mark payment sent
        uint64 confirmDeadline;    // T_confirm: Deadline for seller to confirm receipt before buyer dispute
        uint64 disputeDeadline;    // T_dispute_max: Maximum time an order can remain in dispute freeze
    }

    struct FeeAllocation {
        uint256 totalFeeTokens;    // Total 0.00% (Zero Fee) platform fee in token units
        uint256 treasuryTokens;    // Governed by FeeController (0.00% at launch)
        uint256 coreSgfTokens;     // Governed by FeeController (0.00% at launch)
        uint256 ipfTokens;         // Governed by FeeController (0.00% at launch)
    }

    struct EscrowOrder {
        bytes32 orderId;           // Unique P2P trade identifier (keccak256 order hash)
        address seller;            // Depositor of crypto collateral
        address buyer;             // Designated recipient upon payment confirmation
        address token;             // Escrowed ERC-20 / ERC-3643 / eINR token address
        uint256 amount;            // Quantity of tokens held in escrow
        uint256 fiatPaise;         // Off-chain fiat trade valuation in paise (1 INR = 100 paise)
        bytes32 utrHash;           // Cryptographic hash of the bank UTR / payment proof
        OrderTimelocks timelocks;  // Multi-stage expiration and dispute milestones
        EscrowStatus status;       // Current state machine lifecycle status
    }

    struct ArbitrationPayload {
        bytes32 orderId;           // Trade identifier being arbitrated
        ArbitrationOutcome outcome;// Resolution decision
        address beneficiary;       // Recipient of arbitrated funds (Buyer, Seller, or primary party)
        uint256 buyerTokens;       // Tokens awarded to buyer (net of fee)
        uint256 sellerTokens;      // Tokens refunded to seller
        uint256 feePaise;          // Assessed fee valuation in paise
        uint256 nonce;             // Anti-replay nonce for order arbitration
        uint64 expiry;             // Cryptographic expiration timestamp for the resolution payload
    }

    struct CancellationPayload {
        bytes32 orderId;           // Trade identifier being cancelled
        uint256 nonce;             // Anti-replay nonce
        uint64 expiry;             // Expiration timestamp
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event EscrowCreated(
        bytes32 indexed orderId,
        address indexed seller,
        address indexed buyer,
        address token,
        uint256 amount,
        uint256 fiatPaise,
        uint64 paymentDeadline
    );

    event PaymentMarked(
        bytes32 indexed orderId,
        address indexed buyer,
        bytes32 utrHash,
        uint64 confirmDeadline
    );

    event DisputeOpened(
        bytes32 indexed orderId,
        address indexed initiator,
        bytes32 disputeReasonHash,
        uint64 disputeDeadline
    );

    event EscrowCompleted(
        bytes32 indexed orderId,
        address indexed buyer,
        uint256 netAmount,
        uint256 totalFee
    );

    event EscrowRefunded(
        bytes32 indexed orderId,
        address indexed seller,
        uint256 refundedAmount,
        bytes32 reason
    );

    event EscrowCancelled(
        bytes32 indexed orderId,
        address indexed seller,
        address initiator
    );

    event DisputeResolved(
        bytes32 indexed orderId,
        ArbitrationOutcome outcome,
        address indexed beneficiary,
        address signer1,
        address signer2,
        uint256 buyerAmount,
        uint256 sellerAmount
    );

    event PlatformFeeDistributed(
        bytes32 indexed orderId,
        address token,
        uint256 totalFee,
        uint256 treasuryAmount,
        uint256 coreSgfAmount,
        uint256 ipfAmount
    );

    event TimelockConfigUpdated(
        uint64 minPaymentWindow,
        uint64 maxPaymentWindow,
        uint64 maxDisputeDuration
    );

    event FeeVaultsUpdated(
        address treasuryVault,
        address coreSgfVault,
        address ipfVault
    );

    event TokenAuthorizationSet(
        address indexed token,
        bool authorized
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error OrderAlreadyExists(bytes32 orderId);
    error OrderDoesNotExist(bytes32 orderId);
    error InvalidStatus(bytes32 orderId, EscrowStatus current, EscrowStatus expected);
    error PaymentWindowExpired(bytes32 orderId, uint64 deadline, uint64 currentTimestamp);
    error PaymentWindowActive(bytes32 orderId, uint64 deadline, uint64 currentTimestamp);
    error ConfirmWindowActive(bytes32 orderId, uint64 deadline, uint64 currentTimestamp);
    error DisputeDeadlineExceeded(bytes32 orderId, uint64 deadline, uint64 currentTimestamp);
    error UnauthorizedCaller(address caller);
    error InvalidSignerCount(uint256 provided, uint256 requiredCount);
    error DuplicateSignatures(address duplicateSigner);
    error InvalidThresholdSigners(address signer1, address signer2);
    error SignatureExpired(uint64 expiry, uint64 currentTimestamp);
    error InvalidNonce(bytes32 orderId, uint256 provided, uint256 expected);
    error TokenNotAuthorized(address token);
    error ZeroAddressNotAllowed();
    error ZeroAmountNotAllowed();
    error InvalidWindowParameter(uint64 provided, uint64 minAllowed, uint64 maxAllowed);
    error DisputedOrderLocked(bytes32 orderId);

    // -------------------------------------------------------------------------
    // State Modifying Functions
    // -------------------------------------------------------------------------

    /**
     * @notice Locks seller collateral tokens into P2P escrow upon order match.
     * @param orderId Unique identifier for the P2P order.
     * @param buyer Designated buyer address eligible to receive the crypto.
     * @param token Address of the ERC-20 / ERC-3643 / eINR token being sold.
     * @param amount Token quantity to escrow.
     * @param fiatPaise Total trade consideration in INR paise.
     * @param paymentWindowSeconds Window in seconds for buyer to complete payment (e.g. 900s = 15m).
     */
    function createEscrow(
        bytes32 orderId,
        address buyer,
        address token,
        uint256 amount,
        uint256 fiatPaise,
        uint64 paymentWindowSeconds
    ) external;

    /**
     * @notice Marks that the buyer has dispatched fiat funds and submits cryptographic UTR hash.
     * @param orderId The unique P2P order identifier.
     * @param utrHash SHA-256 or Keccak-256 digest of the banking UTR and payment receipt.
     */
    function markPaymentSent(
        bytes32 orderId,
        bytes32 utrHash
    ) external;

    /**
     * @notice Confirms fiat receipt and releases crypto tokens to the buyer.
     * @dev Callable by the seller under normal non-disputed settlement.
     * @param orderId The unique P2P order identifier.
     */
    function confirmPaymentAndRelease(
        bytes32 orderId
    ) external;

    /**
     * @notice Freezes the escrow into dispute state when payment or confirmation is contested.
     * @dev Callable by either buyer or seller within permissible state windows.
     * @param orderId The unique P2P order identifier.
     * @param disputeReasonHash Keccak-256 hash of the off-chain dispute claim metadata.
     */
    function openDispute(
        bytes32 orderId,
        bytes32 disputeReasonHash
    ) external;

    /**
     * @notice Adjudicates a disputed escrow using 2-of-3 threshold co-signatures.
     * @dev Validates that any 2 distinct authorized keys (Buyer, Seller, Compliance Arbitrator) co-signed the payload.
     * @param orderId The unique P2P order identifier.
     * @param payload Structured resolution terms (outcome, beneficiary, token amounts, nonce, expiry).
     * @param signatures Array containing exactly two 65-byte ECDSA signatures (r, s, v).
     */
    function resolveDisputeWith2of3(
        bytes32 orderId,
        ArbitrationPayload calldata payload,
        bytes[] calldata signatures
    ) external;

    /**
     * @notice Reclaims seller collateral if the buyer failed to mark payment within the allotted payment window.
     * @dev Callable by seller after paymentDeadline expires while status is LOCKED.
     * @param orderId The unique P2P order identifier.
     */
    function refundExpiredPaymentWindow(
        bytes32 orderId
    ) external;

    /**
     * @notice Mutually cancels an active order with explicit co-signatures from both buyer and seller.
     * @param orderId The unique P2P order identifier.
     * @param payload Cancellation terms and anti-replay nonce.
     * @param buyerSignature Valid ECDSA signature from the order buyer.
     * @param sellerSignature Valid ECDSA signature from the order seller.
     */
    function mutualCancel(
        bytes32 orderId,
        CancellationPayload calldata payload,
        bytes calldata buyerSignature,
        bytes calldata sellerSignature
    ) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function getEscrowOrder(bytes32 orderId) external view returns (EscrowOrder memory);
    function getOrderStatus(bytes32 orderId) external view returns (EscrowStatus);
    function getOrderNonce(bytes32 orderId) external view returns (uint256);
    function computeFeeBreakdown(uint256 tokenAmount) external pure returns (FeeAllocation memory fees);
    function hashArbitrationPayload(ArbitrationPayload calldata payload) external view returns (bytes32);
    function isTokenAuthorized(address token) external view returns (bool);
    function getFeeVaults() external view returns (address treasury, address coreSgf, address ipf);
    function getTimelockBounds() external view returns (uint64 minPay, uint64 maxPay, uint64 maxDispute);
}
```

### 2. Implementation Specification Contract Overview (`contracts/src/settlement/P2PEscrow.sol`)

The `P2PEscrow.sol` contract implements `IP2PEscrow` along with the following architecture:
- **Inherited OpenZeppelin Modules:**
  - `Initializable`
  - `UUPSUpgradeable`
  - `AccessControlUpgradeable`
  - `ReentrancyGuardUpgradeable`
  - `PausableUpgradeable`
  - `EIP712Upgradeable`
- **Role Hierarchy:**
  - `bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");` (Assigned to P2P Escrow microservice for relaying payment marks and order setup).
  - `bytes32 public constant ARBITRATOR_ROLE = keccak256("ARBITRATOR_ROLE");` (Assigned to the CloudHSM signing key used by compliance officers in Prompt 717).
  - `bytes32 public constant EMERGENCY_ROLE = keccak256("EMERGENCY_ROLE");` (Circuit breaker pause/unpause).
- **Core Storage Variables:**
  - `mapping(bytes32 => EscrowOrder) private _orders;`
  - `mapping(bytes32 => uint256) private _orderNonces;`
  - `mapping(address => bool) private _authorizedTokens;`
  - `address public treasuryVault;`
  - `address public coreSgfVault;`
  - `address public ipfVault;`
  - `uint64 public minPaymentWindow;` (Default: 900 seconds / 15 minutes).
  - `uint64 public maxPaymentWindow;` (Default: 7,200 seconds / 2 hours).
  - `uint64 public maxDisputeDuration;` (Default: 604,800 seconds / 7 days).
- **EIP-712 Type Hashes:**
  - `ARBITRATION_TYPEHASH = keccak256("ArbitrationPayload(bytes32 orderId,uint8 outcome,address beneficiary,uint256 buyerTokens,uint256 sellerTokens,uint256 feePaise,uint256 nonce,uint64 expiry)");`
  - `CANCELLATION_TYPEHASH = keccak256("CancellationPayload(bytes32 orderId,uint256 nonce,uint64 expiry)");`

## Security & Compliance Notes

### 1. 2-of-3 Multi-Signature Arbitration Security & Threat Model
- **Collusion Resistance:** In a 3-party setup (Buyer, Seller, Arbitrator), no single entity can seize or reallocate funds unilaterally.
  - A malicious seller cannot refuse to release funds if the buyer paid; the Buyer and Arbitrator co-sign to force release to the Buyer.
  - A fraudulent buyer cannot fake payment to steal tokens; the Seller and Arbitrator co-sign to refund tokens back to the Seller.
  - An exchange arbitrator cannot confiscate or misappropriate funds without the valid co-signature of one of the trading counterparties.
- **Strict Signer Identity Verification:** In `resolveDisputeWith2of3`, the contract recovers both ECDSA signers ($s_1, s_2$) and verifies:
  $$s_1 \neq s_2$$
  $$s_1 \in \{\text{order.buyer}, \text{order.seller}, \text{arbitratorAddress}\}$$
  $$s_2 \in \{\text{order.buyer}, \text{order.seller}, \text{arbitratorAddress}\}$$
  If both signatures originate from the same address, or if an unauthorized third-party key is detected, execution immediately reverts with `DuplicateSignatures` or `InvalidThresholdSigners`.

### 2. Timelock Expiration Safety & Griefing Mitigation
- **Buyer Default Griefing Mitigation:** If a buyer locks seller crypto by matching an order but fails to transfer fiat within the 15-minute payment window ($T_{\text{payment}}$), the seller is protected against indefinite capital lockup. Upon $T_{\text{payment}}$ expiry, the seller executes `refundExpiredPaymentWindow()` to unilaterally recover 100% of their tokens.
- **Seller Extortion Griefing Mitigation:** If a buyer marks payment sent within $T_{\text{payment}}$, the funds enter the confirmation state. If the seller goes offline or maliciously refuses to acknowledge receipt, the buyer triggers `openDispute()`. The dispute state freezes the contract timer, preventing the seller from calling a timeout refund while the compliance team conducts UTR verification under RUNBOOK-27.
- **Maximum Dispute Ceiling ($T_{\text{dispute\_max}}$):** If an arbitration remains stalled for more than 7 days, multi-sig governance (`DEFAULT_ADMIN_ROLE`) can intervene to prevent permanent liquidity immobilization.

### 3. Cryptographic Bank Receipt Verification (UTR Hash)
- To prevent on-chain pollution with banking information and adhere to DPDP Act 2023 zero-PII mandates, the buyer submits:
  $$\text{utrHash} = \text{keccak256}(\text{abi.encodePacked}(\text{utrNumber}, \text{orderId}, \text{buyerAddress}))$$
  The plaintext UTR number is verified off-chain by the P2P Escrow Service (Prompt 269) via Open Banking Account Aggregator APIs. When the compliance officer signs the arbitration payload, the signature binds to this `utrHash`, establishing an immutable, auditable bridge between traditional banking rails and on-chain settlement.

### 4. Zero On-Chain PII & Regulatory Audit Trail
- No bank account numbers, IFSC codes, mobile numbers, PAN, or personal identities are recorded on the Besu ledger.
- All emitted events (`PaymentMarked`, `DisputeOpened`, `DisputeResolved`) include cryptographic leaf hashes (`utrHash`, `disputeReasonHash`) and timestamps. These events provide a complete, tamper-proof audit trail for regulatory inspection under SEBI and FIU-IND compliance directives.

### 5. Reentrancy Protection and Checks-Effects-Interactions (CEI)
- All mutator methods (`createEscrow`, `confirmPaymentAndRelease`, `resolveDisputeWith2of3`, `refundExpiredPaymentWindow`, `mutualCancel`) strictly follow the Checks-Effects-Interactions pattern.
- The order status is transitioned (e.g. to `COMPLETED` or `REFUNDED`) and nonces are incremented before invoking external token transfers via `SafeERC20`.
- Inherits OpenZeppelin `ReentrancyGuardUpgradeable` on all external state modifications to prevent reentrancy vectors across non-standard or malicious ERC-20 tokens.

### 6. Growww Fixed Fee Model Integrity
- Universal Zero-Fee Model (0.00% fee - No fee at all) (0.00% fee / 0 bps at launch) is charged on gross token volume exclusively upon successful trade completion or arbitrated release to the buyer.
- Fee distribution: Governed dynamically via FeeController.sol (0.00% at launch).
- Zero fee is assessed on cancelled or refunded trades. Sellers receiving a refund recover 100% of their principal.

## Acceptance Criteria
- [ ] `IP2PEscrow.sol` and `P2PEscrow.sol` compile cleanly under Solidity ^0.8.24 with zero warnings.
- [ ] Contract is fully upgradeable via the UUPS pattern and protected by `Initializable` and `AccessControlUpgradeable`.
- [ ] `createEscrow` correctly transfers collateral tokens from seller to the escrow contract and enforces payment window bounds ($900\text{s} \le \text{window} \le 7200\text{s}$).
- [ ] `markPaymentSent` verifies caller is buyer or operator, confirms payment window has not expired, stores `utrHash`, and transitions status to `PAYMENT_MARKED`.
- [ ] `confirmPaymentAndRelease` callable by seller releases net tokens to buyer, deducts exactly 0.00% fee (No fee at all), and distributes Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
- [ ] `refundExpiredPaymentWindow` allows seller to reclaim 100% of collateral with zero fee deduction if payment window expires without buyer action.
- [ ] `openDispute` freezes escrow state, transitions status to `DISPUTED`, and prevents unilateral seller refunds.
- [ ] `resolveDisputeWith2of3` successfully verifies any 2 distinct signatures out of Buyer, Seller, and Compliance Arbitrator using EIP-712 structured hashing.
- [ ] Submitting identical co-signatures reverts with `DuplicateSignatures`.
- [ ] Submitting signatures with an unauthorized third-party key reverts with `InvalidThresholdSigners`.
- [ ] Submitting expired arbitration payloads reverts with `SignatureExpired`.
- [ ] Arbitrated release to Buyer deducts the canonical 0.00% fee (No fee at all); arbitrated refund to Seller assesses 0% fee.
- [ ] `mutualCancel` allows buyer and seller to co-sign an agreement returning 100% of collateral to seller without fee deduction.
- [ ] Reentrancy attacks against `confirmPaymentAndRelease` and `resolveDisputeWith2of3` are blocked by `ReentrancyGuardUpgradeable`.
- [ ] Zero personally identifiable information (PII) is written to contract storage or event logs.
- [ ] Foundry test suite achieves >95% statement and branch coverage across all escrow lifecycle states, boundary conditions, and cryptographic edge cases.
- [ ] Slither and Mythril static security analyzers report zero high or medium severity vulnerabilities.
- [ ] Zero em dashes and zero en dashes present across the entire specification document (standard ASCII hyphens exclusively).

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301`: Permissioned Blockchain Evaluation & Selection (Hyperledger Besu).
  - Prompt `303`: Token Issuance Smart Contract (ERC-20, ERC-3643, and wrapped eINR).
  - Prompt `305`: Transfer Compliance Hooks Smart Contract (ONCHAINID identity verification).
  - Prompt `307`: MultiSig Governance Smart Contract (Administrative role controller).
- **Parallel Tasks:**
  - Prompt `269`: P2P Escrow Backend Service (Go microservice handling off-chain order state, timers, and Open Banking UTR ingest).
  - Prompt `217`: Admin & Back-Office Service (Maker-Checker dispute arbitration console under RUNBOOK-27).
  - Prompt `717`: CloudHSM Key Lifecycle & Signing Daemon (FIPS 140-2 Level 3 HSM service signing `ARBITRATOR_ROLE` payloads).
  - Prompt `212`: UPI, IMPS, RTGS & NetBanking Payment Gateway Service (Fiat banking integration).
  - Prompt `230`: Settlement Guarantee Fund (SGF) Service (Vault accounting for 25% fee share).
- **Subsequent Prompts Enabled:**
  - Prompt `309`: Blockchain Event Indexing Service (Indexing `EscrowCreated`, `PaymentMarked`, and `DisputeResolved` events).
  - Prompt `315`: Settlement Guarantee Fund Smart Contract (Receiving Core SGF reserve allocations).
  - Prompt `521`: Flutter Mobile P2P Trading & Escrow Dispute Resolution Flow.
  - Prompt `605`: Admin Multi-Party Dispute Arbitration Workbench.
