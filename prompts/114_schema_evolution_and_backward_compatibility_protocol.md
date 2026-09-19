# 114 - Schema Evolution & Wire Protocol Backward Compatibility Standard

## Purpose
Establishes the enterprise-wide schema evolution, wire protocol backward compatibility standards, and zero-downtime contract migration protocols across 30+ microservices, Kafka event streams, and gRPC endpoints in the Growww platform. As an institutional-grade financial exchange and fractional equity settlement ecosystem processing billions of rupees in daily volume across domestic Indian equity and GIFT City international corridors, services deploy continuously and independently (canary, blue-green, and rolling updates) without coordinated cluster-wide lockstep downtime.

A breaking change in an asynchronous Kafka event, an internal gRPC service contract, or an on-chain smart contract ABI can lead to catastrophic deserialization faults, dropped financial transactions, order matching halts, or silent ledger corruption (e.g., misinterpreting an order limit price as an execution quantity due to mismatched field tags). This specification mandates strict wire-level and semantic compatibility invariants across Protocol Buffers v3, Apache Avro, JSON Schema, and Ethereum EVM ABI bindings, enforced by Confluent Schema Registry rules, automated CI/CD breaking change gates, and phased dual-read/dual-write deployment patterns.

## What You Are Building
A comprehensive wire protocol compatibility specification (`docs/standards/schema_evolution_and_wire_compatibility.md`), automated linters and breaking change configurations, schema registry governance policies, and zero-downtime migration runbooks:
- **Wire Protocol Governance Standards:** Canonical rules governing Protocol Buffers v3, Apache Avro, and JSON Schema field lifecycles (allocating, modifying, reserving, deprecating, and retiring wire tags).
- **Confluent Schema Registry Policy:** Enterprise compatibility configurations (`BACKWARD_TRANSITIVE` by default, `FULL_TRANSITIVE` for core financial settlement and accounting topics) utilizing the `TopicRecordNameStrategy` to prevent schema drift and subject collisions.
- **Protobuf v3 Breaking Change Linters (`buf breaking`, `buf lint`):** Strict CI/CD configurations validating syntax, field tag uniqueness, reserved tag ranges, package naming, and wire-level binary layout immutability against git baselines.
- **Apache Avro & CDC Migration Guidelines:** Rules for handling Debezium change-data-capture (CDC) schema propagation, default field values, union evolution, and PostgreSQL DDL-to-Avro synchronization without broker ingestion stalls.
- **Dual-Read / Dual-Write Phased Rollout Blueprint:** A standardized 4-phase rollout methodology (Expand -> Dual-Write -> Dual-Read/Tolerate -> Consumer Cutover -> Contract/Deprecate) ensuring zero-downtime migrations for structural payload changes.
- **Smart Contract ABI & Storage Evolution Framework:** OpenZeppelin UUPS (ERC-1967) and ERC-7201 namespaced storage layout validation protecting EVM state across smart contract upgrades.

## Scope Boundaries
- **In Scope:**
  - Wire protocol compatibility rules for Protocol Buffers v3 (gRPC and Kafka events) and Apache Avro (Debezium CDC and analytical data pipelines).
  - Confluent Schema Registry / Karapace subject configuration, compatibility mode transitions, and pre-deployment registration validation.
  - `buf` CLI rulesets for static linting and binary breaking change detection (`FILE`, `WIRE`, `PACKAGE` tiers).
  - Zero-downtime phased rollout blueprints (Expand/Contract pattern) for state-changing wire protocols.
  - Database schema migration coordination with event streaming (Debezium CDC Avro serialization).
  - Permissioned blockchain (Hyperledger Besu) smart contract ABI versioning and UUPS proxy storage layout preservation.
  - Observability metrics for schema serialization faults, deserialization rejections, and deprecated field access.
