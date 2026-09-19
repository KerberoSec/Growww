# 263 - SEBI Margin Pledge & Re-Pledge Depository Gateway (Go)

## Purpose
Prior to the implementation of the Securities and Exchange Board of India (SEBI) Margin Pledge and Re-Pledge framework, trading members routinely obtained broad Power of Attorney (PoA) authorizations from retail and institutional clients. Brokerages transferred client securities into proprietary pool accounts or client margin trading accounts to post collateral with clearing members and clearing corporations. This practice created significant counterparty risk, systemic vulnerability to member defaults, unauthorized hypothecation, and illegal commingling of client assets.

To eliminate these vulnerabilities, SEBI issued comprehensive circulars `SEBI/HO/MIRSD/DOP/CIR/P/2020/28` (dated February 25, 2020) and `SEBI/HO/MIRSD/DOP/CIR/P/2020/143` (effective September 1, 2020), mandating that:
1. Client securities never leave the investor's own Demat account for margin obligations.
2. Demat securities can only be pledged in favor of the Trading Member (TM) Client Collateral Pledge Account (`TMCCPA`).
3. Trading Members can only re-pledge such securities to the Clearing Member (CM) Client Collateral Pledge Account (`CMCCPA`), which in turn re-pledges them directly to the designated Clearing Corporation (CC).
4. Every pledge and re-pledge instruction requires direct depositor authentication via depository-generated One-Time Passwords (OTP) or Electronic Delivery Instruction Slip (e-DIS) PINs sent directly by depositories to the investor's registered mobile number and email.
5. Client collateral must be strictly segregated with zero commingling across proprietary, clearing, or other client positions under Unique Client Code (UCC) mapping.

The **SEBI Margin Pledge & Re-Pledge Depository Gateway** (`services/margin-pledge-gateway`) provides direct API orchestration and leased-line batch file processing with the two national central depositories: National Securities Depository Limited (NSDL CCPS - Client Collateral Pledge System) and Central Depository Services (India) Limited (CDSL OASIS - Online Authorization & Settlement Instruction System). The gateway automates the entire lifecycle of client collateral securities: pledge creation, depository OTP trigger coordination, confirmation polling, re-pledge allocation, pre-release risk verification for unpledge, default invocation upon margin shortfall, and on-chain tokenized pledge receipt minting with custodial hold flags on Hyperledger Besu.

## What You Are Building
A high-availability, mission-critical distributed Go microservice (`services/margin-pledge-gateway`) operating as the single point of truth and execution for demat collateral pledging across Indian capital markets. Key structural components include:

- **NSDL CCPS & CDSL OASIS Depository Protocol Adapters:** Direct dual-channel integration engine communicating over dedicated leased lines and IPSec VPN tunnels with NSDL CCPS and CDSL OASIS REST/SOAP APIs and batch SFTP servers. Formats and parses both ISO 20022 XML schemas (`colr.003`, `colr.004`, `semt.013`, `sese.023`) and JSON payload variants.
- **Automated Client OTP Trigger & Verification Orchestrator:** Manages the automated trigger of depository-level OTPs directly to client registered contact credentials. Coordinates the out-of-band mobile/email verification loop, validates single-use signed callbacks, and transitions the demat transaction from pending authorization to depository-confirmed pledge.
- **Pledge Lifecycle Finite State Machine (FSM):** Deterministic, event-driven state engine tracking every collateral unit through granular operational states: `PLEDGE_REQUESTED`, `OTP_TRIGGERED`, `OTP_AUTHENTICATED`, `PLEDGED_TO_TM`, `REPLEDGE_INITIATED`, `REPLEDGED_TO_CC`, `UNPLEDGE_REQUESTED`, `UNPLEDGE_RISK_VERIFIED`, `UNPLEDGE_CONFIRMED`, `INVOCATION_INITIATED`, `INVOKED_DEFAULT`, and `PLEDGE_REJECTED`.
- **Pre-Release Risk & Solvency Verifier:** Synchronous gateway client connecting to the Risk & Margin Service (Prompt 206) to ensure that no unpledge request is dispatched to depositories unless the client maintains an excess margin buffer (minimum 105% post-release margin coverage) against active open derivatives and cash exposures.
- **Margin Deficit Default Invocation Pipeline:** High-priority workflow triggered by the Risk Service (Prompt 206) during critical margin calls, invoking pledged shares directly through depository channels into the designated member liquidation demat account for market liquidation.
- **Leased-Line SFTP Batch Ingestion & Reconciliation Poller:** Resilient, cron-driven SFTP worker pool with automated failover for transferring high-volume end-of-day (EOD) and intraday batch files (`*.DAT`, `*.ACK`, `*.RSP`), handling OpenPGP/GPG envelope encryption, SHA-256 checksum validation, and automated ledger reconciliation.
- **Redis Distributed Locking Engine:** Concurrency management layer utilizing Redis distributed mutexes (`redsync/v4`) to eliminate race conditions, double-pledging, and concurrent unpledge requests on identical ISINs and Demat Beneficiary Owner IDs (BOIDs).
- **Hyperledger Besu Tokenized Pledge Receipt Relayer:** Event relayer interfacing with the permissioned Hyperledger Besu ledger via `go-ethereum/ethclient`, minting ERC-1155/ERC-3643 compatible tokenized pledge receipts (`PledgeReceiptToken.sol`) with active custodial hold flags to provide immutable, cryptographic proof of 1:1 custody holding without storing Personally Identifiable Information (PII).
- **Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically calculates, bills, and logs the mandatory platform invariant 0.00% transaction fee (No fee at all) on collateral turnover, routing revenues into the designated deterministic split (0.00% fee at launch; future fee parameters governed by FeeController.sol).

