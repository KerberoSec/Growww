# 534 - Flutter High-Performance Custom Canvas & TradingView Charting Engine

## Purpose
High-frequency financial trading, algorithmic order execution, and technical asset analysis require instantaneous, jitter-free visual feedback across multi-asset classes (tokenized equities, derivatives, government securities, and tokenized commodities). Standard cross-platform charting libraries typically suffer from frame drops, high memory overhead, and unoptimized garbage collection cycles when bombarded with continuous market data tick streams. In institutional trading environments, missing a candle formation, experiencing touch scrubbing latency, or miscalculating a dynamic volatility band can lead to erroneous execution and investor capital impairment.

This specification defines the architectural design, mathematical calculation pipeline, and hardware-accelerated rendering engine for the **Growww Flutter High-Performance Custom Canvas & TradingView Charting Suite** (`apps/growww_flutter/lib/features/charting/`). The engine delivers a sustained 60/120 FPS rendering pipeline on mobile and desktop devices using a dual-engine architecture:
1. A bespoke, Impeller/Skia-accelerated Flutter `CustomPainter` canvas engine optimized for extreme-low-latency tick scrubbing, Level-2 order book depth overlays, volume profiles, and battery-efficient mobile interaction.
2. A tightly coupled TradingView Lightweight Charts bridge utilizing hybrid composition platform views for deep institutional technical analysis, advanced multi-pane indicator charting, and multi-point vector drawing tools.

The charting suite integrates directly with the TimescaleDB historical market data pipeline (Prompt 408), streams sub-10ms binary ticks and market depth via WebSockets (Prompt 207), and plots cryptographic, on-chain corporate action ex-dates and dividend distribution markers verified against Hyperledger Besu smart contracts.

## What You Are Building
A modular, high-performance charting subsystem situated in `apps/growww_flutter/lib/features/charting/`:
- `TradingChartContainerWidget`: Adaptive hybrid container supporting seamless switching or synchronization between the native Flutter `CustomPainter` canvas engine (for ultra-high-speed mobile tick scrubbing, L2 depth overlay, and 120 FPS VSYNC) and the embedded TradingView Lightweight Charts bridge (for heavy multi-pane indicator suites and complex desktop drawing tools).
- `CustomCanvasCandleChart`: Hardware-accelerated Flutter `CustomPainter` and `RenderBox` engine rendering multi-timeframe Candlestick (OHLC), Hollow Candlestick, Heikin-Ashi, and Area/Line charts with sub-pixel alignment, candle body antialiasing, and dynamic price-time grid interpolation.
- `TradingViewLightweightBridge`: Bidirectional JavaScript communication bridge embedding TradingView Lightweight Charts v4.x/v5.x within Flutter WebViews (mobile) and Web platform views (desktop/web), providing zero-copy tick ingestion, theme synchronization, and chart serialization.
- `Level2MarketDepthOverlay`: Real-time cumulative order book depth profile (bids vs asks) rendered directly on the right-hand Y-axis price scale, visually depicting liquidity clusters, wall orders, and bid-ask spread boundaries.
- `VolumeProfileEngine`: Dynamic visible-range volume profile (VRVP) calculating and painting the Point of Control (POC), Value Area High (VAH), and Value Area Low (VAL) across the active viewport.
- `TechnicalIndicatorsSuite`: High-performance vectorized client-side indicator calculation engine supporting Moving Average Convergence Divergence (MACD 12/26/9), Relative Strength Index (RSI 14 with overbought/oversold bands), Bollinger Bands (20, 2-sigma), Exponential Moving Averages (EMA 9/21/50/200), and Volume Weighted Average Price (VWAP).
- `InteractiveDrawingToolsToolbar`: Vector-based overlay layer supporting trendlines, horizontal support/resistance rays, Fibonacci retracements, pitchforks, and text annotations, maintaining normalized $(t, p)$ coordinates that remain mathematically pinned across viewport transformations.
- `CorporateActionOnChainMarkers`: Interactive on-chart timeline markers identifying corporate action ex-dates (dividends, splits, bonus issues) that allow traders to inspect cryptographic execution hashes and depository verification records on Hyperledger Besu.
- `CrosshairAndTooltipEngine`: Multi-touch gestural inspector displaying timestamp, OHLCV data, active indicator values, and percentage distance from cursor to previous close with magnetic candle snapping.

