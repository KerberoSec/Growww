# 506 - Flutter Home / Dashboard Screen

## Purpose
Serves as the central command center for investors upon launching the Growww application. The dashboard provides an immediate, high-clarity snapshot of the user's total investment valuation, daily unrealized P&L, real-time market indices (NIFTY 50, SENSEX, GIFT NIFTY), top market gainers/losers, quick financial actions (Deposit Funds, Buy Fractional Shares, Verify Reserves), and a persistent Proof-of-Reserve transparency card verifying 100% custodian backing of all digital assets.

## What You Are Building
A highly responsive, visually rich dashboard screen in `apps/growww_flutter/lib/features/dashboard/` featuring:
- **Responsive Layout Shell:** Mobile layout with bottom navigation vs desktop layout with an expandable sidebar and multi-column grid.
- **Consolidated Portfolio Hero Card:** Total current balance, invested capital, today's P&L (in INR and percentage), and overall returns.
- **Proof-of-Reserve Live Status Banner:** Real-time badge indicating that 100% of underlying securities are locked in NSDL/CDSL custody, with block timestamp and tap-to-verify modal.
- **Market Indices Carousel:** Live streaming tickers for NIFTY 50, SENSEX, BANK NIFTY, and GIFT NIFTY with mini sparkline charts.
- **Top Movers & Market Breadth:** Tabbed list showing Top Gainers, Top Losers, and Most Active fractional securities.
- **Quick Action Bar:** One-tap shortcuts for Deposit Funds, Withdraw, KYC Verification, and Explore Securities.

## Scope Boundaries
- **In Scope:**
 - Dashboard UI layout, hero card data binding, indices carousel, sparkline rendering, quick-action navigation handlers, and pull-to-refresh.
- **Out of Scope / Handled Elsewhere:**
 - Full portfolio breakdown details (handled in Prompt 510).
 - Interactive technical candlestick charts (handled in Prompt 508).
 - Watchlist curation management (handled in Prompt 507).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Riverpod state management.
 - *Justification:* Riverpod `StreamProvider` enables real-time price streaming to individual index ticker widgets without triggering whole-screen rebuilds, maintaining 60/120 FPS performance.
- **Charts & Sparklines:** `fl_chart` or custom `CustomPainter` for lightweight, hardware-accelerated 24-hour index sparklines.
- **Skeleton Shimmers:** `skeletonizer` for polished placeholder animations while fetching initial API data.

## Backend / Infra Touchpoints
- **Portfolio Microservice:** Fetches user's aggregated portfolio balance (`/api/v1/portfolio/summary`) from Prompt 209.
- **Market Data Service:** WebSocket and REST feeds (`/api/v1/market/indices`, `/api/v1/market/movers`) from Prompt 207.
- **Proof-of-Reserve Registry:** Fetches current custodial reserve snapshot from Prompt 308.

## Blockchain Interaction
- **Live Proof-of-Reserve Attestation Card:** Displays the latest Merkle root commitment published to `ProofOfReserveRegistry.sol`. Shows users the physical share backing ratio (e.g., 100.0% backed) and the exact block number on the Hyperledger Besu consortium ledger, reassuring investors of complete custodial segregation.

## Step-by-Step Build Instructions
1. Create dashboard feature folder structure: `lib/features/dashboard/presentation/screens/`, `presentation/widgets/`, `application/`, `domain/`, `data/`.
2. Define domain models: `DashboardSummary`, `MarketIndexTicker`, `TopMoverItem`, `ProofOfReserveOverview`.
3. Create `dashboard_controller.dart` managing asynchronous fetching and auto-refresh of dashboard data.
4. Implement `PortfolioHeroCard` widget displaying total balance, today's profit/loss with green/red color coding, and hidden balance toggle (`Eye` icon for privacy).
5. Build `ProofOfReserveBanner` featuring an animated lock shield icon, 100% backing badge, and direct tap handler opening the blockchain audit sheet.
6. Implement `IndicesCarousel` with horizontal scrolling on mobile and compact row layout on desktop, binding to live index price streams.
7. Build `SparklinePainter` for rendering 24-hour price curves without performance overhead.
8. Build `QuickActionsBar` with responsive iconography and routing triggers (Deposit, Buy, Wallet, History).
9. Implement `TopMoversSection` featuring tabbed switching between Gainers, Losers, and Volume Leaders with real-time price changes.
10. Add `CustomScrollView` with `SliverAppBar` and `CupertinoSliverRefreshControl` for pull-to-refresh functionality.
11. Implement responsive breakpoint switching: adapt from a single vertical column on mobile to a 3-column dashboard grid on desktop/tablets.
12. Write unit tests for data parsing and golden tests for both light and dark dashboard themes.

## Interfaces / Contracts
```dart
// lib/features/dashboard/domain/models/dashboard_summary.dart
class DashboardSummary {
  final double totalPortfolioValueInr;
  final double totalInvestedInr;
  final double todayPnlInr;
  final double todayPnlPercentage;
  final double totalRealizedPnlInr;
  final double totalPlatformFeeDeductedInr; // Fixed 0.00% transaction fee (No fee at all)
  final ProofOfReserveOverview porOverview;
  final List<MarketIndexTicker> indices;
  final List<TopMoverItem> topGainers;
  final List<TopMoverItem> topLosers;

  const DashboardSummary({
    required this.totalPortfolioValueInr,
    required this.totalInvestedInr,
    required this.todayPnlInr,
    required this.todayPnlPercentage,
    required this.totalRealizedPnlInr,
    required this.totalPlatformFeeDeductedInr,
    required this.porOverview,
    required this.indices,
    required this.topGainers,
    required this.topLosers,
  });
}
```

## Security & Compliance Notes
- **Privacy Mode (Eye Toggle):** Users can tap an eye icon to mask their financial balances (`₹ ••••••`) when viewing in public spaces.
- **SEBI Market Disclosures:** Includes mandatory statutory disclaimer at the bottom of the dashboard: *"Securities investments are subject to market risks. Real assets held in custody with SEBI-registered depositories (NSDL/CDSL)."*
- **Fee Transparency:** Clearly surfaces total platform transaction fees deducted (fixed 0.00% (Zero Fee) / 0 bps (0.00% fee at launch) on turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol)), reaffirming the zero-holding-fee and zero-AUM-fee policy, while showing FIFO capital gains computed strictly for user tax compliance (Section 111A/112A).

## Acceptance Criteria
- [ ] Dashboard loads and renders initial data with smooth skeleton shimmers under 300ms.
- [ ] Market indices update in real time via WebSocket ticks with color flash feedback.
- [ ] Proof-of-Reserve banner displays verified backing ratio and opens the verification modal.
- [ ] Responsive layout adapts flawlessly between mobile (390px) and wide desktop (1920px).
- [ ] Pull-to-refresh triggers fresh state fetch from portfolio and market endpoints.
- [ ] Balance masking toggle instantly conceals/reveals financial amounts.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System), Prompt 505 (Authentication).
- **Backend Dependency:** Prompt 207 (Market Data), Prompt 209 (Portfolio Service), Prompt 308 (Proof-of-Reserve).
- **Enables:** Prompts 507 (Market Watchlist), 509 (Order Placement), 510 (Portfolio Holdings).
