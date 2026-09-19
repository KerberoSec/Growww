# 243 - MCX Commodity & Warehouse Receipt Adapter (Go / Python / WDRA eNWR / NERL / CCRL)

## Purpose
In physical commodity trading, institutional tokenization, and multi-asset clearing ecosystems, digital representations of commodities cannot exist as unbacked synthetic claims. Under Warehousing Development and Regulatory Authority (WDRA) guidelines, Multi Commodity Exchange of India (MCX) delivery bylaws, and SEBI regulations for electronic commodity derivatives, every tokenized commodity unit (such as 1 gram of Gold 999, 1 kilogram of Silver 999, or metric tons of base metals and agricultural commodities) must map 1:1 to underlying physical stock securely lodged in accredited vaults and registered as electronic Negotiable Warehouse Receipts (eNWR) across national electronic repositories: National E-Repository Limited (NERL) and CDSL Commodity Repository Limited (CCRL).

The **MCX Commodity & Warehouse Receipt Adapter** (`services/commodity-vault-adapter`) serves as the mission-critical physical-to-digital custody bridge. It automates repository API/SFTP exchanges with NERL and CCRL, validates Bureau of Indian Standards (BIS) hallmark and London Bullion Market Association (LBMA) good delivery assay certificates, tracks individual bar serial numbers and vault slot coordinates, coordinates 1:1 asset-backed token minting on permissioned Hyperledger Besu, and orchestrates end-to-end physical delivery redemption workflows with armored vault logistics and dual-custody verification.