## Scope Boundaries
- **In Scope:**
  - Hardware-accelerated Flutter `CustomPainter` rendering of OHLCV candles, wicks, volume bars, and price scales at 60/120 FPS.
  - High-performance TradingView Lightweight Charts bridge with bidirectional typed JavaScript messaging.
  - Level-2 market depth histogram overlay rendered on the chart price axis.
  - Visible-range Volume Profile (VRVP) with Point of Control (POC) and Value Area calculations.
  - Client-side vectorized technical indicators: MACD, RSI, Bollinger Bands, EMA, and VWAP.
  - Vector drawing tools overlay with hit-testing, drag handles, and persistent normalized coordinate storage.
  - On-chart corporate action markers linked to Hyperledger Besu smart contract event logs.
  - Viewport navigation: pinch-to-zoom (independent X and Y axes), inertial pan, dynamic time scaling, and double-tap zoom reset.
  - Multi-timeframe switching (1s, 1m, 5m, 15m, 1h, 1D, 1W) with local caching and gap-filling reconciliation.
  - SEBI-compliant financial color palettes, high-contrast monochrome mode, and color-blind accessible modes.
- **Out of Scope / Handled Elsewhere:**
  - Server-side historical tick aggregation and TimescaleDB hypertable maintenance (handled in Prompt 408).
  - Exchange gateway order routing and WebSocket tick multiplexing infrastructure (handled in Prompt 207).
  - Order entry ticket, bracket orders, and trading modal state (handled in Prompt 509).
  - Smart contract dividend distribution and corporate actions clearing logic (handled in Prompt 222, Prompt 306).
  - Physical depository custody settlement and reconciliation (handled in Prompt 213, Prompt 308).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ utilizing `flutter_riverpod` (v2.5+) for reactive state management.
- **Rendering Engines:**
  - *Native Canvas:* Hardware-accelerated Impeller (Vulkan on Android, Metal on iOS/macOS) and Skia (Windows/Linux) via custom `RenderBox` and `CustomPainter` with strict `RepaintBoundary` isolation.
  - *TradingView Bridge:* TradingView Lightweight Charts (v4.2+) embedded via `webview_flutter: ^4.7.0` (mobile hybrid composition) and `dart:html` / `package:web` platform view factories (desktop and Web).
- **Vector Mathematics & Transformations:** `vector_math: ^2.1.4` for coordinate transformations, affine matrix projection, hit-testing algorithms, and Fibonacci ratios.
- **Data Streaming & Transport:** `web_socket_channel: ^3.0.0` for sub-10ms binary tick feeds and `dio: ^5.4.3+1` for REST historical bar pagination.
- **Local Persistence & Caching:** `isar: ^3.1.0` or `sqlite3` for high-throughput local caching of multi-timeframe historical candlestick bars.
- **Typography & Layout:** `dart:ui` `ParagraphBuilder` and `TextPainter` for zero-allocation axis label layout and crosshair callout painting.

## Backend / Infra Touchpoints
- **Historical Market Data Pipeline (Prompt 408):**
  - **Historical Candles API (`GET /api/v1/market/history/candles`):** Queries TimescaleDB continuous aggregates with parameters `symbol`, `resolution` (1s, 1m, 5m, 15m, 1h, 1D, 1W), `from` (UNIX timestamp), and `to` (UNIX timestamp). Supports chunked pagination and server-side gap detection.
  - **Tick-to-Candle Catch-Up API (`GET /api/v1/market/history/catchup`):** Fetches missing micro-bars between the client's last cached timestamp and the current WebSocket subscription sequence.
- **Market Data Service WebSocket Gateway (Prompt 207):**
  - **Live Tick Channel (`wss://ws.growww.in/v1/market/stream`):** Ingests real-time binary or typed JSON tick payloads containing `symbol`, `lastPrice`, `lastQuantity`, `volume`, `turnover`, and sequence number.
  - **Level-2 Depth Channel (`wss://ws.growww.in/v1/market/depth`):** Ingests 5-level and 20-level consolidated order book bid/ask ladders with quantities and order counts for depth overlay rendering.
