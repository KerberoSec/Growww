# SEBI Regulatory Sandbox Application: Growww / NBSE Tokenized Securities & 24/7 Trading Pilot

## 1. Applicant Details & Corporate Structure

- **Applicant Name:** Growww Technologies India Private Limited (in consortium with NBSE Limited)
- **Registration Categories:** Registered Stock Broker (SEBI Reg No: INZ000301838) & Depository Participant (CDSL/NSDL)
- **Sandbox Target Cohort:** FinTech Regulatory Sandbox - Stage II (Live Testing with Limited Users)
- **Legal Entity Segmentation:**
  - **Phase 1 Entity:** Growww Technologies India Pvt. Ltd. (Domestic SEBI Broker/Dealer, Client Onboarding, Order Ingress).
  - **Phase 2 Entity:** NBSE Ltd. (Recognized Stock Exchange & Clearing Corporation FMI under Securities Contracts Regulation Act 1956 §4).

---

## 2. Innovative Product Proposition

The applicant proposes to pilot a 24/7 sovereign financial market infrastructure utilizing permissioned enterprise distributed ledger technology (Hyperledger Besu QBFT) to address liquidity fragmentation, extended settlement cycles, and after-hours volatility in Indian capital markets.

### Key Innovations Tested
1. **Continuous 24/7/365 Trading:** Secondary trading in tokenized sovereign debt (G-Secs, T-Bills) and fractional Indian equities during non-standard market hours (evenings and weekends).
2. **Single Universal Flat 0.00% transaction fee (No fee at all):** Elimination of tier-based brokerage, gas surcharges, and account maintenance fees in favor of a uniform, deterministic 0.00% fee (no fee at all) model.
3. **Atomic Delivery-versus-Payment (DvP) Settlement:** Instantaneous, single-block finality ($< 2.0\text{ seconds}$) replacing $T+1 / T+0$ settlement batch delays.
4. **500μs Asymmetric Speed Bump:** Structural latency barrier protecting retail investors from predatory high-frequency front-running.
5. **Simulated Paper Trading Sandbox:** Pre-onboarding demo environment operating against live market data feeds to educate retail investors prior to committing capital.

---

## 3. Regulatory Exemptions Requested

To conduct the live sandbox pilot, temporary statutory exemptions are requested from specific provisions of SEBI regulations:

| Regulation | Existing Requirement | Sandbox Relief Requested | Safeguard & Risk Mitigation |
| :--- | :--- | :--- | :--- |
| **SEBI (Stock Brokers) Reg 1992, Sch III** | Strict trading session hours (09:15 to 15:30 IST) | Permission to match trades 24/7 on internal crossing book | Dynamic price bands ($\pm 3\%, \pm 5\%$), pre-open call auctions, weekend margin freeze |
| **SEBI Master Circular on Settlement** | End-of-day netting and batch depository settlement | Atomic on-chain DvP with real-time balance reservations | 100% pre-funded collateral and strict 1:1 demat balance parity |
| **SEBI Core SGF Guidelines** | Static broker clearing fund contributions | Dynamic fee split: 25% of all transaction fees routed directly to SGF contract | Autonomous on-chain SGF capitalization and insurance buffer |

---

## 4. Investor Protection & Risk Containment

### 4.1 Investor Cohort Limits During Sandbox
- **Retail Cohort Cap:** Maximum 10,000 verified resident Indian investors.
- **Maximum Investment Per Investor:** Capped at ₹10,00,000 (Ten Lakh Rupees) total portfolio exposure.
- **Accreditation:** Completion of mandatory digital risk disclosure module and 10 simulated paper trades in Demo Mode before real-money activation.

### 4.2 Settlement Guarantee Fund (SGF) Waterfall
1. Tier 1: Defaulter's initial margin and pledged collateral.
2. Tier 2: Defaulter's contribution to SGF.
3. Tier 3: NBSE Core Settlement Guarantee Fund (funded by 25% platform fee allocations).
4. Tier 4: NBSE Corporate Treasury capital buffer (minimum ₹50 Crores escrowed in scheduled commercial bank).

---

## 5. Exit & Transition Strategy

Upon successful completion of the 12-month sandbox period:
- Full migration to formal Securities Contracts (Regulation) Act 1956 Section 4 recognition as an operational stock exchange.
- In the event of pilot termination, all open on-chain tokens are redeemed 1:1 for standard demat shares in NSDL/CDSL without investor penalty.