## Scope Boundaries
- **In Scope:**
  - Direct REST and SFTP leased-line integration with NSDL CCPS and CDSL OASIS gateways.
  - End-to-end demat collateral lifecycle orchestration: creation, OTP trigger, confirmation, re-pledge to CC, unpledge, and invocation.
  - Out-of-band client OTP session management with 180-second TTL nonces.
  - Automated XML (ISO 20022) and JSON serialization, payload validation, and PKCS#7 cryptographic signing.
  - Redis distributed locking on BOID + ISIN pairs to prevent race conditions and double-pledging.
  - Pre-unpledge margin solvency verification with Risk & Margin Service (Prompt 206).
  - Collateral balance synchronization and credit ledger reflection with Wallet Account Service (Prompt 203).
  - Minting and releasing on-chain tokenized pledge receipts with custodial hold flags on Hyperledger Besu.
  - Batch SFTP file generation, PGP signing/encryption, polling, and response reconciliation.
  - Zero-commingling enforcement and client UCC mapping under SEBI circular `SEBI/HO/MIRSD/DOP/CIR/P/2020/28`.
- **Out of Scope / Handled Elsewhere:**
  - Dynamic collateral haircut calculation, VaR/ELM margin percentages, and price tick ingestion (handled in Prompt 264 Dynamic Collateral Haircut Engine).
  - Multi-CCP interoperability routing between NSCCL and ICCL (handled in Prompt 260 Clearing Corporation Interoperability Service).
  - Real-time pre-trade risk checks and SPAN margin mathematics (handled in Prompt 206 and Prompt 241).
  - Secondary market order matching and trade execution (handled in Prompt 205 Order Matching Engine).
  - Fiat bank deposits, UPI collections, and cash payouts (handled in Prompt 203 Wallet Account Service and Prompt 212 Payment Gateway Service).
  - Primary user KYC verification, PAN validation, and biometric onboarding (handled in Prompt 201 User Service and Prompt 202 KYC/AML Service).

## Technology to Use
- **Core Language & Runtime:** **Go 1.22+** utilizing goroutines, structured context propagation, bounded worker pools, and low-latency garbage collection optimized for microsecond IPC.
- **XML / JSON Dual Serialization:**
  - `encoding/xml` with strict schema validation for ISO 20022 XML formats (`colr.003`, `colr.004`, `semt.013`, `sese.023`).
  - `bytedance/sonic` or `github.com/goccy/go-json` for zero-allocation high-speed JSON serialization.
- **Database & Persistence:** **PostgreSQL 16+** with `jackc/pgx/v5` connection pooling, strict ACID transactions, table partitioning by pledge creation date, row-level locking (`SELECT FOR UPDATE`), and immutable audit logging.
- **Distributed Caching & Concurrency:** **Redis 7.2+ Cluster** with `go-redsync/redsync/v4` for multi-node distributed locks, rate-limiting, and ephemeral 180-second OTP authorization session nonces.
- **Event Streaming & Message Broker:** **Apache Kafka 3.7+** using `confluent-kafka-go` (librdkafka C-bindings) with idempotent producers (`enable.idempotence=true`), snappy compression, and strict manual offset commits.
- **Transport Protocols & Depository Leased Lines:**
  - Mutual TLS (mTLS 1.3) with client certificates and RSA-4096 / ECDSA P-256 keys.
  - PKCS#7 / CMS detached cryptographic digital signatures using Hardware Security Module (HSM) FIPS 140-2 Level 3 integration.
  - Leased-line SFTP via `golang.org/x/crypto/ssh` and `github.com/pkg/sftp`.
  - OpenPGP / GnuPG for batch file encryption and decryption (`ProtonMail/gopenpgp/v2`).
- **Internal Microservice RPC:** **gRPC / Protocol Buffers v3** with HTTP/2 multiplexing, keepalive pings, and mutual TLS over private VPC networks.
- **Blockchain Client:** `github.com/ethereum/go-ethereum/ethclient` communicating with permissioned Hyperledger Besu QBFT nodes over mTLS JSON-RPC.

## Backend / Infra Touchpoints
- **External Depository Endpoints:**
  - NSDL CCPS REST Gateway: `https://ccps.nsdl.co.in/api/v2/pledge` (Primary leased line with IPSec VPN backup).
  - NSDL SFTP Leased-Line Server: `sftp://sftp.ccps.nsdl.co.in:22` (Directories: `/inward/pledge`, `/outward/response`, `/reports/daily`).
  - CDSL OASIS REST Gateway: `https://oasis.cdslindia.com/oasisapi/v1/pledge` (Primary leased line with IPSec VPN backup).
  - CDSL SFTP Leased-Line Server: `sftp://sftp.oasis.cdslindia.com:22` (Directories: `/upload/pledge`, `/download/response`, `/reports`).
