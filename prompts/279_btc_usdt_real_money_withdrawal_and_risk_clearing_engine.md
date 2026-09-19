# 279 - BTC & USDT Real-Money Withdrawal & Multi-Sig Risk Clearing Engine (Go)

## Purpose
Real-money cryptocurrency withdrawal represents the highest-risk operational vector across any regulated digital asset trading, clearing, and custody platform. Retail and institutional participants operating within the GIFT City (IFSCA) and domestic regulatory perimeters who hold tokenized or synthetic collateral representations (`wBTC` and `weUSDT`) on Growww's permissioned Hyperledger Besu clearing ledger require a secure, audited, and deterministic pathway to redeem these tokens into native on-chain Bitcoin (Native SegWit / Taproot) and Tether USDT (ERC-20 on Ethereum and TRC-20 on Tron) delivered to verified external wallet addresses.

The **BTC & USDT Real-Money Withdrawal & Multi-Sig Risk Clearing Engine** establishes an institutional-grade, zero-trust gateway that eliminates single points of failure, insider collusion, account takeover drain, and private key exposure. By combining multi-factor risk scoring, hardware-backed WebAuthn/Passkey re-authentication, mandatory 24-hour address whitelisting cooldowns, rolling 24-hour velocity caps, and programmatic Multi-Party Computation Threshold Signature Scheme (MPC-TSS) signing, the service guarantees that no native asset leaves platform custody without rigorous verification.

Crucially, the service coordinates an atomic two-phase burn-and-release settlement: synthetic representations (`wBTC`, `weUSDT`) are irreversibly burned on the Hyperledger Besu clearing ledger prior to broadcasting signed raw transactions to public blockchains. This ensures absolute 1:1 custody reserve invariance, full compliance with statutory FIU-IND and FATF Travel Rule regulations, and immutable auditability across all withdrawal operations.

---

## What You Are Building
A mission-critical, high-concurrency Go microservice (`services/crypto-withdrawal-service`) operating in a hardened zero-trust network environment. Concrete deliverables include:

- **Withdrawal Intake & Cryptographic Attestation Gateway:** Ingests user withdrawal requests with WebAuthn/FIDO2 Passkey digital signatures and Time-Based One-Time Password (TOTP) 2FA. Verifies cryptographic user intent before funds are earmarked or held.
- **Whitelisted Address Registry & 24-Hour Cooldown Engine:** Enforces strict address whitelisting. Any new external withdrawal address registered by a user is locked for an unalterable 24-hour cooldown period before it can receive funds. Provides immediate multi-channel alerting (SMS, email, push) with an instant one-click account freeze mechanism.
- **Dynamic Multi-Factor Risk Scoring Engine:** Computes an automated risk score (0 to 100) per withdrawal by evaluating client IP reputation, device fingerprint variance, recent security credential resets (password/2FA within 48 hours), rolling 24-hour velocity metrics, and counterparty address AML sanctions status via Chainalysis / TRM Labs integrations.
- **Three-Tier Approval Clearing State Machine:**
  - *Tier 1 (Automated Dispatch):* Low-risk score (< 30) and low-notional withdrawals (<= $5,000 equivalent) route automatically to the MPC signing queue without manual intervention.
  - *Tier 2 (Single Maker Review):* Moderate-risk (30 to 70) or intermediate-notional ($5,000 to $50,000) requests require review and sign-off by an authorized compliance officer.
  - *Tier 3 (Dual-Control Maker-Checker Quorum):* High-risk (> 70) or high-notional (> $50,000) requests require an asynchronous dual-control quorum (Maker compliance review + Checker executive officer approval) before cryptographic signing can proceed.
- **Two-Phase Atomic Ledger Hold & Besu Burn Coordinator:** Coordinates with Wallet Account Service (Prompt 203) to place cryptographic balance holds on synthetic assets. Upon approval, invokes the Hyperledger Besu DvP relayer to burn `wBTC` or `weUSDT` tokens, awaiting 1-block QBFT deterministic finality before dispatching the native release.
- **Multi-Chain Transaction Builder & MPC-TSS Signing Client:**
  - *Bitcoin (Native SegWit BIP 84 & Taproot BIP 86):* Executes UTXO coin selection (Branch-and-Bound / Knapsack), calculates dynamic network fees via `estimatesmartfee`, constructs BIP 174 / BIP 370 Partially Signed Bitcoin Transactions (PSBT), and dispatches to MPC Custody Service (Prompt 237) for threshold Schnorr/ECDSA signing.
  - *Ethereum (USDT ERC-20):* Encodes standard `transfer(address,uint256)` calldata for the official USDT smart contract, constructs EIP-1559 Type-2 transactions with dynamic priority fees, and requests Secp256k1 threshold signatures.
  - *Tron (USDT TRC-20):* Encodes TRC-20 `transfer(address,uint256)` TriggerSmartContract protobuf payloads, estimates energy and bandwidth requirements, and requests MPC threshold signatures.
- **Public Chain Broadcast, Mempool Watcher & Finality Verifier:** Broadcasts signed raw transactions across clustered Bitcoin Core full nodes, Ethereum execution clients, and TronGrid RPC nodes. Tracks mempool status, detects Replace-By-Fee (RBF) or stuck transactions, monitors blockchain reorganization events, and confirms final settlement upon reaching network-specific confirmation thresholds.
- **Automated FIU-IND & Travel Rule Compliance Sentinel:** Detects structuring behavior (e.g., multiple sub-threshold withdrawals), evaluates FATF Travel Rule requirements for transactions exceeding statutory thresholds ($1,000 / INR 50,000), and emits Suspicious Activity Report (SAR) alerts directly to Regulatory Reporting Service (Prompt 216 / Prompt 233).

