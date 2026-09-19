# 540 - Flutter BTC/USDT Live TradingView Chart & Market Watch Screen

## Purpose
High-frequency cryptocurrency trading, real-time portfolio margin valuation, and algorithmic technical analysis require instantaneous, jitter-free visual feedback for benchmark digital assets. In volatile crypto markets, Bitcoin (BTC/USDT) experiences intense quote traffic, frequent multi-exchange price discovery events, and sudden liquidity shifts across depth tiers. Standard mobile and desktop chart implementations frequently suffer from micro-stutter, cumulative UI thread garbage collection latency, visual layout jitter caused by variable-width numeric glyphs, and dropped ticks during market-wide volatility spikes.

This specification defines the architecture, state orchestration, hardware-accelerated rendering pipeline, and real-time streaming integration for the **Flutter BTC/USDT Live TradingView Chart & Market Watch Screen** (`lib/screens/trading/btc_chart_screen.dart`). The screen delivers an institutional-grade trading terminal experience on mobile (iOS, Android) and desktop (macOS, Windows, Linux) platforms. It features sub-second price updates, interactive timeframe selection (1m, 5m, 15m, 1h, 1D), client-side technical indicators, a synchronized Level-2 order book depth visualizer, and a hybrid dual-rendering architecture that combines the TradingView Lightweight Charts bridge with an Impeller-accelerated native Flutter CustomPainter fallback.

## What You Are Building
A high-performance trading screen and modular widget ecosystem centered in `apps/growww_flutter/lib/screens/trading/btc_chart_screen.dart` (supported by feature modules in `apps/growww_flutter/lib/features/trading/`):
- `BtcChartScreen`: The root trading screen providing a multi-pane layout that coordinates the live market watch header, timeframe resolution bar, dual-engine charting viewport, Level-2 depth visualizer, and trading action dock.
- `MarketWatchHeaderWidget`: Live ticker banner displaying the BTC/USDT last traded price (LTP), 24-hour absolute change, percentage change with directional uptick (green) and downtick (red) pulse animations, 24-hour high, 24-hour low, 24-hour volume in BTC and USDT, and WebSocket connection status.
- `TradingChartViewport`: Adaptive hybrid container that hosts the TradingView Lightweight Charts bridge for desktop and deep technical analysis, with transparent automatic fallback to the native custom canvas painter on mobile devices or during offline playback.
- `TradingViewBridgeWidget`: Sandboxed platform view and WebView bridge embedding TradingView Lightweight Charts (v4.x/v5.x), executing zero-copy tick streaming via typed JavaScript channels, synchronized color theming, and multi-timeframe bar replacement.
- `CustomCanvasCandleChart`: Hardware-accelerated Flutter `CustomPainter` and `RenderBox` engine delivering sustained 60/120 FPS rendering of OHLC candlesticks, wicks, dynamic price grids, and time axes using Skia and Impeller backends.
- `TimeframeSelectorBar`: Responsive segmented timeframe control enabling instantaneous switching across `1m`, `5m`, `15m`, `1h`, and `1D` resolutions, coordinating local cache lookups, gap detection, and historical backfill requests.
- `OrderBookDepthVisualizer`: Synchronized Level-2 market depth ladder visualizing cumulative bid and ask volume curves, wall orders, instantaneous spread, and order count distribution alongside the price axis.
- `TechnicalIndicatorsToolbar`: Interactive drawer and floating toolbar allowing traders to toggle and configure Moving Average Convergence Divergence (MACD 12/26/9), Relative Strength Index (RSI 14), Bollinger Bands (20, 2-sigma), Exponential Moving Averages (EMA 9/21/50/200), and Volume bars.
- `CrosshairAndTooltipOverlay`: Multi-touch gestural inspector displaying timestamp, open, high, low, close, volume, and percentage distance from candle open, with magnetic snapping to the nearest candlestick body.
- `Riverpod State Layer`: Asynchronous notifiers (`BtcTickerStreamNotifier`, `BtcDepthStreamNotifier`, `BtcChartNotifier`, `TimeframeNotifier`, `IndicatorSelectionNotifier`) providing reactive, unidirectional data flow with frame-rate throttled updates.

