# 812 - Dual-Environment Testnet Sandbox, Mainnet Isolation & Orchestration Suite (Go / Kubernetes / Besu)

## Purpose
Operating a regulated national digital securities and fractional equity platform requires continuous end-to-end testing, partner onboarding, and regulatory simulation without endangering live production ledgers. The Production Mainnet operates under strict SEBI, RBI, and IFSCA compliance with FIPS 140-2 Level 3 Hardware Security Modules (AWS CloudHSM), dedicated NSDL/CDSL leased lines, and live central bank Digital Rupee (e₹) CBDC / RTGS settlement rails. Conversely, the Regulatory Sandbox Testnet must support rapid experimentation, developer integration, automated nightly state resets, chaos injection, simulated KYC validations, mock depository batches, and public token faucets.

The **Dual-Environment Testnet Sandbox, Mainnet Isolation & Orchestration Suite** (`infra/sandbox-orchestrator/`) enforces an absolute, cryptographic and network-level boundary separating the Regulatory Sandbox Testnet from Production Mainnet. It prevents testnet credential or state contamination into production, manages automated testnet environment lifecycles, orchestrates mock depository/banking infrastructure, runs chaos resilience harnesses, and dispenses testnet assets to authorized fintech developers while guaranteeing identical runtime semantics (including the universal 0.00% (No fee at all) platform fee with its 0.00% fee at launch (governed by FeeController.sol) distribution).

## What You Are Building
A robust, automated infrastructure orchestration and environment isolation suite (`infra/sandbox-orchestrator/`) written in Go and declarative infrastructure manifests. Key deliverables include:
- **Cryptographic Boundary & Network Isolation Firewalls:** Strict Cilium eBPF network security policies, VPC peering bans, separate AWS IAM trust boundaries, and distinct mTLS service mesh roots of trust preventing any packet or RPC leakage between Testnet and Mainnet.
- **Sandbox Orchestrator Daemon (`infra/sandbox-orchestrator`):** A Go-based control plane daemon that manages testnet lifecycle events, triggers scheduled or on-demand ledger re-genesis, seeds baseline market data, and manages ephemeral sandbox tenants for third-party participants.
- **Mock Depository & Banking Gateway Adapter:** Simulates NSDL/CDSL demat credit/debit batch files, RTGS/IMPS settlement webhooks, and RBI e-Rupee CBDC token minting in the testnet sandbox, providing 100% API parity with live financial market infrastructure.
- **Synthetic Identity & Mock KYC Generator:** Generates deterministic test identities, emulates DigiLocker and PAN/Aadhaar verification responses, and allows developers to toggle test account verification statuses (`KYC_APPROVED`, `KYC_REJECTED`, `PEP_SUSPECT`, `SANCTIONED`) on demand.
- **Public & Partner Developer Faucet:** Automated token dispersion service delivering testnet native gas tokens, simulated e₹ CBDC, and mock tokenized securities (e.g., mock RELIANCE, mock TATA, mock INFY) with strict rate limits and IP/wallet abuse protection.
- **Chaos & Network Partition Injection Harness:** Injects configurable fault scenarios into the sandbox (such as validator node crashes, 500ms network latency, dropped Kafka partitions, and simulated banking gateway downtime) to validate platform resilience.
- **Universal Fee Invariant Validator:** Continuously asserts that testnet matching engine and smart contract instances enforce the 0.00% (No fee at all) platform fee (0.00% fee at launch; future fee parameters governed by FeeController.sol) identical to mainnet rules.

## Scope Boundaries
- **In Scope:**
  - Complete network, storage, and cryptographic isolation between Testnet (Chain ID `13371`) and Mainnet (Chain ID `13370`).
  - Automated sandbox lifecycle management, state reset scripting, and test data seeding.
  - High-fidelity mock services for NSDL, CDSL, UPI, RTGS, and RBI e-Rupee rails.
  - Testnet faucet rate-limiting, wallet disbursement, and balance replenishment.
  - Chaos fault injection controllers and automated resilience verification suites.
  - Zero-PII synthetic data generators for sandbox testing.
- **Out of Scope / Handled Elsewhere:**
  - GitOps CI/CD promotion pipelines and bytecode verification (handled in Prompt 811).
  - Production AWS CloudHSM key lifecycle daemon (handled in Prompt 717).
  - Public Developer Portal and Web Faucet UI (handled in Prompt 609).
  - Flutter client sandbox toggle interface (handled in Prompt 527).
  - Mainnet dress rehearsal cutover protocol (handled in Prompt 911).

