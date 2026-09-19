# 234 - Bitcoin (BTC), Lightning Network & Taproot Collateral Ingress Service (Go / Bitcoin Core / LND / PSBT / HSM)

## Purpose
Enables international and institutional investors within the GIFT City IFSCA regulatory perimeter to deposit, collateralize, and withdraw Bitcoin (BTC) seamlessly using both on-chain Bitcoin transactions (Native SegWit and Taproot) and instant Layer-2 Lightning Network payment channels.

To eliminate counterparty credit risk and operate 24 hours a day, 7 days a week, this service provides an institutional-grade, non-custodial-level bridge between the native Bitcoin network and Growww's permissioned Hyperledger Besu trading and settlement infrastructure. It automatically validates incoming Bitcoin UTXOs and Lightning Network payments, verifies cryptographic proofs, enforces multi-tier vault security policies, and mints 1:1 asset-backed synthetic Bitcoin collateral (`sBTC`) on Hyperledger Besu to power real-time margin trading and instantaneous DvP settlement.

## What You Are Building
A mission-critical, high-concurrency Go microservice (`services/btc-lightning-ingress`) operating in a hardened zero-trust network environment:
- **Bitcoin Core RPC & Taproot/SegWit Ingress Engine:** Direct interface with clustered Bitcoin Core full nodes over JSON-RPC (with cookie authentication and mTLS). Derives Hierarchical Deterministic (HD) deposit addresses compliant with BIP 84 (P2WPKH `bc1q...`) and BIP 86 (P2TR Taproot `bc1p...` with Schnorr signatures and MAST / Tapscript capability).
- **Lightning Network (LND / Core Lightning) Gateway:** Clustered gRPC client managing high-throughput Layer-2 payment channels, dynamic BOLT 11 invoice generation, BOLT 12 offer handling, invoice settlement monitoring, circular channel rebalancing, and Submarine Swaps.
- **Block Confirmation & Mempool Watcher:** Real-time ZeroMQ/RPC block and transaction ingestion pipeline tracking unconfirmed transactions, mempool depth, and block confirmation progression (enforcing strict 6-block finality for on-chain Bitcoin deposits and instant preimage cryptographic verification for Lightning).
- **Double-Spend & Replace-By-Fee (RBF) Defense Engine:** Active mempool monitor detecting BIP 125 Opt-in RBF replacements, Child-Pays-For-Parent (CPFP) fee escalations, transaction eviction, and blockchain reorganization (reorg) events up to 100 blocks deep.
- **Partially Signed Bitcoin Transaction (PSBT) & HSM Signing Coordinator:** Standardized BIP 174 / BIP 370 PSBT workflow engine that performs coin selection (Branch-and-Bound / Knapsack), builds withdrawal and sweep transactions, routes unsigned PSBT payloads to FIPS 140-2 Level 3 Hardware Security Modules (CloudHSM via PKCS#11) for cryptographic signing, and broadcasts finalized transactions.
- **Multi-Tier Hot/Warm/Cold Vault Sweeper:** Automated liquidity balancer enforcing balance ceilings across hot operational wallets, warm HSM-controlled vaults, and 3-of-5 cold multi-signature storage vaults (BIP 342 Tapscript / MuSig2).
- **Hyperledger Besu Synthetic Collateral Relayer:** Bidirectional bridge that mints 1:1 asset-backed synthetic BTC collateral tokens (`sBTC`) on Hyperledger Besu upon verified deposit, and burns synthetic collateral upon authorized withdrawal requests.

## Scope Boundaries
- **In Scope:**
  - Bitcoin Core v26+ JSON-RPC connection pooling, block header synchronization, and UTXO state management.
  - HD wallet key derivation (BIP 32 / BIP 44 / BIP 84 / BIP 86) for user deposit addresses.
  - LND v0.17+ / Core Lightning (CLN) gRPC integration for BOLT 11 invoices, BOLT 12 offers, and keysend payments.
  - Mempool monitoring via ZeroMQ (`hashtx`, `rawtx`, `hashblock`, `rawblock`) and double-spend detection.
  - BIP 125 Replace-By-Fee (RBF) tracking and conflict detection against internal pending deposits.
  - BIP 174 / BIP 370 PSBT creation, validation, fee rate estimation, and coin selection.
  - PKCS#11 HSM integration for hot/warm key signing and Schnorr/ECDSA signature verification.
  - Automated liquidity sweeps between Hot, Warm, and Cold vaults based on configurable thresholds.
  - Submarine Swaps (Loop In / Loop Out) to maintain balanced Lightning channel liquidity.
  - Minting and burning of synthetic BTC collateral on Hyperledger Besu via smart contract interfaces.
- **Out of Scope / Handled Elsewhere:**
  - Order matching and real-time trade risk calculation (handled in Prompt 205 and Prompt 206).
  - Fiat INR/USD currency conversion and FX rate locking (handled in Prompt 214).
  - User identity onboarding, KYC verification, and AML sanctions screening (handled in Prompt 202 and Prompt 703).
  - Physical cold storage air-gapped ritual ceremony orchestration (handled in Prompt 702 and Prompt 707).

## Technology to Use
- **Language & Runtime:** **Go 1.22+** with `btcsuite/btcd` (RPC client, wire format, txscript, btcutil), `lightningnetwork/lnd` API clients, and `google.golang.org/grpc`.
- **Node Infrastructure:**
  - Clustered Bitcoin Core 26.x+ full archival nodes (`txindex=1`, `blockfilterindex=1`, `coinstatsindex=1`, `prune=0`).
  - Clustered LND v0.17+ and Core Lightning nodes backed by PostgreSQL replicated channel state.
- **Cryptographic & Key Management:**
  - FIPS 140-2 Level 3 Hardware Security Module (AWS CloudHSM / Azure Dedicated HSM) accessed via PKCS#11 (`miekg/pkcs11`).
  - BIP 340 (Schnorr Signatures for secp256k1), BIP 341 (Taproot Segregated Witness), BIP 342 (Tapscript Validation), BIP 174 / BIP 370 (PSBT).
- **Database & State Store:**
  - **PostgreSQL 16+** with `pgx/v5` for relational UTXO ledgers, invoice states, sweep transactions, and audit trails.
  - **Redis 7.2** for sub-millisecond mempool tracking, active deposit address sets, and distributed distributed locks.
- **Event Streaming:** **Apache Kafka** with strictly typed Protobuf / JSON schemas for cross-service event notifications.
- **Ledger Integration:** `go-ethereum/ethclient` for connecting to Hyperledger Besu permissioned QBFT network.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `btc_deposit_addresses`, `btc_utxos`, `btc_onchain_transactions`, `lightning_invoices`, `lightning_payments`, `vault_sweep_batches`, `psbt_signing_sessions`, `synthetic_collateral_bridges`.
- **Redis 7.2 Keys:** `btc:address:watch:{address}`, `btc:mempool:tx:{txid}`, `lightning:invoice:{payment_hash}`, `vault:balance:hot_tier`.
- **Kafka Topics Consumed:**
  - `wallet.collateral.deposit_intent.v1`
  - `wallet.collateral.withdrawal_intent.v1`
  - `compliance.address_screening.cleared.v1`
- **Kafka Topics Produced:**
  - `btc.deposit.detected.v1`
  - `btc.deposit.confirmed.v1`
  - `btc.deposit.reorg_invalidated.v1`
  - `lightning.invoice.settled.v1`
  - `btc.withdrawal.broadcasted.v1`
  - `btc.collateral.mint_dispatched.v1`
  - `btc.collateral.burn_dispatched.v1`
- **Upstream Microservices:**
  - `services/wallet-account-service` (Prompt 203): Account balance updates and double-entry general ledger posting.
  - `services/foreign-investor-funding-service` (Prompt 214): International investor capital accounting.
  - `services/kyc-aml-service` (Prompt 202 & Prompt 703): On-chain address screening, Travel Rule compliance, and sanctions validation.
- **Downstream Microservices:**
  - `services/risk-engine` (Prompt 206): Real-time collateral margin credit for active trading.
  - `services/reconciliation-service` (Prompt 215): 3-way continuous reconciliation (Bitcoin UTXO + Lightning balance vs Besu synthetic tokens vs Internal ledger).
  - `services/audit-log-service` (Prompt 218): Immutable recording of all key generation, signing, and sweep operations.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Synthetic Collateral Contract (`SyntheticBitcoinCollateral.sol`):** A compliance-aware ERC-20 token (`sBTC`) representing 1:1 Bitcoin backing on Hyperledger Besu with 8 decimal places (1 unit = 1 Satoshi = $10^{-8}$ BTC).
- **Atomic Minting Workflow:**
  - Upon 6 on-chain Bitcoin confirmations or instant Lightning invoice preimage settlement, the service compiles a cryptographic mint payload: `(recipient_besu_address, amount_satoshis, btc_txid_or_payment_hash, deposit_vout, block_height, proof_digest)`.
  - The mint payload is submitted to `BitcoinCollateralVault.sol` using an authorized operator key from the HSM.
  - `BitcoinCollateralVault.sol` verifies that the `btc_txid_or_payment_hash` has not been processed previously (replay protection) and mints `sBTC` tokens to the investor's designated trading account.
- **Atomic Burning & Withdrawal Workflow:**
  - When an investor initiates a withdrawal of Bitcoin collateral, the trading platform locks and burns the equivalent `sBTC` balance on Besu via `BitcoinCollateralVault.burnCollateral()`.
  - The smart contract emits a `CollateralBurnedForWithdrawal(address indexed user, uint256 amountSatoshis, bytes32 indexed withdrawalId, string btcDestinationAddress)` event.
  - The Go service ingests the burn event, verifies its finality on Besu, constructs a Bitcoin PSBT or Lightning payment, executes HSM signing, and broadcasts the payout.
- **Zero PII Standard:** Only cryptographic hashes (`investor_account_hash`, `btc_txid`, `payment_hash`, `vault_address`) are written to the Besu ledger. No real names, IP addresses, or KYC references exist on-chain.
- **Proof-of-Reserve Attestation (`ProofOfReserveRegistry.sol`):** Every 10 minutes, the service generates a Merkle tree of all current UTXOs and Lightning channel state balances, signs the root digest using an HSM key, and commits the state to `ProofOfReserveRegistry.sol` (Prompt 308).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Layout:** Initialize Go module `services/btc-lightning-ingress` with clean architectural packages: `rpc/`, `lightning/`, `watcher/`, `rbf/`, `sweeper/`, `psbt/`, `signer/`, `besu/`, and `storage/`.
2. **Define Protobuf Contracts:** Author `proto/growww/btc/v1/btc_ingress.proto` defining gRPC methods for address generation, invoice creation, withdrawal dispatch, sweep execution, and channel health queries.
3. **Generate gRPC Stubs:** Compile Protobuf definitions with `protoc-gen-go` and `protoc-gen-go-grpc`.
4. **Configure PostgreSQL Schema Migrations:** Implement database tables for HD deposit address pools, UTXO state tracking, on-chain transaction histories, Lightning invoices, payments, PSBT sessions, and vault sweeps.
5. **Implement Bitcoin Core RPC Client Pool:** Build resilient JSON-RPC connection pool supporting failover between primary and replica Bitcoin Core nodes, with automatic authentication header generation and keep-alive health checks.
6. **Implement LND / Core Lightning gRPC Engine:** Build bidirectional gRPC connection manager for Lightning nodes with TLS certificate pinning and macaroon authentication. Implement methods for creating BOLT 11 invoices, monitoring invoice settlements via streaming subscriptions, and routing outbound payments.
7. **Implement HD Address Derivation Engine:** Build BIP 32 / BIP 44 / BIP 84 / BIP 86 address generator deriving Native SegWit (P2WPKH) and Taproot (P2TR) addresses from master extended public keys (`xpub`/`zpub`/`tpub`). Store derived addresses in Redis and PostgreSQL with investor account mappings.
8. **Implement Block Confirmation & Mempool Watcher:**
   - Ingest ZeroMQ feeds (`rawtx`, `hashtx`, `rawblock`, `hashblock`) from Bitcoin Core.
   - Detect unconfirmed transactions paying to tracked deposit addresses and record them with `CONFIRMATIONS = 0`.
   - Poll block tip headers, calculate confirmation depth for pending deposits, and trigger collateral minting when confirmation depth reaches exactly 6.
   - Detect blockchain reorganizations by comparing block hashes; invalidate and roll back collateral states if a confirmed deposit is dropped or replaced during a reorg.
9. **Implement Double-Spend & RBF Protection Engine:**
   - Parse transaction inputs for `nSequence < 0xFFFFFFFE` to flag BIP 125 Opt-in RBF transactions.
   - Monitor mempool replacements: if a replacement transaction spends the same inputs but redirects outputs away from the platform address, immediately mark deposit as `CONFLICT_DETECTED` and notify the risk engine.
   - Enforce zero credit for 0-confirmation transactions under all circumstances.
10. **Implement Coin Selection & PSBT Construction Engine:**
    - Implement Branch-and-Bound (BnB) coin selection for exact change-free matches and Knapsack fallback for variable amounts.
    - Build unsigned BIP 174 / BIP 370 PSBTs specifying UTXO inputs, witness scripts, output addresses, and dynamic fee rates obtained from `estimatesmartfee`.
11. **Implement HSM PKCS#11 PSBT Signing Provider:**
    - Interface with CloudHSM via PKCS#11 to load private keys designated for hot and warm vault UTXOs.
    - Compute sighash digests according to BIP 143 (SegWit) and BIP 341 (Taproot Schnorr), submit digests to HSM for signing, attach witness data to PSBT inputs, and finalize the raw transaction.
12. **Implement Multi-Tier Vault Balance Sweeper:**
    - Monitor hot wallet liquid balance against configured minimum and maximum limits.
    - If hot balance exceeds the upper threshold, automatically construct a sweep PSBT transferring surplus funds to the warm HSM vault or 3-of-5 cold multi-sig vault.
    - If hot balance drops below the lower threshold, trigger a prioritized warm-to-hot rebalancing request.
13. **Implement Submarine Swap Liquidity Manager:** Integrate automated Loop In and Loop Out swaps with external Lightning liquidity providers to convert on-chain UTXOs into Lightning channel liquidity and vice versa.
14. **Implement Hyperledger Besu Synthetic Collateral Relayer:**
    - Ingest confirmed deposit events (6 confirmations on Bitcoin or settled Lightning invoice), encode transaction receipts, and call `BitcoinCollateralVault.mintCollateral()` on Besu.
    - Subscribe to Besu `CollateralBurnedForWithdrawal` events, validate burn transaction finality, and dispatch Bitcoin withdrawal processing.
15. **Establish Observability & Comprehensive Test Suites:**
    - Export Prometheus metrics (`btc_mempool_rbf_events_total`, `btc_block_confirmation_latency_seconds`, `lightning_channel_local_balance_sat`, `vault_sweep_duration_seconds`).
    - Build integration test suites using Dockerized `bitcoind -regtest` and dual `lnd` nodes simulating deposits, withdrawals, RBF fee bumps, reorgs, and multi-sig sweeps.

## Mathematical / Algorithmic Formulation

### 1. Coin Selection Optimization (Branch-and-Bound with Knapsack Fallback)
To minimize transaction fees and prevent UTXO pool fragmentation, coin selection selects an optimal subset of UTXOs such that the total value satisfies the withdrawal target plus estimated fees, minimizing change outputs.

For a set of available UTXOs $U = \{u_1, u_2, \dots, u_n\}$ where each UTXO $u_i$ has satoshi value $v(u_i)$ and input virtual size $\text{vsize}(u_i)$:

$$\text{EffectiveValue}(u_i) = v(u_i) - \text{FeeRate} \times \text{vsize}(u_i)$$

The optimization objective for exact match (Branch-and-Bound without change):

$$\text{Minimize } \text{Waste}(S) \quad \text{subject to} \quad \sum_{u_i \in S} \text{EffectiveValue}(u_i) = \text{TargetAmount}$$

Where the waste metric is defined as:

$$\text{Waste}(S) = \sum_{u_i \in S} (\text{FeeRate} - \text{LongTermFeeRate}) \times \text{vsize}(u_i) + \text{CostOfChange}$$

If Branch-and-Bound fails to find an exact match within 100,000 iterations, the Knapsack algorithm executes with a change output:

$$\sum_{u_i \in S} \text{EffectiveValue}(u_i) \ge \text{TargetAmount} + \text{CostOfChange} + \text{DustThreshold}$$

### 2. Opt-in Replace-By-Fee (BIP 125) Validation & Replacement Rules
A transaction $T_{new}$ replaces an unconfirmed transaction $T_{old}$ in the mempool if and only if all BIP 125 conditions are satisfied:

1. **Signaling:** At least one input in $T_{old}$ has $nSequence \le 0xFFFFFFFD$.
2. **No New Unconfirmed Inputs:** $T_{new}$ does not contain any unconfirmed inputs that were not in $T_{old}$.
3. **Absolute Fee Increase:**
   $$\text{Fee}(T_{new}) > \text{Fee}(T_{old}) + \text{IncrementalRelayFee} \times \text{vsize}(T_{new})$$
4. **Fee Rate Increase:**
   $$\text{FeeRate}(T_{new}) > \text{FeeRate}(T_{old})$$

The double-spend engine marks a deposit as compromised if:

$$\exists u_k \in \text{Inputs}(T_{old}) \cap \text{Inputs}(T_{new}) \quad \text{where} \quad \text{Outputs}(T_{new}) \cap \text{DepositAddresses} = \emptyset$$

### 3. Multi-Tier Liquidity Allocation & Rebalancing Model
Let $B_{total}$ be total Bitcoin under custody:

$$B_{total} = B_{hot} + B_{warm} + B_{cold} + B_{lightning}$$

Target allocations are governed by strict parameters:
- Hot Wallet ($B_{hot}$): $5\% \pm 2\%$ of $B_{total}$ (Target $T_{hot} = 0.05 \cdot B_{total}$, Max $L_{max}^{hot} = 0.07 \cdot B_{total}$, Min $L_{min}^{hot} = 0.03 \cdot B_{total}$)
- Warm HSM Vault ($B_{warm}$): $15\% \pm 5\%$ of $B_{total}$
- Lightning Liquidity ($B_{lightning}$): $2\% \pm 1\%$ of $B_{total}$
- Cold Multi-Sig Vault ($B_{cold}$): Remainder ($\ge 75\%$ of $B_{total}$)

Sweep trigger condition:

$$\text{If } B_{hot} > L_{max}^{hot} \implies \text{SweepAmount} = B_{hot} - T_{hot} \longrightarrow \text{Warm / Cold Vault}$$

$$\text{If } B_{hot} < L_{min}^{hot} \implies \text{RebalanceAmount} = T_{hot} - B_{hot} \longleftarrow \text{Warm Vault}$$

### 4. Lightning Channel Balance & Rebalancing Index
For each active payment channel $c$ with local balance $B_{local}(c)$ and remote balance $B_{remote}(c)$:

$$\text{ChannelImbalance}(c) = \frac{B_{local}(c) - B_{remote}(c)}{B_{local}(c) + B_{remote}(c)} \in [-1.0, 1.0]$$

- If $\text{ChannelImbalance}(c) > 0.70$ (local excess, outbound heavy): Trigger **Loop Out** (Submarine Swap to on-chain UTXO).
- If $\text{ChannelImbalance}(c) < -0.70$ (remote excess, inbound depleted): Trigger **Loop In** (Submarine Swap from on-chain UTXO).

### 5. Blockchain Reorganization Depth & Risk Probability
Adversarial reorg probability $P(k)$ after $z$ confirmations with attacker hashrate share $q < 0.5$ and honest share $p = 1 - q$:

$$P(z) = \sum_{k=0}^{\infty} \frac{\lambda^k e^{-\lambda}}{k!} \times \min\left(1, \left(\frac{q}{p}\right)^{\max(0, z-k)}\right) \quad \text{where} \quad \lambda = z \frac{q}{p}$$

For $z = 6$ and $q = 0.10$, $P(6) < 0.0001$, satisfying the risk threshold for clearing on-chain collateral minting.

## Interfaces / Contracts

### Protobuf Service & Message Definitions (`btc_ingress.proto`)
```protobuf
syntax = "proto3";

package growww.btc.v1;

option go_package = "github.com/growww/services/btc-lightning-ingress/gen/v1;btcv1";

service BtcLightningIngressService {
  rpc GenerateDepositAddress (GenerateDepositAddressRequest) returns (GenerateDepositAddressResponse);
  rpc CreateLightningInvoice (CreateLightningInvoiceRequest) returns (CreateLightningInvoiceResponse);
  rpc QueryDepositStatus (QueryDepositStatusRequest) returns (QueryDepositStatusResponse);
  rpc RequestOnChainWithdrawal (RequestOnChainWithdrawalRequest) returns (RequestOnChainWithdrawalResponse);
  rpc RequestLightningWithdrawal (RequestLightningWithdrawalRequest) returns (RequestLightningWithdrawalResponse);
  rpc ExecuteVaultSweep (ExecuteVaultSweepRequest) returns (ExecuteVaultSweepResponse);
  rpc GetVaultBalances (GetVaultBalancesRequest) returns (GetVaultBalancesResponse);
  rpc RebalanceLightningLiquidity (RebalanceLightningLiquidityRequest) returns (RebalanceLightningLiquidityResponse);
}

enum AddressType {
  ADDRESS_TYPE_UNSPECIFIED = 0;
  ADDRESS_TYPE_NATIVE_SEGWIT_P2WPKH = 1; // BIP 84 (bc1q...)
  ADDRESS_TYPE_TAPROOT_P2TR = 2;          // BIP 86 (bc1p...)
}

enum DepositChannel {
  DEPOSIT_CHANNEL_UNSPECIFIED = 0;
  DEPOSIT_CHANNEL_ON_CHAIN_BTC = 1;
  DEPOSIT_CHANNEL_LIGHTNING_NETWORK = 2;
}

enum DepositState {
  DEPOSIT_STATE_UNSPECIFIED = 0;
  DEPOSIT_STATE_DETECTED_MEMPOOL = 1;
  DEPOSIT_STATE_CONFIRMING = 2;
  DEPOSIT_STATE_CONFIRMED = 3;
  DEPOSIT_STATE_COLLATERAL_MINTED = 4;
  DEPOSIT_STATE_RBF_CONFLICT_DETECTED = 5;
  DEPOSIT_STATE_REORG_INVALIDATED = 6;
}

enum VaultTier {
  VAULT_TIER_UNSPECIFIED = 0;
  VAULT_TIER_HOT = 1;
  VAULT_TIER_WARM_HSM = 2;
  VAULT_TIER_COLD_MULTISIG = 3;
  VAULT_TIER_LIGHTNING_CHANNELS = 4;
}

message GenerateDepositAddressRequest {
  string investor_id = 1;
  AddressType address_type = 2;
  string label = 3;
}

message GenerateDepositAddressResponse {
  string address = 1;
  AddressType address_type = 2;
  string derivation_path = 3;
  int64 generated_at_utc = 4;
}

message CreateLightningInvoiceRequest {
  string investor_id = 1;
  uint64 amount_satoshis = 2;
  string memo = 3;
  int64 expiry_seconds = 4;
}

message CreateLightningInvoiceResponse {
  string payment_request = 1; // BOLT 11 payment request string
  string payment_hash = 2;    // SHA-256 payment hash hex
  string r_preimage = 3;      // Preimage (empty until settled)
  uint64 amount_satoshis = 4;
  int64 expires_at_utc = 5;
}

message QueryDepositStatusRequest {
  string deposit_id = 1;
  string txid_or_payment_hash = 2;
}

message QueryDepositStatusResponse {
  string deposit_id = 1;
  DepositChannel channel = 2;
  string txid_or_payment_hash = 3;
  uint32 vout = 4;
  uint64 amount_satoshis = 5;
  uint32 confirmations = 6;
  DepositState state = 7;
  bool is_rbf_enabled = 8;
  string besu_mint_tx_hash = 9;
  int64 detected_at_utc = 10;
  int64 confirmed_at_utc = 11;
}

message RequestOnChainWithdrawalRequest {
  string withdrawal_id = 1;
  string investor_id = 2;
  string destination_address = 3;
  uint64 amount_satoshis = 4;
  uint64 target_fee_rate_sat_per_vbyte = 5;
}

message RequestOnChainWithdrawalResponse {
  string withdrawal_id = 1;
  string txid = 2;
  uint64 gross_satoshis = 3;
  uint64 network_fee_satoshis = 4;
  string psbt_session_id = 5;
  int64 broadcasted_at_utc = 6;
}

message RequestLightningWithdrawalRequest {
  string withdrawal_id = 1;
  string investor_id = 2;
  string bolt11_payment_request = 3;
  uint64 max_fee_satoshis = 4;
}

message RequestLightningWithdrawalResponse {
  string withdrawal_id = 1;
  string payment_hash = 2;
  string payment_preimage = 3;
  uint64 fee_paid_satoshis = 4;
  int64 settled_at_utc = 5;
}

message ExecuteVaultSweepRequest {
  VaultTier source_tier = 1;
  VaultTier destination_tier = 2;
  uint64 amount_satoshis = 3;
  string destination_address = 4;
}

message ExecuteVaultSweepResponse {
  string sweep_batch_id = 1;
  string txid = 2;
  uint64 swept_satoshis = 3;
  uint64 fee_satoshis = 4;
  string psbt_session_id = 5;
}

message GetVaultBalancesRequest {}

message GetVaultBalancesResponse {
  uint64 hot_balance_satoshis = 1;
  uint64 warm_balance_satoshis = 2;
  uint64 cold_balance_satoshis = 3;
  uint64 lightning_local_balance_satoshis = 4;
  uint64 lightning_remote_balance_satoshis = 5;
  uint64 total_custody_satoshis = 6;
  uint64 synthetic_besu_minted_satoshis = 7;
  bool is_solvent = 8;
  int64 reconciled_at_utc = 9;
}

message RebalanceLightningLiquidityRequest {
  string channel_id = 1;
  string swap_direction = 2; // LOOP_IN or LOOP_OUT
  uint64 amount_satoshis = 3;
}

message RebalanceLightningLiquidityResponse {
  string swap_id = 1;
  string htlc_txid = 2;
  uint64 amount_satoshis = 3;
  uint64 cost_satoshis = 4;
  string status = 5;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Bitcoin HD Deposit Addresses Pool
CREATE TABLE btc_deposit_addresses (
    address VARCHAR(90) PRIMARY KEY,
    investor_id UUID NOT NULL,
    address_type VARCHAR(32) NOT NULL CHECK (address_type IN ('NATIVE_SEGWIT_P2WPKH', 'TAPROOT_P2TR')),
    account_index INT NOT NULL,
    address_index INT NOT NULL,
    derivation_path VARCHAR(64) NOT NULL UNIQUE,
    script_pubkey BYTEA NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- On-Chain UTXO State Tracking
CREATE TABLE btc_utxos (
    txid VARCHAR(64) NOT NULL,
    vout INT NOT NULL,
    address VARCHAR(90) NOT NULL REFERENCES btc_deposit_addresses(address),
    amount_satoshis BIGINT NOT NULL CHECK (amount_satoshis > 0),
    script_pubkey BYTEA NOT NULL,
    block_height BIGINT,
    block_hash VARCHAR(64),
    confirmations INT NOT NULL DEFAULT 0,
    vault_tier VARCHAR(32) NOT NULL DEFAULT 'HOT' CHECK (vault_tier IN ('HOT', 'WARM_HSM', 'COLD_MULTISIG')),
    is_spent BOOLEAN NOT NULL DEFAULT FALSE,
    spent_in_txid VARCHAR(64),
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    locked_by_session UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (txid, vout)
);

-- On-Chain Inbound / Outbound Transactions
CREATE TABLE btc_onchain_transactions (
    txid VARCHAR(64) PRIMARY KEY,
    transaction_type VARCHAR(32) NOT NULL CHECK (transaction_type IN ('DEPOSIT', 'WITHDRAWAL', 'SWEEP_HOT_TO_WARM', 'SWEEP_WARM_TO_COLD', 'REBALANCE')),
    raw_hex TEXT NOT NULL,
    vsize INT NOT NULL,
    fee_satoshis BIGINT NOT NULL DEFAULT 0,
    fee_rate_sat_per_vbyte NUMERIC(10, 2) NOT NULL,
    is_rbf_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rbf_sequence_number BIGINT NOT NULL,
    replaces_txid VARCHAR(64),
    replaced_by_txid VARCHAR(64),
    block_height BIGINT,
    block_hash VARCHAR(64),
    confirmations INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'CONFIRMED', 'REPLACED', 'REORG_INVALIDATED', 'FAILED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);

-- Lightning Network Invoices (Inbound Deposits)
CREATE TABLE lightning_invoices (
    payment_hash VARCHAR(64) PRIMARY KEY,
    investor_id UUID NOT NULL,
    payment_request TEXT NOT NULL UNIQUE,
    r_preimage VARCHAR(64),
    amount_satoshis BIGINT NOT NULL CHECK (amount_satoshis > 0),
    fee_budget_satoshis BIGINT NOT NULL DEFAULT 0,
    state VARCHAR(32) NOT NULL DEFAULT 'OPEN' CHECK (state IN ('OPEN', 'SETTLED', 'CANCELED', 'EXPIRED')),
    expiry_seconds INT NOT NULL DEFAULT 3600,
    besu_mint_tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Lightning Network Outbound Payments (Withdrawals)
CREATE TABLE lightning_payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    withdrawal_id UUID NOT NULL UNIQUE,
    investor_id UUID NOT NULL,
    payment_request TEXT NOT NULL,
    payment_hash VARCHAR(64) NOT NULL,
    payment_preimage VARCHAR(64),
    amount_satoshis BIGINT NOT NULL,
    fee_paid_satoshis BIGINT NOT NULL DEFAULT 0,
    route_hops_count INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'IN_FLIGHT' CHECK (status IN ('IN_FLIGHT', 'SUCCEEDED', 'FAILED')),
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ
);

-- PSBT Construction & HSM Signing Sessions
CREATE TABLE psbt_signing_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_type VARCHAR(32) NOT NULL CHECK (session_type IN ('WITHDRAWAL', 'VAULT_SWEEP', 'SUBMARINE_SWAP')),
    unsigned_psbt_base64 TEXT NOT NULL,
    signed_psbt_base64 TEXT,
    target_fee_rate NUMERIC(10, 2) NOT NULL,
    total_input_satoshis BIGINT NOT NULL,
    total_output_satoshis BIGINT NOT NULL,
    fee_satoshis BIGINT NOT NULL,
    hsm_key_identifier VARCHAR(128) NOT NULL,
    hsm_signature_digest VARCHAR(128),
    state VARCHAR(32) NOT NULL DEFAULT 'CREATED' CHECK (state IN ('CREATED', 'SIGNED', 'FINALIZED', 'BROADCASTED', 'REJECTED')),
    broadcasted_txid VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Hyperledger Besu Synthetic Collateral Bridge Ledger
CREATE TABLE synthetic_collateral_bridges (
    bridge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL,
    direction VARCHAR(16) NOT NULL CHECK (direction IN ('MINT_COLLATERAL', 'BURN_COLLATERAL')),
    amount_satoshis BIGINT NOT NULL CHECK (amount_satoshis > 0),
    btc_channel VARCHAR(32) NOT NULL CHECK (btc_channel IN ('ON_CHAIN_BTC', 'LIGHTNING_NETWORK')),
    btc_reference_id VARCHAR(64) NOT NULL, -- txid or payment_hash
    besu_tx_hash VARCHAR(66) NOT NULL UNIQUE,
    besu_block_number BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'COMPLETED' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED')),
    proof_merkle_root VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for Low-Latency Lookups
CREATE INDEX idx_btc_utxo_address ON btc_utxos (address) WHERE NOT is_spent;
CREATE INDEX idx_btc_utxo_tier_unspent ON btc_utxos (vault_tier, is_spent, is_locked);
CREATE INDEX idx_btc_tx_confirmations ON btc_onchain_transactions (status, confirmations);
CREATE INDEX idx_lightning_invoices_state ON lightning_invoices (state, expires_at);
CREATE INDEX idx_synthetic_bridges_investor ON synthetic_collateral_bridges (investor_id, created_at DESC);
```

### Kafka Event Schemas

#### 1. On-Chain Deposit Confirmed Event (`btc.deposit.confirmed.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "BtcDepositConfirmedEvent",
  "type": "object",
  "properties": {
    "deposit_id": { "type": "string", "format": "uuid" },
    "investor_id": { "type": "string", "format": "uuid" },
    "txid": { "type": "string", "pattern": "^[a-fA-F0-9]{64}$" },
    "vout": { "type": "integer", "minimum": 0 },
    "address": { "type": "string" },
    "address_type": { "type": "string", "enum": ["NATIVE_SEGWIT_P2WPKH", "TAPROOT_P2TR"] },
    "amount_satoshis": { "type": "integer", "minimum": 1 },
    "block_height": { "type": "integer" },
    "block_hash": { "type": "string", "pattern": "^[a-fA-F0-9]{64}$" },
    "confirmations": { "type": "integer", "minimum": 6 },
    "confirmed_at_utc": { "type": "integer" }
  },
  "required": [
    "deposit_id",
    "investor_id",
    "txid",
    "vout",
    "address",
    "amount_satoshis",
    "block_height",
    "block_hash",
    "confirmations",
    "confirmed_at_utc"
  ]
}
```

#### 2. Lightning Invoice Settled Event (`lightning.invoice.settled.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "LightningInvoiceSettledEvent",
  "type": "object",
  "properties": {
    "investor_id": { "type": "string", "format": "uuid" },
    "payment_hash": { "type": "string", "pattern": "^[a-fA-F0-9]{64}$" },
    "payment_preimage": { "type": "string", "pattern": "^[a-fA-F0-9]{64}$" },
    "amount_satoshis": { "type": "integer", "minimum": 1 },
    "settled_at_utc": { "type": "integer" }
  },
  "required": [
    "investor_id",
    "payment_hash",
    "payment_preimage",
    "amount_satoshis",
    "settled_at_utc"
  ]
}
```

#### 3. Besu Synthetic Collateral Mint Dispatched Event (`btc.collateral.mint_dispatched.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "BtcCollateralMintDispatchedEvent",
  "type": "object",
  "properties": {
    "bridge_id": { "type": "string", "format": "uuid" },
    "investor_id": { "type": "string", "format": "uuid" },
    "besu_recipient_address": { "type": "string", "pattern": "^0x[a-fA-F0-9]{40}$" },
    "amount_satoshis": { "type": "integer", "minimum": 1 },
    "btc_channel": { "type": "string", "enum": ["ON_CHAIN_BTC", "LIGHTNING_NETWORK"] },
    "btc_reference_id": { "type": "string" },
    "besu_tx_hash": { "type": "string", "pattern": "^0x[a-fA-F0-9]{64}$" },
    "proof_merkle_root": { "type": "string", "pattern": "^[a-fA-F0-9]{64}$" },
    "dispatched_at_utc": { "type": "integer" }
  },
  "required": [
    "bridge_id",
    "investor_id",
    "besu_recipient_address",
    "amount_satoshis",
    "btc_channel",
    "btc_reference_id",
    "besu_tx_hash",
    "proof_merkle_root",
    "dispatched_at_utc"
  ]
}
```

## Security & Compliance Notes
- **FIPS 140-2 Level 3 Hardware Security Module (HSM) Key Isolation:** All master extended private keys (`xprv`), Taproot internal keys, and Schnorr signing seeds are strictly non-exportable and sealed inside HSM hardware partitions. Application services access cryptographic operations via PKCS#11 sessions requiring mTLS authentication and hardware token authorization.
- **Strict 6-Block Confirmation Finality:** To guard against blockchain reorganizations and double-spend attacks on Bitcoin Layer-1, on-chain deposits are strictly quarantined until 6 contiguous blocks have been mined on top of the transaction block. Synthetic collateral is never credited prematurely.
- **Zero-Confirmation Credit Prohibition:** Inbound transactions signaling BIP 125 Opt-in RBF or unconfirmed transactions in the mempool receive zero trading collateral credit under all circumstances.
- **Lightning Channel Watchtower Deployment:** Continuous, 24/7 Watchtower nodes (`wtclient`) monitor the Lightning Network to counter rogue or outdated state broadcasts during any temporary network partition.
- **Travel Rule & AML Address Screening:** Before address derivation and withdrawal broadcasting, recipient and deposit addresses are verified against OFAC, UN, and UAPA sanctions databases via Prompt 202 and Prompt 703.
- **Zero PII on Distributed Ledger:** The Hyperledger Besu permissioned ledger retains only cryptographic transaction hashes, public keys, and token balances. No personal identities, tax identifiers, or KYC documents are ever committed on-chain.

## Acceptance Criteria
- [ ] Successfully derives valid BIP 84 (P2WPKH) and BIP 86 (Taproot P2TR) deposit addresses matching standard cryptographic test vectors.
- [ ] Ingests on-chain Bitcoin blocks and transactions via ZeroMQ and JSON-RPC, correctly tracking confirmation depth from 0 to 6+ blocks.
- [ ] Detects BIP 125 Opt-in RBF transactions in the mempool and pauses processing if an output is redirected away from the platform address.
- [ ] Generates, verifies, and settles BOLT 11 invoices and BOLT 12 offers via LND/CLN gRPC with sub-second response times.
- [ ] Constructs valid BIP 174 / BIP 370 PSBT payloads using Branch-and-Bound and Knapsack coin selection algorithms.
- [ ] Integrates with CloudHSM via PKCS#11 to produce valid ECDSA and BIP 340 Schnorr signatures on PSBT inputs.
- [ ] Executes automated balance sweeps from hot wallets to warm/cold vaults when hot balance exceeds the 7% threshold.
- [ ] Reliably triggers 1:1 synthetic BTC collateral (`sBTC`) minting and burning on Hyperledger Besu with zero balance discrepancy.
- [ ] Handles 100-block testnet chain reorganizations gracefully without creating duplicate collateral or orphaned balances.
- [ ] Unit and regtest integration test coverage exceeds $\ge 90\%$ without application runtime compilation failures.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 003 (Regulatory Pathway Overview), Prompt 103 (API Design Standards), Prompt 104 (Event Schema Standards), Prompt 109 (Secrets Management & HSM), Prompt 203 (Wallet & Account Service), Prompt 303 (Token Issuance Smart Contract), Prompt 308 (On-Chain Proof of Reserve Publishing).
- **Subsequent / Parallel Tasks:** Prompt 206 (Risk and Margin Checks Service), Prompt 214 (Foreign Investor Funding & FX Service), Prompt 215 (Reconciliation Service), Prompt 240 (Perpetuals and Synthetic Derivatives Engine).
