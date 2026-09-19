# 508 - Flutter Security Detail Screen (TradingView-Style Charts)

## Purpose
Provides an advanced, deep-dive analytical view for any listed security (e.g., RELIANCE, TCS, HDFCBANK). Investors require professional charting capabilities (candlesticks, moving averages, RSI, volume profile), full market depth (5-level bid/ask order book), company fundamentals, corporate action timelines, and an immutable on-chain Proof-of-Reserve audit tab confirming the physical custodian share backing of the tokenized asset.

## What You Are Building
A comprehensive asset analytics screen in `apps/growww_flutter/lib/features/security_detail/` including:
- **Interactive Multi-Timeframe Chart:** High-performance candlestick and line chart with timeframes (1D, 1W, 1M, 1Y, 5Y, All) and pinch-to-zoom / pan gestures.
- **Technical Indicators Overlay:** Toggleable indicators including Exponential Moving Averages (EMA 20/50/200), Volume bars, RSI, and MACD.
- **Crosshair & Tooltip Inspector:** Interactive crosshair displaying exact timestamp, Open, High, Low, Close (OHLC), and Volume values.
- **Market Depth (Level 2 Order Book):** Visual bid/ask depth ladder displaying buy/sell quantities, orders count, and cumulative depth bars.
- **Key Financial Metrics & Fundamentals:** P/E Ratio, Market Cap, 52-Week High/Low, Dividend Yield, and Sector classification.
- **On-Chain Custody & Proof-of-Reserve Tab:** Displays the current smart contract token supply, physical shares held in NSDL/CDSL custody, auditor attestation hash, and link to on-chain block explorer.
- **Persistent Bottom Action Bar:** Instant "Buy" and "Sell" buttons launching the order sheet.

## Scope Boundaries
- **In Scope:**
 - Interactive chart rendering, OHLC aggregation, indicator calculations, depth ladder visualization, fundamentals display, and Proof-of-Reserve tab.
- **Out of Scope / Handled Elsewhere:**
 - Order placement modal execution (handled in Prompt 509).
 - Historical market data ingestion and storage (handled in Prompt 207).
 - On-chain custodian attestation publishing (handled in Prompt 308).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with custom Canvas rendering or `interactive_chart`.
 - *Justification:* Drawing candlesticks directly onto Flutter's hardware-accelerated Skia/Impeller `Canvas` via `CustomPainter` avoids the heavyweight overhead and cross-origin iframe security constraints of embedded web-view charting engines, enabling buttery-smooth 120 FPS scrubbing.
- **State Management:** Riverpod `FutureProvider.family` and `StreamProvider.family` scoped to the security's ISIN/symbol.
- **Math & Technical Indicators:** Client-side vector calculation for EMA, RSI, and MACD.

## Backend / Infra Touchpoints
- **Market Data Service:** Historical candle endpoint (`/api/v1/market/history?symbol=RELIANCE&resolution=1D`) and live depth feed (`wss://ws.growww.in/v1/market/depth/RELIANCE`) from Prompt 207.
- **Corporate Actions & Fundamentals API:** REST endpoint (`/api/v1/market/fundamentals/RELIANCE`) from Prompt 222.
- **Custody Attestation API:** Endpoint (`/api/v1/custody/attestation/RELIANCE`) from Prompt 213 & 308.

## Blockchain Interaction
- **On-Chain Custody Verification View:**
 - Connects to `DigitalSecurityToken.sol` and `ProofOfReserveRegistry.sol`.
 - Displays: Total On-Chain Token Supply (e.g., 500,000.0000 units) == Total Physical Shares in Custody (500,000 shares held in NSDL Depository Account #IN300123).
 - Shows the latest SHA-256 Merkle root hash, auditor verification signature, and transaction hash with a direct link to the Besu block explorer.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/security_detail/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/`.
2. Define domain models: `SecurityDetail`, `CandleData`, `OrderBookDepth`, `CompanyFundamentals`, `CustodyAuditRecord`.
3. Implement `CandleChartPainter` extending `CustomPainter` to draw bullish (green) and bearish (red) candle bodies, wicks, and volume bars.
4. Implement gesture recognition on the chart widget for multi-touch pinch-to-zoom (scaling X-axis), drag-to-pan, and long-press crosshair inspection.
5. Implement timeframe selector bar (1D, 1W, 1M, 1Y, 5Y, ALL) triggering dynamic data resolution fetching (1-min, 5-min, daily candles).
6. Implement technical indicator calculations (EMA 20/50, RSI 14, MACD 12/26/9) and render them as customizable canvas overlays.
7. Build `OrderBookDepthWidget` displaying real-time 5-level bid (buy) and ask (sell) queues with horizontal progress bars representing relative volume.
8. Build `FundamentalsGrid` rendering key investment ratios (P/E, Market Cap, 52W Range, PB, Beta).
9. Build `OnChainProofOfReserveTab` rendering the custody verification card, backing ratio badge (100%), auditor digital signature, and smart contract explorer link.
10. Build `CorporateActionsTimeline` showing past and upcoming dividend payouts, stock splits, and bonus issues.
11. Implement sticky bottom bar with responsive "Buy" and "Sell" buttons opening the order sheet.
12. Write widget tests verifying chart touch handling, timeframe data switching, and on-chain hash verification display.

## Interfaces / Contracts
```dart
// lib/features/security_detail/domain/models/candle_data.dart
class CandleData {
  final DateTime timestamp;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;

  const CandleData({
    required this.timestamp,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
  });
}

class CustodyAuditRecord {
  final String isin;
  final String contractAddress;
  final double totalTokenSupply;
  final double totalPhysicalSharesInCustody;
  final String depositoryName; // NSDL or CDSL
  final String custodianName;
  final String latestAttestationTxHash;
  final DateTime verifiedAt;
  final bool isBackingRatioValid; // Must be true (1.0 ratio)

  const CustodyAuditRecord({
    required this.isin,
    required this.contractAddress,
    required this.totalTokenSupply,
    required this.totalPhysicalSharesInCustody,
    required this.depositoryName,
    required this.custodianName,
    required this.latestAttestationTxHash,
    required this.verifiedAt,
    required this.isBackingRatioValid,
  });
}
```

## Security & Compliance Notes
- **SEBI Mandatory Disclosures:** Past performance graphs must explicitly state: *"Past performance is not indicative of future returns."*
- **Audit Hash Tamper-Check:** The client recalculates the SHA-256 hash of the custody report data and verifies it against the on-chain Merkle root before showing the green "Verified" badge.
- **Accurate Decimal Precision:** Candlestick prices and fractional quantities maintain strict 4-decimal precision (e.g., ₹2,450.7525) to prevent rounding discrepancies.

## Acceptance Criteria
- [ ] Candlestick chart smoothly pans, zooms, and scrubs at 60/120 FPS across mobile and desktop.
- [ ] Timeframe switching correctly loads and renders corresponding candle resolutions.
- [ ] Level 2 market depth ladder streams real-time bid/ask quantities over WebSocket.
- [ ] Proof-of-Reserve tab accurately displays 1:1 custody backing ratio and verifiable blockchain transaction link.
- [ ] Buy and Sell buttons launch the order placement flow with pre-populated asset details.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System).
- **Backend Dependency:** Prompt 207 (Market Data), Prompt 213 (Custodian Adapter), Prompt 308 (Proof of Reserve).
- **Enables:** Prompt 509 (Order Placement Flow).
