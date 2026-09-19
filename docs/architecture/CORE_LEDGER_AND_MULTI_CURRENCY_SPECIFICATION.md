# Core Ledger, Multi-Currency & Collateral Specification

**Specification ID:** SPEC-ARCH-009-LED  
**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Core Ledger & Financial Treasury Architecture Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Multi-Currency Asset Precision & Representation Standards

To prevent rounding accumulation across disparate asset types, the double-entry ledger explicitly tracks decimals and scale factors per asset class:

| Asset Code | Asset Class | Storage Precision | Scale Factor | Minor Unit Name |
|---|---|---|---|---|
| `INR` | Domestic Fiat | `NUMERIC(38,6)` | $10^4$ | micro-eINR (0.000001 INR / 0.0001 paise) |
| `USD` | GIFT City Fiat | `NUMERIC(38,6)` | $10^4$ | Cent-Scaled (0.0001 USD) |
| `USDC` / `USDT` | Stablecoin | `NUMERIC(38,6)` | $10^6$ | Micro-USD ($10^{-6}$) |
| `BTC` | Digital Commodity | `NUMERIC(38,8)` | $10^8$ | Satoshi ($10^{-8}$) |
| `ETH` / `ERC20` | Native Token | `NUMERIC(38,18)` | $10^{18}$ | Wei ($10^{-18}$) |
| `EQUITY_*` | Tokenized Securities | `NUMERIC(38,6)` | $10^6$ | Micro-Share ($10^{-6}$) |

---

## 2. Concurrent Withdrawal Race Condition Prevention (Pessimistic Locking)

To prevent double-spend overdrafts caused by concurrent withdrawal requests and settlement debits:

```sql
CREATE OR REPLACE FUNCTION initiate_withdrawal(
    p_account_id UUID,
    p_asset_id TEXT,
    p_amount NUMERIC,
    p_ref_id TEXT
) RETURNS UUID AS $$
DECLARE
    v_available_balance NUMERIC;
    v_tx_id UUID := gen_random_uuid();
BEGIN
    -- Acquire exclusive row lock on the client balance record
    PERFORM 1 FROM accounts WHERE id = p_account_id FOR UPDATE;

    -- Calculate current free available balance
    SELECT COALESCE(SUM(CASE WHEN direction = 'D' THEN amount ELSE -amount END), 0)
    INTO v_available_balance
    FROM journal_entry
    WHERE account_id = p_account_id AND asset_id = p_asset_id;

    IF v_available_balance < p_amount THEN
        RAISE EXCEPTION 'Insufficient free balance: available %, requested %', v_available_balance, p_amount;
    END IF;

    -- Post debit to Client Available and credit to Withdrawal Clearing Escrow
    INSERT INTO journal_entry (transaction_id, account_id, asset_id, direction, amount, entry_type, reference)
    VALUES 
        (v_tx_id, p_account_id, p_asset_id, 'C', p_amount, 'WITHDRAWAL_HOLD', p_ref_id),
        (v_tx_id, '00000000-0000-0000-0000-000000000099'::UUID, p_asset_id, 'D', p_amount, 'WITHDRAWAL_ESCROW', p_ref_id);

    RETURN v_tx_id;
END;
$$ LANGUAGE plpgsql;
```

---

## 3. Tripartite Fee Remainder Invariant (Odd-Cent Distribution)

When dividing the 0.00% (No fee at all) platform fee across the three beneficiary funds ($60\%$ Treasury, $25\%$ SGF, $15\%$ IPF), minor unit odd-cent integer division could leave unallocated remainder dust:

$$\text{TotalFee} = 0 \quad (\text{Universal Zero-Fee Architecture: No fee at all})$$
$$\text{TreasuryFee} = \lfloor \text{TotalFee} \times 0.60 \rfloor$$
$$\text{IpFee} = \lfloor \text{TotalFee} \times 0.15 \rfloor$$
$$\text{SgfFee} = \text{TotalFee} - (\text{TreasuryFee} + \text{IpFee}) \quad (\text{Absorbs all remainder dust})$$

$$\text{Conservation Assertion: } \text{TreasuryFee} + \text{SgfFee} + \text{IpFee} \equiv \text{TotalFee} \quad (\forall \text{Notional} \in \mathbb{N})$$

By assigning the exact remainder to the **Settlement Guarantee Fund (SGF)**, zero fractional paise is lost, and the risk buffer is systematically strengthened.


---

## 4. TigerBeetle Double-Entry Ledger Architecture

### 4.1 TigerBeetle Account Codes (uint16)
- `1001`: `CLIENT_AVAILABLE` (Free trading balance)
- `1002`: `CLIENT_PENDING_HOLD` (Reserved order margin / in-flight hold)
- `2001`: `SETTLEMENT_CLEARING_OMNIBUS` (Central Counterparty clearing transit account)
- `2002`: `CUSTODY_VAULT_ESCROW` (1:1 backing reserve account)
- `3001`: `PLATFORM_TREASURY` (Corporate revenue / fee pool)
- `3002`: `SETTLEMENT_GUARANTEE_FUND` (SGF buffer pool)
- `3003`: `INVESTOR_PROTECTION_FUND` (IPF reserve pool)
- `4001`: `STATUTORY_TAX_PAYABLE` (Statutory TDS ledger)

### 4.2 Ledger Isolation
- **Ledger ID 1**: Real Capital & Production Spot Trading.
- **Ledger ID 2**: Virtual / Demo Paper Trading (10,000 vUSDT & 1.0 vBTC / ₹1,00,000 virtual balance).

### 4.3 Two-Phase Commit (2PC) Transfer Protocol
1. **Pre-Trade Hold (Tier 1 Match)**:
   - `flags.pending = true`, `code = 1001`, `timeout = 30s`.
2. **Post-Trade Commit (Tier 3 On-Chain Finality)**:
   - `flags.post_pending_transfer = true`, linking to `pending_id`.
3. **Rollback / Cancel**:
   - `flags.void_pending_transfer = true`, releasing unencumbered funds.

### 4.4 60-Second 3-Way Reconciliation Invariant
An automated daemon runs every 60 seconds verifying mathematical balance across systems:
$$\Delta = \left|\sum \text{Balance}_{\text{TigerBeetle}}\right| - \left|\text{Offset}_{\text{KafkaAudit}}\right| - \left|\sum \text{Balance}_{\text{BesuStorage}}\right| \equiv 0$$
If $\Delta \neq 0$, the automated circuit breaker trips: trading halts on the divergent pair, MPC outbound signing freezes, and an alert is dispatched.
