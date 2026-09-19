# 811 - Dual-Environment GitOps CI/CD Pipeline: Testnet vs Mainnet, Safe Bytecode Verification & Rollout Gating

## Purpose
In a regulated fractional equity and digital securities platform, deploying smart contracts and backend infrastructure involves dual financial and legal obligations across domestic Indian markets (SEBI jurisdiction) and IFSC GIFT City international markets (IFSCA jurisdiction). A bug, bytecode mismatch, or misconfigured governance parameter pushed directly to production mainnet can cause irreversible asset loss, invalid ledger states, or severe regulatory sanctions.

This prompt specifies the design and implementation of Growww's Dual-Environment GitOps CI/CD Pipeline. The system creates an automated, cryptographic boundary separating the Regulatory Sandbox / Consortium Testnet from the Production Consortium Mainnet. It establishes automated deterministic bytecode compilation, cryptographic SLSA Level 3 build provenance, automated Anvil/Besu mainnet-fork validation, a mandatory 14-day testnet soak gate, multi-sig timelock proposal generation, and automated rollout kill-switches.

## What You Are Building
An enterprise dual-environment GitOps CI/CD deployment and bytecode verification framework:
- `deployments/cicd/testnet-deploy-pipeline.yaml`: Continuous deployment workflow for ephemeral and persistent Regulatory Sandbox / Consortium Testnet environments.
- `deployments/cicd/mainnet-governance-pipeline.yaml`: Gated, cryptographically attested promotion pipeline for Production Consortium Mainnet (Mumbai & GIFT City clusters).
- `scripts/blockchain/verify-bytecode-reproducibility.sh`: Deterministic Solidity compiler and bytecode verification engine that validates exact runtime bytecode against Git commit hashes, compiler flags, and build metadata.
- `scripts/blockchain/simulate-mainnet-fork.sh`: Automated fork-simulation tool using Anvil / Besu dev nodes to execute proposed smart contract upgrades against real mainnet state snapshots before multi-sig execution.
- `deployments/gitops/policies/rollout-gating-rules.rego`: OPA (Open Policy Agent) Gatekeeper / Conftest policy enforcing cryptographic attestations, minimum soak durations, and multi-signature quorum validations.
- `contracts/governance/generate-multisig-payload.sh`: Deterministic proposal generator formatting calldata for `MultiSigGovernance.sol` with automated parameter drift verification.

## Scope Boundaries
- **In Scope:**
  - Dual-environment separation between Sandbox Testnet (Chain ID `13371`) and Production Mainnet (Chain ID `13370`).
  - Deterministic compilation and reproducible bytecode verification with compiler metadata pinning (Solc `0.8.24`, optimizer runs `200`, `via-ir: true`).
  - Cryptographic artifact attestation (Cosign, in-toto metadata, SLSA Level 3 provenance).
  - Automated fork simulation and dry-run state transition testing.
  - Automated parameter drift detection between Testnet and Mainnet (oracle feeds, fee basis points, circuit breaker thresholds, validator sets).
  - 14-day automated soak gating in the Regulatory Sandbox environment prior to mainnet release eligibility.
  - Multi-sig timelock proposal payload formatting and cryptographic signature collection tooling.
- **Out of Scope / Handled Elsewhere:**
  - Base CI unit/integration test execution (Prompt 803).
  - Progressive canary traffic splitting for Kubernetes workloads (Prompt 804).
  - Smart contract core implementations (Prompts 303, 304, 305, 306, 307).
  - Multi-signature contract on-chain execution logic (Prompt 307).

## Technology to Use
- **GitHub Actions & ArgoCD**: Workflow orchestration and declarative GitOps synchronization engine.
- **Foundry (Forge, Cast, Anvil)**: Deterministic compilation framework, testnet deployment scripting, mainnet-forking simulation, and raw calldata generation.
- **Cosign & Sigstore / in-toto**: Cryptographic supply chain security toolchain providing container signing, binary provenance attestations, and commit signature verification.
- **Open Policy Agent (OPA) / Conftest**: Declarative policy engine validating deployment manifests, metadata hashes, and promotion gate requirements.
- **Hyperledger Besu (QBFT Consensus)**: Target EVM-compatible consortium blockchain powering both Testnet and Mainnet nodes.
- **PostgreSQL & Redis**: Storage backends for deployment audit metadata, soak test metrics, and pipeline tracking.

