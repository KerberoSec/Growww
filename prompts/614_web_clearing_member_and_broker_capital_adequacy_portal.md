# 614 - Next.js 14 Clearing Member & Broker Capital Adequacy Portal

## Purpose
In modern electronic and hybrid exchange architectures operating under SEBI (Securities and Exchange Board of India) and IFSCA regulatory frameworks, Central Counterparty (CCP) clearing and settlement integrity depends fundamentally on the financial soundness and capital adequacy of Clearing Members (CMs) and Trading Members (TMs). Clearing Members bear legal counterparty liability for all novated transactions executed by their constituent Trading Members, proprietary trading desks, and direct institutional clients. Under the SEBI Master Circular for Stock Brokers and Clearing Members, the SEBI Upstreaming of Client Funds framework, and CPMI-IOSCO Principles for Financial Market Infrastructures (PFMI), clearing entities must continuously maintain adequate Base Minimum Capital (BMC), enforce a strict 50:50 cash-to-non-cash collateral ratio, monitor intraday peak margin utilization across designated snapshot intervals, execute automated lien-marked Bank Guarantee (e-BG) replenishments, and maintain total segregation between client collateral and proprietary reserves.

This prompt specifies the architecture, user experience, and technical implementation of the **Next.js 14 Clearing Member & Broker Capital Adequacy Portal** (`apps/growww_admin/app/(dashboard)/clearing-member/`). The portal provides Clearing Members, self-clearing Trading Members, and exchange risk officers with a real-time workstation to monitor live capital adequacy ratios (Effective Capital, Net Liquid Capital, and BMC), track peak margin exposure across four daily snapshot windows, reconcile CDSL/NSDL margin pledges and re-pledges, manage electronic Bank Guarantees (e-BG) and Fixed Deposit Receipts (FDRs), audit multi-tiered Settlement Guarantee Fund (SGF) commitments, and verify on-chain tokenized collateral escrowed within Hyperledger Besu smart contracts.

## What You Are Building
An enterprise-grade, high-availability Next.js 14 capital adequacy and risk governance workstation featuring:
- `CapitalAdequacyOverview`: Real-time executive dashboard calculating Net Liquid Capital (NLC), Base Minimum Capital (BMC), Total Deposited Collateral, Haircut-Adjusted Valuation, Free Margin, and Available Margin Utilization Percentage, with visual threshold indicators (Safe: <70%, Advisory Warning: 70-85%, Margin Call: 85-90%, Critical Liquidation Breach: >90%).
- `PeakMarginUtilizationTracker`: Intraday peak margin analytics monitor capturing and visualizing the four statutory SEBI random snapshot intervals per trading session, comparing pre-trade SPAN/VaR requirements against posted member collateral, and alerting on prospective peak margin shortfall penalties.
- `CashCollateralSegregationMonitor`: Real-time ratio compliance monitor enforcing the mandatory 50:50 cash-to-collateral rule (Cash & Cash Equivalents vs Non-Cash Eligible Collateral) and monitoring the upstreaming of client funds to the Clearing Corporation via RBI e-Kuber / CC Clearing House accounts without proprietary commingling.
- `BankGuaranteeManager`: Electronic Bank Guarantee (e-BG) lifecycle portal integrated with National E-Governance Services Ltd (NeSL) and scheduled commercial banks, handling issuance registration, claim expiry tracking, renewal alerts, invocation safeguards, and automated collateral haircut deduction upon approach of expiry (<15 days).
- `MarginPledgeReconciler`: Comprehensive depository pledge and re-pledge reconciliation terminal interfacing with CDSL and NSDL feeds, verifying client-to-TM pledge and TM-to-CM/CC re-pledge chains, validating Demat account segregation, and highlighting unauthorized pledge breaks.
- `MemberRiskLimitTerminal`: Granular Trading Member (TM) risk controller allowing Clearing Members to configure intraday gross exposure multiples, turnover ceilings, margin utilization warning thresholds, and single-click automated order cut-off / square-off triggers upon margin breaches.
- `SettlementGuaranteeFundContributionViewer`: Real-time on-chain and off-chain SGF monitor displaying the Clearing Member's primary Core SGF quota, Cover-1 / Cover-2 extreme stress-test surcharge liabilities, and dynamic default waterfall tranche positions verified via Besu RPC.
- `CollateralDepositWithdrawalModal`: Maker-checker authorization terminal for collateral top-ups (cash via RTGS/virtual accounts, G-Sec/T-Bills pledge, bank guarantees) and margin release withdrawals subject to non-violation of minimum capital adequacy rules.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router administration portal under `apps/growww_admin/app/(dashboard)/clearing-member/`.
  - Real-time WebSocket streaming of member margin utilization, peak snapshots, and collateral valuations.
  - Interactive visualization of the statutory 50:50 cash-to-non-cash collateral ratio and Base Minimum Capital buffers.
  - Electronic Bank Guarantee (e-BG) lifecycle management, NeSL metadata tracking, and renewal workflows.
  - Client Demat pledge / re-pledge reconciliation table with CDSL/NSDL depository discrepancy identification.
  - On-chain Settlement Guarantee Fund (SGF) contribution audit using Viem to read `SettlementGuaranteeFund.sol`.
  - Maker-Checker authorization modals for collateral withdrawal requests and bank guarantee release approvals.
  - High-frequency alert banner and toast notifications for margin utilization crossing 70%, 85%, and 90% risk thresholds.
  - CSV/PDF regulatory capital adequacy report exports formatted to SEBI annexure specifications.
