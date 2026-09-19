# 905 - Automated Security Testing in CI/CD (SAST, DAST & Dependency Scanning)

## Purpose
Implements a comprehensive, automated DevSecOps pipeline across all repositories in the Growww platform. Given that Growww processes high-value INR financial transactions, manages investor PII under the DPDP Act, and settles tokenized securities on a permissioned blockchain, all code, container images, smart contracts, and running APIs must undergo continuous security inspection to eliminate vulnerabilities before reaching staging or production environments.

## What You Are Building
An automated security verification pipeline integrated into CI/CD (`.github/workflows/security.yml` and `security/`):
- Static Application Security Testing (SAST) scanning for Rust, Go, Python, TypeScript, and Dart codebases using Semgrep and SonarQube.
- Smart Contract Security Auditing automation using Slither, Mythril, and Echidna property-based fuzzing for Solidity ledger contracts.
- Container image and open-source software (OSS) dependency vulnerability scanning with Trivy and Grype, generating CycloneDX SBOMs.
- Dynamic Application Security Testing (DAST) pipeline using OWASP ZAP targeting staging API Gateway endpoints and Next.js web applications against OWASP API Top 10 vulnerabilities.
- Pre-commit and CI secret scanning utilizing Gitleaks and TruffleHog to guarantee zero credentials, private keys, or API tokens enter version control.

## Scope Boundaries
- **In Scope:**
 - SAST scanning on all pull requests with automated PR blocking on High/Critical CVEs.
 - Smart contract static analysis and automated symbolic execution.
 - Secret scanning on all git commits and history.
 - Container vulnerability scanning on Docker/OCI images before pushing to registry.
 - DAST scanning on staging environment APIs and Web interfaces.
 - Software Bill of Materials (SBOM) generation and license compliance scanning.
- **Out of Scope / Handled Elsewhere:**
 - Manual external penetration testing and bug bounty program management (Prompt 705).
 - Production runtime threat detection and WAF/SIEM monitoring (Prompt 701, 807).

## Technology to Use
- **Polyglot SAST:** Semgrep (OSS + custom financial rulesets) and SonarQube - provides fast, syntax-aware static analysis across Rust, Go, Python, TypeScript, and Dart.
- **Smart Contract Security:** Slither (Trail of Bits) for AST-based vulnerability detection, Mythril for symbolic execution, and Echidna for EVM bytecode fuzzing.
- **Vulnerability & SBOM Scanning:** Trivy (Aqua Security) and Syft (Anchore) for container image scanning, OS packages, and language dependencies.
- **DAST Scanner:** OWASP ZAP (Zed Attack Proxy) automated daemon executed via GitHub Actions against staging API environments.
- **Secret Scanning:** Gitleaks pre-commit hooks and CI scans.

*Justification:* Combining Semgrep, Slither, and Trivy provides rapid, deterministic, developer-friendly security feedback within CI pipelines (under 3 minutes execution), preventing security debt from compounding.

## Backend / Infra Touchpoints
- GitHub Actions CI / GitLab CI runners.
- Docker registry (Amazon ECR / Harbor) with webhook-based vulnerability gating.
- AWS Secrets Manager / HashiCorp Vault for credential rotation testing.
- SonarQube Server / DefectDojo for centralized vulnerability dashboarding and tracking.