- **Out of Scope / Handled Elsewhere:**
  - High-level REST API URL paths, HTTP verbs, and query parameter conventions (handled in Prompt 103).
  - Kafka topic taxonomy grammar, partition key hashing algorithms, and CloudEvents envelope definitions (handled in Prompt 104).
  - Monorepo directory structures and polyglot build orchestration (handled in Prompt 106).
  - Broker cluster hardware provisioning, KRaft consensus tuning, and disaster recovery replication (handled in Prompt 403 & 802).
  - Specific domain payload field definitions for individual microservices (handled in Category 2 & 3).

## Technology to Use
- **Serialization Formats:**
  - *Internal RPCs & Real-Time Event Streams:* Protocol Buffers v3 (`proto3`) for microsecond-latency binary serialization and compact memory footprint.
  - *Database Change Data Capture (CDC) & Big Data Sinks:* Apache Avro 1.11+ for self-describing, compact binary streaming with rich schema evolution support.
  - *Public Client Gateways & Admin Webhooks:* JSON Schema (Draft 2020-12 / OpenAPI 3.1) with RFC 7807 problem envelopes.
- **Governance & Validation Tooling:**
  - *Buf CLI (v1.35+):* High-performance Protobuf linter, formatter, and breaking change detection suite (`buf lint`, `buf breaking`).
  - *Confluent Schema Registry 7.6+ / Aiven Karapace:* Central schema authority providing schema ID assignment, subject compatibility verification, and client-side serializer caching.
  - *`protoc-gen-validate` (PGV) / `buf validate`:* Runtime declarative semantic constraint enforcement (e.g., regex, numerical ranges, string lengths).
- **Polyglot Serialization Runtimes:**
  - *Go (v1.22+):* `google.golang.org/protobuf` and `github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf`.
  - *Rust (1.78+):* `prost` / `prost-build` with strict tag retention, `apache-avro` for streaming ingestion.
  - *Python (3.11+):* `protobuf` 5.x, `fastavro`, `confluent-kafka[avro,protobuf]`.
- **Blockchain Contract Upgrades:**
  - Solidity 0.8.24, OpenZeppelin Contracts Upgradeable v5.0 (UUPS / ERC-1967).
  - Foundry (`forge`) and OpenZeppelin Upgrades Plugin for automated storage layout diffing (`forge inspect --storage-layout`).

## Backend / Infra Touchpoints
- **Confluent Schema Registry Cluster:** Highly available, multi-AZ deployment backed by the `_schemas` Kafka compacted topic. Acts as the single source of truth for all Kafka message serializers and deserializers.
- **CI/CD Pipeline Schema Gates:** GitHub Actions / GitLab CI pipelines executing `buf breaking --against git://origin/main` and querying Schema Registry compatibility APIs (`POST /compatibility/subjects/{subject}/versions/latest`) prior to merging pull requests.
- **Envoy API Gateway:** Dynamic gRPC-JSON transcoder preserving unknown fields and validating schema version headers (`X-Schema-Version`, `Accept-Version`).
- **Debezium CDC Engine:** Intercepts PostgreSQL Write-Ahead Logs (WAL) via `pgoutput` and publishes Avro-encoded mutation envelopes to Kafka, ensuring database DDL migrations remain compatible with downstream consumers.
- **Observability Stack (Prometheus & OpenTelemetry):** Scrapes client-side serialization failure counters (`growww_schema_deserialization_failures_total`), unknown field count gauges, and deprecated field access logs.

