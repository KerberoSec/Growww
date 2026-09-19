# 529 - Flutter Options Chain and Derivatives Screen

## Purpose
Provides an institutional-grade, continuous 24/7 derivatives trading interface for Indian tokenized equities, indices, and perpetual commodities. Retail and institutional traders require real-time options chain visualization (Calls, Puts, Strike Ladder, Greeks, Implied Volatility, Open Interest, multi-leg payoff diagrams) and perpetual futures execution (dynamic leverage sliders, cross/isolated margin modes, real-time liquidation price calculation, 8-hour funding rate countdown, and mark vs index price tracking).

## What You Are Building
A high-performance derivatives trading module in `apps/growww_flutter/lib/features/derivatives/` comprising:
- **Real-Time Options Chain Matrix:** Split-table Call and Put layout flanking a central strike price ladder with dynamic in-the-money (ITM), at-the-money (ATM), and out-of-the-money (OTM) color coding, streaming bid/ask quotes, Open Interest (OI) depth bars, Implied Volatility (IV) metrics, and option Greeks (Delta, Gamma, Theta, Vega, Rho).
- **Interactive Multi-Leg Strategy Payoff Chart:** Hardware-accelerated visual profit/loss curve plotting payoff at expiry, identifying exact breakeven strikes, maximum upside, maximum downside risk, and risk-reward ratios for single and multi-leg strategies (straddles, strangles, spreads, iron condors).
- **Perpetual Futures Execution Terminal:** Order entry interface with a granular leverage slider (1x to 100x), margin mode selection (Isolated Margin vs Cross Margin), real-time liquidation price computation, and dynamic maintenance margin requirement calculations.
- **Funding Rate & Mark Price Monitor:** Real-time 8-hour funding rate countdown timer, annualized funding rate percentage, and side-by-side display of Mark Price, Index Price, and Last Traded Price (LTP).
- **Liquidation Risk Meter & Warning System:** Dynamic health gauge visualizing current collateral buffer, proximity-to-liquidation alert badges, and high-leverage risk confirmation dialogs.
- **Active Positions & Open Orders Manager:** Real-time list of active derivative contracts, live unrealized PnL updates with color flash animations, margin utilization percentages, and one-tap market close / bracket order controls.

## Scope Boundaries
- **In Scope:**
  - Options chain table layout, strike ladder filtering, and expiry date selection.
  - Greeks display and Implied Volatility model state integration.
  - Interactive multi-leg payoff chart rendering and breakeven calculation visualization.
  - Perpetual futures leverage slider, margin mode toggle, liquidation estimator, and order configuration.
  - Riverpod state providers, notifiers, and WebSocket stream subscription managers.
  - SEBI/IFSCA statutory risk warnings and high-leverage protective alerts.
- **Out of Scope / Handled Elsewhere:**
  - Backend Black-Scholes / Bjerksund-Stensland pricing engine and volatility surface calculation (handled in Prompt 207 / 229).
  - Pre-trade VaR margin validation and liquidation execution engine (handled in Prompt 206 / 229 / 230).
  - Smart contract DvP settlement and on-chain perpetual clearing vault (handled in Prompt 306 / 315).
  - Generic spot equity order execution modal (handled in Prompt 509).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` (v2.5+) using code-generated `AsyncNotifier` and `StreamNotifier` patterns.
  - *Justification:* Granular reactive state isolation enables sub-50ms price updates to refresh isolated table cells or payoff chart painters without triggering whole-screen widget tree rebuilds.
- **Custom Canvas Rendering:** `CustomPainter` on Flutter Impeller/Skia engine for 120 FPS payoff curves, volatility smiles, and horizontal Open Interest comparison bars.
- **High-Throughput Streaming:** `web_socket_channel` with binary Protocol Buffer (Protobuf) deserialization for low-latency market data ticks.
- **Fixed-Point Financial Mathematics:** `decimal` package for exact financial arithmetic, margin calculations, and liquidation thresholds without floating-point rounding errors.

## Backend / Infra Touchpoints
- **Derivatives Market Data Stream:** WebSocket endpoint (`wss://ws.growww.in/v1/derivatives/stream`) for real-time options chain quotes, Greeks, IV, mark prices, index prices, and funding rates (from Prompt 207).
- **Options Chain Snapshot API:** REST endpoint (`GET /api/v1/derivatives/options/chain`) returning strike ladders, open interest distribution, and Greeks snapshot (from Prompt 207).
- **Perpetual Futures Market API:** REST endpoint (`GET /api/v1/derivatives/perpetuals/summary`) returning funding rates, next funding timestamp, open interest, and index reference prices (from Prompt 207).
- **Pre-Trade Risk & Margin Simulation API:** REST endpoint (`POST /api/v1/risk/margin/simulate`) validating required initial margin, maintenance margin, and liquidation price (from Prompt 206 / 229).
- **Derivatives Order Execution API:** REST endpoint (`POST /api/v1/orders/derivatives`) for placing options and perpetual futures orders (from Prompt 204).

