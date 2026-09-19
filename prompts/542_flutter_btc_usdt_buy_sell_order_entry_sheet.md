# 542 - Flutter BTC/USDT Buy/Sell Order Entry Bottom Sheet & Execution Modal

## Purpose
The Flutter BTC/USDT Buy/Sell Order Entry Bottom Sheet (`lib/screens/trading/order_entry_sheet.dart`) is the primary trade placement interaction component for retail and professional traders on both mobile (Android/iOS) and desktop (macOS/Windows/Linux). It provides a responsive, low-latency, and error-resilient modal sheet supporting high-velocity order entry for Bitcoin spot trading across both **Demo / Paper Trading Mode (Testnet)** and **Real-Money Trading Mode (Mainnet)**.

Trading digital assets requires absolute precision, transparent fee accounting, and explicit safeguards against accidental order placement or catastrophic fat-finger errors. This specification establishes the architecture, state lifecycle, form validation, real-time fee calculation, biometric authorization gating, and reactive WebSocket feedback mechanisms for placing Market, Limit, Stop-Limit, and Trailing-Stop orders.

---

## What You Are Building
A modular, production-ready Flutter widget suite comprising:
1. **`OrderEntryBottomSheet`**: The draggable modal bottom sheet (mobile) and docked side-panel (desktop/tablet) that presents the trade composition form.
2. **`OrderTypeSelector`**: Tabbed segment controller switching between Limit, Market, Stop-Limit, and Trailing Stop execution modes.
3. **`BuySellToggle`**: High-contrast dual-state pill toggle clearly demarcating Buy (Emerald Green: `#00C087`) and Sell (Crimson Red: `#FF3B30`) actions.
4. **`PriceAndQuantityInput`**: Number fields formatted with custom `TabularFiguresFormatter`, supporting sub-penny and 8-decimal Satoshi precision with micro-stepper increment buttons.
5. **`PercentageAllocationSlider`**: Interactive percentage chips (25%, 50%, 75%, 100%) dynamically calculating allowable lot sizes against available free collateral.
6. **`FeeAndTotalSummaryCard`**: Real-time breakdown calculating the gross notional, the mandatory platform flat **0.00% fee (No fee at all)** (0 bps at launch / 0.0000), and net settlement amounts.
7. **`OrderSubmissionButton`**: An animated, high-visibility submit button with integrated biometric authorization (Face ID / Touch ID / Local Auth) for Real-Money Mainnet execution, and instant submission for Demo Testnet execution.

---

## Scope Boundaries

### In-Scope
- Reactive price pre-population from the live `MarketDepthProvider` and `TickerStream`.
- Dynamic balance binding: Virtual balances (`vUSDT` / `vBTC`) when in Demo Mode; real balances (`USDT` / `BTC`) when in Real Mode.
- Mathematical precision using `dart:typed_data` and fixed-point scaled integers (paise / satoshis) to prevent floating-point rounding errors.
- Pre-trade validation: minimum order size (10 USDT), maximum tick boundaries, and balance adequacy checks.
- Dual-target submission: routes to `services/demo-matching-engine` (gRPC / HTTP) in Demo Mode and `services/btc-order-service` in Mainnet Mode.
- Biometric verification gating before placing Mainnet real-money orders exceeding configured safety thresholds.
- Zero-PII state logging and strict ASCII character encoding.

### Out-of-Scope
- Complex multi-leg derivatives or options spread builders (covered in Prompt 529).
- Direct fiat on-ramp KYC verification (covered in Prompt 504 and Prompt 543).
- Low-level network socket maintenance (delegated to the typed API client layer, Prompt 525).

---

## Technology to Use
- **Framework:** Flutter 3.22+ / Dart 3.4+
- **State Management:** `flutter_riverpod` (v2.5+) with code generation (`@riverpod`)
- **Animations:** Flutter built-in explicit animation controllers (`AnimationController`, `CurvedAnimation`)
- **Biometric Security:** `local_auth` (v2.2+) with hardware enclave fallback
- **Formatting:** `intl` package with custom `Decimal` / `BigInt` financial formatting utilities
- **Haptics:** `flutter/services.dart` (`HapticFeedback.mediumImpact()` on button presses)

---

## Backend / Infra Touchpoints
- **Demo Mode Target:** `services/demo-matching-engine` via Envoy gRPC-Web proxy (`/nbse.demo.v1.DemoOrderService/PlaceOrder`).
- **Real Mode Target:** `services/btc-order-service` via API Gateway (`/nbse.order.v1.BtcOrderService/PlaceOrder`).
- **Collateral & Balance Stream:** `services/wallet-service` (Real) and `services/demo-wallet-service` (Demo) over WebSocket topic `account:balance:stream`.
- **Live Ticker & Depth Stream:** `services/depth-broadcaster` publishing Level-2 top-of-book quotes over `market:btc_usdt:depth`.

