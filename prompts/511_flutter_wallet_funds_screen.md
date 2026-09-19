# 511 - Flutter Wallet & Funds Screen (Deposit/Withdraw, UPI/Bank Linking)

## Purpose
Provides a secure, frictionless fiat currency (INR) and foreign currency (USD for GIFT City investors) management interface. In strict adherence to RBI and SEBI regulations, all cash deposits, withdrawals, and bank account links are routed exclusively through RBI-licensed banking rails, UPI payment aggregators, and scheduled commercial banks - never through unregulated crypto on-ramps or stablecoins. This screen delivers real-time fund tracking, instant UPI deposit intent launching, and Penny Drop bank verification.

## What You Are Building
A comprehensive funds and banking management screen in `apps/growww_flutter/lib/features/wallet/` including:
- **Available Cash Balance Hero Card:** Total Available Cash to Trade, Funds Locked in Active Orders, Funds Pending T+1 Settlement, and Unsettled Realized Profit.
- **Deposit Funds Flow:** Instant UPI App Intent launcher (GPay, PhonePe, Paytm, BHIM), dynamic UPI QR code generator for desktop, NetBanking portal selector, and NEFT/RTGS virtual account details.
- **Withdraw Funds Flow:** Bank selection, withdrawal amount input, RBI-mandated settlement cooling period countdown, and biometric withdrawal authorization.
- **Linked Bank Accounts Manager:** Multi-bank manager showing Primary Account, IFSC, verified bank logo, and instant "Add Bank via Penny Drop" verification flow.
- **GIFT City Currency Converter (International Leg):** USD-to-INR conversion rate calculator with real-time RBI Reference Rate display for GIFT City onboarding.

## Scope Boundaries
- **In Scope:**
 - Wallet balance dashboard UI, UPI app intent integration, dynamic QR code rendering, bank account linking UI, withdrawal request form, and biometric authorization.
- **Out of Scope / Handled Elsewhere:**
 - Backend payment aggregator webhooks and banking API adapters (handled in Prompt 212).
 - Internal fiat ledger accounting service (handled in Prompt 203).
 - Foreign investor FX conversion settlement service (handled in Prompt 214).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` state management.
 - *Justification:* URL launcher integration on mobile invokes native UPI applications via deep-link intents (`upi://pay`), while rendering dynamic vector QR codes via `qr_flutter` on desktop platforms.
- **UPI Deep Linking:** `url_launcher` configured for Android/iOS UPI application scheme handling.
- **QR Code Rendering:** `qr_flutter` for rendering crisp, real-time UPI payment QR codes.
- **Biometric Signing:** `local_auth` for securing withdrawal requests.

## Backend / Infra Touchpoints
- **Wallet Service:** REST endpoints (`/api/v1/wallet/balances`, `/api/v1/wallet/ledger`) from Prompt 203.
- **Payment Gateway Service:** UPI Intent creation (`/api/v1/payments/upi/create-order`), Webhook polling (`/api/v1/payments/status/:orderId`) from Prompt 212.
- **Bank Verification Service:** Penny drop API (`/api/v1/banking/penny-drop`) from Prompt 212.

## Blockchain Interaction
- **DvP Fiat Escrow Hold Visualization:**
 - Displays the fiat balance locked in Delivery-versus-Payment escrow during active settlement windows.
 - Emphasizes the regulatory guarantee that fiat funds remain in regulated RBI banking escrows while token allocations are synchronized atomically on `SettlementDvP.sol`.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/wallet/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/`.
2. Define domain models: `WalletBalance`, `LinkedBankAccount`, `DepositOrder`, `WithdrawalRequest`, `PaymentMethod`.
3. Implement `WalletController` managing real-time balance polling and deposit transaction status polling.
4. Build `WalletBalanceCard` with breakdowns: Available to Invest, Withdrawable, and Blocked for Trades.
5. Implement `DepositSheet` with quick-amount chips (+₹1,000, +₹5,000, +₹10,000, +₹25,000) and payment method selector.
6. Build `UpiIntentLauncher` detecting installed UPI apps on mobile devices (Google Pay, PhonePe, Paytm, BHIM) and launching via intent URL.
7. Build `DesktopUpiQrWidget` for Windows/macOS/Linux displaying a live UPI QR code with a 5-minute expiry countdown timer.
8. Build `WithdrawalScreen` validating against minimum/maximum limits and withdrawable balance.
9. Implement biometric authorization prompt before final withdrawal dispatch.
10. Build `LinkedBanksScreen` displaying verified bank accounts with primary badge and option to initiate Penny Drop verification for a new bank account.
11. Build `PennyDropVerificationModal` showing step-by-step progress: (1) ₹1 Deposit Sent, (2) Name Matched with KYC PAN, (3) Account Verified.
12. Write widget tests for UPI intent handling, deposit amount limits, and bank account validation.

## Interfaces / Contracts
```dart
// lib/features/wallet/domain/models/wallet_models.dart
class WalletBalance {
  final double totalCashInr;
  final double availableToTradeInr;
  final double withdrawableInr;
  final double lockedInOrdersInr;
  final double pendingSettlementInr;

  const WalletBalance({
    required this.totalCashInr,
    required this.availableToTradeInr,
    required this.withdrawableInr,
    required this.lockedInOrdersInr,
    required this.pendingSettlementInr,
  });
}

class UpiPaymentOrder {
  final String orderId;
  final double amountInr;
  final String upiIntentString; // upi://pay?pa=...
  final String qrCodeData;
  final int expirySeconds;

  const UpiPaymentOrder({
    required this.orderId,
    required this.amountInr,
    required this.upiIntentString,
    required this.qrCodeData,
    required this.expirySeconds,
  });
}
```

## Security & Compliance Notes
- **Strict Third-Party Deposit Ban:** Deposits are accepted ONLY from bank accounts registered in the investor's verified KYC name. The UI explicitly alerts users: *"Deposits from unlinked or third-party bank accounts will be automatically rejected and refunded."*
- **AML Daily Limits:** Daily and monthly deposit limits are strictly enforced on the client side according to the user's KYC tier.
- **Biometric Protected Withdrawals:** Fund withdrawals require mandatory MPIN or biometric confirmation to prevent unauthorized drain.

## Acceptance Criteria
- [ ] Wallet balances correctly reflect available, locked, and withdrawable funds.
- [ ] On mobile, tapping UPI apps launches the native UPI intent.
- [ ] On desktop, dynamic UPI QR code generates and updates on payment confirmation via WebSocket/polling.
- [ ] Penny Drop bank linking validates IFSC, account number, and displays name matching status.
- [ ] Withdrawal flow enforces cooling periods and prompts for biometric confirmation.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System), Prompt 505 (Authentication).
- **Backend Dependency:** Prompt 203 (Wallet Service), Prompt 212 (Payment Gateway), Prompt 214 (FX Service).
- **Enables:** Prompt 509 (Order Placement Flow), Prompt 512 (Transaction History).
