# 110 - Inter-Entity Secure Communication Design (Domestic Regulated Entity <-> GIFT City Gateway)

## Purpose
Establishes the secure, cryptographically attested, and legally compliant communication protocol between Growww's two foundational legal and technical entities:
1. **Domestic Regulated Entity (Entity A):** SEBI/RBI registered broker-dealer and custody interface operating within mainland India, holding real Indian equities in demat custody at NSDL/CDSL.
2. **International Gateway Entity (Entity B):** GIFT City / IFSCA registered entity operating in the International Financial Services Centre, responsible for foreign investor onboarding, multi-currency funding, and international regulatory reporting.

This prompt provides systems architects and security engineers with the canonical specifications for the bilateral communication channel, ensuring strict cross-border regulatory compliance (SEBI, RBI Liberalised Remittance Scheme / FEMA, and IFSCA guidelines), data privacy ring-fencing (zero cross-border domestic PII leakage under DPDP Act / GDPR), mutual TLS tunnels with dedicated CAs, asymmetric Ed25519 request signing, and cryptographic nonces.

## What You Are Building
A comprehensive inter-entity architecture specification (`docs/architecture/inter_entity_communication.md`), gRPC service definitions (`proto/inter_entity/v1/`), cryptographic signature verification middleware, and dual-entity ledger reconciliation protocols:
- Secure Network Architecture: Encrypted WireGuard / AWS PrivateLink / IPSec site-to-site VPN tunnel between Domestic and GIFT City cloud environments.
- Bilateral Mutual TLS (mTLS): Dedicated private Certificate Authority (CA) with certificate rotation protocols and pinned root certificates.
- Asymmetric Request Signing & Non-Repudiation: Every inter-entity HTTP/gRPC request includes cryptographic signature headers (`X-Growww-Signature`, `X-Growww-Nonce`, `X-Growww-Timestamp`) signed using Ed25519 private keys stored in dedicated HSMs.
- Dual-Entity Reconciliation Protocol: Continuous automated reconciliation between foreign investor unit allocations and domestic physical demat custody balances.

## Scope Boundaries
- **In Scope:**
 - Network topology, VPN/PrivateLink routing, and firewall rules between Entity A and Entity B.
 - Bilateral mTLS configuration and Certificate Authority (CA) lifecycle.
 - Inter-entity gRPC service definitions and protobuf message contracts.
 - Cryptographic payload signing, nonce verification, and replay attack prevention.
 - Cross-border data privacy rules and PII sanitization filters.
- **Out of Scope / Handled Elsewhere:**
 - Domestic-only internal microservice communication (handled in Prompt 103 & 105).
 - Foreign investor onboarding and FX currency conversion logic (handled in Prompt 214).
 - Cross-entity blockchain ledger bridge and event indexing (handled in Prompt 313).

## Technology to Use
- **Network & Transport:**
 - *Transport Protocol:* gRPC over HTTP/2 with mTLS (TLS 1.3 only, enforcing modern cipher suites: `TLS_AES_256_GCM_SHA384` and `TLS_CHACHA20_POLY1305_SHA256`).
 - *Network Tunnel:* AWS DirectConnect / PrivateLink or IPSec IKEv2 site-to-site VPN with WireGuard fallback.
- **Cryptographic Algorithms:**
 - *Asymmetric Signing:* Ed25519 for request-level message authentication and non-repudiation.
 - *Hashing & Nonce:* SHA-256 for request payload digest; UUIDv7 for monotonically increasing nonces.
- **Event Streaming & Replication:** Apache Kafka MirrorMaker 2 with selective topic replication and message sanitization filters.

## Backend / Infra Touchpoints
- **Domestic Edge Gateway (Entity A):** Dedicated Envoy gateway instance exposed only to the inter-entity VPN tunnel.
- **GIFT City Edge Gateway (Entity B):** Dedicated Envoy gateway instance terminating mTLS and routing to internal GIFT City services.
- **Dual PostgreSQL Audit Stores:** Entity A and Entity B each maintain an independent, append-only PostgreSQL database logging all inbound and outbound inter-entity transmissions.

