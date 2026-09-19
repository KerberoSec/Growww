# 259 - ERC-4337 Account Abstraction Bundler & Gasless Paymaster Service (Go)

## Purpose
In enterprise digital asset platforms and tokenized securities exchanges operating on Ethereum Virtual Machine (EVM) ledgers, the requirement for end users to maintain, manage, and calculate native blockchain gas tokens creates severe user onboarding friction, custody vulnerabilities, and operational failures. Retail and institutional investors accustomed to seamless equity trading (such as fractional Indian equities under SEBI regulations or sovereign debt markets) should never be exposed to gas mechanics, nonce synchronization errors, or native token volatility. 

The **ERC-4337 Account Abstraction Bundler & Gasless Paymaster Service** (`services/aa-bundler-service`) eliminates crypto-native gas token friction across the platform by implementing a high-throughput, enterprise-grade ERC-4337 Bundler and Gas Sponsorship Paymaster on Hyperledger Besu. Instead of requiring users to fund and sign transactions with native gas tokens, user client devices (Flutter mobile and React web) generate cryptographically signed `UserOperations` executed via smart contract accounts. The Go bundler service validates incoming UserOperations in an off-chain simulation sandbox, enqueues them into an in-memory Redis mempool, verifies KYC whitelisting status, attaches EIP-712 gas sponsorship signatures from `GrowwwPaymaster.sol`, packs validated operations into atomic execution bundles, and dispatches them to the canonical ERC-4337 `EntryPoint` contract (`0x0000000071727De22E5E9d8BAf0edAc6f37da032`). All gas fees on Besu are subsidized through platform treasury fee allocations funded by the Treasury reserve cut of the platform flat 0.00% transaction fee (No fee at all) (with 25% allocated to the Core Settlement Guarantee Fund and 15% to the Investor Protection Fund).

## What You Are Building
A high-throughput, production-grade Go microservice (`services/aa-bundler-service`) implementing standard ERC-4337 JSON-RPC endpoints, off-chain pre-simulation validation, an intelligent Redis mempool, an EIP-712 gas sponsorship paymaster signer, and an atomic bundle packaging engine for Hyperledger Besu. Concrete components include:

- **ERC-4337 Standard JSON-RPC Server:** Implements standard Bundler RPC specifications (`eth_sendUserOperation`, `eth_estimateUserOperationGas`, `eth_getUserOperationByHash`, `eth_getUserOperationReceipt`, `eth_supportedEntryPoints`) handling high-volume concurrent submissions over HTTP and WebSocket.
- **Off-Chain Simulation & Pre-Verification Engine:** Executes pre-flight bytecode simulation (`EntryPoint.simulateValidation`) via Besu `eth_call` to enforce strict EIP-4337 execution rules, prevent illegal opcodes (`GASPRICE`, `BLOCKHASH`, non-isolated storage access), and reject reverting or malformed UserOperations before mempool entry.
- **KYC Whitelist Verification Hook:** Integrates with the KYC/AML Whitelist Registry (Prompt 305) and Redis caching layers to ensure that only fully KYC-verified investor accounts (domestic PAN/Aadhaar or foreign GIFT City passport verified) can receive gasless paymaster sponsorship.
- **EIP-712 Paymaster Sponsorship Signer:** Computes typed structured data hashes conforming to EIP-712 and signs sponsorship payloads using a dedicated secp256k1 private key managed via Hardware Security Modules (CloudHSM / HashiCorp Vault), populating `paymasterAndData` with validity timestamps and sponsorship authorization.
- **Treasury Gas Subsidy & Quota Monitor:** Tracks cumulative gas expenditure per smart account, enforcing daily/hourly sponsorship quotas, rate limits, and reconciling gas consumption against the platform treasury fee reserve per FeeController governance derived from the 0.00% (No fee at all) platform transaction fee.
- **Redis High-Throughput UserOp Mempool:** Distributed priority mempool storing, sorting, deduplicating, and indexing pending UserOperations by sender, nonces, and timestamp, with full support for UserOp gas replacement policies.
- **Bundle Packaging & Batch Dispatcher:** Asynchronous background engine that aggregates up to 30 mutually independent UserOperations into an atomic bundle, constructs the `handleOps` calldata, and broadcasts it to the Besu validator cluster via dedicated relayer signing keys.
- **QBFT Receipt Tracker & Settlement Synchronizer:** Subscribes to Besu new block headers, monitors bundle execution receipts, parses `UserOperationEvent` logs, handles partial bundle execution, updates PostgreSQL transaction tracking tables, and emits Kafka events to downstream settlement services.

