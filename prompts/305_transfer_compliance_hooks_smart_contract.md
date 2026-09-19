# 305 - Smart Contract Transfer Compliance Hooks & KYC/AML Whitelist Registry

## Purpose
Under Indian securities regulations (SEBI) and international cross-border frameworks (GIFT City / IFSCA), equity securities cannot circulate freely without strict identity verification, AML screening, and regulatory transfer controls. Unlike public cryptocurrency tokens, securities tokens must enforce regulatory rules at the blockchain protocol level on every single transaction: peer-to-peer transfers, secondary market trades, and custodial mints/burns.

This prompt specifies the design, implementation, and deployment of the **On-Chain Identity and Modular Compliance Registry Suite** (based on the ERC-3643 / ONCHAINID standard architecture). Every transfer initiated on `DigitalSecurityToken.sol` invokes the `ComplianceRegistry` and `IdentityRegistry` before executing state changes. If either the sender or receiver lacks an active KYC claim, belongs to a sanctioned jurisdiction, or breaches regulatory ownership caps, the transaction reverts deterministically with detailed error codes.

## What You Are Building
A modular compliance framework under `contracts/compliance/` including:
- `IdentityRegistry.sol`: On-chain registry mapping pseudonymous Ethereum addresses to verified investor identity claims (`bytes32 identityId`, country code, KYC tier).
- `ComplianceRegistry.sol`: Modular compliance router orchestrating individual pluggable compliance rule modules.
- `CountryRestrictModule.sol`: Module restricting token holding/transfer based on investor country classification (e.g. domestic Indian vs GIFT City eligible jurisdictions vs FATF blacklisted nations).
- `MaxOwnershipModule.sol`: Module enforcing statutory maximum holding limits (e.g., non-promoter single-entity ownership caps).
- `LockupModule.sol`: Module enforcing regulatory lock-in periods (e.g. IPO / preferential allotment lockups).
- Comprehensive Foundry test suite (`test/compliance/ComplianceRegistry.t.sol`) validating whitelist gating, blacklist enforcement, and modular rule composition.

## Scope Boundaries
- **In Scope:**
 - On-chain investor identity claim verification without storing PII.
 - ERC-3643 compliant `canTransfer(from, to, value)` pre-transfer hook validation.
 - Whitelist registry management (adding, updating, revoking identity claims).
 - Pluggable modular compliance rules (`CountryRestrict`, `MaxOwnership`, `TransferLimit`).
 - Regulatory freeze and emergency circuit-breaker integrations.
- **Out of Scope / Handled Elsewhere:**
 - Off-chain KYC/AML document verification and OCR (handled in Prompt 202).
 - Sanctions list ingestion pipeline (handled in Prompt 703).
 - Domestic vs GIFT City identity verification policies (handled in Prompts 004 & 005).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun).
  *Justification:* Enables modular contract architecture using standardized function interfaces, custom error structures for debugging compliance failures, and efficient bitmap indexing for country/jurisdiction flags.
- **Standards & Frameworks:** ERC-3643 (T-REX Modular Compliance), ONCHAINID interfaces, OpenZeppelin AccessControl.
- **Testing & Tooling:** Foundry (`forge`, `cast`), Slither static analysis.

## Backend / Infra Touchpoints
- **KYC/AML Service:** Backend service (Prompt 202) that triggers on-chain identity claim registrations via relayer after successful Aadhaar/PAN/Passport verification.
- **Sanctions & PEP Screening Service:** Service (Prompt 703) that automatically revokes identity claims or adds addresses to the on-chain freeze registry upon sanctions list hits.
- **Admin Console:** Compliance officer portal (Prompt 604) for manual whitelist inspection and audit logs.

## Blockchain Interaction
- **Invocation Trigger:** `DigitalSecurityToken.sol` executes `require(compliance.canTransfer(from, to, amount), "ComplianceFailure")` inside `_update` / transfer hooks.
- **Zero PII Storage:** `IdentityRegistry.sol` stores only `bytes32 identityHash = keccak256(abi.encode(panHash, kycTier, salt))` and `uint16 countryCode`. No names, addresses, or identity numbers are ever written to state.
- **Immediate Reversal:** Non-compliant transfers revert instantly at zero gas cost in the simulation phase or fail deterministically in consensus without state corruption.

## Step-by-Step Build Instructions
1. Scaffold compliance contract repository structure: `contracts/compliance/`, `contracts/interfaces/`, `test/compliance/`.
2. Define `IIdentityRegistry.sol` and `IComplianceRegistry.sol` interfaces adhering to ERC-3643 specifications.
3. Implement `IdentityRegistry.sol` with mappings: `mapping(address => bytes32) private _identities` and `mapping(bytes32 => InvestorClaim) private _claims`.
4. Implement claim management functions: `registerIdentity(address userAddress, bytes32 identityId, uint16 countryCode, uint8 kycTier)` restricted to `AGENT_ROLE` / KYC relayer.
5. Implement claim revocation `deleteIdentity(address userAddress)` and batch update `batchRegisterIdentity(...)`.
6. Implement `ComplianceRegistry.sol` implementing `IModularCompliance` maintaining an array of active compliance module contracts: `address[] private _modules`.
7. Implement `canTransfer(address from, address to, uint256 amount)` in `ComplianceRegistry.sol`:
 - Verify `identityRegistry.isVerified(from)` and `identityRegistry.isVerified(to)`.
 - Iterate through all bound compliance modules; call `module.moduleCheck(from, to, amount)`.
 - Return `true` if and only if all modules return `true`.