## Scope Boundaries
- **In Scope:**
  - `BtcChartScreen` scaffold and responsive layouts for mobile handsets, tablets, and desktop widescreen displays.
  - Sub-second WebSocket streaming tick ingestion and dynamic forming-candle aggregation.
  - Dual charting engines: TradingView Lightweight Charts bridge and native Flutter `CustomPainter` fallback.
  - Interactive timeframe switcher supporting `1m`, `5m`, `15m`, `1h`, and `1D` resolutions with historical bar pagination.
  - Real-time Level-2 order book depth visualizer (cumulative bid/ask curves and price ladders).
  - Client-side technical indicators: EMA (9, 21, 50, 200), RSI (14), MACD (12/26/9), Bollinger Bands, and Volume bars.
  - Tabular numeric typography (`FontFeature.tabularFigures()`) to prevent visual layout shifting.
  - Monotonic sequence number validation to detect dropped packets or out-of-order tick deliveries.
  - Offline network detection, exponential backoff reconnection, and historical catch-up gap reconciliation.
- **Out of Scope / Handled Elsewhere:**
  - Upstream crypto exchange market data feeder and WebSocket ingress daemon (handled in Prompt 272).
  - High-throughput WebSocket market depth broadcast microservice (handled in Prompt 276).
  - Historical tick aggregation and TimescaleDB hypertable maintenance (handled in Prompt 408).
  - Order entry ticket, limit/market order placement modal, and bracket orders (handled in Prompt 509).
  - Multi-chain crypto deposit and custodial wallet balance tracking (handled in Prompt 528).
  - Pre-trade risk checks, margin collateral calculation, and liquidation engine (handled in Prompt 206, Prompt 241).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+.
- **State Management:** `flutter_riverpod` (v2.5+) utilizing `AsyncNotifier` and `StreamProvider` for reactive, lifecycle-aware state delivery.
- **Rendering & Charting Engine:**
  - *TradingView Bridge:* TradingView Lightweight Charts (v4.2+) embedded via `webview_flutter: ^4.7.0` (mobile hybrid composition) and `package:web` platform view factories (desktop and Web).
  - *Native Canvas Engine:* Flutter `CustomPainter` and `RenderBox` hardware-accelerated by Impeller (Metal on iOS/macOS, Vulkan on Android) and Skia (Windows/Linux) with strict `RepaintBoundary` isolation.
- **Real-Time Data Streaming:** `web_socket_channel: ^3.0.0` for sub-second binary and JSON market data streaming.
- **HTTP & REST Client:** `dio: ^5.4.3+1` for paginated historical candlestick retrieval and gap catch-up requests.
- **Fixed-Point Financial Mathematics:** `decimal: ^2.3.3` for zero-drift price, volume, and percentage calculations.
- **Vector Mathematics & Viewport Projections:** `vector_math: ^2.1.4` for affine coordinate transformations, pinch-to-zoom scaling, and hit-testing.
- **Typography & Font Shaping:** `dart:ui` `FontFeature.tabularFigures()` to ensure equal glyph widths for numerals, preventing visual jitter during high-speed price updates.

## Backend / Infra Touchpoints
- **Market Data Feeder (Prompt 272):**
  - **Live Ticker WebSocket (`wss://ws.growww.in/v1/market/crypto/stream`):** Delivers continuous sub-second BTC/USDT price ticks, trade size, 24-hour volume, 24-hour high/low, and monotonically increasing sequence IDs.
  - **Historical Candlestick REST API (`GET /api/v1/market/crypto/candles`):** Returns historical OHLCV bars parameterized by `symbol=BTCUSDT`, `interval` (1m, 5m, 15m, 1h, 1d), `startTime`, `endTime`, and `limit` (up to 1,000 bars per page).
  - **Candlestick Catch-Up REST API (`GET /api/v1/market/crypto/candles/catchup`):** Fetches missing micro-bars between the client's last cached timestamp and the active WebSocket subscription head after a connection drop.
- **Market Depth Broadcaster (Prompt 276):**
  - **Level-2 Depth WebSocket (`wss://ws.growww.in/v1/market/crypto/depth`):** Streams consolidated Level-2 order book depth snapshots (top 20 bid and ask price levels with quantities and cumulative totals) updated at 100ms intervals.
