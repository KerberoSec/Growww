# Multichain USDT Ingress, Finality Matrix & Automated Gas Station Specification

**Specification ID:** SPEC-ARCH-035-USDT  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Core Treasury & Multichain Custody  
**Owner:** Treasury Engineering & Cross-Chain Infrastructure Group  

---

## 1. Executive Summary & Core Invariants
The multichain USDT ingestion engine processes real-time stablecoin deposits across multiple Layer-1 and Layer-2 blockchains (Ethereum ERC-20, Tron TRC-20, Polygon POS, Arbitrum One, and Binance Smart Chain). It guarantees:
- **Zero Reorg Double-Spend Exposure**: Dynamic block confirmation matrix strictly enforced before crediting trading balances in TigerBeetle.
- **Automated Gas Station Provisioning**: Programmatically funds user deposit addresses with native gas tokens (ETH/TRX/POL/BNB) just-in-time for automated sweeping without exposing hot private keys.
- **1:1 Besu Settlement Token Backing**: Every external USDT deposited is held in segregated MPC custody vaults and mirrors 1:1 with `wUSDT` on Hyperledger Besu.

---

## 2. Dynamic Confirmation Depth & Finality Matrix

| Chain | Transport / Standard | Required Confirmation Depth | Estimated Settlement Latency | Finality Type |
| :--- | :--- | :--- | :--- | :--- |
| **Ethereum** | ERC-20 (`0xdAC1...`) | **12 blocks** (or 2 PoS epochs) | ~2.5 - 3.0 minutes | Casper FFG Checkpoint Finality |
| **Tron** | TRC-20 (`TR7NH...`) | **19 blocks** | ~57 seconds | DPoS Irreversible Solidified Block |
| **Polygon PoS** | ERC-20 (`0xc213...`) | **128 blocks** (or Heimdall Milestone) | ~4.2 minutes | State Checkpoint Commit |
| **Arbitrum One** | ERC-20 (`0xFd08...`) | **64 L1 Batches** | ~1.5 - 2.0 minutes | Rollup L1 Settlement Finality |
| **BNB Smart Chain**| BEP-20 (`0x55d3...`) | **15 blocks** | ~45 seconds | Parlia Consensus Fast Finality |

### Reorg Protection Engine:
If a block reorg occurs prior to reaching target confirmation depth:
1. The chain listener invalidates unconfirmed deposit intent events.
2. The transaction hash is placed into a 30-minute quarantine monitor.
3. Balance crediting in TigerBeetle occurs strictly after the target depth is reached.

---

## 3. Automated Gas Station & Treasury Sweeping Protocol

```
+----------------------------------------------------------------------------------------------------+
| JUST-IN-TIME GAS STATION & MULTICHAIN SWEEPER PIPELINE                                             |
|                                                                                                    |
|  [ User Deposits USDT ] ---> [ Chain Listener Detects Transfer ]                                   |
|                                         |                                                          |
|                                         v                                                          |
|                        [ Confirmation Depth Satisfied? ]                                           |
|                                 |               |                                                  |
|                        YES      |               | NO -> [ Wait for Target Block Height ]           |
|                                 v                                                                  |
|                   [ Credit TigerBeetle Ledger ID 1 ]                                               |
|                   (Trader Can Trade BTC/USDT Instantly)                                            |
|                                 |                                                                  |
|                                 v                                                                  |
|                   [ Deposit Address Balance > $500? ]                                              |
|                                 |                                                                  |
|                   YES           v                                                                  |
|      [ Gas Station Dispatches Micro-Gas Transfer (e.g. 15 TRX / 0.001 ETH) ]                       |
|                                 |                                                                  |
|                                 v                                                                  |
|      [ MPC Sweeper Signs Sweep Transaction -> Moves USDT to Omnibus Cold Vault ]                   |
+----------------------------------------------------------------------------------------------------+
```

### Sweeper Security Measures:
- **Hot Gas Dispatcher**: Operates with restricted daily limits ($2,000 USD equivalent in gas tokens).
- **Anti-Frontrunning Gas Dispatch**: On Ethereum and Arbitrum, sweeping is executed via Flashbots private mempools or ERC-4337 UserOperations to eliminate front-running bots draining gas funds.
- **Sub-Cent Accumulator**: Deposit addresses holding $< 50$ USDT are not swept individually; balances accumulate until crossing profitability thresholds ($> 200$ USDT) to optimize network gas expenditure.