- **Corporate Actions Service (Prompt 222):**
  - **Corporate Action Events API (`GET /api/v1/corporate-actions/events?symbol=RELIANCE`):** Retrieves announced, approved, and executed corporate actions (cash dividends, stock splits, bonus issues, rights offerings) with ex-dates and smart contract transaction hashes.
- **Proof-of-Reserve Registry (Prompt 308):**
  - **Asset Attestation API (`GET /api/v1/custody/attestation?symbol=RELIANCE`):** Supplies custodian depository confirmation hashes verifying 1:1 backing of tokenized shares.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chart Corporate Action Markers:**
  - The chart canvas projects interactive badge flags directly on the horizontal time axis at the precise timestamps corresponding to on-chain corporate action ex-dates.
  - Tapping a corporate action badge summons a slide-up inspection sheet displaying:
    - Event Type: Cash Dividend, Stock Split, Bonus Issue, or Merged Spin-Off.
    - Ratio or Amount: e.g., "Dividend ₹10.00 / share" or "Split 1:5".
    - Smart Contract Address: `CorporateActionManager.sol` deployed on the permissioned Hyperledger Besu network.
    - Transaction Hash & Block Height: Direct link to the internal Besu block explorer verifying cryptographic execution.
    - Depository Reference: Cross-referenced NSDL/CDSL corporate action filing ID confirming 1:1 underlying custodian compliance.
- **Tamper-Evident Dividend Distribution Validation:**
  - Corporate action markers include a cryptographic SHA-256 verification hash matching the on-chain event log. If an attacker attempts to spoof or manipulate dividend yield data client-side, the hash verification fails, and the marker displays an untrusted data warning.
- **Zero On-Chain PII Guarantee:**
  - Only public security-level metadata (tickers, ISINs, ex-dates, ratios, and contract addresses) are plotted and verified. No individual user account numbers, demat IDs, or transaction balances are written to or retrieved from the blockchain.

## State Management Architecture & Dual Charting Engine Pipeline

### 1. Dual-Engine Architecture & Routing Strategy
The charting system utilizes an adaptive dual-engine topology:
- **Mobile Handsets & Quick Inspection (Native CustomPainter Engine):**
  - Uses Flutter Impeller hardware-accelerated canvas.
  - Renders raw candlestick bars, Level-2 depth profiles, and core indicators (EMA, RSI, MACD) at 60/120 FPS with minimal memory allocation and battery drain.
  - Directly handles touch gestures for pan, scale, crosshair, and quick time-range scrubbing.
- **Desktop, Web, & Institutional Deep-Dive (TradingView Bridge Engine):**
  - Embeds TradingView Lightweight Charts within a sandboxed platform view.
  - Activates when users enter full-screen landscape mode or desktop widescreen layouts requiring complex multi-pane drawing suites, horizontal pitchforks, and 50+ concurrent technical indicators.
  - Communicates via an asynchronous JavaScript message channel, streaming candle updates directly into the JS runtime without DOM recreation.

```
+-----------------------------------------------------------------------------------+
|                            TRADING CHART CONTAINER                                |
+-----------------------------------------+-----------------------------------------+
|     NATIVE CUSTOMPAINTER ENGINE         |      TRADINGVIEW LIGHTWEIGHT BRIDGE     |
|   - Impeller / Skia Hardware Canvas     |   - Sandboxed Platform View / WebView   |
|   - 120 FPS Sub-Frame Tick Scrubbing    |   - Bidirectional JS Message Channel    |
|   - Level-2 Depth Price Scale Overlay   |   - Multi-Pane Indicator Layouts        |
|   - Visible-Range Volume Profile        |   - Advanced Drawing Tool Sets          |
|   - Instant Touch Gesture Tracking      |   - Full Historical Replay Engine       |
+-----------------------------------------+-----------------------------------------+
                    |                                          |
                    +--------------------+---------------------+
                                         |
                       [ RIVERPOD CHART ORCHESTRATION ]
                                         |
     +-----------------------------------+-----------------------------------+
     |                                   |                                   |
     v                                   v                                   v
[ HISTORICAL CANDLE CACHE ]    [ REAL-TIME TICK BUFFER ]         [ DRAWING & INDICATOR ]
  - TimescaleDB (Prompt 408)     - 16ms / 8ms VSYNC Flusher        - Normalized (t, p) Store
  - Local Isar Multi-Timeframe   - Current Candle Aggregator       - Vectorized Calculation
```

