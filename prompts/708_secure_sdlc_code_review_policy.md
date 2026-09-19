# 708 - Secure SDLC, Branch Protection & Code Review Policy

## Purpose
Establishes a mandatory Secure Software Development Life Cycle (S-SDLC), cryptographically enforced branch protection rulesets, peer code review requirements, and automated CI/CD security quality gates across the entire Growww engineering organization. Ensures that no unreviewed, unsigned, or vulnerable code - spanning the Flutter multi-platform codebase, Next.js web applications, polyglot microservices (Go, Python, Rust), and Hyperledger Besu smart contracts - can be merged into protected branches or promoted to staging and production environments, satisfying SEBI CSCRF Section 4 and ISO/IEC 27001 Annex A.14.

## What You Are Building
- **Master Secure SDLC Policy Specification:** `docs/security/secure_sdlc.md` outlining development security standards, threat modeling requirements for new features, code review checklists, and vulnerability remediation SLAs.
- **GitHub Enterprise Repository Rulesets:** Declarative JSON/Terraform configurations enforcing branch protection, GPG/SSH commit signing, minimum 2 peer reviews, and linear Git commit histories on `main` and `release/*` branches.
- **Pre-Commit Security Hook Suite:** Developer pre-commit configuration (`.pre-commit-config.yaml`) running local secret detection (Gitleaks), AST security linters (Bandit, Gosec, Clippy), and formatting checks before commit generation.
- **Automated CI Security Gate Pipeline:** GitHub Actions workflow (`.github/workflows/security-gate.yml`) executing SAST, SCA (Software Composition Analysis), container image scanning, and smart contract verification with automated PR blocking on High/Critical findings.
- **Supply Chain Security & Artifact Signing:** Cosign and Sigstore pipeline signing all production container images with cryptographic SLSA Level 3 provenance attestations.

## Scope Boundaries
- **In Scope:**
 - Git commit signature enforcement (GPG / SSH signing keys linked to verified corporate identities).
 - Multi-party peer review requirements: minimum 2 senior approvals for all microservices; 1 approval must be from a designated Security Engineer for changes touching `auth`, `crypto`, or `contracts/`.
 - Automated security scanners: Semgrep (SAST), Snyk/Trivy (SCA/Dependencies), Gitleaks (Secret Detection), Slither/Mythril (Smart Contract Security).
 - Strict vulnerability remediation SLAs: Critical (24 hours), High (7 days), Medium (30 days), Low (90 days).
 - Software Bill of Materials (SBOM) generation (CycloneDX / SPDX).
- **Out of Scope / Handled Elsewhere:**
 - CI runner deployment and Kubernetes infrastructure pipelines (handled in Prompt 803).
 - Third-party external penetration testing and bug bounty management (handled in Prompt 705).
 - Production deployment orchestration and canary rollouts (handled in Prompt 804).

## Technology to Use
- **Version Control & Rules:** GitHub Enterprise / GitLab Ultimate repository rulesets with branch protection policies.
- **Static Analysis (SAST):** Semgrep Enterprise (custom tailored rulesets for financial logic) and SonarQube.
- **Dependency & Container Scanning (SCA):** Aqua Trivy and Snyk scanning npm, pip, cargo, and go dependencies against CVE databases.
- **Secret Detection:** Gitleaks and TruffleHog integrated into both local pre-commit hooks and server-side push filters.
- **Smart Contract Analyzers:** Slither (static analysis) and Mythril (symbolic execution).
- **Artifact Signing & Provenance:** Sigstore Cosign with Kyverno admission controllers on Kubernetes.
- **Justification:** Shift-left security automation prevents secrets, vulnerable dependencies, and logic flaws from ever entering the Git commit history, reducing remediation costs by over 90% compared to post-deployment bug fixing.

## Backend / Infra Touchpoints
- **Source Repositories:** Monorepo / Polyrepos across all Growww services (106).
- **CI/CD Infrastructure:** GitHub Actions Runners (803), Harbor Container Registry, Kubernetes Clusters (802).
- **Vulnerability Tracking:** DefectDojo (705), Jira Software Security Board.

## Blockchain Interaction
Mandatory automated security quality gates for all Hyperledger Besu smart contracts:
- **Zero-Tolerance Static Analysis Gate:** Pull requests modifying Solidity contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `MultiSigGovernance.sol`) must pass 100% of Slither and Mythril security suites with zero unresolved High/Medium findings.
- **Property-Based Fuzzing & Invariant Tests:** CI runs Echidna and Foundry invariant fuzzing tests for a minimum of 500,000 runs to prove mathematical conservation of fractional token balances.
- **Smart Contract Release Multi-Sig Gate:** Smart contract deployment bytecode artifacts must be cryptographically signed by both the Lead Blockchain Architect and the Head of Security before deployment to the testnet or mainnet relayer.

