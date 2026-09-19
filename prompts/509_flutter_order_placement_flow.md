# 509 - Flutter Order Placement Flow (Buy/Sell, Order Types, Confirmation)

## Purpose
Implements the mission-critical, ultra-low-latency order execution workflow for fractional equities and securities. Enabling investors to purchase Indian blue-chip stocks with as little as ₹100 requires a seamless order intake flow that supports both currency-denominated (e.g., invest ₹500) and quantity-denominated (e.g., buy 0.245 shares) inputs, real-time pre-trade margin checks, transparent fee breakdowns (zero holding fee; Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), with FIFO capital gains calculated strictly for user tax compliance under Section 111A/112A), cryptographic Delivery-versus-Payment (DvP) settlement badges, and biometric confirmation.

## What You Are Building
A high-converting, error-proof order placement module in `apps/growww_flutter/lib/features/order_placement/` featuring:
- **Responsive Order Sheet / Modal:** Bottom sheet on mobile and side-docked order ticket on desktop with Buy (Teal/Green) and Sell (Red) dynamic styling.
- **Dual Input Mode Converter:** Seamless toggle between INR Amount input (₹) and Fractional Quantity input (Units up to 4 decimal places).
- **Order Types Selection:** Market Order, Limit Order (with price stepper), and Stop-Loss Limit Order.
- **Pre-Trade Margin & Balance Validator:** Instant real-time check against available INR cash wallet balance.
- **Order Cost & Fee Breakdown Summary:** Detailed view showing Estimated Total, Stamp Duty, GST, Exchange Turnover Fee, and clear "0% Holding Fee" highlight.
- **Swipe-to-Confirm / Biometric Execution Slider:** Interactive slider with haptic feedback to prevent accidental trade submissions.
- **Atomic DvP Settlement Modal:** Live status animation tracking order submission -> matching -> smart contract escrow lock -> custody allocation.

## Scope Boundaries
- **In Scope:**
 - Order entry form, input calculations, pre-trade validations, biometric trade confirmation, order sheet lifecycle, and settlement progress UI.
- **Out of Scope / Handled Elsewhere:**
 - Backend order matching engine (handled in Prompt 205).
 - Risk and margin service validation (handled in Prompt 206).
 - Smart contract DvP settlement execution (handled in Prompt 306).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` state management.
 - *Justification:* Instant state reactivity guarantees that changes in the quantity input automatically recalculate estimated margin, taxes, and fractional shares without input lag.
- **Biometrics & Security:** `local_auth` for high-value trade biometric sign-off.
- **Haptics:** `flutter/services` HapticFeedback for tactile confirmation during swipe actions.

## Backend / Infra Touchpoints
- **Order Service:** REST endpoints (`/api/v1/orders/create`, `/api/v1/orders/estimate`) from Prompt 204.
- **Risk & Margin Service:** Pre-trade margin verification (`/api/v1/risk/pre-trade-check`) from Prompt 206.
- **Wallet Service:** Available cash balance query (`/api/v1/wallet/balance`) from Prompt 203.
- **Fee Engine:** Fee estimation endpoint (`/api/v1/fees/estimate`) from Prompt 210.

## Blockchain Interaction
- **DvP Smart Contract Escrow Guarantee:**
 - Displays the "Atomic DvP Settlement" badge confirming that the trade is governed by `SettlementDvP.sol`.
 - Upon order fill, the post-trade status screen displays the atomic swap lifecycle:
    1. INR Fiat Hold in Regulated Escrow ->
    2. Smart Contract Token Mint / Allocation ->
    3. Immutable On-Chain DvP Receipt with Transaction Hash.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/order_placement/`: `presentation/sheets/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/`.
2. Define domain models: `OrderPlacementRequest`, `OrderEstimateResult`, `OrderType`, `OrderSide`, `SettlementStatus`.
3. Implement `OrderPlacementController` managing form state, calculations, and submission life-cycle.
4. Build `OrderPlacementSheet` with dynamic theme coloring (Green for Buy, Crimson for Sell).
5. Implement `AmountQuantityToggle` allowing frictionless switching between ₹ INR value and fractional share quantity.
6. Build `OrderTypeSelector` handling Market, Limit (with price input), and Stop-Loss conditions.
7. Implement real-time `PreTradeEstimateCalculator` with 150ms debounce, fetching fee breakdown from the backend fee engine.
8. Build `FeeBreakdownAccordion` displaying statutory charges and highlighting the Growww Universal Zero-Fee Model (0.00% fee - No fee at all) (with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol) and zero holding fees), and note FIFO capital gains compliance (Section 111A/112A).
9. Implement `SwipeToConfirmButton` with animated chevron gradient and haptic pulse upon reaching the completion threshold.
10. Integrate biometric prompt requirement for orders exceeding a user-configurable threshold (e.g., >₹50,000).
11. Build `OrderSettlementProgressModal` showing 4-stage progress indicators: (1) Order Placed, (2) Matched, (3) DvP Escrow Locked, (4) Units Credited.
12. Write widget tests verifying margin check errors, decimal input precision, and swipe gesture completions.

## Interfaces / Contracts
```dart
// lib/features/order_placement/domain/models/order_request.dart
enum OrderSide { buy, sell }
enum OrderType { market, limit, stopLossLimit }

class OrderPlacementRequest {
  final String symbol;
  final String isin;
  final OrderSide side;
  final OrderType type;
  final double? inrAmount;
  final double? quantity; // Fractional units up to 4 decimal places
  final double? limitPrice;
  final double? stopPrice;
  final String idempotencyKey;

  const OrderPlacementRequest({
    required this.symbol,
    required this.isin,
    required this.side,
    required this.type,
    this.inrAmount,
    this.quantity,
    this.limitPrice,
    this.stopPrice,
    required this.idempotencyKey,
  });

  Map<String, dynamic> toJson() => {
    'symbol': symbol,
    'isin': isin,
    'side': side.name.toUpperCase(),
    'order_type': type.name.toUpperCase(),
    'inr_amount': inrAmount,
    'quantity': quantity,
    'limit_price': limitPrice,
    'stop_price': stopPrice,
    'idempotency_key': idempotencyKey,
  };
}
```

## Security & Compliance Notes
- **Idempotency Protection:** Every order placement generates a cryptographically random UUIDv4 `idempotency_key` client-side to prevent duplicate order submissions during network retries.
- **Pre-Trade Balance Verification:** Client blocks order submission if available funds are insufficient, preventing unnecessary order rejection penalties.
- **SEBI Order Confirmation Rules:** The order summary explicitly displays the exact price, estimated statutory taxes (STT, Stamp Duty), and execution venue before submission.

## Acceptance Criteria
- [ ] Order sheet toggles seamlessly between Buy and Sell modes with correct thematic styling.
- [ ] Switching between INR amount and fractional quantity updates calculations with 4-decimal accuracy.
- [ ] Swipe-to-confirm gesture triggers haptic feedback and submits order payload with unique idempotency key.
- [ ] Insufficient balance triggers an immediate "Deposit Funds" CTA without resetting the order parameters.
- [ ] Order status screen displays real-time DvP settlement progress and blockchain transaction hash.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System), Prompt 505 (Authentication).
- **Backend Dependency:** Prompt 204 (Order Service), Prompt 206 (Risk Check), Prompt 210 (Fee Engine), Prompt 306 (DvP Contract).
- **Enables:** Prompt 510 (Portfolio Holdings), Prompt 512 (Transaction History).