### 2. High-Frequency Tick Aggregation & VSYNC Flush Pipeline
To prevent UI thread starvation during high-throughput market periods (e.g., 2,000 ticks/second during market open):
1. Incoming WebSocket ticks from Prompt 207 are ingested by a background isolate or lightweight memory ring buffer.
2. The buffer accumulates ticks, updating the high, low, close, volume, and turnover of the active forming candle.
3. A VSYNC frame callback (`SchedulerBinding.instance.scheduleFrameCallback`) flushes the aggregated forming candle to the active chart painter exactly once per display refresh cycle (16.6ms for 60Hz displays, 8.3ms for 120Hz ProMotion displays).
4. Completed historical candles are stored in an immutable contiguous array (`Float64List` memory layout) to maximize CPU L1/L2 cache locality during canvas repainting.

### 3. Vectorized Technical Indicators Engine
Indicators are calculated using incremental vector algorithms to avoid recomputing thousands of bars on every tick:
- **EMA (Exponential Moving Average):** Calculated using $EMA_t = \alpha \cdot Price_t + (1 - \alpha) \cdot EMA_{t-1}$, where $\alpha = 2 / (N + 1)$. Only the current forming bar is updated during live ticks; historical values remain immutable.
- **RSI (Relative Strength Index):** Uses Wilder's Smoothing Technique across 14 periods. Gains and losses are maintained incrementally.
- **MACD (12, 26, 9):** Fast EMA (12) minus Slow EMA (26), paired with a 9-period Signal EMA line and dynamic divergence histogram.
- **Bollinger Bands (20, 2-sigma):** Computes 20-period SMA and rolling population standard deviation $\sigma = \sqrt{\frac{\sum (x_i - \mu)^2}{N}}$, projecting Upper ($\mu + 2\sigma$), Middle ($\mu$), and Lower ($\mu - 2\sigma$) bands.
- **VWAP (Volume Weighted Average Price):** Cumulative $\frac{\sum (TypicalPrice \times Volume)}{\sum Volume}$, reset daily at market open session boundaries.

### 4. Level-2 Depth & Volume Profile Overlay Pipeline
- **Level-2 Depth Overlay:** Bid and ask limit order queues from Prompt 207 are bucketed into discrete price intervals matching the current vertical zoom resolution. Cumulative volumes are drawn as horizontal translucent histograms extending leftward from the right Y-axis scale.
- **Visible-Range Volume Profile (VRVP):** Analyzes all completed candles currently visible within the horizontal viewport bounds $[x_{min}, x_{max}]$. Price range $[Low_{min}, High_{max}]$ is partitioned into $M$ discrete bins (default 50 bins). Volume is allocated per bin. The Point of Control (POC) identifies the bin with maximum volume, highlighted with an amber ray across the chart.