- **Out of Scope / Handled Elsewhere:**
  - Real-time SPAN/VaR margin calculations (handled by Risk Service Prompt 206 and Real-Time VaR Engine Prompt 229).
  - High-throughput depository pledge API orchestration with NSDL/CDSL (handled by Margin Pledge Gateway Prompt 260 / Prompt 263).
  - Off-chain SGF Default Waterfall stress-testing orchestration (handled by SGF Service Prompt 230).
  - On-chain SGF smart contract deployment and state management (handled by Prompt 315).
  - Bank clearing integration via RBI e-Kuber host-to-host banking (handled by Prompt 213 / Prompt 214).
  - Core order matching and pre-trade margin blocking (handled by Prompt 205 and Prompt 206).

## Technology to Use
- **Next.js 14 App Router:** Hybrid architecture leveraging React Server Components (RSC) for initial page hydration, server actions for authenticated mutation workflows, and Client Components for dynamic WebSocket updates.
- **TypeScript 5.x:** Strict type safety across all margin metrics, Protobuf contracts, and depository status enums.
- **Tailwind CSS 3.4+ & shadcn/ui:** Institutional dark/light adaptive dashboard UI styled with clean tabular controls, collapsible panels, and financial data grids.
- **TanStack Table (React Table v8):** High-performance virtualized tables for rendering thousands of client margin pledge records and TM exposure lines without UI lag.
- **Recharts / Tremor:** Responsive data visualization for peak margin intraday timelines, collateral composition pie charts, and historical utilization trends.
- **WebSockets (`native ws` or `@stomp/stompjs`):** Low-latency duplex communication streaming live margin utilization updates and threshold breach alerts from backend risk gateways.
- **PostgreSQL 16 & Prisma/Drizzle:** Persistence layer for storing member configuration, bank guarantee registry records, maker-checker pending proposals, and historical snapshot logs.
- **Redis 7.x:** In-memory caching for sub-millisecond member risk metrics, pub/sub for real-time WebSocket distribution, and rate-limiting.
- **Viem 2.x:** Web3 library for querying on-chain `SettlementGuaranteeFund.sol` contract state and verifying tokenized collateral locks on Hyperledger Besu.
- **Decimal.js:** High-precision arbitrary-precision arithmetic for collateral valuation, haircut calculations, and cash ratio compliance, avoiding IEEE 754 floating-point rounding errors.
- **Zustand with Immer:** Client-side state store for active member selection, margin alert subscriptions, and filter states.