## Scope Boundaries
- **In Scope:**
  - Standard ERC-4337 Bundler JSON-RPC endpoints (`eth_sendUserOperation`, `eth_estimateUserOperationGas`, `eth_getUserOperationByHash`, `eth_getUserOperationReceipt`, `eth_supportedEntryPoints`).
  - Off-chain state simulation and validation of `validateUserOp` execution against EntryPoint `0x0000000071727De22E5E9d8BAf0edAc6f37da032`.
  - Paymaster sponsorship qualification, rate limiting, and EIP-712 signature generation for `GrowwwPaymaster.sol`.
  - Enforcing investor KYC whitelisting checks prior to gas sponsorship authorization.
  - Multi-producer Redis UserOp mempool management with atomic Lua queuing, deduplication, and replacement logic.
  - Atomic bundle generation packing multiple UserOperations into single `handleOps` calls.
  - Integration with CloudHSM / HashiCorp Vault for paymaster and bundler relayer transaction signing.
  - Gas expenditure reconciliation against the platform treasury fee allocation per FeeController governance.
  - PostgreSQL transaction lifecycle tracking and Kafka event publication.
  - Prometheus metrics instrumentation and OpenTelemetry distributed tracing.
- **Out of Scope / Handled Elsewhere:**
  - Smart Contract Account (SCA) Factory and Account implementation contracts (handled in Prompt 303 / Prompt 305).
  - Smart contract implementation of `GrowwwPaymaster.sol` and `EntryPoint.sol` deployment (handled in Prompt 306 / Prompt 329).
  - Central Limit Order Book (CLOB) trade matching (handled in Prompt 205).
  - Off-chain double-entry fiat INR ledger accounting (handled in Prompt 203 / Prompt 210).
  - Multi-chain bridge collateral transfers (handled in Prompt 235 / Prompt 238).
  - UI key generation and client-side passkey/WebAuthn signature generation (handled in Prompt 501 / Prompt 601).

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Leverages Go high-throughput goroutine concurrency, efficient memory layout, low GC pause times, and mature EVM client libraries.
- **EVM Blockchain Client Library:** `github.com/ethereum/go-ethereum` (`ethclient`, `rpc`, `common`, `crypto`, `signer`, `abi`) configured for Hyperledger Besu JSON-RPC and WebSocket endpoints over mTLS.
- **In-Memory UserOp Mempool & Cache:** **Redis 7.2+ Cluster** using `go-redis/v9` with pre-loaded Lua scripts for atomic UserOperation insertion, deduplication, and priority sorting.
- **Relational Persistence:** **PostgreSQL 16+** with `jackc/pgx/v5` connection pool and `sqlc` compile-time type-safe SQL query generation.
- **Message Broker & Event Bus:** **Apache Kafka 3.7+** using `segmentio/kafka-go` for broadcasting UserOperation lifecycle events.
- **Cryptographic Key Management:** **AWS CloudHSM** or **HashiCorp Vault Transit Engine** via REST/mTLS for securing Paymaster EIP-712 signing keys and Bundler execution keys.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with mTLS for internal service RPCs alongside HTTP/JSON-RPC for standard ERC-4337 client interfaces.
- **Observability & Metrics:** **Prometheus** client (`prometheus/client_golang`) and **OpenTelemetry** Go SDK for tracing request simulation, mempool latency, and Besu broadcast pipeline.

## Backend / Infra Touchpoints
- **Hyperledger Besu Validator Cluster:**
  - Primary and backup JSON-RPC endpoints (e.g., `https://besu-node-01.internal:8545`) for state queries and simulation.
  - WebSocket endpoint (e.g., `wss://besu-node-01.internal:8546`) for new block header subscriptions and real-time receipt polling.
- **Redis 7.2 Key Schema:**
  - `bundler:mempool:ops`: Sorted Set of pending UserOperation hashes scored by `maxPriorityFeePerGas` and arrival sequence.
  - `bundler:mempool:data:{opHash}`: String containing serialized UserOperation JSON payload.
  - `bundler:user:nonce:{sender}:{key}`: Current highest observed nonce for the sender account key.
  - `bundler:paymaster:quota:{sender}`: Hash tracking cumulative sponsored gas units and remaining daily allocation.
  - `bundler:kyc:cache:{sender}`: Cached boolean status indicating active KYC compliance verification.
  - `bundler:lock:bundle`: Distributed Redlock key preventing concurrent bundle assembly races across bundler instances.
