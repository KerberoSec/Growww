# 201 - User & Identity Service (FastAPI / PostgreSQL)

## Purpose
The User & Identity Service is the primary gatekeeper and account management authority for the Growww platform. It manages investor identity lifecycle states, authentication credentials (MPIN, biometric FIDO2/WebAuthn public keys), device session bindings, and profile metadata. In alignment with Growww's north-star mission of providing a fully compliant, asset-backed fractional equity infrastructure, this service guarantees that every investor account is strictly bound to a single legal identity, verified hardware device, and cryptographic public address on the permissioned ledger.

By maintaining strict separation between authentication credentials, profile data, and financial ledger accounts, this microservice ensures high developer velocity and modularity while upholding the zero-trust security architecture mandated by SEBI/RBI fintech standards.

## What You Are Building
A production-grade, asynchronous Python FastAPI microservice (`services/user-service`) backed by PostgreSQL and Redis. Concrete deliverables include:
- RESTful HTTP/JSON endpoints for mobile and web client authentication, registration, session renewal, and profile queries.
- High-performance gRPC server implementation (`UserServiceServer`) for internal inter-service authentication token validation and user identity resolution.
- Device fingerprinting and cryptographic session management engine using Redis.
- Asynchronous Kafka event publisher for user lifecycle state changes (`user.registered.v1`, `user.status_changed.v1`, `user.device_bound.v1`).
- Alembic database migration scripts and SQLAlchemy 2.0 async domain models.

## Scope Boundaries
- **In Scope:**
 - Investor registration flow (phone number, email verification via OTP).
 - 6-digit MPIN hashing and verification using Argon2id.
 - WebAuthn/FIDO2 biometric public key enrollment and signature validation.
 - Device binding, hardware ID validation, and refresh token rotation.
 - User status state machine (`PENDING_KYC`, `KYC_SUBMITTED`, `ACTIVE`, `SUSPENDED`, `CLOSED`).
 - Associating a deterministic permissioned ledger address (Hyperledger Besu) with each verified user.
- **Out of Scope / Handled Elsewhere:**
 - KYC document extraction, OCR, and Aadhaar/PAN verification (Prompt 202).
 - Cash wallet balance and INR double-entry ledger (Prompt 203).
 - Pre-trade risk profiling and margin checks (Prompt 206).
 - Admin manual override and compliance freeze actions (Prompt 217).

## Technology to Use
- **Primary Language & Framework:** Python 3.12 with FastAPI (0.111+). FastAPI is chosen for its native asynchronous I/O performance, automated OpenAPI 3.1 documentation generation, strict data validation with Pydantic v2, and rapid integration with cryptographic authentication libraries.
- **Database & Storage:** PostgreSQL 16+ using `asyncpg` and SQLAlchemy 2.0 (async ORM); Redis 7.2+ for distributed session cache and rate limiting.
- **Authentication & Cryptography:** `argon2-cffi` for MPIN password hashing; `py_webauthn` for FIDO2/biometric authentication; `PyJWT` with Ed25519 (EdDSA) asymmetric keys for access/refresh token signing.
- **Polyglot & Transport Integration:** `grpcio` and `grpcio-tools` for internal gRPC service definitions; `aiokafka` for event streaming to Kafka; mTLS for inter-service communication.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Stores transactional records across tables `users`, `user_profiles`, `user_devices`, `auth_sessions`, and `user_ledger_mappings`.
- **Redis 7.2:** Manages active JWT refresh token blacklists, OTP verification tokens (5-minute TTL), and login failure counters.
- **Apache Kafka:** Publishes to topic `user.events.v1` with key partitioned by `user_id`.
- **Notification Service (Prompt 211):** Invoked via gRPC for OTP delivery during registration and login challenges.
- **Audit Log Service (Prompt 218):** Ingests cryptographic audit trails for all authentication and status change events.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Zero On-Chain PII Invariant:** In strict adherence to India's DPDP Act 2023 and global privacy regulations, NO personally identifiable information (names, emails, PANs, phone numbers) is ever written to the blockchain.
- **Address Generation & Association:** For each registered user, the service derives a dedicated cryptographic address (Ethereum-compatible secp256k1 public key derived from internal KMS or user device) that acts as their pseudonymized account identifier on the permissioned Hyperledger Besu network.
- **Compliance Registry Link:** When a user transitions to `ACTIVE` post-KYC, the service triggers the compliance relayer to register the address on `ComplianceRegistry.sol` on Hyperledger Besu, ensuring all subsequent tokenized equity purchases (1:1 backed by NSDL/CDSL custody) map directly to an authorized address.
- **Consensus & Finality:** All ledger transactions operate under Istanbul/QBFT Byzantine Fault Tolerant consensus with deterministic 2-second block finality.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize directory `services/user-service` with Poetry/uv, Python 3.12, strict linters (`ruff`, `mypy`), and directory layout adhering to Prompt 106.
2. **Define Protobuf Schemas:** Create `proto/growww/user/v1/user_service.proto` defining `UserService` RPCs (`GetUser`, `ValidateToken`, `GetUserLedgerAddress`).
3. **Generate gRPC Stubs:** Set up build step to compile `.proto` files into Python async gRPC stubs.
4. **Configure Database & Migrations:** Set up SQLAlchemy 2.0 async engine and Alembic migration directory. Create initial schema migration for `users`, `user_profiles`, `user_devices`, and `auth_sessions`.
5. **Implement Argon2id & Cryptography Module:** Create utility class for MPIN hashing using Argon2id ($m=65536, t=3, p=4$) and Ed25519 JWT asymmetric key signing.
6. **Implement WebAuthn / Biometric Flow:** Build registration and assertion endpoints for FIDO2 biometric credentials using `py_webauthn`.
7. **Build Registration & OTP State Machine:** Create REST endpoints for mobile number initiation, OTP validation, and initial account creation with Redis TTL safeguards.
8. **Build Session & Token Manager:** Implement access token generation (15-minute lifetime) and refresh token rotation with cryptographic revocation in Redis.
9. **Implement Device Binding Middleware:** Enforce hardware fingerprint validation on all sensitive authenticated endpoints; challenge unknown devices with 2FA OTP.
10. **Implement User Lifecycle State Machine:** Enforce strict state transitions (`PENDING_KYC` $\rightarrow$ `KYC_SUBMITTED` $\rightarrow$ `ACTIVE` / `SUSPENDED` / `CLOSED`) with database row-level locking.
11. **Implement Kafka Event Publisher:** Implement transactional outbox pattern or `aiokafka` publisher to stream `user.events.v1` upon state transitions.
12. **Implement gRPC Service Handlers:** Implement `UserServiceServicer` providing ultra-fast (<2ms) session validation and profile lookups for downstream services.
13. **Configure Health Checks & Telemetry:** Expose `/healthz`, `/livez`, `/metrics` (Prometheus), and OpenTelemetry tracing middleware.
14. **Write Comprehensive Test Suite:** Implement unit and integration tests using `pytest-asyncio` and `testcontainers-python` achieving $\ge 85\%$ code coverage.

