# 208 - Trade Settlement & DvP Orchestration Service (Go)

## Purpose
The Trade Settlement & DvP Orchestration Service is the critical settlement bridge linking off-chain trade matching with on-chain cryptographic settlement. In traditional financial markets, equity settlement is subject to counterparty risk and multi-day settlement cycles ($T+1$ or $T+2$). Growww eliminates counterparty settlement risk through instantaneous, atomic Delivery-versus-Payment (DvP) on a permissioned Hyperledger Besu blockchain.

This service guarantees that buyer cash debits in the off-chain INR double-entry ledger and seller fractional token transfers on the blockchain execute atomically: either both legs succeed completely, or both are rolled back. Every settled token maps 1:1 to physical shares held in custody by SEBI-registered depositories (NSDL/CDSL).

## What You Are Building
A mission-critical Go microservice (`services/trade-settlement-service`). Concrete deliverables include:
- Two-legged DvP settlement state machine managing the orchestration lifecycle (`MATCH_INGESTED`, `CASH_COMMITTED`, `CHAIN_SUBMITTED`, `CHAIN_CONFIRMED`, `SETTLED`, `COMPENSATING_ROLLBACK`).
- Web3 client integration with Hyperledger Besu JSON-RPC nodes via mTLS.
- Hardware Security Module (HSM) transaction signing client utilizing AWS CloudHSM / HashiCorp Vault Transit Engine to secure relayer private keys.
- On-chain transaction monitor listening for `SettlementDvP.sol` execution receipts and block finality confirmations under QBFT consensus.
- Compensation coordinator executing automated ledger rollbacks and hold releases if an on-chain transaction reverts.
- Kafka consumer for `engine.matches.v1` and publisher for `trade.settled.v1` and `settlement.failed.v1`.

## Scope Boundaries
- **In Scope:**
 - Ingestion of match records from the matching engine.
 - Coordination with Wallet Service (Prompt 203) for buyer cash debit and seller cash credit.
 - HSM-signed transaction submission to `SettlementDvP.sol` on Hyperledger Besu.
 - Transaction receipt tracking, nonce management, and block confirmation polling.
 - Emitting finalized settlement receipts to Portfolio Service (Prompt 209) and Fee Engine (Prompt 210).
- **Out of Scope / Handled Elsewhere:**
 - Matching engine limit order book execution (Prompt 205).
 - Writing the Solidity smart contract logic for `SettlementDvP.sol` (Prompt 306).
 - Daily physical demat depository share transfers with NSDL/CDSL (Prompt 213).
 - User wallet balance storage (Prompt 203).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+. Selected for its robust concurrency model, excellent EVM client libraries, low latency, and deterministic memory behavior during high-throughput financial orchestration.
- **EVM Blockchain Client:** `github.com/ethereum/go-ethereum/ethclient` configured for Hyperledger Besu JSON-RPC endpoints with custom raw transaction signers.
- **Cryptographic Key Custody:** AWS CloudHSM / HashiCorp Vault Transit Engine via PKCS#11 / REST for secp256k1 ECDSA signing without in-memory private keys.
- **Database & Storage:** PostgreSQL 16+ using `pgx/v5` and `sqlc` for settlement batch and audit records.
- **Streaming & Messaging:** `segmentio/kafka-go` with manual commit acknowledging only after state persistence.

## Backend / Infra Touchpoints
- **Hyperledger Besu Cluster:** Node RPC endpoints over mTLS (e.g. `https://besu-validator-1.internal:8545`).
- **PostgreSQL 16:** Tables `settlement_batches`, `settlement_records`, `chain_transactions`.
- **AWS CloudHSM / Vault:** Manages relayer account keys (`0xRelayerDvP...`).
- **Apache Kafka:** Consumes `engine.matches.v1`; publishes `trade.settled.v1`, `settlement.failed.v1`.
- **Wallet Service (Prompt 203):** Calls `CommitHold` via gRPC.
- **Portfolio Service (Prompt 209):** Ingests `trade.settled.v1` to update fractional tax lots.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network & Consensus:** Hyperledger Besu running Istanbul/QBFT Byzantine Fault Tolerant consensus with 2-second deterministic block finality and zero transaction gas fees (free gas network model with authorized relayers).
- **Core Smart Contracts Interfaced:**
 - `SettlementDvP.sol`: Calls `executeDvPTrade(bytes32 tradeId, address buyer, address seller, address tokenContract, uint256 tokenUnitsRaw, uint256 inrAmountPaise, bytes relayerSig)`.
 - `DigitalSecurityToken.sol`: ERC-3643 compliant permissioned security token representing 1:1 custody-backed equity.
