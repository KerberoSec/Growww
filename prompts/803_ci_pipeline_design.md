# 803 - Polyglot CI Pipeline Architecture & Automated Verification Matrix

## Purpose
In a mission-critical financial system where fractional equity tokens represent real regulated Indian securities, code quality, static security analysis, deterministic compilation, and comprehensive automated testing are paramount. The continuous integration (CI) pipeline guarantees that no pull request can be merged without passing exhaustive unit tests, integration suites, static analysis, container vulnerability scans, smart contract formal checks, and software bill of materials (SBOM) generation.

This prompt defines the enterprise polyglot CI matrix across GitHub Actions / GitLab CI for Growww's multi-language stack (Rust, Go, Python, TypeScript, Dart/Flutter, and Solidity), enforcing sub-10-minute feedback loops via aggressive layer caching, parallelized matrix runners, and strict regulatory compliance gates.

## What You Are Building
Declarative CI workflow definitions, caching strategies, and security analysis scripts:
- `.github/workflows/ci-pr-gate.yml`: Master pull request validation pipeline enforcing linting, unit tests, code coverage thresholds (>=85%), and secret scanning before PR review.
- `.github/workflows/ci-smart-contracts.yml`: Dedicated blockchain contract pipeline executing Foundry test suites, invariant fuzzing, gas benchmarks, Slither static analysis, and bytecode determinism checks.
- `.github/workflows/ci-backend-matrix.yml`: Polyglot test matrix covering Rust (Order Matching Engine), Go (Trade Settlement, API Gateway), Python (KYC, User Service), and TypeScript (Admin/Web).
- `.github/workflows/ci-flutter-client.yml`: Multi-platform Flutter CI testing unit/widget suites, analyzer rules, and compiling headless test binaries for Android, iOS, and Web.
- `.github/workflows/ci-container-publish.yml`: Secure container image build pipeline using Docker Buildx, BuildKit caching, Trivy vulnerability scanning, Syft SBOM generation, and Cosign cryptographic signing.
- `scripts/ci/check-coverage.sh`: Coverage threshold enforcer comparing PR branch coverage against main base branch to prevent coverage regression.

## Scope Boundaries
- **In Scope:**
 - Fast PR verification pipelines for all backend languages (Rust, Go, Python), Frontend/Mobile (Dart, TS), and Smart Contracts (Solidity).
 - Smart contract test execution (Foundry Forge), Slither static analysis, and gas cost diffing.
 - Secret scanning (Gitleaks/TruffleHog), SAST (Semgrep/SonarQube), and container CVE scanning (Trivy).
 - Multi-platform container building with BuildKit and multi-arch support (`linux/amd64`, `linux/arm64`).
 - Cryptographic artifact signing via Cosign and SBOM generation via Syft.
- **Out of Scope / Handled Elsewhere:**
 - Progressive continuous delivery and deployment rollouts (Prompt 804).
 - Dynamic Application Security Testing (DAST) on live staging clusters (Prompt 905).
 - Production release governance and multi-sig promotions (Prompt 810).

## Technology to Use
- **GitHub Actions (with Actions Runner Controller on K8s)**: Scalable, declarative CI orchestrator. Justification: Native integration with GitHub repositories, rich ecosystem of official actions, and ability to run ephemeral, self-hosted container runners inside our private AWS/GCP Kubernetes clusters for high build performance and compliance data isolation.
- **Foundry (Forge/Cast)**: Solidity testing framework. Justification: Executes thousands of smart contract tests and property fuzzing runs in seconds directly in Rust-backed EVM simulation, with built-in gas reporting.
- **Slither & Mythril**: Static analysis tools for Solidity smart contracts.
- **Docker Buildx & BuildKit**: Container builder with remote registry caching (GitHub Cache / Harbor).
- **Trivy (Aqua Security)**: Fast, comprehensive vulnerability and misconfiguration scanner for container images and IaC.
- **Cosign (Sigstore)**: Keyless container signing using OpenID Connect (OIDC) identities.
- **Semgrep & SonarQube**: AST-based static application security testing (SAST) for backend and frontend codebases.

## Backend / Infra Touchpoints
- **Self-Hosted K8s Runners**: Auto-scaling runner pods in `growww-ci-runners` namespace with attached NVMe cache volumes.
- **Container Registry**: AWS ECR / Harbor / GitHub Container Registry with immutable image tag policies.
- **SonarQube / Codecov**: Centralized quality dashboard storing coverage metrics and code debt analytics.
- **Vault / AWS OIDC**: Tokenless authentication between GitHub Actions and cloud providers.

## Blockchain Interaction
The smart contract CI pipeline (`ci-smart-contracts.yml`) validates the integrity of all permissioned ledger contracts:
- **Foundry Invariant & Fuzz Testing**: Runs 50,000 fuzz runs on `SettlementDvP.sol` ensuring that token transfers never execute if fiat payment validation fails, and verifies ERC-3643 transfer whitelist invariants in `DigitalSecurityToken.sol`.
- **Gas Benchmark Regressions**: Generates automated gas snapshot diffs on every PR against the base branch to catch expensive storage operations.
- **Slither Static Analysis**: Scans contracts for reentrancy, uninitialized storage pointers, arbitrary senders, and unchecked external calls, failing the build on any high or medium severity finding.
- **Deterministic Bytecode Verification**: Compiles contracts with locked `solc` version (`0.8.24`) and optimizer runs (`200`), verifying output bytecode hashes match expected staging deployment artifacts.

