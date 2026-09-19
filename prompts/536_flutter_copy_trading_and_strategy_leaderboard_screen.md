# 536 - Flutter Copy Trading & Strategy Leaderboard Screen

## Purpose
Provides retail and institutional investors with an intuitive, transparent, and high-performance social marketplace interface for copy trading and strategy discovery. In volatile equity, derivative, and tokenized real-world asset (RWA) markets, novice investors often lack the algorithmic execution infrastructure, risk discipline, and time required to capture alpha consistently. Conversely, seasoned master traders and algorithmic strategists seek a compliant, automated platform to monetize their trading expertise through regulated profit-sharing.

The **Flutter Copy Trading & Strategy Leaderboard Screen** (`apps/growww_flutter/lib/features/copy_trading/` and `lib/screens/copy_trading/`) bridges this gap in accordance with ADR-0042. It delivers a real-time marketplace where investors can:
1. Discover top-performing, exchange-verified master traders filtered by multi-dimensional risk-adjusted metrics (CAGR, Sharpe ratio, Sortino ratio, maximum drawdown, win rate, profit factor, and Assets Under Copy).
2. Inspect interactive equity curves, historical trade logs, asset allocation breakdowns, and monthly return matrices.
3. Configure granular capital allocations with interactive sliders, custom slippage boundaries (+/- 0.5% clamping), and automated emergency stop-loss detachment circuit breakers.
4. Monitor real-time mirrored positions, live execution latency, and slippage basis points via low-latency WebSocket streams.
5. Execute instantaneous one-tap emergency detachment to decouple capital and flatten or preserve positions.
6. Verify weekly High-Water Mark (HWM) profit-share deductions and cryptographic settlement receipts anchored to the permissioned Hyperledger Besu blockchain.

## What You Are Building
A modular, cross-platform Flutter copy trading subsystem supporting mobile (Android, iOS) and desktop (macOS, Windows, Linux) environments:
- `CopyTradingDashboardScreen`: The root container screen featuring tabbed navigation across "Strategy Marketplace / Leaderboard", "My Active Copied Strategies", "Live Mirrored Positions", and "Settlement History".
- `StrategyLeaderboardWidget`: A virtualized, high-frame-rate list and grid view displaying master trader cards with live rank badges, SEBI registration status, Assets Under Copy (AUC), 30D/90D/1Y return on investment (ROI), Maximum Drawdown (MDD), and verified trader shield badges.
- `LeaderboardFilterBar`: Multi-parameter filtering and sorting toolbar allowing users to filter by asset class (Equities, F&O, Commodities, RWAs), trading horizon (Intraday, Swing, Positional), risk tier (Conservative, Moderate, Aggressive), and sort by Sharpe Ratio, ROI, or Lowest Drawdown.
- `MasterTraderProfileScreen`: Deep-dive strategy inspection view showcasing the trader biography, verified credentials (SEBI RA/IA registration), historical trade log, asset class weighting doughnut chart, monthly return heatmap matrix, and open positions.
- `InteractiveEquityCurveWidget`: Hardware-accelerated `fl_chart` LineChart visualization rendering historical Net Asset Value (NAV) equity curves against benchmark indices (Nifty 50, Nifty Bank), featuring dynamic time horizons (1M, 3M, 6M, 1Y, ALL) and high-water mark overlay points.
- `CapitalAllocationBottomSheet`: Interactive investment configuration modal with dynamic allocation sliders, available unencumbered cash validation, replication mode selection (Equity-Proportional, Fixed-Ratio, Cash-Proportional), maximum slippage tolerance guard (+/- 0.5% default), and emergency drawdown circuit breaker threshold slider (e.g., -5%, -10%, -15%).
- `ActiveCopiedPositionsCard`: Live position tracking card displaying master entry price, follower fill price, execution slippage in basis points (bps), live mark-to-market unrealized PnL in INR and percentage, and stop-loss/take-profit boundaries.
- `EmergencyDetachmentButton`: High-visibility, guarded one-tap safety detachment action triggering instantaneous decoupling from the master trader, prompting the follower to either market-liquidate all replicated positions immediately or retain them for manual exit.
- `OnChainProfitShareReceiptViewer`: Slide-up audit sheet presenting the cryptographic Hyperledger Besu transaction hash, weekly High-Water Mark profit-share breakdown (10% to 15% performance fee cut), Merkle proof verification status, and block confirmation height.
- `MandatorySebiRiskBanner`: Persistent, non-dismissible statutory risk disclosure and warning banner complying with SEBI directives on algorithmic copy trading and investment advisory disclosures.

## Scope Boundaries
- **In Scope:**
  - Complete multi-platform Flutter UI/UX for strategy discovery, leaderboard sorting, master trader profiles, and portfolio allocation.
  - Interactive equity curve charting and benchmark comparison using `fl_chart`.
  - Dynamic capital allocation workflow with interactive sliders, balance guards, and risk parameter validation.
  - Low-latency WebSocket integration for live master fill notifications, child order execution alerts, and real-time position PnL streaming.
  - One-tap emergency detachment workflow with atomic local state update and confirmation modal.
  - High-Water Mark profit-share settlement history displaying on-chain Besu verification receipts.
  - Biometric step-up authentication (`local_auth`) for high-value capital allocations and emergency detachment execution.
  - Mandatory SEBI statutory risk warnings, disclaimers, and accessible UI color modes adhering to WCAG 2.1 AA standards.