## Backend / Infra Touchpoints
- **Git Repositories**: Source of truth with branch protections (`main` for testnet auto-deploy, signed release tags `vX.Y.Z` for mainnet promotion).
- **OCI Container Registry (GHCR / AWS ECR)**: Secure artifact storage with Cosign signature and SLSA provenance attachments.
- **Besu Testnet RPC Cluster**: Sandbox cluster endpoint used for automated regression tests and live pilot validation.
- **Besu Mainnet RPC Cluster**: Isolated production cluster endpoint restricted to multi-sig governance execution.
- **ArgoCD Instance**: GitOps controller running in the management cluster with distinct RBAC roles for `testnet` and `mainnet` namespaces.
- **Vault KMS / AWS KMS**: Secure key management holding testnet deployer keys and verifying hardware-backed multi-sig signers.

## Blockchain Interaction
The dual-environment pipeline manages the entire smart contract release lifecycle across Hyperledger Besu networks:
- **Testnet Deployment Lifecycle**: On every merge to `main`, smart contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`) are automatically compiled, tested, and deployed to the Regulatory Sandbox Testnet. Deployment addresses, transaction hashes, and ABIs are committed to `deployments/testnet/contract-manifest.json`.
- **Reproducible Bytecode Verification**: The verification engine compiles the repository contracts in an isolated Docker container with pinned Solc compiler binaries. It retrieves deployed runtime bytecode from the RPC, strips Swarm/IPFS metadata hashes, and asserts a bitwise 100% match with the newly compiled bytecode.
- **Anvil Mainnet Fork Dry-Run**: Before proposing an upgrade to Mainnet, the pipeline creates an Anvil fork of the live Mainnet at current block height. It executes the candidate bytecode upgrade, runs invariant assertion suites against live state (e.g. token balances, custody reserves, whitelist registrations), and measures gas consumption and state integrity.
- **Multi-Sig Payload Generation**: Once verified, the pipeline generates deterministic calldata for `MultiSigGovernance.proposeUpgrade()` and outputs verifiable cryptographic digests for the multi-sig signers (CTO, Head of Compliance, Custodian Trustee, Independent Director).
- **14-Day Testnet Soak Verification**: Promotion to Mainnet requires an automated query verifying that the target bytecode has run continuously in the Regulatory Sandbox Testnet for at least 14 days without unhandled exceptions or rollbacks.

## Step-by-Step Build Instructions
1. Scaffold directories: `deployments/cicd/`, `scripts/blockchain/`, `deployments/gitops/policies/`, and `deployments/metadata/`.
2. Configure Foundry project settings (`foundry.toml`) ensuring pinned EVM version (`cancun`), optimizer enabled (`200` runs), and deterministic compilation flags (`extra_output = ["metadata", "abi", "evm.bytecode", "evm.deployedBytecode"]`).
3. Implement `scripts/blockchain/verify-bytecode-reproducibility.sh` to compile contracts inside a clean Docker container, extract runtime bytecode, fetch on-chain bytecode, and execute bitwise diff verification.
4. Implement `scripts/blockchain/simulate-mainnet-fork.sh` that launches an Anvil node forking Production Besu Mainnet, deploys the upgrade payload via mock governance impersonation, and runs the full test suite against historical mainnet state.
5. Create `deployments/cicd/testnet-deploy-pipeline.yaml` in GitHub Actions to run tests, compile contracts, deploy to Sandbox Testnet, publish ABI packages, and record deployment metadata in PostgreSQL.
6. Create `deployments/cicd/mainnet-governance-pipeline.yaml` triggered by signed release tags. The workflow executes bytecode verification, runs fork simulations, checks the 14-day soak duration in Testnet, and generates multi-sig proposal artifacts.
7. Author OPA policies in `deployments/gitops/policies/rollout-gating-rules.rego` that fail pipeline progression if container images lack Cosign signatures, if bytecode verification fails, or if soak duration is less than 14 days.
8. Implement `contracts/governance/generate-multisig-payload.sh` which formats raw calldata for `MultiSigGovernance.sol` and outputs QR codes and terminal digests for hardware wallet signers.
9. Implement parameter drift detection: compare genesis allocations, block gas limits, chain IDs, and oracle feed addresses between Testnet and Mainnet manifests.
10. Integrate Slack and PagerDuty notification hooks for deployment starts, verification verdicts, multi-sig proposal broadcasts, and governance execution confirmations.
11. Build emergency circuit breaker gating: automate immediate mainnet deployment halt if unexpected anomaly metrics occur during canary phases.
12. Execute full validation run on local test networks to ensure zero manual interventions required during standard verification paths.

## Interfaces / Contracts

```yaml
# deployments/cicd/mainnet-governance-pipeline.yaml
name: Mainnet Smart Contract Promotion Gate

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  verify-and-propose:
    name: Verify Bytecode & Generate Multi-Sig Proposal
    runs-on: ubuntu-latest
    environment: mainnet-governance
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install Foundry Toolchain
        uses: foundry-rs/foundry-toolchain@v1
        with:
          version: nightly

      - name: Verify Git Commit Signature
        run: |
          git verify-tag ${{ github.ref_name }}

      - name: Compile Deterministic Bytecode
        run: |
          docker run --rm -v $(pwd):/workspace -w /workspace \
            ghcr.io/growww/solc-build-env:0.8.24 \
            forge build --extra-output metadata abi evm.bytecode evm.deployedBytecode

      - name: Execute 14-Day Testnet Soak Verification
        env:
          SANDBOX_DB_URL: ${{ secrets.SANDBOX_METRICS_DB_URL }}
        run: |
          ./scripts/blockchain/verify-soak-duration.sh \
            --git-tag ${{ github.ref_name }} \
            --min-days 14 \
            --network testnet

      - name: Run Mainnet Fork Invariant Simulation
        env:
          MAINNET_RPC_URL: ${{ secrets.BESU_MAINNET_RPC_URL }}
        run: |
          ./scripts/blockchain/simulate-mainnet-fork.sh \
            --rpc-url "${MAINNET_RPC_URL}" \
            --contracts "DigitalSecurityToken,SettlementDvP,ComplianceRegistry"

      - name: Generate Multi-Sig Upgrade Proposal Calldata
        id: multisig_proposal
        env:
          MULTISIG_ADDRESS: ${{ vars.MAINNET_MULTISIG_GOVERNANCE_ADDR }}
        run: |
          ./contracts/governance/generate-multisig-payload.sh \
            --tag ${{ github.ref_name }} \
            --output-dir ./build/governance-proposals

      - name: Sign & Attest Deployment Artifacts (SLSA Level 3)
        uses: sigstore/cosign-installer@v3.4.0
      - run: |
          cosign attest --yes \
            --predicate ./build/governance-proposals/proposal-summary.json \
            --type https://in-toto.io/attestation/v1 \
            ghcr.io/growww/smart-contracts:${{ github.ref_name }}

      - name: Publish Governance Action Items
        run: |
          ./scripts/release/publish-governance-bulletin.sh \
            --proposal-file ./build/governance-proposals/proposal-summary.json
