# 106 - Monorepo Architecture, Repository Layout & Tooling Strategy

## Purpose
Establishes the enterprise repository layout, workspace management tooling, multi-language build orchestration, and code sharing strategy for the Growww investment platform. Growww is a complex polyglot system composed of high-performance backend microservices (Go, Rust, Python), smart contracts (Solidity for Hyperledger Besu), a multi-platform client (Flutter for Android/iOS/Windows/Linux/macOS), web applications (Next.js), and infrastructure definitions.

This prompt resolves the Monorepo vs Polyrepo architectural trade-off, formalizes the canonical repository directory taxonomy, establishes shared Protocol Buffer code-generation workflows, and sets up build orchestration tooling (Turborepo / Taskfile / Cargo Workspaces / Go Workspaces) to enable atomic commits, synchronized API updates, and sub-minute incremental CI build cycles.

## What You Are Building
A comprehensive repository architecture specification (`docs/standards/repo_layout_and_tooling.md`), unified directory layout skeleton, root Taskfile orchestration suite, and Git governance guidelines:
- Monorepo Decision Record (ADR-002) justifying a unified monorepo approach for cross-tier contract consistency, atomic refactoring, and simplified dependency management.
- Canonical Directory Hierarchy: `apps/` (Flutter client, Next.js web, Admin console), `services/` (Go, Rust, and Python microservices), `contracts/` (Solidity smart contracts), `packages/` (Shared Protobuf schemas, design system, common utilities), `infra/` (Terraform, Helm charts, Docker Compose), and `docs/`.
- Build & Task Orchestration: Root `Taskfile.yml` and language-specific workspace configurations (`Cargo.toml` workspace, `go.work`, `pnpm-workspace.yaml`).
- Git Branching & CI Change Detection: Trunk-based development workflow with path-filtered CI execution matrices.

## Scope Boundaries
- **In Scope:**
 - Monorepo directory structure and module isolation rules.
 - Multi-language workspace configuration (`go.work`, `Cargo.toml`, `pnpm-workspace.yaml`).
 - Taskfile automation for local development, linting, testing, and protobuf code generation.
 - Git branching strategy, PR templates, and CODEOWNERS rules.
 - Path-filtering configuration for targeted CI execution.
- **Out of Scope / Handled Elsewhere:**
 - Coding style rules and language linters (handled in Prompt 107).
 - CI/CD pipeline automation scripts and Kubernetes deployment (handled in Category 8).
 - Individual service implementations (handled in Category 2, 3, 5, 6).

## Technology to Use
- **Workspace & Task Orchestration:**
 - *Taskfile (go-task):* Unified, cross-platform CLI task runner replacing complex Makefiles for developer workflows.
 - *Go Workspaces (`go.work`):* Multi-module Go management enabling direct local cross-service module references.
 - *Cargo Workspaces (Rust):* Shared dependency caching and unified compilation for the Order Matching Engine and Settlement modules.
 - *pnpm Workspaces & Turborepo:* Fast, cached monorepo management for TypeScript/Next.js applications and shared UI packages.
 - *Melos:* Monorepo management tool for the multi-platform Flutter client and shared Dart packages.
- **Version Control & Governance:** Git with Trunk-Based Development, Signed GPG/SSH commits, and GitHub CODEOWNERS.

## Backend / Infra Touchpoints
- **CI/CD Pipeline (GitHub Actions / GitLab CI):** Path-based trigger filters ensuring only modified services and their direct dependents are built and tested.
- **Docker Multi-Stage Builds:** Standardized containerization recipes leveraging Layer Caching and monorepo build contexts.
- **Local Dev Orchestration:** Root `docker-compose.yml` linking local services to containerized PostgreSQL, Kafka, Redis, and Besu ledger nodes.