## Backend / Infra Touchpoints
- **Settlement Guarantee Fund Service (Prompt 230):** Connects via gRPC / REST (`/api/v1/clearing/sgf/member/{cmId}`) to retrieve member minimum SGF contribution quotas, stress-test surcharges, and default waterfall standings.
- **Clearing Corporation Interoperability & Margin Pledge Gateway (Prompt 260 / Prompt 263):** Ingests client and TM pledge/re-pledge statuses via `GET /api/v1/clearing/pledges` and streams pledge lifecycle changes (`PLEDGED`, `RE_PLEDGED_TO_CC`, `UNPLEDGE_REQUESTED`, `INVOKED`).
- **Pre-Trade Risk & Margin Checks Service (Prompt 206):** Ingests live pre-trade margin utilization, gross exposure ceilings, and circuit limit evaluations via `wss://api.growww.in/ws/v1/risk/members`.
- **Real-Time VaR & SPAN Margin Engine (Prompt 229, Prompt 241):** Ingests intraday portfolio VaR, Extreme Loss Margin (ELM), and peak margin snapshot calculation batches.
- **Audit Log Service (Prompt 218):** Dispatches immutable audit logs for all administrative actions, risk limit adjustments, collateral release approvals, and manual margin overrides.
- **NeSL & Bank Guarantee Gateway (Prompt 213):** Ingests electronic Bank Guarantee issuance confirmations, digital stamp certificates, and bank API callback webhooks.
- **Hyperledger Besu RPC Node (Prompt 302, Prompt 315):** Directly reads `SettlementGuaranteeFund.sol` and `TokenizedCollateralVault.sol` to verify on-chain collateral commitments.

## Blockchain Interaction
- **Network:** Hyperledger Besu (Permissioned Consortium Network, QBFT Consensus, deterministic 2-second block time).
- **Interfaced Smart Contracts:**
  - `SettlementGuaranteeFund.sol` (Prompt 315): Maintains immutable on-chain records of Clearing Member cash/CBDC and G-Sec allocations across Tier-1 (Member SGF), Tier-2 (Exchange SGF), and Tier-3 (Non-Defaulting Member Tranches). The portal calls:
    - `getMemberContribution(address cmAddress) external view returns (uint256 cashAmount, uint256 gSecValue, uint256 lastUpdatedBlock)`
    - `getWaterfallState() external view returns (uint8 currentTranche, uint256 deficitAmount)`
  - `TokenizedCollateralVault.sol` (Prompt 333 / Prompt 315): Manages on-chain tokenized Government Securities (G-Secs), Treasury Bills (T-Bills), and Sovereign Gold Bonds (SGBs) locked as regulatory capital. The portal reads:
    - `getLockedCollateral(address cmAddress, address tokenContract) external view returns (uint256 lockedUnits, uint256 haircutBps)`
- **Proof-of-Collateral Verification:**
  - The portal provides one-click verification of on-chain collateral balances by hashing the off-chain depository/bank depository receipt against the smart contract's Merkle root, rendering a cryptographic "Verified On-Chain" badge.
- **Zero Exposure of Client PII:**
  - In adherence to DPDP Act 2023 and SEBI regulations, client PAN, Demat account numbers, and retail identity information are cryptographically hashed; only masked identifiers (e.g., `****1234`) and CM/TM clearing codes are displayed or transmitted.

## Step-by-Step Build Instructions

1. **Scaffold Next.js 14 Clearing Member Route Structure:**
   - Create route directory `apps/growww_admin/app/(dashboard)/clearing-member/`.
   - Implement root layout `layout.tsx` with clearing member context provider, persistent risk alert marquee, and navigation header.
   - Configure sub-routes:
     - `page.tsx`: Executive Capital Adequacy & BMC summary dashboard.
     - `peak-margin/page.tsx`: Intraday peak margin tracking and snapshot history.
     - `collateral/page.tsx`: Collateral allocation, 50:50 cash ratio monitor, and deposit/withdrawal terminal.
     - `bank-guarantees/page.tsx`: Electronic Bank Guarantee (e-BG) and FDR lifecycle manager.
     - `pledge-reconciliation/page.tsx`: CDSL/NSDL depository margin pledge and re-pledge reconciliation terminal.
     - `trading-members/page.tsx`: Subordinate Trading Member risk monitoring and limit governance.
     - `sgf-commitments/page.tsx`: Settlement Guarantee Fund contribution and on-chain Besu audit viewer.

