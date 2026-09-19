# 312 - Smart Contract Proxy Upgrades & Blockchain Protocol Hardfork Management

## Purpose
Financial market infrastructure must evolve to support new regulatory mandates, novel asset classes, enhanced compliance logic, and performance optimizations. However, in a multi-stakeholder consortium ledger governing real-world equities, modifications to smart contracts or underlying node protocols cannot occur arbitrarily. Upgrades must be strictly deterministic, backward-compatible, resistant to storage collision vulnerabilities, and subject to institutional governance approvals with verifiable audit trails.

This prompt specifies the end-to-end framework and tooling for: (1) **Smart Contract Upgradeability Architecture** using **ERC-1967 Transparent & UUPS Proxy Patterns** governed by `MultiSigGovernance.sol` and `TimeLockController.sol`, and (2) **Protocol Hardfork & Besu Client Upgrade Playbooks** ensuring coordinated, zero-downtime consensus migrations across all consortium participant nodes.

## What You Are Building
A governance and upgrade tooling suite under `contracts/governance/upgrades/` and `ops/chain-upgrades/` containing:
- Proxy Infrastructure: Standardized ERC-1967 Proxy implementations and `ProxyAdmin.sol` controllers.
- Automated Storage Layout Validator: CLI and CI tool (using Foundry / OpenZeppelin Upgrades) that compares storage layouts between $V_N$ and $V_{N+1}$ to prevent catastrophic storage slot collisions.
- Upgrade Orchestration Scripts: Automated Foundry deployment and timelock proposal generation scripts (`script/UpgradeContract.s.sol`).
- Protocol Hardfork Runbook & Validator Genesis Migration Suite (`ops/chain-upgrades/hardfork_runbook.md`): Step-by-step procedures for activating EVM fork transitions (e.g. Shanghai to Cancun/Prague) in Besu with zero block stalls.

## Scope Boundaries
- **In Scope:**
 - Standardized proxy architecture across all ecosystem contracts (`DigitalSecurityToken`, `SettlementDvP`, `ComplianceRegistry`, `ProofOfReserveRegistry`).
 - Storage gap conventions (`uint256[50] private __gap`) and layout verification.
 - Integration with `MultiSigGovernance.sol` and `TimeLockController.sol` for proposal submission, mandatory timelock delays, and execution.
 - Rollback and emergency freeze procedures during faulty upgrades.
 - Besu node client binary upgrade orchestration across consortium participants.
- **Out of Scope / Handled Elsewhere:**
 - Multisig smart contract core implementation (handled in Prompt 307).
 - Continuous CI/CD deployment pipelines (handled in Prompt 804).
 - Admin governance web UI (handled in Prompt 605).

## Technology to Use
- **Proxy Standard:** **ERC-1967 Transparent Proxy / Universal Upgradeable Proxy Standard (UUPS)** via OpenZeppelin Contracts Upgradeable v5.0.
  *Justification:* ERC-1967 standardizes storage slots (`bytes32(uint256(keccak256('eip1967.proxy.implementation')) - 1)`) to avoid collision with logic contract variables. UUPS reduces gas overhead on runtime calls while concentrating upgrade authorization logic in the implementation contract.
- **Verification Tooling:** Foundry (`forge inspect storage-layout`), OpenZeppelin Upgrades CLI / Plugins, Slither.
- **Coordination Tooling:** Ansible / Kubernetes Helm for rolling Besu client updates.

## Backend / Infra Touchpoints
- **Admin Portal & Governance UI:** Compliance officers review implementation bytecode diffs and approve upgrade proposals (Prompt 605).
- **Admin Service:** Microservice (Prompt 217) monitoring timelock ETAs and triggering proxy upgrade transactions upon expiry.
- **Event Indexer:** Service (Prompt 309) detecting `Upgraded(address indexed implementation)` events to update ABI decoding caches on the fly.

## Blockchain Interaction
- **Upgrade Execution:** Triggered by calling `upgradeToAndCall(address newImplementation, bytes data)` on the proxy contract via `MultiSigGovernance.sol`.
- **Event Emission:** Emits standard ERC-1967 event `Upgraded(address indexed newImplementation)`.
- **Zero State Loss:** All balances, KYC registries, nonces, and historical records remain in the persistent proxy storage; only the executing logic bytecode is redirected.

## Step-by-Step Build Instructions
1. Scaffold upgrade repository structure under `contracts/governance/upgrades/` and `ops/chain-upgrades/`.
2. Define base upgradeable abstract contract `GrowwwUpgradeable.sol` inheriting `Initializable`, `ContextUpgradeable`, and `UUPSUpgradeable`.
3. Reserve 50 storage slots at the end of every stateful contract: `uint256[50] private __gap;` to ensure future variable additions do not corrupt storage offsets.
4. Implement `_authorizeUpgrade(address newImplementation)` in all upgradeable contracts, restricting caller strictly to `MultiSigGovernance.sol`.
5. Create automated storage layout comparison tool in Python / Rust (`tools/storage_diff.py`):
 - Compares JSON output of `forge inspect ContractV1 storage` against `forge inspect ContractV2 storage`.
 - Validates that existing variable slot indexes, offsets, and types are 100% preserved.
 - Throws error if any variable is reordered, deleted, or altered in size.