- **PostgreSQL Database Tables:**
  - `margin_pledge_transactions`: Master ledger of client demat pledge, re-pledge, unpledge, and invocation lifecycles.
  - `margin_repledge_allocations`: Tracks re-pledge assignments from TM Client Collateral Account to CM and CC.
  - `demat_security_collateral`: Real-time inventory of pledged ISINs, held quantities, valuations, and depository references.
  - `depository_sftp_batches`: Tracks outbound and inbound SFTP batch file transmissions, checksums, and parsing states.
  - `pledge_invocation_records`: Audit log of default margin invocations executed into member liquidation accounts.
  - `pledge_otp_sessions`: Ephemeral records of depository OTP triggers, session nonces, and client callback states.
  - `tokenized_pledge_receipts`: Tracks Besu on-chain minted receipt token IDs, transaction hashes, and custodial hold flags.
  - `pledge_gateway_audit_trail`: WORM append-only cryptographic audit log of all raw outbound/inbound payloads.
  - `pledge_platform_fee_distributions`: Ledger recording the 0.00% (Zero Fee) platform fee assessment and 0.00% fee launch policy revenue split.
- **Redis Cache & Distributed Locks:**
  - `lock:pledge:boid:{hashed_boid}:{isin}`: Distributed mutex ensuring single-flight pledge/unpledge execution per asset.
  - `session:otp:pledge:{pledge_id}`: Ephemeral depository OTP authorization session (TTL: 180 seconds).
  - `idempotency:pledge:{request_id}`: Distributed deduplication key preventing duplicate instruction dispatches.
  - `cache:isin:eligibility:{isin}`: In-memory cache of approved collateral scrips and statutory haircut parameters.
- **Kafka Topics:**
  - **Consumes:**
    - `risk.unpledge_verification_response.v1`: Ingests solvency evaluation results from Risk Service for unpledge requests.
    - `risk.margin_call_invocation.v1`: Ingests urgent liquidation triggers requiring collateral invocation.
    - `depository.otp_callback.v1`: Ingests webhook callbacks from client mobile/web interfaces following depository OTP input.
    - `market.price_update.v1`: Real-time price updates for live collateral mark-to-market revaluation.
  - **Publishes:**
    - `pledge.initiated.v1`: Emitted when pledge instruction is generated and submitted to depository.
    - `pledge.otp_triggered.v1`: Emitted when depository confirms OTP dispatch to client registered mobile/email.
    - `pledge.confirmed.v1`: Emitted when depository confirms successful pledge creation in TMCCPA.
    - `pledge.repledged.v1`: Emitted when clearing corporation confirms re-pledge allocation.
    - `pledge.unpledge_initiated.v1`: Emitted when unpledge request passes risk solvency verification.
    - `pledge.unpledge_confirmed.v1`: Emitted when depository confirms collateral release back to free demat balance.
    - `pledge.invoked.v1`: Emitted when default invocation transfers shares to member liquidation account.
    - `pledge.receipt_minted.v1`: Emitted when Hyperledger Besu mints tokenized pledge receipt with custodial hold flag.
    - `pledge.fee_assessed.v1`: Emitted when the 0.00% (Zero Fee) platform fee is recorded.
- **Internal Microservice Interfaces:**
  - **Risk & Margin Service (Prompt 206):** Ingests real-time collateral credit; validates post-unpledge account solvency.
  - **Wallet Account Service (Prompt 203):** Credits non-cash margin balances; debits unpledged collateral values.
  - **Custodian Depository Integration Service (Prompt 213):** Synchronizes broader depository participant holding statements.
  - **Clearing Corporation Interoperability Service (Prompt 260):** Orchestrates multi-CCP allocation of re-pledged assets.
  - **Dynamic Collateral Haircut Engine (Prompt 264):** Ingests dynamic haircut rates for asset valuation.
  - **Notification Service (Prompt 211):** Delivers transactional status alerts and in-app instructions to investors.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Tokenized Pledge Receipt Minting with Custodial Hold:**
  - Upon receiving an authenticated pledge confirmation from NSDL or CDSL (`colr.004` or JSON confirmation), the gateway formats an on-chain transaction calling `PledgeReceiptRegistry.sol`:
    ```solidity
    function mintPledgeReceipt(
        address clientWallet,
        address assetToken,
        uint256 quantity,
        bytes32 depositoryRef,
        bytes32 hashedBoid,
        uint8 depositoryType
    ) external returns (uint256 receiptTokenId);
    ```
  - The minted receipt represents a 1:1 tokenized depository collateral claim.
  - The contract immediately activates an internal `CustodialHold` flag on the receipt:
    ```solidity
    struct PledgeReceipt {
        address clientWallet;
        address assetToken;
        uint256 quantity;
        bytes32 depositoryRef;
        bytes32 hashedBoid;
        bool custodialHold;
        address pledgedToMember;
        address repledgedToCC;
        uint256 timestamp;
    }
    ```
  - While `custodialHold == true`, any attempt to transfer, burn, or wrap the tokenized receipt on-chain reverts automatically.
- **Custodial Hold Release on Unpledge:**
  - When the depository confirms unpledge execution, the gateway invokes:
    ```solidity
    function releasePledgeReceipt(
        uint256 receiptTokenId,
        bytes32 releaseRef
    ) external;
    ```
  - The smart contract verifies the caller is the authorized `MarginPledgeGatewayRelayer`, clears the `CustodialHold` flag, and burns the tokenized receipt, restoring free asset balance.