---

## Scope Boundaries

### In Scope
- End-to-end processing of user cryptocurrency withdrawal requests for Bitcoin (BTC) and Tether (USDT) on Ethereum (ERC-20) and Tron (TRC-20).
- WebAuthn/FIDO2 Passkey authentication and TOTP 2FA verification.
- Whitelisted external address management with mandatory 24-hour security time locks.
- Real-time multi-factor risk evaluation and scoring (0-100).
- Multi-tier clearing policies: Tier 1 (Automated), Tier 2 (Single Maker), Tier 3 (Dual Maker-Checker).
- Two-phase funds hold with Wallet Account Service (Prompt 203).
- On-chain synthetic collateral burning (`wBTC`, `weUSDT`) on Hyperledger Besu.
- Bitcoin PSBT assembly, UTXO coin selection, and change address allocation.
- Ethereum EIP-1559 ERC-20 USDT transfer assembly and fee calculation.
- Tron TRC-20 USDT transfer parameter encoding and bandwidth allocation.
- Integration with MPC-TSS Custody Service (Prompt 237) for 3-of-5 threshold co-signing.
- Public node broadcast, mempool monitoring, and multi-confirmation settlement tracking.
- Automated FIU-IND structuring detection and Kafka audit event publishing.

### Out of Scope / Handled Elsewhere
- Master private key generation, physical CloudHSM key share custody, and threshold DKG (handled in Prompt 237).
- Ingress deposit monitoring and synthetic collateral minting (handled in Prompt 234, Prompt 235, and Prompt 236).
- Central Limit Order Book trading, spot matching, and margin calls (handled in Prompt 204, Prompt 205, and Prompt 206).
- Primary fiat onboarding, INR UPI/IMPS collections, and bank payouts (handled in Prompt 212 and Prompt 214).
- User identity onboarding, PAN/Aadhaar KYC verification, and initial biometric enrollment (handled in Prompt 202).
- Air-gapped cold storage physical ceremonies and offline master recovery vaults (handled in Prompt 702 and Prompt 707).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+ for high-throughput, low-latency concurrent processing, deterministic memory management, and battle-tested cryptographic libraries.
- **Relational Persistence & Audit Store:** PostgreSQL 16+ utilizing `pgx/v5` with connection pooling, strict ACID transactions, row-level locking (`SELECT FOR UPDATE`), foreign key constraints, and monthly partitioned audit tables.
- **Distributed State & In-Memory Coordination:** Redis 7.2 Cluster for distributed mutex locks (Redlock), 24-hour address lock countdowns, rolling velocity sliding windows, and idempotency deduplication keys.
- **Message Broker & Event Bus:** Apache Kafka 3.7+ (`segmentio/kafka-go`) with Protobuf serialization, strict topic partition keys (`user_id` / `withdrawal_id`), and transactional outbox publishing.
- **Blockchain Connectivity & Cryptography:**
  - `btcsuite/btcd` (wire, txscript, btcutil) for Bitcoin transaction creation, PSBT encoding, and RPC interaction with Bitcoin Core v26+.
  - `ethereum/go-ethereum` (`ethclient`, `crypto`, `accounts/abi`) for EIP-1559 transaction creation, USDT ABI encoding, and Besu JSON-RPC communication.
  - `fbsobreira/gotron-sdk` for Tron protobuf transaction assembly and TronGrid RPC interaction.
  - `github.com/go-webauthn/webauthn` for verifying FIDO2/Passkey assertion signatures and client data hashes.
  - `github.com/shopspring/decimal` for exact, zero-drift fixed-point financial arithmetic.
- **Inter-Service Communication:** gRPC (`google.golang.org/grpc`) over HTTP/2 with mutual TLS 1.3 (mTLS) and pinned X.509 certificates.
- **Hardware Security & MPC Integration:** FIPS 140-2/3 Level 3 CloudHSM client and gRPC connection to `services/mpc-tss-vault-service` (Prompt 237).
- **Compliance & AML Screening:** REST/gRPC client to Chainalysis KYT / TRM Labs with circuit breaker protection (`sony/gobreaker`).

---

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `whitelisted_withdrawal_addresses`, `crypto_withdrawal_requests`, `withdrawal_risk_evaluations`, `maker_checker_approvals`, `onchain_broadcast_transactions`, `withdrawal_audit_events`.
- **Redis 7.2 Keys:**
  - `withdrawal:address:lock:{user_id}:{address_hash}`: 24-hour TTL lock key for newly added withdrawal addresses.
  - `withdrawal:velocity:24h:{user_id}`: Redis Sorted Set tracking rolling 24-hour cumulative withdrawal notional.
  - `withdrawal:lock:idempotency:{idempotency_key}`: Distributed execution lock preventing duplicate submissions.
  - `withdrawal:rate_limit:{user_id}`: Sliding window rate limiter for withdrawal submission endpoints.
- **Kafka Topics Consumed:**
  - `wallet.balance.held.v1`: Acknowledges fund hold from Wallet Account Service (Prompt 203).
  - `wallet.balance.hold_failed.v1`: Rejection notice if source synthetic balance is insufficient.
  - `besu.collateral.burned.v1`: Confirmation of on-chain `wBTC` or `weUSDT` token burn.
  - `besu.collateral.burn_failed.v1`: Burn transaction revert or failure notice.
  - `mpc.signing.completed.v1`: Emitted when MPC-TSS produces a valid threshold signature.
  - `mpc.signing.rejected.v1`: Emitted when MPC-TSS co-signing aborts or fails policy checks.
  - `admin.maker_checker.decided.v1`: Decision events from Admin Back-Office (Prompt 217).