## Step-by-Step Build Instructions
1. Scaffold directory `.github/workflows/` and `.github/actions/` containing modular composite actions.
2. Configure OIDC trust relationship between GitHub repository and AWS IAM / GCP Workload Identity to eliminate static cloud access keys.
3. Write `ci-smart-contracts.yml` configuring Foundry setup, Forge unit testing (`forge test -vvv`), fuzz testing (`forge test --fuzz-runs 10000`), and Slither analysis (`slither . --sarif slither.sarif`).
4. Write `ci-backend-matrix.yml` with separate parallel jobs for:
 - **Rust**: `cargo fmt --check`, `cargo clippy -- -D warnings`, `cargo test --all-targets`, `cargo audit`.
 - **Go**: `golangci-lint run`, `go test -race -coverprofile=coverage.out ./...`, `govulncheck ./...`.
 - **Python**: `black --check .`, `flake8 .`, `mypy --strict .`, `pytest --cov=. tests/`.
5. Write `ci-flutter-client.yml` running `flutter analyze --fatal-infos`, `flutter test --coverage`, and building release artifacts in dry-run mode.
6. Write `ci-web.yml` running `eslint`, `prettier --check`, `tsc --noEmit`, and `next build`.
7. Implement Gitleaks and TruffleHog secret scanning step in `ci-pr-gate.yml` scanning the entire commit history of the PR.
8. Configure Semgrep SAST scanning with OWASP Top 10 and CWE rulesets.
9. Implement container build workflow `ci-container-publish.yml` using `docker/build-push-action` with multi-stage targets and BuildKit GitHub Actions cache backend (`type=gha`).
10. Integrate Trivy container image scanning to fail the pipeline on any `CRITICAL` or `HIGH` unfixed CVEs.
11. Implement Syft SBOM generation (`syft scan <image> -o spdx-json`) and attach the SBOM to the build release.
12. Integrate Cosign container signing using GitHub OIDC identity (`cosign sign <image-digest>`).
13. Write `scripts/ci/check-coverage.sh` to enforce >=85% line coverage and post automated PR comments with test breakdowns.
14. Configure GitHub Branch Protection Rules requiring all matrix checks to pass before merging into `main`.

## Interfaces / Contracts
```yaml
# .github/workflows/ci-smart-contracts.yml (Excerpt)
name: Smart Contract Verification Matrix

on:
  pull_request:
    paths:
 - 'contracts/**'
 - '.github/workflows/ci-smart-contracts.yml'
  push:
    branches: [main]

jobs:
  forge-verify:
    runs-on: ubuntu-latest
    steps:
 - uses: actions/checkout@v4
        with:
          submodules: recursive

 - name: Install Foundry
        uses: foundry-rs/foundry-toolchain@v1
        with:
          version: nightly

 - name: Run Forge Format Check
        run: forge fmt --check contracts/

 - name: Run Forge Unit & Invariant Tests
        run: forge test --gas-report --fuzz-runs 10000

 - name: Run Slither Security Analysis
        uses: crytic/slither-action@v0.3.0
        with:
          target: 'contracts/'
          sarif: 'slither.sarif'
          fail-on: 'medium'

 - name: Upload SARIF to GitHub Security Tab
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: slither.sarif
```

```yaml
# .github/workflows/ci-container-publish.yml (Excerpt)
name: Container Build, Scan & Sign

on:
  push:
    branches: [main]

jobs:
  build-and-sign:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write
    steps:
 - uses: actions/checkout@v4
 - uses: docker/setup-buildx-action@v3
 - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
 - name: Build and Push Image
        id: build-image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: services/matching-engine/Dockerfile
          push: true
          tags: ghcr.io/growww/matching-engine:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

 - name: Scan Image for CVEs with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ghcr.io/growww/matching-engine:${{ github.sha }}
          severity: 'CRITICAL,HIGH'
          exit-code: '1'

 - name: Sign Container with Cosign
        uses: sigstore/cosign-installer@v3.4.0
 - run: |
          cosign sign --yes ghcr.io/growww/matching-engine:${{ github.sha }}
```

## Security & Compliance Notes
- Zero Static Cloud Secrets: All CI/CD jobs interacting with AWS/GCP or container registries authenticate exclusively via OIDC tokens with scoped IAM role assumption.
- Immutable Image Tags: Every container image is tagged with the exact Git commit SHA and signed with Cosign; mutable tags like `latest` are prohibited in production pipelines.
- SEBI Audit Readiness: All build logs, test results, code coverage reports, and security scans are archived for 90 days with tamper-evident GitHub audit log streaming.

## Acceptance Criteria
- [ ] PR checks automatically run and complete all lint, test, and security jobs in <10 minutes.
- [ ] Foundry test suite runs with 10,000 fuzz runs and zero invariant failures on `SettlementDvP.sol`.
- [ ] Slither static analysis blocks PR merges on any Medium or High severity finding.
- [ ] Code coverage threshold of 85% is strictly enforced, blocking regression.
- [ ] Container images are built, scanned with Trivy, signed with Cosign, and published to the container registry with valid SBOMs.
- [ ] Secret scanning catches simulated committed API keys and immediately fails the pipeline.

## Suggested Order / Dependencies
- Prerequisites: Prompt 106 (Repository Structure), Prompt 107 (Coding Standards), Prompt 303/306 (Solidity Contracts).
- Parallel Tasks: Prompt 801 (Local Dev Env), Prompt 804 (Continuous Delivery).
