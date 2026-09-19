# 616 - Next.js 14 Pro-Trader Copy Trading & Strategy Management Portal

## Purpose
In modern algorithmic and institutional copy trading ecosystems, professional asset managers, SEBI-registered portfolio managers (PMS), Research Analysts (RA), Investment Advisers (IA), and approved Master Traders require a dedicated, low-latency workstation to oversee automated trade replication across hundreds or thousands of retail and institutional follower accounts. Operating under SEBI algorithmic trading guidelines, SEBI Investment Advisers regulations, and IFSCA market conductor rules, Master Traders must not treat copy trading as informal social betting; they require institutional risk controls, granular replication oversight, strict fee accounting, and real-time execution telemetry.

To prevent unmitigated systemic risks, excessive market impact, or gaming of performance fee structures, copy trading portals must enforce strict High-Water Mark (HWM) accounting, real-time slippage telemetry, child-order fan-out tracking, anti-drawdown throttles, and tamper-evident audit trails. Furthermore, fee distributions must be transparent, verifiable, and reconciled against immutable on-chain payout distribution receipts on Hyperledger Besu.

This prompt specifies the architecture, user experience, data schemas, and technical implementation of the **Next.js 14 Pro-Trader Copy Trading & Strategy Management Portal** (`apps/growww_web/master-trader`), adhering to platform Architecture Decision Record ADR-0042 ("Copy Trading Architecture and Pro-Rata Replication Engine") and Operations RUNBOOK-30 ("Master Trader Incident Square-Off, HWM Settlement, and Replication Fault Recovery"). The portal empowers fund managers and Master Traders to monitor aggregate Assets Under Management (AUM), track sub-100ms replication latency telemetry across follower cohorts, configure real-time risk sliders, audit High-Water Mark profit-share distributions, manage API keys, and export comprehensive regulatory compliance packages.

## What You Are Building
An enterprise-grade, high-availability Next.js 14 institutional management workstation featuring:
- `MasterTraderOverviewDashboard`: Executive mission-control dashboard rendering aggregate replicated AUM, active follower counts, 24-hour strategy P&L, cumulative profit-share accrued, rolling Sharpe and Sortino ratios, and maximum drawdown metrics.
- `FollowerManagementGrid`: High-throughput TanStack Table v8 virtualized grid rendering thousands of follower accounts, displaying follower equity, allocated copy ratio, replication latency percentiles (p50, p90, p99), unrealized/realized P&L, connection status (`ACTIVE`, `PAUSED`, `MARGIN_HALTED`, `DISCONNECTED`), and granular follower override actions.
- `RealTimeRiskSliderControls`: Dynamic parameter guardrail configurator with tactile sliders and numeric inputs for configuring Max Allowed Drawdown (Soft Warning and Hard Square-Off), Max Slippage per Order (bps), Max Position Concentration (% of follower equity), Maximum Open Leverage, and Volatility Dampening Throttles.
- `ReplicationLatencyTelemetryRibbon`: Real-time streaming telemetry ribbon tracking end-to-end signal propagation across the execution pipeline: Master Order Fill -> Copy Trading Service (Prompt 265) -> Pre-Trade Risk Engine (Prompt 206) -> Child Order Generation -> Matching Engine (Prompt 205) -> Ledger Settlement Receipt.
- `HighWaterMarkProfitShareAnalytics`: Comprehensive fee accounting and revenue waterfall terminal displaying historical and unbilled performance fees based on strict High-Water Mark (HWM) watermark ratchets, hurdle rates, crystallization periods (Monthly/Quarterly), and fee split tiers (e.g., 80% Master / 20% Platform).
- `ApiKeyAndWebhookManager`: Institutional developer workspace for generating sub-account API keys with granular scopes (`read`, `signal_broadcast`, `cancel_only`), IP whitelisting (CIDR notation), HMAC SHA-256 secret generation, and real-time webhook endpoint configuration for replication telemetry and margin alerts.
- `StrategyPerformanceAndDrawdownViewer`: Advanced Recharts and Tremor charting panel visualizing cumulative NAV performance curves benchmarked against market indices (Nifty 50, Bank Nifty), rolling drawdown water-level charts, daily return histograms, and win/loss ratios.
- `OnChainProfitReceiptAuditor`: Web3 transparency console querying Hyperledger Besu via Viem, verifying cryptographic payout receipts from `CopyTradingFeeVault.sol` and `ProfitDistributionLedger.sol`, presenting transaction hashes, block numbers, Merkle roots of distribution batches, and fee deductions.
- `SebiComplianceAuditExporter`: Institutional regulatory reporting tool generating audit-ready PDF and CSV logs conforming to SEBI algorithmic trading guidelines, non-discretionary copy limits, investor risk acknowledgment logs, and fee transparency mandates.
- `EmergencyKillswitchModal`: Fail-safe risk control modal allowing Master Traders to pause signal replication, cancel resting child orders, or trigger panic market square-offs in compliance with RUNBOOK-30.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router application structure located under `apps/growww_web/master-trader` (and route group `apps/growww_web/app/(master-trader)/`).
  - Master Trader multi-tenant authentication, session management, and role-based access control (Master Trader Admin, Portfolio Manager, Risk Checker, Auditor).
  - Virtualized follower management tables supporting real-time sorting, filtering, and emergency individual follower disconnects.
  - Real-time risk slider controls with client-side optimistic updates and server synchronization.
  - High-Water Mark (HWM) profit-share analytics, fee crystallization schedules, and payout request submission.
  - WebSocket streaming integration for replication latency metrics and child-order fan-out progress.
  - API key provisioning, scope configuration, IP CIDR whitelisting, and webhook event delivery testing.
  - On-chain fee distribution receipt verification on Hyperledger Besu using Viem.
  - SEBI compliance report generation and audit log export (CSV / PDF).
  - RUNBOOK-30 emergency circuit breaker and killswitch workflows.