## Interfaces / Contracts

### Protobuf Definition (`user_service.proto`)
```protobuf
syntax = "proto3";

package growww.user.v1;

option go_package = "growww/user/v1;userv1";

service UserService {
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  rpc ValidateSession (ValidateSessionRequest) returns (ValidateSessionResponse);
  rpc GetUserLedgerAddress (GetUserLedgerAddressRequest) returns (GetUserLedgerAddressResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  string user_id = 1;
  string phone_number_masked = 2;
  string email_masked = 3;
  string status = 4; // PENDING_KYC, ACTIVE, SUSPENDED, CLOSED
  string jurisdiction = 5; // DOMESTIC_RESIDENT, GIFT_CITY_NRI
  int64 created_at_unix = 6;
}

message ValidateSessionRequest {
  string access_token = 1;
  string device_fingerprint = 2;
}

message ValidateSessionResponse {
  bool is_valid = 1;
  string user_id = 2;
  string status = 3;
  string ledger_address = 4;
}

message GetUserLedgerAddressRequest {
  string user_id = 1;
}

message GetUserLedgerAddressResponse {
  string user_id = 1;
  string ledger_address = 2; // Hex-encoded 0x address for Hyperledger Besu
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE user_status_enum AS ENUM ('PENDING_KYC', 'KYC_SUBMITTED', 'ACTIVE', 'SUSPENDED', 'CLOSED');
CREATE TYPE user_jurisdiction_enum AS ENUM ('DOMESTIC_RESIDENT', 'GIFT_CITY_NRI');

CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number_e164 VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    mpin_hash VARCHAR(255) NOT NULL,
    status user_status_enum NOT NULL DEFAULT 'PENDING_KYC',
    jurisdiction user_jurisdiction_enum NOT NULL DEFAULT 'DOMESTIC_RESIDENT',
    ledger_address VARCHAR(42) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_devices (
    device_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(255) NOT NULL,
    device_model VARCHAR(100) NOT NULL,
    os_version VARCHAR(50) NOT NULL,
    fcm_token TEXT,
    is_trusted BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_device UNIQUE(user_id, device_fingerprint)
);

CREATE TABLE auth_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES user_devices(device_id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_auth_sessions_user ON auth_sessions(user_id) WHERE NOT is_revoked;
```

## Security & Compliance Notes
- **Data Protection & Privacy (DPDP Act 2023):** All PII at rest (phone number, email) is encrypted using AES-256-GCM envelope encryption with keys managed in KMS. Phone numbers and emails are masked before logging or emitting to internal analytics.
- **Password & MPIN Storage:** MPINs are hashed using Argon2id with memory cost 64MB, 3 iterations, and 4 degrees of parallelism. No raw MPINs are logged or cached.
- **Brute Force Protection:** Redis rate-limiters enforce maximum 5 consecutive failed MPIN attempts before locking the account for 30 minutes.
- **Zero-Trust Network:** All inter-service communication requires mutual TLS (mTLS) with short-lived X.509 service certificates.

## Acceptance Criteria
- [ ] User & Identity Service starts cleanly and passes all health check probes (`/healthz`, `/livez`).
- [ ] User registration flow successfully handles phone verification, MPIN hashing, and deterministic ledger address derivation.
- [ ] Biometric WebAuthn enrollment and signature validation pass with hardware key assertions.
- [ ] Refresh token rotation prevents token replay; revoked tokens are immediately rejected across all instances via Redis.
- [ ] gRPC `ValidateSession` RPC responds in $< 2\text{ms}$ under 1,000 req/sec load.
- [ ] Zero PII is written to logs, Kafka topics, or blockchain address associations.
- [ ] Unit and integration test suite achieves $\ge 85\%$ test coverage across all domain modules.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Design Standards), 105 (Auth Architecture), 111 (Domain Model), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 202 (KYC Service), 203 (Wallet Service), 211 (Notification Service).
- **Downstream Blockers:** 204 (Order Service), 505 (Flutter Auth UI), 601 (Web App).