- **Out of Scope / Handled Elsewhere:**
  - Backend proportional sizing mathematics, child order fanout, and sub-15ms Kafka message dispatching (handled in Copy Trading Service Prompt 265).
  - Central order matching engine and order book clearing (handled in Order Matching Engine Prompt 205).
  - Exchange margin calculations, initial margin haircuts, and core risk checks (handled in Risk Service Prompt 206).
  - Primary fiat deposit/withdrawal rails and custodial ledger accounts (handled in Wallet Service Prompt 203 and Payment Gateway Prompt 212).
  - Backend SEBI Research Analyst (RA) / Investment Adviser (IA) license verification workflow (handled in User Service Prompt 201).
  - Hyperledger Besu QBFT consensus validator operations and EVM smart contract compilation (handled in Prompt 309 and Prompt 329).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ utilizing `flutter_riverpod: ^2.5.1` with code generation (`@riverpod`, `AsyncNotifier`, `StreamNotifier`) for predictable, reactive state management.
- **Data Visualization & Charting:** `fl_chart: ^0.68.0` for high-performance, smooth vector rendering of historical equity curves, benchmark overlays, and high-water mark milestone markers.
- **Fixed-Point Financial Mathematics:** `decimal: ^2.3.3` for sub-paise INR precision and exact fractional quantity calculations, preventing floating-point rounding errors during allocation configuration.
- **Real-Time Data Streaming:** `web_socket_channel: ^3.0.0` with typed JSON / binary deserialization for sub-100ms mark-to-market position updates and trade fill events.
- **Hardware Biometric Authorization:** `local_auth: ^2.2.0` for biometric signature validation during capital commitment and emergency detachment actions.
- **Networking & Transport:** `dio: ^5.4.3+1` configured with JWT bearer authentication, automated token refresh interceptors, and exponential backoff retry mechanisms.
- **Accessibility & Internationalization:** `flutter/semantics.dart` providing full TalkBack and VoiceOver screen reader support, paired with WCAG 2.1 AA compliant color palettes (supporting deuteranopia and protanopia).

## Backend / Infra Touchpoints
- **Copy Trading Service (Prompt 265):**
  - **Leaderboard Directory API (`GET /api/v1/copytrading/strategies`):** Queries ranked master strategies with parameters `page`, `limit`, `sort_by` (roi, sharpe, mdd, auc), `asset_class`, and `risk_tier`.
  - **Strategy Detail API (`GET /api/v1/copytrading/strategies/{strategy_id}`):** Retrieves full master profile, trader biography, verified credentials, fee structure (10% to 15% HWM profit share), and trade history.
  - **Historical Equity Curve API (`GET /api/v1/copytrading/strategies/{strategy_id}/equity-curve`):** Returns historical daily/hourly NAV time series alongside Nifty 50 and Nifty Bank benchmark points.
  - **Allocation Intent API (`POST /api/v1/copytrading/subscriptions/allocate`):** Submits capital allocation request with `allocated_amount_inr`, `replication_mode`, `max_slippage_bps`, and `drawdown_stop_loss_pct`.
  - **Emergency Detach API (`POST /api/v1/copytrading/subscriptions/{subscription_id}/detach`):** Triggers immediate decoupling with `close_open_positions` boolean flag (ADR-0042, RUNBOOK-30).
  - **My Copied Strategies API (`GET /api/v1/copytrading/subscriptions/my-strategies`):** Fetches the authenticated follower active subscriptions, allocated capital, current NAV, cumulative PnL, and current High-Water Mark.
  - **Settlement History API (`GET /api/v1/copytrading/settlements/history`):** Returns historical weekly HWM profit-share settlement events with on-chain transaction hashes.
- **Market Data Service (Prompt 207):**
  - **Live Price Stream (`wss://ws.growww.in/v1/market/stream`):** Delivers real-time Last Traded Price (LTP) and depth quotes for underlying symbols held in replicated positions.
- **Portfolio & Holdings Service (Prompt 209):**
  - **Holdings & Margin API (`GET /api/v1/portfolio/holdings`):** Queries unencumbered cash balance and margin utilization to enforce allocation boundaries.
  - **Live Copy Trading WebSocket (`wss://ws.growww.in/v1/copytrading/stream`):** Pushes real-time master execution alerts, child replication fill confirmations, and circuit breaker detachment notifications.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Smart Contract Target:** `PerformanceFeeDistributor.sol` deployed on the permissioned Hyperledger Besu network under QBFT consensus.
- **On-Chain Settlement Verification:**
  - Weekly High-Water Mark profit-share distributions are executed on-chain. When a master trader earns a performance fee, the settlement transaction is mined into a Besu block.
  - The Flutter client displays an interactive on-chain receipt badge for each settlement entry. Tapping the badge opens the `OnChainProfitShareReceiptViewer` displaying:
    - Smart contract address (`0x...` for `PerformanceFeeDistributor.sol`).
    - Transaction hash with deep link to the internal Growww Besu block explorer.
    - Settled gross profit in INR, master fee cut (10% to 15%), and statutory GST deduction.
    - Merkle root proof verifying that the follower weekly NAV snapshot was included in the batch settlement.