### 5. Normalized Drawing Tools Vector Architecture
User drawings (trendlines, horizontal rays, Fibonacci levels) are stored using normalized coordinates $(t, p)$, where $t$ is the exact UTC epoch timestamp and $p$ is the exact double-precision asset price. During canvas rendering, $(t, p)$ coordinates are projected into screen space $(x, y)$ via the active viewport transformation matrix. This guarantees that drawings remain mathematically pinned to specific candle timestamps and price levels regardless of user pinch, zoom, or pan actions.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Feature Directories:** Create `apps/growww_flutter/lib/features/charting/` with subdirectories `presentation/screens/`, `presentation/widgets/`, `presentation/painters/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `domain/algorithms/`, `data/datasources/`, `data/repositories/`, and `assets/tradingview/`.
2. **Define Domain Data Models:** Author immutable models in `domain/models/`: `CandleData`, `ChartViewport`, `ChartTimeframe`, `Level2DepthLevel`, `VolumeProfileData`, `IndicatorConfig`, `CorporateActionMarker`, and `ChartDrawingEntity`.
3. **Implement Vectorized Technical Indicator Calculators:** Create pure Dart calculation algorithms in `domain/algorithms/`:
   - `EmaCalculator` supporting configurable window lengths (9, 21, 50, 200).
   - `RsiCalculator` implementing Wilder's smoothing algorithm with overbought (70) and oversold (30) levels.
   - `MacdCalculator` computing fast/slow EMAs, signal line, and histogram bars.
   - `BollingerBandsCalculator` computing rolling standard deviation and boundary bands.
   - `VwapCalculator` maintaining session-accumulated price-volume ratios.
4. **Build Visible-Range Volume Profile (VRVP) Engine:** Implement `VolumeProfileCalculator` in `domain/algorithms/` to partition visible candles into discrete price bins, compute total traded volume per bin, identify the Point of Control (POC), and delineate the 70% Value Area (VAH and VAL).
5. **Implement Native CandleChartPainter:** Author `CandleChartPainter` extending `CustomPainter`:
   - Implement low-level `Canvas.drawRect` and `Canvas.drawLine` operations for bullish (green) and bearish (red) candle bodies and wicks.
   - Paint volume bars at the base of the viewport with matching sentiment colors.
   - Implement dynamic grid lines, time axis labels, and right-aligned price axis labels using `dart:ui` `ParagraphBuilder`.
   - Wrap the painter in a `RepaintBoundary` to prevent invalidation of surrounding widgets.
6. **Implement Level-2 Market Depth Overlay Painter:** Build `MarketDepthPainter` to render horizontal cumulative bid (blue) and ask (orange) depth ladders projecting from the price axis over the right side of the canvas.
7. **Implement Interactive Gestural Viewport Controller:** Build `ChartGestureHandler` wrapping an `InteractiveViewer` or custom `Listener` to process:
   - Horizontal pinch-to-zoom (scaling candle bar width between 2dp and 40dp).
   - Vertical pinch-to-zoom (scaling price axis amplitude).
   - Smooth horizontal drag panning with kinetic inertia and clamp boundaries.
   - Long-press touch activation of the crosshair inspector with magnetic snapping to the nearest candle close.
8. **Build Corporate Action Marker Painter:** Implement canvas drawing logic for on-chain corporate action badges (D for Dividend, S for Split, B for Bonus) positioned on the time axis, including touch hit-testing to trigger the corporate action verification sheet.
9. **Implement Vector Drawing Tools Layer:** Build `DrawingCanvasPainter` to render user-created trendlines, rays, horizontal support/resistance levels, and Fibonacci retracements with interactive touch handles for resizing and dragging.
10. **Build Sandboxed TradingView Lightweight Charts Bridge:**
    - Scaffold HTML/JavaScript template in `assets/tradingview/chart_bridge.html` loading TradingView Lightweight Charts library.
    - Implement `TradingViewBridgeWidget` using `InAppWebView` or `webview_flutter`, establishing a typed bidirectional `JavaScriptChannel` (`GrowwwChartBridge`).
    - Implement JavaScript handler methods for `setCandles()`, `updateTick()`, `setIndicators()`, and `applyTheme()`.
11. **Implement Chart Repository & Cache Pipeline:** Build `ChartRepository` coordinating REST candle pagination from TimescaleDB (Prompt 408), real-time WebSocket tick ingestion (Prompt 207), and local caching using `Isar` to enable instant offline chart restoration.
12. **Implement Riverpod Chart State Notifiers:** Create `ChartConfigNotifier`, `CandleHistoryNotifier`, `LiveTickStreamNotifier`, `IndicatorSelectionNotifier`, and `DrawingToolsNotifier` to manage chart resolution, active indicators, and real-time tick merging.
13. **Build Adaptive Chart Container Widget:** Create `TradingChartContainerWidget` providing seamless switching between the Native CustomPainter engine (default on mobile handsets) and the TradingView Bridge engine (activated on tablets, desktop, or landscape full-screen mode).
14. **Build Corporate Action Blockchain Inspection Modal:** Construct `CorporateActionVerificationSheet` displaying event details, Besu transaction hash, block height, and depository audit confirmation hash.
15. **Write Unit and Performance Tests:** Author test suites in `test/features/charting/`:
    - Unit tests validating mathematical accuracy of RSI, MACD, Bollinger Bands, and Volume Profile against reference datasets.
    - Golden tests verifying visual rendering of candles, gridlines, and crosshairs.
    - Benchmark tests verifying sustained 60/120 FPS rendering under 1,000 ticks/sec synthetic load with zero UI frame drops.

## Interfaces / Contracts

```dart
// lib/features/charting/domain/models/candle_data.dart

