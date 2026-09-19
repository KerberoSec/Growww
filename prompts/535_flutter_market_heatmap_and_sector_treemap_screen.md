# 535 - Flutter Real-Time Market Heatmap & Sector Treemap Screen

## Purpose
Provides institutional and retail traders with instant visual comprehension of sector performance, market-wide liquidity, and index gainers/losers via real-time hierarchical treemaps. During peak market volatility (market open auctions, RBI monetary policy announcements, Union Budget presentations, and macroeconomic data releases), traditional tabular watchlists and numerical tickers induce severe cognitive fatigue. A 2D spatial, squarified treemap transforms thousands of simultaneous price updates, trading volumes, and market capitalizations into an intuitive visual topology. Traders can instantly evaluate capital rotations across major economic sectors, identify outperforming tokenized Real-World Assets (RWAs), observe liquidity concentrations, and detect market breadth anomalies at a glance.

## What You Are Building
A responsive, high-performance Flutter visualization module rooted in `lib/screens/market/market_treemap_screen.dart` (and supporting sub-packages under `apps/growww_flutter/lib/features/market_treemap/`):
- `MarketTreemapScreen`: The primary container widget providing adaptive multi-platform support across mobile (portrait/landscape) and desktop (macOS, Windows, Linux), featuring index filters, metric controls, breadcrumbs, and real-time connection telemetry.
- `SquarifiedTreemapWidget`: A custom layout widget implementing the Bruls-Huizing-van Wijk squarified treemap algorithm to recursively partition arbitrary screen viewports into rectangular nodes with aspect ratios optimized near 1.0 (golden squares), preventing thin unreadable slivers.
- `HierarchicalSectorNodeWidget`: Nested sector group containers (e.g., Nifty 50, Banking, Information Technology, Energy, Automobile, Pharmaceuticals, Tokenized RWAs) featuring structural boundary borders, sector labels, aggregate market capitalization weightings, and cumulative sector percentage change indicators.
- `SecurityLeafTile`: Individual security leaf node widget displaying ticker symbol, company logo icon, Last Traded Price (LTP) in INR, 1-day percentage change (+/-%), 24-hour trading turnover, and specialized badges for tokenized RWAs.
- `HeatmapGradientEngine`: Bi-directional color interpolation service mapping continuous percentage returns to accessible color scales, including statutory SEBI-compliant palettes (standard emerald green to crimson red), color-blind accessible modes (cobalt blue to amber orange), and high-contrast monochrome modes.
- `ZoomableTreemapViewport`: Interactive canvas supporting multi-touch pinch-to-zoom, smooth panning, double-tap zoom-into-sector drilldown, and breadcrumb navigation trail (`All Markets > Financial Services > Private Banks > HDFC Bank`).
- `TreemapMetricsToolbar`: Metric dimension toggles enabling dynamic recalculation of tile areas based on Market Capitalization, 24-Hour Trading Turnover, Market Depth Liquidity, or Open Interest (for derivative underlyings), paired with temporal return filters (1D, 1W, 1M, 1Y).
- `SecurityPreviewBottomSheet`: Instant inspection modal opened upon tapping any leaf tile, displaying mini candlestick sparklines, bid/ask depth spread, 52-week price ranges, on-chain Proof-of-Reserve verification, and direct Buy/Sell quick-action triggers.

## Scope Boundaries
- **In Scope:**
  - Mathematical execution of the Bruls-Huizing-van Wijk squarified treemap layout algorithm in Dart.
  - Multi-level hierarchical grouping (Index to Sector to Industry to Security Leaf).
  - Real-time WebSocket price, volume, and depth streaming with selective dirty-node repainting optimizations (`CustomPainter` and `RepaintBoundary`).
  - Fluid 60/120 FPS transitions and animated sector drilldown with breadcrumb state preservation.
  - Sizing rectangles dynamically by Market Capitalization, 24-Hour Turnover, or Consolidated Market Depth.
  - Dynamic color interpolation across three accessible palettes adhering to SEBI digital accessibility standards.
  - Interactive multi-touch navigation (pinch, pan, tap, double-tap) and desktop keyboard/mouse wheel zoom.
  - Visualizing tokenized RWA market capitalization ratios alongside traditional equities.
