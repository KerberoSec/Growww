# RWA Launchpad, Dutch Auction & Primary Capital Issuance Specification

**Specification ID:** SPEC-ARCH-047-LAUNCH  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Primary Capital Formation & Token Offering Architecture  
**Owner:** Capital Markets & Token Launchpad Engineering Group  

---

## 1. Executive Summary & Issuance Parameters
The RWA Launchpad provides institutional issuers and asset originators with an on-chain capital raising and primary token distribution platform:
- **Zero Protocol Issuance Fee**: **0.00% platform commission** on primary capital formation (No fee at all).
- **Algorithmic Dutch Auction**: Eliminates gas wars and front-running by lowering price linearly over time until supply meets demand.
- **Linear Smart Contract Vesting**: Continuous block-by-block token vesting governed by `LinearVestingVault.sol` with cliff schedules and multi-sig emergency controls.

---

## 2. Dutch Auction State Machine & Clearing Mechanics

```
+----------------------------------------------------------------------------------------------------+
| RWA PRIMARY ISSUANCE DUTCH AUCTION LIFECYCLE                                                       |
|                                                                                                    |
|  [ 1. Offering Initialized ] ----> Issuer deposits 1,000,000 RWA Tokens into DutchAuction.sol       |
|                 |                  Start Price: $100.00 | Floor Price: $40.00 | Duration: 72 Hours  |
|                 v                                                                                  |
|  [ 2. Bidding Period ] ----------> Participants submit commitment bids with escrowed USDT          |
|                 |                  Cumulative Commitment: B(t) = Sum(bid_amount_i)                |
|                 v                                                                                  |
|  [ 3. Clearing Threshold Met ] --> Occurs when ClearingPrice(t) * TotalOffering == B(t)            |
|                 |                  Uniform Clearing Price Established                              |
|                 v                                                                                  |
|  [ 4. Settlement & Allocation ] -> Tokens minted/transferred to winning bidders at uniform price   |
|                 |                  Excess bid funds refunded immediately with 0.00% fees           |
|                 v                                                                                  |
|  [ 5. Secondary Spot Trading ] --> wUSDT/RWA token pair automatically opens on CLOB exchange       |
+----------------------------------------------------------------------------------------------------+
```