- **Cryptographic Verification:** Every trade settlement records the transaction hash (`tx_hash`), block number, and Merkle receipt proof in the PostgreSQL settlement record.
- **Zero On-Chain PII:** The on-chain call references strictly pseudonymous Ethereum addresses (`buyer_address`, `seller_address`) and the asset contract address; zero names, PANs, or banking details exist on the ledger.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/trade-settlement-service` with strict linter rules and standard layout (`cmd/`, `internal/dvp/`, `internal/chain/`, `internal/hsm/`).
2. **Compile Smart Contract Go Bindings:** Use `abigen` on `SettlementDvP.sol` and `DigitalSecurityToken.sol` ABIs to generate type-safe Go client wrappers.
3. **Design Database Schema:** Write migrations for `settlement_batches`, `settlement_records`, and `chain_transactions`.
4. **Implement CloudHSM / Vault Signer:** Build custom `bind.SignerFn` in Go that dispatches unsigned transaction RLP hashes to AWS CloudHSM / Vault Transit Engine for ECDSA secp256k1 signing.
5. **Implement Nonce Management Engine:** Build distributed Redis/in-memory nonce allocator to prevent transaction replacement collisions under high concurrency.
6. **Implement DvP State Machine:** Build state manager handling transitions: `MATCH_RECEIVED` $\rightarrow$ `CASH_HOLD_COMMITTED` $\rightarrow$ `CHAIN_TX_SUBMITTED` $\rightarrow$ `CHAIN_CONFIRMED` $\rightarrow$ `SETTLED`.
7. **Build Cash Leg Coordinator:** Call `WalletService.CommitHold` for buyer INR debit and seller INR credit.
8. **Build Token Leg Dispatcher:** Call `SettlementDvP.executeDvPTrade` via Besu RPC client using HSM-signed transaction.
9. **Implement Block Confirmation Monitor:** Poll or subscribe via WebSocket to Besu block headers to confirm transaction inclusion with $\ge 1$ QBFT block confirmation.
10. **Implement Automated Compensation Engine:** If the blockchain transaction fails or reverts (e.g. compliance gate block), trigger automated compensation: reverse cash journal postings in Wallet Service, release held shares, and flag trade for operator review.
11. **Implement Kafka Match Consumer:** Consume `engine.matches.v1` with batch buffering (up to 50 matches or 50ms) to support batch settlement submissions.
12. **Implement Kafka Settlement Publisher:** Emit `trade.settled.v1` upon verified block confirmation containing on-chain `tx_hash` and `block_number`.
13. **Configure Telemetry & Alerting:** Instrument Prometheus metrics tracking settlement duration, chain submission latency, revert counts, and gas/nonce metrics.
14. **Perform End-to-End Resilience Testing:** Run integration test suite with Besu testnet and Postgres containers simulating node disconnects, HSM timeouts, and smart contract reverts.

## Interfaces / Contracts

### Protobuf Definition (`settlement_service.proto`)
```protobuf
syntax = "proto3";

package growww.settlement.v1;

option go_package = "growww/settlement/v1;settlementv1";

service SettlementService {
  rpc GetSettlementStatus (GetSettlementStatusRequest) returns (GetSettlementStatusResponse);
  rpc RetryFailedSettlement (RetryFailedSettlementRequest) returns (RetryFailedSettlementResponse);
}

enum SettlementStatus {
  SETTLEMENT_STATUS_UNSPECIFIED = 0;
  SETTLEMENT_STATUS_PROCESSING = 1;
  SETTLEMENT_STATUS_SETTLED = 2;
  SETTLEMENT_STATUS_FAILED = 3;
  SETTLEMENT_STATUS_COMPENSATED = 4;
}

message GetSettlementStatusRequest {
  string trade_id = 1;
}

