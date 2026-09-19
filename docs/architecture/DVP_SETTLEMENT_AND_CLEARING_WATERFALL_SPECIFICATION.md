# Delivery-versus-Payment (DvP) Settlement & Clearing Waterfall Specification

**Specification ID:** SPEC-ARCH-010-DVP  
**Document Version:** 4.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Clearing Corporation & Settlement Operations Group  

---

## 1. Three-Tier Distributed State Clearing Architecture

An order execution spans three coordinated state domains:
1. **Financial Balance Ledger (`tigerbeetle`):** Real-time dual-entry accounting with sub-millisecond pending transfers (`Ledger ID 1` for Real, `Ledger ID 2` for Demo).
2. **In-Memory Matching Engine (`matching-engine`):** Matches orders in Rust memory in $< 1\mu\text{s}$ with dynamic fee parameter logic (starts at 0.00% for all).
3. **Consensus Settlement Ledger (`trade-settlement-service` / Hyperledger Besu):** Settles atomic DvP on-chain in 2.0-second blocks with zero tax withholding and batched netting.

```
+----------------------------------------------------------------------------------------------------+
| THREE-TIER STATE CLEARING & COMPENSATING SGF GUARANTEE                                             |
|                                                                                                    |
|  [TigerBeetle Pending Hold] ---> [In-Memory Match Finality] ---> [Multilateral Netting Relayer]   |
|                                                                                |                   |
|                                                                                v                   |
|                                                                   [Hyperledger Besu DvP Batch]     |
|                                                                                |                   |
|                                                   +----------------------------+                   |
|                                                   | (If On-Chain Batch Reverts)                    |
|                                                   v                                                |
|                                    [Compensating SGF Guarantee]                                    |
|                                    - SGF Acts as Central Counterparty (CCP)                        |
|                                    - Fulfills Buyer Delivery from Reserve Vault                    |
|                                    - Void TigerBeetle Pending Hold & Alert SRE                     |
+----------------------------------------------------------------------------------------------------+
```

## 2. Dynamic Fee Model & Starting Invariants
- **Starting Fee**: Strictly **0.00% for Maker and Taker ("0 means 0 in all")**.
- **Dynamic Parameter Expansion**:
  - `Fee = (Notional * activeFeeBps) / 10000`
  - In starting mode, `activeFeeBps = 0`, resulting in `Fee = 0` and 100% net settlement proceeds delivered.
  - If governance modifies `activeFeeBps` in the future via `FeeController.sol`, the clearing waterfall automatically applies the new schedule upon timelock expiry.
- **TDS Deduction**: Strictly 0.00% (Zero TDS on-chain).

## 3. Nonce Sharding & High-Throughput Relayers
- 32 relayer accounts (`0xRelayer00` to `0xRelayer31`) submit batches in parallel.
- Monotonic nonce tracking eliminates transaction collisions.
- Dynamic EIP-1559 replacement (15% priority fee bump) prevents mempool stalls.

## 4. Continuous 3-Way Reconciliation
- Automated 60-second verification comparing TigerBeetle balances, Kafka processed offsets, and Besu vault tokens.
