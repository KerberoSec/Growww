# 902 - End-to-End Testing Strategy: Flutter Client & Backend Integration

## Purpose
Validates complete, multi-tier user journeys across the Growww ecosystem - from the multi-platform Flutter client (Android, iOS, Windows, macOS, Linux) and Next.js Web App through API Gateways, microservices, the Rust matching engine, and the permissioned Hyperledger Besu blockchain. This ensures that user actions (such as identity onboarding, fund deposits, fractional equity trading, and atomic DvP settlements) operate flawlessly end-to-end without regressions or cross-service state drift.

## What You Are Building
An automated, cross-platform End-to-End (E2E) testing framework and automated test suite (`tests/e2e/`):
- Multi-platform Flutter E2E automation suite powered by Flutter Patrol (`patrol`), automating both in-app widgets and native OS interactions (camera access, biometric authentication, push notifications, file pickers).
- Web E2E automation suite using Playwright for the Next.js investor portal and administrative/compliance console.
- Backend orchestration harnesses with mock external banking (UPI/IMPS) switches, DigiLocker KYC stubs, and NSDL/CDSL custody settlement gateways.
- Continuous E2E testing pipelines executing against containerized staging environments and device clouds (Firebase Test Lab / AWS Device Farm).

## Scope Boundaries
- **In Scope:**
 - Full critical investor user journeys: User Registration -> KYC Verification -> Bank Account Linking -> INR Instant Deposit -> Order Placement -> Matching Engine Execution -> Besu On-Chain DvP Settlement -> Portfolio Update & Proof-of-Reserve Verification.
 - Native mobile interactions: Biometric login prompts, camera frame feeding for video liveness verification, push notification reception.
 - Multi-platform cross-checks: State synchronization between Flutter mobile, Flutter desktop, and Next.js web clients.
 - Automated teardown and state resetting after test scenario execution.
- **Out of Scope / Handled Elsewhere:**
 - Microservice-level unit and integration testing with Testcontainers (Prompt 901).
 - High-concurrency load testing at 50,000 req/s (Prompt 903).
 - Infrastructure chaos and network partition simulation (Prompt 904).

## Technology to Use
- **Mobile & Desktop Client Automation:** Flutter Patrol (`patrol` package v3+) - selected because native Flutter `integration_test` cannot interact with native platform dialogs (camera permissions, FaceID/Fingerprint OS prompts, system notifications).
- **Web App Automation:** Playwright (TypeScript) for high-speed, headless multi-browser (Chromium, WebKit, Firefox) verification of Next.js apps.
- **Test Orchestration & Scenario Definition:** BDD Gherkin syntax with `behave` (Python) / Cucumber for readable regulatory and business compliance scenarios.
- **Device Cloud & CI Runners:** GitHub Actions runners with Linux/macOS virtual runners + Firebase Test Lab / local Android Emulators (x86_64 hardware accelerated) and iOS Simulators.

*Justification:* Flutter Patrol provides unified multi-platform UI control crossing the Flutter-to-native OS barrier, combined with Playwright for web and Gherkin scenarios for verifiable regulatory compliance paths.

## Backend / Infra Touchpoints
- Staging API Gateway (Traefik / Kong / Envoy) over mTLS.
- Staging Microservices (User, KYC, Wallet, Order, Settlement, Portfolio, Market Data).
- Mock Banking Rails (Mock NPCI UPI Server generating instant `SUCCESS` / `FAILURE` / `DEEMED_SUCCESS` callbacks).
- Mock Custodian Gateways (Mock NSDL/CDSL settlement file generators).
- Staging Hyperledger Besu consortium network (4-validator QBFT nodes).

## Blockchain Interaction
- E2E tests initiate orders that trigger off-chain risk checks, pass through the Rust matching engine, and dispatch atomic settlement calls to `SettlementDvP.sol` on Hyperledger Besu.
- The test harness monitors on-chain transaction hash generation, waits for block inclusion (QBFT 2s block finality), and verifies that the `TransferWithCompliance` and `DvPExecuted` events are correctly indexed by off-chain event indexers and reflected in the user's Flutter portfolio view.
- Validates on-chain KYC whitelist checks (`ComplianceRegistry.sol`) by asserting that non-KYC-verified test accounts are blocked at smart contract execution.

