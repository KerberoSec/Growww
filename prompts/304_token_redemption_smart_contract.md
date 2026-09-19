# 304 - Token Redemption & Custodial Asset Release Smart Contract

## Purpose
In a 1:1 asset-backed securities infrastructure, token redemption is the mirror operation of token issuance. When an investor liquidates their fractional equity holdings for INR cash or when an institutional investor requests physical conversion into whole shares transferred to their personal NSDL/CDSL Demat account, the corresponding digital equity tokens on the Hyperledger Besu ledger must be verifiably and irreversibly burned. 

This prompt specifies the architecture, logic, and implementation of the **Token Redemption & Custodial Asset Release Smart Contract** (`TokenRedemption.sol`). To eliminate double-spend and settlement failure risks, the contract implements a cryptographically secure, two-phase commitment pattern: (1) Token Escrow / Lock Phase upon user redemption request, followed by (2) Atomic Token Burn upon verified depository debit confirmation or fiat payout confirmation.

## What You Are Building
A robust Solidity smart contract system under `contracts/redemption/` consisting of:
- `TokenRedemption.sol`: Two-phase redemption manager orchestrating escrow locks, custody signature verification, atomic burning, and emergency timeout refunds.
- `IRedemptionReceiver.sol`: Callback interface for automated settlement notification.
- Comprehensive Foundry unit and stateful invariant test suite (`test/redemption/TokenRedemption.t.sol`) validating escrow invariants, multi-sig execution authorization, and prevention of reentrancy attacks.

## Scope Boundaries
- **In Scope:**
 - On-chain two-phase redemption state machine (`INITIATED`, `ESCROWED`, `SETTLED_BURNED`, `CANCELLED_REFUNDED`).
 - Escrow locking of `DigitalSecurityToken` units from investor balances.
 - Verification of authorized Custodian / Settlement Relayer completion signatures with EIP-712 structured data hashing.
 - Atomic burning of tokens via `DigitalSecurityToken.sol` `burn` interface.
 - Expiry and emergency refund mechanics if custodial settlement fails within the SLA window.
- **Out of Scope / Handled Elsewhere:**
 - INR bank payout execution via payment rails (handled in Prompt 212).
 - Physical Demat transfer instructions to NSDL/CDSL (handled in Prompt 213).
 - Off-chain settlement orchestration service (handled in Prompt 208).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Solidity 0.8.24 provides custom error types, typed hashing via EIP-712, transient storage optimizations, and tight gas economics for multi-step escrow operations.
- **Framework & Libraries:** OpenZeppelin Contracts v5.0 (ReentrancyGuardUpgradeable, PausableUpgradeable, ECDSA, EIP712Upgradeable).
- **Development Toolchain:** Foundry (`forge` test runner, `anvil` local node).
- **Static Analysis:** Slither, Mythril.

## Backend / Infra Touchpoints
- **Trade Settlement Service:** Microservice (Prompt 208) that coordinates redemption requests and relays depository settlement confirmations.
- **Custodian Depository Integration Service:** Microservice (Prompt 213) that dispatches physical share debit requests to NSDL/CDSL APIs.
- **Wallet & Account Service:** Core INR balance service (Prompt 203) releasing fiat funds to investor balances upon on-chain burn confirmation.
- **Blockchain Event Indexer:** Service (Prompt 309) indexing `RedemptionCompleted` and `TokensBurned` events.

## Blockchain Interaction
- **Escrow Mechanics:** When `requestRedemption` is called, tokens are transferred from the investor to the `TokenRedemption.sol` contract (or locked in place via ERC-3643 partition locks).
- **Burn Execution:** When the settlement relayer invokes `executeRedemptionWithProof(redemptionId, dematDebitRef, custodySignature)`, the contract calls `IDigitalSecurityToken(token).burn(address(this), amount)` and emits `RedemptionCompleted`.
- **1:1 Custody Invariant:** Every token burned reduces `totalSupply()` on-chain, matching the exact quantity of physical shares debited from the omnibus custody account at NSDL/CDSL.
- **Zero PII:** Redemptions record only `bytes32 redemptionId`, `address investor`, `uint256 amount`, and `bytes32 dematDebitRefHash`.