2. **Define Type Definitions, Schemas, and Protobuf Stubs:**
   - Establish unified TypeScript models and Protobuf stubs in `packages/types/src/clearing/` for `CapitalAdequacyReport`, `PeakMarginSnapshot`, `BankGuaranteeRecord`, `MarginPledgeItem`, and `SGFContribution`.
   - Create Zod validation schemas for collateral deposit requests, limit modification forms, and maker-checker approval actions.

3. **Construct Real-Time WebSocket Client & Zustand Store:**
   - Build `useClearingMemberSocket` hook in `hooks/use-clearing-member-socket.ts` maintaining an auto-reconnecting WebSocket connection to `wss://api.growww.in/ws/v1/clearing/stream`.
   - Ingest live topics: `cm:metrics:{cmId}`, `cm:peak-margin:{cmId}`, `cm:breach-alerts:{cmId}`, and `cm:pledge-events:{cmId}`.
   - Implement Zustand store `stores/clearing-member-store.ts` with Immer middleware to maintain high-frequency metric state without triggering unnecessary full-tree re-renders.

4. **Implement `CapitalAdequacyOverview` Component:**
   - Calculate and display key regulatory figures:
     - **Base Minimum Capital (BMC):** Mandated non-risk base buffer (₹50 Lakhs for Capital Market + F&O CM).
     - **Total Deposited Collateral:** Aggregate cash, e-BG, FDR, and approved securities.
     - **Effective Collateral (Post-Haircut):** Total collateral after applying SEBI-mandated haircuts (Cash: 0%, G-Sec: 2%, AAA Bonds: 5%, Equities: 20-40%).
     - **Initial Margin (SPAN + VaR) & Extreme Loss Margin (ELM):** Real-time aggregate margin obligation.
     - **Net Liquid Capital (NLC):** $\text{Effective Collateral} - \text{Total Margin Obligation} - \text{BMC}$.
     - **Margin Utilization Percentage:** $(\text{Total Margin Obligation} / \text{Effective Collateral}) \times 100$.
   - Display a dynamic color-coded gauge bar with explicit zones: Green (<70%), Yellow (70-85%), Orange (85-90%), Red (>90%).

5. **Build `PeakMarginUtilizationTracker` Component:**
   - Implement the SEBI peak margin timeline rendering four random snapshot windows per day (Snapshot 1: 09:15-11:00, Snapshot 2: 11:00-12:30, Snapshot 3: 12:30-14:00, Snapshot 4: 14:00-15:30).
   - Render an intraday area chart comparing peak margin requirements against effective capital at each minute.
   - Display peak utilization summary cards indicating the highest recorded peak percentage, the timestamp of occurrence, and projected SEBI short-margin penalty if unaddressed.

6. **Develop `CashCollateralSegregationMonitor` (50:50 Rule):**
   - Implement collateral composition analytics visualizer:
     - Calculate $\text{Cash Component Ratio} = (\text{Cash} + \text{Cash Equivalents}) / \text{Total Eligible Collateral}$.
     - Enforce SEBI rule: At least 50% of effective collateral must consist of Cash / Cash Equivalents (Cash, Bank Guarantee, FDR, T-Bills).
     - If non-cash collateral exceeds 50%, calculate the excess non-cash component and apply an immediate 100% discount on the excess in effective margin calculations.
   - Implement upstreaming of client funds monitor showing exact balances transferred to the Clearing Corporation bank accounts versus balances retained in member settlement accounts.

7. **Implement `BankGuaranteeManager` Component:**
   - Build a searchable, filterable data grid for tracking all active, pending, and expiring Bank Guarantees.
   - Display issuing bank name, IFSC, BG reference number, NeSL e-BG unique identifier, financial/performance type, issuance date, claim expiry date, and face value.
   - Implement dynamic expiry alert badges: Green (>30 days), Yellow (15-30 days), Red (<15 days, highlighting mandatory automated collateral haircut application).
   - Provide an upload and verification modal for registering newly issued NeSL digital e-BGs with digital signature validation.

