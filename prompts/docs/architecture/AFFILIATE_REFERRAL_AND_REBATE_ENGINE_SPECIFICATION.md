# Multi-Tier Affiliate Referral & Rebate Architecture Specification

**Specification ID:** SPEC-ARCH-002-AFFIL  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Growth Engineering & Referral Attribution  
**Owner:** Growth Platform & Referral Engineering Group  

---

## 1. Executive Summary & Attribution Invariants
The Affiliate Referral Engine tracks user onboarding attributions, multi-tier referral graphs, and gamified promotional reward distributions:
- **Universal Zero-Fee Trading Invariant**: All trading operations remain strictly **0.00% fee** (No fee at all).
- **Ecosystem Reward Pool Allocation**: Affiliate rewards and community bonuses are funded from dedicated platform growth incentives and treasury reserves, without assessing fees on trader volume.
- **Sybil-Resistant Attribution Graph**: Directed acyclic graph (DAG) tracking referral trees with automated cyclic attribution rejection and hardware device fingerprinting.

---

## 2. Multi-Tier Attribution & Reward Distribution

```
+----------------------------------------------------------------------------------------------------+
| MULTI-TIER AFFILIATE ATTRIBUTION PIPELINE                                                          |
|                                                                                                    |
|  [ New Trader Signs Up ] ---> [ Attribution Cookie / Referral Code Validated ]                     |
|                                                |                                                   |
|                                                v                                                   |
|                                [ Insert into DAG Referral Tree ]                                   |
|                                - Level 1: Direct Referrer (e.g. 50% reward share)                  |
|                                - Level 2: Parent Referrer (e.g. 20% reward share)                  |
|                                                |                                                   |
|                                                v                                                   |
|                                 [ Anti-Sybil Validation Engine ]                                   |
|                                 - IP Subnet Divergence Check                                       |
|                                 - Device Fingerprint Verification                                  |
|                                                |                                                   |
|                                                v                                                   |
|                         [ Disburse Rewards to TigerBeetle Ledger ID 1 ]                            |
|                         (Funded from Ecosystem Reserve; Zero Trading Fees Deducted)                |
+----------------------------------------------------------------------------------------------------+
```
