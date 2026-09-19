# 315 - On-Chain Settlement Guarantee Fund & Default Waterfall Smart Contract (SettlementGuaranteeFund.sol)

## Purpose
A decentralized or permissioned clearinghouse must guarantee that financial commitments, collateral allocations, and emergency default waterfall executions are transparent, tamper-proof, and mathematically non-bypassable. Off-chain clearing systems are vulnerable to opaque discretion or delayed default resolution during systemic market panics.

The **On-Chain Settlement Guarantee Fund & Default Waterfall Smart Contract (`SettlementGuaranteeFund.sol`)** establishes an immutable, programmable settlement guarantee facility on Hyperledger Besu (QBFT consensus). It locks clearing members' tokenized cash reserves (e₹ CBDC / sovereign tokenized money) and government securities into segregated multi-tranche pools. Upon an authorized multi-sig default declaration, the contract autonomously executes the statutory SEBI / CPMI-IOSCO Default Waterfall sequence, slashing tranches in exact priority order, distributing liquidity to the settlement pool, and issuing programmatic assessment calls for fund replenishment.

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/settlement/` comprising:
- `SettlementGuaranteeFund.sol`: Core contract managing member deposit registries, multi-tranche collateral pools, and default waterfall execution logic.
- `ISettlementGuaranteeFund.sol`: Standardized interface defining all administrative, operational, view, and event signatures.
- Multi-Tranche Reserve Vault: Segregated on-chain accounting for:
 - *Tranche 1:* Member Initial & Variation Margin balances.
 - *Tranche 2:* Member Core SGF contribution quotas.
 - *Tranche 3:* Clearing Corporation (CC) Skin-in-the-Game dedicated capital.
 - *Tranche 4:* Mutualized non-defaulting member SGF corpus.
 - *Tranche 5:* CC emergency reserve capital.
- Default Waterfall State Machine: Enforces non-bypassable sequential tranche drawdown with multi-signature authorization and timelocked safety valves.
- Comprehensive Foundry Test Suite (`test/settlement/SettlementGuaranteeFund.t.sol`): Validating multi-member stress defaults, pro-rata slashing, assessment replenishments, reentrancy guards, and access controls.

## Scope Boundaries
- **In Scope:**
 - On-chain locking and tracking of member SGF collateral deposits (ERC-20 tokenized e₹ / CBDC).
 - Enforcing Clearing Corporation dedicated skin-in-the-game capital minimums.
 - Multi-party default declaration and cryptographic authorization verification.
 - Step-by-step automated Default Waterfall execution (Tranches 1 through 5).
 - Pro-rata mutualized slashing across non-defaulting clearing participants in Tranche 4.
 - Tracking replenishment quotas and assessment call deadlines.
 - Emitting detailed audit events for block indexers and regulatory observers.
- **Out of Scope / Handled Elsewhere:**
 - Off-chain Cover-1 / Cover-2 daily stress scenario simulations (handled in Prompt 230).
 - Real-time pre-trade VaR and margin portfolio calculation (handled in Prompt 229).
 - Primary trade DvP settlement execution (handled in Prompt 306).
 - Physical fiat banking collateral movements at the Reserve Bank of India (handled in Prompt 232).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Cancun/Shanghai with Paris backward compatibility).
- **Security & Standards:** OpenZeppelin Contracts Upgradeable v5.0 (`Initializable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Development & Testing Framework:** **Foundry** (`forge`, `cast`) for deterministic testing, fuzzing, and invariant verification.
- **Static Analysis:** Slither and Mythril automated security analysis pipelines.