- **Invocation Liquidation Transfer:**
  - In the event of client default where securities are invoked by the member, the gateway invokes:
    ```solidity
    function invokePledgeReceipt(
        uint256 receiptTokenId,
        address liquidationAccount,
        bytes32 invocationRef
    ) external;
    ```
  - The smart contract transfers ownership of the tokenized collateral claim to the designated member liquidation pool.
- **Zero Personally Identifiable Information (Zero PII):**
  - Demat Beneficiary Owner IDs (BOIDs), Permanent Account Numbers (PANs), and client names are never committed to the blockchain.
  - BOIDs are cryptographically hashed using salted HMAC-SHA256:
    $$\text{HashedBOID} = \text{HMAC-SHA256}(\text{BOID}, \text{Salt}_{\text{Pledge}})$$
  - All on-chain events reference only Ethereum account addresses, `HashedBOID`, token addresses, quantities, and depository transaction reference hashes.
- **QBFT Consensus & 1:1 Custody Reconciliation:**
  - Every on-chain mint and burn transaction achieves deterministic finality through the Istanbul/QBFT Byzantine Fault Tolerant consensus engine.
  - Daily EOD depository holding statements (`semt.013`) are hashed into a Merkle tree root and notarized into `CustodyNotary.sol`, cryptographically validating 1:1 depository backing against total on-chain receipt tokens.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Go Service Architecture:** Scaffold Go 1.22+ module `services/margin-pledge-gateway` with a modular internal package layout (`cmd/server`, `internal/depository/nsdl`, `internal/depository/cdsl`, `internal/fsm`, `internal/otp`, `internal/sftp`, `internal/risk`, `internal/blockchain`, `internal/storage`).
2. **Define Protobuf Service Contracts:** Author `proto/growww/pledge/v1/margin_pledge_service.proto` defining RPCs for `InitiateMarginPledge`, `TriggerDepositoryOtp`, `ConfirmPledgeOtp`, `InitiateMarginRePledge`, `InitiateMarginUnpledge`, `InvokeMarginPledge`, `GetPledgeStatus`, and `QueryCollateralHoldings`.
3. **Generate Go gRPC Stubs:** Compile protobuf definitions using `protoc-gen-go` and `protoc-gen-go-grpc`, configuring gRPC server interceptors for distributed tracing, metrics, recovery, and request validation.
4. **Implement PostgreSQL Database Migrations:** Write database migration scripts defining tables for pledge transactions, re-pledges, demat holdings, SFTP batches, invocation records, fee ledgers, and audit trails with composite indexes and constraints.
5. **Implement Redis Distributed Locking:** Deploy a robust concurrency control package wrapping `go-redsync/redsync/v4` to acquire distributed locks on `{hashed_boid}:{isin}` before any state mutation, enforcing single-flight execution.
6. **Build Dual XML/JSON Message Engine:** Implement streaming parsers for ISO 20022 XML messages (`colr.003` CollateralProposal, `colr.004` ProposalResponse, `semt.013` IntraPositionReport, `sese.023` SettlementInstruction) and corresponding depository REST JSON schemas with strict field-level validation.
7. **Implement PKCS#7 / CMS Cryptographic Signer:** Author the digital signature utility interfacing with HSM or software keystores, signing outbound depository requests using detached PKCS#7 signatures with SHA-256 and client mTLS certificates.
8. **Build NSDL CCPS Client Adapter:** Implement the NSDL CCPS client supporting REST endpoints for real-time pledge creation, unpledge, and invocation, handling mTLS handshakes, leased-line routing, keepalives, and automatic retries with exponential backoff.
9. **Build CDSL OASIS Client Adapter:** Implement the CDSL OASIS client supporting REST and SOAP endpoints for e-DIS pledge creation, TPIN/OTP verification, and transaction status queries over leased lines.
10. **Implement Client Depository OTP Orchestrator:** Build the OTP lifecycle orchestrator: dispatch depository OTP requests via NSDL/CDSL APIs, generate short-lived cryptographically random 256-bit session tokens in Redis (TTL: 180 seconds), handle client verification callbacks, and transition pledge records upon confirmation.
11. **Implement Leased-Line SFTP Batch Worker:** Build automated SFTP transfer and polling workers using `golang.org/x/crypto/ssh` and `pkg/sftp`. Implement OpenPGP file encryption/decryption, SHA-256 checksum verification, and batch parser for depository acknowledgment files (`*.ACK`, `*.RSP`).
12. **Construct Pledge Lifecycle Finite State Machine (FSM):** Build a resilient, event-driven state machine managing atomic transitions from `PLEDGE_REQUESTED` through to `REPLEDGED_TO_CC`, `UNPLEDGE_CONFIRMED`, or `INVOKED_DEFAULT`. Enforce that invalid state transitions fail with deterministic error codes.
13. **Implement Pre-Release Risk & Solvency Pipeline:** Connect synchronously to Risk & Margin Service (Prompt 206) via gRPC. Before dispatching any unpledge instruction to depositories, evaluate post-release portfolio margin solvency; block instruction immediately if post-release margin buffer is under 105%.
14. **Implement Hyperledger Besu Tokenized Receipt Relayer:** Integrate `go-ethereum/ethclient` to invoke `mintPledgeReceipt()`, `releasePledgeReceipt()`, and `invokePledgeReceipt()` on `PledgeReceiptRegistry.sol`, validating transaction receipts and QBFT finality.
15. **Implement Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically calculate the 0.00% (No fee at all) platform transaction fee on collateral value, recording the deterministic 0.00% fee at launch (governed by FeeController.sol) revenue split in PostgreSQL.
16. **Author Unit, Integration, and Chaos Test Suites:** Develop automated test suites covering ISO 20022 serialization, mock NSDL/CDSL gateway responses, Redis distributed lock race condition simulations, OTP expiration timeouts, and end-to-end unpledge risk blocking.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/pledge/v1/margin_pledge_service.proto`)
```protobuf
syntax = "proto3";