- **PostgreSQL 16 Tables:**
  - `user_operations`: Records incoming UserOperations, sender addresses, nonces, gas limits, paymaster status, execution status, and bundle associations.
  - `bundles`: Records packed bundles, relayer transaction hashes, Besu block numbers, gas used, and execution outcomes.
  - `paymaster_sponsorship_ledgers`: Logs every sponsored UserOperation, actual gas consumed, INR equivalent cost, and treasury allocation debit.
  - `bundler_account_quotas`: Configured rate limits, tier thresholds, and sponsorship allowances per smart account.
- **Kafka Topics:**
  - Consumes: `user.kyc_status_updated.v1` (from KYC Service, Prompt 202), `treasury.fee_allocated.v1` (from Fee Engine, Prompt 210).
  - Publishes: `aa.userop_received.v1`, `aa.userop_sponsored.v1`, `aa.bundle_submitted.v1`, `aa.userop_mined.v1`, `aa.userop_reverted.v1`.
- **Hardware Security Module (HSM):**
  - Key Alias `growww-paymaster-signer`: Dedicated secp256k1 key for signing EIP-712 paymaster authorizations.
  - Key Alias `growww-bundler-relayer`: Dedicated secp256k1 key for broadcasting raw transactions calling `EntryPoint.handleOps`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Network Topology:** Hyperledger Besu running Enterprise Istanbul/QBFT consensus with 2-second block intervals, deterministic single-block finality, and zero reorg risk past block confirmation.
- **Canonical EntryPoint Deployment:** Interacts directly with canonical ERC-4337 EntryPoint (`0x0000000071727De22E5E9d8BAf0edAc6f37da032`).
- **Paymaster Contract (`GrowwwPaymaster.sol`):**
  - Registered as a trusted paymaster on the EntryPoint.
  - Holds an on-chain native gas balance deposited via `entryPoint.depositTo{value: amount}(paymasterAddress)`.
  - Implements `validatePaymasterUserOp(UserOperation userOp, bytes32 userOpHash, uint256 maxCost)` returning context and validation status.
  - Verifies EIP-712 signature over:
    `PaymasterData(address sender, uint256 nonce, bytes initCode, bytes callData, uint256 callGasLimit, uint256 verificationGasLimit, uint256 preVerificationGas, uint256 maxFeePerGas, uint256 maxPriorityFeePerGas, uint48 validUntil, uint48 validAfter)`
  - Rejects sponsorship if the sender account is not active in `KYCWhitelistRegistry.sol`.
- **Bundle Execution Invocation:**
  - The bundler calls `EntryPoint.handleOps(UserOperation[] ops, address payable beneficiary)` where `beneficiary` is the bundler relayer address.
  - Bundler signs and broadcasts an EIP-1559 transaction (`DynamicFeeTx`) to the Besu network containing the batch payload.
- **Zero On-Chain PII Invariant:** All UserOperations, paymaster signatures, and bundle calldata contain exclusively pseudonymous contract addresses, nonces, method selectors, token identifiers (ISINs), and cryptographic proofs. Absolutely no investor Personally Identifiable Information (PAN, Aadhaar, email, phone) is ever encoded in calldata or logged on-chain.

## Account Abstraction & Gasless Paymaster Mechanics

### 1. ERC-4337 UserOperation Architecture & Besu Lifecycle
In standard EVM transactions, an Externally Owned Account (EOA) must directly initiate execution and pay gas. Under ERC-4337 on Besu:
1. The user creates and signs an off-chain `UserOperation` struct:
   $$\text{UserOp} = \{ \text{sender}, \text{nonce}, \text{initCode}, \text{callData}, \text{callGasLimit}, \text{verificationGasLimit}, \text{preVerificationGas}, \text{maxFeePerGas}, \text{maxPriorityFeePerGas}, \text{paymasterAndData}, \text{signature} \}$$
2. The user client dispatches the struct to the bundler via `eth_sendUserOperation`.
3. The bundler performs off-chain static and dynamic checks, verifying that the smart account exists or will be created via `initCode`.
4. Once verified, the operation is added to the Redis mempool.
5. The bundler packs multiple UserOperations into a single transaction calling `EntryPoint.handleOps(ops, beneficiary)`.
6. Inside the EntryPoint, execution proceeds in two deterministic loops:
   - **Verification Loop:** Calls `validateUserOp` on each sender account (and `validatePaymasterUserOp` on `GrowwwPaymaster.sol`). If any verification fails, that operation is discarded or reverts the transaction.
   - **Execution Loop:** Executes the actual `callData` on each sender account and computes the actual gas consumed.
   - **Refund/Settlement:** The EntryPoint deducts the gas cost from the Paymaster deposit and refunds any unused pre-allocated gas.