## Blockchain Interaction
Maintains continuous backward compatibility and prevents state corruption during smart contract lifecycle upgrades on Hyperledger Besu:
- **ERC-1967 / UUPS Proxy Architecture:** All financial state contracts (`DigitalSecurityToken.sol`, `SettlementDvP.sol`, `CashLedgerToken.sol`) deploy behind ERC-1967 transparent or UUPS proxies. Upgrade implementations must NEVER alter or shift historical storage slots.
- **ERC-7201 Namespaced Storage Layouts:** New state variables for upgraded contracts utilize ERC-7201 namespaced storage (`keccak256(abi.encode(uint256(keccak256("growww.storage.SettlementDvP")) - 1)) & ~bytes32(uint256(0xff))`), preventing storage collisions with parent or legacy contracts.
- **Smart Contract ABI Versioning:** Smart contract ABIs are cataloged in versioned directories (`contracts/abi/v1/`, `contracts/abi/v2/`). Backend relayer services implement dual-ABI bindings: when a contract upgrade alters an event signature or function selector, relayers maintain listeners for both `DvPExecutedV1` and `DvPExecutedV2` until transaction in-flight pools drain.
- **Immutable Event Signatures:** Because topic hash signatures (`keccak256("EventName(type1,type2)")`) cannot be mutated retroactively, any modifications to emitted event parameters require publishing a new distinct event signature (e.g., `SettlementFinalizedV2`) alongside the legacy event during deprecation windows.

## Step-by-Step Build Instructions
1. Initialize repository directory `docs/standards/` and author the master wire governance charter: `docs/standards/schema_evolution_and_wire_compatibility.md`.
2. Configure the enterprise `buf.yaml` ruleset enforcing binary and source wire compatibility:
   - Configure linter rules under `lint.use`: `DEFAULT`, `COMMENTS`, `FILE_LOWER_SNAKE_CASE`, `PACKAGE_VERSION_SUFFIX`.
   - Configure breaking change rules under `breaking.use`: `WIRE`, `FILE`.
   - Configure breaking change comparison target to track the canonical production branch (`against: "git://github.com/growww/growww-proto.git#branch=main"`).
3. Establish the **Protobuf v3 Evolution Standard**:
   - *Rule 1 (Field Numbers):* Field numbers (tags) are permanently immutable once assigned. They can never be changed, reassigned, or repurposed.
   - *Rule 2 (Field Deletion & `reserved`):* Fields cannot be deleted without immediately adding their field number and name to a `reserved` statement (e.g., `reserved 4, 11; reserved "legacy_broker_id";`).
   - *Rule 3 (Types):* Field wire types cannot be changed arbitrarily. Conversions between wire-incompatible types (e.g., `string` to `bytes`, `int32` to `fixed32`, `int32` to `string`) are strictly prohibited.
   - *Rule 4 (Field Cardinality):* Adding or removing `repeated` is prohibited; changing a scalar into a `repeated` breaks wire binary layout.
   - *Rule 5 (Unknown Fields):* All service deserializers across Go, Rust, and Python must preserve unknown fields during decode/encode roundtrips (Proto3 unknown field preservation) to allow newer producers to communicate through older intermediary proxies.
   - *Rule 6 (Enums):* Every enum must define tag `0` as `<ENUM_NAME>_UNSPECIFIED`. Removed enum members must be marked `reserved`.
4. Configure **Confluent Schema Registry Compatibility Policies**:
   - Set global default compatibility mode to `BACKWARD_TRANSITIVE`.
   - For critical financial accounting topics (`domestic.ledger.*`, `giftcity.settlement.*`), enforce `FULL_TRANSITIVE` compatibility (ensuring all old messages can be read by new schemas, and new messages can be read by old schemas).
   - Set Subject Name Strategy to `io.confluent.kafka.serializers.subject.TopicRecordNameStrategy` to decouple record schemas from individual topic names and permit polymorphic message schemas within audit topics without collisions.
5. Author the **Apache Avro Evolution Rules**:
   - New fields added to Avro records MUST specify a `default` value (e.g., `"default": null` with a union type `["null", "string"]`).
   - Fields can never be deleted from an Avro schema unless all active consumer versions have deployed code that no longer expects that field.
   - Modifying union types must only append new allowed types to the end of the union array.
6. Design the **4-Phase Dual-Read / Dual-Write Migration Blueprint**:
   - *Phase 1 (Expand - Dual Write):* Producer publishes both the legacy field/topic and the new field/topic simultaneously. New field is optional.
   - *Phase 2 (Tolerate - Dual Read):* Consumers are updated to read the new field if present, falling back to the legacy field if absent.
   - *Phase 3 (Cutover - Exclusive Read):* Once 100% of consumers are on Phase 2, producers cease writing to the legacy field and write exclusively to the new field.
   - *Phase 4 (Contract - Deprecate):* Legacy fields are marked `deprecated = true` in Protobuf, observed for zero access over a 90-day grace period, and subsequently moved to `reserved`.