```

```bash
# scripts/blockchain/verify-bytecode-reproducibility.sh
#!/usr/bin/env bash
set -euo pipefail

CONTRACT_NAME="${1:-DigitalSecurityToken}"
DEPLOYED_ADDRESS="${2:-0x0000000000000000000000000000000000000000}"
RPC_URL="${3:-http://127.0.0.1:8545}"

echo "=== Verifying Bytecode Reproducibility for ${CONTRACT_NAME} at ${DEPLOYED_ADDRESS} ==="

# 1. Fetch live deployed runtime bytecode from blockchain
FETCHED_BYTECODE=$(cast code "${DEPLOYED_ADDRESS}" --rpc-url "${RPC_URL}" | tr -d '\n' | tr '[:upper:]' '[:lower:]')

# 2. Extract freshly compiled runtime bytecode from local build artifact
LOCAL_BYTECODE="0x$(jq -r '.deployedBytecode.object' "out/${CONTRACT_NAME}.sol/${CONTRACT_NAME}.json" | tr -d '\n' | tr '[:upper:]' '[:lower:]')"

# 3. Strip CBOR metadata hash (last 53 bytes / 106 hex characters) for invariant comparison
FETCHED_STRIPPED="${FETCHED_BYTECODE:0:${#FETCHED_BYTECODE}-106}"
LOCAL_STRIPPED="${LOCAL_BYTECODE:0:${#LOCAL_BYTECODE}-106}"