- **Out of Scope / Handled Elsewhere:**
  - Full candlestick technical chart analysis and multi-timeframe indicator suites (handled in Prompt 508).
  - Order execution sheet, order routing, and bracket order placement (handled in Prompt 509).
  - Server-side WebSocket market feed multiplexing and order book depth aggregation (handled in Prompt 207).
  - Master data sector classification taxonomy ingestion and corporate action rebalancing (handled in Prompt 407).
  - On-chain smart contract token minting, burning, and custodian DvP clearing (handled in Prompt 304, Prompt 306, Prompt 329).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (v2.5+) with code-generated `AsyncNotifier` and `StreamNotifier` state primitives.
- **Rendering Pipeline:** Hardware-accelerated Flutter Impeller / Skia engine utilizing custom `RenderBox` and `CustomPainter` implementations wrapped with `RepaintBoundary` to prevent canvas invalidation across unmodified nodes.
- **Mathematical Layout Engine:** Custom recursive squarified treemap algorithm written in pure Dart, targeting aspect ratios $\max(w/h, h/w) \approx 1.0$ with zero external heavyweight layout dependencies.
- **Real-Time Data Streaming:** `web_socket_channel: ^3.0.0` with binary Protocol Buffer / typed JSON deserialization for high-frequency tick ingestion.
- **Vector Mathematics & Viewport Physics:** `vector_math: ^2.1.4` for 2D matrix affine transformations during viewport panning, scaling, and constraint clamping.
- **Accessibility & Contrast Compliance:** `flutter/semantics.dart` providing screen reader (TalkBack / VoiceOver) semantic labels and dynamic contrast ratio calculation fulfilling WCAG 2.1 AA standards.

## Backend / Infra Touchpoints
- **Market Data Service (Prompt 207):**
  - **Treemap Snapshot API (`GET /api/v1/market/treemap/snapshot`):** Fetches the baseline hierarchical market tree for a selected index (e.g., Nifty 50, Nifty 500, Tokenized RWAs), including sector classifications, baseline previous close prices, issued shares, market caps, and initial trading turnover.
  - **Real-Time Treemap WebSocket (`wss://ws.growww.in/v1/market/treemap/stream`):** Delivers sub-100ms delta packets containing security symbol, last traded price (LTP), net change, percentage change, cumulative turnover, and top-5 bid/ask depth volume.
- **Master Data Management (Prompt 407):**
  - **Sector Classification Taxonomy API (`GET /api/v1/master/sectors/hierarchy`):** Delivers the authoritative SEBI/NSE industry classification hierarchy (Macro-Economic Sector, Sector, Industry, Basic Industry) and constituent mapping.
- **Tokenized RWA Service (Prompt 308 / 203):**
  - **RWA Market Cap & Reserve API (`GET /api/v1/rwa/market-caps`):** Retrieves circulating token supply, oracle price feeds, and verified vault reserve balances for tokenized commodities (`gGOLD`, `gSILVER`) and tokenized sovereign debt (`gGSEC`).

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Tokenized RWA Market Capitalization Visualization:**
  - Tokenized real-world assets (e.g., `gGOLD`, `gSILVER`, `gGSEC`) are integrated directly into the treemap as a dedicated top-level sector or alongside traditional commodities.
  - Sizing for RWA nodes is derived dynamically from total circulating token supply on Hyperledger Besu multiplied by verified reference oracle prices.
- **On-Chain Proof-of-Reserve Verification Badge:**
  - Each tokenized RWA leaf tile displays a cryptographic shield badge. Tapping this badge opens an audit dialog showing the ERC-3643 smart contract address (`DigitalSecurityToken.sol`), the latest on-chain Sparse Merkle Tree Proof-of-Reserve root hash, and accredited custodian vault verification timestamps.
- **Zero On-Chain PII Policy:**
  - The client displays only publicly attested asset-level telemetry: verified token supply, smart contract addresses, and cryptographic reserve proofs. No personal investor identities, demat account numbers, or individual wallet balances are queried or exposed on-chain.