package growww.pledge.v1;

option go_package = "github.com/growww/proto/gen/go/pledge/v1;pledgev1";

enum DepositoryType {
  DEPOSITORY_TYPE_UNSPECIFIED = 0;
  DEPOSITORY_TYPE_NSDL = 1;
  DEPOSITORY_TYPE_CDSL = 2;
}

enum PledgeType {
  PLEDGE_TYPE_UNSPECIFIED = 0;
  PLEDGE_TYPE_MARGIN_PLEDGE = 1;     // Client Demat -> TM Client Collateral Account
  PLEDGE_TYPE_MARGIN_REPLEDGE = 2;   // TM Client Collateral Account -> CM / CC
  PLEDGE_TYPE_MARGIN_UNPLEDGE = 3;   // Release collateral back to Client Demat
  PLEDGE_TYPE_MARGIN_INVOCATION = 4; // Member default invocation into liquidation demat
}

enum PledgeStatus {
  PLEDGE_STATUS_UNSPECIFIED = 0;
  PLEDGE_STATUS_REQUESTED = 1;
  PLEDGE_STATUS_OTP_TRIGGERED = 2;
  PLEDGE_STATUS_OTP_AUTHENTICATED = 3;
  PLEDGE_STATUS_PLEDGED_TO_TM = 4;
  PLEDGE_STATUS_REPLEDGE_INITIATED = 5;
  PLEDGE_STATUS_REPLEDGED_TO_CC = 6;
  PLEDGE_STATUS_UNPLEDGE_REQUESTED = 7;
  PLEDGE_STATUS_UNPLEDGE_RISK_VERIFIED = 8;
  PLEDGE_STATUS_UNPLEDGE_CONFIRMED = 9;
  PLEDGE_STATUS_INVOCATION_INITIATED = 10;
  PLEDGE_STATUS_INVOKED_DEFAULT = 11;
  PLEDGE_STATUS_REJECTED = 12;
  PLEDGE_STATUS_EXPIRED = 13;
}

message CollateralSecuritiesItem {
  string isin = 1;
  string symbol = 2;
  uint64 quantity = 3;
  uint64 ltp_paise = 4;
  uint32 haircut_bps = 5; // e.g., 2000 = 20.00%
  uint64 collateral_value_paise = 6;
  uint64 effective_margin_credit_paise = 7;
}

message InitiateMarginPledgeRequest {
  string request_id = 1;
  string account_id = 2;
  string hashed_boid = 3;
  DepositoryType depository = 4;
  repeated CollateralSecuritiesItem securities = 5;
  string ucc_code = 6;
  string client_ip = 7;
  string user_agent = 8;
}

message InitiateMarginPledgeResponse {
  string pledge_id = 1;
  PledgeStatus status = 2;
  string depository_reference = 3;
  string otp_session_token = 4;
  int64 session_expires_at_unix_ns = 5;
  string message = 6;
}

message TriggerDepositoryOtpRequest {
  string pledge_id = 1;
  string otp_session_token = 2;
}

message TriggerDepositoryOtpResponse {
  string pledge_id = 1;
  PledgeStatus status = 2;
  string masked_mobile_number = 3;
  string masked_email = 4;
  int64 otp_expires_at_unix_ns = 5;
  bool otp_dispatched = 6;
}

message ConfirmPledgeOtpRequest {
  string pledge_id = 1;
  string otp_session_token = 2;
  string otp_code = 3;
}

message ConfirmPledgeOtpResponse {
  string pledge_id = 1;
  PledgeStatus status = 2;
  string depository_confirmation_ref = 3;
  uint64 total_collateral_value_paise = 4;
  uint64 total_margin_credit_paise = 5;
  uint64 platform_fee_paise = 6; // 0.00% (No fee at all) platform fee
  string onchain_receipt_token_id = 7;
  string onchain_tx_hash = 8;
  bool custodial_hold_active = 9;
}

message InitiateMarginRePledgeRequest {
  string pledge_id = 1;
  string target_clearing_corp = 2; // "NSCCL" or "ICCL"
  string cm_ccpa_account = 3;      // Clearing Member Client Collateral Pledge Account
  string cc_account = 4;           // Clearing Corporation Account
}

message InitiateMarginRePledgeResponse {
  string repledge_id = 1;
  PledgeStatus status = 2;
  string cc_confirmation_ref = 3;
  int64 repledged_at_unix_ns = 4;
}

message InitiateMarginUnpledgeRequest {
  string request_id = 1;
  string account_id = 2;
  string pledge_id = 3;
  string isin = 4;
  uint64 quantity = 5;
  string reason = 6;
}