- **Out of Scope / Handled Elsewhere:**
  - Core Copy Trading replication engine, order fan-out logic, and pro-rata lot calculations (handled by Copy Trading Service, Prompt 265).
  - Core fee engine billing, ledger journalization, and automated tax withholding (handled by Fee Engine Prompt 210 and Automated Tax Ledger Prompt 336).
  - Core order matching and memory-mapped write-ahead logging (handled by Matching Engine Prompt 205 and Prompt 246).
  - Pre-trade risk and margin evaluation for follower accounts (handled by Pre-Trade Risk Service Prompt 206).
  - Retail investor strategy discovery marketplace, investor onboarding, and copy subscription management (handled by Prompt 601, Prompt 603, and Flutter App Prompt 506).
  - Smart contract authoring and bytecode deployment of fee vaults (handled by Prompt 310 and Prompt 326).

## Technology to Use
- **Framework:** Next.js 14 App Router leveraging React Server Components (RSC) for initial page hydration, Streaming SSR with Suspense, and Server Actions for authenticated mutations.
- **Language:** TypeScript 5.x with strict type checking, strict null checks, and zero `any` declarations.
- **Styling & UI Kit:** Tailwind CSS 3.4+ combined with shadcn/ui components (Radix UI primitives), styled with an institutional dark palette (Obsidian `#0A0E17`, Slate `#1E293B`, Emerald `#10B981`, Crimson `#EF4444`, Amber `#F59E0B`).
- **Data Tables & Virtualization:** TanStack Table v8 (`@tanstack/react-table`) paired with TanStack Virtual (`@tanstack/react-virtual`) for rendering large datasets of follower accounts without browser thread stalls.
- **Data Visualization:** Recharts and Tremor for NAV performance curves, drawdown underwater areas, replication latency histograms, and AUM allocation donuts.
- **State Management & Server Cache:** Zustand with Immer for local UI and filter state; TanStack Query v5 for server data hydration, optimistic updates, and cache invalidation.
- **Real-Time Communication:** Native browser WebSocket API and Server-Sent Events (SSE) for streaming sub-second replication latency metrics and follower status transitions.
- **Blockchain Client:** Viem 2.x for connecting to Hyperledger Besu JSON-RPC endpoints, querying smart contract state, decoding event logs, and verifying Merkle proofs.
- **Precision Financial Math:** Decimal.js for High-Water Mark calculations, fee splits, hurdle rates, and P&L tracking without IEEE 754 floating-point inaccuracies.
- **Form Management & Validation:** React Hook Form coupled with Zod schemas for risk configuration sliders, API key generation, and webhook settings.