7. Configure **Debezium CDC Schema Synchronization**:
   - Configure Debezium PostgreSQL connector with Avro serialization and Schema Registry integration.
   - Mandate that all database schema DDL migrations (Alembic/Flyway/Liquibase) follow additive-only semantics: column additions must be `NULLABLE` or have a database `DEFAULT` value before application code updates.
8. Establish the **Smart Contract ABI & Storage Layout Upgrade Verification Process**:
   - Integrate `forge inspect <ContractName> storage-layout` into the smart contract CI pipeline.
   - Enforce automated comparison between implementation upgrades to verify zero slot shifting for existing variables.
   - Verify that all upgrades inherit from `Initializable` and `UUPSUpgradeable`, with empty constructor blocks containing `_disableInitializers()`.
9. Implement CI/CD Schema Validation Gates:
   - Build a GitHub Actions reusable workflow (`.github/workflows/schema-compat-check.yml`).
   - Step 1: Run `buf lint`.
   - Step 2: Run `buf breaking --against 'https://github.com/growww/growww-proto.git#branch=main'`.
   - Step 3: Run Python/Go test suites verifying that Avro and JSON schemas registered in `schemas/` pass Schema Registry compatibility against live staging registries via `POST /compatibility/subjects/{subject}/versions/latest`.
10. Implement runtime telemetry and observability for schema operations:
    - Standardize Prometheus metrics: `growww_schema_registry_request_latency_seconds`, `growww_wire_deserialization_errors_total`, `growww_deprecated_field_access_total`.
    - Emit warning log events whenever a service deserializes a payload using a field marked with `[deprecated = true]`.
11. Design the `schema_evolution_registry` database metadata table to track version rollout states across all production microservices and active Kafka topics.
12. Construct end-to-end chaos integration tests simulating mixed-version cluster environments:
    - Deploy Producer v2 (with new fields) while Consumer v1 (legacy code) is actively consuming from the topic; assert zero deserialization errors and successful message processing.
    - Deploy Consumer v2 (expecting new fields) while Producer v1 (legacy code) is actively emitting; assert graceful fallback to defaults with zero panics.
13. Finalize documentation and publish `docs/standards/schema_evolution_and_wire_compatibility.md`.

## Interfaces / Contracts

### 1. Protobuf Breaking Change Configuration (`buf.yaml`)
```yaml
version: v1
name: buf.build/growww/financial-platform
lint:
  use:
    - DEFAULT
    - COMMENTS
    - FILE_LOWER_SNAKE_CASE
    - PACKAGE_VERSION_SUFFIX
  except:
    - PACKAGE_NO_IMPORT_CYCLE
breaking:
  use:
    - WIRE
    - FILE
    - PACKAGE
  ignore:
    - proto/internal/
```