- **Kafka Topics Published:**
  - `crypto.withdrawal.requested.v1`: Emitted when a new withdrawal request is validated and registered.
  - `crypto.withdrawal.risk_evaluated.v1`: Emitted with calculated risk score and assigned approval tier.
  - `crypto.withdrawal.burn_initiated.v1`: Dispatched to trigger synthetic collateral burning on Besu.
  - `crypto.withdrawal.mpc_sign_requested.v1`: Dispatched to MPC Custody Service for threshold signing.
  - `crypto.withdrawal.broadcasted.v1`: Emitted when raw signed transaction is broadcast to public nodes.
  - `crypto.withdrawal.confirmed.v1`: Emitted when public transaction achieves required confirmations.
  - `crypto.withdrawal.failed.v1`: Emitted upon failure, triggering automatic balance hold release.
  - `crypto.withdrawal.fiu_suspicious_alert.v1`: Direct alert to Regulatory Reporting Service (Prompt 216 / 233).
- **Key Upstream & Downstream Services:**
  - **Wallet Account Service (Prompt 203):** Places two-phase cryptographic holds on source synthetic balances and executes final ledger balance debits.
  - **MPC Custody Service (Prompt 237):** Executes 3-of-5 threshold signing rounds for raw Bitcoin and Ethereum/Tron transactions.
  - **Admin Back-Office (Prompt 217):** Provides the human Maker-Checker UI queue for compliance and executive officers.
  - **KYC/AML Service (Prompt 202):** Provides user KYC tier, verification state, and sanctions data.
  - **Notification Service (Prompt 211):** Dispatches real-time withdrawal alerts (SMS, email, push notifications).
  - **Audit Log Service (Prompt 218):** Ingests immutable cryptographic audit trails for compliance reporting.

---

## Blockchain Interaction (Burning on Besu, signing on Bitcoin and Ethereum/Tron via MPC)
The service bridges three distinct blockchain environments under strict zero-trust invariants:

### 1. Hyperledger Besu Clearinghouse (Permissioned Ledger, QBFT Consensus)
- Synthetic representations of real-money crypto assets reside on Hyperledger Besu as permissioned ERC-20 smart contracts:
  - `wBTC`: Synthetic Wrapped Bitcoin (8 decimals).
  - `weUSDT`: Synthetic Wrapped Tether USD (6 decimals).
- To maintain an exact 1:1 Proof-of-Reserve invariant, the service cannot release native on-chain funds until the corresponding synthetic collateral is destroyed.
- Upon risk approval, the service constructs a burn transaction targeting the synthetic token contract's `burn(uint256 amount)` method via the platform's authorized DvP relayer key.
- The service monitors the Besu JSON-RPC interface until the burn transaction achieves 1-block deterministic finality under QBFT consensus (2-second block time).
- The resulting Besu `burn_tx_hash` is permanently recorded in the database record before any native transaction can be dispatched.

### 2. Native Bitcoin Blockchain (Layer-1 Mainnet)
- Supports Native SegWit (BIP 84 P2WPKH starting with `bc1q...`) and Taproot (BIP 86 P2TR starting with `bc1p...`).
- Constructs BIP 174 / BIP 370 Partially Signed Bitcoin Transactions (PSBT):
  - Fetches available unspent outputs (UTXOs) from the platform's hot vault pool.
  - Runs coin selection algorithms to match the withdrawal amount plus dynamic mining fees.
  - Constructs output 0: user's whitelisted destination address with exact satoshi amount.
  - Constructs output 1: internal hot-vault change address derived from HD path.
  - Computes exact transaction fee using live estimates from Bitcoin Core `estimatesmartfee 3 CONSERVATIVE`.
- The unsigned PSBT is transmitted to MPC Custody Service (Prompt 237) via gRPC. The 5 MPC nodes execute threshold ECDSA / Schnorr signing without assembling the private key.
- The service receives the finalized, fully signed raw transaction hex and broadcasts it to clustered Bitcoin Core full nodes via `sendrawtransaction`.

### 3. Native Ethereum & Tron Blockchains (USDT Settlement)
- **Ethereum ERC-20 USDT:**
  - Encodes the standard ERC-20 function selector `transfer(address _to, uint256 _value)` for the official Tether smart contract (`0xdAC17F958D2ee523a2206206994597C13D831ec7`).
  - Constructs an EIP-1559 Type-2 dynamic fee transaction specifying `maxPriorityFeePerGas`, `maxFeePerGas`, and gas limit (standard 65,000 gas for USDT transfers).
  - Acquires an MPC threshold Secp256k1 signature from Prompt 237, reconstructs the RLP-encoded signed envelope, and broadcasts via `eth_sendRawTransaction`.
- **Tron TRC-20 USDT:**
  - Encodes the TRC-20 smart contract trigger call targeting the Tether contract (`TR7NHqjekTsxG5Z8upEd25B448nK68Ki7c`).
  - Estimates energy and bandwidth costs, setting `fee_limit` (typically 30 TRX equivalent).
  - Acquires an MPC threshold signature, packs the transaction protobuf, and broadcasts via TronGrid JSON-RPC.

