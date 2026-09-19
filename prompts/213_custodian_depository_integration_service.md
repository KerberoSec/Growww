# 213 - Custodian & Depository Integration Service (NSDL/CDSL API Adapter)

## Purpose
The Custodian & Depository Integration Service is the foundational physical bridge between Growww's digital fractional investment ecosystem and India's regulated national depository infrastructure (National Securities Depository Limited - NSDL, and Central Depository Services Limited - CDSL). Under SEBI regulations, fractional digital representations cannot exist as synthetic or unbacked instruments; every digital unit issued to investors must map 1:1 to underlying dematerialized Indian equity shares safely lodged in institutional custody pool accounts.

This service automates the secure exchange of depository instruction files, trade settlement confirmations, daily Demat holding statements, and corporate action feeds. By providing continuous, verified synchronization between depository physical records and platform asset accounting, it ensures that platform operations maintain strict legal backing, zero-variance custody reconciliations, and full auditability.

## What You Are Building
A production-grade, highly secure Go microservice (`services/custody-adapter`) providing:
- **ISO 20022 / NSDL SPEED-e / CDSL Easiest Adapters:** Bidirectional batch file exchange and SFTP file processing engine.
- **Depository API Connectors:** Real-time and scheduled REST/SOAP API integration with Custodian banks and depositories.
- **Demat Settlement Confirmation Pipeline:** Ingestion and verification of physical Delivery-versus-Payment (DvP) and pool account transfers.
- **Physical Custody Proof Generator:** Extraction of daily Demat holding snapshots and cryptographic Merkle proof generation for on-chain reserve attestation.
- **Artifacts Delivered:**
 - `services/custody-adapter/cmd/server/main.go` - Go microservice entry point.
 - `services/custody-adapter/internal/depository/` - NSDL/CDSL parsers and protocol handlers.
 - `services/custody-adapter/internal/iso20022/` - ISO 20022 `sese.023`, `sese.024`, `sese.025`, `semt.002` parsers.
 - `services/custody-adapter/internal/hsm/` - PKCS#11 hardware security module digital signing client.
 - `proto/growww/custody/v1/custody.proto` - Internal gRPC service contracts.

## Scope Boundaries
- **In Scope:**
 - Secure SFTP/API communication with NSDL, CDSL, and institutional custodian banks (e.g., ICICI Custody, HDFC Bank Custody, SBI SG).
 - Parsing, validation, and generation of ISO 20022 settlement messages and proprietary depository flat files.
 - Signing outgoing depository transfer instructions via FIPS 140-2 Level 3 HSM.
 - Aggregating and publishing verified Demat holding snapshots to Kafka for downstream reconciliation and on-chain proof-of-reserve.
- **Out of Scope / Handled Elsewhere:**
 - On-chain token minting/burning execution (handled by Token Issuance Service, Prompt 303/304).
 - Internal three-way reconciliation logic between SQL, chain, and custody (handled by Reconciliation Service, Prompt 215).
 - Order matching and internal order book management (handled by Prompt 205).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `pgx/v5` for PostgreSQL access, `go-pkcs11` for hardware security module signing, and `confluent-kafka-go` for reliable event streaming.
- **Justification:** Go is chosen for its superior concurrency model when handling multiple simultaneous SFTP and API streams, low memory overhead, strong static typing, and robust standard library cryptography support required for processing high-volume depository transaction files under strict latency windows.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for custody metadata and transaction audit storage.
 - Redis 7.2+ for file transfer deduplication and transmission state locking.
 - Apache Kafka 3.7+ for high-throughput event messaging.
 - `github.com/miekg/pkcs11` for HSM PKI integration.
 - `github.com/pkg/sftp` for resilient, encrypted file transfers.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Stores custody accounts, batch file metadata, settlement transaction logs, and daily snapshot records.
- **Redis 7.2:** Distributed file locks, SFTP polling markers, and duplicate message prevention.
- **Kafka Topics:**
 - Publishes: `custody.settlement.confirmed`, `custody.holdings.snapshot`, `custody.file.processed`, `custody.discrepancy.flagged`.
 - Subscribes: `trade.settlement.instruction`, `corporate_action.request`.
- **External Depository Endpoints:** SEBI Custodian SFTP server, NSDL SPEED-e Gateway, CDSL Easiest API over leased lines/IPSec VPN.
- **Key Management:** AWS CloudHSM or HashiCorp Vault Transit engine for digital signature creation.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium ledger running QBFT consensus with 2-second block finality.
- **Interaction Model:** The Custody Adapter generates cryptographic Merkle root proofs of daily Demat share balances. It forwards these attestations to the Proof-of-Reserve Relayer, which writes them to `ProofOfReserveRegistry.sol`.
- **Zero PII Guarantee:** No individual investor Demat account details or identity data are ever published to the ledger. All records on-chain reference institutional omnibus custody pool accounts, aggregate ISIN share quantities, and cryptographic depository settlement reference hashes.
- **Event Synchronization:** Listens for `TokenMintRequested` and `TokenRedemptionApproved` on-chain events via Kafka indexers to trigger corresponding depository pool allocations or releases.