## State Management Architecture & Treemap Layout Engine
- **Squarified Treemap Layout Algorithm (Bruls-Huizing-van Wijk):**
  - Given an arbitrary bounding rectangle with dimensions $(W, H)$, items are sorted in descending order of weight (e.g., market capitalization or volume).
  - The algorithm fills sub-rectangles along the shorter edge $S = \min(W, H)$ by calculating aspect ratios $\max(w/h, h/w)$. It continues accumulating items into the active row as long as the maximum aspect ratio improves (approaches 1.0). Once adding an additional item degrades the aspect ratio, the current row is locked, its geometry is subtracted from the container, and the algorithm recurses on the remaining space.
  - Sub-sectors are recursively partitioned within their parent sector bounding boxes, guaranteeing strict visual hierarchy.
- **Riverpod State Providers:**
  - `TreemapConfigNotifier (`AutoDisposeNotifier<TreemapConfigState>`):` Manages selected index (Nifty 50, Nifty Bank, Nifty 500, RWAs), active sizing metric (Market Cap, Turnover, Market Depth), active time horizon (1D, 1W, 1M, 1Y), and active color palette mode.
  - `TreemapHierarchyNotifier (`AutoDisposeAsyncNotifier<TreemapHierarchyState>`):` Loads and caches the hierarchical sector tree from REST snapshots, merging static taxonomy with live baseline quotes.
  - `TreemapRealtimeStreamNotifier (`AutoDisposeStreamNotifier<TreemapTickBatch>`):` Consumes high-frequency WebSocket tick bursts, batching updates every 16ms (60 FPS) or 8ms (120 FPS) to match display VSYNC, updating in-memory leaf nodes without triggering full widget tree reconstructions.
  - `TreemapViewportNotifier (`AutoDisposeNotifier<TreemapViewportState>`):` Maintains matrix translation, zoom scale factor, active breadcrumb navigation path, and animated focus transitions during sector drilldowns.
- **Selective Dirty Repainting Architecture:**
  - Individual leaf nodes are rendered with dedicated `RepaintBoundary` wrappers or directly painted via a lightweight `CustomPainter` that evaluates node equality. Unchanged tiles bypass repaint cycles completely, preventing mobile CPU/GPU thermal throttling.

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold feature directories under `apps/growww_flutter/lib/features/market_treemap/`: `domain/models/`, `domain/algorithms/`, `domain/interfaces/`, `data/datasources/`, `data/repositories/`, `presentation/controllers/`, `presentation/painters/`, `presentation/screens/`, and `presentation/widgets/`.
2. Implement domain model contracts in `domain/models/`: `TreemapNode`, `SectorGroupNode`, `SecurityLeafNode`, `TreemapMetricType`, `HeatmapColorMode`, and `TreemapTickUpdate`.
3. Implement the pure Dart `SquarifiedTreemapAlgorithm` in `domain/algorithms/squarified_treemap.dart`, supporting recursive area partitioning, aspect ratio minimization, padding boundaries, and container aspect ratio adaptation.
4. Implement `HeatmapColorEngine` in `domain/services/heatmap_color_engine.dart` with mathematical linear and non-linear color interpolation across:
   - Standard Palette: Crimson Red (-3% or worse) to Charcoal Neutral (0%) to Emerald Green (+3% or better).
   - Accessible Color-Blind Palette: Amber Orange (-3%) to Charcoal Neutral (0%) to Cobalt Blue (+3%).
   - High-Contrast Monochrome: Charcoal Gray (-3%) to Pure White/Black with prominent signed numerical text tags.
