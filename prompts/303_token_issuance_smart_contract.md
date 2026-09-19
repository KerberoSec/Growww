# 303 - Permissioned Asset Token Issuance Smart Contract (ERC-3643 / CMTA)

## Purpose
In the Growww ecosystem, digital asset tokens represent fractional or whole ownership of real-world Indian equities held in custody by a SEBI-registered custodian / depository participant (NSDL/CDSL). To prevent unbacked synthetic token generation, token issuance must follow a strict, cryptographic 1:1 custody-matching invariant. Tokens can only be minted when a physical share credit confirmation (demat statement receipt) is validated, signed off by multi-signature governance, and committed on-chain.

This prompt specifies the design, implementation, testing, and deployment of the **ERC-3643 (T-REX standard) compliant Digital Security Token** smart contracts on the permissioned Hyperledger Besu ledger. The contract standard guarantees that security tokens are legally compliant, natively non-transferable to non-KYC'd entities, support corporate action balance adjustments (splits/bonuses), and strictly enforce 18-decimal fractionalization.

## What You Are Building
A modular Solidity smart contract suite under `contracts/tokens/` containing:
- `DigitalSecurityToken.sol`: Core ERC-3643 compliant security token implementing restricted minting, compliant transfers, burning, pausing, and multi-partition share tracking.
- `TokenFactory.sol`: Factory contract deployed to instantiate new standardized token contracts for each newly listed Indian equity ISIN (e.g. `GROWWW-TCS-INE467B01029`).
- `TokenStorage.sol` / `TokenProxy.sol`: ERC-1967 upgradeable storage proxy contract for deterministic upgradeability.
- Comprehensive Foundry unit and fuzz test suite (`test/tokens/DigitalSecurityToken.t.sol`) validating 100% minting invariants, custodian signature verifications, and overflow protections.

## Scope Boundaries
- **In Scope:**
 - Implementation of ERC-3643 interface (`IERC3643`, `IToken`, `IERC20`).
 - Minting logic requiring dual-signature custodian confirmation and multi-sig authorization (Prompt 307).
 - Linking to `IdentityRegistry.sol` for on-chain KYC/AML verification checks on token mint recipient addresses.
 - Recording custodian reference metadata (ISIN, Depository Demat Account ID, Batch Ingestion Hash).
 - Fractional precision handling (18 decimal places, supporting units down to $10^{-18}$ shares).
- **Out of Scope / Handled Elsewhere:**
 - Token redemption and burn orchestration (handled in Prompt 304).
 - Detailed transfer compliance rule engines and lockup logic (handled in Prompt 305).
 - Off-chain Custodian API integration adapter (handled in Prompt 213).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Solidity 0.8.24 provides native checked arithmetic (preventing integer underflows/overflows), user-defined value types, transient storage opcodes, and mature ecosystem tooling.
- **Contract Framework:** OpenZeppelin Contracts v5.0 (Upgradeable) & ERC-3643 T-REX reference implementation.
- **Development & Testing Toolchain:** Foundry (`forge` for compilation and ultra-fast fuzz testing, `cast` for Besu RPC contract interactions).
- **Static Analysis Tools:** Slither (Trail of Bits), Mythril, and Solhint.

## Backend / Infra Touchpoints
- **Custodian Depository Service:** Microservice (Prompt 213) that ingests NSDL/CDSL DIS/demat credit slips and triggers multi-sig mint proposals.
- **Event Indexer:** Blockchain Event Indexer (Prompt 309) that listens for `TokensMinted` and `CustodyProofLinked` events to update off-chain PostgreSQL portfolio balances.
- **Reconciliation Service:** Daily reconciliation engine (Prompt 215) validating on-chain `totalSupply()` against NSDL/CDSL depository holding statements.

