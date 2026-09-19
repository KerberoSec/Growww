# 002 - Two-Entity Legal & Technical Separation Architecture

## Purpose
Growww operates under a dual-entity paradigm designed to satisfy strict jurisdictional boundaries:
1. **Entity A (Domestic Indian Regulated Entity):** Operates under SEBI/RBI jurisdiction in Mumbai/Bangalore, holding underlying equity shares in SEBI-registered custodian/depository accounts (NSDL/CDSL), integrating with domestic banking rails (UPI/IMPS/NEFT), and managing domestic investor accounts.
2. **Entity B (International Gateway Entity):** Operates under IFSCA jurisdiction within Gujarat International Finance Tec-City (GIFT City IFSC), managing foreign investor onboarding (FATF/sanctions KYC), foreign currency (USD/EUR/AED) funding, FX conversion, and cross-border investor access.

This document establishes the technical, data, network, and cryptographic boundaries separating both entities to guarantee regulatory compliance, prevent cross-jurisdiction data contamination, and enforce zero cross-border PII leakage.

## What You Are Building
A system boundary specification and network architecture blueprint (`docs/architecture/two_entity_structure.md`) that details:
- Technical and operational separation between Domestic Custody Entity and GIFT City Gateway Entity.
- The secure, audited Inter-Entity Gateway Protocol (IEGP) operating over mutual TLS (mTLS).
- Jurisdictional data residency policies (Domestic data residing exclusively in AWS/GCP `ap-south-1` Mumbai; IFSC data residing in dedicated GIFT City data centers).
- Cross-entity settlement workflows, proof-of-custody requests, and FX conversion pipelines.
- Independent cryptographic certificate authorities (CAs) and Hardware Security Modules (HSMs) per entity.

## Scope Boundaries
- **In Scope:**
 - Entity boundary definition, responsibility matrix, and regulatory jurisdiction mapping.
 - Network isolation topology, VPC peering constraints, and private leased line / Direct Connect configurations.
 - Inter-entity API authentication, authorization, and mutual audit logging standards.
 - Jurisdictional PII containment and pseudonymized identifier mapping.
- **Out of Scope / Handled Elsewhere:**
 - Specific microservice implementations for domestic custody (covered in Prompt 213) and foreign funding (covered in Prompt 214).
 - Detailed inter-entity mTLS network configurations (covered in Prompt 110).
 - Cross-border ledger bridging smart contracts (covered in Prompt 313).

## Technology to Use
- Primary Protocol: gRPC over mTLS with Protocol Buffers v3 for binary, strongly-typed inter-entity communication.
- Cloud / Infrastructure: Dual-VPC architecture across AWS `ap-south-1` (Mumbai) and GIFT City IFSC Tier-IV Data Center / AWS local zone with dedicated AWS Direct Connect.
- Key Custody / PKI: Independent HashiCorp Vault clusters and dedicated FIPS 140-2 Level 3 HSMs per entity with independent Root CAs.
- Justification: Protocol Buffers over mTLS provides high-throughput, low-latency, strictly typed communication with non-repudiable cryptographic identity validation, preventing accidental data leakage across jurisdictional perimeters.

## Backend / Infra Touchpoints
- Domestic Entity Cloud Infrastructure (VPC-Domestic, Mumbai region, PostgreSQL domestic cluster, Vault-Domestic).
- GIFT City Entity Cloud Infrastructure (VPC-IFSC, GIFT City zone, PostgreSQL IFSC cluster, Vault-IFSC).
- AWS Direct Connect / PrivateLink dedicated redundant fiber links with MACsec hardware encryption.
- Multi-region Kafka clusters operating with unidirectional schema-validated mirroring.