5. Create `ITreemapRepository` and concrete `TreemapRepository` in `data/repositories/` to coordinate REST snapshot fetching (`GET /api/v1/market/treemap/snapshot`) and WebSocket streaming (`wss://ws.growww.in/v1/market/treemap/stream`).
6. Implement `TreemapConfigNotifier`, `TreemapHierarchyNotifier`, and `TreemapRealtimeStreamNotifier` using Riverpod code-generation, enforcing stream throttling to respect display VSYNC boundaries.
7. Build `TreemapCustomPainter` extending `CustomPainter` to efficiently render nested sector containers, sector title headers, leaf tile rectangles, and micro-borders using low-level `Canvas.drawRect` and `ParagraphBuilder` text rendering.
8. Create `SecurityLeafTileWidget` featuring ticker symbol, formatted INR price, signed percentage change, volume badges, and tokenized RWA shield indicators with dynamic font-size scaling based on tile dimensions.
9. Implement `ZoomableTreemapViewport` wrapping an `InteractiveViewer` with matrix transformation constraints, supporting pinch-to-zoom (0.5x to 5.0x), pan inertia, and double-tap zoom gestures.
10. Build `TreemapBreadcrumbBar` widget displaying interactive breadcrumb trails (`All > Sector > Industry > Stock`) allowing users to navigate upwards from drilled-down sector views.
11. Build `TreemapMetricsToolbar` allowing traders to toggle between sizing dimensions (Market Cap, Volume, Depth) and performance return horizons (1D, 1W, 1M, 1Y).
12. Build `SecurityPreviewBottomSheet` displaying mini candlestick sparklines, bid/ask spread, 52-week high/low bar, RWA Proof-of-Reserve verification, and direct order routing buttons.
13. Integrate SEBI digital accessibility compliance by embedding `Semantics` widgets across all nodes, announcing company name, sector, market cap, and percentage change to TalkBack and VoiceOver screen readers.
14. Optimize battery and rendering performance by establishing an application lifecycle observer that pauses WebSocket tick consumption and cancels ticker frames when the screen is hidden or backgrounded.
15. Write unit tests for `SquarifiedTreemapAlgorithm` verifying zero overlapping rects and aspect ratio bounds, and write widget tests verifying color interpolation and metric switching.

## Interfaces / Contracts