8. **Build `MarginPledgeReconciler` Component:**
   - Implement a high-performance TanStack virtualized table rendering client collateral pledges across CDSL and NSDL.
   - Columns: Client UCC, Depository (CDSL/NSDL), ISIN, Security Name, Pledged Quantity, Market Price, Haircut %, Post-Haircut Valuation, Pledge Date, Status (`PLEDGED_TO_TM`, `RE_PLEDGED_TO_CM`, `ALLOCATED_TO_CC`).
   - Implement automated discrepancy scanner highlighting reconciliation breaks:
     - Mismatch between depository ledger and clearing house allocation.
     - Unallocated client collateral sitting in member pool accounts (>24 hours).
   - Provide CSV export formatted for direct submission in daily SEBI collateral reporting files.

9. **Build `MemberRiskLimitTerminal` Component:**
   - Implement subordinate Trading Member (TM) risk governance panel.
   - Allow Clearing Members to set intraday limits per TM:
     - Max Gross Turnover Limit (INR).
     - Maximum Margin Utilization Threshold (default 85%).
     - Pro-Trading vs Client-Trading margin allocation partitions.
   - Add emergency control triggers:
     - "Suspend New Orders" (Square-off only mode).
     - "Force Cancel Open Orders".
     - "Trigger Member Square-Off".
   - Require dual-confirmation security modals before executing restrictive member risk commands.

10. **Implement `SettlementGuaranteeFundContributionViewer`:**
    - Render Core SGF participation breakdown:
      - Minimum Member SGF Quota (based on historical trade turnover and open interest).
      - Dynamic Cover-1 and Cover-2 stress-test scenario charges computed during daily simulation runs.
      - Total committed funds and current SGF health score.
    - Connect Viem RPC hook to read `SettlementGuaranteeFund.sol` on Hyperledger Besu:
      - Display confirmed on-chain block number of last deposit.
      - Display smart contract transaction hash, tokenized G-Sec escrow contract address, and on-chain verification badge.

11. **Develop `CollateralDepositWithdrawalModal` with Maker-Checker Controls:**
    - Create a multi-step modal for collateral actions:
      - **Deposit Flow:** Select asset type (Cash via RTGS, G-Sec pledge, e-BG, FDR), enter amounts, generate unique virtual account payment reference or depository pledge transaction ID.
      - **Withdrawal Flow:** Specify release amount, perform automated real-time pre-check ensuring post-withdrawal capital remains above BMC + initial margin + 50:50 ratio buffers.
    - Enforce Maker-Checker workflow: High-value collateral withdrawals or risk limit adjustments enter `PENDING_CHECKER_APPROVAL` state, requiring an authorized secondary admin signature before execution.

12. **Build Global Risk Alerting & Toast Notification System:**
    - Implement a sticky top-level Alert Banner that turns:
      - **Amber:** When aggregate member margin utilization exceeds 80% or any single TM exceeds 85%.
      - **Red (Flashing):** When any member breaches 90% utilization, Base Minimum Capital is breached, or an e-BG is within 7 days of expiry without renewal.
    - Integrate sound-enabled audio chime (user-toggleable) and browser desktop notifications for critical margin breaches.

13. **Implement Audit Trail & Regulatory Export Capabilities:**
    - Create `ClearingAuditTrailDrawer` displaying immutable timestamped logs of every limit change, collateral action, and emergency restriction.
    - Integrate client-side PDF generation (`@react-pdf/renderer` or server-side Puppeteer) and CSV export for SEBI Capital Adequacy Statements, Form A/B reports, and Daily Margin Segregation files.

14. **Testing, Simulation & Verification:**
    - Write unit tests using Vitest for 50:50 collateral ratio calculations, haircut deductions, and peak margin interpolation.
    - Implement Playwright end-to-end test suites simulating:
      - Live WebSocket margin ticks pushing utilization from 60% to 92% and validating alert banner activation.
      - NeSL e-BG registration and verification flow.
      - Depository pledge break identification and reconciliation filter.
      - Maker-checker collateral withdrawal approval workflow.

## Interfaces / Contracts