- **TimescaleDB & Redis Infrastructure (Prompt 402, Prompt 408):**
  - Redis Pub/Sub channels (`market:btc_usdt:ticker`, `market:btc_usdt:depth_l2`) backing the external WebSocket gateways.
  - TimescaleDB hypertable `crypto_candles_1m` continuous aggregates powering multi-timeframe rollups.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Zero Direct Blockchain Interaction on UI Thread:**
  - The Flutter UI thread maintains strictly zero direct RPC, Web3, or smart contract interaction during live chart rendering.
  - Streaming market data, price ticks, and depth ladders originate exclusively from the backend Market Data Feeder (Prompt 272) and Market Depth Broadcaster (Prompt 276).
  - All crypto collateral custody attestations, proof-of-reserve balances, and on-chain settlement operations happen asynchronously through background backend microservices (Prompt 234, Prompt 308, Prompt 528).
  - This strict separation guarantees uninterrupted 60/120 FPS rendering, zero UI thread freezing, and complete immunity from blockchain network latency, gas fluctuations, or block reorg delays.

## State Management & Real-Time Data Pipeline Architecture

### 1. Dual-Engine Architecture & Viewport Orchestration
The trading screen employs an adaptive dual-engine topology:
- **Desktop, Web, and Deep Technical Analysis (TradingView Bridge Engine):**
  - Sandboxed platform view running TradingView Lightweight Charts.
  - Supports multi-pane indicator charting, complex drawing tools, and standard institutional keyboard shortcuts.
  - Streams ticks into the JavaScript runtime over a typed bidirectional communication channel.
- **Mobile Handsets and Battery-Efficient Native View (CustomPainter Engine):**
  - Native Flutter Impeller/Skia canvas painter with sub-pixel antialiasing.
  - Optimized for 120 FPS touch scrubbing, low memory footprint, and high-frequency tick updates.
  - Direct touch gesture dispatch for independent time (X-axis) and price (Y-axis) pinch-to-zoom.

```
+-----------------------------------------------------------------------------------+
|                           BTC/USDT LIVE CHART SCREEN                              |
+-----------------------------------------------------------------------------------+
|  [ MARKET WATCH HEADER ]  - LTP, 24h Change, High/Low, Vol, Stream Health Badge   |
|  [ TIMEFRAME SELECTOR  ]  - 1m  |  5m  |  15m  |  1h  |  1D                       |
+-----------------------------------------+-----------------------------------------+
|     TRADINGVIEW LIGHTWEIGHT BRIDGE      |         NATIVE CUSTOM CANVAS            |
|   - Sandboxed Platform View / WebView   |   - Impeller / Skia Hardware Canvas     |
|   - Multi-Pane Technical Indicators     |   - Sub-10ms Live Candlestick Painter   |
|   - Vectorized Drawing Tools            |   - 120 FPS VSYNC Frame Buffer          |
+-----------------------------------------+-----------------------------------------+
|  [ ORDER BOOK DEPTH VISUALIZER ]        - L2 Cumulative Bid/Ask Liquidity Profile |
|  [ ACTION DOCK ]                        - Quick Buy / Quick Sell Action Buttons   |
+-----------------------------------------------------------------------------------+
                                        |
                   [ RIVERPOD REACTIVE ORCHESTRATOR ]
                                        |
     +----------------------------------+----------------------------------+
     |                                  |                                  |
     v                                  v                                  v
[ TICK INGESTION ISOLATE ]     [ L2 DEPTH BUFFER ]            [ CACHE & RECONCILIATION ]
  - WebSocket Stream (272)       - Depth Stream (276)           - Local Bar Cache
  - Monotonic Sequence Guard     - 100ms Ladder Snapshot        - Catch-Up Gap Backfill
  - VSYNC Aggregation Buffer     - Cumulative Volume Calc       - Historical Restorer
```

### 2. High-Frequency Tick Aggregation & VSYNC Throttling
1. Raw WebSocket ticks from Prompt 272 arrive at sub-second intervals (up to 500 ticks per second during high volatility).
2. Ticks pass through a background memory ring buffer that validates monotonically increasing sequence IDs.
3. The active forming candlestick updates its high, low, close, volume, and trade count in place.
4. A VSYNC frame callback (`SchedulerBinding.instance.scheduleFrameCallback`) flushes the updated forming candle to the active chart painter at the display refresh rate (16.6ms for 60Hz, 8.3ms for 120Hz ProMotion).
5. Off-screen historical candlesticks remain immutable in memory, preventing unnecessary re-allocations and garbage collection pauses.