```dart
// lib/features/market_treemap/domain/models/treemap_enums.dart

enum TreemapMetricType {
  marketCap,
  volume24h,
  marketDepthLiquidity,
  openInterest,
}

enum HeatmapColorMode {
  standardRedGreen,
  accessibleBlueOrange,
  highContrastMonochrome,
}

enum TreemapTimeHorizon {
  oneDay,
  oneWeek,
  oneMonth,
  oneYear,
}

// lib/features/market_treemap/domain/models/treemap_geometry.dart

class TreemapRect {
  final double left;
  final double top;
  final double width;
  final double height;

  const TreemapRect({
    required this.left,
    required this.top,
    required this.width,
    required this.height,
  });

  double get right => left + width;
  double get bottom => top + height;
  double get area => width * height;
  double get aspectRatio => width > height ? width / height : height / width;

  TreemapRect copyWith({
    double? left,
    double? top,
    double? width,
    double? height,
  }) {
    return TreemapRect(
      left: left ?? this.left,
      top: top ?? this.top,
      width: width ?? this.width,
      height: height ?? this.height,
    );
  }
}

// lib/features/market_treemap/domain/models/treemap_node.dart

abstract class TreemapNode {
  final String id;
  final String name;
  final double value; // Sizing weight (Market Cap or Volume or Depth)
  final TreemapRect rect;

  const TreemapNode({
    required this.id,
    required this.name,
    required this.value,
    this.rect = const TreemapRect(left: 0, top: 0, width: 0, height: 0),
  });

  TreemapNode copyWithRect(TreemapRect newRect);
}

class SectorGroupNode extends TreemapNode {
  final List<TreemapNode> children;
  final double cumulativeChangePercentage;
  final int totalConstituents;

  const SectorGroupNode({
    required super.id,
    required super.name,
    required super.value,
    super.rect,
    required this.children,
    required this.cumulativeChangePercentage,
    required this.totalConstituents,
  });

  @override
  SectorGroupNode copyWithRect(TreemapRect newRect) {
    return SectorGroupNode(
      id: id,
      name: name,
      value: value,
      rect: newRect,
      children: children,
      cumulativeChangePercentage: cumulativeChangePercentage,
      totalConstituents: totalConstituents,
    );
  }
}

class SecurityLeafNode extends TreemapNode {
  final String symbol;
  final String isin;
  final String sectorId;
  final double ltp;
  final double previousClose;
  final double changeInr;
  final double changePercentage;
  final double turnoverInr;
  final double marketDepthBidVolume;
  final double marketDepthAskVolume;
  final bool isTokenizedRwa;
  final String? rwaContractAddress;
  final String? rwaProofOfReserveHash;

  const SecurityLeafNode({
    required super.id,
    required super.name,
    required super.value,
    super.rect,
    required this.symbol,
    required this.isin,
    required this.sectorId,
    required this.ltp,
    required this.previousClose,
    required this.changeInr,
    required this.changePercentage,
    required this.turnoverInr,
    required this.marketDepthBidVolume,
    required this.marketDepthAskVolume,
    this.isTokenizedRwa = false,
    this.rwaContractAddress,
    this.rwaProofOfReserveHash,
  });

  @override
  SecurityLeafNode copyWithRect(TreemapRect newRect) {
    return SecurityLeafNode(
      id: id,
      name: name,
      value: value,
      rect: newRect,
      symbol: symbol,
      isin: isin,
      sectorId: sectorId,
      ltp: ltp,
      previousClose: previousClose,
      changeInr: changeInr,
      changePercentage: changePercentage,
      turnoverInr: turnoverInr,
      marketDepthBidVolume: marketDepthBidVolume,
      marketDepthAskVolume: marketDepthAskVolume,
      isTokenizedRwa: isTokenizedRwa,
      rwaContractAddress: rwaContractAddress,
      rwaProofOfReserveHash: rwaProofOfReserveHash,
    );
  }

  SecurityLeafNode copyWithTick({
    required double newLtp,
    required double newChangeInr,
    required double newChangePercentage,
    required double newTurnoverInr,
    double? newBidVolume,
    double? newAskVolume,
  }) {
    return SecurityLeafNode(
      id: id,
      name: name,
      value: value,
      rect: rect,
      symbol: symbol,
      isin: isin,
      sectorId: sectorId,
      ltp: newLtp,
      previousClose: previousClose,
      changeInr: newChangeInr,
      changePercentage: newChangePercentage,
      turnoverInr: newTurnoverInr,
      marketDepthBidVolume: newBidVolume ?? marketDepthBidVolume,
      marketDepthAskVolume: newAskVolume ?? marketDepthAskVolume,
      isTokenizedRwa: isTokenizedRwa,
      rwaContractAddress: rwaContractAddress,
      rwaProofOfReserveHash: rwaProofOfReserveHash,
    );
  }
}

// lib/features/market_treemap/domain/models/treemap_tick_update.dart

class TreemapTickUpdate {
  final String symbol;
  final double ltp;
  final double changeInr;
  final double changePercentage;
  final double turnoverInr;
  final double totalBidVolume;
  final double totalAskVolume;
  final DateTime timestamp;

  const TreemapTickUpdate({
    required this.symbol,
    required this.ltp,
    required this.changeInr,
    required this.changePercentage,
    required this.turnoverInr,
    required this.totalBidVolume,
    required this.totalAskVolume,
    required this.timestamp,
  });
}

// lib/features/market_treemap/domain/interfaces/i_treemap_repository.dart

abstract class ITreemapRepository {
  Future<SectorGroupNode> fetchTreemapSnapshot({
    required String indexCode,
    required TreemapMetricType metricType,
    required TreemapTimeHorizon timeHorizon,
  });

  Stream<List<TreemapTickUpdate>> streamTreemapTicks({
    required String indexCode,
  });

  Future<List<String>> fetchAvailableIndices();
}
```