## Blockchain Interaction
Standardizes smart contract repository organization and artifact sharing:
- **Contract Workspace (`contracts/`):** Houses Foundry / Hardhat project suites for Hyperledger Besu smart contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`).
- **Automated ABI & Stub Generation:** A central Taskfile command (`task proto:gen` / `task contracts:compile`) automatically compiles Solidity contracts and exports TypeScript/Go/Rust/Dart typed contract bindings into `packages/contracts-bindings/` for direct consumption by backend services and Flutter clients.

## Step-by-Step Build Instructions
1. Author `docs/architecture/adr/ADR-002-monorepo-strategy.md` detailing the rationale for a unified monorepo over polyrepos (contract synchronization, atomic cross-tier commits, simplified end-to-end testing).
2. Create the root directory hierarchy:
   ```
   ├── apps/
   │   ├── client_flutter/
   │   ├── web_investor/
   │   └── web_admin/
   ├── services/
   │   ├── user_service/
   │   ├── kyc_service/
   │   ├── wallet_service/
   │   ├── order_service/
   │   ├── matching_engine/
   │   ├── settlement_service/
   │   └── custody_adapter/
   ├── contracts/
   │   └── besu_contracts/
   ├── packages/
   │   ├── proto/
   │   ├── contracts_bindings/
   │   └── ui_design_system/
   ├── infra/
   │   ├── docker/
   │   ├── k8s/
   │   └── terraform/
   └── docs/
   ```
3. Configure `go.work` in the repository root to link all Go microservices and shared Go libraries.
4. Configure root `Cargo.toml` with workspace members for all Rust services (`services/matching_engine`, `services/settlement_engine`).
5. Configure `pnpm-workspace.yaml` and `turbo.json` in root for web apps (`apps/web_investor`, `apps/web_admin`) and shared TypeScript libraries.
6. Configure `melos.yaml` for Flutter multi-platform apps and shared Dart packages.
7. Create `Taskfile.yml` in the repository root with standardized commands:
 - `task setup`: Installs all language toolchains, linters, and Buf CLI.
 - `task proto:gen`: Compiles Protocol Buffers into Go, Rust, Python, Dart, and TypeScript stubs.
 - `task contracts:compile`: Compiles Solidity contracts and generates typed bindings.
 - `task lint`: Runs linters across all languages.
 - `task test`: Runs unit tests across all workspaces.
 - `task dev:up`: Boots local Docker Compose environment (Kafka, Postgres, Redis, Besu).
8. Create `.gitignore` comprehensive across all 5 languages (ignoring `target/`, `node_modules/`, `.dart_tool/`, `venv/`, `bin/`, `.turbo/`).
9. Create `.github/CODEOWNERS` establishing mandatory multi-party approvals for critical paths (`contracts/`, `infra/`, `packages/proto/`).
10. Establish trunk-based development rules: short-lived feature branches (`feat/*`, `fix/*`), squash-merge to `main`, mandatory passing CI checks and signed commits.
11. Document the complete workflow in `docs/standards/repo_layout_and_tooling.md`.
12. Verify local task execution and monorepo build passes cleanly from a fresh checkout.

## Interfaces / Contracts

### Root Taskfile Configuration (`Taskfile.yml`)
```yaml
version: '3'

vars:
  PROTO_DIR: packages/proto
  CONTRACTS_DIR: contracts/besu_contracts

tasks:
  default:
    cmds:
 - task --list

  setup:
    desc: Install developer toolchains, Buf, linters, and dependencies
    cmds:
 - buf --version || echo "Please install Buf CLI (https://buf.build)"
 - pnpm install
 - cargo --version
 - go version
 - flutter --version

  proto:gen:
    desc: Generate typed Protobuf stubs for Go, Rust, Python, Dart, and TypeScript
    dir: '{{.PROTO_DIR}}'
    cmds:
 - buf lint
 - buf generate

  contracts:compile:
    desc: Compile Solidity smart contracts and generate bindings
    dir: '{{.CONTRACTS_DIR}}'
    cmds:
 - forge build
 - forge bind --bindings-path ../../packages/contracts_bindings/go --module go

  lint:
    desc: Run linters across all monorepo workspaces
    cmds:
 - task: lint:go
 - task: lint:rust
 - task: lint:python
 - task: lint:flutter
 - task: lint:web

  test:
    desc: Run unit tests across all monorepo workspaces
    cmds:
 - go test ./services/...
 - cargo test --workspace
 - pytest services/user_service services/kyc_service
 - flutter test apps/client_flutter
 - pnpm test

  dev:up:
    desc: Start local infrastructure stack (Postgres, Kafka, Redis, Besu Node)
    cmds:
 - docker compose -f infra/docker/docker-compose.yml up -d
```

### GitHub CODEOWNERS Specification (`.github/CODEOWNERS`)
```text
# Global Monorepo Approvers
* @growww-core-architects

# Smart Contracts & Blockchain Layer (Mandatory Multi-Party Approval)
/contracts/ @growww-blockchain-lead @growww-security-team
/packages/contracts_bindings/ @growww-blockchain-lead

# Protocol Buffers & Canonical API Contracts
/packages/proto/ @growww-core-architects @growww-api-governance

# Infrastructure & Production Deployments
/infra/ @growww-devops-lead @growww-security-team

# Core Matching & Financial Settlement
/services/matching_engine/ @growww-trading-engineers
/services/settlement_service/ @growww-settlement-engineers
```

## Security & Compliance Notes
- **Mandatory Cryptographic Commit Signing:** All commits to the repository must be signed with GPG or SSH keys verified against developer identities; unsigned commits are rejected at the Git push hook and GitHub branch protection layer.
- **Automated Secret Scanning:** Pre-commit hooks and CI pipelines run automated scans (TruffleHog / GitGuardian) to block hardcoded private keys, API secrets, or passwords from entering the Git history.
- **Strict CODEOWNERS Enforcement:** Modifications to smart contracts (`/contracts/`), API schemas (`/packages/proto/`), and infrastructure (`/infra/`) require mandatory approval from at least two senior domain architects before merging.

## Acceptance Criteria
- [ ] Comprehensive Monorepo specification (`docs/standards/repo_layout_and_tooling.md`) and ADR-002 published.
- [ ] Root directory structure created with all workspace files (`go.work`, `Cargo.toml`, `pnpm-workspace.yaml`, `melos.yaml`).
- [ ] Root `Taskfile.yml` configured and verified executing `task setup`, `task proto:gen`, `task contracts:compile`, `task lint`, and `task test`.
- [ ] `.github/CODEOWNERS` and branch protection rules formalized.
- [ ] Comprehensive `.gitignore` committed covering all 5 language ecosystems.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture Overview).
- **Parallel Work:** Prompt 107 (Coding Standards), Prompt 108 (Environment Strategy).
- **Blocks:** Prompt 801 (Docker Compose Dev Env), Prompt 803 (CI Pipeline), All microservice and client scaffolding prompts.
