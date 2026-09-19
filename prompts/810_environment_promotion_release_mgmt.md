# 810 - Environment Promotion, Release Management & Governance Workflows

## Purpose
In a regulated securities and blockchain infrastructure environment, code releases, database migrations, infrastructure upgrades, and smart contract deployments carry profound financial and regulatory obligations. Direct, uncontrolled production deployments are prohibited by SEBI cyber resilience frameworks and institutional custody standards.

This prompt formalizes Growww's end-to-end environment promotion lifecycle (Local -> Dev -> Staging -> UAT/Sandbox -> Production Mumbai & GIFT City). It establishes cryptographic artifact provenance, semantic release tagging, automated zero-downtime database migration governance (expand/contract schema pattern), multi-party authorization for on-chain smart contract upgrades (`MultiSigGovernance.sol` with 48-hour timelocks), and digital Change Advisory Board (CAB) sign-offs.

## What You Are Building
A formalized release governance and automated environment promotion pipeline:
- `deployments/release/promotion-pipeline.yaml`: Master environment promotion pipeline managing gated progression across environments based on test verdicts and cryptographic signatures.
- `scripts/release/version-bump.sh`: Automated Semantic Versioning (SemVer 2.0) tool evaluating Conventional Commits to generate version tags, release notes, and changelogs.
- `deployments/db/migrations/`: Database migration governance framework using Flyway / Goose with strict forward/backward compatibility checks (Expand/Contract schema pattern).
- `contracts/governance/timelock-upgrade-proposal.sh`: Multi-sig smart contract upgrade proposal script that generates deterministic proposal payloads, verifies contract bytecode against verified source repositories, and submits proposals to `MultiSigGovernance.sol`.
- `docs/release/cab_standard_operating_procedure.md`: Change Advisory Board (CAB) SOP outlining pre-release checklists, rollback criteria, regulatory notification rules, and post-deployment sanity verification.

## Scope Boundaries
- **In Scope:**
 - Standardized environment progression: Dev -> Staging -> UAT/Regulatory Sandbox -> Production (Mumbai & GIFT City).
 - Semantic versioning, release tagging, and automated changelog generation.
 - Zero-downtime database schema migration lifecycle with automated dry-runs and rollback plans.
 - Smart contract deployment and parameter governance via on-chain Multi-Sig and Timelock contracts.
 - Cryptographic artifact verification (Cosign container signatures, Git commit signatures).
 - Emergency hotfix fast-track protocol with post-facto audit review.
- **Out of Scope / Handled Elsewhere:**
 - CI test matrix execution (Prompt 803).
 - ArgoCD progressive canary rollout execution (Prompt 804).
 - Smart contract multi-sig contract implementation (Prompt 307).

## Technology to Use
- **GitHub Actions & ArgoCD**: Release promotion orchestrators. Justification: Native integration with Git branch models, environment protection rules, and GitOps synchronization engines.
- **Release-Please / Semantic Release**: Automated versioning and changelog generator based on Conventional Commits standard.
- **Flyway / Goose**: Deterministic database migration tools. Justification: Enforces versioned, immutable SQL migration scripts with checksum validation and automated rollback script verification.
- **Foundry (Forge/Cast)**: Used for compiling, verifying, and submitting on-chain multi-sig upgrade transactions.
- **Cosign & Sigstore**: Verifies container image signatures before promoting images to staging or production Kubernetes manifests.

## Backend / Infra Touchpoints
- **Git Repositories**: Source of truth with signed GPG commits and protected release tags (`vX.Y.Z`).
- **Container Registry**: Promotion between staging and production repositories based on Cosign signature verification.
- **PostgreSQL / Aurora Databases**: Target databases for Flyway migrations.
- **Hyperledger Besu Production RPC**: Target blockchain endpoint for smart contract timelock proposals.
- **ServiceNow / Jira Service Management**: Automated CAB ticket generation and approval logging.

