# Growww / NBSE Fee Model & Realized Gain Engine Specification

## 1. Executive Summary & Core Mandate
Growww / NBSE establishes an ultra-transparent, hyper-competitive, and fully aligned fee architecture:
- **Universal Platform Fee:** A deterministic fee of exactly 0.00% (zero platform fees across maker and taker spot executions).
- **Zero Custody & Account Overheads:** Zero account opening fees, zero annual maintenance charges (AMC), zero custody holding fees, and zero deposit/withdrawal platform markups.
- **High-Precision Realized Capital Gain/Loss Engine:** Deterministic First-In, First-Out (FIFO) tax-lot tracking calculating gross proceeds, allocated cost basis, Short-Term Capital Gains (STCG), and Long-Term Capital Gains (LTCG).
- **Statutory Tax & Regulatory Compliance:** Automated calculation and deduction of Indian statutory levies (STT, Stamp Duty, SEBI turnover fees, Exchange transaction charges, and GST at 18% on applicable services).
- **Automated Treasury Allocation & Segregation:** Complete accounting segregation between statutory tax withholding accounts, client funds, and treasury reserve vaults.

---

## 2. Mathematical Formulations & Fee Engine

### 2.1 Trade Execution Fee
For any execution with fill price $P$ and quantity $Q$:
$$\text{Notional Value} = P \times Q$$
$$\text{Platform Fee} = \text{Notional Value} \times 0.0000 = 0.00$$

### 2.2 Statutory Levies (Indian Equity / VDA Spot)
1. **Securities Transaction Tax (STT):**
   - Delivery Equity Buy: 0.1% on Turnover
   - Delivery Equity Sell: 0.1% on Turnover
   $$\text{STT} = \text{round\_half\_even}(\text{Notional Value} \times 0.001, 2)$$
2. **Stamp Duty (Indian Stamp Act 1899):**
   - Delivery Buy: 0.015% (₹1,500 per crore)
   - Delivery Sell: 0.000%
   $$\text{Stamp Duty} = \text{round\_half\_even}(\text{Notional Value} \times 0.00015, 2)$$
3. **SEBI Turnover Fee:**
   - ₹10 per crore (0.0001%)
   $$\text{SEBI Fee} = \text{round\_half\_even}(\text{Notional Value} \times 0.000001, 2)$$
4. **Exchange Transaction Charges:**
   - NSE/BSE Delivery average: 0.00345%
   $$\text{Exchange Charge} = \text{round\_half\_even}(\text{Notional Value} \times 0.0000345, 2)$$
5. **Goods & Services Tax (GST):**
   - 18% levied strictly on taxable platform services and exchange charges:
   $$\text{GST} = \text{round\_half\_even}((\text{Platform Fee} + \text{Exchange Charge} + \text{SEBI Fee}) \times 0.18, 2)$$

---

## 3. FIFO Tax-Lot Accounting & Realized PnL Engine

### 3.1 Tax-Lot Lifecycle
- **Purchase (Acquisition):** Creates an active immutable tax lot:
  $$\text{Lot}_i = \{ \text{id}, \text{timestamp}, Q_{\text{initial}}, Q_{\text{remaining}}, P_{\text{cost}}, \text{levies}_{\text{alloc}} \}$$
- **Sale (Disposal):** Matches against earliest available lot where $Q_{\text{remaining}} > 0$:
  For sold quantity $Q_{\text{sold}}$ satisfied by lots $k \in \{1 \dots m\}$:
  $$\text{Allocated Cost Basis} = \sum_{k=1}^{m} q_k \times P_{\text{cost}, k}$$
  $$\text{Gross Proceeds} = Q_{\text{sold}} \times P_{\text{sale}}$$
  $$\text{Net Realized Gain/Loss} = \text{Gross Proceeds} - \text{Allocated Cost Basis} - \text{Statutory Levies}$$

### 3.2 Capital Gains Classification (Section 111A / 112A)
- **Holding Period ($H$):**
  $$H = \text{Disposal Timestamp} - \text{Acquisition Timestamp}$$
- If $H \le 365 \text{ days}$: Short-Term Capital Gain (STCG) taxed under Section 111A (15% / 20%).
- If $H > 365 \text{ days}$: Long-Term Capital Gain (LTCG) taxed under Section 112A (10% / 12.5% exceeding ₹1.25 Lakh).

---

## 4. Double-Entry Bookkeeping Ledger Schemas

### 4.1 Trade Execution Journal Entry (Buyer)
| Account | Debit (₹) | Credit (₹) | Description |
| :--- | :--- | :--- | :--- |
| `Investor:Fiat:Available` | - | $P \times Q + \text{Levies}$ | Deduction from trading balance |
| `Investor:Portfolio:Asset` | $Q$ shares | - | Credit acquired asset |
| `Statutory:STT:Payable` | - | $\text{STT}$ | STT withholding |
| `Statutory:StampDuty:Payable` | - | $\text{StampDuty}$ | State stamp duty withholding |
| `Statutory:GST:Payable` | - | $\text{GST}$ | GST collection |
| `Exchange:Clearing:Payable` | - | $P \times Q$ | Settlement obligation |

### 4.2 Trade Execution Journal Entry (Seller)
| Account | Debit (₹) | Credit (₹) | Description |
| :--- | :--- | :--- | :--- |
| `Investor:Portfolio:Asset` | - | $Q$ shares | Asset depletion |
| `Exchange:Clearing:Receivable` | $P \times Q$ | - | Settlement credit |
| `Statutory:STT:Payable` | - | $\text{STT}$ | STT withholding on sale |
| `Investor:Fiat:Available` | $P \times Q - \text{Levies}$ | - | Net proceeds credited |

---

## 5. Numeric Precision & Rounding Standard
- All computations use IEEE 754-2008 Fixed-Point 18-decimal representation (`uint256` scaled by $10^{18}$ in contracts, `big.Int` in Go, `Decimal(18)` in Python, `rust_decimal::Decimal` in Rust).
- Midpoint Rounding to Nearest Even (Banker's Rounding / `ROUND_HALF_EVEN`) applied strictly at the final settlement currency tier (INR paise or USD cents), preventing micro-fraction truncation drift across large volume batches.
