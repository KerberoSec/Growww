# Crypto Earn, PoS Staking & Yield Vaults Specification

**Specification ID:** SPEC-ARCH-012-EARN  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Institutional Staking Infrastructure & Yield Architecture  
**Owner:** Staking Infrastructure & Financial Products Group  

---

## 1. Executive Summary & Zero-Fee Invariant
The Crypto Earn subsystem provides users with non-custodial and custodial staking yields across major proof-of-stake cryptocurrencies (wETH, wBTC, USDT) and tokenized sovereign debt:
- **Zero Platform Commission**: 100% of blockchain network staking rewards and protocol yields are distributed directly to users (**0.00% exchange fee**).
- **Institutional Slashing Protection**: A pre-funded insurance reserve absorbs any validator downtime or consensus slashing penalties, guaranteeing principal protection.
- **Algorithmic Instant Unbonding**: An automated liquidity buffer allows users to instantly exit locked staking positions without waiting for native protocol unbonding queues (e.g. 7-day to 21-day unbonding periods).

---

## 2. Staking Topology & Yield Distribution Pipeline

```
+----------------------------------------------------------------------------------------------------+
| STAKING VALIDATOR ORCHESTRATION & YIELD DISTRIBUTION PIPELINE                                      |
|                                                                                                    |
|  [ User Stakes wETH / wBTC ] ---> [ StakeCoordinator.sol on Besu ]                                 |
|                                                |                                                   |
|                                                v                                                   |
|                            [ Route to Verified Institutional Validator ]                           |
|                            - Slashing Protection Enclave Monitored                                 |
|                            - Zero-Fee Direct Staking Contract                                     |
|                                                |                                                   |
|                                                v                                                   |
|                               [ Epoch Reward Accrual Engine ]                                      |
|                               - Daily Reward Compounding (e.g. 4.2% APY)                           |
|                                                |                                                   |
|                                                v                                                   |
|                       [ 100% Rewards Credited to TigerBeetle Ledger ID 1 ]                         |
|                       (Zero Exchange Deduction; Clean Financial Pass-Through)                      |
|                                                |                                                   |
|                                                v                                                   |
|                       [ Instant Exit Requested? ]                                                  |
|                               |                  |                                                 |
|                      YES      |                  | NO (Standard Protocol Unbonding)                |
|                               v                  v                                                 |
|       [ Liquidity Buffer Pool Swaps Locked Asset -> Instant Available Balance ]                    |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Slashing Protection Mechanics:
$$\text{NetReward} = \text{GrossYield} \times 1.00$$
If a validator is slashed by the underlying network, the platform's Slashing Guarantee Fund automatically replenishes the loss from the corporate reserve buffer, ensuring the user's principal balance never experiences a deficit.