message GetSettlementStatusResponse {
  string trade_id = 1;
  SettlementStatus status = 2;
  string buyer_user_id = 3;
  string seller_user_id = 4;
  string isin = 5;
  string token_amount = 6;
  string inr_amount = 7;
  string tx_hash = 8; // On-chain Hyperledger Besu transaction hash
  int64 block_number = 9;
  int64 settled_at_unix = 10;
  string error_message = 11;
}

message RetryFailedSettlementRequest {
  string trade_id = 1;
  string operator_reason = 2;
}

message RetryFailedSettlementResponse {
  bool success = 1;
  string new_settlement_status = 2;
}
```

### Smart Contract Solidity Interface Sketch (`ISettlementDvP.sol`)
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface ISettlementDvP {
    event DvPExecuted(
        bytes32 indexed tradeId,
        address indexed buyer,
        address indexed seller,
        address tokenAddress,
        uint256 tokenUnitsRaw,
        uint256 inrAmountPaise,
        uint256 timestamp
    );

    function executeDvPTrade(
        bytes32 tradeId,
        address buyer,
        address seller,
        address tokenAddress,
        uint256 tokenUnitsRaw,
        uint256 inrAmountPaise,
        bytes calldata relayerSignature
    ) external returns (bool success);
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE settlement_status_enum AS ENUM ('MATCH_RECEIVED', 'CASH_COMMITTED', 'CHAIN_SUBMITTED', 'CHAIN_CONFIRMED', 'SETTLED', 'FAILED', 'COMPENSATED');

CREATE TABLE settlement_records (
    settlement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_id UUID NOT NULL UNIQUE,
    buyer_user_id UUID NOT NULL,
    seller_user_id UUID NOT NULL,
    buyer_ledger_address VARCHAR(42) NOT NULL,
    seller_ledger_address VARCHAR(42) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    token_contract_address VARCHAR(42) NOT NULL,
    token_quantity_raw NUMERIC(18, 6) NOT NULL,
    gross_amount_inr NUMERIC(18, 4) NOT NULL,
    status settlement_status_enum NOT NULL DEFAULT 'MATCH_RECEIVED',
    tx_hash VARCHAR(66), -- 0x + 64 hex characters
    block_number BIGINT,
    revert_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ
);

CREATE TABLE chain_transactions (
    tx_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id UUID NOT NULL REFERENCES settlement_records(settlement_id),
    nonce BIGINT NOT NULL,
    relayer_address VARCHAR(42) NOT NULL,
    tx_hash VARCHAR(66) NOT NULL UNIQUE,
    raw_payload BYTEA NOT NULL,
    gas_used BIGINT,
    status VARCHAR(20) NOT NULL, -- PENDING, MINED, REVERTED
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);

CREATE INDEX idx_settlement_status ON settlement_records(status);
CREATE INDEX idx_settlement_tx_hash ON settlement_records(tx_hash);
```

## Security & Compliance Notes
- **Strict DvP Invariant:** Zero counterparty risk: cash leg and token transfer leg are guaranteed to be atomic.
- **HSM Key Custody:** Relayer private keys never reside in application memory or environment variables; all signing requests are dispatched to FIPS 140-2 Level 3 HSM / Vault.
- **Auditability:** Every settlement record stores the immutable blockchain transaction hash, enabling external auditors and SEBI regulators to verify settlement validity on-chain.
- **Zero On-Chain PII:** The blockchain transaction payload contains strictly pseudonymous addresses and transaction amounts.

## Acceptance Criteria
- [ ] Settlement service successfully ingests trade matches and executes two-legged DvP workflow.
- [ ] Relayer transactions are signed via HSM and confirmed on Hyperledger Besu under QBFT consensus within 3 seconds.
- [ ] Successful settlements persist `tx_hash` and `block_number` and emit `trade.settled.v1`.
- [ ] If on-chain transaction reverts, automated compensation releases holds and returns cash/shares cleanly.
- [ ] Nonce manager prevents transaction collisions under concurrent multi-worker execution.
- [ ] Zero PII is submitted to the blockchain network.

## Suggested Order / Dependencies
- **Prerequisites:** 203 (Wallet Service), 205 (Matching Engine), 303 (Token Issuance), 305 (Compliance Hooks), 306 (DvP Contract), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 209 (Portfolio Service), 210 (Fee Engine).
- **Downstream Blockers:** 215 (Reconciliation Service), 512 (Flutter Trade History).