## Step-by-Step Build Instructions
1. Initialize the `tests/e2e/` directory structure with modular suites for `flutter_patrol/`, `web_playwright/`, and `scenarios/`.
2. Configure `patrol` in the Flutter root project (`pubspec.yaml`), setting up custom runners for Android, iOS, and Desktop platforms.
3. Implement mock camera and document injector services in Flutter staging builds to feed synthetic PAN/Aadhaar card images and simulated liveness video streams.
4. Implement mock biometric authentication bridges in Patrol to simulate successful FaceID/Fingerprint authentication responses.
5. Create Gherkin feature definitions (`tests/e2e/features/`) covering core flows: `01_onboarding_kyc.feature`, `02_funds_deposit.feature`, `03_fractional_stock_trading.feature`, `04_dvp_settlement.feature`, and `05_proof_of_reserve_audit.feature`.
6. Build automated Playwright test scripts (`tests/e2e/web_playwright/`) for Next.js investor web dashboard and the admin compliance review console.
7. Implement backend test fixture seeding utilities that pre-populate staging databases with test stock tickers (e.g., `GROWWW-TCS-001`, `GROWWW-RELIANCE-001`) and custodial reserves.
8. Implement an automated webhook callback simulator that listens for UPI collect requests and sends signed mock bank approval webhooks.
9. Implement on-chain ledger verification assertions in the E2E harness that query the Hyperledger Besu RPC endpoint to verify contract state transitions.
10. Integrate real-time WebSocket assertions to verify that order book level-2 depth and executed trade tickets stream live to connected clients during test runs.
11. Implement an automated cleanup and teardown utility that rolls back or archives test user balances and resets order book memory state.
12. Configure CI matrix workflows (`.github/workflows/e2e-matrix.yml`) to execute E2E tests across Android emulator, iOS simulator, desktop Linux/macOS, and web browsers on scheduled nightly runs and pre-release tags.

## Interfaces / Contracts
```gherkin
# Feature: End-to-End Fractional Stock Purchase and Atomic DvP Settlement
# File: tests/e2e/features/03_fractional_stock_trading.feature

Feature: Fractional Equity Trading & On-Chain DvP Settlement
  As a verified Indian Retail Investor
  I want to buy 0.25 shares of INFOSYS using deposited INR
  So that my trade is matched off-chain and settled atomically on the permissioned blockchain

  Background:
    Given user "investor_rajesh" is authenticated and KYC-approved
    And user has an available wallet balance of INR 50,000.00
    And depository custody holds 100 whole shares of "INFY" in escrow

  Scenario: Place Fractional Buy Order and Verify On-Chain Receipt
    When the user opens the Flutter stock detail screen for "INFY"
    And submits a Limit Buy Order for "0.25" units at INR "1,600.00" per unit
    Then the order status updates to "PLACED" within 200 milliseconds
    And the Rust matching engine matches against resting sell liquidity
    And the settlement service triggers atomic DvP execution on Hyperledger Besu
    And the user receives a push notification "Trade Executed: 0.25 INFY"
    And the Flutter portfolio screen displays "0.25 INFY" with valid blockchain tx_hash
    And the on-chain ProofOfReserveRegistry attests 1:1 physical backing
```

```yaml
# Patrol E2E Configuration (tests/e2e/patrol_config.yaml)
patrol:
  app_name: "Growww Client"
  package_name: "com.growww.investor.app"
  android:
    flavor: "staging"
    package_name: "com.growww.investor.staging"
  ios:
    bundle_id: "com.growww.investor.staging"
  timeouts:
    find_timeout_ms: 10000
    action_timeout_ms: 5000
  mock_sensors:
    camera_feed_dir: "tests/fixtures/synthetic_faces/"
    biometrics: "always_pass"
```

## Security & Compliance Notes
- All E2E test runs must strictly utilize synthetic accounts and zero production customer records.
- Staging API keys, test JWT secrets, and mock banking webhooks must be segregated from production secrets and rotated periodically via Vault.
- Mock KYC flows must explicitly verify that rejected KYC test profiles are halted with strict error codes and cannot execute any trading actions.

## Acceptance Criteria
- [ ] Flutter Patrol E2E tests successfully execute full onboarding and trading journeys on Android, iOS, Windows, macOS, and Linux runners.
- [ ] Playwright web test suite validates 100% of investor web portal and admin console critical workflows.
- [ ] On-chain settlement transactions generated during E2E runs are verified on Besu nodes with confirmed receipt logs.
- [ ] CI pipeline executes the complete smoke E2E suite within 15 minutes, generating full video recordings, screenshots on failure, and JUnit XML test reports.

## Suggested Order / Dependencies
- **Prerequisites:** 201-208 (Backend Services), 303-306 (Smart Contracts), 501-512 (Flutter Client UI), 601-604 (Web/Admin App), 901 (Unit Testing).
- **Parallel Tasks:** 903 (Load Testing), 904 (Chaos Resilience Testing).