### 2. EIP-712 Gas Sponsorship Paymaster & Treasury Allocation
To achieve 100% gasless transactions for verified investors, `GrowwwPaymaster.sol` acts as a Verifying Paymaster. The bundler service operates the Paymaster Signing Service:
1. **Qualification Verification:** Upon receiving a request to sponsor a UserOp (or during `eth_sendUserOperation` when `paymasterAndData` specifies GrowwwPaymaster), the service verifies that the user is entitled to gasless trading.
2. **EIP-712 Structured Data Hashing:**
   The service builds the typed data structure following EIP-712:
   $$\text{DomainSeparator} = \text{Keccak-256}(\text{EIP712Domain}(\text{"GrowwwPaymaster"}, \text{"1.0.0"}, \text{ChainID}, \text{PaymasterAddress}))$$
   $$\text{StructHash} = \text{Keccak-256}(\text{UserOpSponsorship}(\text{sender}, \text{nonce}, \dots, \text{validUntil}, \text{validAfter}))$$
   $$\text{Digest} = \text{Keccak-256}(\text{"\x19\x01"} \parallel \text{DomainSeparator} \parallel \text{StructHash})$$
3. **HSM secp256k1 Signature:** The digest is signed by the HSM-secured Paymaster private key, producing $(R, S, V)$.
4. **Encoding `paymasterAndData`:**
   $$\text{paymasterAndData} = \text{PaymasterAddress} \ (20 \text{ bytes}) \parallel \text{validUntil} \ (6 \text{ bytes}) \parallel \text{validAfter} \ (6 \text{ bytes}) \parallel \text{Signature} \ (65 \text{ bytes})$$
5. **Platform Treasury Funding Model:**
   The entire gas liability is backed by Growww platform fee economics:
   $$\text{Platform Fee} = 0.01\% \ (1 \text{ bps on trade turnover})$$
   $$\text{Treasury Allocation} = \text{Platform Fee} \times 60\%$$
   $$\text{Core Settlement Guarantee Fund (SGF)} = \text{Platform Fee} \times 25\%$$
   $$\text{Investor Protection Fund (IPF)} = \text{Platform Fee} \times 15\%$$
   A portion of the Treasury reserve allocation is continuously swept into the `EntryPoint` contract via automated top-up scripts (`entryPoint.depositTo{value: budget}(GrowwwPaymaster)`), maintaining a minimum balance of 500,000,000 gas units at all times.

### 3. Off-Chain UserOp Simulation & Anti-DDoS Validation
To prevent Denial of Service (DDoS) attacks where invalid UserOperations waste bundler transaction fees, the Go bundler performs rigorous off-chain simulation before accepting any UserOp into the mempool:
1. **Simulation Invocation:** The bundler invokes `EntryPoint.simulateValidation(userOp)` via `eth_call` against a Besu full node.
2. **Return Parsing:** The EntryPoint reverts with custom error `ValidationResult(ReturnInfo, StakeInfo, StakeInfo, StakeInfo)` containing gas limits, timestamps, and validation status.
3. **Opcode & Storage Access Rule Enforcement:**
   - The UserOp verification phase cannot use banned opcodes: `ORIGIN`, `TIMESTAMP`, `BLOCKHASH`, `DIFFICULTY`, `GASPRICE`.
   - The verification code cannot access storage slots of other accounts, preventing front-running attacks that invalidate multiple operations simultaneously.
4. **Economic Feasibility Check:**
   $$\text{MaxPossibleCost} = (\text{callGasLimit} + \text{verificationGasLimit} \times 3 + \text{preVerificationGas}) \times \text{maxFeePerGas}$$
   The bundler validates that the Paymaster deposit on EntryPoint $\ge \sum \text{MaxPossibleCost}$ for all in-flight UserOps.

### 4. Redis Mempool Queuing & Bundle Packaging Engine
1. **Deduplication & Replacement:**
   - UserOps are keyed by `(sender, nonce)`.
   - If a UserOp with identical `(sender, nonce)` is submitted, it is rejected unless `maxPriorityFeePerGas` and `maxFeePerGas` are at least 10% higher than the pending operation (EIP-4337 replacement rule).
2. **Prioritization:**
   - UserOperations are queued in a Redis Sorted Set scored by effective priority fee and timestamp.