if [ "${FETCHED_STRIPPED}" != "${LOCAL_STRIPPED}" ]; then
  echo "[-] CRITICAL ERROR: Bytecode mismatch detected!"
  echo "Expected Local Bytecode Length: ${#LOCAL_STRIPPED}"
  echo "Actual Remote Bytecode Length: ${#FETCHED_STRIPPED}"
  diff <(echo "${LOCAL_STRIPPED}") <(echo "${FETCHED_STRIPPED}") || true
  exit 1
fi

echo "[+] SUCCESS: Bytecode bitwise match verified between compiler output and on-chain contract!"
```

```rego
# deployments/gitops/policies/rollout-gating-rules.rego
package growww.governance.gating

default allow = false

# Rule 1: Allow promotion only if SLSA provenance attestation is valid
allow {
    input.attestation.verified == true
    input.attestation.slsa_level >= 3
    input.soak_test.duration_days >= 14
    input.soak_test.unhandled_exceptions == 0
    input.fork_simulation.status == "PASSED"
    input.bytecode_verification.bitwise_match == true
}

# Rule 2: Reject promotion if any unapproved parameter drift exists
deny[msg] {
    input.parameter_drift.detected == true
    msg := sprintf("Parameter drift detected between testnet and mainnet: %v", [input.parameter_drift.fields])
}
```

## Security & Compliance Notes
- **Separation of Network Keys**: Testnet private keys are isolated in GitHub repository secrets for automated deployment, whereas Mainnet multi-sig signer keys are hardware-backed Ledger/Trezor devices or AWS CloudHSM modules requiring physical authorization.
- **SLSA Level 3 Provenance**: All smart contract artifacts and container images must include cryptographic in-toto build attestations generated inside isolated, hardened CI runners to prevent source tampering.
- **Parameter Drift Prevention**: The pipeline strictly enforces that oracle addresses, consensus validator addresses, and chain IDs are explicitly sourced from cryptographically signed configuration manifests rather than environment variables.
- **Zero-Downtime Governance Execution**: Smart contract upgrades must use the UUPS (Universal Upgradeable Proxy Standard) or Beacon proxy pattern, ensuring state variables are preserved without storage layout collision.

## Acceptance Criteria
- [ ] Automated Testnet pipeline compiles, tests, and deploys contracts to Regulatory Sandbox on merges to `main`.
- [ ] Deterministic Docker build environment reproduces identical Solidity bytecode across clean runner instances.
- [ ] `verify-bytecode-reproducibility.sh` accurately strips CBOR metadata and performs bitwise bytecode verification.
- [ ] `simulate-mainnet-fork.sh` forks production Besu node state via Anvil and executes upgrade invariant suites.
- [ ] 14-day continuous testnet soak rule is programmatically verified before mainnet promotion eligibility.
- [ ] OPA Gatekeeper policy blocks any mainnet GitOps synchronization lacking cryptographic Cosign and SLSA attestations.
- [ ] Deterministic multi-sig proposal generator outputs verifiable calldata with zero manual parameter editing.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 108 (Environment Strategy), Prompt 307 (Multi-Sig Governance), Prompt 803 (CI Pipeline Design), Prompt 804 (CD Progressive Delivery).
- **Parallel Tasks:** Prompt 810 (Release Management & Governance), Prompt 812 (Mock Depository & Banking Sandbox Suite).
- **Next Steps:** Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios), Prompt 907 (Regulatory Sandbox Pilot Launch Plan).