## Blockchain Interaction
Establishes cross-entity ledger transparency while protecting domestic investor privacy:
- **Shared Consortium View:** Both Domestic Entity A and GIFT City Entity B run validator / observer nodes on the permissioned Hyperledger Besu consortium network.
- **On-Chain Proof-of-Reserve Attestation:** Domestic Entity A publishes cryptographic Merkle roots of physical NSDL/CDSL custody positions to `ProofOfReserveRegistry.sol`. GIFT City Entity B validates these on-chain attestations before allowing foreign investors to purchase tokenized fractional units.
- **Zero Domestic PII on Cross-Entity Ledger:** On-chain tokens allocated to GIFT City investors reflect only foreign investor wallet addresses; no domestic investor identities or domestic bank details are ever transmitted across entity boundaries or written to shared ledger states.

## Step-by-Step Build Instructions
1. Author `docs/architecture/inter_entity_communication.md` establishing the legal boundaries, data sovereignty rules, and network architecture.
2. Design the dedicated Inter-Entity Private Certificate Authority (CA) using Vault PKI:
 - Root CA (`growww-inter-entity-root-ca`) with 10-year validity.
 - Intermediate CAs for Entity A (`entity-a-ca`) and Entity B (`entity-b-ca`) issuing 90-day certificates.
3. Define the network interconnect: Setup AWS PrivateLink / IPSec VPN between VPC-A (Domestic, Mumbai region) and VPC-B (GIFT City, Gujarat / Singapore international cloud region).
4. Author the foundational inter-entity Protocol Buffer contract `proto/inter_entity/v1/gateway.proto` defining RPCs for:
 - `CheckCustodyAvailability`: Verifies physical share allocation before foreign order acceptance.
 - `ReserveCustodyShares`: Locks physical shares in demat pool for foreign purchase.
 - `ExecuteCrossEntitySettlement`: Signals DvP completion and custody balance transfer.
 - `PublishProofOfReserve`: Shares audited custody Merkle root.
5. Define the cryptographic request-signing algorithm:
 - Construct canonical signature string: `METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + SHA256(BODY)`.
 - Sign canonical string with Entity Ed25519 private key.
 - Inject signature into HTTP/gRPC metadata headers.
6. Implement bilateral signature verification interceptors in Go and Rust:
 - Check timestamp freshness (reject requests where `|now - timestamp| > 30 seconds`).
 - Check nonce uniqueness in Redis (reject duplicate nonces within a 5-minute sliding window).
 - Verify Ed25519 signature against the sender's known public key certificate.
7. Configure Kafka MirrorMaker 2 for asynchronous cross-entity event streaming:
 - Define strict topic whitelist (`giftcity.export.*`, `domestic.export.*`).
 - Implement custom Kafka Connect SMT (Single Message Transform) to scrub all domestic PII fields before cross-border transmission.
8. Implement dual-entity audit logging: Every transmitted message is hashed (SHA-256) and recorded in both Entity A and Entity B audit tables with receipt timestamps.
9. Establish an automated daily custody reconciliation job: Compares total fractional units held by GIFT City foreign investors against the omnibus custody holding account at NSDL/CDSL.
10. Formulate cross-entity incident response protocols for network partitions, certificate revocations, and key compromise.
11. Perform compliance review against RBI FEMA guidelines, SEBI cross-border investment regulations, and IFSCA Capital Markets regulations.
12. Publish `docs/architecture/inter_entity_communication.md` to repository.

## Interfaces / Contracts