## Backend / Infra Touchpoints
- **Copy Trading Service (Prompt 265):** Connects via REST (`/api/v1/copytrading/master/{masterId}`) and duplex WebSocket (`wss://api.growww.in/ws/v1/copytrading/telemetry`) to ingest live follower states, replication latency telemetry, signal broadcast statuses, and fan-out progress.
- **Fee Engine (Prompt 210):** Connects via REST (`/api/v1/fees/copytrading/hwm`) to retrieve High-Water Mark accounting ledgers, crystallization batch records, unbilled fee projections, and submit payout withdrawal requests.
- **API Gateway & BFF (Prompt 219):** Terminates external TLS, performs JWT bearer authentication, enforces mTLS for institutional connectors, verifies API keys, and enforces rate-limiting.
- **Pre-Trade Risk & Margin Checks Service (Prompt 206):** Ingests strategy risk parameters (max drawdown, leverage limits, slippage caps) and emits risk breach callbacks to the portal.
- **Order Service & Matching Engine (Prompt 204, Prompt 205):** Correlates parent order execution reports with child follower orders for end-to-end replication latency calculation.
- **Audit Log Service (Prompt 218):** Ingests immutable, hash-chained audit events for every risk parameter adjustment, API key creation, emergency pause action, and payout request.
- **Notification Service (Prompt 211):** Emits automated event triggers (email, SMS, Webhooks) notifying Master Traders and risk officers of follower margin breaches or system anomalies.
- **Hyperledger Besu RPC Node (Prompt 302):** Directly queries `CopyTradingFeeVault.sol` and `ProfitDistributionLedger.sol` to audit on-chain fee distribution receipts.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu (Enterprise Permissioned Consortium Ledger, QBFT Consensus, 2-second block intervals, zero gas volatility).
- **Interfaced Smart Contracts:**
  - `CopyTradingFeeVault.sol`: On-chain escrow vault holding performance fee distributions. The portal queries:
    ```solidity
    function getDistributionReceipt(bytes32 batchId) external view returns (
        bytes32 merkleRoot,
        uint256 totalAmount,
        uint256 masterShare,
        uint256 platformShare,
        uint256 timestamp,
        bool isSettled
    );
    function getMasterTraderCumulativePayout(address masterAddress) external view returns (
        uint256 totalPaidOut,
        uint256 pendingSettlement
    );
    ```
  - `ProfitDistributionLedger.sol`: Immutable record of follower profit-share allocations. The portal verifies Merkle proofs:
    ```solidity
    function verifyFollowerProfitShare(
        bytes32 batchId,
        address followerAddress,
        uint256 profitAmount,
        uint256 feeDeducted,
        bytes32[] calldata merkleProof
    ) external view returns (bool);
    ```
- **Cryptographic Custody Linking & Zero PII:**
  - Payout batches are finalized on-chain by committing the Merkle root of all follower profit deductions.
  - Zero Personally Identifiable Information (PII) is published on-chain. Master Traders and followers are identified on-chain solely by pseudonymous Ethereum addresses and hashed batch identifiers (`bytes32 batchId = SHA256(masterId + crystallizationPeriod + timestamp)`).

## Step-by-Step Build Instructions

1. **Scaffold Next.js 14 Master Trader Route Architecture:**
   - Initialize the directory structure at `apps/growww_web/master-trader` and route group `apps/growww_web/app/(master-trader)/`.
   - Create root layouts with navigation sidebar, status ribbons, breadcrumbs, and context providers:
     - `app/(master-trader)/dashboard/page.tsx`: Executive dashboard displaying aggregate AUM, follower summary, 24h P&L, and quick metrics.
     - `app/(master-trader)/followers/page.tsx`: Virtualized follower management table with filtering and individual controls.
     - `app/(master-trader)/risk-controls/page.tsx`: Real-time risk sliders, drawdown thresholds, and leverage caps.
     - `app/(master-trader)/telemetry/page.tsx`: Real-time replication latency telemetry, pipeline diagnostics, and slip monitors.
     - `app/(master-trader)/profit-share/page.tsx`: High-Water Mark analytics, fee crystallization history, and payout ledger.
     - `app/(master-trader)/api-keys/page.tsx`: API key provisioning, CIDR IP whitelist configuration, and webhook manager.
     - `app/(master-trader)/performance/page.tsx`: Strategy NAV curves, benchmark comparisons, and trade win/loss distribution.
     - `app/(master-trader)/compliance/page.tsx`: SEBI audit log export, regulatory disclosures, and risk acknowledgments.

2. **Establish Data Models & Prisma Migrations:**
   - Define PostgreSQL tables in `packages/database/prisma/schema.prisma`:
     - `MasterTrader`: Stores master trader ID, user ID, SEBI registration number (RA/IA/PMS), accreditation tier, public profile, and status (`PENDING`, `APPROVED`, `SUSPENDED`).
     - `MasterStrategy`: Stores strategy title, description, asset classes, base currency, target risk profile, min follower equity, max AUM capacity, and active status.
     - `FollowerSubscription`: Records follower user ID, master strategy ID, allocated equity, copy multiplier, max slippage tolerance, subscription timestamp, and status (`ACTIVE`, `PAUSED`, `MARGIN_HALTED`, `DISCONNECTED`).
     - `ReplicationTelemetryLog`: High-frequency table tracking signal broadcast timestamp, fan-out completion timestamp, p50/p90/p99 latency, child order count, fill success rate, and average slippage.
     - `HwmAccountingRecord`: Tracks strategy crystallization cycles, starting NAV, ending NAV, high-water mark peak, gross profit, hurdle deduction, performance fee accrued, platform fee share, and settlement status (`PROJECTED`, `CRYSTALLIZED`, `PAID_OUT`).
     - `ApiKey`: Stores hashed key, prefix, label, encrypted secret, permitted IP CIDRs, granular scopes, and expiration.
     - `WebhookSubscription`: Stores target webhook URL, secret token, subscribed events, active state, and failure retry count.

