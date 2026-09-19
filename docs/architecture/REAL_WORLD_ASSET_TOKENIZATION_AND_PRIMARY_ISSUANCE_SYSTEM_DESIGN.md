# Real-World Asset (RWA) Tokenization & Primary Issuance Specification

**Specification ID:** SPEC-ARCH-044-RWA  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Asset Tokenization & Primary Capital Markets  
**Owner:** Capital Markets & RWA Tokenization Group  

---

## 1. Executive Summary & Legal Token Standards
The RWA Tokenization framework enables compliant on-chain issuance and primary trading of tokenized real-world assets (equity shares, sovereign bonds, and bullion) on Hyperledger Besu:
- **ERC-3643 (T-REX) Standard**: Enforces permissioned token transfers, on-chain identity verification, and investor accreditation hooks.
- **100% Real Depository Backing**: Every security token is matched 1:1 with physical securities vaulted in NSDL / CDSL demat accounts or LBMA-accredited gold vaults.
- **Universal Zero-Fee Model**: **0.00% issuance and trading fees** (No fee at all).
- **Dutch Auction Price Discovery**: Primary offerings use algorithmic Dutch auctions to ensure fair market clearing prices without front-running.

---

## 2. Primary Issuance & Dutch Auction Lifecycle

```
+----------------------------------------------------------------------------------------------------+
| RWA TOKENIZATION & DUTCH AUCTION ISSUANCE LIFECYCLE                                                |
|                                                                                                    |
|  [ 1. Asset Vaulting ] --------> NSDL/CDSL custodian confirms shares received in pool account       |
|            |                                                                                       |
|            v                                                                                       |
|  [ 2. Smart Contract Mint ] ---> Issuer deploys ERC-3643 token with 1:1 total supply cap            |
|            |                                                                                       |
|            v                                                                                       |
|  [ 3. Dutch Auction ] ---------> Price drops linearly from MaxPrice to ReservePrice over 48 hours   |
|            |                     Bids accumulate until Total Bids == Total Offering Units          |
|            v                                                                                       |
|  [ 4. Clearing Finality ] -----> Uniform clearing price determined; all winning bidders pay        |
|            |                     the same clearing price; excess bid funds refunded                |
|            v                                                                                       |
|  [ 5. Secondary Trading ] -----> Unrestricted 24/7 spot trading enabled on Hyperledger Besu        |
+----------------------------------------------------------------------------------------------------+
```