8. Implement `CountryRestrictModule.sol`:
 - Maintain bitmask or mapping of allowed/forbidden ISO-3166 numeric country codes.
 - Enforce that both `from` and `to` reside in permissible investment jurisdictions.
9. Implement `MaxOwnershipModule.sol`:
 - Query target token balance + `amount` against maximum allowable percentage of `totalSupply()`.
 - Prevent market cornering or regulatory threshold breaches (e.g. >5% without disclosures).
10. Implement `LockupModule.sol`:
 - Record time-locked balances per investor address with expiration timestamps.
 - Check available unlocked balance: `balanceOf(from) - lockedBalance(from) >= amount`.
11. Implement emergency freeze hook: allow `COMPLIANCE_ADMIN` to freeze individual addresses or globally halt transfers in `DigitalSecurityToken.sol`.
12. Write extensive Foundry unit tests covering compliant transfers, unverified sender/receiver reverts, country blacklist triggers, and ownership cap breaches.
13. Write fuzz tests testing arbitrary transfer amounts and verifying compliance invariants hold across 10,000 randomized scenarios.
14. Run Slither analysis and verify contract modularity introduces no delegatecall or storage collision vulnerabilities.
15. Deploy compliance suite to Hyperledger Besu devnet and bind to test `DigitalSecurityToken.sol` instances.

## Interfaces / Contracts

### Identity Registry Interface (`IIdentityRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IIdentityRegistry {
    struct InvestorClaim {
        bytes32 identityId;
        uint16 countryCode; // ISO 3166-1 numeric
        uint8 kycTier;      // 1 = Basic, 2 = Verified Domestic, 3 = Accredited Foreign
        bool isSanctioned;
        uint64 registeredAt;
    }

    event IdentityRegistered(address indexed userAddress, bytes32 indexed identityId, uint16 countryCode);
    event IdentityRemoved(address indexed userAddress, bytes32 indexed identityId);
    event CountryUpdated(address indexed userAddress, uint16 newCountryCode);
    event SanctionStatusUpdated(bytes32 indexed identityId, bool isSanctioned);

    error IdentityAlreadyExists(address userAddress);
    error IdentityNotFound(address userAddress);
    error UnauthorizedAgent();

    function registerIdentity(
        address userAddress,
        bytes32 identityId,
        uint16 countryCode,
        uint8 kycTier
    ) external;

    function deleteIdentity(address userAddress) external;
    function updateSanctionStatus(bytes32 identityId, bool isSanctioned) external;

    function isVerified(address userAddress) external view returns (bool);
    function getInvestorClaim(address userAddress) external view returns (InvestorClaim memory);
    function getInvestorCountry(address userAddress) external view returns (uint16);
}
```

### Modular Compliance Router Interface (`IComplianceRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IComplianceRegistry {
    event ModuleBound(address indexed moduleAddress);
    event ModuleUnbound(address indexed moduleAddress);

    function canTransfer(address from, address to, uint256 amount) external view returns (bool);
    function transferred(address from, address to, uint256 amount) external;
    function created(address to, uint256 amount) external;
    function destroyed(address from, uint256 amount) external;

    function bindModule(address moduleAddress) external;
    function unbindModule(address moduleAddress) external;
    function getModules() external view returns (address[] memory);
}

interface IComplianceModule {
    function moduleCheck(address from, address to, uint256 amount, address complianceRegistry) external view returns (bool);
    function moduleTransferAction(address from, address to, uint256 amount) external;
    function moduleMintAction(address to, uint256 amount) external;
    function moduleBurnAction(address from, uint256 amount) external;
}
```

## Security & Compliance Notes
- **Zero-Bypass Architecture:** Transfer hooks are executed inside the internal `_update` logic of `DigitalSecurityToken.sol`. There is no execution path that allows tokens to move without passing compliance checks.
- **Privacy & GDPR/DPDP Compliance:** No investor PII (name, tax ID, Aadhaar) is written to the blockchain. Identity is verified off-chain; only cryptographic identity hashes and anonymized regulatory claims are anchored on-chain.
- **Immediate Sanctions Freezing:** When a sanction alert triggers off-chain, the compliance relayer updates `updateSanctionStatus(identityId, true)`, which immediately and automatically freezes all associated wallet addresses across all tokens.

## Acceptance Criteria
- [ ] `IdentityRegistry.sol` and `ComplianceRegistry.sol` fully deployed and integrated with `DigitalSecurityToken.sol`.
- [ ] Transfer between two KYC-verified addresses succeeds seamlessly.
- [ ] Transfer reverts immediately if either sender or recipient is not registered or is marked sanctioned.
- [ ] `CountryRestrictModule` successfully rejects transfers to prohibited country codes.
- [ ] `MaxOwnershipModule` successfully reverts transfers that would cause an investor to exceed the defined ownership cap.
- [ ] 100% test coverage achieved in Foundry with zero critical/high Slither findings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Platform Selection), Prompt `303` (Token Issuance Smart Contract), Prompt `004`/`005` (KYC/AML Policies).
- **Parallel Tasks:** Prompt `202` (KYC/AML Backend Service), Prompt `703` (Sanctions Screening).
- **Subsequent Prompts Enabled:** Prompt `306` (Settlement DvP Smart Contract).