## Blockchain Interaction
Establishes a rigorous on-chain release governance protocol for all smart contract upgrades:
- **UAT Sandbox Staging**: All smart contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`) must be deployed and validated in the SEBI/IFSCA Regulatory Sandbox test environment for a minimum of 14 continuous days without error before production deployment.
- **Multi-Sig & Timelock Protocol**: Production smart contract upgrades, parameter modifications (e.g. fee basis points or whitelist authority), or contract pause actions cannot be executed by any single administrator. Proposals must be submitted to `MultiSigGovernance.sol` requiring M-of-N threshold signatures (e.g. 3 of 5 keys: CTO, Head of Compliance, Principal Architect, Custodian Trustee, Independent Director).
- **48-Hour Timelock Enforcement**: Non-emergency contract upgrades enforce a mandatory 48-hour on-chain timelock delay between proposal approval and execution, giving regulators and market participants time to inspect proposed bytecode.

## Step-by-Step Build Instructions
1. Scaffold directory `deployments/release/`, `scripts/release/`, `deployments/db/migrations/`, and `docs/release/`.
2. Configure GitHub repository Environment Protection Rules requiring designated code owners and QA leads to approve promotions to `staging` and `production`.
3. Implement `scripts/release/version-bump.sh` integrating Release-Please to analyze `feat:`, `fix:`, and `perf:` commits and generate automated release tags (`v1.2.0`).
4. Establish database migration guidelines requiring the 3-step Expand/Contract pattern: (1) Expand: Add new nullable columns or tables, (2) Migrate: Dual-write application data, (3) Contract: Remove old unused columns in a subsequent release.
5. Create Flyway CI validation step in GitHub Actions that spins up a transient PostgreSQL container, runs all historic migrations, applies new migration, runs the rollback script, and re-applies to prove migration reversibility.
6. Author the automated CAB integration workflow generating a standardized Change Request ticket detailing Git commit diffs, database migration impact, test results, and rollback runbooks.
7. Write `contracts/governance/timelock-upgrade-proposal.sh` that compiles the updated smart contract bytecode, executes `forge verify-contract`, and creates a proposal on `MultiSigGovernance.sol`.
8. Configure Cosign image verification in the GitOps promotion pipeline: ArgoCD rejects any image tag whose cryptographic signature does not originate from the authorized GitHub CI OIDC builder.
9. Implement the emergency hotfix procedure: allows expedited promotion for P1 security/trading fixes with automatic escalation to C-level executives and post-deployment audit review within 24 hours.
10. Test full end-to-end promotion: create a PR -> merge to main -> automated tag generation -> automatic staging deploy -> QA sign-off -> manual CAB approval -> canary production rollout.
11. Test database migration rollback under simulated load to confirm zero locking or downtime.
12. Conduct a mock smart contract multi-sig upgrade rehearsal on the UAT network.
13. Publish the CAB Standard Operating Procedure to internal compliance portals.

## Interfaces / Contracts
```yaml
# deployments/release/promotion-pipeline.yaml (Excerpt)
name: Production Promotion Gate

on:
  workflow_dispatch:
    inputs:
      release_version:
        description: 'Release Tag to Promote (e.g. v1.4.0)'
        required: true
      cab_ticket_id:
        description: 'Approved CAB Ticket ID (e.g. CAB-2026-8941)'
        required: true

jobs:
  promote-to-production:
    runs-on: ubuntu-latest
    environment: production-mumbai
    steps:
 - uses: actions/checkout@v4
      
 - name: Verify Container Cosign Signatures
        uses: sigstore/cosign-installer@v3.4.0
 - run: |
          cosign verify --certificate-identity-regexp="https://github.com/growww/.*" \
                        --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
                        ghcr.io/growww/matching-engine:${{ github.event.inputs.release_version }}

 - name: Validate CAB Ticket Approval
        run: |
          ./scripts/release/verify-cab-ticket.sh ${{ github.event.inputs.cab_ticket_id }}

 - name: Trigger GitOps Overlay Update
        run: |
          ./scripts/gitops/promote-release.sh --env=prod-mumbai \
                                             --version=${{ github.event.inputs.release_version }} \
                                             --cab=${{ github.event.inputs.cab_ticket_id }}
```

```bash
# contracts/governance/timelock-upgrade-proposal.sh (Excerpt)
#!/usr/bin/env bash
set -euo pipefail

CONTRACT_NAME=$1
NEW_IMPLEMENTATION_ADDRESS=$2
DESCRIPTION=$3

echo "Submitting upgrade proposal for ${CONTRACT_NAME} to MultiSig Governance..."
cast send "${MULTISIG_GOVERNANCE_ADDRESS}" \
  "proposeUpgrade(string,address,string)" \
  "${CONTRACT_NAME}" \
  "${NEW_IMPLEMENTATION_ADDRESS}" \
  "${DESCRIPTION}" \
  --rpc-url "${BESU_RPC_URL}" \
  --private-key "${PROPOSER_KEY}"

echo "Upgrade proposal submitted. 48-hour timelock countdown initiated."
```

## Security & Compliance Notes
- Separation of Duties: Engineers authoring code cannot approve their own production promotions; dual sign-off is enforced by GitHub branch protection and environment approval gates.
- SEBI Cyber Resilience Compliance: All production changes, database alterations, and contract upgrades are recorded in tamper-evident logs (Prompt 807) and cross-referenced with authorized change tickets.
- Emergency Hotfix Accountability: Hotfixes bypass standard scheduling but require retroactive CAB audit approval and an incident post-mortem within 24 hours.

## Acceptance Criteria
- [ ] End-to-end environment progression (Dev -> Staging -> UAT -> Prod) is governed by automated promotion workflows.
- [ ] Container image signatures are cryptographically verified via Cosign before deployment to production clusters.
- [ ] Database migrations adhere to the Expand/Contract pattern and have verified, automated rollback scripts.
- [ ] Smart contract upgrades enforce M-of-N multi-sig approvals and 48-hour timelock delays via `MultiSigGovernance.sol`.
- [ ] Release notes, changelogs, and SemVer tags are automatically generated from Conventional Commits.
- [ ] Emergency hotfix protocol is fully documented with mandatory compliance audit trail generation.

## Suggested Order / Dependencies
- Prerequisites: Prompt 108 (Environment Strategy), Prompt 307 (Multi-Sig Governance), Prompt 312 (Chain Upgrade Process), Prompt 803 (CI Pipeline), Prompt 804 (CD Pipeline).
- Parallel Tasks: Prompt 906 (UAT Plan), Prompt 908 (Production Launch Runbook).