### Protobuf Definition: Capital Adequacy & Member Risk
```protobuf
syntax = "proto3";

package growww.clearing.v1;

import "google/protobuf/timestamp.proto";

enum CollateralType {
  COLLATERAL_TYPE_UNSPECIFIED = 0;
  COLLATERAL_TYPE_CASH = 1;
  COLLATERAL_TYPE_BANK_GUARANTEE = 2;
  COLLATERAL_TYPE_FIXED_DEPOSIT = 3;
  COLLATERAL_TYPE_GOVT_SECURITY = 4;
  COLLATERAL_TYPE_TREASURY_BILL = 5;
  COLLATERAL_TYPE_APPROVED_EQUITY = 6;
}

enum RiskHealthStatus {
  RISK_HEALTH_NORMAL = 0;      // Utilization < 70%
  RISK_HEALTH_ADVISORY = 1;    // Utilization 70% - 84.99%
  RISK_HEALTH_MARGIN_CALL = 2; // Utilization 85% - 89.99%
  RISK_HEALTH_CRITICAL = 3;    // Utilization >= 90%
}

message CapitalAdequacySnapshot {
  string clearing_member_id = 1;
  string clearing_member_code = 2;
  string exchange_segment = 3; // e.g. "NSE_FO", "BSE_EQUITY", "MCX_COMMODITY"
  
  double base_minimum_capital_inr = 4;
  double total_gross_collateral_inr = 5;
  double total_effective_collateral_inr = 6;
  
  double cash_and_equivalents_inr = 7;
  double non_cash_collateral_inr = 8;
  double cash_ratio_percentage = 9; // Must be >= 50.0
  bool is_cash_ratio_compliant = 10;
  
  double initial_margin_inr = 11; // SPAN / VaR
  double extreme_loss_margin_inr = 12;
  double peak_margin_obligation_inr = 13;
  double total_margin_obligation_inr = 14;
  
  double net_liquid_capital_inr = 15;
  double free_margin_inr = 16;
  double margin_utilization_percentage = 17;
  
  RiskHealthStatus health_status = 18;
  google.protobuf.Timestamp timestamp = 19;
}

message PeakMarginIntradayRecord {
  string clearing_member_id = 1;
  string trading_date = 2; // YYYY-MM-DD
  int32 snapshot_window = 3; // 1, 2, 3, or 4
  google.protobuf.Timestamp snapshot_time = 4;
  double margin_required_inr = 5;
  double effective_collateral_inr = 6;
  double utilization_percentage = 7;
  double shortfall_amount_inr = 8;
  bool is_breached = 9;
}

message BankGuaranteeItem {
  string bg_id = 1;
  string clearing_member_id = 2;
  string issuing_bank_name = 3;
  string bank_branch_ifsc = 4;
  string bg_reference_number = 5;
  string nesl_unique_identifier = 6;
  CollateralType collateral_type = 7;
  double face_value_inr = 8;
  double haircut_percentage = 9;
  double effective_value_inr = 10;
  google.protobuf.Timestamp issue_date = 11;
  google.protobuf.Timestamp claim_expiry_date = 12;
  int32 days_to_expiry = 13;
  string status = 14; // "ACTIVE", "EXPIRING_SOON", "EXPIRED", "INVOKED"
}
```

