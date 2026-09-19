# Peer-to-Peer (P2P) Fiat Escrow & Dispute Resolution Specification

This document defines the state machine, escrow lockbox mechanics, payment verification protocols, and arbitration workflows for the Peer-to-Peer (P2P) Fiat-Crypto trading subsystem on the Growww RWA Exchange.

---

## 1. P2P Escrow State Machine & Lockbox Mechanics

### 1.1 Automated Escrow Lockbox
When a maker or taker initiates a P2P trade (e.g. buying USDT with INR via UPI/IMPS):
1. **Collateral Lock:** The seller's crypto (USDT) is immediately deducted from their available balance and locked in the **P2P Escrow Account** (Account 2800).
2. **Payment Window Timer:** A strict countdown timer (default **15 minutes**) begins for the buyer to transfer fiat via their chosen payment method (UPI, IMPS, NetBanking).
3. **Escrow State Transitions:**

```
[1. ORDER_CREATED]
  - Seller crypto locked in Escrow Account (Account 2800)
  - 15-minute payment countdown started
                  |
         +--------+--------+
         |                 | (Buyer taps "Transferred, Notify Seller")
         v                 v
   [EXPIRED / CANCELLED]  [2. PAYMENT_MARKED_PAID]
   - If timer hits 0:       - Seller must confirm receipt within 30 min
     funds unlocked         - Buyer uploads bank payment UTR / reference
     back to seller        |
                  +--------+--------+
                  |                 | (Seller taps "I have received payment")
                  v                 v
          [3. DISPUTE_OPENED]      [4. COMPLETED & RELEASED]
          - Either party flags       - Escrow unlocks and transfers crypto
            non-receipt or fraud       to Buyer available trading balance
          - Funds remain locked      - Status = 'SUCCESS'
```

---

## 2. Fraud Prevention & Bank Payment Invariants

### 2.1 Name Matching Invariant
- The buyer's verified bank account / UPI VPA holder name **MUST exactly match** the KYC-verified name on their Growww account.
- **Third-Party Payments Prohibited:** Payments from accounts belonging to third parties are strictly rejected.

### 2.2 Maker-Checker Operational Arbitration (RUNBOOK-27)
If a dispute is opened:
1. **Evidence Ingestion:** Buyer submits bank statement PDF, UPI transaction screenshot, and UTR number; Seller submits bank statement showing non-receipt.
2. **Automated Verification:** The system queries banking aggregator APIs or Open Banking account aggregators to verify UTR settlement.
3. **Arbitration Execution:**
   - If payment verified: Crypto is force-released to buyer; seller reputation score penalized.
   - If payment false/unpaid: Crypto is returned to seller; buyer account banned for fraud.