## Blockchain Interaction
- **Contract Deployment:** Deployed via `TokenFactory.sol` on Hyperledger Besu consortium ledger.
- **Minting Functionality:** Restricted exclusively to the `MINTER_ROLE`, granted solely to the `MultiSigGovernance.sol` contract (Prompt 307).
- **Compliance Gating:** Before minting tokens to `to` address, `DigitalSecurityToken.sol` calls `IIdentityRegistry(identityRegistry).isVerified(to)`; if false, execution reverts with `IdentityNotVerified()`.
- **1:1 Custody Invariant:** Mint transactions emit `TokensMintedWithCustodyProof(to, amount, isin, depositoryBatchId, dematTxRefHash)` linking on-chain mints to legal depository records.
- **Zero PII:** Investor addresses are pseudonymous Ethereum addresses mapped to `bytes32 identityId` within the Identity Registry. No personal details exist on-chain.

## Step-by-Step Build Instructions
1. Initialize Foundry project under `contracts/` with standard directory layout: `src/tokens/`, `src/interfaces/`, `test/tokens/`, `script/`.
2. Install OpenZeppelin Contracts Upgradeable v5.0 and ERC-3643 standard interfaces as git submodules.
3. Define `IERC3643.sol` and `IDigitalSecurityToken.sol` interfaces incorporating standard ERC-20 functions alongside security token extensions (`mint`, `burn`, `freeze`, `setIdentityRegistry`, `compliance`).
4. Implement `DigitalSecurityToken.sol` inheriting from `Initializable`, `ContextUpgradeable`, `AccessControlEnumerableUpgradeable`, `ERC1967UpgradeUpgradeable`, and `IERC3643`.
5. Define constructor and `initialize(string name, string symbol, uint8 decimals, string isin, address identityRegistry, address compliance, address defaultAdmin)` initializer function with reinitialization guards.
6. Implement `mint(address to, uint256 amount, bytes32 depositoryBatchId, bytes calldata custodyProof)`:
 - Check caller has `MINTER_ROLE`.
 - Validate `amount > 0`.
 - Call `identityRegistry.isVerified(to)` to ensure compliance.
 - Call `compliance.canValidateTokenAction(to, amount)` for regulatory transfer checks.
 - Increment `_totalSupply` and `_balances[to]`.
 - Emit `Transfer(address(0), to, amount)` and `TokensMintedWithCustodyProof(...)`.
7. Implement `batchMint(address[] calldata recipients, uint256[] calldata amounts, bytes32 batchId, bytes calldata custodyProof)` to enable gas-efficient batched allocation.
8. Implement address-level freezing (`freezeAddress`, `unfreezeAddress`) and token-level recovery functions mandated by regulatory authorities (for court-ordered freezes or lost key recovery).
9. Implement `TokenFactory.sol` with `createToken(...)` method using `ERC1967Proxy` to deploy upgradeable token instances for new equity listings.
10. Write extensive Foundry unit tests in `test/tokens/DigitalSecurityToken.t.sol` covering initialization, unauthorized mint attempts, non-verified recipient mints, and batch mint edge cases.
11. Write fuzz tests in Foundry testing invariant: `totalSupply() == sum(balances)` and checking that no mint succeeds without valid `MINTER_ROLE`.
12. Run Slither static analyzer: `slither src/tokens/DigitalSecurityToken.sol` and resolve any compiler warnings or reentrancy flags.
13. Write deployment scripts in Foundry (`script/DeployTokens.s.sol`) configured for Hyperledger Besu QBFT RPC endpoint.
14. Deploy token contracts to local Besu testnet, verify contract bytecode, and simulate minting 1,000,000 units of `GROWWW-RELIANCE-INE002A01018` against mock NSDL custody receipts.
15. Document contract ABI, method selectors, and error definitions for consumption by Backend Service teams.

## Interfaces / Contracts