class CandleData {
  final DateTime timestamp;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;
  final double turnover;

  const CandleData({
    required this.timestamp,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
    required this.turnover,
  });

  bool get isBullish => close >= open;
  double get bodyRange => (close - open).abs();
  double get totalRange => high - low;

  CandleData copyWithUpdatedTick({
    required double tickPrice,
    required double tickVolume,
    required double tickTurnover,
  }) {
    return CandleData(
      timestamp: timestamp,
      open: open,
      high: tickPrice > high ? tickPrice : high,
      low: tickPrice < low ? tickPrice : low,
      close: tickPrice,
      volume: volume + tickVolume,
      turnover: turnover + tickTurnover,
    );
  }
}

// lib/features/charting/domain/models/chart_enums.dart

enum ChartResolution {
  oneSecond('1s', 1),
  oneMinute('1m', 60),
  fiveMinutes('5m', 300),
  fifteenMinutes('15m', 900),
  oneHour('1h', 3600),
  oneDay('1D', 86400),
  oneWeek('1W', 604800);

  final String label;
  final int durationSeconds;
  const ChartResolution(this.label, this.durationSeconds);
}

enum ChartStyle {
  candlestick,
  hollowCandlestick,
  heikinAshi,
  areaLine,
  baseline,
}

enum IndicatorType {
  ema,
  rsi,
  macd,
  bollingerBands,
  vwap,
}

enum DrawingToolType {
  cursor,
  trendline,
  horizontalRay,
  fibonacciRetracement,
  verticalLine,
  rectangleHighlight,
}

// lib/features/charting/domain/models/level2_depth_model.dart

class Level2DepthLevel {
  final double price;
  final double quantity;
  final int orderCount;
  final double cumulativeVolume;

  const Level2DepthLevel({
    required this.price,
    required this.quantity,
    required this.orderCount,
    required this.cumulativeVolume,
  });
}

class MarketDepthSnapshot {
  final String symbol;
  final List<Level2DepthLevel> bids;
  final List<Level2DepthLevel> asks;
  final double maxCumulativeVolume;
  final DateTime timestamp;

  const MarketDepthSnapshot({
    required this.symbol,
    required this.bids,
    required this.asks,
    required this.maxCumulativeVolume,
    required this.timestamp,
  });
}

// lib/features/charting/domain/models/volume_profile_model.dart

class VolumeProfileBucket {
  final double lowerPrice;
  final double upperPrice;
  final double buyVolume;
  final double sellVolume;
  final double totalVolume;

  const VolumeProfileBucket({
    required this.lowerPrice,
    required this.upperPrice,
    required this.buyVolume,
    required this.sellVolume,
    required this.totalVolume,
  });
}

class VolumeProfileResult {
  final List<VolumeProfileBucket> buckets;
  final double pointOfControlPrice; // POC
  final double valueAreaHighPrice;   // VAH
  final double valueAreaLowPrice;    // VAL
  final double totalTradedVolume;

  const VolumeProfileResult({
    required this.buckets,
    required this.pointOfControlPrice,
    required this.valueAreaHighPrice,
    required this.valueAreaLowPrice,
    required this.totalTradedVolume,
  });
}

// lib/features/charting/domain/models/corporate_action_marker.dart

enum CorporateActionType {
  cashDividend,
  stockSplit,
  bonusIssue,
  rightsOffering,
}

class CorporateActionMarker {
  final String id;
  final String symbol;
  final DateTime exDate;
  final CorporateActionType actionType;
  final String title;
  final String description;
  final double? dividendAmountInr;
  final String? splitRatio;
  final String smartContractAddress;
  final String transactionHash;
  final int blockNumber;
  final String depositoryReference;