- **Zero On-Chain PII Policy:**
  - Blockchain records contain zero personally identifiable information (zero PII). All events reference strictly pseudonymous identifiers: `strategy_id` UUID, masked `account_hash` (keccak256 digest of user ID and cryptographic salt), and tokenized asset contracts. No names, PAN numbers, or bank account details are ever committed to or read from the ledger.

## State Management Architecture & Copy Trading Engine
The feature leverages Riverpod code-generation to establish a reactive, unidirectional data flow:
- **`LeaderboardFilterNotifier` (`AutoDisposeNotifier<LeaderboardFilterState>`):** Maintains the current filter parameters (asset class, risk tier, time horizon, minimum win rate) and active sorting criteria.
- **`StrategyLeaderboardNotifier` (`AutoDisposeAsyncNotifier<List<MasterTraderSummary>>`):** Manages pagination, pull-to-refresh, and caching for the master strategy directory.
- **`MasterStrategyDetailNotifier` (`AutoDisposeFamilyAsyncNotifier<MasterTraderDetail, String>`):** Fetches comprehensive profile details and historical metrics for a specific master trader.
- **`EquityCurveNotifier` (`AutoDisposeFamilyAsyncNotifier<EquityCurveData, String>`):** Loads and processes historical NAV data points, normalizes dates, and computes relative performance against benchmark indices.
- **`CopyAllocationNotifier` (`AutoDisposeAsyncNotifier<AllocationWizardState>`):** Orchestrates the capital allocation journey: validates slider input against available wallet cash, validates minimum ticket size (e.g., INR 10,000), checks slippage tolerance (+/- 0.5%), and dispatches the allocation intent.
- **`ActiveCopiedPositionsNotifier` (`AutoDisposeStreamNotifier<List<CopiedPosition>>`):** Subscribes to the live copy trading WebSocket channel, continuously recalculating unrealized PnL, current slippage dispersion, and distance to the user-configured stop-loss circuit breaker.
- **`EmergencyDetachmentNotifier` (`AutoDisposeAsyncNotifier<DetachmentResult?>`):** Coordinates the one-tap emergency detachment protocol: dispatches biometric challenge, calls the detachment API, optimistically updates UI state, and handles position liquidation progress.
- **`SettlementHistoryNotifier` (`AutoDisposeAsyncNotifier<List<ProfitShareSettlementReceipt>>`):** Loads weekly HWM settlement receipts and handles cryptographic Merkle proof validation.

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold feature directory structure in `apps/growww_flutter/lib/features/copy_trading/`:
   - `domain/models/`: Domain entities and immutable data classes.
   - `domain/interfaces/`: Abstract repository contracts.
   - `data/datasources/`: REST and WebSocket API client implementations.
   - `data/repositories/`: Concrete repository implementations coordinating local caching and remote services.
   - `presentation/controllers/`: Riverpod state notifiers and providers.
   - `presentation/screens/`: Top-level screens (`CopyTradingDashboardScreen`, `MasterTraderProfileScreen`).
   - `presentation/widgets/`: Reusable presentation widgets (leaderboard card, equity curve chart, allocation sheet, detachment dialog).
2. Define domain enums and models in `domain/models/`:
   - `ReplicationMode`, `RiskTier`, `AssetClassFilter`, `StrategySortBy`, `DetachmentMode`.
   - `MasterTraderSummary`, `MasterTraderDetail`, `TraderPerformanceMetrics`, `EquityCurvePoint`, `CopiedPosition`, `CopySubscription`, `ProfitShareSettlementReceipt`.
3. Implement `ICopyTradingRepository` contract in `domain/interfaces/copy_trading_repository.dart` defining methods for fetching strategies, streaming positions, allocating capital, triggering detachment, and querying settlement history.
4. Implement `CopyTradingRemoteDataSource` in `data/datasources/` utilizing `Dio` for REST endpoints and `WebSocketChannel` for the live copy trading event feed.
5. Create `CopyTradingRepositoryImpl` in `data/repositories/` to coordinate network calls, error mapping, and local state caching.
6. Build Riverpod state notifiers in `presentation/controllers/`:
   - Implement `StrategyLeaderboardNotifier` supporting infinite scrolling and debounce search.
   - Implement `EquityCurveNotifier` with timeframe caching (1M, 3M, 6M, 1Y, ALL).
   - Implement `ActiveCopiedPositionsNotifier` consuming the real-time WebSocket tick stream.
   - Implement `EmergencyDetachmentNotifier` with atomic optimistic state mutation.
7. Build `MandatorySebiRiskBanner` widget:
   - Render prominent statutory warning: "Copy Trading involves substantial risk of loss. Past performance is not indicative of future returns. 9 out of 10 individual traders in equity F&O incur net losses."
   - Embed accessible high-contrast warning icon and un-dismissible container styling.
8. Construct `StrategyLeaderboardWidget` and `MasterTraderCard`:
   - Display trader avatar, SEBI verification badge, trader handle, and strategy moniker.
   - Render key metric chips: 30D/90D ROI (+/-%), Max Drawdown (MDD), Sharpe Ratio, and Assets Under Copy (AUC in INR Crores/Lakhs).
   - Integrate visual sparkline preview of the 30-day equity trajectory.