### Zero On-Chain PII Invariant
In strict compliance with DPDP Act 2023 and global privacy standards, neither Hyperledger Besu nor public blockchains (Bitcoin, Ethereum, Tron) ever receive user names, PANs, email addresses, IP addresses, or internal user IDs. All on-chain transactions reference only cryptographic wallet addresses, public key hashes, and numeric amounts.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Directory Layout:** Initialize Go module `services/crypto-withdrawal-service` with standardized layout:
   - `cmd/server/`: Main application lifecycle, dependency injection, and signal handling.
   - `internal/api/`: gRPC and REST handler implementations.
   - `internal/auth/`: WebAuthn/FIDO2 Passkey and TOTP 2FA verification logic.
   - `internal/whitelist/`: 24-hour address lock manager and verification state machine.
   - `internal/risk/`: Dynamic risk scoring engine and velocity tracking.
   - `internal/coordinator/`: Two-phase atomic settlement orchestrator.
   - `internal/besu/`: Hyperledger Besu relayer client for synthetic collateral burning.
   - `internal/builder/`: Multi-chain transaction builders (Bitcoin PSBT, Ethereum EIP-1559, Tron TRC-20).
   - `internal/mpc/`: Client adapter interfacing with MPC Custody Service (Prompt 237).
   - `internal/broadcaster/`: Public node transaction broadcast and mempool tracking worker.
   - `internal/store/`: PostgreSQL schemas, migrations, and `pgx/v5` repository queries.
2. **Define Protobuf Service Specifications:** Author `proto/growww/crypto/withdrawal/v1/crypto_withdrawal_service.proto` defining RPCs: `RegisterWithdrawalAddress`, `ListWithdrawalAddresses`, `InitiateCryptoWithdrawal`, `GetWithdrawalStatus`, `ListWithdrawalHistory`, and `MakerCheckerReviewWithdrawal`.
3. **Generate gRPC Stubs:** Compile Protobuf contracts into type-safe Go server stubs and client interfaces using `buf` and `protoc-gen-go-grpc`.
4. **Design PostgreSQL Schema & Migrations:** Write database migration scripts creating tables: `whitelisted_withdrawal_addresses`, `crypto_withdrawal_requests`, `withdrawal_risk_evaluations`, `maker_checker_approvals`, `onchain_broadcast_transactions`, and `withdrawal_audit_events`. Implement foreign key constraints, check constraints, and performance indexes.
5. **Implement Whitelisted Address Registry & 24-Hour Cooldown:** Build the address management module:
   - Validate address format via checksum parsing (BIP 84/86 Bech32 for Bitcoin, EIP-55 for Ethereum, Base58Check for Tron).
   - Enforce WebAuthn Passkey re-authentication upon address registration.
   - Store address in PostgreSQL with status `PENDING_ACTIVATION` and calculate `activated_at = NOW() + INTERVAL '24 hours'`.
   - Set Redis key `withdrawal:address:lock:{user_id}:{address_hash}` with 86,400-second TTL.
   - Dispatch immediate security alerts via Notification Service (Prompt 211).
6. **Implement WebAuthn & 2FA Verification Module:** Integrate `github.com/go-webauthn/webauthn` to verify FIDO2 assertion signatures against user public keys stored in User Service (Prompt 201). Verify TOTP tokens using HMAC-SHA1 RFC 6238.
7. **Build Dynamic Risk Scoring & Velocity Engine:** Implement the multi-factor risk assessment pipeline:
   - Query Redis sorted sets for 24-hour cumulative withdrawal notional.
   - Check user profile metadata for security updates within the prior 48 hours (password reset, email change, 2FA re-enrollment).
   - Evaluate client IP and device fingerprint consistency.
   - Query Chainalysis / TRM Labs API to verify external address sanctions status.
   - Compute final risk score (0-100) and assign approval tier (Tier 1 Automated, Tier 2 Maker, Tier 3 Maker-Checker).
8. **Build Two-Phase Atomic Settlement Coordinator:** Implement the execution state machine:
   - *Phase 1 (Hold):* Request synchronous cryptographic balance hold from Wallet Account Service (Prompt 203) for the required synthetic asset amount (`wBTC` or `weUSDT`) plus platform withdrawal fee.
   - *Phase 2 (Risk/Queue Gate):* If Tier 1, proceed immediately. If Tier 2 or 3, hold funds and insert entry into `maker_checker_approvals` queue.
   - *Phase 3 (Besu Burn):* Upon approval, call Hyperledger Besu relayer to execute on-chain burn.
   - *Phase 4 (Dispatch):* Send transaction payload to MPC Custody Service for threshold signing.
   - *Rollback:* If rejected or failed at any point prior to Besu burn, immediately release balance holds.
9. **Implement Hyperledger Besu Burn Relayer:** Configure `go-ethereum/ethclient` to submit EIP-1559 burn transactions to the `wBTC` / `weUSDT` token contracts on Besu. Subscribe to block headers to verify 1-block deterministic confirmation under QBFT consensus before signaling success to the coordinator.
10. **Build Bitcoin Transaction Builder & PSBT Assembler:** Implement Bitcoin transaction logic:
    - Query hot-vault UTXO set from PostgreSQL.
    - Run coin selection (Knapsack / Branch-and-Bound) to minimize change outputs and fee cost.
    - Build BIP 174 PSBT structure with user destination output and internal change output.
    - Fetch dynamic network fee rate via `estimatesmartfee` and attach to PSBT.
11. **Build Ethereum & Tron Transaction Builders:**
    - Ethereum: Encode ABI calldata for USDT `transfer(address,uint256)`, construct EIP-1559 transaction envelope with dynamic gas parameters.
    - Tron: Construct TRC-20 transfer protobuf message, calculate fee limits, and serialize for signing.
