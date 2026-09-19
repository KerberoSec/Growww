# 901 - Unit & Integration Testing Strategy Across Polyglot Stack

## Purpose
Establishes a unified, automated testing architecture across all polyglot microservices in the Growww platform (Rust, Go, Python/FastAPI, TypeScript, and Solidity). Because Growww handles real financial assets, fiat INR transactions, and permissioned ledger states under SEBI/RBI regulatory oversight, every individual microservice must guarantee high determinism, strict contract verification, and zero regression through isolated unit tests and containerized integration tests before merging.

## What You Are Building
A comprehensive test framework and execution harness across the monorepo/polyrepo workspaces:
- Polyglot test runner configurations and test harnesses (`tests/harness/`, `tests/fixtures/`).
- Standardized Testcontainers setup for ephemeral PostgreSQL, Redis, Kafka, and mock Hyperledger Besu blockchain nodes.
- Mock server fixtures for third-party upstream providers (NSDL/CDSL Depository APIs, UPI/Payment Aggregators, Aadhaar/DigiLocker KYC, CVL KRA).
- Automated CI test matrix executing unit and integration test suites with enforced >=85% code coverage and zero-leakage memory/sanitizer checks.

## Scope Boundaries
- **In Scope:**
 - Unit test standards and mocking paradigms for Rust (matching engine, risk), Go (settlement, ingestion), Python (FastAPI user/KYC/reporting), and TypeScript (BFF/admin).
 - Integration testing with Testcontainers for local and CI container orchestration.
 - Contract testing for gRPC and Kafka event schemas (Pact / Protobuf compatibility checks).
 - Smart contract unit and integration tests using Foundry / Hardhat against a local Besu EVM instance.
- **Out of Scope / Handled Elsewhere:**
 - End-to-end multi-tier UI/API integration flows (Prompt 902).
 - High-throughput load and stress testing at 50,000 req/s (Prompt 903).
 - Fault injection and network partition chaos testing (Prompt 904).

## Technology to Use
- **Test Frameworks:** `cargo test` + `mockall` (Rust), `go test` + `testify` + `gomock` (Go), `pytest` + `pytest-asyncio` + `pytest-mock` (Python), `Vitest` / `Jest` (TypeScript/Next.js), `Foundry` (`forge test`) for Solidity smart contracts.
- **Testcontainers:** `testcontainers-go`, `testcontainers-python`, `testcontainers-rs`, and `testcontainers-node` spinning up PostgreSQL 16+, Redis 7+, Kafka (KRaft mode), and Hyperledger Besu ephemeral nodes.
- **Mock Servers:** `WireMock` / `prism` for OpenAPI/gRPC service virtualization of banking and depository APIs.
- **Coverage & Reporting:** `tarpaulin` (Rust), `go-acc` (Go), `pytest-cov` (Python), `c8`/`istanbul` (TypeScript), aggregating into SonarQube and GitHub Actions CI reports.

*Justification:* Utilizing native language runners combined with Testcontainers provides high-fidelity integration testing with zero external network dependencies, ensuring tests are deterministic, fast, and hermetic.

## Backend / Infra Touchpoints
- Docker / Podman socket on local dev environments and CI runner instances.
- Ephemeral PostgreSQL database instances with automatic flyway/alembic migration execution.
- Ephemeral Redis clusters for cache and distributed lock verification.
- Ephemeral Kafka brokers for producer/consumer partition and serialization validation.
- Mock NSDL/CDSL REST/SOAP servers and Mock UPI Payment Switch endpoints.

## Blockchain Interaction
- Smart contract unit tests execute in Foundry local EVM environments validating `DigitalSecurityToken.sol`, `SettlementDvP.sol`, and `ComplianceRegistry.sol`.
- Microservice ledger adapters (Go settlement service, Python reporting) interact with containerized Hyperledger Besu test nodes running QBFT consensus.
- Integration tests assert correct transaction signing via mock HSM/Vault transit engine interfaces and verify event emission parsing (`TransferWithCompliance`, `DvPExecuted`, `ReserveAttestationPublished`).