9. Construct `MasterTraderProfileScreen`:
   - Layout sliver app bar with master overview, bio, SEBI RA/IA credentials, and trade frequency stats.
   - Embed `InteractiveEquityCurveWidget` using `fl_chart` with interactive touch tooltips, high-water mark overlay points, and benchmark index toggles (Nifty 50).
   - Render monthly return matrix heatmap showing percentage returns across months and years.
   - Add asset allocation breakdown doughnut chart (Equities vs Derivatives vs Commodities vs RWAs).
10. Build `CapitalAllocationBottomSheet`:
    - Display unencumbered wallet cash balance retrieved from Portfolio Service (Prompt 209).
    - Implement interactive slider and text field for INR capital allocation with minimum ticket boundary validation (e.g., INR 10,000).
    - Implement replication mode segmented control: Equity-Proportional ($Q_f = Q_m \times E_f / E_m$), Fixed-Ratio, or Cash-Proportional.
    - Implement slippage tolerance selector locked to +/- 0.5% maximum boundary with warning indicator.
    - Implement emergency stop-loss drawdown slider (e.g., -5%, -10%, -15%) that sets the local and server-side circuit breaker.
    - Integrate biometric prompt (`local_auth`) before confirming capital commitment.
11. Build `ActiveCopiedPositionsCard` and `CopiedPositionsListView`:
    - Render active replicated holdings with real-time mark-to-market PnL in INR and percentage.
    - Display master entry price alongside follower execution price, highlighting slippage in basis points.
    - Show real-time distance gauge indicating proximity to the emergency stop-loss threshold.
12. Build `EmergencyDetachmentDialog` and button:
    - Prominent crimson action trigger in the app bar and subscription detail view.
    - Two-step detachment confirmation modal presenting choice: "Liquidate All Open Positions at Market" or "Keep Positions Open for Manual Exit".
    - Execute instantaneous API call to `POST /api/v1/copytrading/subscriptions/{id}/detach`, immediately aborting active replication and canceling pending child orders.
13. Build `OnChainProfitShareReceiptViewer` modal:
    - Present weekly HWM settlement details, gross profit, master fee deduction, and net return.
    - Render Hyperledger Besu transaction hash, block height, and interactive link to internal block explorer.
    - Display cryptographic Merkle proof attestation confirming inclusion in `PerformanceFeeDistributor.sol`.
14. Implement digital accessibility and theme support:
    - Add semantic labels for TalkBack and VoiceOver on all metric chips, sliders, and chart points.
    - Support standard SEBI emerald/crimson palette, accessible cobalt blue / amber orange palette, and high-contrast monochrome mode.
15. Author unit and widget test suites:
    - Unit tests for allocation math, equity curve scaling, and slippage calculation.
    - Widget tests for the allocation slider, emergency detachment flow, and risk banner visibility.

## Interfaces / Contracts