## Blockchain Interaction
- Automated CI jobs execute Slither and Mythril analyzers against all Solidity contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`, `MultiSigGovernance.sol`).
- Checks for smart contract specific anti-patterns: Reentrancy vulnerabilities (`reentrancy-eth`, `reentrancy-no-eth`), missing access control modifiers (`unprotected-upgrade`, `arbitrary-send-erc20`), unchecked return values, block timestamp manipulation, and integer overflow/underflow edge cases.
- Executes Echidna property-based fuzzing tests asserting that total circulating digital tokens never exceed the custodian-attested Merkle reserve root under any sequence of transactions.

## Step-by-Step Build Instructions
1. Establish the `security/` directory containing custom rulesets (`security/semgrep/`), contract analysis scripts (`security/smart_contracts/`), and DAST configurations (`security/zap/`).
2. Configure `.pre-commit-config.yaml` to enforce Gitleaks secret scanning and Semgrep pre-commit checks on local developer workstations.
3. Author custom Semgrep rules detecting common financial coding anti-patterns (e.g., floating-point arithmetic for monetary calculations, unencrypted PII in loggers, missing authorization checks).
4. Implement the smart contract security pipeline (`security/smart_contracts/analyze_contracts.sh`) running Slither AST analysis and Mythril symbolic analysis on all PRs modifying `contracts/`.
5. Write Echidna fuzz testing harnesses (`contracts/test/fuzz/`) verifying invariant properties of `SettlementDvP.sol` and `ComplianceRegistry.sol`.
6. Configure Trivy container image scanning in the CI build workflow, failing builds on any `CRITICAL` or `HIGH` vulnerabilities without an available fix.
7. Configure Syft to generate CycloneDX format Software Bill of Materials (SBOM) for every released container artifact.
8. Implement an automated OWASP ZAP DAST scan workflow (`.github/workflows/dast-scan.yml`) that runs against the staging API Gateway every night.
9. Configure OWASP ZAP authentication scripts to obtain staging JWT tokens and systematically probe all authenticated endpoints for BOLA/IDOR, broken authentication, and injection flaws.
10. Integrate DefectDojo / SonarQube to ingest scan results from Semgrep, Slither, Trivy, and ZAP into a single unified security dashboard.
11. Configure automated Slack/PagerDuty security alerts notifying the AppSec lead when new High/Critical vulnerabilities are detected.
12. Establish compliance audit export scripts generating quarterly SEBI CSCRF (Cyber Security and Cyber Resilience Framework) automated security compliance reports.

## Interfaces / Contracts
```yaml
# GitHub Actions Security Pipeline (.github/workflows/security-scan.yml)
name: Security Scan & Vulnerability Gate

on:
  pull_request:
    branches: [ main, develop ]
  schedule:
 - cron: '0 1 * * *' # Daily midnight scan

jobs:
  sast_semgrep:
    name: SAST Code Analysis (Semgrep)
    runs-on: ubuntu-latest
    steps:
 - uses: actions/checkout@v4
 - name: Run Semgrep
        run: |
          docker run --rm -v $(pwd):/src returntocorp/semgrep \
            semgrep scan --config=auto --config=/src/security/semgrep/rules.yaml \
            --error --severity=ERROR

  smart_contract_audit:
    name: Solidity Security Scan (Slither & Mythril)
    runs-on: ubuntu-latest
    steps:
 - uses: actions/checkout@v4
 - name: Run Slither
        uses: crytic/slither-action@v0.3.0
        with:
          target: 'contracts/'
          fail-on: 'high'
          slither-config: 'security/smart_contracts/slither.config.json'

  container_trivy:
    name: Container Vulnerability Scan (Trivy)
    runs-on: ubuntu-latest
    steps:
 - uses: actions/checkout@v4
 - name: Run Trivy Scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          severity: 'CRITICAL,HIGH'
          exit-code: '1'
          ignore-unfixed: true
```

```json
// Slither Custom Configuration (security/smart_contracts/slither.config.json)
{
  "detectors_to_run": "reentrancy-eth,reentrancy-no-eth,unprotected-upgrade,arbitrary-send-erc20,controlled-delegatecall",
  "filter_paths": "node_modules|contracts/mocks",
  "legacy_ast": false
}
```

## Security & Compliance Notes
- Adheres strictly to the SEBI Cyber Security and Cyber Resilience Framework (CSCRF) and RBI Master Directions on IT Governance and Risk Management.
- Zero-tolerance policy on hardcoded API keys, JWT secrets, database passwords, or private keys across all git histories.
- All scanned third-party open-source libraries must have permissive licenses compliant with corporate IP policy (MIT, Apache-2.0, BSD-3; GPL-3 strictly flagged for review).

## Acceptance Criteria
- [ ] Gitleaks and Semgrep scan 100% of pull requests with zero bypassed checks.
- [ ] Slither and Mythril smart contract scans execute cleanly with 0 high/critical findings on all deployed contracts.
- [ ] Trivy container scans block images with unpatched Critical/High CVEs from being pushed to container registries.
- [ ] OWASP ZAP DAST scans run automatically against staging APIs with zero OWASP API Top 10 vulnerabilities detected.
- [ ] Automated CycloneDX SBOM generation is verified for all microservice releases.

## Suggested Order / Dependencies
- **Prerequisites:** 107 (Coding Standards), 303-307 (Smart Contracts), 803 (CI Pipelines).
- **Parallel Tasks:** 701 (Threat Model), 705 (Penetration Testing), 901 (Unit Testing).