message InitiateMarginUnpledgeResponse {
  string unpledge_request_id = 1;
  PledgeStatus status = 2;
  bool risk_margin_check_passed = 3;
  uint64 remaining_margin_buffer_paise = 4;
  string depository_unpledge_ref = 5;
  string message = 6;
}

message InvokeMarginPledgeRequest {
  string pledge_id = 1;
  string account_id = 2;
  string isin = 3;
  uint64 quantity = 4;
  string default_notice_id = 5;
  string liquidation_demat_account = 6;
  string authorized_by = 7;
}

message InvokeMarginPledgeResponse {
  string invocation_id = 1;
  PledgeStatus status = 2;
  string depository_invocation_ref = 3;
  uint64 invoked_value_paise = 4;
  string onchain_invocation_tx_hash = 5;
}

message GetPledgeStatusRequest {
  string pledge_id = 1;
}

message GetPledgeStatusResponse {
  string pledge_id = 1;
  string account_id = 2;
  string hashed_boid = 3;
  DepositoryType depository = 4;
  PledgeStatus status = 5;
  repeated CollateralSecuritiesItem securities = 6;
  string depository_ref = 7;
  string cc_repledge_ref = 8;
  uint64 total_collateral_value_paise = 9;
  uint64 effective_margin_credit_paise = 10;
  bool custodial_hold_active = 11;
  int64 created_at_unix_ns = 12;
  int64 updated_at_unix_ns = 13;
}

message QueryCollateralHoldingsRequest {
  string account_id = 1;
  bool active_only = 2;
}

message QueryCollateralHoldingsResponse {
  string account_id = 1;
  repeated CollateralSecuritiesItem holdings = 2;
  uint64 aggregate_collateral_value_paise = 3;
  uint64 aggregate_margin_credit_paise = 4;
}

service MarginPledgeService {
  rpc InitiateMarginPledge (InitiateMarginPledgeRequest) returns (InitiateMarginPledgeResponse);
  rpc TriggerDepositoryOtp (TriggerDepositoryOtpRequest) returns (TriggerDepositoryOtpResponse);
  rpc ConfirmPledgeOtp (ConfirmPledgeOtpRequest) returns (ConfirmPledgeOtpResponse);
  rpc InitiateMarginRePledge (InitiateMarginRePledgeRequest) returns (InitiateMarginRePledgeResponse);
  rpc InitiateMarginUnpledge (InitiateMarginUnpledgeRequest) returns (InitiateMarginUnpledgeResponse);
  rpc InvokeMarginPledge (InvokeMarginPledgeRequest) returns (InvokeMarginPledgeResponse);
  rpc GetPledgeStatus (GetPledgeStatusRequest) returns (GetPledgeStatusResponse);
  rpc QueryCollateralHoldings (QueryCollateralHoldingsRequest) returns (QueryCollateralHoldingsResponse);
}
```

### PostgreSQL Database Schema (`services/margin-pledge-gateway/migrations/001_initial_schema.sql`)
```sql
-- PostgreSQL Migration: SEBI Margin Pledge & Re-Pledge Depository Gateway
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums for Depository, Pledge Type, and Lifecycle Status
CREATE TYPE depository_type_enum AS ENUM ('NSDL', 'CDSL');
CREATE TYPE pledge_type_enum AS ENUM ('MARGIN_PLEDGE', 'MARGIN_REPLEDGE', 'MARGIN_UNPLEDGE', 'MARGIN_INVOCATION');
CREATE TYPE pledge_status_enum AS ENUM (
    'REQUESTED',
    'OTP_TRIGGERED',
    'OTP_AUTHENTICATED',
    'PLEDGED_TO_TM',
    'REPLEDGE_INITIATED',
    'REPLEDGED_TO_CC',
    'UNPLEDGE_REQUESTED',
    'UNPLEDGE_RISK_VERIFIED',
    'UNPLEDGE_CONFIRMED',
    'INVOCATION_INITIATED',
    'INVOKED_DEFAULT',
    'REJECTED',
    'EXPIRED'
);