## Backend / Infra Touchpoints
- **Settlement Guarantee Fund Service (Prompt 230):** Off-chain orchestrator dispatching waterfall execution commands.
- **Trade Settlement Service (Prompt 208):** Ingests liquidity released from SGF contract to finalize pending trade settlements.
- **CBDC Settlement Adapter (Prompt 232):** Mints/burns and transfers tokenized e₹ into the SGF contract vault.
- **Blockchain Event Indexer (Prompt 309):** Ingests `TrancheSlashed`, `DefaultResolved`, and `AssessmentCallIssued` events to update off-chain relational databases.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Direct Asset Movement:** SGF contract interacts with approved settlement tokens (e.g., `eINRToken.sol`, `TokenizedGSec.sol`) using `SafeERC20.safeTransfer` and `SafeERC20.safeTransferFrom`.
- **Atomic Liquidity Injection:** Slashed collateral is transferred directly to `SettlementDvP.sol` (Prompt 306) to fulfill cash obligations of defaulted transactions within the same block.
- **Role-Based Governance:** High-risk actions (default declaration, tranche slashing, parameter updates) require threshold multi-sig signatures from `MultiSigGovernance.sol` (Prompt 307) backed by CloudHSM validator keys.
- **Zero On-Chain PII:** The contract stores only `bytes32 memberIdHash`, `address memberWallet`, `uint256 depositAmount`, and `bytes32 defaultId`.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Contract Directory:** Create `contracts/settlement/SettlementGuaranteeFund.sol` and interface `contracts/interfaces/ISettlementGuaranteeFund.sol`.
2. **Define Data Structures & Enums:** Define `TrancheType` enum (MARGIN, MEMBER_SGF, CC_SKIN_IN_THE_GAME, POOLED_SGF, CC_RESERVES) and structs `MemberAllocation`, `DefaultRecord`, and `AssessmentCall`.
3. **Implement ERC-1967 Upgradeability:** Inherit from OpenZeppelin `UUPSUpgradeable` and `AccessControlUpgradeable` with storage gap reservations for future upgrades.
4. **Implement Access Control Roles:** Define `DEFAULT_ADMIN_ROLE`, `RISK_COMMITTEE_ROLE`, `SETTLEMENT_OPERATOR_ROLE`, and `EMERGENCY_GUARDIAN_ROLE`.
5. **Implement Collateral Deposit & Allocation Logic:**
 - `depositMemberSGF(bytes32 memberIdHash, address token, uint256 amount)`: Pulls tokens into contract, updates `memberAllocations[memberIdHash]`, and emits `SGFDepositReceived`.
 - `depositCCContribution(address token, uint256 amount)`: Dedicated Clearing Corporation skin-in-the-game capital.
6. **Implement Default Declaration Module:**
 - `declareMemberDefault(bytes32 defaultId, bytes32 defaulterIdHash, uint256 defaultAmountINR)` restricted to `RISK_COMMITTEE_ROLE`.
 - Freezes the defaulting member's SGF assets and initializes `DefaultRecord`.
7. **Implement Waterfall Stage 1 (Defaulter Margins):**
 - `slashDefaulterMargins(bytes32 defaultId, address settlementVault)`: Slashes locked margin tokens and transfers them to the settlement pool.
8. **Implement Waterfall Stage 2 (Defaulter SGF Contribution):**
 - `slashDefaulterSGF(bytes32 defaultId, address settlementVault)`: Deducts the defaulter's own SGF quota.
9. **Implement Waterfall Stage 3 (Clearing Corporation Capital):**
 - `slashCCContribution(bytes32 defaultId, address settlementVault, uint256 amount)`: Draws down CC skin-in-the-game capital before touching non-defaulting member funds.
10. **Implement Waterfall Stage 4 (Pooled Non-Defaulting Member SGF):**
 - `slashPooledSGF(bytes32 defaultId, address settlementVault, uint256 requiredAmount)`: Calculates pro-rata share per solvent member:
      $$\text{Slashing Share}_i = \text{Required Amount} \times \frac{\text{MemberDeposit}_i}{\sum \text{SolventDeposits}}$$
 - Deducts balances atomically across all solvent members.
11. **Implement Waterfall Stage 5 (CC Reserves & Final Resolution):**
 - `slashCCReserves(bytes32 defaultId, address settlementVault, uint256 amount)` and mark default as resolved.
12. **Implement Assessment Calls & Replenishment:**
 - `issueAssessmentCall(bytes32 defaultId, bytes32 memberIdHash, uint256 assessedAmount, uint256 deadline)`: Emits on-chain assessment requirements with maximum cap.
13. **Implement Emergency Pause & Migration:** Add circuit-breaker emergency pause callable by `EMERGENCY_GUARDIAN_ROLE`.
14. **Write Comprehensive Foundry Test Suite:**
 - Unit tests: single-member deposits, withdrawals, and role validations.
 - Integration tests: Full waterfall execution (Stage 1 $\rightarrow$ Stage 4) with 10 simulated clearing members.
 - Invariant tests: Invariant that $\sum \text{Tracked Allocations} == \text{Contract Token Balance}$ at all times.
15. **Run Static Analysis & Formal Verification:** Execute Slither static analyzer and Certora/Foundry invariant fuzzing to verify zero reentrancy or math overflow vulnerabilities.

## Interfaces / Contracts