## Step-by-Step Build Instructions
1. Author the Master Secure SDLC Policy under `docs/security/secure_sdlc.md`, defining security milestones across Requirements, Design, Development, Testing, and Deployment phases.
2. Establish the Code Review Checklist: input validation, authentication/authorization checks, SQL injection prevention, safe arithmetic (integer overflow/underflow), cryptographic correctness, and zero PII logging.
3. Configure GitHub Enterprise Branch Protection Rulesets for `main` and `release/*`:
 - Require linear commit history (no merge commits; rebase or squash).
 - Require pull request reviews (minimum 2 approvals; code owners review required).
 - Dismiss stale pull request approvals when new commits are pushed.
 - Require signed commits (reject any unsigned commits).
 - Require all status checks to pass before merging.
 - Block force pushes and branch deletions for all users (including repository administrators).
4. Implement the root `.pre-commit-config.yaml` enforcing local execution of Gitleaks, Prettier, Black/Ruff, and Gosec before commit creation.
5. Create `.github/workflows/secret-scanning.yml` to execute Gitleaks and TruffleHog across all branches and commit histories on push and pull request events.
6. Configure Semgrep SAST workflow (`.github/workflows/sast.yml`) with custom rules preventing insecure crypto (e.g. MD5, DES), raw SQL string formatting, and unprotected API endpoints.
7. Configure Trivy SCA workflow (`.github/workflows/dependency-scan.yml`) generating CycloneDX SBOMs and failing PR builds if any Critical or High severity CVE without an approved exception is found.
8. Implement smart contract CI workflow (`.github/workflows/smart-contract-security.yml`) executing Hardhat/Foundry unit tests (100% branch coverage required) and Slither static analysis.
9. Implement container image signing in CD pipelines using Sigstore Cosign: sign image digests and publish SBOM attestations to the container registry.
10. Deploy Kyverno admission policies on staging and production Kubernetes clusters, blocking pods from launching if container images lack a valid Cosign signature.
11. Build an automated Jira integration bot: automatically open Jira security tickets when new CVEs are discovered in production dependencies and track remediation SLAs.
12. Establish an automated quarterly developer secure coding training program and track completion across the engineering organization.

## Interfaces / Contracts
```yaml
# .pre-commit-config.yaml
repos:
 - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.2
    hooks:
 - id: gitleaks
 - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.3.0
    hooks:
 - id: ruff
        args: [--fix, --exit-non-zero-on-fix]
 - repo: https://github.com/crytic/slither
    rev: 0.10.1
    hooks:
 - id: slither
        files: ^contracts/.*\.sol$
```

```json
// github_branch_ruleset.json
{
  "name": "Production and Release Branch Protection",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["refs/heads/main", "refs/heads/release/*"],
      "exclude": []
    }
  },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "required_signatures" },
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 2,
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": true,
        "require_last_push_approval": true
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "strict_required_status_checks_policy": true,
        "required_status_checks": [
          { "context": "security-gate / sast-semgrep" },
          { "context": "security-gate / sca-trivy" },
          { "context": "security-gate / secret-scan-gitleaks" },
          { "context": "security-gate / slither-contracts" },
          { "context": "tests / unit-and-integration" }
        ]
      }
    }
  ]
}
```

## Security & Compliance Notes
- **SEBI CSCRF Section 4 (SDLC Security):** Requires institutional intermediaries to enforce documented secure coding standards, independent code reviews, vulnerability scanning before release, and artifact integrity protection.
- **ISO/IEC 27001 Annex A.14.2:** Enforces security throughout development and support processes, including separation of development, test, and operational environments.
- **SLSA Level 3 Compliance:** Guarantees end-to-end supply chain integrity, preventing tampering with source code, build pipelines, and production release binaries.

## Acceptance Criteria
- [ ] Master Secure SDLC policy published under `docs/security/secure_sdlc.md`.
- [ ] GitHub branch protection rulesets active on `main` and `release/*` with signed commits and 2-reviewer approvals enforced.
- [ ] Pre-commit hooks (`.pre-commit-config.yaml`) configured and verified across local dev environments.
- [ ] CI security workflows (SAST, SCA, Secret Scanning, Smart Contract Security) active and blocking PRs on Critical/High violations.
- [ ] Slither and Foundry invariant tests execute on smart contract PRs with 100% test coverage requirement.
- [ ] Container images signed with Cosign; Kubernetes Kyverno admission controller blocks unsigned images in staging.
- [ ] Automated CycloneDX SBOM generation and CVE tracking operational.

## Suggested Order / Dependencies
- **Prerequisites:** 106 (Repository Convention), 107 (Coding Standards), 803 (CI Pipeline Design).
- **Parallel Tasks:** 701 (Threat Model), 705 (Pentesting Plan).
- **Downstream Dependents:** 804 (CD Pipeline), 901 (Testing Strategy), 905 (Security Testing Automation).