---

## Blockchain Interaction
- **Demo Mode:** Zero direct blockchain interaction from UI; trades emit simulated execution receipts on Besu Testnet via the backend relayer.
- **Real Mode:** The backend matches orders and queues them for atomic Delivery-versus-Payment (DvP) settlement on Hyperledger Besu Mainnet via `SettlementDvP.sol`. The UI displays the pending transaction hash and links to the internal block explorer once confirmed.

---

## Step-by-Step Build Instructions

### Step 1: Data Model and Enums Definition
Define strict, immutable Dart enums and models representing order attributes:
- `OrderSide`: `buy`, `sell`
- `OrderType`: `limit`, `market`, `stopLimit`, `trailingStop`
- `TimeInForce`: `goodTillCancelled` (GTC), `immediateOrCancel` (IOC), `fillOrKill` (FOK)
- `OrderValidationState`: `valid`, `insufficientBalance`, `belowMinimumNotional`, `invalidPriceTick`

### Step 2: Currency and Fixed-Point Arithmetic Utilities
Create `FinancialFormatter` utilizing `BigInt` scaling:
- Bitcoin amounts scaled by $10^8$ (1 Satoshi = 0.00000001 BTC).
- USDT amounts scaled by $10^6$ (1 micro-USDT = 0.000001 USDT).
- Guarantee that all arithmetic operations (Notional = Price * Quantity, Fee = Notional * 0.0000 (0% at launch)) are executed using fixed-point math to avoid IEEE 754 precision loss.

### Step 3: Riverpod State Notifier Setup (`OrderEntryNotifier`)
Implement the state management controller managing:
- Selected order side, type, price input, quantity input, and active percentage chip.
- Reactive listener to `currentTradingModeProvider` (`demo` vs `real`).
- Reactive listener to `liveTickerProvider` to pre-fill current market price on side change.
- Computed getters for `notionalValue`, `estimatedFee` (0.00% (Zero Fee)), and `netTotalRequired`.

### Step 4: Bottom Sheet Scaffolding & Responsive Layout
Construct the responsive container widget:
- On mobile devices ($width < 600$), present as an expandable modal bottom sheet using `showModalBottomSheet()` with `isScrollControlled: true` and keyboard inset padding.
- On desktop/tablet devices ($width \ge 600$), render as a docked persistent right-side execution drawer within the primary trading screen.

### Step 5: High-Contrast Buy/Sell Pill Toggle
Implement the segmented pill toggle:
- Emerald Green (`#00C087`) background with white text when Buy is active.
- Crimson Red (`#FF3B30`) background with white text when Sell is active.
- Trigger haptic feedback on tab change. Automatically adjust available balance display (USDT balance for Buy, BTC balance for Sell).

### Step 6: Order Type Tab Bar
Implement horizontal chips for selecting execution type:
- **Limit:** Exposes Price and Quantity fields.
- **Market:** Disables the Price field, displaying `"Best Market Price"` with real-time slippage warning indicator.
- **Stop-Limit:** Exposes Stop Trigger Price, Limit Price, and Quantity fields.

### Step 7: Price Input Field with Micro-Steppers
Construct the price entry field:
- Display tick size increment buttons (`+` and `-`) adjusting by $0.10 USDT per tap or continuous increment on long-press.
- Pre-populate with current Best Bid (for Sell) or Best Ask (for Buy) when the user taps on order book depth rows.

### Step 8: Quantity Input Field and Satoshi Unit Formatter
Construct the quantity entry field:
- Supports numeric input up to 8 decimal places.
- Incorporates a secondary display label showing the approximate fiat equivalent (in INR or USD) based on the latest FX rate.

### Step 9: Percentage Allocation Quick-Select Chips
Build 4 quick-select chips (25%, 50%, 75%, 100%):
- Tapping a chip calculates the exact maximum allowable quantity:
  $$\text{Quantity} = \frac{\text{AvailableBalance} \times \text{Percentage}}{\text{Price} \times (1 + \text{FeeRate})}$$
- For 100% Buy orders, reserve the fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) in free collateral so the transaction does not fail due to insufficient fee margin.

### Step 10: Fee and Total Breakdown Card
Implement an expandable summary card showing:
- **Gross Order Value:** $\text{Quantity} \times \text{Price}$
- **Exchange Platform Fee (0.00% (Zero Fee)):** Explicitly labelled `"Platform Fee (0.00% (No fee at all) / Zero Brokerage)"`.
- **Net Required Collateral:** $\text{Gross Value} + \text{Fee}$ (for Buy) or $\text{Gross Value} - \text{Fee}$ (for Sell).
- Mode Badge: Displays `"Testnet Virtual Escrow"` or `"Mainnet Besu DvP Settlement"`.