### TypeScript Interfaces & Schemas
```typescript
export type RiskLevel = 'SAFE' | 'ADVISORY' | 'MARGIN_CALL' | 'CRITICAL';

export type DepositoryType = 'CDSL' | 'NSDL';

export type PledgeStatus = 
  | 'PLEDGED_TO_TM'
  | 'RE_PLEDGED_TO_CM'
  | 'ALLOCATED_TO_CC'
  | 'UNPLEDGE_IN_PROGRESS'
  | 'RELEASED';

export interface MemberCapitalMetrics {
  clearingMemberId: string;
  clearingMemberCode: string;
  clearingMemberName: string;
  segment: 'CAPITAL_MARKET' | 'FUTURES_OPTIONS' | 'CURRENCY' | 'COMMODITIES';
  baseMinimumCapital: string; // Decimal string
  grossCollateral: string;
  effectiveCollateral: string;
  cashComponent: string;
  nonCashComponent: string;
  cashRatioPercentage: number;
  cashRatioCompliant: boolean;
  totalMarginRequirement: string;
  spanMargin: string;
  extremeLossMargin: string;
  peakMarginRequirement: string;
  netLiquidCapital: string;
  freeMargin: string;
  marginUtilizationPct: number;
  healthStatus: RiskLevel;
  lastUpdatedTimestamp: string;
}

export interface IntradayPeakSnapshot {
  windowId: 1 | 2 | 3 | 4;
  windowName: string;
  timeRange: string;
  snapshotTimestamp: string | null;
  marginRequired: string;
  collateralAvailable: string;
  utilizationPct: number;
  isPeakOfDay: boolean;
  shortfallAmount: string;
}

export interface DepositoryPledgeRecord {
  pledgeId: string;
  clientUcc: string;
  clientNameMasked: string;
  tradingMemberCode: string;
  depository: DepositoryType;
  isin: string;
  symbol: string;
  quantity: number;
  lastClosingPrice: string;
  grossValue: string;
  haircutPercentage: number;
  effectiveValue: string;
  pledgeStatus: PledgeStatus;
  depositoryPledgeRef: string;
  pledgeTimestamp: string;
  allocatedToClearingCorp: boolean;
  reconciliationMatch: boolean;
}

export interface ElectronicBankGuarantee {
  bgId: string;
  bankName: string;
  branchIfsc: string;
  bgNumber: string;
  neslRefNumber: string;
  faceValue: string;
  haircutPct: number;
  effectiveValue: string;
  issuanceDate: string;
  expiryDate: string;
  claimExpiryDate: string;
  daysRemaining: number;
  lienMarkedTo: string; // e.g. "NSE Clearing Limited" or "NBSE Clearing Corp"
  status: 'ACTIVE' | 'WARNING_EXPIRING' | 'EXPIRED' | 'INVOKED';
  documentDownloadUrl?: string;
}

export interface SGFOnChainContribution {
  clearingMemberAddress: `0x${string}`;
  tier1CashEquivalentCbdc: string;
  tier1TokenizedGSecValue: string;
  totalOnChainSgfValue: string;
  mandatoryQuotaRequirement: string;
  quotaCompliancePct: number;
  lastBesuBlockNumber: number;
  contractTxHash: `0x${string}`;
  isCryptographicallyVerified: boolean;
}

export interface CollateralActionRequest {
  requestId: string;
  clearingMemberId: string;
  actionType: 'DEPOSIT' | 'WITHDRAWAL';
  collateralType: 'CASH_RTGS' | 'BANK_GUARANTEE' | 'FDR' | 'GSEC_PLEDGE';
  requestedAmount: string;
  makerUserId: string;
  makerComment: string;
  checkerUserId?: string;
  status: 'PENDING_MAKER' | 'PENDING_CHECKER' | 'APPROVED' | 'REJECTED' | 'EXECUTED';
  postWithdrawalProjectedUtilizationPct: number;
  isEligibleForExecution: boolean;
  createdAt: string;
}
```

### REST & WebSocket API Contracts
- `GET /api/v1/clearing/members/{cmId}/capital-adequacy`:
  - Returns `MemberCapitalMetrics` including BMC, cash ratio, and utilization.
- `GET /api/v1/clearing/members/{cmId}/peak-margin/today`:
  - Returns array of `IntradayPeakSnapshot` for windows 1 through 4.
- `GET /api/v1/clearing/members/{cmId}/bank-guarantees`:
  - Returns list of `ElectronicBankGuarantee` items with NeSL verification statuses.
- `GET /api/v1/clearing/members/{cmId}/pledges?page=1&limit=50&depository=ALL`:
  - Returns paginated `DepositoryPledgeRecord` items and break summary counts.
- `POST /api/v1/clearing/collateral/request`:
  - Body: `CollateralActionRequest` (triggers Maker-Checker workflow).
- `POST /api/v1/clearing/members/{cmId}/trading-members/{tmId}/risk-limits`:
  - Modifies TM intraday exposure multipliers and order restriction flags.
- `WSS wss://api.growww.in/ws/v1/clearing/stream?token={jwt}`:
  - Subscribes to dynamic topics:
    - `SUB cm:metrics:{cmId}`
    - `SUB cm:alerts:{cmId}`
    - `SUB cm:pledge:{cmId}`