3. **Batch Generation:**
   - Every 500 milliseconds (or immediately when the mempool reaches 30 operations), the bundle worker executes a packaging cycle.
   - The worker pulls up to 30 valid UserOperations, ensuring no two operations in the same bundle share the same sender address (preventing internal nonce collision).
   - Constructs the aggregated call `EntryPoint.handleOps(bundleOps, relayerAddress)`.
   - Simulates the entire bundle via `eth_call` (`simulateHandleOp`).
   - Dispatches the raw transaction to Besu via `eth_sendRawTransaction`.

### 5. KYC Compliance & Rate Limiting Guardrails
1. **Mandatory KYC Verification:**
   - Before signing any Paymaster sponsorship payload, the bundler queries the on-chain `KYCWhitelistRegistry.sol` (cached in Redis with a 60-second TTL).
   - If the sender address is not registered or flagged as suspended, sponsorship is immediately denied with RPC error code `-32500` (Transaction Rejected: KYC Whitelist Required).
2. **Rate Limiting & Anti-Drain Quotas:**
   - Tier 1 (Retail Investors): Maximum 100 sponsored operations per 24-hour window, capped at 15,000,000 cumulative gas.
   - Tier 2 (Institutional / HNI Investors): Maximum 2,000 sponsored operations per 24-hour window, capped at 300,000,000 cumulative gas.
   - Any attempt to exceed these quotas triggers an administrative alert and requires either user fee contribution or treasury authorization.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Go Microservice Architecture:** Initialize Go 1.22+ module `services/aa-bundler-service` structured with domain boundaries: `cmd/bundler/`, `internal/rpc/`, `internal/mempool/`, `internal/simulator/`, `internal/paymaster/`, `internal/packager/`, `internal/besu/`, `internal/store/`, and `pkg/types/`.
2. **Author Protobuf Specifications:** Define internal gRPC contracts in `proto/growww/bundler/v1/bundler_service.proto` for inter-service communication (sponsorship requests, quota queries, bundle status). Generate Go stubs via `protoc-gen-go` and `protoc-gen-go-grpc`.
3. **Create PostgreSQL Database Schema Migrations:** Author database migration scripts in `services/aa-bundler-service/migrations/001_aa_bundler_schema.sql` creating tables for UserOperations, execution bundles, paymaster sponsorships, and account rate limits.
4. **Implement Core ERC-4337 Go Types & Hashers:** Define typed Go structs for `UserOperation`, `PackedUserOperation`, and `ValidationResult`. Implement Keccak-256 UserOp hash calculation strictly adhering to EIP-4337 EntryPoint specification.
5. **Implement Besu RPC Client & Contract Bindings:** Generate Go contract bindings for `EntryPoint.sol` (`0x0000000071727De22E5E9d8BAf0edAc6f37da032`) and `GrowwwPaymaster.sol` using `abigen`. Establish robust mTLS connection pools to Besu validator nodes.
6. **Implement Off-Chain Simulation Engine:** Build `Simulator` module executing `EntryPoint.simulateValidation` via `eth_call`. Parse returned custom revert bytes, extract execution timeframes (`validUntil`, `validAfter`), and verify storage access constraints.
7. **Implement KYC Whitelisting & Quota Guard:** Create pre-validation filter verifying sender address against `KYCWhitelistRegistry` (and local Redis cache). Enforce daily/hourly gas quotas per smart account before granting sponsorship.
8. **Implement EIP-712 Paymaster Sponsorship Signer:** Build `PaymasterSigner` module constructing EIP-712 domain separators and typed struct hashes. Connect to AWS CloudHSM or HashiCorp Vault Transit Engine to generate secp256k1 ECDSA signatures and assemble `paymasterAndData`.
9. **Implement Gas Estimation Engine:** Implement JSON-RPC method `eth_estimateUserOperationGas`, dynamically computing `verificationGasLimit`, `callGasLimit`, and `preVerificationGas` by running differential simulations against current Besu block state.
10. **Implement Redis UserOp Mempool:** Implement in-memory mempool with atomic Redis Lua scripts (`mempool_push.lua`, `mempool_pop.lua`, `mempool_replace.lua`) ensuring zero-gap ordering, nonce tracking, and atomic 10% fee-bump replacement.
11. **Implement Bundle Packaging & Execution Worker:** Build background worker that wakes up on 500ms intervals or mempool thresholds, selects up to 30 non-colliding UserOperations, performs batch simulation, generates raw `handleOps` transaction, and broadcasts it to Besu.
12. **Implement Block Receipt Monitor & Finality Tracker:** Build WebSocket subscription to Besu `newHeads`, inspecting transaction receipts for `UserOperationEvent`, updating PostgreSQL records, and evicting confirmed UserOperations from Redis.
13. **Implement Treasury Gas Reconciliation Engine:** Build asynchronous reconciliation worker calculating actual gas consumed by sponsored operations, converting gas to INR fiat equivalents, and auditing debits against the Treasury reserve fee ledger.
14. **Implement Standard ERC-4337 JSON-RPC Server:** Build high-performance HTTP and WebSocket JSON-RPC 2.0 handler exposing `eth_sendUserOperation`, `eth_estimateUserOperationGas`, `eth_getUserOperationByHash`, `eth_getUserOperationReceipt`, and `eth_supportedEntryPoints`.
15. **Implement Observability, Metrics & Stress Harness:** Instrument Prometheus metrics for mempool depth, simulation latency, bundle pack rate, and gas expenditure. Build end-to-end simulation test suite verifying high-concurrency throughput under simulated flash-trading conditions.