3. **Implement Authentication, RBAC, and Institutional Context Provider:**
   - Configure session validation middleware verifying JWT tokens issued by API Gateway (Prompt 219).
   - Enforce Role-Based Access Control (RBAC) checking for roles: `MASTER_TRADER_ADMIN`, `PORTFOLIO_MANAGER`, `RISK_CHECKER`, and `COMPLIANCE_AUDITOR`.
   - Implement `MasterTraderContext` providing active strategy selection, global status, unread risk alerts, and real-time connection status.

4. **Build `MasterTraderOverviewDashboard` Component:**
   - Construct executive cockpit showing top-line KPI cards:
     - Replicated AUM (with day-over-day delta percentage).
     - Active Follower Count (with pending onboarding count).
     - 24-Hour Strategy Net P&L (currency and percentage).
     - Cumulative Accrued Profit-Share (current crystallization period).
     - Strategy Health Index (Sharpe Ratio, Sortino Ratio, Calmar Ratio, Current Drawdown).
   - Embed real-time notification toast for critical system events (e.g., follower margin shortfall, signal latency spike > 150ms).

5. **Build `FollowerManagementGrid` with TanStack Table and Virtualization:**
   - Implement virtualized follower data table using `@tanstack/react-table` and `@tanstack/react-virtual`:
     - Columns: Follower ID (anonymized UCC), Equity (INR), Copy Ratio, Allocation Mode (Pro-Rata / Fixed Lot), Current P&L, Replication Latency (p90), Connection Status, and Actions.
     - Filter controls: Filter by status (`ACTIVE`, `PAUSED`, `MARGIN_HALTED`), equity size range, and search by UCC.
     - Individual Follower Actions: "Pause Replication", "Resume Replication", "Force Square-Off & Sever Connection".
     - Action confirmations requiring explicit modal verification to prevent accidental severance.

6. **Implement `RealTimeRiskSliderControls` Component:**
   - Develop intuitive, responsive risk configuration interface powered by React Hook Form and Zod:
     - Soft Drawdown Warning Slider (5% to 25%): Triggers notification when portfolio drawdown crosses threshold.
     - Hard Drawdown Stop Slider (10% to 40%): Automatically pauses signal replication and rejects new master orders.
     - Maximum Open Leverage Slider (1x to 5x): Limits maximum aggregate margin multiplier for child order replication.
     - Maximum Slippage per Order (1 to 50 bps): Rejects child orders if market execution exceeds slippage boundary.
     - Max Position Concentration (% of follower equity, 5% to 50%): Restricts single instrument exposure.
   - Implement optimistic UI updates with debounced Server Action synchronization to Copy Trading Service (Prompt 265) and Risk Service (Prompt 206).
   - Include Maker-Checker dual confirmation requirement when modifying hard stop levels or leverage ceilings.

7. **Build `ReplicationLatencyTelemetryRibbon` Component:**
   - Construct real-time telemetry component connected to WebSocket stream `wss://api.growww.in/ws/v1/copytrading/telemetry`:
     - Visualizes multi-stage latency breakdown across the replication pipeline:
       1. Master Execution Ingestion ($t_1$).
       2. Pro-Rata Sizing & Risk Check ($t_2$).
       3. Child Order Fan-Out Broadcast ($t_3$).
       4. Exchange Matching Engine Ack ($t_4$).
       5. Overall Round-Trip Time ($t_{\text{total}} = t_4 - t_1$).
     - Renders rolling p50, p90, and p99 latency gauge with visual color thresholds:
       - Green (Optimal): < 50ms.
       - Amber (Elevated): 50ms - 150ms.
       - Red (Degraded): > 150ms.
     - Displays active child-order fill rate (% of followers successfully matched) and execution slippage distribution.

8. **Construct `HighWaterMarkProfitShareAnalytics` Component:**
   - Build fee accounting and profit-share distribution terminal:
     - Visual HWM Water-Level Chart: Plots cumulative strategy NAV against historical High-Water Mark line, clearly highlighting the "Profit-Share Generation Zone" where NAV exceeds HWM.
     - Crystallization Calendar: Displays next scheduled fee crystallization date (e.g., end of calendar month) and countdown ticker.
     - Fee Breakdown Calculator: Implements Decimal.js logic computing:
       $$\text{Eligible Profit} = \max(0, \text{NAV}_{\text{current}} - \text{HWM}_{\text{previous}} - \text{Hurdle})$$
       $$\text{Total Performance Fee} = \text{Eligible Profit} \times \text{Fee Rate}$$
       $$\text{Master Payout} = \text{Total Performance Fee} \times \text{Master Split \%}$$
       $$\text{Platform Share} = \text{Total Performance Fee} \times \text{Platform Split \%}$$
     - Historical Crystallization Table: Paginated list of past cycles showing gross profit, fee deducted, payout transaction hash, and payment status.
     - Payout Request Action: Button allowing Master Traders to initiate withdrawal of crystallized earnings to registered bank accounts.