```dart
// lib/features/copy_trading/domain/models/copy_trading_enums.dart

enum ReplicationMode {
  equityProportional,
  fixedRatio,
  cashProportional,
}

enum RiskTier {
  conservative,
  moderate,
  aggressive,
}

enum AssetClassFilter {
  all,
  equities,
  derivatives,
  commodities,
  tokenizedRwa,
}

enum StrategySortBy {
  roi,
  sharpeRatio,
  lowestDrawdown,
  assetsUnderCopy,
  winRate,
}

enum DetachmentMode {
  liquidateAllMarket,
  retainPositionsManual,
}

enum SubscriptionStatus {
  active,
  paused,
  circuitBreakerDetached,
  manuallyDetached,
  closed,
}

// lib/features/copy_trading/domain/models/trader_performance_metrics.dart

class TraderPerformanceMetrics {
  final double roi30d;
  final double roi90d;
  final double roi1y;
  final double roiAllTime;
  final double sharpeRatio;
  final double sortinoRatio;
  final double maxDrawdownPct;
  final double winRatePct;
  final double profitFactor;
  final int totalTrades;
  final int winningTrades;
  final int losingTrades;
  final double avgHoldingDurationHours;
  final double assetsUnderCopyInr;
  final int totalFollowers;

  const TraderPerformanceMetrics({
    required this.roi30d,
    required this.roi90d,
    required this.roi1y,
    required this.roiAllTime,
    required this.sharpeRatio,
    required this.sortinoRatio,
    required this.maxDrawdownPct,
    required this.winRatePct,
    required this.profitFactor,
    required this.totalTrades,
    required this.winningTrades,
    required this.losingTrades,
    required this.avgHoldingDurationHours,
    required this.assetsUnderCopyInr,
    required this.totalFollowers,
  });

  factory TraderPerformanceMetrics.fromJson(Map<String, dynamic> json) {
    return TraderPerformanceMetrics(
      roi30d: (json['roi_30d'] as num).toDouble(),
      roi90d: (json['roi_90d'] as num).toDouble(),
      roi1y: (json['roi_1y'] as num).toDouble(),
      roiAllTime: (json['roi_all_time'] as num).toDouble(),
      sharpeRatio: (json['sharpe_ratio'] as num).toDouble(),
      sortinoRatio: (json['sortino_ratio'] as num).toDouble(),
      maxDrawdownPct: (json['max_drawdown_pct'] as num).toDouble(),
      winRatePct: (json['win_rate_pct'] as num).toDouble(),
      profitFactor: (json['profit_factor'] as num).toDouble(),
      totalTrades: json['total_trades'] as int,
      winningTrades: json['winning_trades'] as int,
      losingTrades: json['losing_trades'] as int,
      avgHoldingDurationHours: (json['avg_holding_duration_hours'] as num).toDouble(),
      assetsUnderCopyInr: (json['assets_under_copy_inr'] as num).toDouble(),
      totalFollowers: json['total_followers'] as int,
    );
  }
}

// lib/features/copy_trading/domain/models/master_trader.dart

class MasterTraderSummary {
  final String strategyId;
  final String masterUserId;
  final String traderHandle;
  final String strategyName;
  final String avatarUrl;
  final bool isSebiRegistered;
  final String? sebiRegistrationNumber;
  final RiskTier riskTier;
  final List<String> primaryAssetClasses;
  final double performanceFeeRate; // e.g. 0.10 for 10%
  final TraderPerformanceMetrics metrics;

  const MasterTraderSummary({
    required this.strategyId,
    required this.masterUserId,
    required this.traderHandle,
    required this.strategyName,
    required this.avatarUrl,
    required this.isSebiRegistered,
    this.sebiRegistrationNumber,
    required this.riskTier,
    required this.primaryAssetClasses,
    required this.performanceFeeRate,
    required this.metrics,
  });

  factory MasterTraderSummary.fromJson(Map<String, dynamic> json) {
    return MasterTraderSummary(
      strategyId: json['strategy_id'] as String,
      masterUserId: json['master_user_id'] as String,
      traderHandle: json['trader_handle'] as String,
      strategyName: json['strategy_name'] as String,
      avatarUrl: json['avatar_url'] as String,
      isSebiRegistered: json['is_sebi_registered'] as bool,
      sebiRegistrationNumber: json['sebi_registration_number'] as String?,
      riskTier: RiskTier.values.firstWhere(
        (e) => e.name.toLowerCase() == (json['risk_tier'] as String).toLowerCase(),
      ),
      primaryAssetClasses: List<String>.from(json['primary_asset_classes'] as List),
      performanceFeeRate: (json['performance_fee_rate'] as num).toDouble(),
      metrics: TraderPerformanceMetrics.fromJson(
        json['metrics'] as Map<String, dynamic>,
      ),
    );
  }
}

class MasterTraderDetail extends MasterTraderSummary {
  final String biography;
  final String tradingMethodology;
  final Map<String, double> assetAllocationBreakdown;
  final List<MonthlyReturnEntry> monthlyReturns;
  final DateTime inceptionDate;

  const MasterTraderDetail({
    required super.strategyId,
    required super.masterUserId,
    required super.traderHandle,
    required super.strategyName,
    required super.avatarUrl,
    required super.isSebiRegistered,
    super.sebiRegistrationNumber,
    required super.riskTier,
    required super.primaryAssetClasses,
    required super.performanceFeeRate,
    required super.metrics,
    required this.biography,
    required this.tradingMethodology,
    required this.assetAllocationBreakdown,
    required this.monthlyReturns,
    required this.inceptionDate,
  });

  factory MasterTraderDetail.fromJson(Map<String, dynamic> json) {
    return MasterTraderDetail(
      strategyId: json['strategy_id'] as String,
      masterUserId: json['master_user_id'] as String,
      traderHandle: json['trader_handle'] as String,
      strategyName: json['strategy_name'] as String,
      avatarUrl: json['avatar_url'] as String,
      isSebiRegistered: json['is_sebi_registered'] as bool,
      sebiRegistrationNumber: json['sebi_registration_number'] as String?,
      riskTier: RiskTier.values.firstWhere(
        (e) => e.name.toLowerCase() == (json['risk_tier'] as String).toLowerCase(),
      ),
      primaryAssetClasses: List<String>.from(json['primary_asset_classes'] as List),
      performanceFeeRate: (json['performance_fee_rate'] as num).toDouble(),
      metrics: TraderPerformanceMetrics.fromJson(
        json['metrics'] as Map<String, dynamic>,
      ),
      biography: json['biography'] as String,
      tradingMethodology: json['trading_methodology'] as String,
      assetAllocationBreakdown: Map<String, double>.from(
        json['asset_allocation_breakdown'] as Map,
      ),
      monthlyReturns: (json['monthly_returns'] as List)
          .map((e) => MonthlyReturnEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
      inceptionDate: DateTime.parse(json['inception_date'] as String),
    );
  }
}

class MonthlyReturnEntry {
  final int year;
  final int month;
  final double returnPct;

  const MonthlyReturnEntry({
    required this.year,
    required this.month,
    required this.returnPct,
  });

  factory MonthlyReturnEntry.fromJson(Map<String, dynamic> json) {
    return MonthlyReturnEntry(
      year: json['year'] as int,
      month: json['month'] as int,
      returnPct: (json['return_pct'] as num).toDouble(),
    );
  }
}

// lib/features/copy_trading/domain/models/equity_curve_point.dart

class EquityCurvePoint {
  final DateTime timestamp;
  final double strategyNav;
  final double benchmarkNav;
  final double? highWaterMark;

  const EquityCurvePoint({
    required this.timestamp,
    required this.strategyNav,
    required this.benchmarkNav,
    this.highWaterMark,
  });

  factory EquityCurvePoint.fromJson(Map<String, dynamic> json) {
    return EquityCurvePoint(
      timestamp: DateTime.parse(json['timestamp'] as String),
      strategyNav: (json['strategy_nav'] as num).toDouble(),
      benchmarkNav: (json['benchmark_nav'] as num).toDouble(),
      highWaterMark: json['high_water_mark'] != null
          ? (json['high_water_mark'] as num).toDouble()
          : null,
    );
  }
}

// lib/features/copy_trading/domain/models/copied_position.dart

class CopiedPosition {
  final String positionId;
  final String subscriptionId;
  final String symbol;
  final String isin;
  final String side; // BUY or SELL
  final double quantity;
  final double masterAvgPrice;
  final double followerAvgPrice;
  final double slippageBps;
  final double currentLtp;
  final double unrealizedPnlInr;
  final double unrealizedPnlPct;
  final double? stopLossPrice;
  final double? takeProfitPrice;
  final DateTime executedAt;

  const CopiedPosition({
    required this.positionId,
    required this.subscriptionId,
    required this.symbol,
    required this.isin,
    required this.side,
    required this.quantity,
    required this.masterAvgPrice,
    required this.followerAvgPrice,
    required this.slippageBps,
    required this.currentLtp,
    required this.unrealizedPnlInr,
    required this.unrealizedPnlPct,
    this.stopLossPrice,
    this.takeProfitPrice,
    required this.executedAt,
  });

  factory CopiedPosition.fromJson(Map<String, dynamic> json) {
    return CopiedPosition(
      positionId: json['position_id'] as String,
      subscriptionId: json['subscription_id'] as String,
      symbol: json['symbol'] as String,
      isin: json['isin'] as String,
      side: json['side'] as String,
      quantity: (json['quantity'] as num).toDouble(),
      masterAvgPrice: (json['master_avg_price'] as num).toDouble(),
      followerAvgPrice: (json['follower_avg_price'] as num).toDouble(),
      slippageBps: (json['slippage_bps'] as num).toDouble(),
      currentLtp: (json['current_ltp'] as num).toDouble(),
      unrealizedPnlInr: (json['unrealized_pnl_inr'] as num).toDouble(),
      unrealizedPnlPct: (json['unrealized_pnl_pct'] as num).toDouble(),
      stopLossPrice: json['stop_loss_price'] != null
          ? (json['stop_loss_price'] as num).toDouble()
          : null,
      takeProfitPrice: json['take_profit_price'] != null
          ? (json['take_profit_price'] as num).toDouble()
          : null,
      executedAt: DateTime.parse(json['executed_at'] as String),
    );
  }
}

// lib/features/copy_trading/domain/models/copy_subscription.dart

class CopySubscription {
  final String subscriptionId;
  final String strategyId;
  final String strategyName;
  final String traderHandle;
  final double allocatedCapitalInr;
  final double currentEquityInr;
  final double highWaterMarkInr;
  final double cumulativeRealizedPnlInr;
  final double cumulativeUnrealizedPnlInr;
  final ReplicationMode replicationMode;
  final double maxSlippageBps;
  final double drawdownCircuitBreakerPct;
  final SubscriptionStatus status;
  final DateTime subscribedAt;

  const CopySubscription({
    required this.subscriptionId,
    required this.strategyId,
    required this.strategyName,
    required this.traderHandle,
    required this.allocatedCapitalInr,
    required this.currentEquityInr,
    required this.highWaterMarkInr,
    required this.cumulativeRealizedPnlInr,
    required this.cumulativeUnrealizedPnlInr,
    required this.replicationMode,
    required this.maxSlippageBps,
    required this.drawdownCircuitBreakerPct,
    required this.status,
    required this.subscribedAt,
  });

  factory CopySubscription.fromJson(Map<String, dynamic> json) {
    return CopySubscription(
      subscriptionId: json['subscription_id'] as String,
      strategyId: json['strategy_id'] as String,
      strategyName: json['strategy_name'] as String,
      traderHandle: json['trader_handle'] as String,
      allocatedCapitalInr: (json['allocated_capital_inr'] as num).toDouble(),
      currentEquityInr: (json['current_equity_inr'] as num).toDouble(),
      highWaterMarkInr: (json['high_water_mark_inr'] as num).toDouble(),
      cumulativeRealizedPnlInr: (json['cumulative_realized_pnl_inr'] as num).toDouble(),
      cumulativeUnrealizedPnlInr: (json['cumulative_unrealized_pnl_inr'] as num).toDouble(),
      replicationMode: ReplicationMode.values.firstWhere(
        (e) => e.name.toLowerCase() == (json['replication_mode'] as String).toLowerCase(),
      ),
      maxSlippageBps: (json['max_slippage_bps'] as num).toDouble(),
      drawdownCircuitBreakerPct: (json['drawdown_circuit_breaker_pct'] as num).toDouble(),
      status: SubscriptionStatus.values.firstWhere(
        (e) => e.name.toLowerCase() == (json['status'] as String).toLowerCase(),
      ),
      subscribedAt: DateTime.parse(json['subscribed_at'] as String),
    );
  }
}

// lib/features/copy_trading/domain/models/copy_allocation_request.dart

class CopyAllocationRequest {
  final String strategyId;
  final double allocatedAmountInr;
  final ReplicationMode replicationMode;
  final double maxSlippageBps; // default 50 bps (0.5%)
  final double drawdownStopLossPct; // e.g. 15.0 for -15%

  const CopyAllocationRequest({
    required this.strategyId,
    required this.allocatedAmountInr,
    required this.replicationMode,
    this.maxSlippageBps = 50.0,
    required this.drawdownStopLossPct,
  });

  Map<String, dynamic> toJson() => {
    'strategy_id': strategyId,
    'allocated_amount_inr': allocatedAmountInr,
    'replication_mode': replicationMode.name,
    'max_slippage_bps': maxSlippageBps,
    'drawdown_stop_loss_pct': drawdownStopLossPct,
  };
}

// lib/features/copy_trading/domain/models/profit_share_settlement_receipt.dart

class ProfitShareSettlementReceipt {
  final String settlementId;
  final String subscriptionId;
  final String strategyId;
  final String strategyName;
  final DateTime settlementPeriodStart;
  final DateTime settlementPeriodEnd;
  final double previousHighWaterMarkInr;
  final double newHighWaterMarkInr;
  final double netProfitInr;
  final double performanceFeeRate;
  final double performanceFeeAmountInr;
  final double gstDeductionInr;
  final double netFollowerCreditInr;
  final String onChainTxHash;
  final int blockNumber;
  final String smartContractAddress;
  final bool isMerkleProofVerified;

  const ProfitShareSettlementReceipt({
    required this.settlementId,
    required this.subscriptionId,
    required this.strategyId,
    required this.strategyName,
    required this.settlementPeriodStart,
    required this.settlementPeriodEnd,
    required this.previousHighWaterMarkInr,
    required this.newHighWaterMarkInr,
    required this.netProfitInr,
    required this.performanceFeeRate,
    required this.performanceFeeAmountInr,
    required this.gstDeductionInr,
    required this.netFollowerCreditInr,
    required this.onChainTxHash,
    required this.blockNumber,
    required this.smartContractAddress,
    required this.isMerkleProofVerified,
  });

  factory ProfitShareSettlementReceipt.fromJson(Map<String, dynamic> json) {
    return ProfitShareSettlementReceipt(
      settlementId: json['settlement_id'] as String,
      subscriptionId: json['subscription_id'] as String,
      strategyId: json['strategy_id'] as String,
      strategyName: json['strategy_name'] as String,
      settlementPeriodStart: DateTime.parse(json['settlement_period_start'] as String),
      settlementPeriodEnd: DateTime.parse(json['settlement_period_end'] as String),
      previousHighWaterMarkInr: (json['previous_hwm_inr'] as num).toDouble(),
      newHighWaterMarkInr: (json['new_hwm_inr'] as num).toDouble(),
      netProfitInr: (json['net_profit_inr'] as num).toDouble(),
      performanceFeeRate: (json['performance_fee_rate'] as num).toDouble(),
      performanceFeeAmountInr: (json['performance_fee_amount_inr'] as num).toDouble(),
      gstDeductionInr: (json['gst_deduction_inr'] as num).toDouble(),
      netFollowerCreditInr: (json['net_follower_credit_inr'] as num).toDouble(),
      onChainTxHash: json['on_chain_tx_hash'] as String,
      blockNumber: json['block_number'] as int,
      smartContractAddress: json['smart_contract_address'] as String,
      isMerkleProofVerified: json['is_merkle_proof_verified'] as bool,
    );
  }
}

// lib/features/copy_trading/domain/interfaces/copy_trading_repository.dart

abstract class ICopyTradingRepository {
  Future<List<MasterTraderSummary>> fetchLeaderboard({
    required int page,
    required int limit,
    required StrategySortBy sortBy,
    AssetClassFilter? assetClass,
    RiskTier? riskTier,
  });

  Future<MasterTraderDetail> fetchStrategyDetail(String strategyId);

  Future<List<EquityCurvePoint>> fetchEquityCurve({
    required String strategyId,
    required String timeframe, // 1M, 3M, 6M, 1Y, ALL
  });

  Future<CopySubscription> allocateCapital(CopyAllocationRequest request);

  Future<void> emergencyDetach({
    required String subscriptionId,
    required DetachmentMode mode,
  });

  Future<List<CopySubscription>> fetchMySubscriptions();

  Stream<List<CopiedPosition>> streamActivePositions({
    required String subscriptionId,
  });

  Future<List<ProfitShareSettlementReceipt>> fetchSettlementHistory();
}
```