## Technology to Use
- **Primary Language & Runtime:** **Go 1.22+** utilizing `client-go` for Kubernetes orchestration, `gin-gonic` for REST APIs, and `pgx/v5` for PostgreSQL state persistence.
- **Network Isolation & Policy Engine:** **Cilium (eBPF)** and **Kubernetes NetworkPolicies** enforcing strict namespace segmentation and multi-VPC egress firewalls.
- **Consortium Blockchain:** **Hyperledger Besu** configured in dual clusters:
  - *Regulatory Sandbox Testnet:* Chain ID `13371`, 4-node QBFT dev cluster with fast block time (1 second) and reset automation.
  - *Production Mainnet:* Chain ID `13370`, Multi-institutional FIPS 140-2 Level 3 HSM-backed QBFT cluster.
- **Relational Persistence:** **PostgreSQL 16+** with isolated database clusters for sandbox orchestrator metadata and mock depository states.
- **In-Memory Cache & Faucet Limiter:** **Redis Cluster** for token bucket rate-limiting and active sandbox tenant tracking.
- **Secret & Key Isolation:** **HashiCorp Vault / AWS KMS** for testnet keys; dedicated **AWS CloudHSM** enclaves for mainnet keys with physical role separation.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `sandbox_environments`, `sandbox_mock_entities`, `sandbox_reset_schedules`, `chaos_experiment_runs`, `faucet_disbursement_logs`.
- **Kafka Topics (Testnet Cluster):** Consumes `testnet.sandbox.reset.v1`, `testnet.chaos.inject.v1`; publishes `testnet.sandbox.state_ready.v1`, `testnet.faucet.claim.v1`, `testnet.chaos.metric_collected.v1`. Note: Testnet Kafka clusters are physically isolated from Production Kafka clusters.
- **Developer Portal (Prompt 609):** Routes interactive API requests and faucet claims to the sandbox orchestrator.
- **Mock Depository Service (Prompt 213 / Prompt 812):** Executes mock depository settlements and trade confirmations.
- **Compliance Registry (Prompt 305):** Ingests synthetic KYC claims in the testnet environment.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Separate Cryptographic Roots:** Testnet (Chain ID `13371`) and Mainnet (Chain ID `13370`) possess distinct genesis blocks, validator sets, contract addresses, and signing keys. No mainnet private key is ever accessible to testnet infrastructure or developers.
- **Automated Testnet Re-Genesis:** The sandbox orchestrator can tear down the testnet Besu network, generate a fresh QBFT genesis state, deploy core contracts (`SettlementDvP.sol`, `DigitalSecurityToken.sol`, `ComplianceRegistry.sol`), and seed initial balances within 60 seconds.
- **Zero PII Compliance:** All synthetic KYC records and investor accounts in the sandbox use mathematically valid yet provably synthetic PAN/Aadhaar patterns (e.g. `AAAAA0000A`) with zero linkage to real individuals, complying with DPDP Act 2023.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Create Go project `infra/sandbox-orchestrator` with packages: `cmd`, `config`, `controller`, `faucet`, `mockdepository`, `chaos`, and `storage`.
2. **Define Protobuf Schema:** Author `proto/growww/sandbox/v1/sandbox_orchestrator.proto` declaring RPCs: `TriggerSandboxReset`, `SeedMockData`, `DisburseFaucetTokens`, `InjectChaosScenario`, and `GetSandboxStatus`.
3. **Generate Go Stubs:** Compile Protobuf schemas into Go server stubs using `buf`.
4. **Deploy Network Isolation Policies:** Author Cilium eBPF and Kubernetes NetworkPolicy manifests denying all cross-namespace and cross-VPC traffic between `testnet-*` and `prod-*` environments.
5. **Implement Testnet Besu Cluster Automation:** Create Kubernetes Helm charts and operator scripts to automate the deployment, peering, snapshotting, and re-genesis of the 4-node QBFT testnet cluster.
6. **Implement Sandbox Reset Coordinator:** Build the Go orchestration engine that drops test databases, resets Besu ledger data volumes, runs Flyway/SQL migrations, and redeploys smart contract suites from versioned Git tags.
7. **Implement Mock Depository Engine:** Create high-fidelity mock generators producing NSDL/CDSL daily position files, demat allocation responses, and corporate action schedules.
8. **Implement Synthetic KYC Generator:** Implement synthetic PAN, Aadhaar, and DigiLocker XML mock endpoints with configurable verification outcomes for developer testing.
9. **Implement Multi-Chain Developer Faucet:** Build faucet distribution logic supporting Besu testnet gas, simulated Digital Rupee CBDC tokens, and mock equity tokens with IP and wallet rate-limiting.
10. **Implement Chaos Injection Controller:** Build chaos harnesses that induce pod kills, network packet latency, Byzantine node partitioning, and mock payment gateway timeout errors.
11. **Implement Fee Invariant Monitor:** Build automated integration test worker asserting that all testnet trades execute with the standard 0.00% (Zero Fee) platform fee (0.00% fee at launch (governed by FeeController.sol)).
12. **Configure Observability & Alerting:** Instrument the sandbox orchestrator with Prometheus metrics and Grafana dashboards tracking faucet reserve balances, reset latencies, and active mock tenants.
13. **Write Full Verification Suite:** Execute end-to-end integration tests verifying zero network cross-talk from testnet pods to mainnet RPCs and confirming clean testnet resets under 60 seconds.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/sandbox/v1/sandbox_orchestrator.proto`)
```protobuf
syntax = "proto3";