12. **Implement MPC Custody Service Client:** Construct gRPC client connecting to `services/mpc-tss-vault-service` (Prompt 237) over mTLS 1.3. Transmit raw transaction hashes / PSBT payloads and await 3-of-5 threshold co-signing completion.
13. **Build Public Node Broadcaster & Mempool Confirmation Watcher:**
    - Broadcast signed raw transactions to public nodes (Bitcoin Core `sendrawtransaction`, Ethereum `eth_sendRawTransaction`, TronGrid `wallet/broadcasttransaction`).
    - Run background worker polling transaction receipt status, mempool position, and block depth.
    - Confirm withdrawal as `COMPLETED` when confirmation thresholds are satisfied (BTC: 3 blocks, ETH: 12 blocks, Tron: 20 blocks).
    - Emit Kafka event `crypto.withdrawal.confirmed.v1` to trigger final double-entry ledger debit in Wallet Account Service (Prompt 203).
14. **Build Maker-Checker Review Queue API:** Implement admin gRPC endpoints allowing compliance and executive officers to view pending Tier 2 and Tier 3 withdrawals, inspect risk scoring factor breakdowns, and submit cryptographically signed Maker/Checker approval or rejection decisions.
15. **Implement Comprehensive Test Suite:** Author end-to-end integration tests using `testcontainers-go` for PostgreSQL and Redis, verifying 24-hour address lock enforcement, rejection of invalid Passkey assertions, zero double-spending under concurrent submissions, correct Besu collateral burning, and fail-safe rollback upon MPC timeout.

---

## Interfaces / Contracts

### Protobuf Definition (`crypto_withdrawal_service.proto`)

```protobuf
syntax = "proto3";

package growww.crypto.withdrawal.v1;

option go_package = "growww/crypto/withdrawal/v1;withdrawalv1";

// Service managing secure real-money crypto withdrawals and risk clearing.
service CryptoWithdrawalService {
  // Register a new external withdrawal address subject to a mandatory 24-hour cooldown.
  rpc RegisterWithdrawalAddress (RegisterWithdrawalAddressRequest) returns (RegisterWithdrawalAddressResponse);

  // List all registered withdrawal addresses and their activation countdown status.
  rpc ListWithdrawalAddresses (ListWithdrawalAddressesRequest) returns (ListWithdrawalAddressesResponse);

  // Submit a real-money crypto withdrawal request with Passkey and 2FA authentication.
  rpc InitiateCryptoWithdrawal (InitiateCryptoWithdrawalRequest) returns (InitiateCryptoWithdrawalResponse);

  // Query real-time status, risk tier, and on-chain confirmation progress of a withdrawal.
  rpc GetWithdrawalStatus (GetWithdrawalStatusRequest) returns (GetWithdrawalStatusResponse);

  // List historical withdrawals for an authenticated user.
  rpc ListWithdrawalHistory (ListWithdrawalHistoryRequest) returns (ListWithdrawalHistoryResponse);

  // Administrative Maker-Checker decision endpoint for Tier 2 and Tier 3 reviews.
  rpc MakerCheckerReviewWithdrawal (MakerCheckerReviewWithdrawalRequest) returns (MakerCheckerReviewWithdrawalResponse);
}

enum BlockchainNetwork {
  BLOCKCHAIN_NETWORK_UNSPECIFIED = 0;
  BLOCKCHAIN_NETWORK_BITCOIN = 1;
  BLOCKCHAIN_NETWORK_ETHEREUM = 2;
  BLOCKCHAIN_NETWORK_TRON = 3;
}

enum CryptoAsset {
  CRYPTO_ASSET_UNSPECIFIED = 0;
  CRYPTO_ASSET_BTC = 1;
  CRYPTO_ASSET_USDT = 2;
}

enum WithdrawalStatus {
  WITHDRAWAL_STATUS_UNSPECIFIED = 0;
  WITHDRAWAL_STATUS_PENDING_RISK_CHECK = 1;
  WITHDRAWAL_STATUS_PENDING_MAKER_REVIEW = 2;
  WITHDRAWAL_STATUS_PENDING_CHECKER_APPROVAL = 3;
  WITHDRAWAL_STATUS_HELD_IN_WALLET = 4;
  WITHDRAWAL_STATUS_BURNING_ON_BESU = 5;
  WITHDRAWAL_STATUS_AWAITING_MPC_SIGNING = 6;
  WITHDRAWAL_STATUS_BROADCASTED = 7;
  WITHDRAWAL_STATUS_CONFIRMED = 8;
  WITHDRAWAL_STATUS_REJECTED = 9;
  WITHDRAWAL_STATUS_FAILED = 10;
}

enum RiskTier {
  RISK_TIER_UNSPECIFIED = 0;
  RISK_TIER_1_AUTOMATED = 1;
  RISK_TIER_2_SINGLE_MAKER = 2;
  RISK_TIER_3_DUAL_MAKER_CHECKER = 3;
}

message RegisterWithdrawalAddressRequest {
  string user_id = 1;
  BlockchainNetwork network = 2;
  CryptoAsset asset = 3;
  string destination_address = 4;
  string address_label = 5;
  string totp_code = 6;
  string passkey_assertion_json = 7; // WebAuthn assertion payload
  string idempotency_key = 8;
}

message RegisterWithdrawalAddressResponse {
  string address_id = 1;
  string destination_address = 2;
  BlockchainNetwork network = 3;
  bool is_active = 4;                // False until 24-hour cooldown expires
  int64 registered_at_unix_ms = 5;
  int64 activates_at_unix_ms = 6;    // Exactly registered_at + 24 hours
  int64 remaining_lock_seconds = 7;
}

message ListWithdrawalAddressesRequest {
  string user_id = 1;
  BlockchainNetwork network_filter = 2;
}

message ListWithdrawalAddressesResponse {
  repeated RegisterWithdrawalAddressResponse addresses = 1;
}

message InitiateCryptoWithdrawalRequest {
  string user_id = 1;
  string address_id = 2;              // Must reference an active whitelisted address
  BlockchainNetwork network = 3;
  CryptoAsset asset = 4;
  string amount = 5;                  // Decimal string representation of amount
  string totp_code = 6;
  string passkey_assertion_json = 7;  // Cryptographic user authorization
  string client_ip = 8;
  string device_fingerprint = 9;
  string idempotency_key = 10;
}

message InitiateCryptoWithdrawalResponse {
  string withdrawal_id = 1;
  WithdrawalStatus status = 2;
  RiskTier assigned_risk_tier = 3;
  int32 risk_score = 4;               // 0 to 100
  string network_fee_estimated = 5;
  string platform_fee = 6;
  string net_payout_amount = 7;
  int64 created_at_unix_ms = 8;
}

message GetWithdrawalStatusRequest {
  string withdrawal_id = 1;
  string user_id = 2;
}

message GetWithdrawalStatusResponse {
  string withdrawal_id = 1;
  string user_id = 2;
  WithdrawalStatus status = 3;
  RiskTier risk_tier = 4;
  BlockchainNetwork network = 5;
  CryptoAsset asset = 6;
  string amount = 7;
  string destination_address = 8;
  string besu_burn_tx_hash = 9;      // Hash of collateral burn on Besu
  string public_chain_tx_hash = 10;   // Bitcoin/Ethereum/Tron transaction hash
  int32 confirmations = 11;
  int32 required_confirmations = 12;
  string failure_reason = 13;
  int64 created_at_unix_ms = 14;
  int64 updated_at_unix_ms = 15;
}

message ListWithdrawalHistoryRequest {
  string user_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListWithdrawalHistoryResponse {
  repeated GetWithdrawalStatusResponse withdrawals = 1;
  string next_page_token = 2;
}

message MakerCheckerReviewWithdrawalRequest {
  string withdrawal_id = 1;
  string reviewer_admin_id = 2;
  string reviewer_role = 3;          // "MAKER" or "CHECKER"
  bool decision_approve = 4;         // True to approve, False to reject
  string rejection_reason = 5;
  string admin_passkey_assertion = 6; // Cryptographic attestation by administrator
}

message MakerCheckerReviewWithdrawalResponse {
  string withdrawal_id = 1;
  WithdrawalStatus new_status = 2;
  string message = 3;
  int64 reviewed_at_unix_ms = 4;
}
```