9. **Build `StrategyPerformanceAndDrawdownViewer` Component:**
   - Implement advanced charting suite using Recharts:
     - Cumulative Performance Area Chart: Interactive time-series showing Strategy NAV vs Nifty 50 and Bank Nifty over 1D, 1W, 1M, 3M, 1Y, and ALL timeframes.
     - Underwater Drawdown Chart: Downward-facing area chart plotting daily peak-to-trough drawdown percentages.
     - Monthly Returns Heatmap: Grid visualizing monthly percentage returns across calendar years.
     - Trade Analytics Panel: Win/Loss Ratio, Profit Factor, Average Win vs Average Loss, and Average Holding Time.

10. **Implement `ApiKeyAndWebhookManager` Component:**
    - Develop institutional API management interface:
      - API Key Creation Modal: Generate cryptographically secure API keys (`gw_master_live_...`) with user-defined label, expiry, and granular permissions (`read:telemetry`, `write:signals`, `manage:risk`).
      - IP CIDR Restriction Input: Enforces strict IP address whitelisting for all API key actions.
      - Secret Display Card: Displays generated API secret exactly once with client-side copy button and security warning.
      - Active Keys Table: Shows key prefix, label, created date, last used date, IP restrictions, and instant Revoke action.
      - Webhook Configuration Terminal: Input form for webhook delivery URL and secret, event type checkboxes (`follower.disconnected`, `risk.breach`, `fee.crystallized`), and a "Send Test Ping" diagnostic tool.

11. **Implement `OnChainProfitReceiptAuditor` Component:**
    - Build Web3 audit console powered by Viem connecting to Hyperledger Besu:
      - Interrogates `CopyTradingFeeVault.sol` to fetch batch receipt metadata for any selected crystallization period.
      - Displays on-chain transaction hash, block height, confirmation timestamp, total distributed amount, and platform fee deduction.
      - Merkle Proof Verifier: Allows inputting a follower's pseudonymous address to cryptographically verify their exact profit-share deduction against the on-chain Merkle root.
      - Provides direct hyperlinks to the internal Besu Consortium Explorer (`https://explorer.growww.in/tx/0x...`).

12. **Build `SebiComplianceAuditExporter` Component:**
    - Develop regulatory export interface compliant with SEBI guidelines:
      - Exports standardized audit packages including:
        - Timestamped log of all risk limit adjustments with actor identity and dual-authorization signatures.
        - Strategy replication disclosure statement confirming non-discretionary execution within pre-set risk parameters.
        - Follower consent and risk profiling audit trail.
        - High-Water Mark calculation audit trail proving no fees were billed below peak NAV.
      - Supports single-click generation of encrypted PDF and CSV reports with SHA-256 integrity hash.

13. **Implement `EmergencyKillswitchModal` (RUNBOOK-30 Compliance):**
    - Construct high-visibility emergency intervention modal:
      - Action 1: "Halt Signal Broadcast" (prevents new parent orders from replicating without affecting open positions).
      - Action 2: "Cancel All Pending Child Orders" (pulls all unexecuted follower limit orders across the book).
      - Action 3: "Emergency Market Square-Off" (liquidates all open follower positions at market price).
    - Requires two-stroke confirmation (typing the strategy name and entering a 6-digit TOTP / SMS MFA code) to prevent accidental actuation.
    - Emits high-priority audit log and triggers immediate WebSocket notification to exchange risk desks.

14. **Automated Unit, Integration, and E2E Testing:**
    - Unit tests using Vitest covering Decimal.js High-Water Mark calculations, fee split logic, Zod schema validations, and latency histogram bucketing.
    - Integration tests validating TanStack Table virtual scrolling performance with 5,000 mock follower rows.
    - Playwright E2E test suites validating:
      - Executive dashboard KPI loading and navigation.
      - Follower filtering, sorting, and pause action execution.
      - Risk slider adjustment and Server Action optimistic update round-trip.
      - High-Water Mark payout request submission.
      - API key generation with IP CIDR validation.
      - Emergency killswitch two-stroke confirmation modal.

## Interfaces / Contracts