## Blockchain Interaction
Establishes the dual-tier permissioned ledger topology across jurisdictional nodes:
- **Domestic Permissioned Ledger:** Hyperledger Besu QBFT network deployed in Mumbai, maintaining the master 1:1 custody-backed `DigitalSecurityToken` contracts and NSDL/CDSL depository mappings.
- **GIFT City Validator Nodes:** Permissioned validator nodes located in GIFT City participating in QBFT consensus, validating cross-border settlement without holding domestic investor PII.
- **On-Chain Zero-PII Invariant:** Cross-border transactions reference only globally unique, cryptographically blind investor account hashes (`bytes32`). No cross-border exchange of PII occurs on-chain or via ledger state.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Map regulatory mandates and legal boundaries between SEBI/RBI (Domestic Entity) and IFSCA (GIFT City Entity).
2. Establish the operational responsibility matrix defining which entity executes depository custody, fiat settlement, FX conversion, and KYC onboarding.
3. Design the network isolation topology diagramming VPC-Domestic and VPC-IFSC with zero public internet exposure.
4. Define the Inter-Entity Gateway Protocol (IEGP) specification using Protocol Buffers v3.
5. Define jurisdictional data residency rules: Domestic PII must never leave `ap-south-1`; International PII stays in GIFT City IFSC.
6. Design the cryptographic trust model establishing two independent Certificate Authorities (CA-Domestic and CA-IFSC) with mutual certificate verification.
7. Detail the Cross-Entity Asset Allocation Workflow: Foreign investor deposits USD -> GIFT City converts to INR -> Domestic entity allocates custody shares -> Domestic ledger issues fractional units to GIFT City omnibus balance.
8. Detail the Cross-Entity Proof-of-Reserve Protocol enabling GIFT City nodes to query real-time depository custody backing without exposing domestic Demat account numbers.
9. Specify the audit logging and non-repudiation mechanism requiring dual-signed ledger receipts for every inter-entity transfer.
10. Design disaster recovery and failover isolation mechanisms ensuring an outage in one entity does not compromise data integrity in the other.
11. Create Mermaid architecture and sequence diagrams for all inter-entity interactions.
12. Review the specification with Chief Compliance Officers, Lead Security Architects, and Legal Counsel across both jurisdictions.

## Interfaces / Contracts
```protobuf
// Inter-Entity Gateway Protocol (docs/architecture/proto/inter_entity_gateway.proto)
syntax = "proto3";

package growww.interentity.v1;

option go_package = "growww/interentity/v1;interentityv1";

service InterEntityGateway {
  // Requests allocation of custody-backed equity units from Domestic Entity
  rpc RequestCustodyAllocation(CustodyAllocationRequest) returns (CustodyAllocationResponse);
  
  // Queries cryptographic Proof-of-Reserve from Domestic Depository Custody
  rpc VerifyProofOfReserve(ReserveQuery) returns (ReserveAttestation);
  
  // Submits settlement batch confirmation across jurisdictional boundaries
  rpc ConfirmBatchSettlement(BatchSettlementRequest) returns (BatchSettlementResponse);
}

message CustodyAllocationRequest {
  string request_id = 1;             // UUID v4 idempotency key
  string isin = 2;                   // Indian Security ISIN (e.g., INE002A01018)
  string fractional_units = 3;       // 18-decimal fixed-point string
  string inr_settlement_amount = 4;  // INR settlement value
  string omnibus_account_id = 5;     // GIFT City IFSC omnibus account hash
  bytes gift_city_signature = 6;     // HSM-backed signature from GIFT City entity
  int64 timestamp = 7;               // Unix timestamp (milliseconds)
}

message CustodyAllocationResponse {
  string request_id = 1;
  string allocation_id = 2;
  string status = 3;                 // ALLOCATED, REJECTED, PENDING_CUSTODY
  string on_chain_tx_hash = 4;       // Besu ledger transaction hash
  bytes domestic_signature = 5;      // HSM-backed signature from Domestic entity
  int64 timestamp = 6;
}

message ReserveQuery {
  string query_id = 1;
  string isin = 2;
  int64 as_of_block_number = 3;
}

message ReserveAttestation {
  string isin = 4;
  string total_shares_in_custody = 5;
  string total_tokens_minted = 6;
  bytes merkle_root = 7;
  bytes custodian_signature = 8;
  int64 timestamp = 9;
}
```

## Security & Compliance Notes
- Strict physical and logical network separation; zero shared administrative credentials or root keys across entities.
- Full compliance with Section 43A of the IT Act, DPDP Act 2023 (India), and IFSCA (Global Administrative and Financial Services) Framework.
- All inter-entity gRPC requests require mutual TLS (TLS 1.3 only, restricted to ECDHE-ECDSA-AES256-GCM-SHA384 cipher suites) and dual HSM payload signing.

## Acceptance Criteria
- [ ] `docs/architecture/two_entity_structure.md` is complete with detailed network diagrams and responsibility matrices.
- [ ] Protobuf service contracts for `InterEntityGateway` are defined and pass proto-linter checks.
- [ ] Strict data residency boundaries (Domestic vs. GIFT City IFSC) are formally specified with zero cross-border PII leakage.
- [ ] Dual-entity cryptographic trust model and independent PKI hierarchies are documented.
- [ ] Sign-off obtained from Legal Counsel (India & GIFT City IFSC) and Chief Information Security Officer.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 001 (Glossary).
- Parallel Tasks: 003 (Regulatory Pathway), 110 (Inter-Entity Communication), 214 (Foreign Investor Funding).
- Downstream Blockers: Blocks Prompt 110, Prompt 213, Prompt 214, and Prompt 313.
