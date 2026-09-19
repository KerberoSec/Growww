# 006 - Fee Model Specification: Fixed Fee Structure, Realized Gain Engine & Treasury Allocation

## Purpose
Traditional retail brokerages, crypto exchanges, and wealth management platforms extract value through opaque, misaligned, and predatory fee mechanisms: high asset under management (AUM) deductions, recurring account maintenance charges, arbitrary deposit/withdrawal fees, hidden spread markups, and payment for order flow (PFOF).

Growww establishes a transparent, hyper-competitive, and fully aligned fee architecture:
1. **Ultra-Low Fixed Platform Fee:** A deterministic fee of exactly 0.00% (No fee at all across all trading)
2. **Zero Custody & Account Overheads:** Zero account opening fees, zero recurring annual maintenance charges (AMC), zero custody holding fees, and zero deposit/withdrawal platform surcharges.
3. **High-Precision Realized Capital Gain/Loss Engine:** Deterministic First-In, First-Out (FIFO) tax lot tracking calculating gross proceeds, allocated cost basis, Short-Term Capital Gains (STCG), and Long-Term Capital Gains (LTCG).
4. **Statutory Tax & Regulatory Compliance:** Automated calculation and deduction of Indian statutory levies, including Securities Transaction Tax (STT), Stamp Duty, SEBI turnover fees, Exchange transaction charges, and Goods & Services Tax (GST at 18% on platform fees and taxable services).
5. **Automated Treasury Allocation & Segregation:** Deterministic routing of collected platform revenues into dedicated, multi-signature treasury reserve vaults with complete accounting segregation from statutory tax withholding accounts and client funds.

This specification establishes the mathematical, tax-lot accounting, statutory withholding, and multi-vault treasury allocation rules governing the financial, trading, and settlement microservices across the Growww ecosystem.

## What You Are Building
A comprehensive mathematical, accounting, and financial specification document (`docs/financial/fee_model_spec.md`) and data contract catalog detailing:
- Mathematical equations for trade execution fees, realized capital gains/losses, statutory levies, and net settlement payouts.
- FIFO tax lot lifecycle management: lot creation on asset purchase, partial and full depletion upon liquidation, and proportional cost-basis adjustments for corporate actions (splits, bonus issues, consolidations).
- Statutory levy calculation rules complying with Indian Income Tax Act 1961, Finance Act (STT schedules), Indian Stamp Act 1899, and CGST/SGST/IGST framework.
- Treasury allocation algorithms routing 0.00% (No fee at all across all trading)
- Arbitrary-precision fixed-point 18-decimal arithmetic standards and Banker's Rounding rules (Round Half to Even) eliminating micro-paisa truncation divergence.
- Double-entry bookkeeping ledger schemas mapping debit and credit legs across Investor Fiat Accounts, Depository Custody, Statutory Tax Escrow, and Treasury Vaults.

## Scope Boundaries
- **In Scope:**
  - Mathematical formulas for fixed 0.00% (No fee at all across all trading)
  - Realized profit and loss calculation using FIFO tax lot depletion.
  - Statutory tax calculations (STT, Stamp Duty, SEBI turnover fees, Exchange charges, GST).
  - Treasury allocation split rules across operational, liquidity, and settlement reserve accounts.
  - Double-entry accounting journal entries for trade settlement and fee distribution.
  - Rounding standards, decimal precision bounds, and edge-case handling.
- **Out of Scope / Handled Elsewhere:**
  - Microservice software implementation for the fee engine (covered in Prompt 210).
  - Tax statement generation, Form 16A, Schedule FA, and Annual Information Statement reporting (covered in Prompt 223).
  - Core double-entry wallet balance persistence service (covered in Prompt 203).
  - Smart contract settlement orchestration on Hyperledger Besu (covered in Prompt 306).