### Digital Security Token Interface (`IDigitalSecurityToken.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IDigitalSecurityToken is IERC20 {
    // Custom Events
    event TokensMintedWithCustodyProof(
        address indexed to,
        uint256 amount,
        string indexed isin,
        bytes32 indexed depositoryBatchId,
        bytes custodyProof
    );
    event AddressFrozen(address indexed target, bool isFrozen, address indexed operator);
    event IdentityRegistryUpdated(address indexed oldRegistry, address indexed newRegistry);
    event ComplianceUpdated(address indexed oldCompliance, address indexed newCompliance);

    // Custom Errors
    error CallerNotMinter(address caller);
    error RecipientNotCompliant(address recipient);
    error InvalidZeroAddress();
    error InvalidMintAmount();
    error AddressIsFrozen(address account);
    error ISINMismatch(string expected, string provided);

    // View Functions
    function isin() external view returns (string memory);
    function identityRegistry() external view returns (address);
    function compliance() external view returns (address);
    function isFrozen(address account) external view returns (bool);

    // Privileged Minting Functions
    function mint(
        address to,
        uint256 amount,
        bytes32 depositoryBatchId,
        bytes calldata custodyProof
    ) external;

    function batchMint(
        address[] calldata recipients,
        uint256[] calldata amounts,
        bytes32 depositoryBatchId,
        bytes calldata custodyProof
    ) external;

    // Regulatory / Compliance Management
    function freezeAddress(address account) external;
    function unfreezeAddress(address account) external;
    function setIdentityRegistry(address newIdentityRegistry) external;
    function setCompliance(address newCompliance) external;
}
```

### Factory Contract Interface (`ITokenFactory.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ITokenFactory {
    event TokenDeployed(
        string indexed isin,
        address indexed tokenAddress,
        address indexed tokenProxy,
        string name,
        string symbol
    );

    function deploySecurityToken(
        string calldata name,
        string calldata symbol,
        string calldata isin,
        address identityRegistry,
        address compliance,
        address adminMultisig
    ) external returns (address proxyAddress);

    function getTokenByISIN(string calldata isin) external view returns (address);
}
```

## Security & Compliance Notes
- **Multi-Party Mint Authorization:** The `MINTER_ROLE` is strictly granted to `MultiSigGovernance.sol` (Prompt 307) requiring 3-of-5 institutional threshold signatures (Growww Compliance + Custodian Officer + Clearing Officer). No single private key can mint tokens.
- **1:1 Custodial Invariant Enforcement:** Mint transactions require a cryptographic hash `depositoryBatchId` referencing signed NSDL/CDSL credit files; periodic reconciliations (Prompt 215) will halt the platform if ledger supply exceeds depository custody.
- **Zero PII on Chain:** All token holders are identified strictly by their Ethereum addresses. The on-chain `IdentityRegistry` checks identity validity without storing names, PAN numbers, or emails.
- **Regulatory Freeze & Court Order Compliance:** Under SEBI / regulatory directives, the compliance officer can invoke `freezeAddress(account)` via multi-sig to block transfers and dividend disbursements for disputed or sanctioned accounts.

## Acceptance Criteria
- [ ] `DigitalSecurityToken.sol` and `TokenFactory.sol` fully implemented adhering to ERC-3643 standards.
- [ ] 100% test coverage in Foundry across unit tests, edge cases, and fuzz testing suites.
- [ ] Minting fails unconditionally if recipient address is not verified in `IdentityRegistry.sol`.
- [ ] Minting fails unconditionally if caller lacks `MINTER_ROLE`.
- [ ] Slither static analysis runs with 0 high/medium issues.
- [ ] Factory deployment produces deterministic, verified proxy contracts on local Besu QBFT network.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Platform Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `000` (Project North Star).
- **Parallel Tasks:** Prompt `305` (Transfer Compliance Hooks), Prompt `307` (Multisig Governance).
- **Subsequent Prompts Enabled:** Prompt `304` (Token Redemption), Prompt `306` (Settlement DvP), Prompt `308` (Proof of Reserve).
