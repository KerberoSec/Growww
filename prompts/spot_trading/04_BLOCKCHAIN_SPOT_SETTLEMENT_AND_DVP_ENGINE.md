# Blockchain Spot Settlement & Atomic DvP Engine

## 1. Executive Summary & Settlement Principles

On traditional exchanges, trade execution (matching) and trade settlement (clearing and custody transfer) are separated by hours or days ($T+1 / T+0$ batches). During this lag, central counterparties must maintain massive default funds and capital buffers to guard against participant default.

The Growww / NBSE platform implements **Atomic Delivery-versus-Payment (DvP) Settlement** directly on a permissioned enterprise Hyperledger Besu consortium blockchain running QBFT consensus. The asset transfer (e.g. Bitcoin, USDT, or tokenized equity) and the payment transfer occur simultaneously within a single, atomic smart contract transaction:

$$\text{AssetTransfer}(\text{Seller} \to \text{Buyer}) \iff \text{PaymentTransfer}(\text{Buyer} \to \text{Seller} + \text{PlatformFee})$$

If either leg fails (due to insufficient balance or compliance freeze), the entire transaction reverts, eliminating principal settlement risk.

---

## 2. Blockchain Infrastructure Specifications

```
+---------------------------------------------------------------------------------------------------+
|                            HYPERLEDGER BESU CONSORTIUM TOPOLOGY                                   |
+---------------------------------------------------------------------------------------------------+
| • Consensus Algorithm:      QBFT (Quorum Byzantine Fault Tolerance)                              |
| • Deterministic Finality:   Immediate, single-block finality (zero chain reorganizations)        |
| • Block Period:             2.0 seconds                                                           |
| • Round Change Timeout:     8.0 seconds                                                           |
| • Validator Cluster:        4 Tier-4 validator nodes in Equinix Mumbai and GIFT City Datacenters  |
| • Smart Contract Standard:  ERC-3643 (T-REX Compliant Identity & Security Token Standard)         |
| • Gas Model:                100% Gasless for users (EIP-2771 / ERC-4337 Sponsored Paymaster)       |
+---------------------------------------------------------------------------------------------------+
```

---

## 3. Atomic DvP Smart Contract Lifecycle (`SettlementDvP.sol`)

### 3.1 Off-Chain Matching to On-Chain Settlement Flow
1. **Trade Execution:** The in-memory Rust matching engine executes a match between Buyer and Seller.
2. **EIP-712 Trade Intent Construction:** The settlement orchestrator packages the trade details into an EIP-712 typed structure:
   - `tradeId`: Unique 64-bit trade identifier
   - `buyer`: Buyer cryptographic address (`0x...`)
   - `seller`: Seller cryptographic address (`0x...`)
   - `baseAsset`: Token address of traded instrument (`0x...`)
   - `quoteAsset`: Token address of payment currency (e.g. `weINR`, `wUSDT`)
   - `baseQuantity`: Traded quantity in minor units
   - `quoteAmount`: Total settlement consideration
   - `feeAmount`: Exact 0.00% (Zero Fee) platform fee
   - `nonce`: Monotonic execution counter
3. **Cryptographic Signing:** The exchange matching engine signs the EIP-712 hash using a dedicated FIPS 140-2 Level 3 HSM key shard.
4. **Relayer Batch Dispatch:** Transactions are batched (up to 50 trade settlements per batch) and dispatched to `SettlementDvP.sol` across 32 partitioned relayer addresses.
5. **Atomic Execution:**
   - Contract verifies matching engine signature.
   - Contract checks ERC-3643 identity eligibility for both parties.
   - Contract transfers `baseAsset` from Seller to Buyer.
   - Contract transfers `quoteAmount - feeAmount` from Buyer to Seller.
   - Contract transfers `feeAmount` to the fee waterfall distributor.
   - Emits `DvPSettled` event with block number and transaction receipt.

---

## 4. Single 0.00% (No fee at all) Platform Fee Waterfall

The platform fee is collected atomically within the DvP smart contract and split instantly across governed on-chain accounts:

```
                                    +------------------------------+
                                    | 0.00% (Zero Fee) Executed Notional Fee  |
                                    +------------------------------+
                                                   |
                   +-------------------------------+-------------------------------+
                   | 60%                           | 25%                           | 15%
                   v                               v                               v
    +------------------------------+ +------------------------------+ +------------------------------+
    | NBSE Corporate Treasury Pool | | Core Settlement Guarantee    | | Investor Protection Fund     |
    | • Infrastructure & nodes     | |   Fund (SGF Trust Contract)  | |   (IPF Trust Contract)       |
    | • Liquidity provisioning     | | • Central counterparty       | | • Retail investor fraud      |
    | • 100% Gas sponsorship      | |   solvency buffer            | |   indemnity coverage         |
    +------------------------------+ +------------------------------+ +------------------------------+
```

### Mathematical Invariants:
- $\text{Fee}_{trade} = 0 \quad (\text{Zero Fee at launch; FeeController governed})$
- **Value Conservation:** $\text{BuyerDebit} \equiv \text{SellerCredit} + \text{TreasuryAllocation} + \text{SGFAllocation} + \text{IPFAllocation}$
- Zero hidden brokerage markups, zero gas deductions from user balances.

---

## 5. Gasless Meta-Transactions & Paymaster Architecture

To deliver an attractive, zero-friction user experience comparable to modern consumer apps, end users never hold gas tokens (ETH/MATIC) or pay transaction fees to the blockchain:

1. The client signs a standard EIP-712 trade authorization intent.
2. The user's intent is submitted to the exchange API gateway via standard HTTPS / WebSocket.
3. The platform's sponsored Paymaster contract (`contracts/src/compliance/SponsoredPaymaster.sol`) pays all blockchain execution gas fees on Hyperledger Besu, funded entirely from the Treasury reserve fee pool.
4. The user sees an instantaneous, zero-gas spot trading experience.