## Blockchain Interaction
- **On-Chain Collateral & Perpetual Vault Verification:**
  - Connects to `PerpetualClearingVault.sol` and `SettlementDvP.sol` on the permissioned Hyperledger Besu network.
  - Displays smart contract vault address, total collateral locked (digital INR / USDC backing), and on-chain solvency attestation proof.
  - Validates that derivative positions and margin accounts are backed 1:1 by segregated custodial reserves with zero PII recorded on-chain.

## State Management Architecture
- **Options Chain Stream Provider:** Subscribes to real-time WebSocket ticks for all strikes in the active underlying asset and selected expiry.
- **Options Chain Filter State Provider:** Manages selected underlying symbol, active expiry date, strike count window (e.g., ATM +/- 10 strikes), and display column mode (LTP, Greeks, OI, IV).
- **Strategy Payoff Notifier Provider:** Computes payoff coordinates, zero-line intersections, max profit, and max loss across combined option legs.
- **Perpetual Trading Notifier Provider:** Manages leverage level (1x to 100x), margin mode (Cross vs Isolated), order type (Limit, Market, Stop), limit price, order size, dynamic collateral requirement, and computed liquidation price.
- **Derivatives Positions Stream Provider:** Streams open positions, live unrealized PnL, margin utilization ratios, and distance-to-liquidation percentages.
- **Funding Rate Stream Provider:** Streams live funding rates and drives the local high-precision countdown timer to the next 8-hour funding settlement.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `apps/growww_flutter/lib/features/derivatives/`: `presentation/screens/`, `presentation/widgets/`, `presentation/controllers/`, `domain/models/`, `domain/services/`, `data/repositories/`, `data/sources/`.
2. Define domain model data contracts for options chains: `OptionContract`, `OptionGreeks`, `OptionStrikeLadder`, `OptionsChainSnapshot`, and `OptionPayoffPoint`.
3. Define domain model data contracts for perpetual futures: `PerpetualMarketInfo`, `PerpetualPosition`, `LeverageConfig`, `MarginMode`, and `LiquidationEstimate`.
4. Define Riverpod state providers and AsyncNotifier/StreamNotifier signatures for options chain streaming, perpetual state management, and payoff computation.
5. Create `OptionsChainScreen` containing top-level underlying selector, expiry date horizontal chip carousel, and Call/Put split table view.
6. Implement `StrikeLadderTable` widget with sticky center strike column, ITM/ATM/OTM background shading, and real-time bid/ask quote updates.
7. Build `OptionsGreeksView` toggle displaying Delta, Gamma, Theta, Vega, Rho, and Implied Volatility per strike.
8. Build `OpenInterestBarChart` rendering call OI versus put OI horizontal visual bars for each strike price.
9. Implement `PayoffChartPainter` extending `CustomPainter` to draw multi-leg profit/loss payoff curves, zero-line breakeven markers, and current underlying price indicator.
10. Build `PerpetualFuturesScreen` featuring order type selector (Limit, Market, Stop-Market), size input in contracts/INR, and dynamic margin calculator.
11. Implement `LeverageSliderWidget` with preset snap stops (1x, 2x, 5x, 10x, 25x, 50x, 100x), visual risk tier color grading, and high-leverage risk confirmation dialog.
12. Build `LiquidationWarningBanner` and `MarginHealthMeter` dynamically displaying liquidation price, distance to liquidation percentage, and maintenance margin ratio.
13. Implement `FundingRateTimer` widget displaying the current 8-hour funding rate (annualized percentage) and real-time countdown to the next funding settlement.
14. Build `ActivePositionsSheet` displaying open options and perpetual positions with live unrealized PnL, margin utilization, and quick-close buttons.
15. Write unit and widget test contracts verifying state transitions, leverage slider calculations, liquidation warning triggers, and WebSocket tick updates.