### 2. Canonical Protobuf Schema Evolution Specification (`proto/trading/v1/order_event.proto`)
```protobuf
syntax = "proto3";

package growww.trading.v1;

option go_package = "github.com/growww/proto/gen/go/trading/v1;tradingv1";
option java_package = "com.growww.proto.trading.v1";

import "google/protobuf/timestamp.proto";
import "proto/common/v1/money.proto";

// OrderPlacedEvent represents an order successfully matched and accepted.
// Wire Protocol Invariant: Tags 1 through 10 are frozen in v1.
message OrderPlacedEvent {
  // RESERVED: Tags and field names of retired fields to prevent accidental reassignment.
  reserved 4, 11 to 15;
  reserved "broker_sub_account_code", "legacy_terminal_id";

  // Tag 1: Permanent Order Identifier (UUIDv7 string)
  string order_id = 1;

  // Tag 2: Authenticated Investor Identifier
  string user_id = 2;

  // Tag 3: Security ISIN identifier (e.g., "INE002A01018")
  string security_isin = 3;

  // Tag 5: Total order quantity (Fixed-point 6 decimals as string or integer units)
  int64 quantity_micros = 5;

  // Tag 6: Limit price representation (safe money object)
  growww.common.v1.Money limit_price = 6;

  // Tag 7: Order Type enumeration
  OrderType order_type = 7;

  // Tag 8: Time In Force enumeration
  TimeInForce time_in_force = 8;

  // Tag 9: Timestamp of acceptance
  google.protobuf.Timestamp placed_at = 9;

  // Tag 10: DEPRECATED field. Use execution_venue instead.
  // Marked deprecated; consumers must transition before Phase 4 removal.
  string routing_desk_id = 10 [deprecated = true];

  // Tag 16: ADDED in v1.1.0 - Zero-downtime backward-compatible expansion.
  // Wire compatible: scalar string defaults to empty string in older consumers.
  string execution_venue = 16;

  // Tag 17: ADDED in v1.2.0 - Optional institutional algorithmic routing parameters.
  optional AlgorithmicRoutingParams algo_params = 17;
}

enum OrderType {
  ORDER_TYPE_UNSPECIFIED = 0;
  ORDER_TYPE_LIMIT = 1;
  ORDER_TYPE_MARKET = 2;
  ORDER_TYPE_STOP_LIMIT = 3;
}

enum TimeInForce {
  TIME_IN_FORCE_UNSPECIFIED = 0;
  TIME_IN_FORCE_DAY = 1;
  TIME_IN_FORCE_IOC = 2; // Immediate or Cancel
  TIME_IN_FORCE_GTC = 3; // Good 'Til Cancelled
}

message AlgorithmicRoutingParams {
  string algorithm_name = 1;
  int32 max_participation_rate_bps = 2;
}
```

### 3. Confluent Schema Registry Subject Compatibility Configuration
```json
{
  "compatibility": "BACKWARD_TRANSITIVE"
}
```

### 4. Schema Rollout Tracking PostgreSQL Table
```sql
CREATE TABLE schema_evolution_registry (
    subject_name VARCHAR(255) PRIMARY KEY,
    schema_type VARCHAR(32) NOT NULL, -- 'PROTOBUF', 'AVRO', 'JSON'
    current_version INT NOT NULL,
    compatibility_mode VARCHAR(32) NOT NULL DEFAULT 'BACKWARD_TRANSITIVE',
    rollout_phase VARCHAR(32) NOT NULL DEFAULT 'PHASE_1_EXPAND',
    legacy_fields_deprecated TEXT[] DEFAULT '{}',
    reserved_tags_documented INT[] DEFAULT '{}',
    ci_validation_hash VARCHAR(64) NOT NULL,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deprecated_at TIMESTAMPTZ,
    CONSTRAINT check_rollout_phase CHECK (
        rollout_phase IN ('PHASE_1_EXPAND', 'PHASE_2_DUAL_READ', 'PHASE_3_CUTOVER', 'PHASE_4_CONTRACT')
    )
);

CREATE INDEX idx_schema_rollout ON schema_evolution_registry (rollout_phase);
```

### 5. Smart Contract ERC-7201 Namespaced Storage Upgrade Layout (Solidity)
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";