## What You Are Building
A production-grade, fault-tolerant microservice (`services/commodity-vault-adapter`) built using Go 1.22+ for high-throughput repository streaming and gRPC interfaces, and Python 3.12+ for assay document OCR/validation and statistical inventory reconciliation. Concrete deliverables include:
- **NERL & CCRL eNWR Repository Connectors:** ISO 20022 and proprietary REST/SOAP/SFTP API adapters for National E-Repository Limited (NERL) and CDSL Commodity Repository Limited (CCRL), managing eNWR creation, electronic pledging, unpledging, title transfer, and rematerialization.
- **MCX Clearing & Settlement Delivery Adapter:** Integration gateway interfacing with MCX Clearing Corporation (MCXCCL) for physical commodity delivery pay-in/pay-out allocations, tender period delivery notices, and settlement confirmations.
- **Bullion & Commodity Assay Certification Engine:** Python parsing engine extracting and validating BIS hallmark data, X-ray fluorescence (XRF) and fire assay purity reports, refinery origin certifications, and assayer digital signatures against accredited laboratory public key registries.
- **Individual Bar Serial & Physical Vault Inventory Registry:** Real-time tracking system managing bar-level metadata (refinery serial number, brand, gross weight, net fine weight, fineness 999.0 / 999.9, vault operator, vault bay/rack/slot coordinates, and tamper-evident RFID/QR seal identifiers).
- **1:1 Commodity Token Minting Bridge:** Dual-control attestation pipeline that verifies physical deposit and eNWR immobilization before generating cryptographically signed minting authorizations consumed by the Token Issuance Service (Prompt 303).
- **Physical Delivery Redemption & Armored Logistics Orchestrator:** Secure redemption engine managing on-chain token lockup and burn, eNWR rematerialization into physical withdrawal orders, time-locked OTP/PIN generation for vault pickup, and API integration with insured armored logistics carriers (e.g., Brink's, Loomis, Sequel Logistics, Malca-Amit).
- **Continuous IoT Vault Telemetry & Proof-of-Reserve Attestation Daemon:** Scheduled auditing daemon ingesting real-time IoT vault environmental telemetry (temperature, humidity, vibration, optical door sensors) and publishing daily Sparse Merkle Tree (SMT) inventory root hashes to `ProofOfReserveRegistry.sol` (Prompt 308).

## Scope Boundaries
- **In Scope:**
  - Connectivity with WDRA repositories (NERL, CCRL) and MCXCCL delivery settlement endpoints over mTLS and secure VPN.
  - Automated eNWR lifecycle operations: deposit acknowledgment, lien marking/pledging, title transfer, rematerialization, and extinguishment.
  - Assay certificate ingestion, optical character recognition (OCR), signature validation, and fineness verification (Gold: 995, 999, 999.9; Silver: 999).
  - Bar serial registry tracking gross weight, tare, net fine weight, refiner accreditation, and vault location coordinates.
  - Token mint authorization generation (DvP asset backing proof) and token burn redemption reconciliation.
  - Physical delivery order management, armored carrier tracking, gate pass generation, and dual-custody vault handoff verification.
  - Daily cryptographic Proof-of-Reserve Merkle root calculation over all vaulted commodity inventory.
- **Out of Scope / Handled Elsewhere:**
  - High-frequency matching engine central limit order book execution (handled in Prompt 205).
  - Cash wallet ledger double-entry accounting (handled in Prompt 203).
  - On-chain smart contract token minting/burning execution (handled in Prompt 303 and Prompt 304).
  - General equity depository settlement with NSDL/CDSL (handled in Prompt 213).
  - SPAN margin calculations and derivatives liquidations (handled in Prompt 241).

## Technology to Use
- **Primary Languages & Runtimes:**
  - **Go 1.22+:** Core daemon, repository API/SFTP connectors, gRPC services, Kafka event streaming, and dual-custody redemption workflows (`services/commodity-vault-adapter`).
  - **Python 3.12+:** Assay certificate OCR parsing worker, PDF metadata extraction (`pdfplumber`, `pydantic`), and statistical assay variance analysis.
- **Relational Database:** **PostgreSQL 16+** with `pgx/v5` and `sqlc` for storing vault registries, bar serials, eNWR records, assay certificates, and redemption orders.
- **In-Memory Cache & Distributed Lock:** **Redis 7.2+ Cluster** for idempotency locking, temporary redemption OTP/PIN caching, and repository transaction state tracking.
- **Message Broker:** **Apache Kafka 3.7+** for asynchronous event ingestion and broadcast (`commodity.enwr.*`, `commodity.bar.*`, `commodity.redemption.*`).
- **Cryptographic Security & Signing:** **PKCS#11 / FIPS 140-2 Level 3 HSM** (AWS CloudHSM or HashiCorp Vault Transit Engine) for signing eNWR instructions and mint authorization payloads.
- **File Transfer & Network Security:** Secure SFTP client (`golang.org/x/crypto/ssh`), ISO 20022 XML parsers, and mutual TLS (mTLS) over dedicated IPSec VPN tunnels.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:**
  - `commodity_vault_locations`, `commodity_batches`, `bar_serial_inventory`, `enwr_records`, `assay_certificates`, `redemption_delivery_orders`, `vault_audit_snapshots`, `vault_iot_telemetry`.
- **Redis 7.2 Keys:**
  - `commodity:vault:{vault_id}:capacity`: Vault capacity and current fine weight utilization.
  - `commodity:bar:{serial_no}:lock`: Distributed lock for bar-level status modifications.
  - `commodity:enwr:{enwr_number}:state`: Active repository state and lien status.
  - `commodity:redemption:{order_id}:otp`: Secure hash of collection PIN/OTP.
- **Apache Kafka Topics:**
  - Subscribes to: `commodity.deposit.initiated`, `commodity.token_burn.confirmed`, `trade.settlement.commodity_allocated`.
  - Publishes to: `commodity.enwr.pledged.v1`, `commodity.bar.registered.v1`, `commodity.token_mint_authorized.v1`, `commodity.redemption.dispatched.v1`, `commodity.audit.attestation.v1`.
- **Token Issuance Service (Prompt 303) & Token Redemption Service (Prompt 304):** Exchanges cryptographically signed mint authorizations and burn completion receipts.
- **Reconciliation Service (Prompt 215):** Supplies daily physical inventory feeds for three-way reconciliation (Physical Vaults vs eNWR Repositories vs Hyperledger Besu Ledger).
- **Proof-of-Reserve Registry (Prompt 308):** Ingests daily Sparse Merkle Tree (SMT) root hashes representing verified physical grams in vault custody.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **1:1 Physical Custody Backing Invariant:** Every on-chain commodity token (e.g., `tGOLD`, `tSILVER`) issued under ERC-3643 / CMTA standards on Hyperledger Besu must have an immutable reference to a specific eNWR number, bar serial hash list, and accredited vault repository certificate.
- **On-Chain Reserve Root Publication:** The service computes a daily Sparse Merkle Tree (SMT) of all vaulted bar serials, gross weights, fineness, and eNWR IDs. The resulting Merkle root hash is signed by the vault custodian HSM and submitted to `ProofOfReserveRegistry.sol`.
- **Cryptographic Mint Authorization:** When a new physical deposit or eNWR transfer is finalized, the service generates an EIP-712 typed signature containing `{deposit_id, commodity_code, fine_weight_grams, enwr_hash, bar_count, recipient_address, nonce, expiry}`. The smart contract validates this signature against the authorized Custodian signer key before minting tokens.
- **Atomic Redemption Escrow & Burn Verification:** For physical redemptions, tokens are locked in `TokenRedemption.sol`. Upon physical bar release and gate pass dispatch from the vault, the service submits the signed dispatch confirmation to trigger the on-chain token burn, ensuring tokens cannot circulate once physical custody is relinquished.
- **Zero On-Chain PII Guarantee:** No investor identities, residential delivery addresses, or KYC documents are ever written to the blockchain. The ledger stores only pseudonymous `wallet_address`, `redemption_id` UUID, commodity fine weight, and cryptographic hash digests of delivery receipts.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Directory & Architecture:** Initialize Go microservice under `services/commodity-vault-adapter` and Python OCR worker under `services/commodity-vault-adapter/ocr_worker` with clean architecture separation (domain, repository, usecases, transport, adapters).
2. **Define Protocol Buffers Contracts:** Author `proto/growww/commodity/v1/commodity_vault.proto` specifying RPCs for eNWR pledging, bar registration, assay validation, mint authorization generation, physical redemption request, and vault audit reporting. Generate Go and Python stubs.
3. **Configure PostgreSQL Schema Migrations:** Write database migration scripts creating tables for vaults, commodity batches, individual bar serials, eNWR states, assay reports, redemption delivery orders, and IoT telemetry records.
4. **Implement NERL & CCRL Repository Clients:** Build REST/SOAP and SFTP connectors with mTLS authentication, ISO 20022 message parsers (`setr`, `semt` extensions for commodities), and automated XML signature generation for WDRA eNWR electronic title transfers and lien markings.
5. **Implement MCXCCL Delivery Gateway:** Build integration adapter consuming MCX tender notice files, delivery allocation reports, and settlement variance adjustments to reconcile exchange-cleared commodity positions.
6. **Implement Python Assay Certificate OCR & Validation Engine:** Construct worker utilizing PyMuPDF and pdfplumber to ingest PDF assay reports from BIS-recognized assaying centers (e.g., MMTC-PAMP, India Government Mint, certified private assayers). Extract refinery stamp, fineness percentage, gross weight, melted weight, and assayer digital signatures with Pydantic schema validation.
7. **Implement Bar Serial Inventory Registry:** Build Go domain logic for registering individual commodity bars. Validate that $\text{Net Fine Weight} = \text{Gross Weight} \times \frac{\text{Fineness}}{1000}$. Assign physical vault coordinates (`vault_id`, `room_id`, `rack_id`, `shelf_id`, `slot_id`) and record tamper-evident security seal numbers.
8. **Build 1:1 Token Mint Authorization Generator:** Implement dual-control validation verifying that physical bars are deposited, eNWR is successfully pledged/transferred to the platform omnibus repository account, and assay tests pass. Generate EIP-712 structured cryptographic mint authorizations signed via PKCS#11 HSM.
9. **Implement Physical Delivery Redemption State Machine:**
   - **Step 1 (Redemption Request):** Ingest user redemption request, verify token balance, and generate on-chain lockup instruction.
   - **Step 2 (Bar Allocation):** Execute FIFO/LIFO or optimal weight-matching algorithm to allocate specific physical bars matching requested weight.
   - **Step 3 (eNWR Rematerialization):** Submit rematerialization instruction to NERL/CCRL repository to convert electronic receipt to physical delivery order.
   - **Step 4 (Vault Gate Pass & Armored Carrier Integration):** Dispatch delivery instruction to vault operator and armored carrier (Brink's/Sequel) with pickup window, vehicle registration, and driver identity verification.
   - **Step 5 (Dual-Custody Handoff):** Generate secure time-locked OTP/PIN for armored courier pickup and final customer delivery handoff.
   - **Step 6 (Dispatch & Token Burn):** Upon confirmed vault dispatch, emit Kafka event to trigger on-chain token burn in `TokenRedemption.sol`.
10. **Build Armored Carrier Logistics Tracking Adapter:** Integrate webhook and REST tracking APIs for armored logistics providers (Sequel Logistics, Brink's Secure Logistics) to monitor transit status, GPS telemetry, transit insurance policy binding, and doorstep recipient biometric/OTP verification.
11. **Implement IoT Vault Environmental Monitor:** Build ingestion daemon for vault sensor feeds (temperature, relative humidity, vibration, seismic, PIR motion, optical door contacts). Trigger automated anomaly alerts if environmental thresholds are breached (e.g., humidity $> 65\%$ for silver storage).
12. **Implement Daily Cryptographic Proof-of-Reserve Generator:** Build scheduled worker that iterates all unencumbered, active vaulted bars, constructs a 256-bit Sparse Merkle Tree (SMT), and publishes the root hash alongside total verified fine grams to Kafka topic `commodity.audit.attestation.v1` and on-chain registry.
13. **Implement Idempotency & Distributed Locking:** Use Redis Redlock for all bar allocation and eNWR state transitions to prevent race conditions, double-allocation of physical bars, or duplicate rematerialization calls.
14. **Configure Prometheus Metrics & OpenTelemetry Instrumentation:** Instrument metrics for `commodity_vault_bars_total`, `commodity_fine_grams_vaulted`, `enwr_operations_duration_seconds`, `assay_ocr_processing_latency_seconds`, `redemption_orders_active`, and `vault_iot_anomaly_events_total`.
15. **Implement Comprehensive Test Suite & Mocks:** Write unit, integration, and mock repository tests simulating NERL/CCRL SOAP/SFTP exchanges, assayer PDF fixtures, bar weight matching edge cases, and end-to-end redemption lifecycle execution with $\ge 85\%$ code coverage.

## Interfaces / Contracts

### Protobuf Service & Message Definitions (`proto/growww/commodity/v1/commodity_vault.proto`)

```protobuf
syntax = "proto3";

package growww.commodity.v1;

option go_package = "growww/commodity/v1;commodityv1";

service CommodityVaultAdapterService {
  rpc RegisterPhysicalDeposit (RegisterPhysicalDepositRequest) returns (RegisterPhysicalDepositResponse);
  rpc VerifyAssayCertificate (VerifyAssayCertificateRequest) returns (VerifyAssayCertificateResponse);
  rpc PledgeENWR (PledgeENWRRequest) returns (PledgeENWRResponse);
  rpc GenerateMintAuthorization (GenerateMintAuthorizationRequest) returns (GenerateMintAuthorizationResponse);
  rpc RequestPhysicalRedemption (RequestPhysicalRedemptionRequest) returns (RequestPhysicalRedemptionResponse);
  rpc ConfirmVaultDispatch (ConfirmVaultDispatchRequest) returns (ConfirmVaultDispatchResponse);
  rpc ConfirmDeliveryReceipt (ConfirmDeliveryReceiptRequest) returns (ConfirmDeliveryReceiptResponse);
  rpc GetVaultInventorySummary (GetVaultInventorySummaryRequest) returns (GetVaultInventorySummaryResponse);
  rpc GenerateProofOfReserveSnapshot (GenerateProofOfReserveSnapshotRequest) returns (GenerateProofOfReserveSnapshotResponse);
}

enum CommodityType {
  COMMODITY_TYPE_UNSPECIFIED = 0;
  COMMODITY_TYPE_GOLD_995 = 1;
  COMMODITY_TYPE_GOLD_999 = 2;
  COMMODITY_TYPE_GOLD_9999 = 3;
  COMMODITY_TYPE_SILVER_999 = 4;
  COMMODITY_TYPE_COPPER = 5;
  COMMODITY_TYPE_ALUMINIUM = 6;
}

enum RepositoryType {
  REPOSITORY_TYPE_UNSPECIFIED = 0;
  REPOSITORY_TYPE_NERL = 1;
  REPOSITORY_TYPE_CCRL = 2;
  REPOSITORY_TYPE_MCX_ACCREDITED_VAULT = 3;
}

enum BarStatus {
  BAR_STATUS_UNSPECIFIED = 0;
  BAR_STATUS_DEPOSITED = 1;
  BAR_STATUS_ASSAY_VERIFIED = 2;
  BAR_STATUS_ENWR_PLEDGED = 3;
  BAR_STATUS_TOKENIZED = 4;
  BAR_STATUS_REDEMPTION_LOCKED = 5;
  BAR_STATUS_DISPATCHED = 6;
  BAR_STATUS_EXTINGUISHED = 7;
}

enum RedemptionStatus {
  REDEMPTION_STATUS_UNSPECIFIED = 0;
  REDEMPTION_STATUS_REQUESTED = 1;
  REDEMPTION_STATUS_BARS_ALLOCATED = 2;
  REDEMPTION_STATUS_ENWR_REMATERIALIZED = 3;
  REDEMPTION_STATUS_VAULT_GATE_PASS_ISSUED = 4;
  REDEMPTION_STATUS_COURIER_PICKED_UP = 5;
  REDEMPTION_STATUS_IN_TRANSIT = 6;
  REDEMPTION_STATUS_DELIVERED = 7;
  REDEMPTION_STATUS_CANCELLED = 8;
}

message BarDetail {
  string bar_serial_number = 1;
  string refiner_name = 2;
  string refiner_code = 3; // LBMA / BIS refiner ID
  CommodityType commodity_type = 4;
  string gross_weight_grams = 5; // Decimal string
  string fineness = 6; // e.g., "999.00"
  string net_fine_weight_grams = 7; // Decimal string
  string vault_id = 8;
  string vault_location_code = 9; // Bay-Rack-Shelf-Slot
  string rfid_tag_id = 10;
  string security_seal_number = 11;
  int64 manufacture_year = 12;
}

message RegisterPhysicalDepositRequest {
  string deposit_id = 1;
  RepositoryType repository = 2;
  string enwr_number = 3;
  string vault_id = 4;
  CommodityType commodity_type = 5;
  repeated BarDetail bars = 6;
  string assayer_id = 7;
  string assay_certificate_ref = 8;
  int64 deposit_timestamp_ns = 9;
}

message RegisterPhysicalDepositResponse {
  bool success = 1;
  string deposit_id = 2;
  int32 registered_bar_count = 3;
  string total_fine_weight_grams = 4;
  string status = 5;
}

message VerifyAssayCertificateRequest {
  string certificate_id = 1;
  string assayer_id = 2;
  string certificate_pdf_base64 = 3;
  CommodityType commodity_type = 4;
  repeated string expected_bar_serials = 5;
}

message VerifyAssayCertificateResponse {
  bool is_valid = 1;
  string certificate_id = 2;
  string assayer_name = 3;
  string tested_fineness = 4;
  bool bis_hallmark_present = 5;
  string digital_signature_status = 6; // VERIFIED, INVALID, UNTRUSTED
  repeated string validated_bar_serials = 7;
  string verification_error = 8;
}

message PledgeENWRRequest {
  string enwr_number = 1;
  RepositoryType repository = 2;
  string platform_pledgee_id = 3;
  string pledge_reason = 4; // TOKENIZATION_CUSTODY
  int64 requested_at_ns = 5;
}

message PledgeENWRResponse {
  bool success = 1;
  string enwr_number = 2;
  string repository_transaction_ref = 3;
  string lien_status = 4; // PLEDGED, PENDING, REJECTED
  int64 confirmed_at_ns = 5;
}

message GenerateMintAuthorizationRequest {
  string deposit_id = 1;
  string enwr_number = 2;
  string recipient_wallet_address = 3;
  int64 authorization_expiry_ns = 4;
}

message GenerateMintAuthorizationResponse {
  bool authorized = 1;
  string deposit_id = 2;
  string commodity_symbol = 3; // e.g., "tGOLD"
  string fine_weight_grams = 4;
  uint64 token_amount_atomic = 5;
  string signature_eip712 = 6;
  string signer_public_key = 7;
  int64 expires_at_ns = 8;
}

message RequestPhysicalRedemptionRequest {
  string redemption_id = 1;
  string user_id = 2;
  string wallet_address = 3;
  CommodityType commodity_type = 4;
  string requested_fine_weight_grams = 5;
  string delivery_mode = 6; // VAULT_PICKUP or ARMORED_DOORSTEP
  string destination_vault_or_hub_id = 7;
  string delivery_address_encrypted = 8;
  int64 requested_at_ns = 9;
}

message RequestPhysicalRedemptionResponse {
  bool accepted = 1;
  string redemption_id = 2;
  RedemptionStatus status = 3;
  repeated BarDetail allocated_bars = 4;
  string total_allocated_fine_weight_grams = 5;
  string rematerialization_fee_inr = 6;
  string logistics_insurance_fee_inr = 7;
}

message ConfirmVaultDispatchRequest {
  string redemption_id = 1;
  string vault_id = 2;
  string gate_pass_number = 3;
  string courier_company = 4;
  string courier_consignment_number = 5;
  string courier_driver_id_hash = 6;
  repeated string dispatched_bar_serials = 7;
  int64 dispatched_at_ns = 8;
}

message ConfirmVaultDispatchResponse {
  bool success = 1;
  string redemption_id = 2;
  string token_burn_receipt_id = 3;
  int64 processed_at_ns = 4;
}

message ConfirmDeliveryReceiptRequest {
  string redemption_id = 1;
  string delivery_otp = 2;
  string recipient_signature_hash = 3;
  int64 delivered_at_ns = 4;
}

message ConfirmDeliveryReceiptResponse {
  bool completed = 1;
  string redemption_id = 2;
  RedemptionStatus final_status = 3;
  int64 finalized_at_ns = 4;
}

message GetVaultInventorySummaryRequest {
  string vault_id = 1; // Empty for all vaults
  CommodityType commodity_type = 2;
}

message GetVaultInventorySummaryResponse {
  int64 total_active_bars = 1;
  string total_gross_weight_grams = 2;
  string total_fine_weight_grams = 3;
  string tokenized_fine_weight_grams = 4;
  string unallocated_fine_weight_grams = 5;
  string locked_for_redemption_grams = 6;
  int64 last_audit_timestamp_ns = 7;
}

message GenerateProofOfReserveSnapshotRequest {
  int64 snapshot_timestamp_ns = 1;
}

message GenerateProofOfReserveSnapshotResponse {
  string snapshot_id = 1;
  string merkle_root_hex = 2;
  int64 total_leaf_count = 3;
  string total_verified_fine_grams = 4;
  string attestation_signature = 5;
  int64 generated_at_ns = 6;
}
```

### PostgreSQL Schema DDL (`schema/commodity_vault_schema.sql`)

```sql
-- Schema: commodity_vault_adapter

CREATE TYPE commodity_asset_type AS ENUM (
    'GOLD_995',
    'GOLD_999',
    'GOLD_9999',
    'SILVER_999',
    'COPPER',
    'ALUMINIUM'
);

CREATE TYPE bar_inventory_status AS ENUM (
    'DEPOSITED',
    'ASSAY_VERIFIED',
    'ENWR_PLEDGED',
    'TOKENIZED',
    'REDEMPTION_LOCKED',
    'DISPATCHED',
    'EXTINGUISHED'
);

CREATE TYPE enwr_repository_enum AS ENUM (
    'NERL',
    'CCRL',
    'MCX_ACCREDITED_VAULT'
);

CREATE TYPE redemption_order_status AS ENUM (
    'REQUESTED',
    'BARS_ALLOCATED',
    'ENWR_REMATERIALIZED',
    'VAULT_GATE_PASS_ISSUED',
    'COURIER_PICKED_UP',
    'IN_TRANSIT',
    'DELIVERED',
    'CANCELLED'
);

CREATE TABLE commodity_vault_locations (
    vault_id VARCHAR(64) PRIMARY KEY,
    vault_operator VARCHAR(128) NOT NULL,
    wdra_accreditation_number VARCHAR(128) NOT NULL UNIQUE,
    mcx_delivery_center_code VARCHAR(64),
    facility_name VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    pincode VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    max_capacity_kg NUMERIC(16, 4) NOT NULL,
    current_stored_kg NUMERIC(16, 4) NOT NULL DEFAULT 0.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE commodity_batches (
    batch_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deposit_ref VARCHAR(128) NOT NULL UNIQUE,
    vault_id VARCHAR(64) NOT NULL REFERENCES commodity_vault_locations(vault_id),
    commodity_type commodity_asset_type NOT NULL,
    enwr_number VARCHAR(128) NOT NULL,
    repository enwr_repository_enum NOT NULL,
    total_bars INT NOT NULL,
    total_gross_weight_grams NUMERIC(18, 4) NOT NULL,
    total_fine_weight_grams NUMERIC(18, 4) NOT NULL,
    declared_fineness NUMERIC(6, 2) NOT NULL,
    assayer_id VARCHAR(128) NOT NULL,
    assay_certificate_id VARCHAR(128),
    deposit_date DATE NOT NULL,
    is_pledged BOOLEAN NOT NULL DEFAULT FALSE,
    is_tokenized BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bar_serial_inventory (
    bar_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bar_serial_number VARCHAR(128) NOT NULL,
    refiner_name VARCHAR(128) NOT NULL,
    refiner_code VARCHAR(64) NOT NULL,
    commodity_type commodity_asset_type NOT NULL,
    batch_id UUID NOT NULL REFERENCES commodity_batches(batch_id),
    vault_id VARCHAR(64) NOT NULL REFERENCES commodity_vault_locations(vault_id),
    vault_location_code VARCHAR(64) NOT NULL, -- e.g., BAY-04/RACK-12/SLOT-08
    gross_weight_grams NUMERIC(14, 4) NOT NULL,
    fineness NUMERIC(6, 2) NOT NULL,
    net_fine_weight_grams NUMERIC(14, 4) NOT NULL,
    rfid_tag_id VARCHAR(128) UNIQUE,
    security_seal_number VARCHAR(128) UNIQUE,
    manufacture_year INT NOT NULL,
    status bar_inventory_status NOT NULL DEFAULT 'DEPOSITED',
    token_mint_tx_hash VARCHAR(66),
    on_chain_token_id VARCHAR(128),
    allocated_redemption_id UUID,
    dispatched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_bar_refiner_serial UNIQUE (refiner_code, bar_serial_number)
);

CREATE TABLE enwr_records (
    enwr_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enwr_number VARCHAR(128) NOT NULL UNIQUE,
    repository enwr_repository_enum NOT NULL,
    vault_id VARCHAR(64) NOT NULL REFERENCES commodity_vault_locations(vault_id),
    commodity_type commodity_asset_type NOT NULL,
    quantity_units INT NOT NULL,
    gross_weight_grams NUMERIC(18, 4) NOT NULL,
    net_fine_weight_grams NUMERIC(18, 4) NOT NULL,
    holder_client_id VARCHAR(128) NOT NULL,
    pledgee_client_id VARCHAR(128),
    lien_marked BOOLEAN NOT NULL DEFAULT FALSE,
    lien_reference VARCHAR(128),
    validity_start DATE NOT NULL,
    validity_expiry DATE NOT NULL,
    repository_status VARCHAR(64) NOT NULL, -- ACTIVE, PLEDGED, REMATERIALIZED, EXTINGUISHED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assay_certificates (
    certificate_id VARCHAR(128) PRIMARY KEY,
    assayer_name VARCHAR(128) NOT NULL,
    assayer_license_number VARCHAR(128) NOT NULL,
    commodity_type commodity_asset_type NOT NULL,
    batch_id UUID REFERENCES commodity_batches(batch_id),
    testing_method VARCHAR(64) NOT NULL, -- FIRE_ASSAY, XRF, TITRATION
    tested_fineness NUMERIC(6, 2) NOT NULL,
    bis_hallmark_verified BOOLEAN NOT NULL DEFAULT FALSE,
    certificate_document_hash VARCHAR(64) NOT NULL, -- SHA-256 of PDF
    digital_signature_hash VARCHAR(128),
    is_valid BOOLEAN NOT NULL DEFAULT TRUE,
    issued_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE redemption_delivery_orders (
    redemption_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    commodity_type commodity_asset_type NOT NULL,
    requested_fine_weight_grams NUMERIC(14, 4) NOT NULL,
    allocated_fine_weight_grams NUMERIC(14, 4) NOT NULL DEFAULT 0.0000,
    delivery_mode VARCHAR(32) NOT NULL, -- VAULT_PICKUP, ARMORED_DOORSTEP
    source_vault_id VARCHAR(64) NOT NULL REFERENCES commodity_vault_locations(vault_id),
    delivery_address_encrypted BYTEA,
    rematerialization_fee_inr NUMERIC(12, 2) NOT NULL,
    logistics_fee_inr NUMERIC(12, 2) NOT NULL,
    status redemption_order_status NOT NULL DEFAULT 'REQUESTED',
    gate_pass_number VARCHAR(128),
    armored_carrier_code VARCHAR(64), -- BRINKS, SEQUEL, LOOMIS
    consignment_tracking_number VARCHAR(128),
    pickup_otp_hash VARCHAR(128),
    delivery_otp_hash VARCHAR(128),
    token_lock_tx_hash VARCHAR(66),
    token_burn_tx_hash VARCHAR(66),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vault_audit_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_timestamp TIMESTAMPTZ NOT NULL,
    total_bars_audited INT NOT NULL,
    total_gross_weight_grams NUMERIC(18, 4) NOT NULL,
    total_fine_weight_grams NUMERIC(18, 4) NOT NULL,
    sparse_merkle_root_hex VARCHAR(66) NOT NULL,
    on_chain_tx_hash VARCHAR(66),
    auditor_identity VARCHAR(128) NOT NULL,
    is_reconciled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vault_iot_telemetry (
    telemetry_id BIGSERIAL PRIMARY KEY,
    vault_id VARCHAR(64) NOT NULL REFERENCES commodity_vault_locations(vault_id),
    sensor_type VARCHAR(64) NOT NULL, -- TEMPERATURE, HUMIDITY, VIBRATION, OPTICAL_DOOR
    reading_value NUMERIC(10, 4) NOT NULL,
    unit_of_measure VARCHAR(32) NOT NULL, -- CELSIUS, PERCENT, G_FORCE, STATE_BOOL
    is_alert_triggered BOOLEAN NOT NULL DEFAULT FALSE,
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for high-frequency queries
CREATE INDEX idx_bar_serial_batch ON bar_serial_inventory(batch_id);
CREATE INDEX idx_bar_serial_vault_status ON bar_serial_inventory(vault_id, status);
CREATE INDEX idx_bar_serial_commodity_status ON bar_serial_inventory(commodity_type, status);
CREATE INDEX idx_enwr_records_number ON enwr_records(enwr_number);
CREATE INDEX idx_redemption_orders_user ON redemption_delivery_orders(user_id, status);
CREATE INDEX idx_redemption_orders_status ON redemption_delivery_orders(status);
CREATE INDEX idx_iot_telemetry_vault_time ON vault_iot_telemetry(vault_id, recorded_at DESC);
```

### Kafka Event Schemas

#### Topic: `commodity.enwr.pledged.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "ENWRPledgedEvent",
  "type": "object",
  "required": [
    "event_id",
    "enwr_number",
    "repository",
    "vault_id",
    "commodity_type",
    "net_fine_weight_grams",
    "pledgee_client_id",
    "repository_tx_ref",
    "timestamp_unix_ns"
  ],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "enwr_number": { "type": "string" },
    "repository": { "type": "string", "enum": ["NERL", "CCRL", "MCX_ACCREDITED_VAULT"] },
    "vault_id": { "type": "string" },
    "commodity_type": { "type": "string", "enum": ["GOLD_995", "GOLD_999", "GOLD_9999", "SILVER_999", "COPPER", "ALUMINIUM"] },
    "gross_weight_grams": { "type": "string" },
    "net_fine_weight_grams": { "type": "string" },
    "pledgee_client_id": { "type": "string" },
    "repository_tx_ref": { "type": "string" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `commodity.token_mint_authorized.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CommodityTokenMintAuthorizedEvent",
  "type": "object",
  "required": [
    "authorization_id",
    "deposit_id",
    "enwr_number",
    "commodity_symbol",
    "fine_weight_grams",
    "recipient_wallet_address",
    "signature_eip712",
    "signer_public_key",
    "timestamp_unix_ns"
  ],
  "properties": {
    "authorization_id": { "type": "string", "format": "uuid" },
    "deposit_id": { "type": "string" },
    "enwr_number": { "type": "string" },
    "commodity_symbol": { "type": "string", "enum": ["tGOLD", "tSILVER", "tCOPPER", "tALUMINIUM"] },
    "fine_weight_grams": { "type": "string" },
    "bar_serial_hashes": {
      "type": "array",
      "items": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" }
    },
    "recipient_wallet_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "signature_eip712": { "type": "string", "pattern": "^0x[0-9a-fA-F]{130}$" },
    "signer_public_key": { "type": "string" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `commodity.redemption.dispatched.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CommodityRedemptionDispatchedEvent",
  "type": "object",
  "required": [
    "redemption_id",
    "user_id",
    "wallet_address",
    "commodity_type",
    "dispatched_fine_weight_grams",
    "dispatched_bar_serials",
    "vault_id",
    "gate_pass_number",
    "armored_carrier_code",
    "tracking_number",
    "timestamp_unix_ns"
  ],
  "properties": {
    "redemption_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "wallet_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "commodity_type": { "type": "string" },
    "dispatched_fine_weight_grams": { "type": "string" },
    "dispatched_bar_serials": {
      "type": "array",
      "items": { "type": "string" }
    },
    "vault_id": { "type": "string" },
    "gate_pass_number": { "type": "string" },
    "armored_carrier_code": { "type": "string" },
    "tracking_number": { "type": "string" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `commodity.audit.attestation.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CommodityVaultAuditAttestationEvent",
  "type": "object",
  "required": [
    "snapshot_id",
    "snapshot_timestamp_unix_ns",
    "total_bars_audited",
    "total_fine_weight_grams",
    "sparse_merkle_root_hex",
    "auditor_hsm_signature",
    "vault_breakdown"
  ],
  "properties": {
    "snapshot_id": { "type": "string", "format": "uuid" },
    "snapshot_timestamp_unix_ns": { "type": "integer" },
    "total_bars_audited": { "type": "integer" },
    "total_fine_weight_grams": { "type": "string" },
    "sparse_merkle_root_hex": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "auditor_hsm_signature": { "type": "string" },
    "vault_breakdown": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["vault_id", "commodity_type", "bar_count", "fine_weight_grams"],
        "properties": {
          "vault_id": { "type": "string" },
          "commodity_type": { "type": "string" },
          "bar_count": { "type": "integer" },
          "fine_weight_grams": { "type": "string" }
        }
      }
    }
  }
}
```

## Security & Compliance Notes
- **WDRA & SEBI Regulatory Compliance:** All vaulted commodities backing on-chain tokens must reside exclusively in WDRA-registered warehouses and MCX-approved delivery centers. Physical stock must be electronically registered as eNWR on NERL or CCRL before any token issuance authorization is signed.
- **BIS Hallmarking & LBMA Good Delivery Standards:** Gold bars must strictly adhere to BIS 1417 (Purity & Hallmarking) or LBMA Good Delivery standards with minimum 995.0 or 999.9 fineness. Silver bars must meet minimum 999.0 fineness. Ingested assay certificates lacking valid cryptographic signatures from accredited assayers must be rejected automatically.
- **Dual-Custody Physical Handoff & Armored Transport:** Vault withdrawals require two-man rule validation (Vault Operations Lead + Compliance Officer sign-off). Vault gate passes generate cryptographically salted, time-limited OTP/PIN hashes. Physical transfer from vault to client must be executed via insured armored logistics providers (e.g., Brink's, Loomis, Sequel) with real-time transit telemetry and full transit insurance coverage.
- **Fail-Closed Anti-Double-Pledging Invariant:** Individual bar serial numbers and eNWR receipts are guarded by atomic Redis distributed locks and unique database constraints. A physical bar can never be pledged against multiple tokens, loaned, or unpledged while active token supply exists on Hyperledger Besu.
- **Zero On-Chain PII & Data Privacy:** All recipient names, physical addresses, contact numbers, and KYC details remain strictly in off-chain PostgreSQL encrypted storage (`delivery_address_encrypted` via AES-256-GCM envelope encryption). Blockchain transactions record only pseudonymous wallet addresses, transaction IDs, and cryptographic Merkle proofs.

## Acceptance Criteria
- [ ] Go microservice and Python OCR worker compile and launch with zero configuration errors.
- [ ] NERL and CCRL repository adapters successfully connect via mTLS, submit eNWR pledge requests, and parse confirmation XML/JSON feeds.
- [ ] Python assay parsing engine extracts purity, gross weight, refiner stamp, and BIS hallmark status from PDF reports with 100% field accuracy on valid test fixtures.
- [ ] Bar serial inventory registry prevents duplicate bar entries across refiner codes and enforces $\text{Net Fine Weight} = \text{Gross Weight} \times \frac{\text{Fineness}}{1000}$ with arbitrary-precision decimal math.
- [ ] Dual-control mint authorization pipeline validates eNWR lien marking and produces EIP-712 cryptographic signatures signed by the HSM key.
- [ ] Physical delivery redemption engine correctly executes bar allocation matching, rematerialization instruction dispatch, and gate pass generation.
- [ ] Vault dispatch confirmation emits `commodity.redemption.dispatched.v1` Kafka event and triggers on-chain token burn verification.
- [ ] Armored carrier tracking adapter ingests transit milestones and validates OTP/PIN before marking order as delivered.
- [ ] IoT telemetry monitor ingests vault sensor readings and raises alerts when environmental parameters exceed safety thresholds.
- [ ] Daily Proof-of-Reserve worker calculates 256-bit Sparse Merkle Tree over all vaulted inventory and publishes verified Merkle root to `commodity.audit.attestation.v1`.
- [ ] Integration test suite achieves $\ge 85\%$ code coverage across all repository workflows, OCR parsers, and redemption state transitions.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `104` (Event Schema Standards), Prompt `203` (Wallet & Account Ledger Service), Prompt `213` (Custodian & Depository Integration Service), Prompt `401` (PostgreSQL Schema Design), Prompt `402` (Redis Patterns & Caching).
- **Parallel Tasks:** Prompt `208` (Trade Settlement & DvP Orchestration Service), Prompt `215` (On-Chain vs Off-Chain Reconciliation Engine), Prompt `232` (CBDC Digital Rupee Settlement Adapter).
- **Downstream Blockers:** Prompt `303` (Token Issuance Smart Contract), Prompt `304` (Token Redemption Smart Contract), Prompt `308` (On-Chain Proof-of-Reserve Publishing), Prompt `508` (Flutter Security & Asset Detail Screen).