## Security & Compliance Notes
- **Mandatory SEBI Risk Warning Disclosures:**
  - Copy trading and algorithmic replication involve substantial financial risk, especially when replicating intraday derivative strategies. In strict compliance with SEBI circulars on social trading, copy trading, and investor risk awareness, the client renders an un-dismissible, high-contrast banner at the top of the marketplace:
    `"Risk Disclosure on Copy Trading: 9 out of 10 individual traders in equity Futures and Options Segment incurred net losses. Past performance of a Master Trader is not indicative of future returns. Replication involves execution slippage and market risk."`
  - Prior to capital commitment in `CapitalAllocationBottomSheet`, investors must review and acknowledge a SEBI-mandated risk declaration modal detailing potential loss of capital, execution delay risks, and performance fee deductions.
- **Verification & Anti-Spoofing of Master Performance:**
  - To prevent manipulated or simulated track records, the client displays verified shield badges exclusively for master traders whose metrics are derived directly from exchange matching engine trade execution logs (Prompt 205).
  - Unregistered or unverified traders cannot publish strategies publicly. Traders with SEBI Research Analyst (RA) or Investment Adviser (IA) registrations display their authenticated registration numbers with a clickable verification certificate.
- **Strict Slippage Clamping & Follower Transparency (ADR-0042):**
  - Followers are shielded from predatory front-running and adverse selection through a client-configurable slippage barrier capped at +/- 0.5% (50 basis points) relative to the master fill price.
  - The UI explicitly visualizes the master entry price alongside the follower fill price and displays the exact slippage incurred in basis points on every replicated position card.