contract SettlementDvPUpgradeable is Initializable, UUPSUpgradeable, AccessControlUpgradeable {
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    // ERC-7201: Storage namespace prevents storage slot collisions across upgrades
    // keccak256(abi.encode(uint256(keccak256("growww.storage.SettlementDvP")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant SETTLEMENT_STORAGE_LOCATION =
        0x5a1811a2f643194a50d2e8b23f2b406e987c88b7f73967b5eb6ce0144f808f00;

    struct SettlementStorage {
        mapping(bytes32 => bool) executedSettlements;
        uint256 totalSettledVolumePaise;
        // APPEND-ONLY: New variables are appended to the end of the struct.
        // Existing variables can NEVER be reordered or have their types altered.
        mapping(bytes32 => uint256) settlementTimestamps; // Added in Upgrade V2
    }

    function _getSettlementStorage() private pure returns (SettlementStorage storage $) {
        assembly {
            $.slot := SETTLEMENT_STORAGE_LOCATION
        }
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address defaultAdmin) external initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        _grantRole(DEFAULT_ADMIN_ROLE, defaultAdmin);
        _grantRole(UPGRADER_ROLE, defaultAdmin);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {
        // Upgrade authorization gated by multi-sig governance
    }
}
```

## Security & Compliance Notes
- **Prevention of Silent Field Collision & Data Corruption:** In Protobuf v3, reusing a previously deleted field number causes newer consumers to interpret old payload bytes as the new data type, potentially leading to critical mathematical errors in financial calculations (e.g., trading order pricing vs. quantity). Reserving all retired field tags and names is mandatory and enforced by `buf breaking`.
- **Financial Audit Trail Replayability:** Under SEBI Master Circulars and IFSCA capital market regulations, electronic trading records and clearing ledger entries must be reconstructible for up to 10 years. Serialized Kafka events stored in cold object stores (S3/GCS) must remain decodable indefinitely; this requires that schemas are never deleted from the Schema Registry and all mutations maintain `BACKWARD_TRANSITIVE` compatibility.
- **Access Control & Schema Registry Authentication:** Only authorized CI/CD deployment service principals may register or alter schema versions on Confluent Schema Registry. Registry write access is locked with mTLS and RBAC, preventing rogue microservice instances or unauthorized developers from overriding production schemas.
- **PII Protection in Schema Evolutions:** Under the Digital Personal Data Protection (DPDP) Act 2023, any newly introduced fields carrying Personally Identifiable Information (PAN, Aadhaar hash, bank account details, investor phone numbers) must be decorated with custom Protobuf field options `[(growww.common.v1.pii) = true]`. CI pipeline checks verify that PII fields automatically trigger masking filters in log sinks and analytics connectors.

## Acceptance Criteria
- [ ] Master wire governance charter published at `docs/standards/schema_evolution_and_wire_compatibility.md`.
- [ ] Enterprise `buf.yaml` configured with `WIRE` and `FILE` breaking change enforcement enabled.
- [ ] Confluent Schema Registry configured with global `BACKWARD_TRANSITIVE` default and `FULL_TRANSITIVE` for core financial settlement topics.
- [ ] Confluent Schema Registry configured to use `TopicRecordNameStrategy` for polymorphic event topics.
- [ ] CI/CD breaking change pipeline step implemented and tested, successfully blocking PRs that modify assigned field tags, change wire types, or delete fields without reservation.
- [ ] Complete 4-phase dual-read / dual-write migration runbook documented with concrete Go and Python code examples.
- [ ] Debezium CDC Avro evolution rules established, verifying that database DDL migrations remain backward-compatible with active consumers.
- [ ] Foundry storage layout test harness configured to detect slot shifting in UUPS upgradeable smart contracts (`forge inspect --storage-layout`).
- [ ] Prometheus metrics for deserialization failures and deprecated field access integrated into standard service templates.
- [ ] Mixed-version integration tests pass: Producer v2 with Consumer v1, and Producer v1 with Consumer v2 exhibit zero unhandled exceptions or data loss.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture Overview), Prompt 103 (API Design Standards), Prompt 104 (Event Schema & Kafka Topic Standards), Prompt 107 (Coding Standards & Linting).
- **Parallel Work:** Prompt 106 (Monorepo Layout & Tooling), Prompt 110 (Inter-Entity Secure Communication), Prompt 112 (Idempotency & Exactly-Once Processing).
- **Blocks:** Prompt 203 (Wallet Account Service), Prompt 204 (Order Service), Prompt 205 (Order Matching Engine), Prompt 208 (Trade Settlement Service), Prompt 306 (DvP Settlement Smart Contract), Prompt 803 (CI/CD Pipeline Architecture).