### 3. Order Book Depth Visualizer Pipeline
1. Level-2 order book snapshots arrive from Prompt 276 every 100ms.
2. The depth engine calculates cumulative bid volume ($V_{bid}(p) = \sum_{i \ge p} q_i$) and cumulative ask volume ($V_{ask}(p) = \sum_{i \le p} q_i$).
3. Maximum cumulative volume determines the horizontal scale factor.
4. Translucent bid (green) and ask (red) histograms are painted adjacent to the price scale, providing instantaneous visual depth without obscuring active candlestick wicks.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure:** Create directories under `apps/growww_flutter/lib/screens/trading/` and `apps/growww_flutter/lib/features/trading/`: `presentation/screens/`, `presentation/widgets/`, `presentation/painters/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `data/datasources/`, `data/repositories/`, and `assets/tradingview/`.
2. **Define Domain Models:** Implement immutable Dart models in `domain/models/`: `CandlestickData`, `TickerUpdate`, `DepthSnapshot`, `OrderBookLevel`, `TimeframeInterval`, `ChartEngineType`, and `TechnicalIndicatorConfig`.
3. **Implement Tabular Numeric Typography Styles:** Configure app text themes using `TextStyle` with `fontFeatures: [FontFeature.tabularFigures()]` to ensure mono-spaced numeric digits across price, percentage, and volume readouts.
4. **Implement Market Data WebSocket Service:** Build `BtcMarketDataWebSocketService` in `data/datasources/` to establish and maintain WebSocket streams to Market Data Feeder (`/v1/market/crypto/stream`) and Market Depth Broadcaster (`/v1/market/crypto/depth`) with ping/pong heartbeats and monotonic sequence validation.
5. **Implement Forming Candle Aggregator:** Build `FormingCandleAggregator` in `domain/algorithms/` to ingest live ticks, update the current active candle (open, high, low, close, volume), and trigger VSYNC-aligned UI frame callbacks.
6. **Implement Riverpod State Notifiers:** Create `BtcTickerNotifier`, `BtcDepthNotifier`, `BtcChartConfigNotifier`, and `BtcHistoricalCandlesNotifier` to manage reactive data flow, connection state transitions, and user preferences.
7. **Build TradingView Lightweight Charts Bridge:**
   - Scaffold `assets/tradingview/btc_chart_bridge.html` bundling TradingView Lightweight Charts library.
   - Build `TradingViewBridgeWidget` wrapping an `InAppWebView` or platform view, exposing a typed bidirectional `JavaScriptChannel` (`GrowwwBtcBridge`).
   - Implement JavaScript handlers: `setCandles()`, `updateBar()`, `setTheme()`, `setResolution()`, and `toggleIndicator()`.
8. **Build Native Custom Canvas Painter:** Implement `CustomCanvasCandleChart` extending `CustomPainter`:
   - Paint bullish (green) and bearish (red) candle bodies and wicks with sub-pixel alignment.
   - Paint volume histogram bars at the bottom with matching directional colors.
   - Paint dynamic horizontal price grid lines and vertical time grid lines with formatted labels.
   - Wrap the painter in a `RepaintBoundary` to prevent invalidating the parent layout.
9. **Build Market Watch Header Widget:** Create `MarketWatchHeaderWidget` displaying BTC/USDT ticker symbol, last traded price in tabular numerals, 24-hour price change with directional flash animation (400ms green/red fade), 24-hour high, 24-hour low, 24-hour volume, and a live stream health indicator badge.
10. **Build Timeframe Selector Bar:** Construct `TimeframeSelectorBar` presenting segmented chips for `1m`, `5m`, `15m`, `1h`, and `1D`. Tapping a chip queries cached historical bars, triggers gap catch-up if needed, and switches the WebSocket aggregation resolution.
11. **Build Level-2 Order Book Depth Visualizer:** Construct `OrderBookDepthVisualizer` to render horizontal cumulative bid and ask volume histograms alongside the price scale, highlighting liquidity walls and the active bid-ask spread.
12. **Build Technical Indicators Toolbar:** Implement `TechnicalIndicatorsToolbar` enabling users to toggle overlays (EMA 9/21/50/200, Bollinger Bands) and sub-pane oscillators (RSI, MACD) with custom parameter sliders.
13. **Implement Gestural Crosshair and Tooltip Overlay:** Build `CrosshairAndTooltipOverlay` with gesture handlers for long-press scrubbing, displaying a floating badge with timestamp, OHLCV values, and percentage change relative to the previous close.
14. **Build Root BTC Chart Screen Scaffold:** Construct `BtcChartScreen` in `lib/screens/trading/btc_chart_screen.dart`, unifying the market watch header, timeframe selector, chart container, depth visualizer, and bottom action buttons (Quick Buy, Quick Sell).
15. **Implement Offline Recovery and Gap Reconciliation:** Add network connectivity listeners (`connectivity_plus`) and implement exponential backoff reconnection. Upon reconnecting, invoke `GET /api/v1/market/crypto/candles/catchup` to fill missing candlestick bars before resuming live tick aggregation.
16. **Write Unit, Golden, and Benchmark Tests:** Author test suites in `test/features/trading/` verifying candlestick aggregation arithmetic, tabular typography rendering, sequence gap detection, and sustained 60/120 FPS rendering under synthetic load.

## Interfaces / Contracts

```dart
// lib/features/trading/domain/models/timeframe_interval.dart