### Protobuf Definition: Copy Trading Master Telemetry & HWM
```protobuf
syntax = "proto3";

package growww.copytrading.v1;

import "google/protobuf/timestamp.proto";

enum FollowerStatus {
  FOLLOWER_STATUS_UNSPECIFIED = 0;
  FOLLOWER_STATUS_ACTIVE = 1;
  FOLLOWER_STATUS_PAUSED = 2;
  FOLLOWER_STATUS_MARGIN_HALTED = 3;
  FOLLOWER_STATUS_DISCONNECTED = 4;
}

enum CrystallizationStatus {
  CRYSTALLIZATION_STATUS_UNSPECIFIED = 0;
  CRYSTALLIZATION_STATUS_PROJECTED = 1;
  CRYSTALLIZATION_STATUS_CRYSTALLIZED = 2;
  CRYSTALLIZATION_STATUS_SETTLING = 3;
  CRYSTALLIZATION_STATUS_PAID_OUT = 4;
}

message StrategyRiskProfile {
  string strategy_id = 1;
  string master_id = 2;
  double soft_drawdown_warning_pct = 3;
  double hard_drawdown_stop_pct = 4;
  double max_leverage_multiplier = 5;
  uint32 max_slippage_bps = 6;
  double max_concentration_pct = 7;
  bool is_replication_active = 8;
  google.protobuf.Timestamp updated_at = 9;
}

message FollowerReplicationStatus {
  string follower_id = 1;
  string strategy_id = 2;
  string follower_ucc = 3;
  double allocated_equity_inr = 4;
  double copy_multiplier = 5;
  double current_pnl_inr = 6;
  uint32 replication_latency_p90_ms = 7;
  FollowerStatus status = 8;
  google.protobuf.Timestamp subscribed_at = 9;
}

message ReplicationTelemetryEvent {
  string strategy_id = 1;
  string signal_id = 2;
  google.protobuf.Timestamp master_fill_time = 3;
  uint32 latency_ingest_to_risk_ms = 4;
  uint32 latency_risk_to_fanout_ms = 5;
  uint32 latency_fanout_to_match_p50_ms = 6;
  uint32 latency_fanout_to_match_p99_ms = 7;
  uint32 total_child_orders = 8;
  uint32 filled_child_orders = 9;
  uint32 failed_child_orders = 10;
  double average_slippage_bps = 11;
}

message HwmDistributionBatch {
  string batch_id = 1;
  string strategy_id = 2;
  string master_id = 3;
  google.protobuf.Timestamp period_start = 4;
  google.protobuf.Timestamp period_end = 5;
  double starting_nav = 6;
  double ending_nav = 7;
  double previous_hwm = 8;
  double new_hwm = 9;
  double total_eligible_profit_inr = 10;
  double master_payout_inr = 11;
  double platform_share_inr = 12;
  string on_chain_tx_hash = 13;
  bytes merkle_root = 14;
  CrystallizationStatus status = 15;
}

message DistributionReceipt {
  string batch_id = 1;
  string tx_hash = 2;
  uint64 block_number = 3;
  string contract_address = 4;
  string merkle_root_hex = 5;
  double total_payout_amount_inr = 6;
  google.protobuf.Timestamp confirmed_at = 7;
  bool is_verified = 8;
}
```

### TypeScript Domain Interfaces
```typescript
export type FollowerStatus = 
  | 'ACTIVE' 
  | 'PAUSED' 
  | 'MARGIN_HALTED' 
  | 'DISCONNECTED';

export type CrystallizationStatus = 
  | 'PROJECTED' 
  | 'CRYSTALLIZED' 
  | 'SETTLING' 
  | 'PAID_OUT';

export interface MasterStrategySummary {
  strategyId: string;
  masterId: string;
  title: string;
  assetClass: 'EQUITY_CASH' | 'F_AND_O' | 'COMMODITY' | 'MULTI_ASSET';
  totalAumInr: string;
  activeFollowersCount: number;
  netPnl24hInr: string;
  netPnl24hPct: number;
  accruedProfitShareInr: string;
  currentNav: string;
  highWaterMark: string;
  sharpeRatio: number;
  sortinoRatio: number;
  maxDrawdownPct: number;
  isReplicationActive: boolean;
}

export interface FollowerItem {
  followerId: string;
  followerUcc: string;
  allocatedEquityInr: string;
  copyMultiplier: number;
  unrealizedPnlInr: string;
  realizedPnlInr: string;
  latencyP90Ms: number;
  status: FollowerStatus;
  subscribedAt: string;
  lastSignalReplicatedAt?: string;
}

export interface StrategyRiskConfig {
  strategyId: string;
  softDrawdownWarningPct: number; // e.g. 10.0 for 10%
  hardDrawdownStopPct: number;    // e.g. 20.0 for 20%
  maxLeverageMultiplier: number;  // e.g. 2.5 for 2.5x
  maxSlippageBps: number;         // e.g. 15 for 15 bps
  maxConcentrationPct: number;    // e.g. 25.0 for 25%
  isReplicationActive: boolean;
  updatedAt: string;
}

export interface ReplicationTelemetrySnapshot {
  signalId: string;
  timestamp: string;
  totalFollowersTargeted: number;
  successfulFills: number;
  failedFills: number;
  latencyP50Ms: number;
  latencyP90Ms: number;
  latencyP99Ms: number;
  averageSlippageBps: number;
  pipelineStages: {
    ingestToRiskMs: number;
    riskToFanoutMs: number;
    fanoutToMatchMs: number;
  };
}

export interface HwmAccountingItem {
  batchId: string;
  periodLabel: string; // e.g. "Q2 2026" or "Sep 2026"
  periodStart: string;
  periodEnd: string;
  startingNav: string;
  endingNav: string;
  previousHwm: string;
  newHwm: string;
  eligibleProfitInr: string;
  performanceFeeRatePct: number; // e.g. 20.0 for 20%
  masterPayoutInr: string;
  platformShareInr: string;
  status: CrystallizationStatus;
  onChainTxHash?: `0x${string}`;
  merkleRoot?: `0x${string}`;
}

export interface ApiKeyItem {
  keyId: string;
  prefix: string;
  label: string;
  scopes: ('read:telemetry' | 'write:signals' | 'manage:risk')[];
  ipWhitelistCidrs: string[];
  createdAt: string;
  lastUsedAt?: string;
  expiresAt?: string;
}

export interface WebhookConfig {
  webhookId: string;
  targetUrl: string;
  subscribedEvents: string[];
  isActive: boolean;
  createdAt: string;
}
```