## Interfaces / Contracts
```dart
// lib/features/derivatives/domain/models/option_contract.dart

enum OptionType { call, put }

enum MoneynessType { inTheMoney, atTheMoney, outOfTheMoney }

enum MarginMode { cross, isolated }

enum DerivativeOrderType { limit, market, stopMarket, stopLimit }

class OptionGreeks {
  final double delta;
  final double gamma;
  final double theta;
  final double vega;
  final double rho;
  final double impliedVolatility;

  const OptionGreeks({
    required this.delta,
    required this.gamma,
    required this.theta,
    required this.vega,
    required this.rho,
    required this.impliedVolatility,
  });
}

class OptionContractQuote {
  final String contractId;
  final String symbol;
  final OptionType optionType;
  final double strikePrice;
  final DateTime expiryDate;
  final double lastPrice;
  final double priceChange;
  final double priceChangePercent;
  final double bidPrice;
  final double bidQuantity;
  final double askPrice;
  final double askQuantity;
  final double openInterest;
  final double openInterestChange;
  final double volume;
  final OptionGreeks greeks;
  final MoneynessType moneyness;

  const OptionContractQuote({
    required this.contractId,
    required this.symbol,
    required this.optionType,
    required this.strikePrice,
    required this.expiryDate,
    required this.lastPrice,
    required this.priceChange,
    required this.priceChangePercent,
    required this.bidPrice,
    required this.bidQuantity,
    required this.askPrice,
    required this.askQuantity,
    required this.openInterest,
    required this.openInterestChange,
    required this.volume,
    required this.greeks,
    required this.moneyness,
  });
}

class OptionStrikeRow {
  final double strikePrice;
  final OptionContractQuote callQuote;
  final OptionContractQuote putQuote;

  const OptionStrikeRow({
    required this.strikePrice,
    required this.callQuote,
    required this.putQuote,
  });
}

class OptionsChainSnapshot {
  final String underlyingSymbol;
  final double underlyingSpotPrice;
  final List<DateTime> availableExpiries;
  final DateTime selectedExpiry;
  final List<OptionStrikeRow> strikeRows;
  final double totalCallOpenInterest;
  final double totalPutOpenInterest;
  final double putCallRatio;
  final DateTime timestamp;

  const OptionsChainSnapshot({
    required this.underlyingSymbol,
    required this.underlyingSpotPrice,
    required this.availableExpiries,
    required this.selectedExpiry,
    required this.strikeRows,
    required this.totalCallOpenInterest,
    required this.totalPutOpenInterest,
    required this.putCallRatio,
    required this.timestamp,
  });
}

class OptionLegStrategyItem {
  final OptionContractQuote contract;
  final int quantity; // Positive for Buy/Long, Negative for Sell/Short
  final double entryPremium;

  const OptionLegStrategyItem({
    required this.contract,
    required this.quantity,
    required this.entryPremium,
  });
}

class OptionPayoffPoint {
  final double spotPriceAtExpiry;
  final double netProfitLoss;

  const OptionPayoffPoint({
    required this.spotPriceAtExpiry,
    required this.netProfitLoss,
  });
}

class OptionStrategyPayoff {
  final List<OptionLegStrategyItem> legs;
  final List<OptionPayoffPoint> payoffCurve;
  final List<double> breakevenPrices;
  final double maxProfit; // Double.infinity if uncapped
  final double maxLoss;   // Double.infinity if uncapped
  final double netPremiumPaidOrReceived;

  const OptionStrategyPayoff({
    required this.legs,
    required this.payoffCurve,
    required this.breakevenPrices,
    required this.maxProfit,
    required this.maxLoss,
    required this.netPremiumPaidOrReceived,
  });
}

class PerpetualMarketInfo {
  final String perpetualSymbol;
  final String underlyingAsset;
  final double markPrice;
  final double indexPrice;
  final double lastTradedPrice;
  final double fundingRate8h;
  final double fundingRateAnnualized;
  final DateTime nextFundingTimestamp;
  final double openInterest;
  final double volume24h;
  final int maxLeverage;
  final double maintenanceMarginPercent;

  const PerpetualMarketInfo({
    required this.perpetualSymbol,
    required this.underlyingAsset,
    required this.markPrice,
    required this.indexPrice,
    required this.lastTradedPrice,
    required this.fundingRate8h,
    required this.fundingRateAnnualized,
    required this.nextFundingTimestamp,
    required this.openInterest,
    required this.volume24h,
    required this.maxLeverage,
    required this.maintenanceMarginPercent,
  });
}

class LiquidationEstimate {
  final double liquidationPrice;
  final double distancePercentToLiquidation;
  final double requiredInitialMargin;
  final double requiredMaintenanceMargin;
  final double bankruptcyPrice;
  final bool isHighRiskWarning;

  const LiquidationEstimate({
    required this.liquidationPrice,
    required this.distancePercentToLiquidation,
    required this.requiredInitialMargin,
    required this.requiredMaintenanceMargin,
    required this.bankruptcyPrice,
    required this.isHighRiskWarning,
  });
}

class PerpetualPosition {
  final String positionId;
  final String perpetualSymbol;
  final int leverage;
  final MarginMode marginMode;
  final double positionSizeContracts;
  final double entryPrice;
  final double markPrice;
  final double liquidationPrice;
  final double marginAllocated;
  final double unrealizedPnL;
  final double unrealizedPnLPercentage;
  final double maintenanceMarginRequirement;
  final double marginHealthRatio;

  const PerpetualPosition({
    required this.positionId,
    required this.perpetualSymbol,
    required this.leverage,
    required this.marginMode,
    required this.positionSizeContracts,
    required this.entryPrice,
    required this.markPrice,
    required this.liquidationPrice,
    required this.marginAllocated,
    required this.unrealizedPnL,
    required this.unrealizedPnLPercentage,
    required this.maintenanceMarginRequirement,
    required this.marginHealthRatio,
  });
}

abstract class IDerivativesRepository {
  Future<OptionsChainSnapshot> fetchOptionsChainSnapshot({
    required String underlyingSymbol,
    required DateTime expiryDate,
  });

  Future<PerpetualMarketInfo> fetchPerpetualMarketInfo({
    required String perpetualSymbol,
  });

  Future<LiquidationEstimate> calculateLiquidationRisk({
    required String perpetualSymbol,
    required int leverage,
    required MarginMode marginMode,
    required double orderSize,
    required double entryPrice,
  });

  Future<List<PerpetualPosition>> fetchActivePerpetualPositions();

  Stream<OptionContractQuote> streamOptionsContractTicks({
    required String contractId,
  });

  Stream<PerpetualMarketInfo> streamPerpetualMarketTicks({
    required String perpetualSymbol,
  });
}
```