enum TimeframeInterval {
  oneMinute('1m', 60),
  fiveMinutes('5m', 300),
  fifteenMinutes('15m', 900),
  oneHour('1h', 3600),
  oneDay('1D', 86400);

  final String label;
  final int durationSeconds;

  const TimeframeInterval(this.label, this.durationSeconds);
}

// lib/features/trading/domain/models/candlestick_data.dart

class CandlestickData {
  final DateTime timestamp;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;
  final double quoteVolume;
  final int tradeCount;

  const CandlestickData({
    required this.timestamp,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
    required this.quoteVolume,
    required this.tradeCount,
  });

  bool get isBullish => close >= open;
  double get bodyRange => (close - open).abs();
  double get totalRange => high - low;

  CandlestickData copyWith({
    DateTime? timestamp,
    double? open,
    double? high,
    double? low,
    double? close,
    double? volume,
    double? quoteVolume,
    int? tradeCount,
  }) {
    return CandlestickData(
      timestamp: timestamp ?? this.timestamp,
      open: open ?? this.open,
      high: high ?? this.high,
      low: low ?? this.low,
      close: close ?? this.close,
      volume: volume ?? this.volume,
      quoteVolume: quoteVolume ?? this.quoteVolume,
      tradeCount: tradeCount ?? this.tradeCount,
    );
  }

  Map<String, dynamic> toJson() => {
    'timestamp': timestamp.millisecondsSinceEpoch ~/ 1000,
    'open': open,
    'high': high,
    'low': low,
    'close': close,
    'volume': volume,
    'quoteVolume': quoteVolume,
    'tradeCount': tradeCount,
  };
}

// lib/features/trading/domain/models/ticker_update.dart

class TickerUpdate {
  final String symbol;
  final double lastPrice;
  final double priceChange24h;
  final double priceChangePercent24h;
  final double high24h;
  final double low24h;
  final double volumeBtc24h;
  final double volumeUsdt24h;
  final int sequenceId;
  final DateTime timestamp;

  const TickerUpdate({
    required this.symbol,
    required this.lastPrice,
    required this.priceChange24h,
    required this.priceChangePercent24h,
    required this.high24h,
    required this.low24h,
    required this.volumeBtc24h,
    required this.volumeUsdt24h,
    required this.sequenceId,
    required this.timestamp,
  });

  factory TickerUpdate.fromJson(Map<String, dynamic> json) {
    return TickerUpdate(
      symbol: json['symbol'] as String,
      lastPrice: (json['lastPrice'] as num).toDouble(),
      priceChange24h: (json['priceChange24h'] as num).toDouble(),
      priceChangePercent24h: (json['priceChangePercent24h'] as num).toDouble(),
      high24h: (json['high24h'] as num).toDouble(),
      low24h: (json['low24h'] as num).toDouble(),
      volumeBtc24h: (json['volumeBtc24h'] as num).toDouble(),
      volumeUsdt24h: (json['volumeUsdt24h'] as num).toDouble(),
      sequenceId: json['sequenceId'] as int,
      timestamp: DateTime.fromMillisecondsSinceEpoch(json['timestamp'] as int),
    );
  }
}

// lib/features/trading/domain/models/depth_snapshot.dart