### REST & WebSocket API Contracts
- `GET /api/v1/copytrading/master/strategies`:
  - Retrieves all active and archived strategies managed by the authenticated Master Trader.
- `GET /api/v1/copytrading/master/strategies/{strategyId}/summary`:
  - Retrieves real-time summary statistics, AUM, P&L, and HWM watermarks.
- `GET /api/v1/copytrading/master/strategies/{strategyId}/followers`:
  - Query Parameters: `page`, `limit`, `status`, `searchQuery`, `sortBy`, `sortOrder`.
  - Returns paginated follower list with equity, latency, and replication status.
- `POST /api/v1/copytrading/master/strategies/{strategyId}/followers/{followerId}/action`:
  - Request: `{ action: 'PAUSE' | 'RESUME' | 'FORCE_SQUARE_OFF' }`
  - Performs administrative override on an individual follower replication state.
- `GET /api/v1/copytrading/master/strategies/{strategyId}/risk`:
  - Retrieves current risk sliders and guardrail limits.
- `PUT /api/v1/copytrading/master/strategies/{strategyId}/risk`:
  - Request: `StrategyRiskConfig` payload.
  - Updates risk thresholds; triggers validation in Pre-Trade Risk Service (Prompt 206).
- `GET /api/v1/copytrading/master/strategies/{strategyId}/hwm`:
  - Retrieves historical HWM accounting batches and unbilled profit-share accruals.
- `POST /api/v1/copytrading/master/strategies/{strategyId}/payout/request`:
  - Request: `{ batchId: string, destinationBankAccountId: string }`
  - Submits payout withdrawal request for crystallized performance fees.
- `POST /api/v1/copytrading/master/strategies/{strategyId}/emergency-halt`:
  - Request: `{ confirmationCode: string, actionType: 'HALT_BROADCAST' | 'CANCEL_ORDERS' | 'SQUARE_OFF_ALL' }`
  - Executes RUNBOOK-30 emergency circuit breaker.
- `GET /api/v1/copytrading/master/api-keys`:
  - Lists provisioned API keys with metadata and IP whitelist restrictions.
- `POST /api/v1/copytrading/master/api-keys`:
  - Request: `{ label: string, scopes: string[], ipWhitelistCidrs: string[] }`
  - Returns generated API key and secret (displayed once).
- `DELETE /api/v1/copytrading/master/api-keys/{keyId}`:
  - Revokes an active API key immediately.
- `WSS wss://api.growww.in/ws/v1/copytrading/telemetry`:
  - WebSocket subscription streaming:
    - `SUB strategy:telemetry:{strategyId}` (sub-second latency and fan-out updates).
    - `SUB strategy:follower-status:{strategyId}` (follower state transitions).
    - `SUB strategy:risk-alert:{strategyId}` (drawdown warning notifications).

## Security & Compliance Notes
- **Anti-Gaming and Drawdown Controls:**
  - *High-Water Mark Ratchet Enforcement:* Performance fees are strictly calculated only on new net profits above the historical maximum peak NAV (HWM). If the strategy suffers a drawdown, zero performance fees accrue until the previous watermark is fully recovered.
  - *Capital Inflow/Outflow Adjustments:* Formula adjusts HWM dynamically for follower deposits and withdrawals using modified Dietz or time-weighted return methods to prevent artificial fee inflation.
  - *Anti-Churn Collar:* Prohibits rapid intraday round-trip trades intended solely to generate turnover or artificial volume. Strategies exceeding excessive order-to-trade ratios (OTR) face automated throttling.
- **SEBI Regulatory Compliance:**
  - *Adherence to SEBI RA and IA Regulations:* Master Traders must hold valid SEBI Research Analyst (RA), Investment Adviser (IA), or Portfolio Manager (PMS) credentials. The portal displays SEBI registration badges and disallows unaccredited marketing.
  - *Prohibition of Guaranteed Return Claims:* The portal interface strictly excludes terms promising fixed or risk-free returns. All marketing and strategy summaries mandate prominent risk disclosure banners.
  - *Investor Risk Categorization:* Follower replication is restricted if the strategy's risk classification exceeds the follower's CKYC risk tolerance profile.
  - *Non-Discretionary Parameter Lock:* While signals replicate automatically, followers retain full authority to configure independent stop-loss levels or sever copy replication at any time.