## Step-by-Step Build Instructions
1. Establish the centralized test directory structure (`tests/harness/`, `tests/fixtures/`, `tests/mocks/`) and shared test utility libraries.
2. Configure base Testcontainers harnesses for PostgreSQL, Redis, and Kafka in Go, Python, and Rust.
3. Implement database migration bootstrap hooks in test harnesses to ensure ephemeral DBs match production schemas before test execution.
4. Build WireMock / HTTP stub definitions for external depository interfaces (NSDL DP settlement feed, CDSL holding verification).
5. Build WireMock / HTTP stub definitions for banking rails (NPCI UPI collect request, refund callback, webhook signatures).
6. Implement Rust unit test suites for the matching engine (`services/matching-engine`), verifying order book price-time priority, FIFO queues, and fractional share arithmetic.
7. Implement Go integration tests for the trade settlement service (`services/settlement-service`), validating two-phase commit and DvP state transitions against ephemeral Postgres and Kafka.
8. Implement Python/FastAPI integration tests for KYC and User services using `pytest-asyncio` and Testcontainers Postgres.
9. Implement Foundry test suites (`contracts/test/`) for all Solidity contracts covering reentrancy guards, KYC whitelist modifiers, and multi-sig thresholds.
10. Configure Pact / Schema-Registry contract verification to prevent breaking schema changes across gRPC and Kafka payloads.
11. Implement SonarQube and CI test report aggregation enforcing >=85% branch and line coverage thresholds.
12. Establish pre-commit git hooks and parallelized CI pipeline workflows (`.github/workflows/test.yml`) executing tests under 5 minutes total runtime.

## Interfaces / Contracts
```yaml
# Test Harness Configuration (tests/harness/config.yaml)
version: "1.0"
test_environment:
  type: "ephemeral-containers"
  containers:
    postgres:
      image: "postgres:16-alpine"
      db_name: "growww_test"
      migrations_path: "migrations/sql"
    redis:
      image: "redis:7.2-alpine"
      port: 6379
    kafka:
      image: "confluentinc/cp-kafka:7.5.0"
      kraft_mode: true
      topics:
 - name: "orders.placed.v1"
          partitions: 4
 - name: "settlement.dvp.v1"
          partitions: 4
    besu_node:
      image: "hyperledger/besu:24.1.0"
      consensus: "qbft"
      mining: true
      block_time_sec: 1
mock_gateways:
  wiremock:
    port: 8089
    stubs_dir: "tests/mocks/stubs"
coverage_gates:
  unit_line_coverage_min: 85.0
  unit_branch_coverage_min: 80.0
  integration_coverage_min: 75.0
```

## Security & Compliance Notes
- Ensure zero production PII, real Aadhaar/PAN data, or live cryptographic keys are present in test fixtures or mock databases.
- Test data generators must generate synthetic, mathematically valid PANs (e.g., matching the `[A-Z]{5}[0-9]{4}[A-Z]{1}` regex) and verifiably mock bank accounts.
- Test execution must run inside non-root container contexts with network isolation preventing outbound external internet calls.

## Acceptance Criteria
- [ ] Testcontainers setup runs reliably in local dev environments and GitHub Actions CI without port collisions.
- [ ] Rust matching engine unit tests achieve 100% logic coverage for fractional matching and price-time priority.
- [ ] Go, Python, and TypeScript services achieve >=85% code coverage across all core domain paths.
- [ ] Foundry smart contract test suite passes with 100% branch coverage on access control and DvP escrow flows.
- [ ] Mock fixtures accurately simulate 100% of external banking (UPI) and depository (NSDL/CDSL) API error codes.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Standards), 104 (Event Schemas), 107 (Coding Standards), 201-208 (Microservices), 303-306 (Smart Contracts).
- **Parallel Tasks:** 803 (CI Pipeline Design), 902 (End-to-End Testing Strategy).