## Security & Compliance Notes
- **SEBI Digital Accessibility Guidelines & Color-Blind Usability:** Standard red/green financial heatmap representations create severe visibility barriers for approximately 8% of male and 0.5% of female traders experiencing deuteranopia or protanopia. In compliance with the SEBI Master Circular on Digital Accessibility and Web Content Accessibility Guidelines (WCAG 2.1 AA), the application provides an instantaneous toggle to an Accessible Cobalt Blue / Amber Orange palette and a High-Contrast Monochrome mode. Text labels rendered on top of colored tiles must maintain a minimum contrast ratio of 4.5:1 against the computed background color.
- **Battery-Efficient Repainting & Client Thermal Protection:** Unbounded rendering of high-throughput market ticks (up to 1,000 updates/second during market open) can quickly drain mobile batteries and induce CPU thermal throttling. The treemap engine enforces a 16ms VSYNC render throttle (capped at 60 FPS on standard displays and 120 FPS on ProMotion/high-refresh displays). Ticks are accumulated into an in-memory buffer and flushed synchronously on display frame callbacks. Furthermore, an `AppLifecycleListener` halts WebSocket stream subscription and cancels painter tickers when the app transitions into the background or the screen loses focus.
- **Selective Repaint Boundary Architecture:** Each individual sector group and leaf node maintains isolated render state. Modifying a single stock's percentage return does not invalidate the full-screen render canvas, isolating repaints strictly to the updated leaf geometry.
- **Market Data Integrity & Consolidated Depth Guardrails:** Visual treemap depths must reflect verified consolidated book depth from the exchange matching engine (Prompt 207) and not unverified mock feeds, preventing spoofed bid/ask visual inflation.
- **Zero PII & Secure Communication Channels:** The treemap module queries only public market metadata and signed telemetry. All network communications are conducted over TLS 1.3 for HTTPS REST calls and WSS for WebSocket streams, enforcing certificate pinning without capturing or transmitting personal identifiable information (PII).

## Acceptance Criteria
- [ ] `MarketTreemapScreen` renders a responsive, squarified treemap layout on both mobile (Android, iOS) and desktop (macOS, Windows, Linux) platforms.
- [ ] Squarified treemap layout algorithm accurately computes non-overlapping rectangular coordinates with aspect ratios targeting 1.0.
- [ ] Hierarchical sector boundaries display sector names, constituent counts, and cumulative sector percentage changes.
- [ ] Metric selector dynamically recalculates tile dimensions according to Market Capitalization, 24-Hour Turnover, or Market Depth Liquidity without layout failure.
- [ ] Color interpolation engine smoothly grades returns from -3% (deep loss) through 0% (neutral charcoal) to +3% (deep gain).
- [ ] SEBI accessibility compliant color-blind mode (Cobalt Blue / Amber Orange) and High-Contrast mode fulfill WCAG 2.1 AA minimum 4.5:1 contrast standards.
- [ ] WebSocket streaming updates leaf prices, turnover, and percentage changes in real time while throttling repaints to display VSYNC.
- [ ] Selective `RepaintBoundary` architecture ensures sustained 60/120 FPS performance during rapid market updates with zero UI jank.
- [ ] Pinch-to-zoom, pan, double-tap zoom, and breadcrumb bar allow fluid drilldown from Top Index to Sector to Individual Security.
- [ ] Tapping a leaf node opens `SecurityPreviewBottomSheet` presenting price metrics, depth spread, and quick trade action buttons.
- [ ] Tokenized RWA assets render distinct on-chain badges displaying ERC-3643 contract metadata and Proof-of-Reserve verification hashes.
- [ ] Application lifecycle observer automatically pauses WebSocket streaming and cancels animation tickers when the app is backgrounded.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Multi-Platform Project Scaffolding), Prompt 502 (Flutter App Architecture & Riverpod State Management), Prompt 503 (Flutter Design System & Theming), Prompt 523 (Accessibility & Multi-Language Localization), Prompt 525 (Flutter API Client Layer).
- **Backend Dependencies:** Prompt 207 (Market Data & Real-Time Broadcast Service), Prompt 407 (Master Data Management & Sector Taxonomy Service), Prompt 203 (Wallet & Ledger Service), Prompt 243 (MCX Commodity & Warehouse Receipt Gateway).
- **Blockchain Dependencies:** Prompt 308 (Proof-of-Reserve Registry Contract), Prompt 306 (Digital Security Token ERC-3643 Engine).
- **Downstream Blockers:** Prompt 506 (Flutter Home / Dashboard Screen), Prompt 507 (Flutter Market & Watchlist Screen), Prompt 508 (Flutter Security Detail Screen), Prompt 509 (Flutter Order Placement Flow), Prompt 902 (Cross-Platform End-to-End Test Suite).