package growww.sandbox.v1;

option go_package = "github.com/growww/proto/gen/go/sandbox/v1;sandboxv1";

enum SandboxResetScope {
  SANDBOX_RESET_SCOPE_UNSPECIFIED = 0;
  SANDBOX_RESET_SCOPE_FULL_ENVIRONMENT = 1; // Blockchain + DBs + Kafka
  SANDBOX_RESET_SCOPE_BLOCKCHAIN_ONLY = 2;  // Reset Besu ledger to genesis
  SANDBOX_RESET_SCOPE_MOCK_DATA_RESEED = 3; // Clear and reseed orders/users
}

enum FaucetTokenType {
  FAUCET_TOKEN_TYPE_UNSPECIFIED = 0;
  FAUCET_TOKEN_TYPE_GAS_NATIVE = 1;        // Testnet Besu Gas Token
  FAUCET_TOKEN_TYPE_CBDC_E_RUPEE = 2;      // Simulated RBI e-Rupee
  FAUCET_TOKEN_TYPE_MOCK_EQUITY_RELIANCE = 3;
  FAUCET_TOKEN_TYPE_MOCK_EQUITY_TATA = 4;
}

message TriggerSandboxResetRequest {
  string request_id = 1;
  SandboxResetScope scope = 2;
  string target_git_tag = 3;
  bool seed_default_market_depth = 4;
}

message TriggerSandboxResetResponse {
  string reset_id = 1;
  string status = 2; // "COMPLETED", "IN_PROGRESS", "FAILED"
  uint64 block_height = 3;
  repeated string deployed_contracts = 4;
  int64 elapsed_ms = 5;
}

message DisburseFaucetTokensRequest {
  string beneficiary_wallet_address = 1;
  FaucetTokenType token_type = 2;
  uint64 requested_amount_e6 = 3;
  string client_ip = 4;
}

message DisburseFaucetTokensResponse {
  string tx_hash = 1;
  uint64 disbursed_amount_e6 = 2;
  uint64 remaining_daily_quota_e6 = 3;
  int64 disbursed_at_ns = 4;
}

service SandboxOrchestratorService {
  rpc TriggerSandboxReset (TriggerSandboxResetRequest) returns (TriggerSandboxResetResponse);
  rpc DisburseFaucetTokens (DisburseFaucetTokensRequest) returns (DisburseFaucetTokensResponse);
}
```

### Cilium eBPF Network Isolation Manifest (`infra/sandbox-orchestrator/manifests/cilium-network-isolation.yaml`)
```yaml
apiVersion: "cilium.io/v2"
kind: CiliumNetworkPolicy
metadata:
  name: enforce-sandbox-mainnet-isolation
  namespace: testnet-sandbox
spec:
  description: "Strict isolation policy preventing testnet workloads from accessing mainnet resources"
  endpointSelector:
    matchLabels:
      environment: testnet
  ingress:
    - fromEndpoints:
        - matchLabels:
            environment: testnet
    - fromCIDR:
        - 10.100.0.0/16 # Testnet VPC CIDR only
  egress:
    - toEndpoints:
        - matchLabels:
            environment: testnet
    - toCIDR:
        - 10.100.0.0/16 # Testnet internal CIDR
    - toEntities:
        - world # Public internet for external developer webhooks
    - toEntities:
        - cluster # Kubernetes DNS within testnet cluster
  # Hard denial of any route toward Mainnet VPC (10.200.0.0/16)