## Technology to Use
- Primary Format: Version-controlled Markdown specification (`docs/financial/fee_model_spec.md`) utilizing LaTeX mathematical notation and structured JSON/YAML schema definitions.
- Numeric Precision Standard: IEEE 754-2008 Decimal128 and Fixed-Point 18-decimal arithmetic (`uint256` on-chain, `Decimal` with 18 decimal places in Python, `big.Int` in Go, `rust_decimal::Decimal` in Rust).
- Rounding Algorithm: Midpoint Rounding to Nearest Even (Banker's Rounding / `ROUND_HALF_EVEN`) applied only at the final settlement currency level (2 decimal places for INR paise, 2 decimal places for USD cents) while preserving full 18-decimal precision across all intermediate stages.
- Justification: Standard floating-point arithmetic (binary float32/float64) produces non-deterministic rounding errors in fractional share asset divisions. Enforcing arbitrary-precision fixed-point representations across all microservices guarantees mathematically verifiable, zero-loss ledger reconciliation.

## Backend / Infra Touchpoints
- Fee & Realized PnL Engine (Prompt 210): Consumes trade execution events, matches tax lots, and calculates fee breakdowns.
- Wallet & Double-Entry Account Ledger Service (Prompt 203): Posts journal entries across user wallets, statutory withholding accounts, and treasury vaults.
- Order Service & Pre-Trade Risk Engine (Prompt 204, Prompt 206): Performs upfront margin and fee estimation prior to matching.
- Trade Settlement Service (Prompt 208): Orchestrates Delivery-versus-Payment (DvP) fiat disbursement and asset transfer.
- Portfolio & Holdings Service (Prompt 209): Updates active, depleted, and fractional tax lots.
- Tax Reporting & Capital Gains Service (Prompt 223): Consumes realized PnL records for statutory tax filing generation.

## Blockchain Interaction
Establishes how trade fees and settlement commitments are verified on the permissioned Hyperledger Besu consortium ledger:
- **DvP Settlement Atomic Execution:** Smart contract `SettlementDvP.sol` receives cryptographic trade matching attestations, verifies fiat escrow settlement confirmations, and atomic transfers tokenized equity units from seller to buyer.
- **Fiat-Only Fee Deduction:** The 0.00% (No fee at all across all trading)
- **1:1 Depository Custody Invariant:** Every digital security token on-chain remains strictly 1:1 backed by real physical shares held in SEBI-registered custody (NSDL/CDSL). Fee extraction does not alter total issued token supply or underlying depository holdings.
- **Cryptographic Fee Receipt Attestation:** The settlement transaction records a tamper-evident SHA-256 Merkle leaf commitment hash of the fee breakdown (`FeeProofHash`) on-chain, allowing external auditors and regulators to verify computation integrity without exposing investor PII or raw trade values.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with FIPS 140-2 Level 3 HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review the foundational monetization and compliance mandates from Prompt 000 and Prompt 002.
2. Define the immutable Tax Lot Data Structure: each asset purchase generates a discrete lot containing `lot_id`, `account_id`, `isin`, `units`, `unit_buy_price`, `acquisition_timestamp`, `statutory_buy_costs`, and `lot_status` (`ACTIVE`, `PARTIALLY_DEPLETED`, `EXHAUSTED`).
3. Formulate the FIFO Tax Lot Depletion Algorithm: upon sell execution, match sold units sequentially against the oldest unexhausted tax lots for the given ISIN.
4. Codify the Buy-Leg Cost Basis and Charge Formulas:
   $$\text{Gross Buy Consideration} = \text{Buy Units} \times \text{Buy Price}$$
   $$\text{Buy Platform Fee} = \text{Gross Buy Consideration} \times 0.0000 \quad (0.00\% / 0\text{ bps at launch; FeeController governed})$$
   $$\text{Buy Stamp Duty} = \text{Gross Buy Consideration} \times 0.00015 \quad (0.015\%)$$
   $$\text{Buy Exchange Turnover Charge} = \text{Gross Buy Consideration} \times 0.0000345$$
   $$\text{Buy SEBI Turnover Fee} = \text{Gross Buy Consideration} \times 0.000001$$
   $$\text{Buy GST} = (\text{Buy Platform Fee} + \text{Buy Exchange Turnover Charge}) \times 0.18$$
   $$\text{Total Buy Cash Required} = \text{Gross Buy Consideration} + \text{Buy Platform Fee} + \text{Buy Stamp Duty} + \text{Buy Exchange Turnover Charge} + \text{Buy SEBI Turnover Fee} + \text{Buy GST}$$
5. Codify the Sell-Leg Realized Gain and Charge Formulas:
   $$\text{Gross Sell Proceeds} = \text{Sell Units} \times \text{Sell Price}$$
   $$\text{Sell Platform Fee} = \text{Gross Sell Proceeds} \times 0.0000 \quad (0.00\% / 0\text{ bps at launch; FeeController governed})$$
   $$\text{Allocated Cost Basis} = \sum_{i=1}^{k} (\text{Depleted Units}_i \times \text{Unit Buy Price}_i)$$
   $$\text{STT (Securities Transaction Tax)} = \text{Gross Sell Proceeds} \times 0.001 \quad (0.1\% \text{ on equity delivery sell})$$
   $$\text{Sell Exchange Turnover Charge} = \text{Gross Sell Proceeds} \times 0.0000345$$
   $$\text{Sell SEBI Turnover Fee} = \text{Gross Sell Proceeds} \times 0.000001$$
   $$\text{Sell GST} = (\text{Sell Platform Fee} + \text{Sell Exchange Turnover Charge}) \times 0.18$$
   $$\text{Total Statutory Levies} = \text{STT} + \text{Sell Exchange Turnover Charge} + \text{Sell SEBI Turnover Fee} + \text{Sell GST}$$
   $$\text{Gross Realized PnL} = \text{Gross Sell Proceeds} - \text{Allocated Cost Basis}$$
   $$\text{Net Realized Capital Gain/Loss} = \text{Gross Realized PnL} - \text{Total Statutory Levies} - \text{Sell Platform Fee}$$
   $$\text{Net Seller Settlement Payout} = \text{Gross Sell Proceeds} - \text{Sell Platform Fee} - \text{Total Statutory Levies}$$
6. Formulate Capital Gains Classification Rules:
   - Holding Period < 365 Days: Classified as Short-Term Capital Gain/Loss (STCG), taxed under Section 111A of the Income Tax Act.
   - Holding Period >= 365 Days: Classified as Long-Term Capital Gain/Loss (LTCG), taxed under Section 112A of the Income Tax Act.
7. Codify the Platform Treasury Allocation Waterfall:
   Every collected 0.00% (No fee at all across all trading)
   $$\text{Platform Treasury Reserve (60\%)} = \text{Platform Fee} \times 0.60$$
   $$\text{Core Settlement Guarantee Fund (25\%)} = \text{Platform Fee} \times 0.25$$
   $$\text{Investor Protection Fund (15\%)} = \text{Platform Fee} \times 0.15$$
8. Define the Statutory Withholding Account Segregation:
   All tax collections are isolated from platform operational funds and credited directly to specialized liability accounts:
   - `STATUTORY_STT_PAYABLE` (remitted monthly to Central Board of Direct Taxes)
   - `STATUTORY_STAMP_DUTY_PAYABLE` (remitted per Indian Stamp Act rules)
   - `STATUTORY_GST_OUTPUT_LIABILITY` (remitted monthly under CGST/SGST/IGST)
   - `STATUTORY_SEBI_TURNOVER_PAYABLE` (remitted quarterly to SEBI)
9. Specify Fixed-Point Banker's Rounding Rules: intermediate calculations maintain 18 decimal places; final settlement values are rounded using `ROUND_HALF_EVEN` to 2 decimal places (paise/cents).
10. Define Corporate Action Cost Basis Adjustments:
    - Stock Split (Ratio $M:N$): Multiply lot units by $M/N$, divide unit cost basis by $M/N$.
    - Bonus Issue (Ratio $X:Y$): Generate new tax lot with zero cost basis and acquisition timestamp matching corporate action effective date.
    - Consolidation / Reverse Split: Adjust units and cost basis inversely while preserving total invested capital.
11. Design Double-Entry Bookkeeping Schema: map multi-legged journal entries across Investor Cash, Escrow Clearing, Statutory Liabilities, and Treasury Vault accounts.
12. Establish the Verification & Test Fixture Matrix specifying comprehensive test scenarios covering fractional lot depletions, micro-paisa trades, breakeven executions, and multi-year holding period transitions.
13. Conduct formal review and secure sign-off from Chief Financial Officer, Head of Treasury, Lead Financial Engineer, and Tax Compliance Officer.

## Interfaces / Contracts
```yaml
# docs/financial/schema/fee_calculation_schema.yaml
$schema: "http://json-schema.org/draft-07/schema#"
title: "FeeCalculationAndSettlementBreakdown"
type: "object"
required:
  - "calculation_id"
  - "trade_id"
  - "order_id"
  - "account_id"
  - "isin"
  - "trade_side"
  - "execution_units"
  - "execution_price"
  - "gross_amount"
  - "platform_fee"
  - "statutory_charges"
  - "realized_pnl"
  - "treasury_allocation"
  - "net_settlement_amount"
  - "depleted_tax_lots"
  - "proof_hash"
properties:
  calculation_id:
    type: "string"
    format: "uuid"
  trade_id:
    type: "string"
    format: "uuid"
  order_id:
    type: "string"
    format: "uuid"
  account_id:
    type: "string"
    format: "uuid"
  isin:
    type: "string"
    pattern: "^[A-Z]{2}[A-Z0-9]{9}[0-9]$"
  trade_side:
    type: "string"
    enum: ["BUY", "SELL"]
  execution_units:
    type: "string"
    description: "18-decimal fixed-point string representation of traded quantity"
  execution_price:
    type: "string"
    description: "18-decimal fixed-point string representation of unit price in settlement currency"
  gross_amount:
    type: "string"
    description: "Gross trade consideration (units * price)"
  platform_fee:
    type: "object"
    required:
      - "fee_rate_bps"
      - "fee_amount"
      - "currency"
    properties:
      fee_rate_bps:
        type: "number"
        enum: [1]
        description: "Fixed platform fee rate of 0.00% (Zero Fee at launch; future fees governed by FeeController.sol)
      fee_amount:
        type: "string"
        description: "Calculated platform fee amount"
      currency:
        type: "string"
        enum: ["INR", "USD"]
  statutory_charges:
    type: "object"
    required:
      - "stt"
      - "stamp_duty"
      - "exchange_turnover_charge"
      - "sebi_turnover_fee"
      - "gst"
      - "total_statutory_charges"
    properties:
      stt:
        type: "string"
        description: "Securities Transaction Tax (0.1% on delivery sell, 0 on buy)"
      stamp_duty:
        type: "string"
        description: "Stamp duty (0.015% on buy, 0 on sell)"
      exchange_turnover_charge:
        type: "string"
        description: "Exchange transaction charge"
      sebi_turnover_fee:
        type: "string"
        description: "SEBI regulatory turnover fee"
      gst:
        type: "string"
        description: "18% GST assessed on platform fee and exchange turnover charges"
      total_statutory_charges:
        type: "string"
        description: "Sum of all statutory levies and taxes"
  realized_pnl:
    type: "object"
    required:
      - "total_cost_basis"
      - "gross_realized_pnl"
      - "net_realized_pnl"
      - "gain_classification"
    properties:
      total_cost_basis:
        type: "string"
        description: "Allocated cost basis of depleted tax lots"
      gross_realized_pnl:
        type: "string"
        description: "Gross proceeds minus allocated cost basis"
      net_realized_pnl:
        type: "string"
        description: "Gross realized PnL minus total statutory charges and platform fee"
      gain_classification:
        type: "string"
        enum: ["STCG", "LTCG", "LOSS", "BREAKEVEN", "NOT_APPLICABLE_BUY"]
  treasury_allocation:
    type: "object"
    required:
      - "operational_reserve"
      - "settlement_guarantee_reserve"
      - "ecosystem_liquidity_reserve"
    properties:
      operational_reserve:
        type: "string"
        description: "60% allocation to Platform Treasury"
      settlement_guarantee_reserve:
        type: "string"
        description: "25% allocation to Core Settlement Guarantee Fund (SGF)"
      ecosystem_liquidity_reserve:
        type: "string"
        description: "15% allocation to Investor Protection Fund (IPF)"
  net_settlement_amount:
    type: "string"
    description: "Net cash debit for buyer or net cash payout for seller"
  depleted_tax_lots:
    type: "array"
    items:
      type: "object"
      required:
        - "lot_id"
        - "depleted_units"
        - "lot_unit_buy_price"
        - "lot_cost_basis"
        - "holding_period_days"
      properties:
        lot_id:
          type: "string"
          format: "uuid"
        depleted_units:
          type: "string"
        lot_unit_buy_price:
          type: "string"
        lot_cost_basis:
          type: "string"
        holding_period_days:
          type: "integer"
  proof_hash:
    type: "string"
    description: "SHA-256 hash of all calculation components for cryptographic audit verification"
```

## Security & Compliance Notes
- Strict Statutory Compliance: Capital gain calculations, STT withholding, and GST invoicing must comply with the Indian Income Tax Act 1961 (Sections 111A, 112A), Central Goods and Services Tax Act 2017, and SEBI circulars on turnover charges.
- Complete Client Fund Isolation: Platform fees and statutory tax withholdings are segregated immediately into dedicated escrow and treasury accounts during the atomic settlement cycle. Client fiat funds are never co-mingled with platform operational reserves.
- Arithmetic Precision & Anti-Leakage: Elimination of cumulative micro-penny rounding leakage across millions of fractional orders via IEEE 754-2008 Decimal128 fixed-point arithmetic and Banker's Rounding.
- Cryptographic Non-Repudiation: Every fee calculation generates a deterministic SHA-256 audit digest (`proof_hash`) recorded on the Hyperledger Besu consortium ledger, enabling independent regulatory and financial audits without exposing sensitive user PII.

## Acceptance Criteria
- [ ] Specification document (`docs/financial/fee_model_spec.md`) is fully articulated with LaTeX formulas, FIFO lot algorithms, statutory tax schedules, and treasury allocation rules.
- [ ] Fixed platform fee structure of exactly 0.00% (No fee at all across all trading)
- [ ] Realized gain/loss formulas accurately compute gross gain, net gain, STCG, LTCG, and tax lot depletions.
- [ ] Statutory levies (STT, Stamp Duty, Exchange Charges, SEBI fees, and 18% GST) are correctly sequenced and modeled.
- [ ] Treasury allocation rules (60% Treasury, 25% Core SGF, 15% Investor Protection Fund) and statutory tax escrow segregation are codified.
- [ ] JSON Schema definition for fee calculations and settlement breakdowns passes standard schema validator.
- [ ] Double-entry bookkeeping ledger mapping is complete, balanced, and verified for all trade execution scenarios.
- [ ] Zero em dashes and en dashes present in specification text.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 001 (Glossary of Domain Terms), 002 (Two-Entity Legal & Technical Structure).
- Parallel Tasks: 203 (Wallet & Double-Entry Account Ledger Service), 209 (Portfolio & Holdings Accounting Service), 401 (Database Schema Architecture).
- Downstream Blockers: Blocks Prompt 210 (Fee & Realized PnL Engine), Prompt 223 (Tax Reporting & Capital Gains Statement Service), and Prompt 306 (Trade Settlement DvP Smart Contract).