## Interfaces / Contracts

### 1. JSON-RPC ERC-4337 Bundler Interface (`cmd/bundler/rpc.go`)
```json
// Request: eth_sendUserOperation
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "eth_sendUserOperation",
  "params": [
    {
      "sender": "0x3B6C9B0B76378A529C52c9388B97a30364d92Fe1",
      "nonce": "0x1",
      "initCode": "0x",
      "callData": "0xb61d27f6000000000000000000000000...",
      "callGasLimit": "0x186a0",
      "verificationGasLimit": "0x249f0",
      "preVerificationGas": "0xc350",
      "maxFeePerGas": "0x3b9aca00",
      "maxPriorityFeePerGas": "0x3b9aca00",
      "paymasterAndData": "0xGrowwwPaymasterAddress0000673b0000000000000000...",
      "signature": "0x7a8e..."
    },
    "0x0000000071727De22E5E9d8BAf0edAc6f37da032"
  ]
}

// Response: Success
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x8f03b57e4b529a6bbfdf24df233c847e1d52ba441b8a514d791242fb835bdfd9"
}
```

### 2. Protobuf Service Contract (`proto/growww/bundler/v1/bundler_service.proto`)
```protobuf
syntax = "proto3";

package growww.bundler.v1;

option go_package = "github.com/growww/proto/gen/go/bundler/v1;bundlerv1";

message SponsorUserOpRequest {
  string sender_address = 1;
  uint64 nonce = 2;
  bytes init_code = 3;
  bytes call_data = 4;
  uint64 call_gas_limit = 5;
  uint64 verification_gas_limit = 6;
  uint64 pre_verification_gas = 7;
  uint64 max_fee_per_gas = 8;
  uint64 max_priority_fee_per_gas = 9;
  string isin = 10;
  uint64 trade_turnover_paise = 11;
}

message SponsorUserOpResponse {
  bool sponsored = 1;
  string paymaster_address = 2;
  bytes paymaster_and_data = 3;
  uint64 valid_until = 4;
  uint64 valid_after = 5;
  string rejection_reason = 6;
}

message UserOpStatusRequest {
  string user_op_hash = 1;
}

message UserOpStatusResponse {
  string user_op_hash = 1;
  string status = 2; // PENDING_MEMPOOL, BUNDLED, MINED_SUCCESS, REVERTED, DROPPED
  string transaction_hash = 3;
  uint64 block_number = 4;
  uint64 actual_gas_cost_wei = 5;
  uint64 actual_gas_used = 6;
  int64 mined_at_timestamp_ns = 7;
}

message BundlerPoolTelemetryResponse {
  uint32 pending_mempool_size = 1;
  uint64 total_ops_processed = 2;
  uint64 total_bundles_mined = 3;
  uint64 paymaster_deposit_remaining_wei = 4;
  uint64 treasury_allocated_gas_budget_wei = 5;
}

message EmptyRequest {}

service AABundlerService {
  rpc RequestSponsorship (SponsorUserOpRequest) returns (SponsorUserOpResponse);
  rpc GetUserOpStatus (UserOpStatusRequest) returns (UserOpStatusResponse);
  rpc GetBundlerTelemetry (EmptyRequest) returns (BundlerPoolTelemetryResponse);
}
```