  const CorporateActionMarker({
    required this.id,
    required this.symbol,
    required this.exDate,
    required this.actionType,
    required this.title,
    required this.description,
    this.dividendAmountInr,
    this.splitRatio,
    required this.smartContractAddress,
    required this.transactionHash,
    required this.blockNumber,
    required this.depositoryReference,
  });
}

// lib/features/charting/domain/models/drawing_entity.dart

class ChartCoordinate {
  final DateTime time;
  final double price;

  const ChartCoordinate({
    required this.time,
    required this.price,
  });
}

class ChartDrawingEntity {
  final String id;
  final DrawingToolType type;
  final List<ChartCoordinate> coordinates;
  final int strokeColorHex;
  final double strokeWidth;
  final bool isLocked;

  const ChartDrawingEntity({
    required this.id,
    required this.type,
    required this.coordinates,
    required this.strokeColorHex,
    this.strokeWidth = 2.0,
    this.isLocked = false,
  });
}

// lib/features/charting/domain/interfaces/i_technical_indicator_calculator.dart

abstract class ITechnicalIndicatorCalculator<T> {
  T calculate(List<CandleData> candles);
}

class MacdOutput {
  final List<double?> macdLine;
  final List<double?> signalLine;
  final List<double?> histogram;

  const MacdOutput({
    required this.macdLine,
    required this.signalLine,
    required this.histogram,
  });
}

class BollingerBandsOutput {
  final List<double?> upperBand;
  final List<double?> middleBand;
  final List<double?> lowerBand;

  const BollingerBandsOutput({
    required this.upperBand,
    required this.middleBand,
    required this.lowerBand,
  });
}

// lib/features/charting/domain/interfaces/i_chart_repository.dart

abstract class IChartRepository {
  Future<List<CandleData>> fetchHistoricalCandles({
    required String symbol,
    required ChartResolution resolution,
    required DateTime from,
    required DateTime to,
  });

  Stream<CandleData> streamRealtimeCandleUpdates({
    required String symbol,
    required ChartResolution resolution,
  });

  Stream<MarketDepthSnapshot> streamMarketDepth({
    required String symbol,
  });

  Future<List<CorporateActionMarker>> fetchCorporateActionMarkers({
    required String symbol,
  });
}

// lib/features/charting/data/datasources/tradingview_js_bridge_contracts.dart

abstract class ITradingViewJsBridge {
  void postCandles(List<CandleData> candles);
  void postTickUpdate(CandleData updatedCandle);
  void setChartResolution(ChartResolution resolution);
  void setChartTheme(String themeName);
  void toggleIndicator(IndicatorType type, bool isEnabled);
}
```

## Security & Compliance Notes
- **Client-Side Mathematical Accuracy & Statutory Integrity:** Technical indicators, especially RSI and Bollinger Bands, directly inform retail and institutional trading decisions. Implementing flawed mathematical formulas (such as simple moving averages instead of Wilder's Exponential Smoothing for RSI, or biased sample variance for Bollinger Bands) exposes the platform to regulatory liability under SEBI (Investment Advisers) Regulations and SEBI Master Circular on algorithmic and digital interfaces. All calculation algorithms must undergo unit testing against reference standard outputs from quantitative financial benchmarks.
- **Tamper-Evident Tick Validation:** High-frequency market data streams transmitted over WebSockets can be susceptible to packet injection or sequence tampering in compromised network environments. The charting data ingestion pipeline validates sequential monotonically increasing sequence numbers emitted by the exchange gateway (Prompt 207). Ticks arriving out of order or with duplicate timestamps are quarantined, preventing chart distortion or synthetic flash-crash visual anomalies.
- **Sandboxed WebView & Content Security Policy (CSP):** The TradingView Lightweight Charts bridge runs inside a sandboxed platform view (`InAppWebView` / platform view iframe). To prevent cross-site scripting (XSS) and arbitrary script execution:
  - The embedded HTML page enforces a strict Content Security Policy (`default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'`).
  - Local file system access, geolocation, camera, and external web navigation are explicitly disabled on the WebView controller.
  - The bidirectional communication bridge passes strictly validated, typed JSON payloads over the registered `JavaScriptChannel`, rejecting raw strings or executable code chunks.