---

### PostgreSQL Database Schema DDL

```sql
-- PostgreSQL 16+ DDL Schema for Crypto Withdrawal and Risk Clearing Engine

CREATE TYPE crypto_network_enum AS ENUM (
    'BITCOIN',
    'ETHEREUM',
    'TRON'
);

CREATE TYPE crypto_asset_enum AS ENUM (
    'BTC',
    'USDT'
);

CREATE TYPE crypto_withdrawal_status_enum AS ENUM (
    'PENDING_RISK_CHECK',
    'PENDING_MAKER_REVIEW',
    'PENDING_CHECKER_APPROVAL',
    'HELD_IN_WALLET',
    'BURNING_ON_BESU',
    'AWAITING_MPC_SIGNING',
    'BROADCASTED',
    'CONFIRMED',
    'REJECTED',
    'FAILED'
);

CREATE TYPE risk_tier_enum AS ENUM (
    'TIER_1_AUTOMATED',
    'TIER_2_SINGLE_MAKER',
    'TIER_3_DUAL_MAKER_CHECKER'
);

-- Table 1: Whitelisted Withdrawal Addresses with Mandatory 24-Hour Cooldown
CREATE TABLE whitelisted_withdrawal_addresses (
    address_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    network crypto_network_enum NOT NULL,
    asset crypto_asset_enum NOT NULL,
    destination_address VARCHAR(128) NOT NULL,
    address_label VARCHAR(64) NOT NULL,
    address_hash VARCHAR(64) NOT NULL, -- SHA-256 hash of (network || destination_address)
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activates_at TIMESTAMPTZ NOT NULL, -- registered_at + 24 hours
    passkey_credential_id VARCHAR(255) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_address_network UNIQUE (user_id, network, destination_address)
);

-- Table 2: Crypto Withdrawal Requests Master Table
CREATE TABLE crypto_withdrawal_requests (
    withdrawal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    address_id UUID NOT NULL REFERENCES whitelisted_withdrawal_addresses(address_id),
    network crypto_network_enum NOT NULL,
    asset crypto_asset_enum NOT NULL,
    amount NUMERIC(28, 8) NOT NULL CHECK (amount > 0),
    platform_fee NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000 CHECK (platform_fee >= 0),
    estimated_network_fee NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000 CHECK (estimated_network_fee >= 0),
    net_payout_amount NUMERIC(28, 8) NOT NULL CHECK (net_payout_amount > 0),
    destination_address VARCHAR(128) NOT NULL,
    status crypto_withdrawal_status_enum NOT NULL DEFAULT 'PENDING_RISK_CHECK',
    assigned_risk_tier risk_tier_enum NOT NULL,
    risk_score INT NOT NULL CHECK (risk_score >= 0 AND risk_score <= 100),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    wallet_hold_id UUID,
    besu_burn_tx_hash VARCHAR(66), -- 0x-prefixed 32-byte hash
    public_chain_tx_hash VARCHAR(128),
    public_chain_nonce BIGINT,
    confirmations_count INT NOT NULL DEFAULT 0,
    required_confirmations INT NOT NULL DEFAULT 3,
    failure_reason TEXT,
    client_ip VARCHAR(45) NOT NULL,
    device_fingerprint VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 3: Withdrawal Risk Scoring Factor Breakdowns
CREATE TABLE withdrawal_risk_evaluations (
    evaluation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    withdrawal_id UUID NOT NULL REFERENCES crypto_withdrawal_requests(withdrawal_id) ON DELETE CASCADE,
    velocity_24h_usd NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    velocity_score INT NOT NULL DEFAULT 0,
    ip_reputation_score INT NOT NULL DEFAULT 0,
    device_anomaly_score INT NOT NULL DEFAULT 0,
    credential_reset_flag BOOLEAN NOT NULL DEFAULT FALSE,
    sanction_screening_flag BOOLEAN NOT NULL DEFAULT FALSE,
    travel_rule_compliant BOOLEAN NOT NULL DEFAULT TRUE,
    total_computed_score INT NOT NULL,
    evaluation_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 4: Maker-Checker Review Queue and Decisions
CREATE TABLE maker_checker_approvals (
    approval_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    withdrawal_id UUID NOT NULL REFERENCES crypto_withdrawal_requests(withdrawal_id) ON DELETE CASCADE,
    required_tier risk_tier_enum NOT NULL,
    maker_admin_id UUID,
    maker_decision VARCHAR(32), -- 'APPROVED' or 'REJECTED'
    maker_notes TEXT,
    maker_passkey_sig TEXT,
    maker_decided_at TIMESTAMPTZ,
    checker_admin_id UUID,
    checker_decision VARCHAR(32), -- 'APPROVED' or 'REJECTED'
    checker_notes TEXT,
    checker_passkey_sig TEXT,
    checker_decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 5: On-Chain Broadcast Transactions Log
CREATE TABLE onchain_broadcast_transactions (
    broadcast_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    withdrawal_id UUID NOT NULL REFERENCES crypto_withdrawal_requests(withdrawal_id) ON DELETE RESTRICT,
    network crypto_network_enum NOT NULL,
    raw_tx_payload TEXT NOT NULL,
    tx_hash VARCHAR(128) NOT NULL UNIQUE,
    gas_or_miner_fee NUMERIC(28, 8) NOT NULL,
    broadcast_attempts INT NOT NULL DEFAULT 1,
    last_broadcast_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    block_height BIGINT,
    block_hash VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 6: Immutable Withdrawal Audit Events
CREATE TABLE withdrawal_audit_events (
    audit_id BIGSERIAL PRIMARY KEY,
    withdrawal_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor_id VARCHAR(128) NOT NULL,
    actor_role VARCHAR(64) NOT NULL,
    payload_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance and Query Optimization Indexes
CREATE INDEX idx_whitelisted_addr_user ON whitelisted_withdrawal_addresses (user_id, network) WHERE is_deleted = FALSE;
CREATE INDEX idx_whitelisted_addr_activation ON whitelisted_withdrawal_addresses (activates_at) WHERE is_active = FALSE;
CREATE INDEX idx_withdrawal_req_user ON crypto_withdrawal_requests (user_id, created_at DESC);
CREATE INDEX idx_withdrawal_req_status ON crypto_withdrawal_requests (status);
CREATE INDEX idx_withdrawal_req_tier ON crypto_withdrawal_requests (assigned_risk_tier) WHERE status IN ('PENDING_MAKER_REVIEW', 'PENDING_CHECKER_APPROVAL');
CREATE INDEX idx_onchain_broadcast_tx ON onchain_broadcast_transactions (tx_hash);
CREATE INDEX idx_withdrawal_audit_w_id ON withdrawal_audit_events (withdrawal_id, created_at ASC);
```