```

### PostgreSQL Database Schema (`infra/sandbox-orchestrator/migrations/001_sandbox_schema.sql`)
```sql
CREATE TABLE sandbox_environments (
    env_id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    chain_id BIGINT NOT NULL DEFAULT 13371,
    besu_rpc_endpoint VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_reset_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sandbox_mock_entities (
    entity_id VARCHAR(64) PRIMARY KEY,
    entity_type VARCHAR(32) NOT NULL, -- 'INDIVIDUAL', 'CORPORATE', 'FPI_CAT_1'
    synthetic_pan VARCHAR(10) NOT NULL UNIQUE,
    synthetic_aadhaar_hash VARCHAR(64) NOT NULL,
    wallet_address VARCHAR(42) NOT NULL UNIQUE,
    kyc_status VARCHAR(24) NOT NULL DEFAULT 'KYC_APPROVED',
    is_pep_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE faucet_disbursement_logs (
    disbursement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    beneficiary_wallet VARCHAR(42) NOT NULL,
    token_type VARCHAR(32) NOT NULL,
    amount_e6 BIGINT NOT NULL,
    tx_hash VARCHAR(66) NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    disbursed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chaos_experiment_runs (
    experiment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_name VARCHAR(64) NOT NULL,
    target_service VARCHAR(64) NOT NULL,
    fault_type VARCHAR(32) NOT NULL, -- 'LATENCY', 'PACKET_LOSS', 'POD_KILL', 'GATEWAY_TIMEOUT'
    duration_seconds INT NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'RUNNING',
    recovery_detected BOOLEAN NOT NULL DEFAULT FALSE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_faucet_beneficiary_time ON faucet_disbursement_logs(beneficiary_wallet, disbursed_at);
CREATE INDEX idx_chaos_runs_status ON chaos_experiment_runs(status, started_at);
```

## Security & Compliance Notes
- **Cryptographic Boundary Enforcement:** Testnet and Mainnet keys are maintained under completely independent key hierarchies. Testnet deployer private keys have zero cryptographic authority on Mainnet contracts or validator nodes.
- **Zero Real-World PII:** The mock identity generator creates mathematically syntactically valid yet strictly synthetic PAN, Aadhaar, and bank account numbers, ensuring zero DPDP Act compliance exposure in test environments.
- **Network Air-Gap Assertion:** Cilium network policies prohibit all routing, DNS lookups, or TCP connections originating from testnet namespaces targeted at production VPC CIDRs.
- **Deterministic Fee Enforcement:** Testnet smart contracts and matching engines run the identical 0.00% fee (No fee at all) logic (0.00% fee at launch; future fee parameters governed by FeeController.sol) to ensure realistic economic behavior during sandbox simulations.

## Acceptance Criteria
- [ ] Sandbox orchestrator completes a full environment reset (ledger + databases + seed data) in under 60 seconds.
- [ ] Cilium network policies drop 100% of egress packets from testnet pods directed at mainnet VPC IP ranges.
- [ ] Mock depository engine successfully simulates NSDL/CDSL end-of-day allocation batch feeds.
- [ ] Faucet service dispenses testnet gas and synthetic e₹ CBDC tokens while enforcing daily per-IP/wallet rate limits.
- [ ] Mock KYC service responds to DigiLocker and PAN validation requests with deterministic synthetic payloads.
- [ ] Chaos injection harness triggers validator and network faults without corrupting testnet state consistency.
- [ ] 0.00% (Zero Fee) flat platform fee (0.00% fee at launch (future fee parameters governed by FeeController.sol)) is verified across all testnet simulated trades.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 108 (Environment Strategy), Prompt 302 (Network Topology), Prompt 802 (Kubernetes Architecture), Prompt 811 (Dual-Environment CI/CD).
- **Parallel Work:** Prompt 320 (Testnet Cluster & Faucet), Prompt 527 (Flutter Sandbox Mode), Prompt 609 (Developer Portal & Faucet UI).
- **Enables:** Safe, high-velocity developer integration, regulatory sandbox piloting, and comprehensive chaos resilience testing.