## Security & Compliance Notes
- **SEBI Member Capital Adequacy Norms:** The portal strictly enforces SEBI Master Circular guidelines. Base Minimum Capital (BMC) cannot be utilized towards risk margin requirements. If effective collateral drops below BMC, all trading terminal privileges are automatically restricted to square-off only.
- **50:50 Cash-to-Non-Cash Collateral Ratio:** The platform calculates the cash component dynamically. If non-cash collateral (equity shares, mutual funds) exceeds 50% of the total margin requirement, the excess non-cash collateral receives zero collateral credit (100% haircut) to prevent over-reliance on volatile collateral.
- **Client Fund Segregation & Non-Commingling:** In compliance with SEBI Upstreaming circulars, client funds collected by Trading Members must be upstreamed directly to the Clearing Corporation bank accounts. The portal visually audits and flags any client funds improperly retained in member proprietary bank accounts.
- **NeSL e-BG Verification & Expiry Guard:** Electronic Bank Guarantees must be verified via NeSL cryptographic digital signatures. The system automatically enforces an escalating collateral haircut schedule as e-BGs approach expiry (30 days: Notification; 15 days: 50% Haircut; 7 days: 100% Haircut / complete margin disqualification).
- **Maker-Checker Governance for Capital Actions:** Any manual collateral release, capital withdrawal, or risk ceiling expansion exceeding ₹10 Lakhs strictly mandates dual-authorization (Maker submits request, authorized Checker approves with hardware MFA or Web3 signature).
- **Immutable Audit Logging:** All risk threshold modifications, order cut-offs, and pledge adjustments generate tamper-evident audit records submitted to the internal Audit Log Service (Prompt 218).

## Acceptance Criteria
- [ ] `CapitalAdequacyOverview` displays real-time Net Liquid Capital, Base Minimum Capital (BMC), and color-coded Margin Utilization percentage.
- [ ] Visual gauge changes state dynamically across Safe (<70%), Advisory (70-85%), Margin Call (85-90%), and Critical (>90%).
- [ ] `PeakMarginUtilizationTracker` renders all four statutory SEBI intraday snapshot windows and accurately identifies peak margin liabilities.
- [ ] `CashCollateralSegregationMonitor` calculates the 50:50 cash-to-non-cash collateral ratio and automatically discounts excess non-cash collateral.
- [ ] Upstreaming of client funds view shows exact allocations deposited with the Clearing Corporation versus clearing member accounts.
- [ ] `BankGuaranteeManager` renders NeSL e-BGs with active days-to-expiry alerts and enforces mandatory haircut discounts within 15 days of expiry.
- [ ] `MarginPledgeReconciler` displays CDSL and NSDL client pledge records and flags depository ledger reconciliation breaks.
- [ ] `MemberRiskLimitTerminal` permits setting Trading Member (TM) intraday gross turnover limits and provides emergency order cut-off toggles.
- [ ] `SettlementGuaranteeFundContributionViewer` reads on-chain `SettlementGuaranteeFund.sol` state via Viem, displaying block confirmations and verified status.
- [ ] `CollateralDepositWithdrawalModal` enforces Maker-Checker dual authorization and blocks withdrawals that would violate BMC or 50:50 buffers.
- [ ] WebSocket streaming client handles high-frequency margin utilization ticks without frame drops or UI latency.
- [ ] Vitest unit tests pass for 50:50 ratio logic, haircut calculations, and peak margin shortfall rules.
- [ ] Playwright E2E tests verify the complete capital adequacy dashboard experience and collateral approval workflow.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 206: Pre-Trade Risk & Margin Checks Service
  - Prompt 229: Real-Time VaR Margin Engine
  - Prompt 230: Settlement Guarantee Fund & Default Waterfall Service
  - Prompt 260: Clearing Corporation Interoperability & SEBI Margin Pledge Gateway
  - Prompt 315: On-Chain Settlement Guarantee Fund Contract (`SettlementGuaranteeFund.sol`)
  - Prompt 601: Next.js Investor & Admin Web Scaffolding
- **Parallel Tasks:**
  - Prompt 605: Admin Risk Exception & Multi-Party Approval Portal
  - Prompt 606: Admin Proof of Reserve Reconciliation Workstation
  - Prompt 607: Admin Regulatory Reporting Dashboard
- **Downstream Blockers:**
  - Prompt 906: UAT Plan & Regulatory Sandbox Scenarios
  - Prompt 907: Regulatory Sandbox Pilot Launch Plan