- **API Key Security & Zero Trust Architecture:**
  - API keys use HMAC SHA-256 signatures with millisecond timestamp verification to prevent replay attacks.
  - Strict IP CIDR whitelisting is enforced at the API Gateway level; requests from unlisted IPs are dropped immediately.
  - Raw API secrets are hashed with bcrypt/Argon2 before storage; secrets are shown only once upon generation.
- **Zero PII on Distributed Ledger:**
  - In compliance with the Digital Personal Data Protection (DPDP) Act 2023, investor names, PAN numbers, and bank account numbers are never committed to Hyperledger Besu.
  - On-chain receipts in `CopyTradingFeeVault.sol` record only cryptographic batch Merkle roots and anonymized addresses.
- **Maker-Checker Governance:**
  - Adjusting hard drawdown stop-loss levels, altering hurdle rates, or requesting fee payout withdrawals mandates dual confirmation from a designated Risk Checker or Auditor within the Master Trader organization.

## Acceptance Criteria
- [ ] Next.js 14 application compiles without errors under `apps/growww_web/master-trader` with zero TypeScript lint or build failures.
- [ ] `MasterTraderOverviewDashboard` accurately renders aggregate AUM, follower count, 24h P&L, Sharpe/Sortino ratios, and HWM watermarks.
- [ ] `FollowerManagementGrid` renders 5,000+ follower accounts smoothly using `@tanstack/react-virtual` with zero frame drops during scrolling.
- [ ] Follower filtering by status (`ACTIVE`, `PAUSED`, `MARGIN_HALTED`) and search by UCC return correct filtered results in sub-100ms.
- [ ] Individual follower actions ("Pause", "Resume", "Force Square-Off") successfully dispatch backend commands and update table state optimistically.
- [ ] `RealTimeRiskSliderControls` dynamic inputs (Soft Drawdown, Hard Drawdown, Leverage, Slippage) persist changes to Copy Trading Service (Prompt 265) and Risk Service (Prompt 206).
- [ ] `ReplicationLatencyTelemetryRibbon` connects to WebSocket feed and displays live p50, p90, and p99 latency metrics with color-coded warning thresholds.
- [ ] High-Water Mark analytics correctly calculates performance fees using Decimal.js, ensuring zero fees accrue below previous peak NAV.
- [ ] `OnChainProfitReceiptAuditor` successfully queries `CopyTradingFeeVault.sol` on Hyperledger Besu via Viem, verifying distribution receipts and Merkle roots.
- [ ] API Key management interface provisions new keys with IP CIDR validation, displays the secret exactly once, and supports instant revocation.
- [ ] Webhook management interface allows configuring target URLs, selecting events, and executing successful test ping dispatches.
- [ ] `EmergencyKillswitchModal` enforces two-stroke confirmation (strategy name confirmation + MFA verification) and successfully broadcasts halt signals adhering to RUNBOOK-30.
- [ ] SEBI compliance report generator exports timestamped PDF and CSV audit logs with SHA-256 integrity checksums.
- [ ] Vitest test suite covers financial math calculations (HWM, hurdle rate, fee waterfall) and Zod schema validations with 100% pass rate.
- [ ] Playwright E2E tests validate complete workflow: dashboard hydration, follower management, risk slider updates, HWM payout request, and emergency killswitch actuation.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 205: Matching Engine & In-Memory Order Book (for execution report ingestion).
  - Prompt 206: Pre-Trade Risk & Margin Checks Microservice (for risk slider integration and margin breach alerts).
  - Prompt 210: Dynamic Fee Calculation & Billing Engine (for High-Water Mark calculation rules and fee splits).
  - Prompt 219: API Gateway & BFF Architecture (for JWT authentication, rate limiting, and API key routing).
  - Prompt 265: High-Performance Copy Trading & Strategy Replication Service (core backend for signal fan-out and replication telemetry).
  - Prompt 302: Hyperledger Besu Private Consortium Network Deployment (consortium blockchain infrastructure).
  - Prompt 310: Automated Copy Trading Fee Vault & Profit Distribution Ledger Smart Contract (for on-chain receipt verification).
  - Prompt 601: Next.js Investor & Admin Web Scaffolding (monorepo layout, shared UI primitives, and design system).
- **Parallel Tasks:**
  - Prompt 603: Web Trading Dashboard & Advanced Charting Terminal.
  - Prompt 604: Admin User KYC & KYB Review Dashboard.
  - Prompt 605: Admin Risk Exception & Multi-Party Approval Workstation.
- **Downstream Blockers:**
  - Prompt 506: Flutter Mobile Home & Copy Trading Follower Subscription Screen.
  - Prompt 906: Comprehensive UAT Plan & Regulatory Sandbox Scenarios.
  - Prompt 907: Regulatory Sandbox Pilot Launch Plan.
  - RUNBOOK-30: Operational Disaster Recovery Drills for Replication Failures.