### 3. Core Go Structs & Schemas (`pkg/types/userop.go`)
```go
package types

import (
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

// UserOperation represents the ERC-4337 v0.6 / v0.7 UserOperation payload
type UserOperation struct {
	Sender               common.Address `json:"sender"`
	Nonce                *big.Int       `json:"nonce"`
	InitCode             []byte         `json:"initCode"`
	CallData             []byte         `json:"callData"`
	CallGasLimit         *big.Int       `json:"callGasLimit"`
	VerificationGasLimit *big.Int       `json:"verificationGasLimit"`
	PreVerificationGas   *big.Int       `json:"preVerificationGas"`
	MaxFeePerGas         *big.Int       `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int       `json:"maxPriorityFeePerGas"`
	PaymasterAndData     []byte         `json:"paymasterAndData"`
	Signature            []byte         `json:"signature"`
}

// PaymasterSponsorshipClaim represents structured data signed under EIP-712
type PaymasterSponsorshipClaim struct {
	Sender               common.Address
	Nonce                *big.Int
	InitCodeHash         [32]byte
	CallDataHash         [32]byte
	CallGasLimit         *big.Int
	VerificationGasLimit *big.Int
	PreVerificationGas   *big.Int
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
	ValidUntil           uint64
	ValidAfter           uint64
}