- **High-Water Mark (HWM) Fee Transparency:**
  - Performance fees (10% to 15%) are deducted strictly when cumulative net equity exceeds the historical High-Water Mark.
  - The client provides total transparency by displaying the historical HWM baseline, gross new profits, statutory GST deduction, and the net performance fee retained by the master trader. No performance fee is charged during recovery periods following a drawdown.
- **One-Tap Emergency Detachment & Circuit Breaker (ADR-0042, RUNBOOK-30):**
  - Follower autonomy is guaranteed via an instantaneous, one-tap emergency detachment button. When pressed, the client issues a high-priority cancellation command that immediately aborts child order replication.
  - The follower is presented with a clear choice to either market-liquidate all replicated positions immediately or retain them for manual disposal.
  - If the follower account reaches the user-set maximum drawdown limit (e.g., -15%), the system automatically activates the circuit breaker, notifies the user via push/WebSocket, and halts further replication.
- **Zero On-Chain PII Guarantee:**
  - All on-chain settlement receipts and Merkle tree attestations published to Hyperledger Besu contain zero Personally Identifiable Information (PII). No investor names, Aadhaar/PAN details, contact information, or bank details are stored on-chain or rendered from blockchain logs.

## Acceptance Criteria
- [ ] `CopyTradingDashboardScreen` renders responsive tab navigation ("Strategy Marketplace", "My Strategies", "Mirrored Positions", "Settlement History") across Android, iOS, macOS, Windows, and Linux.
- [ ] `StrategyLeaderboardWidget` displays verified master trader cards with live rank badges, SEBI registration icons, Assets Under Copy, 30D/90D ROI, and Maximum Drawdown metrics.
- [ ] `LeaderboardFilterBar` enables instant sorting by ROI, Sharpe Ratio, Lowest Drawdown, and Assets Under Copy, as well as filtering by asset class and risk tier.
- [ ] `InteractiveEquityCurveWidget` utilizes `fl_chart` to render smooth historical NAV lines against benchmark indices (Nifty 50) with interactive touch tooltips and timeframe buttons (1M, 3M, 6M, 1Y, ALL).
- [ ] `MasterTraderProfileScreen` displays trader biography, verified SEBI credentials, asset allocation doughnut chart, and monthly return heatmap matrix.
- [ ] `CapitalAllocationBottomSheet` enforces available unencumbered cash limits, validates minimum ticket size (e.g. INR 10,000), enforces +/- 0.5% maximum slippage clamping, and sets emergency drawdown circuit breakers.
- [ ] High-value capital allocation and emergency detachment actions require biometric step-up authorization (`local_auth`).
- [ ] `ActiveCopiedPositionsCard` streams real-time mark-to-market PnL in INR and percentage, displaying master fill price vs follower execution price with slippage in basis points.
- [ ] Emergency detachment button triggers immediate decoupling from master strategy, providing the option to market-liquidate positions or retain them for manual management.
- [ ] `OnChainProfitShareReceiptViewer` renders weekly HWM settlement details, Hyperledger Besu transaction hashes, and Merkle proof verification status.
- [ ] Persistent SEBI statutory risk warning banner is rendered at the top of the copy trading marketplace without dismissibility.
- [ ] Application lifecycle observer suspends WebSocket streaming and cancels animation tickers when the app is placed in the background.
- [ ] Accessible color modes (standard red/green, color-blind cobalt blue / amber orange, high-contrast monochrome) fulfill WCAG 2.1 AA 4.5:1 contrast ratio requirements.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 501 (Flutter Multi-Platform Project Scaffolding)
  - Prompt 502 (Flutter App Architecture & Riverpod State Management)
  - Prompt 503 (Flutter Design System & Theming)
  - Prompt 523 (Accessibility & Multi-Language Localization)
  - Prompt 525 (Flutter API Client Layer & WebSocket Gateway)
- **Backend Dependencies:**
  - Prompt 265 (Copy Trading & Proportional Replication Service)
  - Prompt 207 (Market Data & Real-Time Broadcast Service)
  - Prompt 209 (Portfolio & Holdings Service)
  - Prompt 204 (Order Service & Lifecycle Gateway)
  - Prompt 206 (Risk & Margin Checks Service)
  - Prompt 210 (Fee & Realized PnL Engine)
  - Prompt 203 (Wallet & Account Service)
- **Blockchain Dependencies:**
  - Prompt 309 (Settlement & Performance Fee Distributor Smart Contracts on Hyperledger Besu)
  - Prompt 329 (On-Chain Proof-of-Reserve & Verification Gateway)
- **Downstream Blockers:**
  - Prompt 506 (Flutter Home / Dashboard Screen integration)
  - Prompt 508 (Flutter Security Detail Screen copy-trading recommendations)
  - Prompt 509 (Flutter Order Placement Flow manual position takeovers)
  - Prompt 902 (Cross-Platform End-to-End Test Suite)