---

## Security & Compliance Notes

### Mandatory 24-Hour Address Lock & Cooldown
- Any new withdrawal address registered by a user is immediately subjected to an unalterable 24-hour cooldown enforced at both the application and database levels (`activates_at = NOW() + INTERVAL '24 hours'`).
- During this window, any withdrawal submitted targeting the address is rejected with code `ADDRESS_IN_COOLDOWN`.
- Upon address registration, multi-channel security notifications (SMS, push, email) are instantly dispatched to the user detailing the newly added destination address.
- The notification contains a high-priority "Freeze Account Immediately" action button allowing the user to freeze all account withdrawals with a single click if the address was added maliciously.

### Multi-Factor Dynamic Risk Scoring & Anti-Theft Velocity Limits
- Every withdrawal request is evaluated by the risk scoring engine before any financial lock occurs:
  - **Account Security Variance (+40 risk points):** If user password, email, or 2FA credentials were modified within the preceding 48 hours.
  - **Device & IP Variance (+25 risk points):** If the withdrawal originates from an unrecognized ASN, foreign country IP, or unregistered device fingerprint.
  - **24-Hour Velocity Thresholds (+30 risk points):** Cumulative rolling 24-hour withdrawal volume tracked in Redis. Threshold breaches automatically escalate low-risk withdrawals to Tier 2 or Tier 3.
  - **Sanctions & Darknet Cluster Hits (Immediate Rejection):** If destination address is flagged by Chainalysis / TRM Labs as associated with OFAC sanctions, mixer protocols (Tornado Cash), or darknet exploits.
