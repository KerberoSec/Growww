# RBI e₹ (Digital Rupee CBDC) & Payment Systems Integration Charter

## 1. Statutory Framework & Scope

- **Regulatory Body:** Reserve Bank of India (RBI), Department of Payment and Settlement Systems (DPSS) & FinTech Department.
- **Governing Legislation:** Reserve Bank of India Act 1934 (as amended in 2022 to include Digital Rupee §2(a-iv)) and Payment and Settlement Systems Act 2007.
- **Objective:** Establishing the operational and cryptographic framework for integrating 24/7 Central Bank Digital Currency (e₹ - Digital Rupee Wholesale `e₹-W` and Retail `e₹-R`) for instantaneous, atomic Delivery-versus-Payment (DvP) settlement on the Growww / NBSE platform.

---

## 2. Technical Architecture & RBI Gateway Interface

```
+------------------------------------+                  +------------------------------------+
| GROWWW / NBSE PLATFORM             |                  | RESERVE BANK OF INDIA CBDC FABRIC  |
+------------------------------------+                  +------------------------------------+
| • services/cbdc-settlement-adapter |   ISO 20022      | • RBI Digital Rupee Core Gateway   |
| • Encrypted mTLS Endpoint          |=================>| • 24/7 Interbank RTGS Liquidity    |
| • Wrapped Token Contract (weINR)   |   mTLS Channel   | • Scheduled Commercial Bank Nodes  |
| • Atomic DvP Settlement Contract   |                  | • Real-Time Wholesale CBDC Ledger  |
+------------------------------------+                  +------------------------------------+
```

### 2.1 1:1 Peg Invariant & Wrapped Digital Rupee (`weINR`)
Every digital rupee represented on the Hyperledger Besu settlement fabric (`weINR`) is strictly backed 1:1 by sovereign digital currency held in the platform's designated CBDC wallet at the Reserve Bank of India:

$$\text{TotalSupply}(weINR) \equiv \text{Balance}_{CBDC}(\text{RBI Vault})$$

### 2.2 Token Life Cycle
1. **Ingress (Lock & Mint):**
   - Participant transfers `e₹-W` or funds via 24/7 RTGS to the platform's escrow account.
   - The CBDC adapter (`services/cbdc-settlement-adapter`) verifies the digital signature and emits a `CBDC_DEPOSIT_CONFIRMED` CloudEvent.
   - The token controller mints an equivalent amount of ERC-3643 compliant `weINR` into the participant's on-chain trading address.
2. **Atomic Settlement:** `weINR` moves atomically across the buyer-seller boundary during the DvP smart contract execution in the same block as the asset token.
3. **Egress (Burn & Unlock):**
   - Participant requests redemption. `weINR` is burned on Hyperledger Besu.
   - The adapter executes an immediate outward `e₹-W` push to the participant's registered bank wallet.

---

## 3. Financial Stability & Systemic Risk Protections

1. **Zero Fractional Reserve:** The platform never lends, rehypothecates, or stakes CBDC reserves. 100% of user balances remain fully funded at all times.
2. **Zero Interest Bearing:** In accordance with RBI guidelines, `weINR` balances earn zero interest, preventing disintermediation of commercial bank deposits.
3. **Circuit Breakers & Outflow Limits:** Automated velocity limits cap daily net CBDC redemptions per entity to prevent bank run cascades during market shocks.