class OrderBookLevel {
  final double price;
  final double quantity;
  final double cumulativeTotal;
  final int orderCount;

  const OrderBookLevel({
    required this.price,
    required this.quantity,
    required this.cumulativeTotal,
    required this.orderCount,
  });

  factory OrderBookLevel.fromJson(List<dynamic> json, double runningTotal) {
    final price = (json[0] as num).toDouble();
    final quantity = (json[1] as num).toDouble();
    final orderCount = json.length > 2 ? (json[2] as num).toInt() : 1;
    return OrderBookLevel(
      price: price,
      quantity: quantity,
      cumulativeTotal: runningTotal + quantity,
      orderCount: orderCount,
    );
  }
}

class DepthSnapshot {
  final String symbol;
  final int sequenceId;
  final DateTime timestamp;
  final List<OrderBookLevel> bids;
  final List<OrderBookLevel> asks;

  const DepthSnapshot({
    required this.symbol,
    required this.sequenceId,
    required this.timestamp,
    required this.bids,
    required this.asks,
  });

  double get bestBid => bids.isNotEmpty ? bids.first.price : 0.0;
  double get bestAsk => asks.isNotEmpty ? asks.first.price : 0.0;
  double get spread => bestAsk > 0 && bestBid > 0 ? bestAsk - bestBid : 0.0;
  double get spreadPercent => bestBid > 0 ? (spread / bestBid) * 100 : 0.0;

  factory DepthSnapshot.fromJson(Map<String, dynamic> json) {
    double runningBid = 0.0;
    final bidsList = <OrderBookLevel>[];
    for (final item in json['bids'] as List<dynamic>) {
      final level = OrderBookLevel.fromJson(item as List<dynamic>, runningBid);
      runningBid = level.cumulativeTotal;
      bidsList.add(level);
    }

    double runningAsk = 0.0;
    final asksList = <OrderBookLevel>[];
    for (final item in json['asks'] as List<dynamic>) {
      final level = OrderBookLevel.fromJson(item as List<dynamic>, runningAsk);
      runningAsk = level.cumulativeTotal;
      asksList.add(level);
    }

    return DepthSnapshot(
      symbol: json['symbol'] as String,
      sequenceId: json['sequenceId'] as int,
      timestamp: DateTime.fromMillisecondsSinceEpoch(json['timestamp'] as int),
      bids: bidsList,
      asks: asksList,
    );
  }
}

// lib/features/trading/domain/interfaces/i_btc_market_data_repository.dart

abstract class IBtcMarketDataRepository {
  Stream<TickerUpdate> streamBtcTicker();
  Stream<DepthSnapshot> streamBtcDepth();
  Future<List<CandlestickData>> fetchHistoricalCandles({
    required TimeframeInterval interval,
    required DateTime startTime,
    required DateTime endTime,
    int limit = 500,
  });
  Future<List<CandlestickData>> catchUpCandles({
    required TimeframeInterval interval,
    required DateTime lastKnownTimestamp,
  });
}

// lib/features/trading/domain/interfaces/i_trading_view_bridge_controller.dart