-- Master Table for Margin Pledge Transactions
CREATE TABLE margin_pledge_transactions (
    pledge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(64) NOT NULL UNIQUE,
    account_id VARCHAR(64) NOT NULL,
    hashed_boid VARCHAR(64) NOT NULL,
    depository depository_type_enum NOT NULL,
    pledge_type pledge_type_enum NOT NULL DEFAULT 'MARGIN_PLEDGE',
    status pledge_status_enum NOT NULL DEFAULT 'REQUESTED',
    ucc_code VARCHAR(16) NOT NULL,
    tm_ccpa_boid VARCHAR(16) NOT NULL,
    cm_ccpa_boid VARCHAR(16),
    cc_account_id VARCHAR(32),
    depository_reference VARCHAR(64),
    cc_repledge_reference VARCHAR(64),
    gross_collateral_value_paise BIGINT NOT NULL DEFAULT 0,
    net_margin_credit_paise BIGINT NOT NULL DEFAULT 0,
    platform_fee_paise BIGINT NOT NULL DEFAULT 0, -- 0.00% (Zero Fee) platform fee
    error_code VARCHAR(32),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Itemized Demat Collateral Securities Lines
CREATE TABLE demat_security_collateral (
    collateral_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_id UUID NOT NULL REFERENCES margin_pledge_transactions(pledge_id) ON DELETE CASCADE,
    account_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    repledged_quantity BIGINT NOT NULL DEFAULT 0,
    ltp_paise BIGINT NOT NULL DEFAULT 0,
    haircut_bps INT NOT NULL DEFAULT 2000, -- 20.00%
    collateral_value_paise BIGINT NOT NULL DEFAULT 0,
    effective_margin_credit_paise BIGINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_repledged BOOLEAN NOT NULL DEFAULT FALSE,
    is_unpledged BOOLEAN NOT NULL DEFAULT FALSE,
    is_invoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Margin Re-Pledge Allocation Records
CREATE TABLE margin_repledge_allocations (
    repledge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_id UUID NOT NULL REFERENCES margin_pledge_transactions(pledge_id),
    target_clearing_corp VARCHAR(16) NOT NULL, -- 'NSCCL' or 'ICCL'
    cm_ccpa_account VARCHAR(32) NOT NULL,
    cc_account VARCHAR(32) NOT NULL,
    allocated_margin_credit_paise BIGINT NOT NULL,
    depository_repledge_ref VARCHAR(64),
    cc_acknowledgment_ref VARCHAR(64),
    status pledge_status_enum NOT NULL DEFAULT 'REPLEDGE_INITIATED',
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Depository OTP Session Management (Ephemeral Ledger)
CREATE TABLE pledge_otp_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_id UUID NOT NULL REFERENCES margin_pledge_transactions(pledge_id) ON DELETE CASCADE,
    session_token VARCHAR(64) NOT NULL UNIQUE,
    depository_txn_id VARCHAR(64),
    masked_mobile VARCHAR(16),
    masked_email VARCHAR(64),
    otp_dispatched_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts_count INT NOT NULL DEFAULT 0,
    is_consumed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Depository Batch SFTP Transmission & Parsing Ledger
CREATE TABLE depository_sftp_batches (
    batch_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    depository depository_type_enum NOT NULL,
    batch_direction VARCHAR(8) NOT NULL CHECK (batch_direction IN ('INWARD', 'OUTWARD')),
    file_name VARCHAR(128) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    sha256_checksum VARCHAR(64) NOT NULL,
    records_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    is_pgp_encrypted BOOLEAN NOT NULL DEFAULT TRUE,
    processing_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    processed_at TIMESTAMPTZ,
    error_summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Default Margin Invocations
CREATE TABLE pledge_invocation_records (
    invocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_id UUID NOT NULL REFERENCES margin_pledge_transactions(pledge_id),
    account_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    invoked_quantity BIGINT NOT NULL CHECK (invoked_quantity > 0),
    invoked_value_paise BIGINT NOT NULL,
    default_notice_id VARCHAR(64) NOT NULL,
    liquidation_demat_account VARCHAR(32) NOT NULL,
    depository_invocation_ref VARCHAR(64),
    authorized_by VARCHAR(64) NOT NULL,
    invoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Hyperledger Besu Tokenized Pledge Receipts
CREATE TABLE tokenized_pledge_receipts (
    receipt_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_id UUID NOT NULL REFERENCES margin_pledge_transactions(pledge_id),
    receipt_token_id VARCHAR(78) NOT NULL, -- uint256 string
    token_contract_address VARCHAR(42) NOT NULL,
    client_wallet_address VARCHAR(42) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    token_quantity NUMERIC(78, 0) NOT NULL,
    custodial_hold_active BOOLEAN NOT NULL DEFAULT TRUE,
    mint_tx_hash VARCHAR(66) NOT NULL,
    release_tx_hash VARCHAR(66),
    block_number BIGINT NOT NULL,
    minted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ
);

-- Platform 0.00% (Zero Fee) Collateral Fee Ledger
CREATE TABLE pledge_platform_fee_distributions (
    fee_distribution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_id UUID NOT NULL REFERENCES margin_pledge_transactions(pledge_id),
    account_id VARCHAR(64) NOT NULL,
    collateral_turnover_paise BIGINT NOT NULL,
    total_fee_paise BIGINT NOT NULL,       -- 0.00% (Zero Fee)
    platform_treasury_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    core_sgf_paise BIGINT NOT NULL,          -- Governed by FeeController (0.00% at launch)
    ipf_paise BIGINT NOT NULL,               -- Governed by FeeController (0.00% at launch)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Cryptographic Append-Only WORM Audit Trail
CREATE TABLE pledge_gateway_audit_trail (
    audit_id BIGSERIAL PRIMARY KEY,
    pledge_id UUID NOT NULL,
    account_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    previous_state VARCHAR(32),
    new_state VARCHAR(32) NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    raw_payload_json JSONB NOT NULL,
    digital_signature_base64 TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance Indexes
CREATE INDEX idx_margin_pledge_acc_status ON margin_pledge_transactions(account_id, status);
CREATE INDEX idx_margin_pledge_boid ON margin_pledge_transactions(hashed_boid);
CREATE INDEX idx_demat_collateral_acc_isin ON demat_security_collateral(account_id, isin) WHERE is_active = TRUE;
CREATE INDEX idx_repledge_pledge_id ON margin_repledge_allocations(pledge_id);
CREATE INDEX idx_otp_sessions_token ON pledge_otp_sessions(session_token);
CREATE INDEX idx_sftp_batches_depo_status ON depository_sftp_batches(depository, processing_status);
CREATE INDEX idx_tokenized_receipt_pledge ON tokenized_pledge_receipts(pledge_id);
CREATE INDEX idx_audit_pledge_event ON pledge_gateway_audit_trail(pledge_id, recorded_at);
```

## Security & Compliance Notes
- **Strict Compliance with SEBI Circular `SEBI/HO/MIRSD/DOP/CIR/P/2020/28`:**
  - Securities remain credited exclusively within the investor's own Demat account. Broker pool accounts or third-party custody transfers are strictly blocked at both the API adapter and FSM validation levels.
  - Pledges can only target the Trading Member Client Collateral Pledge Account (`TMCCPA`). Re-pledges can only target the Clearing Member Client Collateral Pledge Account (`CMCCPA`) and the designated Clearing Corporation account.
- **Zero Client Commingling & Unique Client Code (UCC) Tagging:**
  - Every pledge and re-pledge instruction carries the investor's statutory Unique Client Code (UCC).
  - Client collateral is segregated at the Clearing Corporation level. Under no circumstances can client securities be re-pledged or utilized to satisfy proprietary trading member obligations or margins of other clients.
- **Automated Depository OTP Trigger & Out-of-Band Verification:**
  - Depository OTPs are generated and dispatched exclusively by NSDL or CDSL directly to the client's registered mobile number and email address recorded in the depository master.
  - The gateway coordinates the trigger, provides status callbacks, and validates OTP verification tokens via short-lived Redis session nonces (TTL: 180 seconds).
  - Plaintext OTPs are never stored in databases, logs, caches, or traces.
- **Hardware Security Module (HSM) Cryptographic Signing:**
  - All outbound API payloads and batch SFTP instruction files to NSDL and CDSL must be digitally signed with PKCS#7 detached signatures using an RSA-4096 private key secured within an HSM certified to FIPS 140-2 Level 3.
- **Zero Personally Identifiable Information (Zero PII):**
  - Demat BOIDs, PANs, phone numbers, and email addresses are masked or hashed using salted HMAC-SHA256 before logging, caching, or on-chain transmission.
- **Redis Distributed Locking & Race Condition Elimination:**
  - All pledge and unpledge mutations on a specific `{hashed_boid}:{isin}` pair require acquiring a distributed Redis lock with a 5-second automatic TTL to eliminate concurrent double-pledging or unpledge collision attacks.
- **Immutable WORM Audit Trail:**
  - All raw ISO 20022 XML, JSON payloads, responses, status transitions, and digital signatures are recorded in an append-only audit trail table with SHA-256 payload hashing for 8-year regulatory compliance inspection under SEBI inspection guidelines.

## Acceptance Criteria
- [ ] Direct NSDL CCPS and CDSL OASIS API clients successfully serialize, sign (PKCS#7 SHA-256), transmit, and parse ISO 20022 XML and JSON requests with 100% schema compliance.
- [ ] Automated client OTP trigger coordinates with NSDL/CDSL endpoints, delivering verification prompts to clients within 2 seconds of pledge initiation.
- [ ] OTP session manager enforces a strict 180-second TTL in Redis, invalidating expired tokens and preventing replay attacks.
- [ ] Pledge lifecycle Finite State Machine executes deterministic transitions across all states with zero invalid or orphaned state transitions.
- [ ] Redis distributed locking reliably blocks concurrent pledge or unpledge attempts on the same `{hashed_boid}:{isin}` pair.
- [ ] Pre-release risk verification synchronously connects to Risk & Margin Service (Prompt 206) and aborts unpledge requests whenever the post-release account margin buffer falls below 105%.
- [ ] Margin deficit invocation workflow successfully routes invocation instructions directly to NSDL/CDSL, moving collateral shares to the member liquidation demat account.
- [ ] Leased-line SFTP worker successfully connects, decrypts PGP files, parses batch responses (`*.ACK`, `*.RSP`), and reconciles depository holding statements.
- [ ] Hyperledger Besu relayer invokes `mintPledgeReceipt()`, activates the `CustodialHold` flag, and burns/releases the receipt upon verified unpledge confirmation with QBFT consensus finality.
- [ ] Universal 0.00% (No fee at all) platform fee is accurately computed and logged across collateral transactions, distributing revenues according to the 0.00% fee at launch (governed by FeeController.sol) rule.
- [ ] End-to-end pledge initiation, OTP verification, and TM collateral confirmation completes within 3 seconds under standard depository operational conditions.
- [ ] 100% zero commingling verification: every pledged unit is mapped strictly to the client's UCC, with zero proprietary account usage.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 101 (System Architecture & Master Blueprint)
  - Prompt 102 (Bounded Contexts & Domain Models)
  - Prompt 201 (User Service & Account Context)
  - Prompt 203 (Multi-Asset Wallet & Account Service)
  - Prompt 206 (Risk & Margin Checks Service)
- **Parallel Work:**
  - Prompt 213 (Custodian Depository Integration Service)
  - Prompt 260 (Clearing Corporation Interoperability & SEBI Margin Pledge Service)
  - Prompt 264 (Dynamic Collateral Haircut & Margin Call Notification Engine)
- **Enables:**
  - Prompt 204 (Order Service - collateral margin credit utilization)
  - Prompt 205 (Order Matching Engine - pre-funded collateral backing)
  - Prompt 241 (SPAN Margin Engine - demat collateral margin offset)
  - Comprehensive, automated, SEBI-compliant non-cash margin trading for equity, currency, and derivatives participants.