### Solidity Interface (`ISettlementGuaranteeFund.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ISettlementGuaranteeFund {
    enum TrancheType {
        DEFAULTER_MARGINS,
        DEFAULTER_SGF,
        CC_SKIN_IN_THE_GAME,
        POOLED_SGF,
        CC_RESERVES
    }

    enum DefaultStatus {
        UNSPECIFIED,
        DECLARED,
        PROCESSING_WATERFALL,
        RESOLVED,
        FAILED
    }

    struct MemberAllocation {
        bytes32 memberIdHash;
        address depositToken;
        uint256 sgfDepositBalance;
        uint256 lockedMarginBalance;
        uint256 lastUpdatedTimestamp;
        bool isDefaulted;
    }

    struct DefaultRecord {
        bytes32 defaultId;
        bytes32 defaulterIdHash;
        uint256 totalDefaultAmount;
        uint256 totalRecoveredAmount;
        TrancheType currentTranche;
        DefaultStatus status;
        uint256 declaredTimestamp;
        uint256 resolvedTimestamp;
    }

    // Events
    event SGFDepositReceived(bytes32 indexed memberIdHash, address indexed token, uint256 amount);
    event CCContributionDeposited(address indexed token, uint256 amount);
    event DefaultDeclared(bytes32 indexed defaultId, bytes32 indexed defaulterIdHash, uint256 defaultAmount);
    event TrancheSlashed(bytes32 indexed defaultId, TrancheType indexed tranche, uint256 amountDrawn, address destinationVault);
    event ProRataMemberSlashed(bytes32 indexed defaultId, bytes32 indexed memberIdHash, uint256 amountSlashed);
    event DefaultResolved(bytes32 indexed defaultId, uint256 totalRecovered, uint256 remainingDeficit);
    event AssessmentCallIssued(bytes32 indexed defaultId, bytes32 indexed memberIdHash, uint256 amountDue, uint256 deadline);

    // Custom Errors
    error UnauthorizedCaller(address caller);
    error DefaultAlreadyDeclared(bytes32 defaultId);
    error DefaultNotFound(bytes32 defaultId);
    error InvalidTrancheSequence(TrancheType expected, TrancheType provided);
    error InsufficientContractBalance(address token, uint256 available, uint256 requested);
    error MemberAlreadyDefaulted(bytes32 memberIdHash);
    error ZeroDepositAmount();
    error AssessmentExceedsCap(bytes32 memberIdHash, uint256 amount);

    // Operational Functions
    function depositMemberSGF(bytes32 memberIdHash, address token, uint256 amount) external;
    function depositCCContribution(address token, uint256 amount) external;
    function declareMemberDefault(bytes32 defaultId, bytes32 defaulterIdHash, uint256 defaultAmount) external;
    function executeWaterfallStep(bytes32 defaultId, TrancheType tranche, uint256 drawAmount, address destinationVault) external;
    function resolveDefault(bytes32 defaultId) external;
    function issueAssessmentCall(bytes32 defaultId, bytes32 memberIdHash, uint256 assessedAmount, uint256 deadline) external;

    // View Functions
    function getMemberAllocation(bytes32 memberIdHash) external view returns (MemberAllocation memory);
    function getDefaultRecord(bytes32 defaultId) external view returns (DefaultRecord memory);
    function getTotalSGFCorpus(address token) external view returns (uint256 totalMemberDeposits, uint256 ccContribution);
}
```

## Security & Compliance Notes
- **SEBI Core SGF Regulatory Compliance:** The smart contract enforces that Clearing Corporation capital (Tranche 3) must be fully drawn before non-defaulting members' SGF contributions (Tranche 4) can be slashed.
- **Reentrancy Protection:** All fund disbursement functions are guarded by OpenZeppelin `nonReentrant` modifiers and adhere strictly to the Checks-Effects-Interactions pattern.
- **Strict Mathematical Invariants:** SGF balances and allocations use safe fixed-point arithmetic (0.8.24 native overflow checks) ensuring that token deductions never exceed member or pool balances.
- **Multi-Party Timelock Controls:** Structural contract parameter changes (e.g., minimum CC contribution percentage or token whitelists) are subject to a 48-hour on-chain timelock managed by `MultiSigGovernance.sol` (Prompt 307).

## Acceptance Criteria
- [ ] `SettlementGuaranteeFund.sol` deployed, initialized, and verified on Hyperledger Besu devnet.
- [ ] Member SGF deposits and CC skin-in-the-game capital deposits correctly track balances.
- [ ] Default declaration freezes the defaulter's assets and locks the waterfall state machine.
- [ ] Tranche 1 (Defaulter Margins), Tranche 2 (Defaulter SGF), and Tranche 3 (CC Contribution) execute in strict non-bypassable sequence.
- [ ] Tranche 4 slashes non-defaulting member SGF balances strictly on a pro-rata basis.
- [ ] Invariant fuzz tests pass 10,000 runs in Foundry verifying zero token balance leakage.
- [ ] Slither and Mythril security scans pass with zero high or medium severity vulnerabilities.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Evaluation), Prompt `303` (Token Issuance), Prompt `306` (Settlement DvP Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `230` (SGF Orchestration Service), Prompt `232` (CBDC Settlement Adapter).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (On-Chain Proof of Reserve).