abstract class ITradingViewBridgeController {
  void initializeBridge();
  void setCandles(List<CandlestickData> candles);
  void updateFormingCandle(CandlestickData formingCandle);
  void setTimeframe(TimeframeInterval interval);
  void setTheme(String themeMode);
  void toggleIndicator(String indicatorName, bool isEnabled, Map<String, dynamic> params);
  void dispose();
}
```

## Security & Compliance Notes
- **Tabular Numeric Typography & Layout Stability:** Rapidly changing price and depth numerals must render using monospace digits (`FontFeature.tabularFigures()`). Proportional numerals cause character widths to expand and contract on digit changes (such as transitioning from 1 to 8), producing severe visual horizontal jitter, distracting traders, and causing accidental taps on adjacent UI controls during rapid price moves.
- **Monotonic Sequence Integrity & Replay Defense:** All WebSocket packets received from Prompt 272 and Prompt 276 include a monotonically increasing `sequenceId`. The client-side stream parser validates that $sequenceId_{n} = sequenceId_{n-1} + 1$. Any packet arriving out of sequence, with a duplicate ID, or with an older timestamp is flagged. If more than 3 consecutive sequence IDs are dropped, the client initiates an automatic socket reset and triggers the catch-up reconciliation API (`GET /api/v1/market/crypto/candles/catchup`) to ensure zero visual candle distortions.
- **Graceful Offline Reconnection & Gap Reconciliation:** The application monitors device network reachability using `connectivity_plus`. When connectivity drops, the UI displays a non-intrusive warning badge (*"Reconnecting to live feed..."*) without wiping historical chart data. Reconnection attempts use an exponential backoff schedule (1s, 2s, 4s, 8s, up to 30s max) with random jitter to prevent thundering-herd effects on the WebSocket gateway. Once reconnected, missing bars between the disconnection time and the live head are fetched and merged into the active dataset.
- **Sandboxed WebView & Content Security Policy (CSP):** The embedded TradingView bridge executes inside an isolated platform view with restrictive permissions. The local HTML wrapper enforces a strict Content Security Policy (`default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'`). Camera, microphone, geolocation, and external network navigation are explicitly disabled on the WebView controller. All data passes exclusively through the typed `GrowwwBtcBridge` JavaScript channel, rejecting arbitrary code evaluation.
- **Battery & Thermal Throttling Protection:** Sub-second streaming data can rapidly deplete battery life and trigger thermal throttling on mobile devices. An `AppLifecycleListener` detects when the application is minimized or when the user navigates away from `BtcChartScreen`, immediately unsubscribing from WebSocket channels and canceling rendering tickers. Streaming resumes automatically when the screen returns to the foreground.

## Acceptance Criteria
- [ ] `BtcChartScreen` renders the complete trading layout (market watch header, timeframe selector, chart viewport, depth visualizer, action dock) on Android, iOS, macOS, Windows, Linux, and Web.
- [ ] Market watch header displays real-time BTC/USDT price updates with sub-100ms latency upon WebSocket message arrival.
- [ ] Price updates trigger a subtle green (uptick) or red (downtick) flash animation that decays smoothly over 400ms without frame stutter.
- [ ] Numeric values for price, percentage, volume, and depth levels render strictly with tabular figures (`FontFeature.tabularFigures()`), with zero horizontal layout jitter during price fluctuations.
- [ ] Timeframe selector seamlessly transitions between `1m`, `5m`, `15m`, `1h`, and `1D` intervals, reloading historical data and updating the forming candle aggregation period in <300ms.
- [ ] TradingView Lightweight Charts bridge successfully loads, synchronizes dark/light theme, and receives zero-copy streaming tick updates over the JavaScript channel.
- [ ] Native Flutter `CustomPainter` fallback renders candlestick bodies, wicks, volume bars, dynamic price grids, and crosshair overlays at a sustained 60 FPS on standard devices and 120 FPS on high-refresh ProMotion displays.
- [ ] Level-2 order book depth visualizer correctly parses top 20 bids and asks, computes cumulative volume totals, and renders an accurate liquidity profile adjacent to the price scale.
- [ ] Technical indicators toolbar successfully toggles and updates EMA (9, 21, 50, 200), RSI (14), MACD (12/26/9), Bollinger Bands, and Volume bars.
- [ ] Crosshair gesture handling provides magnetic snapping to candle close prices, displaying accurate OHLCV values and percentage differences.
- [ ] Simulating network disconnections triggers the offline reconnection banner, applies exponential backoff, and fetches missing bars via the catch-up API without creating data gaps.
- [ ] Minimizing the application or navigating away suspends WebSocket subscriptions and stops animation controllers, maintaining battery efficiency.
- [ ] Specification adheres strictly to the 12-section template, contains zero raw implementation code in the workspace, and uses standard ASCII hyphens exclusively.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Project Scaffolding), Prompt 502 (App Architecture & State Management), Prompt 503 (Design System & Theming), Prompt 525 (API Client Layer), Prompt 534 (TradingView Charting & Technical Indicators Engine).
- **Backend Dependencies:** Prompt 272 (Market Data Feeder), Prompt 276 (Market Depth Broadcaster), Prompt 408 (Historical Market Data TimescaleDB Pipeline).
- **Parallel Tasks:** Prompt 508 (Security Detail Screen), Prompt 535 (Market Heatmap & Sector Treemap Screen).
- **Downstream Blockers:** Prompt 509 (Flutter Order Placement Flow), Prompt 532 (Desktop Multi-Window Institutional Terminal), Prompt 902 (Cross-Platform End-to-End Test Suite).