6. Integrate storage diff check into GitHub Actions CI pipeline to block pull requests with breaking storage layout mutations.
7. Implement upgrade script `script/ProposeUpgrade.s.sol`:
 - Deploys new implementation contract $V_2$ to Hyperledger Besu.
 - Generates encoded calldata for `upgradeToAndCall(address(v2), reinitData)`.
 - Submits upgrade proposal to `MultiSigGovernance.sol` with `ProposalCategory.CONTRACT_UPGRADE`.
 - Emits proposal ID and timelock ETA timestamp.
8. Implement verification script `script/ExecuteUpgrade.s.sol` that executes proposal once timelock delay (e.g. 48 hours) has elapsed and required multisig confirmations are collected.
9. Implement post-upgrade smoke test suite verifying contract storage variables, token balances, and compliance hooks remain intact.
10. Define emergency rollback playbook: if $V_2$ exhibits a critical defect post-upgrade, the multisig can submit an immediate emergency upgrade to deploy patch $V_{2\text{-patch}}$ or revert proxy implementation pointer to verified $V_1$.
11. Draft consortium protocol hardfork coordination procedure (`ops/chain-upgrades/hardfork_runbook.md`):
 - Agree on activation block height $H_{\text{fork}}$ across all consortium partners (Growww, Custodian, Clearing, GIFT City).
 - Distribute updated Besu node binaries (e.g. v24.1.0 to v24.4.0) with updated `genesis.json` / fork configurations.
 - Perform rolling upgrade of validator nodes one by one at least 48 hours prior to $H_{\text{fork}}$.
 - Monitor block production at block $H_{\text{fork}}$ in Prometheus/Grafana to verify zero round stalls or consensus forks.
12. Execute simulated upgrade and protocol fork tests on local 4-validator testnet harness.

## Interfaces / Contracts

### Upgradeable Base Contract Template (`GrowwwUpgradeable.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";

abstract contract GrowwwUpgradeable is Initializable, AccessControlUpgradeable, UUPSUpgradeable {
    bytes32 public constant UPGRADE_ROLE = keccak256("UPGRADE_ROLE");

    event ContractUpgraded(address indexed newImplementation, uint256 version, address indexed upgradedBy);

    function __GrowwwUpgradeable_init(address adminMultisig) internal onlyInitializing {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        _grantRole(DEFAULT_ADMIN_ROLE, adminMultisig);
        _grantRole(UPGRADE_ROLE, adminMultisig);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADE_ROLE) {
        // Upgrade permitted only by MultiSigGovernance contract
    }

    // Reserved storage space to allow for layout expansion in future versions
    uint256[50] private __gap;
}
```

### Storage Diff Validator Configuration Schema (`storage-diff-config.json`)
```json
{
  "contracts": [
    {
      "name": "DigitalSecurityToken",
      "v1_path": "contracts/tokens/v1/DigitalSecurityToken.sol",
      "v2_path": "contracts/tokens/DigitalSecurityToken.sol"
    },
    {
      "name": "SettlementDvP",
      "v1_path": "contracts/settlement/v1/SettlementDvP.sol",
      "v2_path": "contracts/settlement/SettlementDvP.sol"
    }
  ],
  "rules": {
    "allow_variable_append": true,
    "allow_variable_type_mutation": false,
    "allow_variable_rename": false,
    "enforce_gap_decrement": true
  }
}
```

## Security & Compliance Notes
- **Mandatory Timelock & Multi-Party Review:** No smart contract upgrade can be executed instantaneously. A mandatory 48-hour timelock period ensures SEBI, institutional custodians, and security auditors can inspect bytecode diffs before activation.
- **Storage Collision Prevention:** Automated storage layout validation is an absolute hard gate in CI. Corrupting proxy storage slots would lead to permanent loss of investor balance records.
- **Immutable Historical Logs:** While smart contract logic can be upgraded, all historical transaction receipts, transfer events, and proof-of-reserve records remain permanently immutably preserved in the underlying Besu blockchain ledger.

## Acceptance Criteria
- [ ] Base upgradeable template (`GrowwwUpgradeable.sol`) implemented and integrated into all core contracts.
- [ ] Storage layout validation tool successfully detects incompatible variable mutations and prevents deployment.
- [ ] Upgrade workflow tested end-to-end on Besu: $V_1 \rightarrow V_2$ migration preserves 100% of state variables, balances, and mappings.
- [ ] MultiSig + Timelock enforcement verified: upgrade call from unauthorized address or before timelock expiry reverts deterministically.
- [ ] Protocol hardfork runbook tested on testnet with zero consensus interruption at activation height.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance), Prompt `306` (Settlement DvP), Prompt `307` (Multisig Governance).
- **Parallel Tasks:** Prompt `804` (CD Pipeline Design), Prompt `810` (Release Management Process).
- **Subsequent Prompts Enabled:** Prompt `314` (Chain Disaster Recovery & Backup).