## Security & Compliance Notes
- **SEBI Mandatory Risk Warnings:** Prominent, persistent risk disclosure banner displayed at the top of the derivatives screen: *"Risk Warning: 9 out of 10 individual traders in equity Futures and Options segment incur net financial losses."*
- **Leverage Guardrails & Risk Warnings:** Selecting leverage above 10x requires explicit user confirmation via an alert modal detailing liquidation risk, spread slippage, and maintenance margin depletion.
- **Decimal Precision Compliance:** All strike prices, premiums, margins, and collateral figures are calculated using high-precision decimal representation to eliminate floating-point calculation errors.
- **Zero PII on Blockchain:** Smart contract perpetual clearing vault records strictly anonymized trade hashes and cryptographic account identifiers with zero customer identity exposure.

## Acceptance Criteria
- [ ] Options chain renders live strike prices, Calls, Puts, Greeks, and IV with sub-50ms update latency over WebSocket.
- [ ] ITM, ATM, and OTM strikes are visually distinguished using standard theme color shading.
- [ ] Multi-leg payoff chart accurately plots profit and loss curves and flags exact breakeven price points.
- [ ] Perpetual futures leverage slider smoothly adjusts from 1x to 100x with real-time recalculation of required margin and liquidation price.
- [ ] Liquidation price warnings prominently trigger when market price approaches within 5% of liquidation threshold.
- [ ] Funding rate timer accurately counts down to the 8-hour settlement window and displays the current rate.
- [ ] Active positions sheet reflects live unrealized PnL and enables one-tap market exit order submission.
- [ ] Sustained 60/120 FPS rendering during high-frequency market data tick bursts without UI frame drops.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System), Prompt 505 (Authentication).
- **Backend Dependency:** Prompt 204 (Order Service), Prompt 206 (Risk Engine), Prompt 207 (Market Data Service), Prompt 229 (Real-Time VaR Margin Engine).
- **Blockchain Dependency:** Prompt 306 (Settlement DvP Smart Contract), Prompt 315 (Settlement Guarantee Fund Contract).
- **Enables:** Prompt 510 (Portfolio Holdings Screen), Prompt 512 (Trade History Screen).