### Inter-Entity Gateway Protocol Buffer (`proto/inter_entity/v1/gateway.proto`)
```protobuf
syntax = "proto3";

package growww.inter_entity.v1;

import "google/protobuf/timestamp.proto";
import "proto/common/v1/fractional_share.proto";
import "proto/common/v1/money.proto";

option go_package = "github.com/growww/proto/gen/go/inter_entity/v1;interentityv1";

service InterEntityGatewayService {
  // Check available physical demat custody for an Indian equity
  rpc CheckCustodyAvailability(CustodyCheckRequest) returns (CustodyCheckResponse);

  // Reserve custody units for an incoming international purchase order
  rpc ReserveCustody(ReserveCustodyRequest) returns (ReserveCustodyResponse);

  // Confirm settlement and release/transfer custody units
  rpc ConfirmSettlement(ConfirmSettlementRequest) returns (ConfirmSettlementResponse);

  // Synchronize proof-of-reserve Merkle tree attestation
  rpc SyncProofOfReserve(ProofOfReserveSyncRequest) returns (ProofOfReserveSyncResponse);
}

message CustodyCheckRequest {
  string isin = 1; // e.g., "INE002A01018" (Reliance Industries)
  string request_id = 2;
  google.protobuf.Timestamp timestamp = 3;
}

message CustodyCheckResponse {
  string isin = 1;
  growww.common.v1.FractionalShare available_shares = 2;
  growww.common.v1.FractionalShare total_custody_shares = 3;
  bool is_trading_halted = 4;
}

message ReserveCustodyRequest {
  string reservation_id = 1; // UUIDv7 generated by GIFT City gateway
  string isin = 2;
  growww.common.v1.FractionalShare quantity = 3;
  growww.common.v1.Money max_fiat_inr_value = 4;
  int64 expiry_timestamp_ms = 5;
}

message ReserveCustodyResponse {
  string reservation_id = 1;
  bool status = 2; // TRUE = RESERVED, FALSE = REJECTED
  string custody_lock_reference = 3;
  string error_reason = 4;
}

message ConfirmSettlementRequest {
  string reservation_id = 1;
  string settlement_id = 2;
  string onchain_tx_hash = 3;
  google.protobuf.Timestamp settled_at = 4;
}

message ConfirmSettlementResponse {
  string settlement_id = 1;
  bool confirmed = 2;
  string custody_demat_voucher_id = 3;
}

message ProofOfReserveSyncRequest {
  string isin = 1;
  bytes merkle_root = 2;
  int64 block_height = 3;
  google.protobuf.Timestamp timestamp = 4;
}

message ProofOfReserveSyncResponse {
  bool verified = 1;
  string attestation_id = 2;
}
```

### Inter-Entity Cryptographic Request Headers
```http
POST /growww.inter_entity.v1.InterEntityGatewayService/ReserveCustody HTTP/2
Host: gateway.domestic.growww.internal
Content-Type: application/grpc
X-Growww-Entity-Source: GIFT_CITY
X-Growww-Entity-Target: DOMESTIC
X-Growww-Request-ID: req_01HZX89AB72K9M12P5QRSTUVWX
X-Growww-Timestamp: 1718811900
X-Growww-Nonce: non_01HZX89AB98CDEF1234567890A
X-Growww-Signature: base64(ed25519_sign(entity_b_private_key, canonical_string))
```

## Security & Compliance Notes
- **Data Sovereignty & DPDP Compliance:** Domestic investor personal identifiable information (PII) is strictly prohibited from crossing the inter-entity boundary into GIFT City or overseas cloud regions.
- **Non-Repudiation & Audit Trail:** All inter-entity messages are cryptographically signed with Ed25519 keys held in HSMs; neither entity can repudiate a custody reservation or settlement confirmation once transmitted.
- **Replay Attack Defense:** Strict 30-second timestamp freshness validation combined with Redis-backed nonce deduplication guarantees protection against message replay attacks.
- **FEMA & LRS Compliance:** All cross-entity financial flows are reconciled against RBI Liberalised Remittance Scheme (LRS) and FEMA reporting parameters on a T+0 basis.

## Acceptance Criteria
- [ ] Complete Inter-Entity Communication architecture document (`docs/architecture/inter_entity_communication.md`) published.
- [ ] Inter-entity gRPC service contract (`proto/inter_entity/v1/gateway.proto`) authored and verified with Buf.
- [ ] Bilateral mTLS setup and private CA certificate lifecycle documented.
- [ ] Ed25519 request signature and timestamp/nonce verification algorithm specified with unit test test-vectors.
- [ ] Kafka MirrorMaker 2 PII-sanitizing replication topology finalized.
- [ ] Automated daily custody reconciliation protocol defined between domestic demat pool and foreign holdings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 002 (Two-Entity Legal Structure), Prompt 101 (System Architecture), Prompt 103 (API Standards), Prompt 105 (Auth Architecture).
- **Parallel Work:** Prompt 109 (Secrets Management), Prompt 111 (Domain Model).
- **Blocks:** Prompt 214 (Foreign-Investor Funding & FX Service), Prompt 313 (Cross-Entity Ledger Bridge), Prompt 701 (Threat Model).