- Tier Boundaries:
  - **Tier 1 (Automated):** Score < 30 and Notional <= $5,000 equivalent. Hot-wallet automated MPC dispatch.
  - **Tier 2 (Single Maker):** Score 30 to 70 or Notional $5,000 to $50,000. Compliance officer review required.
  - **Tier 3 (Dual Maker-Checker):** Score > 70 or Notional > $50,000. Dual human sign-off (Compliance Maker + Executive Checker) required.

### FIU-IND Reporting & Anti-Structuring Enforcement
- The service monitors withdrawal patterns for structuring tactics (e.g., submitting several $9,500 withdrawals within a short window to circumvent the $10,000 / INR 10,00,000 reporting threshold).
- Detected structuring attempts emit a high-priority event to Kafka topic `crypto.withdrawal.fiu_suspicious_alert.v1` for automatic Suspicious Transaction Report (STR) filing by Regulatory Reporting Service (Prompt 216 / Prompt 233).
- Enforces FATF Recommendation 16 (Travel Rule) for all transactions exceeding $1,000 / INR 50,000, requiring counterparty VASP identity attestation prior to execution.

### Cryptographic Attestation & Replay Protection
- Client submissions require a WebAuthn/FIDO2 Passkey assertion signature over an ephemeral server challenge binding `{user_id, destination_address, amount, network, timestamp}`.
- Prevents session hijacking and MITM attacks: an attacker with stolen session tokens cannot forge a withdrawal without access to the user's hardware biometric authenticator.
- All requests require a unique UUIDv4 `idempotency_key`, guarded in Redis with a 24-hour expiration window to completely eliminate duplicate execution risk.

---

## Acceptance Criteria

- [ ] New external withdrawal addresses are strictly locked for 24 hours; withdrawal attempts prior to expiration fail with code `ADDRESS_IN_COOLDOWN`.
- [ ] Address registration and withdrawal submission require valid WebAuthn/FIDO2 Passkey assertion signatures; tampered challenge hashes are rejected.
- [ ] Dynamic risk scoring accurately computes scores (0-100) and categorizes requests into Tier 1 (Automated), Tier 2 (Single Maker), or Tier 3 (Maker-Checker).
- [ ] Tier 1 withdrawals (score < 30 and <= $5,000) complete automated dispatch to MPC signing within <= 2.5 seconds end-to-end latency.
- [ ] Tier 2 and Tier 3 withdrawals remain in pending queue until approved by authorized administrators via `MakerCheckerReviewWithdrawal`.
- [ ] Two-phase atomic settlement coordinates with Wallet Account Service (Prompt 203) to place cryptographic balance holds prior to on-chain burn.
- [ ] Collateral burn on Hyperledger Besu executes successfully and confirms 1-block deterministic finality before native transaction broadcast.
- [ ] If on-chain Besu burn reverts or MPC signing times out, synthetic balance holds in Wallet Account Service are automatically released without orphan balances.
- [ ] Bitcoin transaction builder performs accurate coin selection, change output assignment, and dynamic fee estimation using live Bitcoin Core RPCs.
- [ ] Ethereum and Tron transaction builders format exact USDT transfer calldata and broadcast successfully to public testnet/mainnet nodes.
- [ ] Mempool watcher tracks block confirmations accurately (BTC: 3 blocks, ETH: 12 blocks, Tron: 20 blocks) and emits `crypto.withdrawal.confirmed.v1`.
- [ ] Structuring detection flags repeated sub-threshold withdrawals and emits `crypto.withdrawal.fiu_suspicious_alert.v1` to Kafka.
- [ ] All PostgreSQL database tables, foreign keys, and indexes compile cleanly against PostgreSQL 16+.

---

## Suggested Order / Dependencies

- **Prerequisites:**
  - `103_api_design_standards.md`: gRPC and RESTful API standards and error handling conventions.
  - `104_event_schema_and_kafka_topic_standards.md`: Protobuf event serialization and Kafka topic naming.
  - `111_domain_model_core_entities.md`: Core asset, wallet, and user entity definitions.
  - `112_idempotency_and_exactly_once_processing.md`: Distributed idempotency key standards.
  - `201_user_service.md`: WebAuthn credential retrieval and user session validation.
  - `202_kyc_aml_service.md`: User KYC tier limits and sanctions checks.
  - `203_wallet_account_service.md`: Double-entry ledgers, synthetic balance holds, and account debits.
  - `237_multichain_mpc_tss_vault_custody_service.md`: 3-of-5 threshold co-signing rounds for BTC/ETH/TRON.
- **Parallel Tasks:**
  - `211_notification_service.md`: Delivery of 24h address whitelist alerts and withdrawal notifications.
  - `216_regulatory_reporting_service.md`: Ingestion of FIU-IND suspicious activity alerts.
  - `217_admin_back_office_service.md`: Maker-Checker human approval portal interface.
  - `234_bitcoin_lightning_and_taproot_ingress_service.md`: Bitcoin Core full node peering and UTXO state management.
  - `235_evm_chainlink_ccip_multi_token_ingress_service.md`: Ethereum node connectivity and ERC-20 token handling.
- **Downstream Blockers:**
  - `509_flutter_order_placement_and_trading_interface.md`: Mobile Crypto Withdrawal flow, address book, and Passkey prompt.
  - `603_web_trading_dashboard_and_order_entry.md`: Web Crypto Withdrawal modal, whitelisting UI, and transaction tracker.
  - `703_aml_kyt_travel_rule_compliance_operations.md`: Back-office Travel Rule audit and SAR escalation procedures.