## Step-by-Step Build Instructions
1. Initialize Go module structure under `services/custody-adapter` adhering to clean architecture principles (Prompt 106).
2. Define Protobuf definitions in `proto/growww/custody/v1/custody.proto` and generate Go gRPC client/server stubs.
3. Configure PostgreSQL schema migrations for `custody_pool_accounts`, `depository_transactions`, and `demat_snapshots`.
4. Implement the PKCS#11 cryptographic signer module to sign ISO 20022 and NSDL batch instructions using HSM keys.
5. Build the SFTP communication client featuring exponential backoff, automatic keep-alive, session pooling, and mTLS over IPSec VPN.
6. Implement parsers for NSDL SPEED-e fixed-width files and CDSL XML/CSV settlement report feeds.
7. Implement ISO 20022 XML serializers and deserializers for `sese.023` (Securities Settlement Transaction Instruction) and `sese.025` (Securities Settlement Transaction Confirmation).
8. Develop the batch processing pipeline with idempotency filters to prevent duplicate trade instruction uploads.
9. Implement the daily Demat holding snapshot aggregator that computes SHA-256 Merkle roots over ISIN balance sets.
10. Integrate Kafka event publishers producing `custody.settlement.confirmed` and `custody.holdings.snapshot` messages.
11. Build health checks (`/healthz`, `/livez`), Prometheus metric collectors (file processing duration, SFTP error rates), and OpenTelemetry tracing.
12. Write end-to-end integration tests using mock SFTP servers and depository simulator fixtures achieving >=85% code coverage.

## Interfaces / Contracts

### Protobuf Definition (`custody.proto`)
```protobuf
syntax = "proto3";

package growww.custody.v1;

option go_package = "github.com/growww/services/custody-adapter/gen/v1;custodyv1";

service CustodyAdapterService {
  rpc SubmitSettlementInstruction (SettlementInstructionRequest) returns (SettlementInstructionResponse);
  rpc GetHoldingSnapshot (HoldingSnapshotRequest) returns (HoldingSnapshotResponse);
  rpc VerifyCustodyHealth (CustodyHealthRequest) returns (CustodyHealthResponse);
}

message SettlementInstructionRequest {
  string instruction_id = 1;
  string isin = 2;
  int64 share_quantity = 3;
  string trade_type = 4; // BUY / SELL
  string depository = 5; // NSDL / CDSL
  string settlement_date = 6;
  string pool_account_id = 7;
}

message SettlementInstructionResponse {
  string instruction_id = 1;
  string depository_reference = 2;
  string status = 3; // ACCEPTED / REJECTED / PENDING
  string message = 4;
  int64 timestamp = 5;
}

message HoldingSnapshotRequest {
  string depository = 1;
  string snapshot_date = 2;
}

message HoldingSnapshotResponse {
  string snapshot_id = 1;
  string depository = 2;
  string merkle_root = 3;
  int64 total_securities_count = 4;
  repeated SecurityBalance balances = 5;
  int64 generated_at = 6;
}

message SecurityBalance {
  string isin = 1;
  string symbol = 2;
  int64 settled_quantity = 3;
  int64 pledged_quantity = 4;
  int64 in_transit_quantity = 5;
}

message CustodyHealthRequest {}
message CustodyHealthResponse {
  bool sftp_connected = 1;
  bool hsm_healthy = 2;
  bool depository_api_reachable = 3;
  int64 last_sync_timestamp = 4;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE custody_pool_accounts (
    account_id VARCHAR(64) PRIMARY KEY,
    depository VARCHAR(16) NOT NULL CHECK (depository IN ('NSDL', 'CDSL')),
    dp_id VARCHAR(32) NOT NULL,
    client_id VARCHAR(32) NOT NULL,
    custodian_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE depository_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instruction_id VARCHAR(64) UNIQUE NOT NULL,
    depository_ref VARCHAR(128),
    isin VARCHAR(12) NOT NULL,
    quantity NUMERIC(18, 4) NOT NULL,
    direction VARCHAR(8) NOT NULL CHECK (direction IN ('INFLOW', 'OUTFLOW')),
    status VARCHAR(32) NOT NULL,
    file_batch_id VARCHAR(64),
    raw_payload_hash VARCHAR(64) NOT NULL,
    executed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE demat_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date DATE NOT NULL,
    depository VARCHAR(16) NOT NULL,
    merkle_root VARCHAR(64) NOT NULL,
    total_isin_count INT NOT NULL,
    payload_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_depository_snapshot_date UNIQUE (snapshot_date, depository)
);
```

## Security & Compliance Notes
- **SEBI Custodian Regulations (1996) Compliance:** Ensures clear segregation between platform client funds/securities and institutional proprietary assets. All assets reside strictly in designated Client Omnibus Pool Accounts.
- **FIPS 140-2 Level 3 Cryptographic Signing:** Depository instruction files must be signed using private keys residing exclusively within HSM; plaintext private keys are prohibited from application memory.
- **Zero Investor PII:** Only institutional Demat account IDs and ISIN quantities are transmitted to depository gateways; no individual user PII is ever leaked.
- **mTLS & Leased Line Isolation:** All network connectivity to depository servers operates over encrypted mTLS tunnels terminated within a dedicated VPC with static IP whitelisting.

## Acceptance Criteria
- [ ] Go microservice cleanly compiles with zero linter errors using Go 1.22+.
- [ ] ISO 20022 and NSDL flat-file parsers correctly parse test vector files with 100% roundtrip accuracy.
- [ ] HSM digital signing integration successfully signs test payloads without exporting private keys.
- [ ] SFTP client reliably handles network disconnects and reconnects using exponential backoff without dropping batch files.
- [ ] Merkle root computation over Demat balances matches independent verification test harness.
- [ ] Kafka events are successfully published with Avro/JSON schemas and ingested by the reconciliation test pipeline.
- [ ] Unit and integration test suite achieves >=85% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Standards), Prompt 104 (Kafka Standards), Prompt 109 (Secrets & HSM).
- **Subsequent / Parallel Tasks:** Prompt 215 (Reconciliation Service), Prompt 303 (Token Issuance Smart Contract), Prompt 308 (On-Chain Proof-of-Reserve).