- **Battery & Thermal Throttling Protection:** Rendering continuous 120 FPS canvas updates under heavy market data bursts can rapidly deplete mobile battery reserves and trigger hardware thermal throttling. The charting engine enforces a VSYNC-aligned render throttle: regardless of whether 500 or 5,000 ticks arrive per second, the canvas repaints strictly on display refresh intervals (16.6ms or 8.3ms). When the application transitions to the background or the chart screen loses focus, an `AppLifecycleListener` pauses WebSocket subscriptions and terminates animation tickers immediately.
- **Zero On-Chain PII Policy:** In strict compliance with Indian Digital Personal Data Protection (DPDP) Act and SEBI data guidelines, on-chain corporate action verification queries query only immutable corporate event parameters (ISIN, ex-date, dividend amount, split ratio, smart contract address). No demat account numbers, investor PANs, or individual portfolio holdings are written to or queried from the blockchain ledger.

## Acceptance Criteria
- [ ] `TradingChartContainerWidget` renders multi-timeframe Candlestick, Hollow Candlestick, Heikin-Ashi, and Area/Line charts with sub-pixel clarity on Android, iOS, macOS, Windows, Linux, and Web.
- [ ] Native Flutter `CustomPainter` engine maintains a verified sustained 60 FPS on standard devices and 120 FPS on high-refresh displays during active market streaming with zero visual stutter.
- [ ] Embedded TradingView Lightweight Charts bridge successfully loads, synchronizes theme colors, and ingests live ticks via typed bidirectional JavaScript messages.
- [ ] Level-2 market depth ladder renders real-time cumulative bid and ask order histograms projecting horizontally from the price scale.
- [ ] Visible-Range Volume Profile (VRVP) dynamically calculates and renders Point of Control (POC), Value Area High (VAH), and Value Area Low (VAL) across the active viewport.
- [ ] Vectorized technical indicator suite accurately computes and displays EMA (9, 21, 50, 200), RSI (14 with 70/30 levels), MACD (12/26/9 with histogram), Bollinger Bands (20, 2-sigma), and VWAP.
- [ ] Interactive drawing tools layer allows traders to draw, move, and edit trendlines, horizontal rays, and Fibonacci retracements, maintaining locked normalized $(t, p)$ coordinates across panning and zooming.
- [ ] Interactive touch gestures handle independent horizontal time-zoom, vertical price-zoom, smooth kinetic panning, and long-press crosshair inspection with magnetic candle snapping.
- [ ] Corporate action markers render on the time axis at exact ex-dates, opening an inspection sheet displaying Besu block heights, smart contract addresses, and NSDL/CDSL depository verification hashes.
- [ ] High-frequency tick aggregation flushes forming candle updates strictly on VSYNC frame callbacks without dropping historical candles.
- [ ] Backgrounding the application or navigating away immediately suspends WebSocket tick streaming and cancels animation tickers to prevent battery drain.
- [ ] Full specification adheres strictly to the 12-section template, containing zero code implementation in the workspace, and uses standard ASCII hyphens exclusively.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Multi-Platform Project Scaffolding), Prompt 502 (Flutter App Architecture & Riverpod State Management), Prompt 503 (Flutter Design System & Theming), Prompt 525 (Flutter API Client Layer).
- **Backend Dependencies:** Prompt 408 (Historical Market Data TimescaleDB Aggregation Pipeline), Prompt 207 (Market Data Service & Real-Time WebSocket Broadcast Gateway), Prompt 222 (Corporate Actions & Dividends Service).
- **Blockchain Dependencies:** Prompt 306 (Digital Security Token ERC-3643 Engine), Prompt 308 (Proof-of-Reserve Registry Contract on Hyperledger Besu).
- **Downstream Blockers:** Prompt 508 (Flutter Security Detail Screen), Prompt 509 (Flutter Order Placement Flow), Prompt 529 (Flutter Options Chain & Derivatives Screen), Prompt 530 (Flutter Commodity Physical Delivery Flow), Prompt 902 (Cross-Platform End-to-End Test Suite).