// UserOpReceipt encapsulates on-chain confirmation details
type UserOpReceipt struct {
	UserOpHash    common.Hash    `json:"userOpHash"`
	Sender        common.Address `json:"sender"`
	Nonce         *big.Int       `json:"nonce"`
	Paymaster     common.Address `json:"paymaster"`
	ActualGasCost *big.Int       `json:"actualGasCost"`
	ActualGasUsed *big.Int       `json:"actualGasUsed"`
	Success       bool           `json:"success"`
	Reason        string         `json:"reason,omitempty"`
	TxHash        common.Hash    `json:"transactionHash"`
	BlockNumber   uint64         `json:"blockNumber"`
}
```

### 4. PostgreSQL Database Schema (`services/aa-bundler-service/migrations/001_aa_bundler_schema.sql`)
```sql
CREATE TABLE bundler_accounts (
    sender_address VARCHAR(42) PRIMARY KEY,
    user_id UUID NOT NULL,
    kyc_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    tier_level VARCHAR(32) NOT NULL DEFAULT 'RETAIL_TIER_1',
    daily_gas_quota_limit BIGINT NOT NULL DEFAULT 15000000,
    daily_gas_quota_consumed BIGINT NOT NULL DEFAULT 0,
    is_sponsored_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE execution_bundles (
    bundle_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relayer_address VARCHAR(42) NOT NULL,
    tx_hash VARCHAR(66) UNIQUE,
    block_number BIGINT,
    op_count INT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED', -- SUBMITTED, MINED, REVERTED
    total_gas_used BIGINT,
    effective_gas_price BIGINT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    mined_at TIMESTAMPTZ
);

CREATE TABLE user_operations (
    user_op_hash VARCHAR(66) PRIMARY KEY,
    sender_address VARCHAR(42) NOT NULL REFERENCES bundler_accounts(sender_address),
    nonce NUMERIC(78, 0) NOT NULL,
    call_data_selector VARCHAR(10),
    call_gas_limit BIGINT NOT NULL,
    verification_gas_limit BIGINT NOT NULL,
    pre_verification_gas BIGINT NOT NULL,
    max_fee_per_gas BIGINT NOT NULL,
    max_priority_fee_per_gas BIGINT NOT NULL,
    paymaster_address VARCHAR(42),
    is_sponsored BOOLEAN NOT NULL DEFAULT FALSE,
    bundle_id UUID REFERENCES execution_bundles(bundle_id),
    status VARCHAR(32) NOT NULL DEFAULT 'MEMPOOL_PENDING', -- MEMPOOL_PENDING, IN_BUNDLE, MINED_SUCCESS, MINED_REVERT, DROPPED
    actual_gas_used BIGINT,
    actual_gas_cost_wei NUMERIC(78, 0),
    revert_reason TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    mined_at TIMESTAMPTZ
);

CREATE TABLE paymaster_sponsorship_ledgers (
    sponsorship_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_op_hash VARCHAR(66) NOT NULL REFERENCES user_operations(user_op_hash),
    sender_address VARCHAR(42) NOT NULL REFERENCES bundler_accounts(sender_address),
    actual_gas_cost_wei NUMERIC(78, 0) NOT NULL,
    inr_equivalent_paise BIGINT NOT NULL,
    treasury_fee_allocated_paise BIGINT NOT NULL, -- Sourced from Treasury reserve allocation of 0.00% (Zero Fee) trade fee
    sponsored_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_userops_sender_nonce ON user_operations(sender_address, nonce);
CREATE INDEX idx_userops_status ON user_operations(status);
CREATE INDEX idx_userops_bundle ON user_operations(bundle_id);
CREATE INDEX idx_sponsorship_sender ON paymaster_sponsorship_ledgers(sender_address, sponsored_at DESC);
```

## Security & Compliance Notes
- **Paymaster Anti-Drain Rate Limits:** Every smart account is bounded by deterministic sliding-window gas quotas in Redis. In the event of algorithmic runaway, an investor account cannot consume more than its statutory daily allocation, neutralizing economic drain on the platform treasury.
- **Anti-DDoS Simulation Isolation:** The off-chain simulation engine guarantees that no UserOp entering the mempool accesses storage slots outside its own account perimeter or invokes ungrounded opcodes, preventing multi-op invalidation attacks that could exhaust bundler processing threads.
- **Mandatory KYC Whitelisting Hook:** Prior to attaching an EIP-712 gas sponsorship signature, the service cryptographically validates that the account is active in `KYCWhitelistRegistry.sol`. Non-compliant, blacklisted, or unverified entities are blocked from receiving gas subsidies.
- **Hardware Security Module Key Protection:** All Paymaster EIP-712 signing keys and Bundler Relayer execution keys are hosted inside FIPS 140-2 Level 3 certified CloudHSM or HashiCorp Vault enclaves. Private keys are never exposed in application memory or config files.
- **Zero-PII Compliance:** The bundler operates strictly over cryptographic public addresses, hashed identifiers, and raw bytecode payloads. No names, PAN numbers, tax IDs, or phone numbers enter the bundler data pipeline or transaction logs.
- **Platform Fee Treasury Allocation Guarantee:** All paymaster gas expenditures are fully backed by the Treasury reserve allocation originating from the platform invariant 0.00% transaction fee (No fee at all) (0.00% fee at launch; future fee parameters governed by FeeController.sol), preserving strict economic solvency without cross-subsidizing clearing guarantee funds.

## Acceptance Criteria
- [ ] Protobuf service contracts compile cleanly generating typed Go stubs with zero linting or generation warnings.
- [ ] Implements full standard ERC-4337 Bundler JSON-RPC methods (`eth_sendUserOperation`, `eth_estimateUserOperationGas`, `eth_getUserOperationByHash`, `eth_getUserOperationReceipt`, `eth_supportedEntryPoints`).
- [ ] Off-chain simulation correctly executes `EntryPoint.simulateValidation` and rejects malformed or reverting UserOperations before mempool entry.
- [ ] EIP-712 structured data signing generates valid cryptographic signatures matching `GrowwwPaymaster.sol` on-chain verification specifications.
- [ ] KYC whitelisting check blocks unverified accounts from receiving paymaster sponsorship with standard RPC error `-32500`.
- [ ] Redis mempool enforces atomic deduplication, monotonic nonce sequencing, and 10% fee replacement rules via Lua scripts.
- [ ] Bundle packaging engine aggregates up to 30 non-colliding UserOperations into a single atomic `handleOps` transaction.
- [ ] Single-block deterministic QBFT confirmation correctly parses `UserOperationEvent` logs and updates PostgreSQL state within 2 seconds.
- [ ] Sponsored gas costs are logged and reconciled accurately against the Treasury reserve allocation of the 0.00% (Zero Fee) platform fee.
- [ ] Prometheus metrics accurately export mempool depth, bundle dispatch rates, simulation errors, and paymaster gas spending.
- [ ] Strictly adheres to the 12 mandatory sections with zero raw application code, zero em dashes, and zero en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `202` (KYC/AML Service), Prompt `203` (Wallet & Account Service), Prompt `245` (Settlement Relayer Nonce Partitioning & Gas Escalator).
- **Parallel Work:** Prompt `305` (Smart Contract Transfer Compliance Hooks & KYC/AML Whitelist Registry), Prompt `306` (Atomic Delivery-versus-Payment Settlement Smart Contract), Prompt `329` (NBSE Delivery-versus-Payment Settlement & Automated Fee Collector Smart Contracts).
- **Subsequent Prompts Enabled:** Prompt `501` (Flutter Mobile Trading Interface), Prompt `509` (Flutter Order Placement Flow), Prompt `601` (Web Trading Terminal), Prompt `603` (Web Trading Dashboard).