### Step 11: Biometric and Multi-Factor Security Gate
Before dispatching a Real-Money Mainnet order:
- If the order value exceeds 1,000 USDT or if `"Require Biometrics for Orders"` is enabled in settings, invoke `local_auth` to authenticate the user via fingerprint or facial recognition.
- If in Demo Mode, bypass biometric gating and provide instantaneous one-tap submission.

### Step 12: Order Submission and Optimistic UI Update
Implement the submit action:
- Generate a unique client-side UUIDv4 `Idempotency-Key`.
- Display a micro-spinner inside the button.
- Dispatch the gRPC / REST request to the appropriate backend service.
- On success, emit a brief toast notification and trigger a slight vibration.
- On error (e.g., price band breach or insufficient margin), display an inline amber alert banner without closing the sheet.

---

## Interfaces / Contracts

### Dart PlaceOrderRequest Model
```dart
class PlaceOrderRequest {
  final String idempotencyKey;
  final String accountId;
  final String symbol; // "BTC/USDT"
  final OrderSide side; // buy | sell
  final OrderType type; // limit | market | stopLimit
  final BigInt priceScaled; // Scaled by 10^6 (USDT micro-units)
  final BigInt quantityScaled; // Scaled by 10^8 (Satoshis)
  final BigInt? stopPriceScaled;
  final TimeInForce timeInForce;
  final bool isDemo;

  PlaceOrderRequest({
    required this.idempotencyKey,
    required this.accountId,
    required this.symbol,
    required this.side,
    required this.type,
    required this.priceScaled,
    required this.quantityScaled,
    this.stopPriceScaled,
    required this.timeInForce,
    required this.isDemo,
  });

  Map<String, dynamic> toJson() => {
    'idempotency_key': idempotencyKey,
    'account_id': accountId,
    'symbol': symbol,
    'side': side.name.toUpperCase(),
    'type': type.name.toUpperCase(),
    'price_scaled': priceScaled.toString(),
    'quantity_scaled': quantityScaled.toString(),
    if (stopPriceScaled != null) 'stop_price_scaled': stopPriceScaled.toString(),
    'time_in_force': timeInForce.name.toUpperCase(),
    'is_demo': isDemo,
  };
}
```

### Order Placement Response Contract
```json
{
  "order_id": "ord_btc_883921049281",
  "client_order_id": "c9284729-2810-4f93-bc81-192837465920",
  "symbol": "BTC/USDT",
  "side": "BUY",
  "type": "LIMIT",
  "price": "64250.500000",
  "quantity": "0.15000000",
  "executed_quantity": "0.00000000",
  "status": "ACCEPTED",
  "fee_rate": "0.0000",
  "estimated_fee": "0.000000",
  "is_demo": true,
  "created_at_utc": "2026-09-19T09:25:00.124Z"
}
```

---

## Security & Compliance Notes
1. **Accidental Order Safeguards:** Slippage warning modal triggered automatically if a Market Order has estimated market impact $> 0.5\%$.
2. **Strict Fee Transparency:** Displays exact 0.00% fee (No fee at all) upfront; prohibits hidden markups or spreads.
3. **Environment Segregation:** The active mode (`demo` vs `real`) is cryptographically checked against the auth token header (`X-NBSE-Environment: TESTNET` or `MAINNET`).
4. **Biometric Security:** Real-money trades can require biometric confirmation, preventing unauthorized order entry on unlocked devices.
5. **Zero-PII Telemetry:** Client crash logs and analytics never record user account IDs or trade dollar amounts.

---

## Acceptance Criteria
- [ ] Sheet renders correctly on mobile bottom modal and desktop docked panel without UI overflow.
- [ ] Buy/Sell toggle transitions seamlessly with emerald green / crimson red color shifts and haptic feedback.
- [ ] Percentage chips calculate the exact maximum quantity reserving the 0.00% (Zero Fee) platform fee without rounding drift.
- [ ] Tapping order book depth entries instantly updates the Price input field with sub-5ms UI responsiveness.
- [ ] Demo Mode places simulated orders against `services/demo-matching-engine` without requiring biometrics.
- [ ] Real Mode routes orders to `services/btc-order-service` and verifies biometric sign-off for orders exceeding threshold.
- [ ] All monetary calculations use fixed-point arithmetic with zero floating-point imprecision.
- [ ] Client uses standard ASCII hyphens exclusively across all code and documentation.

---

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 272 (Market Feeder), Prompt 273 (Demo Engine), Prompt 275 (Spot Order Service), Prompt 540 (Flutter Chart Screen).
- **Parallel Work:** Prompt 541 (Mode Switcher Banner), Prompt 543 (Crypto Transfer Modal).
- **Downstream Consumers:** Prompt 620 (Web Trading Terminal), Prompt 918 (BTC/USDT E2E Testing Suite).