## Step-by-Step Build Instructions
1. Create smart contract source file `contracts/redemption/TokenRedemption.sol` and interface `contracts/interfaces/ITokenRedemption.sol`.
2. Define data structures: `enum RedemptionStatus { NONE, INITIATED, ESCROWED, SETTLED_BURNED, CANCELLED }` and `struct RedemptionRequest { bytes32 id; address investor; address tokenAddress; uint256 amount; uint64 requestTimestamp; uint64 expiryTimestamp; RedemptionStatus status; bytes32 dematDebitRef; }`.
3. Implement ERC-1967 upgradeable storage layout and initializer `initialize(address adminMultisig, address initialRelayer, uint64 defaultTimeoutSeconds)`.
4. Inherit `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, and `EIP712Upgradeable` ("GrowwwRedemptionManager", "1").
5. Implement `requestRedemption(address token, uint256 amount)`:
 - Validate `token` is an authorized security token and `amount > 0`.
 - Generate unique `bytes32 redemptionId = keccak256(abi.encode(msg.sender, token, amount, block.timestamp, nonces[msg.sender]++))`.
 - Call `IERC20(token).transferFrom(msg.sender, address(this), amount)`.
 - Store `RedemptionRequest` with status `ESCROWED`.
 - Emit `RedemptionInitiated(redemptionId, msg.sender, token, amount, expiryTimestamp)`.
6. Implement `executeRedemptionWithProof(bytes32 redemptionId, bytes32 dematDebitRef, bytes calldata signature)`:
 - Ensure caller has `RELAYER_ROLE` or verify cryptographic EIP-712 signature from the authorized Custodian key.
 - Verify request status is `ESCROWED` and not expired (`block.timestamp <= request.expiryTimestamp`).
 - Transition status to `SETTLED_BURNED`.
 - Call `IDigitalSecurityToken(request.tokenAddress).burn(address(this), request.amount)`.
 - Emit `RedemptionCompleted(redemptionId, dematDebitRef, block.timestamp)`.
7. Implement `cancelExpiredRedemption(bytes32 redemptionId)`:
 - Check request exists and status is `ESCROWED`.
 - Check `block.timestamp > request.expiryTimestamp`.
 - Transition status to `CANCELLED`.
 - Refund tokens: `IERC20(request.tokenAddress).transfer(request.investor, request.amount)`.
 - Emit `RedemptionCancelled(redemptionId, "SLA_EXPIRED")`.
8. Implement administrative emergency cancel `adminCancelRedemption(bytes32 redemptionId, string calldata reason)` restricted to multi-sig governance.
9. Implement batch settlement method `executeBatchRedemption(...)` for high-volume settlement cycles.
10. Write unit tests in Foundry covering standard lifecycle (Request -> Lock -> Execute -> Burn).
11. Write edge case tests covering double execution, execution after expiry, unauthorized relayer calls, and reentrancy vectors.
12. Write stateful invariant fuzz tests verifying that contract balance plus burned supply always equals total tokens received into escrow.
13. Run Slither analysis and verify zero reentrancy or state variable shadowing vulnerabilities.
14. Configure deployment scripts in Foundry targeting Hyperledger Besu network.
15. Test deployment and execution against mock `DigitalSecurityToken.sol` on local Besu devnet.

## Interfaces / Contracts

### Token Redemption Interface (`ITokenRedemption.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ITokenRedemption {
    enum RedemptionStatus {
        NONE,
        ESCROWED,
        SETTLED_BURNED,
        CANCELLED
    }

    struct RedemptionRequest {
        bytes32 id;
        address investor;
        address tokenAddress;
        uint256 amount;
        uint64 requestTimestamp;
        uint64 expiryTimestamp;
        RedemptionStatus status;
        bytes32 dematDebitRef;
    }

    event RedemptionInitiated(
        bytes32 indexed id,
        address indexed investor,
        address indexed token,
        uint256 amount,
        uint64 expiryTimestamp
    );

    event RedemptionCompleted(
        bytes32 indexed id,
        bytes32 indexed dematDebitRef,
        uint256 timestamp
    );

    event RedemptionCancelled(
        bytes32 indexed id,
        string reason
    );

    // Errors
    error InvalidAmount();
    error TokenNotSupported(address token);
    error RequestNotFound(bytes32 id);
    error InvalidStatus(RedemptionStatus currentStatus);
    error RequestExpired(uint64 expiryTimestamp, uint256 currentTimestamp);
    error RequestNotExpired(uint64 expiryTimestamp, uint256 currentTimestamp);
    error InvalidSignature();

    // User Operations
    function requestRedemption(address token, uint256 amount) external returns (bytes32 redemptionId);

    // Custodian / Relayer Settlement Operations
    function executeRedemptionWithProof(
        bytes32 redemptionId,
        bytes32 dematDebitRef,
        bytes calldata custodianSignature
    ) external;

    function executeBatchRedemption(
        bytes32[] calldata redemptionIds,
        bytes32[] calldata dematDebitRefs,
        bytes[] calldata signatures
    ) external;

    // Failure / Expiry Recovery
    function cancelExpiredRedemption(bytes32 redemptionId) external;
    function adminCancelRedemption(bytes32 redemptionId, string calldata reason) external;

    // View Functions
    function getRequest(bytes32 redemptionId) external view returns (RedemptionRequest memory);
}
```

## Security & Compliance Notes
- **Non-Reversible Burning:** Once tokens are burned via `executeRedemptionWithProof`, they are removed permanently from on-chain supply. The contract ensures burn only occurs after cryptographic proof of physical asset release/fiat reservation.
- **Fail-Safe Timeout Mechanics:** If the custodian API or banking rail experiences an outage exceeding the configured SLA (e.g. 24 hours), the investor can safely recover their locked tokens by calling `cancelExpiredRedemption`.
- **EIP-712 Structured Proofs:** Off-chain settlement proofs signed by the Custodian HSM include domain separator, chain ID, `redemptionId`, `dematDebitRef`, and timestamp to prevent replay attacks across environments or chains.
- **Reentrancy Protection:** All state transitions occur prior to external token transfers or burning operations (`nonReentrant` modifier enforced).

## Acceptance Criteria
- [ ] `TokenRedemption.sol` contract deployed and verified on Hyperledger Besu.
- [ ] Two-phase lifecycle verified: token transfer to escrow -> burn confirmation reducing `totalSupply()`.
- [ ] Expiry refund verified: tokens cannot be burned after expiry and are returned to the user wallet.
- [ ] EIP-712 cryptographic signature verification tested with valid and invalid signer keys.
- [ ] 100% test coverage across Foundry unit and fuzzing suites with zero Slither issues.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance Smart Contract), Prompt `305` (Compliance Hooks), Prompt `101` (System Architecture).
- **Parallel Tasks:** Prompt `208` (Trade Settlement Service), Prompt `213` (Custodian Depository Integration).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (Proof of Reserve).
